package main

import (
	"genesis/initializers"
	"genesis/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDB()
}

func main() {
	initializers.DB.AutoMigrate(&models.Post{})
	initializers.DB.AutoMigrate(&models.Account{})
	initializers.DB.AutoMigrate(&models.Wallet{})
	initializers.DB.AutoMigrate(&models.Transfer{})
	initializers.DB.AutoMigrate(&models.Entry{})
}
