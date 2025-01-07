package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/models"
)

func HandleShopData(w http.ResponseWriter, r *http.Request) {
	shopID := r.URL.Query().Get("shopId")
	if shopID == "" {
		http.Error(w, "shopId is required", http.StatusBadRequest)
		return
	}

	var shop models.Shop
	err := db.Conn.QueryRow(
		context.Background(),
		"SELECT shop_id, shop_name, location, '詳細情報は未実装' AS description FROM shops WHERE shop_id = $1",
		shopID,
	).Scan(&shop.ShopID, &shop.ShopName, &shop.Location, &shop.Description)

	if err != nil {
		http.Error(w, "Shop not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shop)
}
