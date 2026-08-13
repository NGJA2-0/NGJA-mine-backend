package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Env contains environment variables
type Env struct {
	MongoURI  string
	DBName    string
	JWTSecret string
	Port      string
}

// LoadEnv loads environment variables from .env file
func LoadEnv() *Env {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "mine_details_db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	return &Env{
		MongoURI:  mongoURI,
		DBName:    dbName,
		JWTSecret: jwtSecret,
		Port:      port,
	}
}
