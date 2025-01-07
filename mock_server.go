package main

import (
	"log"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/handlers"
	"websystem-backend/middlewares"
)

func main() {
	// データベース接続
	db.ConnectDB()
	defer db.CloseDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/shop", handlers.HandleShopData)
	mux.HandleFunc("/api/shops", handlers.HandleStoresData)
	mux.HandleFunc("/api/sales", handlers.HandleSalesData)
	mux.HandleFunc("/api/shop-sales-summary", handlers.HandleShopsSalesSummary)
	mux.HandleFunc("/api/login", handlers.HandleLogin)

	handler := middlewares.CORS(mux)

	port := "4000"
	log.Printf("API server running on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
