package models

type Sale struct {
	ProductName string `json:"productName"`
	Quantity    int    `json:"quantity"`
	SaleDate    string `json:"saleDate"`
}
