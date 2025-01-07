package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"websystem-backend/models"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var db *pgx.Conn

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func connectDB() {
	var err error
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in .env")
	}

	db, err = pgx.Connect(context.Background(), databaseURL+"?statement_cache_mode=describe")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	log.Println("Connected to the database.")
}

func handleShopData(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shopId")
	if shopID == "" {
		http.Error(w, "shopId is required", http.StatusBadRequest)
		return
	}

	var shop models.Shop
	err := db.QueryRow(
		context.Background(),
		"SELECT shop_id, shop_name, location, '詳細情報は未実装' AS description FROM shops WHERE shop_id = $1",
		shopID,
	).Scan(&shop.ShopID, &shop.ShopName, &shop.Location, &shop.Description)

	if err != nil {
		log.Printf("Error fetching shop data for shopId %s: %v", shopID, err)
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}

func handleStoresData(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT shop_id, shop_name, location FROM shops")
	if err != nil {
		log.Printf("Error fetching stores data: %v", err)
		http.Error(w, "Failed to fetch stores", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stores []models.Shop
	for rows.Next() {
		var shop models.Shop
		if err := rows.Scan(&shop.ShopID, &shop.ShopName, &shop.Location); err != nil {
			log.Printf("Error scanning store data: %v", err)
			http.Error(w, "Failed to parse store data", http.StatusInternalServerError)
			return
		}
		stores = append(stores, shop)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stores)
}

func handleSalesData(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shopId")
	if shopID == "" {
		http.Error(w, "shopId is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT p.product_name, s.quantity, s.sale_date
		FROM sales s
		JOIN products p ON s.product_id = p.product_id
		WHERE s.shop_id = $1
		ORDER BY s.sale_date
	`
	rows, err := db.Query(context.Background(), query, shopID)
	if err != nil {
		log.Printf("Error fetching sales data for shopId %s: %v", shopID, err)
		http.Error(w, "Failed to fetch sales data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var sale models.Sale
		if err := rows.Scan(&sale.ProductName, &sale.Quantity, &sale.SaleDate); err != nil {
			log.Printf("Error scanning sales data for shopId %s: %v", shopID, err)
			http.Error(w, "Failed to parse sales data", http.StatusInternalServerError)
			return
		}
		sales = append(sales, sale)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sales)
}

func handleAllShopsData(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT s.shop_name, SUM(sa.quantity * p.price) AS total_sales
		FROM shops s
		LEFT JOIN sales sa ON s.shop_id = sa.shop_id
		LEFT JOIN products p ON sa.product_id = p.product_id
		GROUP BY s.shop_id
		ORDER BY total_sales DESC
	`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		log.Printf("Error fetching all shops data: %v", err)
		http.Error(w, "Failed to fetch all shops data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allShops []models.ShopSales
	for rows.Next() {
		var shopSales models.ShopSales
		if err := rows.Scan(&shopSales.ShopName, &shopSales.TotalSales); err != nil {
			log.Printf("Error scanning all shops data: %v", err)
			http.Error(w, "Failed to parse all shops data", http.StatusInternalServerError)
			return
		}
		allShops = append(allShops, shopSales)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allShops)
}

func main() {
	connectDB()
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/shop", handleShopData)
	mux.HandleFunc("/api/shops", handleStoresData)
	mux.HandleFunc("/api/sales", handleSalesData)
	mux.HandleFunc("/api/all-shops", handleAllShopsData)

	handler := corsMiddleware(mux)

	port := "4000"
	log.Printf("API server running on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
