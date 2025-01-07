package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/models"
)

func HandleShopsSalesSummary(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT s.shop_name, SUM(sa.quantity * p.price) AS total_sales
		FROM shops s
		LEFT JOIN sales sa ON s.shop_id = sa.shop_id
		LEFT JOIN products p ON sa.product_id = p.product_id
		GROUP BY s.shop_id
		ORDER BY total_sales DESC
	`
	rows, err := db.Conn.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to fetch shop sales summary", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var salesSummaries []models.ShopSales
	for rows.Next() {
		var shopSales models.ShopSales
		if err := rows.Scan(&shopSales.ShopName, &shopSales.TotalSales); err != nil {
			http.Error(w, "Failed to parse shop sales summary", http.StatusInternalServerError)
			return
		}
		salesSummaries = append(salesSummaries, shopSales)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(salesSummaries)
}
