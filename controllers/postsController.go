package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"genesis/initializers"
	"genesis/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequestPost represents the expected JSON payload for creating a post.
type RequestPostBody struct {
	Title string `json:"title" binding:"required,max=255"`
	Body  string `json:"body"  binding:"required,max=65535"`
}

// ResponsePost represents the response structure for a post.
type ResponsePost struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	AccountID uint      `json:"accountID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PostsCreate handles POST requests to create a new post.
func PostsCreate(c *gin.Context) {
	accountID, exists := c.Get("accountID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User is unauthenticated"})
		return
	}

	// Parse and validate request body
	var req RequestPostBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	fmt.Printf("accountID: %v, Type: %T\n", accountID, accountID)
	var account models.Account
	if err := initializers.DB.Select("id").First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Create post
	post := models.Post{
		Title:     req.Title,
		Body:      req.Body,
		AccountID: account.ID,
	}

	if err := initializers.DB.Create(&post).Error; err != nil {
		log.Printf("Failed to create post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create post",
		})
		return
	}

	// Prepare response
	resp := ResponsePost{
		ID:        post.ID,
		Title:     post.Title,
		Body:      post.Body,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"post": resp,
	})
}

// PostGet handles GET requests to retrieve a post by ID.
func PostGet(c *gin.Context) {
	// Get and validate post ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid post ID",
		})
		return
	}

	// Fetch post
	var post models.Post
	if err := initializers.DB.Select("id, title, body, created_at, updated_at").
		First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
			return
		}
		log.Printf("Failed to fetch post %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch post",
		})
		return
	}

	// Prepare response
	resp := ResponsePost{
		ID:        post.ID,
		Title:     post.Title,
		Body:      post.Body,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"post": resp,
	})
}

func PostList(c *gin.Context) {
	// Get auth user ID
	accountID, exists := c.Get("accountID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User is unauthenticated"})
		return
	}
	// removed the account check because it's already done in the auth middleware
	// var account models.Account
	// if err := initializers.DB.Select("id").First(&account, accountID).Error; err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	// 	return
	// }

	// Get Posts
	var posts []models.Post

	if err := initializers.DB.Where("account_id = ?", accountID).Find(&posts).Error; err != nil {
		log.Printf("Failed to fetch posts %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch post",
		})
	}

	resps := make([]ResponsePost, len(posts))
	for i, post := range posts {
		resps[i] = ResponsePost{
			ID:        post.ID,
			Title:     post.Title,
			Body:      post.Body,
			AccountID: post.AccountID,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": resps,
	})
}

func PostUpdate(c *gin.Context) {
	//Get post Id
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request param"})
		return
	}
	var req RequestPostBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
	}
	accountID, ok := c.Get("accountID")
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "No acccount associated with jwt"})
		return
	}
	var post models.Post
	if err := initializers.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	fmt.Println("Account Id is: ", post.Title)
	if post.AccountID != accountID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Action"})
		return
	}

	post = models.Post{
		Title:     req.Title,
		Body:      req.Body,
		AccountID: post.AccountID,
	}
	// Or post.title = newTitle -> db.save(&post)
	if err := initializers.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update post"})
	}
	resp := ResponsePost{
		ID:        post.ID,
		Title:     post.Title,
		Body:      post.Body,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
	c.JSON(http.StatusOK, gin.H{"post": resp})

}

func PostDelete(c *gin.Context) {
	// Get post ID param
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid post ID",
		})
		return
	}
	accountID, exists := c.Get("accountID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User is unauthenticated"})
		return
	}

	// Fetch post by ID in DB
	var post models.Post

	if err := initializers.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Post not found",
		})
		return
	}
	if post.AccountID != accountID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Action"})
		return
	}
	if err := initializers.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to Delete post",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post Deleted",
	})
}
