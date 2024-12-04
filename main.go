package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Type definitions
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Size        string    `json:"size"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Quantity    int       `json:"quantity"`
	SellerID    string    `json:"seller_id"`
	SellerName  string    `json:"seller_name"`
	Images      []string  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
}

type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	AddressID   string    `json:"address_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderStatus struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Address struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Street    string    `json:"street"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	ZipCode   string    `json:"zip_code"`
	Country   string    `json:"country"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderItem struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	SellerID   string  `json:"seller_id"`
	SellerName string  `json:"seller_name"`
}

type Message struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	SenderID  string    `json:"sender_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type UnreadMessage struct {
	OrderID       string    `json:"order_id"`
	ID            string    `json:"id"`
	LatestMessage string    `json:"latest_message"`
	Timestamp     time.Time `json:"latest_timestamp"`
	Count         int       `json:"count"`
}

// Constants
const (
	maxFileSize = 10 << 20 // 10MB
	maxImages   = 3
	uploadDir   = "./uploads"
)

var db *sql.DB

func initDB() {
	connStr := "postgres://killian@localhost/marketplace?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully connected to database")
}

// Helper functions
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	return token.SignedString([]byte("your-secret-key"))
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveImage(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s%s",
		uuid.New().String(), filepath.Ext(fileHeader.Filename))

	// Save to full path
	fullPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		return "", err
	}

	// Return relative path for database storage
	return filepath.Join("uploads", filename), nil
}

// Auth Handlers
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(creds.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userID string
	err = db.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id`,
		creds.Name, creds.Email, hashedPassword).Scan(&userID)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			http.Error(w, "Email already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := generateToken(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{
		"token":   token,
		"user_id": userID,
		"name":    creds.Name,
		"email":   creds.Email,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow(`
		SELECT id, name, email, password_hash
		FROM users
		WHERE email = $1`,
		creds.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !checkPasswordHash(creds.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{
		"token":   token,
		"user_id": user.ID,
		"name":    user.Name,
	})
}

func getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	var user User
	err := db.QueryRow(`
			SELECT id, name, email
			FROM users
			WHERE id = $1`,
		userID).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, user)
}

func getUserNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/users/")

	var name string
	err := db.QueryRow("SELECT name FROM users WHERE id = $1", userID).Scan(&name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"name": name})
}

// Item Handlers
func searchItemsHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	category := r.URL.Query().Get("category")
	size := r.URL.Query().Get("size")
	minPrice := r.URL.Query().Get("min_price")
	maxPrice := r.URL.Query().Get("max_price")

	sqlQuery := `
      SELECT i.id, i.title, i.description, i.price, i.size, i.category,
             i.status, i.quantity, i.seller_id, u.name as seller_name,
             i.created_at, array_agg(im.image_path) as images
      FROM items i
      LEFT JOIN item_images im ON i.id = im.item_id
      JOIN users u ON i.seller_id = u.id
      WHERE 1=1`

	var params []interface{}
	paramCount := 1

	if query != "" {
		sqlQuery += fmt.Sprintf(` AND (LOWER(i.title) LIKE $%d OR LOWER(i.description) LIKE $%d)`, paramCount, paramCount)
		params = append(params, "%"+query+"%")
		paramCount++
	}

	if category != "" {
		sqlQuery += fmt.Sprintf(` AND i.category = $%d`, paramCount)
		params = append(params, category)
		paramCount++
	}

	if size != "" {
		sqlQuery += fmt.Sprintf(` AND i.size = $%d`, paramCount)
		params = append(params, size)
		paramCount++
	}

	if minPrice != "" {
		sqlQuery += fmt.Sprintf(` AND i.price >= $%d`, paramCount)
		params = append(params, minPrice)
		paramCount++
	}

	if maxPrice != "" {
		sqlQuery += fmt.Sprintf(` AND i.price <= $%d`, paramCount)
		params = append(params, maxPrice)
		paramCount++
	}

	sqlQuery += ` GROUP BY i.id, u.name
              ORDER BY
                (CASE WHEN i.quantity > 0 THEN 0 ELSE 1 END),
                i.created_at DESC`

	rows, err := db.Query(sqlQuery, params...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var images []sql.NullString
		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.Price,
			&item.Size, &item.Category, &item.Status, &item.Quantity,
			&item.SellerID, &item.SellerName, &item.CreatedAt, pq.Array(&images))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		item.Images = make([]string, 0)
		for _, img := range images {
			if img.Valid {
				item.Images = append(item.Images, img.String)
			}
		}

		items = append(items, item)
	}

	sendJSON(w, items)
}

func createOrderNotification(orderID, userID, message string) error {
	_, err := db.Exec(`
			INSERT INTO notifications (
					user_id,
					type,
					reference_id,
					message,
					read
			) VALUES ($1, 'order_status', $2, $3, false)`,
		userID, orderID, message)
	return err
}

func createItemWithImagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var item Item
	itemData := r.FormValue("item")
	if err := json.Unmarshal([]byte(itemData), &item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item.Quantity = 1
	log.Printf("Setting initial quantity to: %d", item.Quantity)

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "At least one image required", http.StatusBadRequest)
		return
	}
	if len(files) > maxImages {
		http.Error(w, fmt.Sprintf("Maximum %d images allowed", maxImages), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var itemID string
	err = tx.QueryRow(`
        INSERT INTO items (title, description, price, size, category, seller_id, quantity, status)
        VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, 1), 'available'::item_status_enum)
        RETURNING id, quantity`,
		item.Title, item.Description, item.Price, item.Size, item.Category, userID, item.Quantity).Scan(&itemID, &item.Quantity)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting item: %v", err), http.StatusInternalServerError)
		return
	}

	var checkQuantity int
	err = tx.QueryRow(`SELECT quantity FROM items WHERE id = $1`, itemID).Scan(&checkQuantity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking quantity: %v", err), http.StatusInternalServerError)
		return
	}

	if checkQuantity != 1 {
		_, err = tx.Exec(`UPDATE items SET quantity = 1 WHERE id = $1`, itemID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fixing quantity: %v", err), http.StatusInternalServerError)
			return
		}
	}

	var imagePaths []string
	for _, fileHeader := range files {
		imagePath, err := saveImage(fileHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imagePaths = append(imagePaths, imagePath)

		_, err = tx.Exec(`
			INSERT INTO item_images (item_id, image_path)
			VALUES ($1, $2)`,
			itemID, imagePath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var createdItem Item
	var images []sql.NullString
	err = db.QueryRow(`
		SELECT i.id, i.title, i.description, i.price, i.size,
			   i.category, i.status, i.quantity, i.seller_id, u.name as seller_name, i.created_at,
			   array_agg(im.image_path) as images
		FROM items i
		LEFT JOIN item_images im ON i.id = im.item_id
		JOIN users u ON i.seller_id = u.id
		WHERE i.id = $1
		GROUP BY i.id, u.name`,
		itemID).Scan(
		&createdItem.ID, &createdItem.Title, &createdItem.Description,
		&createdItem.Price, &createdItem.Size, &createdItem.Category,
		&createdItem.Status, &createdItem.Quantity, &createdItem.SellerID,
		&createdItem.SellerName, &createdItem.CreatedAt, pq.Array(&images))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdItem.Images = make([]string, 0)
	for _, img := range images {
		if img.Valid {
			createdItem.Images = append(createdItem.Images, img.String)
		}
	}

	sendJSON(w, createdItem)
}

// Cart Handlers
func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user is the seller
	var sellerID string
	err := db.QueryRow(`SELECT seller_id FROM items WHERE id = $1`, req.ItemID).Scan(&sellerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sellerID == userID {
		http.Error(w, "Cannot purchase your own item", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var quantity int
	err = tx.QueryRow(`
      SELECT quantity FROM items
      WHERE id = $1 AND status = 'available'`,
		req.ItemID).Scan(&quantity)

	if err == sql.ErrNoRows {
		http.Error(w, "Item not found or unavailable", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if quantity <= 0 {
		http.Error(w, "Item out of stock", http.StatusBadRequest)
		return
	}

	var cartCount int
	err = tx.QueryRow(`
      SELECT COUNT(*) FROM cart_items
      WHERE user_id = $1 AND item_id = $2`,
		userID, req.ItemID).Scan(&cartCount)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if cartCount >= quantity {
		http.Error(w, "Cannot add more of this item - quantity limit reached", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(`
      INSERT INTO cart_items (user_id, item_id)
      VALUES ($1, $2)`,
		userID, req.ItemID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	viewCartHandler(w, r)
}

func viewCartHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	rows, err := db.Query(`
		SELECT i.id, i.title, i.description, i.price, i.size,
			   i.category, i.status, i.quantity, i.seller_id,
			   u.name as seller_name, i.created_at,
			   array_agg(im.image_path) as images
		FROM cart_items c
		JOIN items i ON c.item_id = i.id
		LEFT JOIN item_images im ON i.id = im.item_id
		JOIN users u ON i.seller_id = u.id
		WHERE c.user_id = $1
		GROUP BY i.id, u.name`,
		userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var images []sql.NullString
		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.Price,
			&item.Size, &item.Category, &item.Status, &item.Quantity,
			&item.SellerID, &item.SellerName, &item.CreatedAt, pq.Array(&images))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		item.Images = make([]string, 0)
		for _, img := range images {
			if img.Valid {
				item.Images = append(item.Images, img.String)
			}
		}

		items = append(items, item)
	}

	sendJSON(w, items)
}

func removeFromCartHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
		DELETE FROM cart_items
		WHERE user_id = $1 AND item_id = $2`,
		userID, req.ItemID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	viewCartHandler(w, r)
}

// Address Handlers
func saveAddressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	var address Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if address.IsDefault {
		_, err = tx.Exec(`
			UPDATE addresses
			SET is_default = false
			WHERE user_id = $1`,
			userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.QueryRow(`
		INSERT INTO addresses (user_id, street, city, state, zip_code, country, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		userID, address.Street, address.City, address.State,
		address.ZipCode, address.Country, address.IsDefault).Scan(&address.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, address)
}

func deleteAddressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	addressID := r.URL.Query().Get("id")

	result, err := db.Exec(`
			UPDATE addresses
			SET deleted_at = CURRENT_TIMESTAMP
			WHERE id = $1
			AND user_id = $2
			AND NOT EXISTS (
					SELECT 1 FROM orders
					WHERE address_id = addresses.id
					AND status NOT IN ('delivered', 'cancelled')
			)`,
		addressID, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Address not found or cannot be deleted (active orders exist)", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getUserAddressesHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	rows, err := db.Query(`
			SELECT id, street, city, state, zip_code, country, is_default, created_at
			FROM addresses
			WHERE user_id = $1
			AND deleted_at IS NULL
			ORDER BY is_default DESC, created_at DESC`,
		userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var addresses []Address
	for rows.Next() {
		var addr Address
		err := rows.Scan(
			&addr.ID, &addr.Street, &addr.City, &addr.State,
			&addr.ZipCode, &addr.Country, &addr.IsDefault, &addr.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		addresses = append(addresses, addr)
	}

	sendJSON(w, addresses)
}

func archiveOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	orderID := r.URL.Query().Get("order_id")

	// If specific order ID provided, archive just that order
	if orderID != "" {
		_, err := db.Exec(`
					UPDATE orders
					SET archived = true
					WHERE id = $1
					AND (
							user_id = $2
							OR EXISTS (
									SELECT 1 FROM order_items oi
									JOIN items i ON oi.item_id = i.id
									WHERE oi.order_id = orders.id
									AND i.seller_id = $2
							)
					)
					AND status IN ('delivered', 'cancelled')`,
			orderID, userID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Archive all completed orders
		_, err := db.Exec(`SELECT archive_completed_orders($1)`, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func getUnreadMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	type UnreadMessage struct {
		OrderID       string    `json:"order_id"`
		ID            string    `json:"id"`
		LatestMessage string    `json:"latest_message"`
		Timestamp     time.Time `json:"latest_timestamp"`
		Count         int       `json:"count"`
	}

	rows, err := db.Query(`
			SELECT
					m.order_id,
					m.id,
					m.message as latest_message,
					m.created_at as latest_timestamp,
					COUNT(*) OVER (PARTITION BY m.order_id) as message_count
			FROM messages m
			LEFT JOIN message_seen ms ON m.id = ms.message_id AND ms.user_id = $1
			JOIN orders o ON m.order_id = o.id
			WHERE ms.id IS NULL
			AND m.sender_id != $1
			AND (o.user_id = $1 OR EXISTS (
					SELECT 1 FROM order_items oi
					JOIN items i ON oi.item_id = i.id
					WHERE oi.order_id = o.id AND i.seller_id = $1
			))
			AND m.created_at = (
					SELECT MAX(created_at)
					FROM messages
					WHERE order_id = m.order_id
			)
			ORDER BY m.created_at DESC`,
		userID)

	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []UnreadMessage
	for rows.Next() {
		var msg UnreadMessage
		if err := rows.Scan(&msg.OrderID, &msg.ID, &msg.LatestMessage, &msg.Timestamp, &msg.Count); err != nil {
			log.Printf("Scan error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		log.Printf("Encode error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func markMessagesAsSeenHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())
	var req struct {
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
			INSERT INTO message_seen (message_id, user_id)
			SELECT m.id, $1
			FROM messages m
			WHERE m.order_id = $2 AND NOT EXISTS (
					SELECT 1 FROM message_seen ms
					WHERE ms.message_id = m.id AND ms.user_id = $1
			)`,
		userID, req.OrderID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getUnreadNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	rows, err := db.Query(`
			SELECT id, type, reference_id, message, created_at
			FROM notifications
			WHERE user_id = $1 AND read = false
			ORDER BY created_at DESC`,
		userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []struct {
		ID          string    `json:"id"`
		Type        string    `json:"type"`
		ReferenceID string    `json:"reference_id"`
		Message     string    `json:"message"`
		CreatedAt   time.Time `json:"created_at"`
	}

	for rows.Next() {
		var n struct {
			ID          string    `json:"id"`
			Type        string    `json:"type"`
			ReferenceID string    `json:"reference_id"`
			Message     string    `json:"message"`
			CreatedAt   time.Time `json:"created_at"`
		}
		if err := rows.Scan(&n.ID, &n.Type, &n.ReferenceID, &n.Message, &n.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		notifications = append(notifications, n)
	}

	sendJSON(w, notifications)
}

func markNotificationAsSeenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	notificationID := strings.TrimPrefix(r.URL.Path, "/notifications/seen/")

	_, err := db.Exec(`
			UPDATE notifications
			SET read = true
			WHERE id = $1 AND user_id = $2`,
		notificationID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func clearNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	_, err := db.Exec(`
			UPDATE notifications
			SET read = true
			WHERE user_id = $1 AND read = false`,
		userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Order Handlers
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	var req struct {
		Address struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Street    string `json:"street"`
			City      string `json:"city"`
			State     string `json:"state"`
			ZipCode   string `json:"zipCode"`
			Country   string `json:"country"`
		} `json:"address"`
		SaveAddress bool `json:"save_address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request data: %v", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var total float64
	err = tx.QueryRow(`
			SELECT COALESCE(SUM(i.price), 0)
			FROM cart_items c
			JOIN items i ON c.item_id = i.id
			WHERE c.user_id = $1`,
		userID).Scan(&total)
	if err != nil {
		log.Printf("Error calculating total: %v", err)
		http.Error(w, "Failed to calculate total", http.StatusInternalServerError)
		return
	}

	var addressID string
	log.Printf("Creating address for order with name: %s %s",
		req.Address.FirstName, req.Address.LastName)
	err = tx.QueryRow(`
			INSERT INTO addresses (
					user_id,
					first_name,
					last_name,
					street,
					city,
					state,
					zip_code,
					country,
					is_default
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
		userID,
		req.Address.FirstName,
		req.Address.LastName,
		req.Address.Street,
		req.Address.City,
		req.Address.State,
		req.Address.ZipCode,
		req.Address.Country,
		req.SaveAddress).Scan(&addressID)
	if err != nil {
		log.Printf("Error saving address: %v", err)
		http.Error(w, "Failed to save address", http.StatusInternalServerError)
		return
	}

	var orderID string
	err = tx.QueryRow(`
			INSERT INTO orders (
					user_id,
					address_id,
					total,
					status
			) VALUES ($1, $2, $3, 'pending')
			RETURNING id`,
		userID, addressID, total).Scan(&orderID)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
			INSERT INTO order_items (order_id, item_id, price_at_time)
			SELECT $1, i.id, i.price
			FROM cart_items c
			JOIN items i ON c.item_id = i.id
			WHERE c.user_id = $2`,
		orderID, userID)
	if err != nil {
		log.Printf("Error creating order items: %v", err)
		http.Error(w, "Failed to create order items", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
			UPDATE items i
			SET quantity = quantity - 1,
					status = CASE
							WHEN quantity - 1 <= 0 THEN 'sold'::item_status_enum
							ELSE status
					END
			FROM order_items oi
			WHERE oi.item_id = i.id AND oi.order_id = $1`,
		orderID)
	if err != nil {
		log.Printf("Error updating inventory: %v", err)
		http.Error(w, "Failed to update inventory", http.StatusInternalServerError)
		return
	}

	if err := notifySellers(tx, userID, orderID); err != nil {
		log.Printf("Error notifying sellers: %v", err)
	}

	_, err = tx.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("Error clearing cart: %v", err)
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to complete checkout", http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"order_id": orderID,
		"status":   "success",
	})
}

func notifySellers(tx *sql.Tx, userID, orderID string) error {
	log.Printf("Starting seller notifications for order %s", orderID)

	rows, err := tx.Query(`
			SELECT DISTINCT i.seller_id
			FROM cart_items c
			JOIN items i ON c.item_id = i.id
			WHERE c.user_id = $1`,
		userID)
	if err != nil {
		log.Printf("Error fetching sellers: %v", err)
		return fmt.Errorf("error fetching sellers: %v", err)
	}
	defer rows.Close()

	var notifiedCount int
	for rows.Next() {
		var sellerID string
		if err := rows.Scan(&sellerID); err != nil {
			log.Printf("Error scanning seller ID: %v", err)
			continue
		}
		notificationMsg := fmt.Sprintf("New order #%s received", orderID)
		if err := createOrderNotification(orderID, sellerID, notificationMsg); err != nil {
			log.Printf("Error creating notification for seller %s: %v", sellerID, err)
		} else {
			notifiedCount++
		}
	}

	log.Printf("Notified %d sellers for order %s", notifiedCount, orderID)
	return rows.Err()
}

func getUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	rows, err := db.Query(`
			SELECT
					o.id,
					o.user_id,
					o.status,
					o.created_at,
					o.updated_at,
					a.id as address_id,
					a.first_name,
					a.last_name,
					a.street,
					a.city,
					a.state,
					a.zip_code,
					a.country,
					COALESCE(
							json_agg(
									json_build_object(
											'id', i.id,
											'title', i.title,
											'price', oi.price_at_time,
											'seller_id', i.seller_id,
											'seller_name', u.name
									)
							) FILTER (WHERE i.id IS NOT NULL),
							'[]'::json
					) as items
			FROM orders o
			JOIN addresses a ON o.address_id = a.id
			LEFT JOIN order_items oi ON o.id = oi.order_id
			LEFT JOIN items i ON oi.item_id = i.id
			LEFT JOIN users u ON i.seller_id = u.id
			WHERE o.user_id = $1 OR i.seller_id = $1
			GROUP BY o.id, o.user_id, a.id, a.first_name, a.last_name, a.street, a.city,
							 a.state, a.zip_code, a.country
			ORDER BY o.created_at DESC`,
		userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrderAddress struct {
		ID        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Street    string `json:"street"`
		City      string `json:"city"`
		State     string `json:"state"`
		ZipCode   string `json:"zip_code"`
		Country   string `json:"country"`
	}

	type Order struct {
		ID        string       `json:"id"`
		UserID    string       `json:"user_id"`
		Status    string       `json:"status"`
		CreatedAt time.Time    `json:"created_at"`
		UpdatedAt time.Time    `json:"updated_at"`
		Address   OrderAddress `json:"address"`
		Items     []OrderItem  `json:"items"`
	}

	var orders []Order
	for rows.Next() {
		var o Order
		var addr OrderAddress
		var itemsJSON []byte

		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
			&addr.ID,
			&addr.FirstName,
			&addr.LastName,
			&addr.Street,
			&addr.City,
			&addr.State,
			&addr.ZipCode,
			&addr.Country,
			&itemsJSON,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		o.Address = addr
		err = json.Unmarshal(itemsJSON, &o.Items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, orders)
}

func updateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	orderID := r.URL.Query().Get("order_id")

	var req struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var buyerID string
	err = tx.QueryRow(`SELECT user_id FROM orders WHERE id = $1`, orderID).Scan(&buyerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update order status
	_, err = tx.Exec(`
			UPDATE orders
			SET status = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2`,
		req.Status, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create notification for status change
	var notificationMsg string
	switch req.Status {
	case "shipped":
		notificationMsg = fmt.Sprintf("Your order #%s has been shipped", orderID)
	case "cancelled":
		notificationMsg = fmt.Sprintf("Your order #%s has been cancelled. Reason: %s", orderID, req.Message)
	case "delivered":
		// Notify seller of delivery confirmation
		notificationMsg = fmt.Sprintf("Order #%s has been confirmed as delivered", orderID)
		// Create notification for seller
		var sellerID string
		err = tx.QueryRow(`
        SELECT DISTINCT i.seller_id
        FROM order_items oi
        JOIN items i ON oi.item_id = i.id
        WHERE oi.order_id = $1
        LIMIT 1`, orderID).Scan(&sellerID)
		if err == nil && sellerID != userID {
			err = createOrderNotification(orderID, sellerID, notificationMsg)
			if err != nil {
				log.Printf("Error creating seller notification: %v", err)
			}
		}
	}

	if notificationMsg != "" && buyerID != userID {
		err = createOrderNotification(orderID, buyerID, notificationMsg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{
		"status": "success",
	})
}

func getMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID = strings.TrimSuffix(orderID, "/messages")

	rows, err := db.Query(`
      SELECT id, sender_id, message, created_at
      FROM messages
      WHERE order_id = $1
      ORDER BY created_at ASC`,
		orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Message, &msg.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	sendJSON(w, messages)
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID = strings.TrimSuffix(orderID, "/messages")

	var msg struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
      INSERT INTO messages (order_id, sender_id, message)
      VALUES ($1, $2, $3)`,
		orderID, userID, msg.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func enableCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h(w, r)
	}
}

func authMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte("your-secret-key"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userID", userID))
		h(w, r)
	}
}

func getUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("userID").(string)
	return userID, ok
}

// Add these just before the main() function

// getUserItemsHandler()
func getUserItemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	rows, err := db.Query(`
			SELECT
					i.id,
					i.title,
					i.description,
					i.price,
					i.size,
					i.category,
					get_actual_item_status(i.id) as status,
					CASE
							WHEN get_actual_item_status(i.id) IN ('reserved', 'delivered', 'cancelled') THEN 0
							ELSE i.quantity
					END as display_quantity,
					i.seller_id,
					u.name as seller_name,
					i.created_at,
					array_agg(COALESCE(im.image_path, '')) as images,
					EXISTS (
							SELECT 1 FROM order_items oi
							JOIN orders o ON oi.order_id = o.id
							WHERE oi.item_id = i.id
							AND o.status IN ('pending', 'shipped')
					) as has_active_order
			FROM items i
			LEFT JOIN item_images im ON i.id = im.item_id
			JOIN users u ON i.seller_id = u.id
			WHERE i.seller_id = $1
			GROUP BY i.id, u.name
			ORDER BY i.created_at DESC`,
		userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var images []sql.NullString
		var hasActiveOrder bool

		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Price,
			&item.Size,
			&item.Category,
			&item.Status,
			&item.Quantity,
			&item.SellerID,
			&item.SellerName,
			&item.CreatedAt,
			pq.Array(&images),
			&hasActiveOrder,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		item.Images = make([]string, 0)
		for _, img := range images {
			if img.Valid && img.String != "" {
				item.Images = append(item.Images, img.String)
			}
		}

		items = append(items, item)
	}

	sendJSON(w, items)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	itemID := r.URL.Query().Get("id")

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// First verify the item exists and belongs to the user
	var exists bool
	err = tx.QueryRow(`
			SELECT EXISTS (
					SELECT 1 FROM items
					WHERE id = $1 AND seller_id = $2
			)`, itemID, userID).Scan(&exists)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Item not found or not authorized", http.StatusNotFound)
		return
	}

	// Delete related cart items
	_, err = tx.Exec(`DELETE FROM cart_items WHERE item_id = $1`, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete from images first
	rows, err := tx.Query("SELECT image_path FROM item_images WHERE item_id = $1", itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var imagePaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imagePaths = append(imagePaths, path)
	}

	// Delete the item images
	_, err = tx.Exec("DELETE FROM item_images WHERE item_id = $1", itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the item
	result, err := tx.Exec("DELETE FROM items WHERE id = $1 AND seller_id = $2", itemID, userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" { // Foreign key violation
				http.Error(w, "Cannot delete item: it is part of existing orders", http.StatusBadRequest)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Item not found or not authorized", http.StatusNotFound)
		return
	}

	// Delete physical image files
	for _, path := range imagePaths {
		os.Remove(filepath.Join(".", path))
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func serveImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		http.Error(w, "Image path is required", http.StatusBadRequest)
		return
	}

	// Convert relative path to absolute path
	fullPath := filepath.Join(".", imagePath)

	// Basic security check to prevent directory traversal
	if !strings.Contains(fullPath, uploadDir) && !strings.Contains(fullPath, "uploads") {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "image/jpeg") // You might want to detect the actual content type
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	http.ServeFile(w, r, fullPath)
}

func main() {
	initDB()

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/signup", enableCors(signupHandler))
	mux.HandleFunc("/login", enableCors(loginHandler))
	mux.HandleFunc("/users/", enableCors(authMiddleware(getUserNameHandler)))
	mux.HandleFunc("/messages/unread", enableCors(authMiddleware(getUnreadMessagesHandler)))

	// Protected routes
	mux.HandleFunc("/user/items", authMiddleware(getUserItemsHandler))
	mux.HandleFunc("/user/addresses", authMiddleware(getUserAddressesHandler))
	mux.HandleFunc("/addresses/delete", authMiddleware(deleteAddressHandler))
	mux.HandleFunc("/user/orders", authMiddleware(getUserOrdersHandler))
	mux.HandleFunc("/items/create", authMiddleware(createItemWithImagesHandler))
	mux.HandleFunc("/items/delete", authMiddleware(deleteItemHandler))
	mux.HandleFunc("/orders/update", authMiddleware(updateOrderStatusHandler))
	mux.HandleFunc("/orders/archive", authMiddleware(archiveOrderHandler))
	mux.HandleFunc("/cart/add", authMiddleware(addToCartHandler))
	mux.HandleFunc("/cart", authMiddleware(viewCartHandler))
	mux.HandleFunc("/cart/remove", authMiddleware(removeFromCartHandler))
	mux.HandleFunc("/checkout", authMiddleware(checkoutHandler))
	mux.HandleFunc("/user/current", authMiddleware(getCurrentUserHandler))
	mux.HandleFunc("/messages/seen", enableCors(authMiddleware(markMessagesAsSeenHandler)))
	mux.HandleFunc("/orders/", enableCors(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/messages") {
			switch r.Method {
			case http.MethodGet:
				getMessageHandler(w, r)
			case http.MethodPost:
				sendMessageHandler(w, r)
			case http.MethodOptions:
				w.WriteHeader(http.StatusOK)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})))
	// Update your endpoint handlers in main() function
	mux.HandleFunc("/notifications/unread", enableCors(authMiddleware(getUnreadNotificationsHandler)))
	mux.HandleFunc("/notifications/seen/", enableCors(authMiddleware(markNotificationAsSeenHandler)))
	mux.HandleFunc("/notifications/clear", enableCors(authMiddleware(clearNotificationsHandler)))

	// Public routes
	mux.HandleFunc("/items/search", enableCors(searchItemsHandler))
	mux.HandleFunc("/images", enableCors(serveImageHandler))

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Error creating uploads directory:", err)
	}

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
