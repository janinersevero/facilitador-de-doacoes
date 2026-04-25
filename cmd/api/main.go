package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"facilitador-de-doacoes/internal/handler"
	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/pkg/abacatepay"
	"facilitador-de-doacoes/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Donation{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	abacateClient := abacatepay.NewClient(os.Getenv("ABACATEPAY_API_KEY"))

	userRepo := repository.NewUserRepository(db)
	userUC := usecase.NewUserUseCase(userRepo)
	userHandler := handler.NewUserHandler(userUC)

	donationRepo := repository.NewDonationRepository(db)
	donationUC := usecase.NewDonationUseCase(donationRepo, userRepo, abacateClient)
	donationHandler := handler.NewDonationHandler(donationUC)

	r := gin.Default()

	api := r.Group("/api/v1")
	userHandler.RegisterRoutes(api)
	donationHandler.RegisterRoutes(api)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
