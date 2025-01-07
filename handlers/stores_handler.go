package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/models"
)

func HandleStoresData(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Conn.Query(context.Background(), "SELECT shop_id, shop_name, location FROM shops")
	if err != nil {
		http.Error(w, "Failed to fetch stores", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stores []models.Shop
	for rows.Next() {
		var shop models.Shop
		if err := rows.Scan(&shop.ShopID, &shop.ShopName, &shop.Location); err != nil {
			http.Error(w, "Failed to parse store data", http.StatusInternalServerError)
			return
		}
		stores = append(stores, shop)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stores)
}
