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
	Password string `json:"-"` // "-" means this won't be sent in JSON responses
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

// Constants
const (
	maxFileSize = 10 << 20 // 10MB
	maxImages   = 3
	uploadDir   = "./uploads"
)

var db *sql.DB

func initDB() {
	// Update with your username from whoami command
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

	filename := filepath.Join(uploadDir, fmt.Sprintf("%s%s",
		uuid.New().String(), filepath.Ext(fileHeader.Filename)))

	dst, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		return "", err
	}

	return filename, nil
}

// Handlers
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

	// Set default quantity to 1
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

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert item with explicit quantity = 1
	var itemID string
	err = tx.QueryRow(`
        INSERT INTO items (title, description, price, size, category, seller_id, quantity, status)
        VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, 1), 'available'::item_status_enum)
        RETURNING id, quantity`, // Also return quantity to verify
		item.Title, item.Description, item.Price, item.Size, item.Category, userID, item.Quantity).Scan(&itemID, &item.Quantity)

	log.Printf("After insert, quantity is: %d", item.Quantity)

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
	log.Printf("Verification query shows quantity: %d", checkQuantity)

	if checkQuantity != 1 {
		// Update it if somehow it's not 1
		_, err = tx.Exec(`UPDATE items SET quantity = 1 WHERE id = $1`, itemID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fixing quantity: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Save images and create records
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

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch complete item data
	var createdItem Item
	var images []sql.NullString
	err = db.QueryRow(`
		SELECT i.id, i.title, i.description, i.price, i.size,
			   i.category, i.status, i.quantity, i.seller_id,
			   u.name as seller_name, i.created_at,
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

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	var req struct {
		ItemID string `json:"item_id"`
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

	// Check item availability
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

	// Allow items to be added only if quantity > 0
	if quantity <= 0 {
		http.Error(w, "Item out of stock", http.StatusBadRequest)
		return
	}

	// Check how many of this item are in the user's cart
	var cartCount int
	err = tx.QueryRow(`
      SELECT COUNT(*) FROM cart_items
      WHERE user_id = $1 AND item_id = $2`,
		userID, req.ItemID).Scan(&cartCount)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make sure we don't exceed available quantity
	if cartCount >= quantity {
		http.Error(w, "Cannot add more of this item - quantity limit reached", http.StatusBadRequest)
		return
	}

	// Add to cart
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

	// Return updated cart
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

// User dashboard handlers

func getUserItemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())

	rows, err := db.Query(`
      SELECT i.id, i.title, i.description, i.price, i.size, i.category,
             i.status, i.quantity, i.seller_id, u.name as seller_name,
             i.created_at, array_agg(im.image_path) as images
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

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := getUserIDFromContext(r.Context())
	itemID := r.URL.Query().Get("id")

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// First get the image paths
	var imagePaths []string
	rows, err := tx.Query("SELECT image_path FROM item_images WHERE item_id = $1", itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imagePaths = append(imagePaths, path)
	}

	// Delete item (cascade will handle item_images)
	result, err := tx.Exec(`
      DELETE FROM items
      WHERE id = $1 AND seller_id = $2`,
		itemID, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Item not found or unauthorized", http.StatusNotFound)
		return
	}

	// Delete the physical image files
	for _, path := range imagePaths {
		os.Remove(path)
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r.Context())

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get items in cart with their quantities
	rows, err := tx.Query(`
      SELECT item_id, COUNT(*) as count_in_cart
      FROM cart_items
      WHERE user_id = $1
      GROUP BY item_id`,
		userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// For each item, decrease quantity
	for rows.Next() {
		var itemID string
		var countInCart int
		if err := rows.Scan(&itemID, &countInCart); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update item quantity
		result, err := tx.Exec(`
          UPDATE items
          SET quantity = quantity - $1
          WHERE id = $2 AND quantity >= $1`,
			countInCart, itemID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if update was successful
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			http.Error(w, "Some items are no longer available in requested quantity", http.StatusBadRequest)
			return
		}

		// Update status to 'sold' if quantity reaches 0
		_, err = tx.Exec(`
          UPDATE items
          SET status = CASE
              WHEN quantity = 0 THEN 'sold'::item_status_enum
              ELSE status
              END
          WHERE id = $1`,
			itemID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Clear user's cart
	_, err = tx.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"message": "Checkout successful"})
}

func serveImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imagePath := r.URL.Query().Get("path")
	http.ServeFile(w, r, imagePath)
}

func enableCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all responses
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h(w, r)
	}
}

func authMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check authorization
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

func main() {
	initDB()

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/signup", enableCors(signupHandler))
	mux.HandleFunc("/login", enableCors(loginHandler))

	// Protected routes - note the order of middleware
	mux.HandleFunc("/user/items", authMiddleware(getUserItemsHandler)) // User dashboard routes
	mux.HandleFunc("/items/delete", authMiddleware(deleteItemHandler))
	mux.HandleFunc("/items/create", authMiddleware(createItemWithImagesHandler))
	mux.HandleFunc("/cart/add", authMiddleware(addToCartHandler))
	mux.HandleFunc("/cart", authMiddleware(viewCartHandler))
	mux.HandleFunc("/cart/remove", authMiddleware(removeFromCartHandler))
	mux.HandleFunc("/checkout", authMiddleware(checkoutHandler))

	// Public routes
	mux.HandleFunc("/items/search", enableCors(searchItemsHandler))
	mux.HandleFunc("/images", enableCors(serveImageHandler))

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Error creating uploads directory:", err)
	}

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
