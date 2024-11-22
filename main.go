package main

import (
	"context"
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
)

// Type definitions
type Size string

const (
	XS Size = "XS"
	S  Size = "S"
	M  Size = "M"
	L  Size = "L"
	XL Size = "XL"
)

type Category string

const (
	Tops        Category = "tops"
	Bottoms     Category = "bottoms"
	Outerwear   Category = "outerwear"
	Footwear    Category = "footwear"
	Accessories Category = "accessories"
)

type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Size        Size      `json:"size"`
	Category    Category  `json:"category"`
	Images      []string  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	Quantity    int       `json:"quantity"`
	SellerID    string    `json:"seller_id"`
	SellerName  string    `json:"seller_name"`
}

type CartItem struct {
	ItemID  string    `json:"item_id"`
	AddedAt time.Time `json:"added_at"`
}

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Items     []Item    `json:"items"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Address   string    `json:"address"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Storage
var (
	items  = make(map[string]Item)
	cart   = make(map[string][]CartItem)
	orders = make(map[string]Order)
	users  = make(map[string]User)
)

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// Constants
const (
	maxFileSize = 10 << 20 // 10MB
	maxImages   = 3
	uploadDir   = "./uploads"
)

// Handlers
func createItemWithImagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get userID from context
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "At least one image required", http.StatusBadRequest)
		return
	}
	if len(files) > maxImages {
		http.Error(w, "Maximum 3 images allowed", http.StatusBadRequest)
		return
	}

	var imagePaths []string
	for _, fileHeader := range files {
		imagePath, err := saveImage(fileHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		imagePaths = append(imagePaths, imagePath)
	}

	item.ID = generateID()
	item.CreatedAt = time.Now()
	item.Status = "available"
	item.Images = imagePaths
	item.Quantity = 1
	item.SellerID = userID

	// Get seller name from users map
	if seller, exists := users[userID]; exists {
		item.SellerName = seller.Name
	}

	items[item.ID] = item
	sendJSON(w, item)
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

	filename := filepath.Join(uploadDir, generateID()+filepath.Ext(fileHeader.Filename))

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

func searchItemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	category := Category(r.URL.Query().Get("category"))
	size := Size(r.URL.Query().Get("size"))
	minPrice := parseFloat(r.URL.Query().Get("min_price"))
	maxPrice := parseFloat(r.URL.Query().Get("max_price"))

	var results []Item
	for _, item := range items {
		if matchesSearch(item, query, category, size, minPrice, maxPrice) {
			results = append(results, item)
		}
	}

	sendJSON(w, results)
}

func matchesSearch(item Item, query string, category Category, size Size, minPrice, maxPrice float64) bool {
	if query != "" {
		title := strings.ToLower(item.Title)
		desc := strings.ToLower(item.Description)
		query = strings.ToLower(query)
		words := strings.Fields(query)

		matched := false
		for _, word := range words {
			if strings.Contains(title, word) || strings.Contains(desc, word) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if category != "" && item.Category != category {
		return false
	}

	if size != "" && item.Size != size {
		return false
	}

	if minPrice > 0 && item.Price < minPrice {
		return false
	}

	if maxPrice > 0 && item.Price > maxPrice {
		return false
	}

	return true
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if item exists and has available quantity
	item, exists := items[req.ItemID]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Count how many times this item is in the user's cart
	itemCount := 0
	for _, cartItem := range cart[userID] {
		if cartItem.ItemID == req.ItemID {
			itemCount++
		}
	}

	if itemCount >= item.Quantity {
		http.Error(w, "Item out of stock", http.StatusBadRequest)
		return
	}

	cartItem := CartItem{
		ItemID:  req.ItemID,
		AddedAt: time.Now(),
	}

	cart[userID] = append(cart[userID], cartItem)
	sendJSON(w, cart[userID])
}

func viewCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userCart := cart[userID]

	var cartItems []Item
	for _, cartItem := range userCart {
		if item, exists := items[cartItem.ItemID]; exists {
			cartItems = append(cartItems, item)
		}
	}

	sendJSON(w, cartItems)
}

func removeFromCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userCart := cart[userID]
	for i, item := range userCart {
		if item.ItemID == req.ItemID {
			cart[userID] = append(userCart[:i], userCart[i+1:]...)
			break
		}
	}

	sendJSON(w, cart[userID])
}

func serveImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imagePath := r.URL.Query().Get("path")
	http.ServeFile(w, r, imagePath)
}

func getUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("userID").(string)
	return userID, ok
}

func generateID() string {
	return fmt.Sprintf("item-%d", len(items)+1)
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func enableCors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		h(w, r)
	}
}

func authMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return enableCors(func(w http.ResponseWriter, r *http.Request) {
		// Implement this middleware to handle authentication
		h(w, r)
	})
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	// Implement this handler to handle user signup
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Implement this handler to handle user login
}

func main() {
	// Initialize quantities
	for id, item := range items {
		if item.Quantity == 0 {
			item.Quantity = 1
			items[id] = item
		}
	}

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/signup", enableCors(signupHandler))
	mux.HandleFunc("/login", enableCors(loginHandler))

	// Protected routes with auth middleware
	mux.HandleFunc("/items/create", authMiddleware(createItemWithImagesHandler))
	mux.HandleFunc("/cart/add", authMiddleware(addToCartHandler))
	mux.HandleFunc("/cart", authMiddleware(viewCartHandler))
	mux.HandleFunc("/cart/remove", authMiddleware(removeFromCartHandler))

	// Public routes
	mux.HandleFunc("/items/search", enableCors(searchItemsHandler))
	mux.HandleFunc("/images", enableCors(serveImageHandler))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
