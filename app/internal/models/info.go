package models

type ReceiveHistory struct {
	FromUser string `json:"fromUser" db:"from"`
	Amount   int    `json:"amount" db:"amount"`
}

type SendHistory struct {
	ToUser string `json:"toUser" db:"to"`
	Amount int    `json:"amount" db:"amount"`
}

type CoinHistory struct {
	Received []ReceiveHistory `json:"received"`
	Sent     []SendHistory    `json:"sent"`
}

type ItemInfo struct {
	Type     string `json:"type" db:"type"`
	Quantity int    `json:"quantity" db:"quantity"`
}

type Info struct {
	Coins     int         `json:"coins"`
	History   CoinHistory `json:"coinHistory"`
	Inventory []ItemInfo  `json:"inventory"`
}
