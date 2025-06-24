package controllers

import "time"

type WalletBody struct {
	ID        uint      `json:"id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	AccountID uint      `json:"accountID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Add Transfer and Entries from SIMPLE BANK postgres file
