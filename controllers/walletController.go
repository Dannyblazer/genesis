package controllers

import (
	"genesis/initializers"
	"genesis/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletBody struct {
	ID        uint      `json:"id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	AccountID uint      `json:"accountID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransferBody struct {
	Email  string `json:"email" binding:"required,email,max=255"`
	Amount int64  `json:"amount" binding:"required"`
}

// Add Transfer and Entries from SIMPLE BANK postgres file

func WalletDetail(c *gin.Context) {
	// Get user from jwt auth
	accountID, ok := c.Get("accountID")

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "No user associated in jwt",
		})
		return
	}

	var wallet models.Wallet

	if err := initializers.DB.Where("account_id = ?", accountID).First(&wallet).Error; err != nil {
		log.Printf("Failed to fetch wallet %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch wallet",
		})
		return
	}

	resp := WalletBody{
		ID:        wallet.ID,
		Balance:   float64(wallet.Balance),
		Currency:  wallet.Currency,
		AccountID: wallet.AccountID,
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet": resp,
	})

}

func WalletTransfer(c *gin.Context) {
	accountID, ok := c.Get("accountID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No Wallet Associated with your account"})
		return
	}

	var req TransferBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("Transfer request: accountID=%v, receiverEmail=%s, amount=%f", accountID, req.Email, req.Amount)

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {
		var receiverWallet models.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Joins("JOIN accounts ON accounts.id = wallets.account_id").
			Where("accounts.email = ?", req.Email).
			First(&receiverWallet).Error; err != nil {
			log.Printf("Receiver wallet query failed: %v", err)
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Receiver account not found"})
				return err
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return err
		}

		var senderWallet models.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Joins("JOIN accounts ON accounts.id = wallets.account_id").
			Where("accounts.id = ?", accountID).
			First(&senderWallet).Error; err != nil {
			log.Printf("Sender wallet query failed: %v", err)
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Sender account not found"})
				return err
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return err
		}

		if senderWallet.Balance < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
			return gorm.ErrInvalidTransaction
		}

		senderWallet.Balance -= req.Amount
		receiverWallet.Balance += req.Amount

		// Create Transfer object (history)
		transfer := models.Transfer{
			FromAccountID: uint64(senderWallet.AccountID),
			ToAccountID:   uint64(receiverWallet.AccountID),
			Amount:        req.Amount,
		}
		if err := initializers.DB.Create(&transfer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error Creating transfer record",
			})
			return err
		}

		if err := tx.Save(&senderWallet).Error; err != nil {
			log.Printf("Failed to update sender wallet: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source account"})
			return err
		}
		if err := tx.Save(&receiverWallet).Error; err != nil {
			log.Printf("Failed to update receiver wallet: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update destination account"})
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	var wallet models.Wallet
	if err := initializers.DB.Where("account_id = ?", accountID).First(&wallet).Error; err != nil {
		log.Printf("Failed to fetch wallet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer successful",
		"balance": wallet.Balance,
	})
}
