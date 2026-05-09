package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"facilitador-de-doacoes/internal/handler"
	"facilitador-de-doacoes/internal/middleware"
	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/internal/usecase"
	"github.com/gin-contrib/cors"

	"facilitador-de-doacoes/pkg/abacatepay"
	"facilitador-de-doacoes/pkg/auth0mgmt"
	"facilitador-de-doacoes/pkg/database"
	"facilitador-de-doacoes/pkg/supabase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Donation{}, &model.Institution{}, &model.Campaign{}, &model.Necessity{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	abacateClient := abacatepay.NewClient(os.Getenv("ABACATEPAY_API_KEY"))
	supabaseClient := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"), os.Getenv("SUPABASE_BUCKET_NAME"))
	mgmtClient := auth0mgmt.NewClient(
		os.Getenv("AUTH0_DOMAIN"),
		os.Getenv("AUTH0_MGMT_CLIENT_ID"),
		os.Getenv("AUTH0_MGMT_CLIENT_SECRET"),
	)

	userRepo := repository.NewUserRepository(db)
	userUC := usecase.NewUserUseCase(userRepo, supabaseClient, mgmtClient)
	userHandler := handler.NewUserHandler(userUC)

	institutionRepo := repository.NewInstitutionRepository(db)
	institutionUC := usecase.NewInstitutionUseCase(institutionRepo, mgmtClient)
	institutionHandler := handler.NewInstitutionHandler(institutionUC)

	campaignRepo := repository.NewCampaignRepository(db)
	campaignUC := usecase.NewCampaignUseCase(campaignRepo)
	campaignHandler := handler.NewCampaignHandler(campaignUC)

	donationRepo := repository.NewDonationRepository(db)
	donationUC := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, abacateClient)
	donationHandler := handler.NewDonationHandler(donationUC)

	necessityRepo := repository.NewNecessityRepository(db)
	necessityUC := usecase.NewNecessityUseCase(necessityRepo)
	necessityHandler := handler.NewNecessityHandler(necessityUC)

	jwtValidator, err := middleware.NewAuth0Validator(
		os.Getenv("AUTH0_DOMAIN"),
		os.Getenv("AUTH0_AUDIENCE"),
	)
	if err != nil {
		log.Fatalf("failed to create auth0 validator: %v", err)
	}

	authMiddleware := middleware.AuthMiddleware(jwtValidator)
	requireUser := middleware.RequireUser(userUC)
	requireInstitution := middleware.RequireInstitution(institutionUC)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_URL"), "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api/v1")
	userHandler.RegisterRoutes(api,
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{authMiddleware, requireUser},
	)
	donationHandler.RegisterRoutes(api)
	institutionHandler.RegisterRoutes(api,
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{authMiddleware, requireInstitution},
	)
	campaignHandler.RegisterRoutes(api, authMiddleware, requireInstitution)
	necessityHandler.RegisterRoutes(api, authMiddleware, requireInstitution)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
