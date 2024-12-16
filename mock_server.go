package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time" 

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var db *pgx.Conn

// CORS ミドルウェア
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

// データベース接続
func connectDB() {
	var err error

	// .env ファイルから環境変数を読み込み
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// DATABASE_URL を取得
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in .env")
	}

	// PostgreSQL に接続 (ステートメントキャッシュを無効化)
	db, err = pgx.Connect(context.Background(), databaseURL+"?statement_cache_mode=describe")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	log.Println("Connected to the database.")
}

// ハンドラー: 店舗データ取得
func handleShopData(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shopId")
	if shopID == "" {
		http.Error(w, "shopId is required", http.StatusBadRequest)
		return
	}

	var shop struct {
		ShopID      string `json:"shopId"`
		ShopName    string `json:"shopName"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}

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

// ハンドラー: 全店舗データ取得
func handleStoresData(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT shop_id, shop_name, location FROM shops")
	if err != nil {
		log.Printf("Error fetching stores data: %v", err)
		http.Error(w, "Failed to fetch stores", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stores []map[string]interface{}
	for rows.Next() {
		var shop struct {
			ShopID   string `json:"shopId"`
			ShopName string `json:"shopName"`
			Location string `json:"location"`
		}

		if err := rows.Scan(&shop.ShopID, &shop.ShopName, &shop.Location); err != nil {
			log.Printf("Error scanning store data: %v", err)
			http.Error(w, "Failed to parse store data", http.StatusInternalServerError)
			return
		}

		stores = append(stores, map[string]interface{}{
			"shopId":   shop.ShopID,
			"shopName": shop.ShopName,
			"location": shop.Location,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stores)
}

// ハンドラー: 販売データ取得
func handleSalesData(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shopId")
	if shopID == "" {
		http.Error(w, "shopId is required", http.StatusBadRequest)
		return
	}

	// 販売データ取得クエリ
	query := `
		SELECT p.product_name, s.quantity, s.sale_date
		FROM sales s
		JOIN products p ON s.product_id = p.product_id
		WHERE s.shop_id = $1
		ORDER BY s.sale_date
	`

	// クエリ実行
	rows, err := db.Query(context.Background(), query, shopID) // `db` に変更
	if err != nil {
		log.Printf("Error fetching sales data for shopId %s: %v", shopID, err)
		http.Error(w, "Failed to fetch sales data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 結果を構造体に変換
	var sales []map[string]interface{}
	for rows.Next() {
		var sale struct {
			ProductName string    `json:"productName"`
			Quantity    int       `json:"quantity"`
			SaleDate    time.Time `json:"saleDate"` // `time.Time` 型
		}

		if err := rows.Scan(&sale.ProductName, &sale.Quantity, &sale.SaleDate); err != nil {
			log.Printf("Error scanning sales data for shopId %s: %v", shopID, err)
			http.Error(w, "Failed to parse sales data", http.StatusInternalServerError)
			return
		}

		// `time.Time` 型を ISO 8601 形式の文字列に変換
		sales = append(sales, map[string]interface{}{
			"productName": sale.ProductName,
			"quantity":    sale.Quantity,
			"saleDate":    sale.SaleDate.Format(time.RFC3339), // ISO 8601 形式
		})
	}

	// JSON にエンコードしてクライアントに送信
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sales); err != nil {
		log.Printf("Error encoding sales data for shopId %s: %v", shopID, err)
		http.Error(w, "Failed to encode sales data", http.StatusInternalServerError)
	}
}

// ハンドラー: 全店舗の売上データ取得
func handleAllShopsData(w http.ResponseWriter, r *http.Request) {
	// 全店舗データを取得するクエリ
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

	var allShops []map[string]interface{}
	for rows.Next() {
		var shop struct {
			ShopName   string  `json:"shopName"`
			TotalSales float64 `json:"totalSales"`
		}

		if err := rows.Scan(&shop.ShopName, &shop.TotalSales); err != nil {
			log.Printf("Error scanning all shops data: %v", err)
			http.Error(w, "Failed to parse all shops data", http.StatusInternalServerError)
			return
		}

		allShops = append(allShops, map[string]interface{}{
			"shopName":   shop.ShopName,
			"totalSales": shop.TotalSales,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(allShops); err != nil {
		log.Printf("Error encoding all shops data: %v", err)
		http.Error(w, "Failed to encode all shops data", http.StatusInternalServerError)
	}
}

func main() {
	// データベース接続
	connectDB()
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// ルーティング
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shop", handleShopData)
	mux.HandleFunc("/api/shops", handleStoresData)
	mux.HandleFunc("/api/sales", handleSalesData)
  mux.HandleFunc("/api/all-shops", handleAllShopsData)

	// CORS ミドルウェア適用
	handler := corsMiddleware(mux)

	// サーバー起動
	port := "4000"
	log.Printf("API server running on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
