package models

type Shop struct {
	ShopID      string `json:"shopId"`
	ShopName    string `json:"shopName"`
	Location    string `json:"location"`
	Description string `json:"description"`
}
