package controllers

import (
	"genesis/initializers"
	"genesis/models"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AccountBody struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=255"`
}

type AccountResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func AccountCreate(c *gin.Context) {
	// Clean and Get response Body
	var req AccountBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Object",
		})
		return
	}

	// Hash Password and Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to Hash Password",
		})
		return
	}

	// Check and validate email uniqueness
	var existingAccount models.Account
	if err := initializers.DB.Where("email = ?", req.Email).First(&existingAccount).Error; err == nil {
		log.Printf("Account already exists for email: %s", req.Email)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email Already Exists",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("Error checking existing account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error checking account",
		})
		return
	}

	account := models.Account{
		Email:    req.Email,
		Password: string(hash),
	}

	if err := initializers.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error Creating Account",
		})
		return
	}
	resp := AccountResponse{
		ID:        account.ID,
		Email:     account.Email,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
	c.JSON(http.StatusCreated, gin.H{
		"account": resp,
	})
}

func AccountLogin(c *gin.Context) {
	// Get and Sanitize the input request
	var req AccountBody
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request Object")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request Object",
		})
		return
	}

	// Fetch user data and compare password hash

	var existingAccount models.Account
	if err := initializers.DB.Where("email = ?", req.Email).First(&existingAccount).Error; err != nil {
		log.Println("Account Does not exist")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid Login",
		})
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(existingAccount.Password), []byte(req.Password))
	if err != nil { // If err != nil then password incorrect
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Incorrect Login",
		})
		return
	}
	// Correct password
	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": existingAccount.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	//log.Printf("Token: %v", token)
	// Sign and get the complete encoded token as string using secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		log.Printf("Unable to sign jwt token with secret %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Something Unexpected Happened",
		})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		//"token": tokenString,
	})

}

func Validate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "I'm logged In",
	})
}
