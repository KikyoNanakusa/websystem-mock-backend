package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"websystem-backend/db"
	"websystem-backend/models"
)

func HandleSalesData(w http.ResponseWriter, r *http.Request) {
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
	rows, err := db.Conn.Query(context.Background(), query, shopID)
	if err != nil {
		http.Error(w, "Failed to fetch sales data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var sale models.Sale
		if err := rows.Scan(&sale.ProductName, &sale.Quantity, &sale.SaleDate); err != nil {
			http.Error(w, "Failed to parse sales data", http.StatusInternalServerError)
			return
		}
		sales = append(sales, sale)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sales)
}
