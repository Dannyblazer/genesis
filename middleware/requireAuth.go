package middleware

import (
	"genesis/initializers"
	"genesis/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	// Get cookie off req
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		log.Println("No token found in cookie")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	log.Println(tokenString)
	// Decode/Validate token within cookie

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		log.Fatal(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check Expiration
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// Find Account attached to token sub
		var existingAccount models.Account
		initializers.DB.First(&existingAccount, claims["sub"])
		if existingAccount.ID == 0 {
			log.Println("Account Does not exist")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// Attach to req
		c.Set("accountID", existingAccount.ID)
		// Continue
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
