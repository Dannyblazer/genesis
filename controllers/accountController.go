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

type EmailChange struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type AccountResponse struct {
	ID        uint           `json:"id"`
	Email     string         `json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	Posts     []ResponsePost `json:"posts"`
	UpdatedAt time.Time      `json:"updated_at"`
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

func AccountDetail(c *gin.Context) {
	// Get user from jwt auth
	accountID, ok := c.Get("accountID")

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user associated in jwt"})
		return
	}

	var account models.Account
	if err := initializers.DB.Preload("Posts").First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Associated user not found"})
		return
	}

	var responsePosts []ResponsePost
	for _, post := range account.Posts {
		responsePosts = append(responsePosts, ResponsePost{
			ID:        post.ID,
			Title:     post.Title,
			Body:      post.Body,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		})
	}
	resp := AccountResponse{
		ID:        account.ID,
		Email:     account.Email,
		CreatedAt: account.CreatedAt,
		Posts:     responsePosts,
		UpdatedAt: account.UpdatedAt,
	}
	c.JSON(http.StatusOK, gin.H{
		"account": resp,
	})

}

func AccountUpdate(c *gin.Context) {
	// Get auth user account
	accountID, err := c.Get("accountID")
	if !err {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not associated with jwt"})
		return
	}
	var req EmailChange
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}
	var existingAccount models.Account
	if err := initializers.DB.Where("email = ?", req.Email).First(&existingAccount).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already used"})
		return
	}

	var account models.Account
	if err := initializers.DB.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch account"})
		return
	}

	if err := initializers.DB.Model(&account).Update("email", req.Email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to Update Account"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account Updated"})
}

func AccountDelete(c *gin.Context) {
	// Get Account ID jwt
	accountID, ok := c.Get("accountID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No account associated with jwt"})
		return
	}
	if err := initializers.DB.Delete(&models.Account{}, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete account"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account Deleted"})
}
