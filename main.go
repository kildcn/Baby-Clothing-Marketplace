package main

import (
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

func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

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
}

type CartItem struct {
	ItemID  string    `json:"item_id"`
	AddedAt time.Time `json:"added_at"`
}

// Storage
var (
	items = make(map[string]Item)
	cart  = make(map[string][]CartItem) // user_id -> cart items
)

// Helper functions
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

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

	var req struct {
		UserID string `json:"user_id"`
		ItemID string `json:"item_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cartItem := CartItem{
		ItemID:  req.ItemID,
		AddedAt: time.Now(),
	}

	cart[req.UserID] = append(cart[req.UserID], cartItem)
	sendJSON(w, cart[req.UserID])
}

func viewCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	userCart := cart[userID]

	var cartItems []Item
	for _, cartItem := range userCart {
		if item, exists := items[cartItem.ItemID]; exists {
			cartItems = append(cartItems, item)
		}
	}

	sendJSON(w, cartItems)
}

func serveImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imagePath := r.URL.Query().Get("path")
	http.ServeFile(w, r, imagePath)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/create", enableCors(createItemWithImagesHandler))
	mux.HandleFunc("/items/search", enableCors(searchItemsHandler))
	mux.HandleFunc("/cart/add", enableCors(addToCartHandler))
	mux.HandleFunc("/cart", enableCors(viewCartHandler))
	mux.HandleFunc("/images", enableCors(serveImageHandler))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
