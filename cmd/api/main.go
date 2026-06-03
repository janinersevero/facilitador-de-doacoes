package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	"facilitador-de-doacoes/internal/handler"
	"facilitador-de-doacoes/internal/middleware"
	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/internal/usecase"
	"github.com/gin-contrib/cors"

	"facilitador-de-doacoes/pkg/asaas"
	"facilitador-de-doacoes/pkg/auth0mgmt"
	"facilitador-de-doacoes/pkg/database"
	"facilitador-de-doacoes/pkg/queue"
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

	// NOTE: AutoMigrate adds payment_method and payment_id columns.
	// If upgrading from the previous schema, run manually:
	//   ALTER TABLE donations RENAME COLUMN pix_id TO payment_id;
	//   ALTER TABLE donations ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20) NOT NULL DEFAULT 'PIX';
	if err := db.AutoMigrate(&model.User{}, &model.Donation{}, &model.Institution{}, &model.Campaign{}, &model.Necessity{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	sandbox := os.Getenv("ASAAS_SANDBOX") == "true"
	asaasClient := asaas.NewClient(os.Getenv("ASAAS_API_KEY"), sandbox)
	supabaseClient := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"), os.Getenv("SUPABASE_BUCKET_NAME"))
	mgmtClient := auth0mgmt.NewClient(
		os.Getenv("AUTH0_DOMAIN"),
		os.Getenv("AUTH0_MGMT_CLIENT_ID"),
		os.Getenv("AUTH0_MGMT_CLIENT_SECRET"),
	)

	// Queue — optional: if REDIS_URL is not set, notifications are disabled
	var queueClient *asynq.Client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		redisOpt, err := asynq.ParseRedisURI(redisURL)
		if err != nil {
			log.Fatalf("invalid REDIS_URL: %v", err)
		}
		queueClient = asynq.NewClient(redisOpt)
		defer queueClient.Close()
		log.Println("queue: conectado ao Redis")
	} else {
		log.Println("queue: REDIS_URL não configurado, notificações desativadas")
	}

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
	donationUC := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, asaasClient, queueClient)
	donationHandler := handler.NewDonationHandler(donationUC)

	rankingUC := usecase.NewRankingUseCase(donationRepo)
	rankingHandler := handler.NewRankingHandler(rankingUC)

	transferUC := usecase.NewTransferUseCase(asaasClient)
	transferHandler := handler.NewTransferHandler(transferUC)

	necessityRepo := repository.NewNecessityRepository(db)
	necessityUC := usecase.NewNecessityUseCase(necessityRepo)
	necessityHandler := handler.NewNecessityHandler(necessityUC)

	// Worker — runs in a goroutine alongside the API
	if queueClient != nil {
		notifyURL := os.Getenv("CHATBOT_NOTIFY_URL")
		if notifyURL == "" {
			notifyURL = "http://localhost:3000/notify" // local fallback
		}

		notificationHandler := queue.NewNotificationHandler(userRepo, institutionRepo, campaignRepo, notifyURL)

		workerRedisOpt, _ := asynq.ParseRedisURI(redisURL)
		srv := asynq.NewServer(
			workerRedisOpt,
			asynq.Config{
				Concurrency: 5,
				RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
					return 15 * time.Minute
				},
			},
		)

		mux := asynq.NewServeMux()
		mux.HandleFunc(queue.TypeDonationPaid, notificationHandler.ProcessTask)

		go func() {
			if err := srv.Run(mux); err != nil {
				log.Fatalf("worker error: %v", err)
			}
		}()
		log.Println("worker: iniciado")
	}

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

	allowedOrigins := []string{"http://localhost:5174"}
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
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
	rankingHandler.RegisterRoutes(api)
	institutionHandler.RegisterRoutes(api,
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{authMiddleware, requireInstitution},
	)
	campaignHandler.RegisterRoutes(api, authMiddleware, requireInstitution)
	necessityHandler.RegisterRoutes(api, authMiddleware, requireInstitution)
	transferHandler.RegisterRoutes(api, authMiddleware, requireInstitution)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
