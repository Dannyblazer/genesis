package models

import (
	"gorm.io/gorm"
)

// Account represents the accounts table (equivalent to Wallet)
type Wallet struct {
	gorm.Model
	AccountID         uint
	Balance           int64      `gorm:"not null"`
	Currency          string     `gorm:"type:varchar;not null"`
	Entries           []Entry    `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
	SentTransfers     []Transfer `gorm:"foreignKey:FromAccountID;constraint:OnDelete:CASCADE"`
	ReceivedTransfers []Transfer `gorm:"foreignKey:ToAccountID;constraint:OnDelete:CASCADE"`
}

// Entry represents the entries table
type Entry struct {
	gorm.Model
	AccountID uint64 `gorm:"not null;index"`
	Amount    int64  `gorm:"not null;comment:can be negative or positive"`
}

// Transfer represents the transfers table
type Transfer struct {
	gorm.Model
	FromAccountID uint64 `gorm:"not null;index:idx_from_account,priority:1;index:idx_from_to_account,priority:1"`
	ToAccountID   uint64 `gorm:"not null;index:idx_to_account,priority:1;index:idx_from_to_account,priority:2"`
	Amount        int64  `gorm:"not null;comment:must be positive"`
}
