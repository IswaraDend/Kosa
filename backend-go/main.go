package main

import (
	"log"

	"backend-go/config"
	"backend-go/database"
	"backend-go/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	config.Load()
	database.Connect()

	r := gin.Default()
	routes.RegisterRoutes(r)

	r.Run(":8080")
}
