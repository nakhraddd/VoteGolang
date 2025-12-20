package app

import (
	"VoteGolang/conf"
	_ "VoteGolang/docs" // ← This is important for registering the generated Swagger docs
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/app/migrations"
	"VoteGolang/internals/controller/blockchain_routes"
	"VoteGolang/internals/controller/candidate_routes"
	middleware "VoteGolang/internals/controller/http"
	"VoteGolang/internals/controller/login_routes"
	"VoteGolang/internals/controller/petition_routes"
	"VoteGolang/internals/controller/search_routes"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/email"
	"VoteGolang/internals/infrastructure/repositories"
	"VoteGolang/internals/infrastructure/search"
	"VoteGolang/internals/service" // <-- NEW IMPORT
	"VoteGolang/internals/usecases/auth_usecase"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/usecases/petition_usecase"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

type App struct {
	Config     *conf.Config
	DB         *gorm.DB
	Blockchain service.BlockchainService // <-- CHANGED
}

func NewApp(kafkaLogger *logging.KafkaLogger) (*App, *auth_usecase.AuthUseCase, domain.TokenManager, *redis.Client, *elasticsearch.Client, error) {
	kafkaLogger.Log("INFO", "Initializing application components...")

	config := conf.LoadConfig(kafkaLogger)
	kafkaLogger.Log("INFO", "Configuration loaded successfully")

	db, err := connect.ConnectDB(config, kafkaLogger)
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Database connection failed: %v", err))
		return nil, nil, nil, nil, nil, err
	}
	kafkaLogger.Log("INFO", "Connected to MySQL database successfully")

	err = migrations.MigrateAllTables(db)
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Database migration failed: %v", err))
		return nil, nil, nil, nil, nil, err
	}
	kafkaLogger.Log("INFO", "Database migrations applied successfully")

	esClient, err := connect.ConnectElasticsearch()
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Failed to connect to Elasticsearch: %v", err))
	} else {
		kafkaLogger.Log("INFO", "Connected to Elasticsearch successfully")
	}
	bc, err := service.NewBnbService(config.BNB) // <-- CHANGED
	if err != nil {
		kafkaLogger.Log("INFO", "Blockchain initialized")
	}
	app := &App{
		Config:     config,
		DB:         db,
		Blockchain: bc, // <-- CHANGED
	}

	userRepo := repositories.NewUserRepository(db)
	tokenManager := domain.NewJwtToken(config.JWTSecret)

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		DB:   0,
	})

	ctx := context.Background()
	status, err := rdb.Ping(ctx).Result()
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Redis connection failed: %v", err))
		log.Fatalln("Redis connection was refused:", err)
	} else {
		kafkaLogger.Log("INFO", fmt.Sprintf("Redis connected successfully: %s", status))
	}

	roleRepo := repositories.NewRoleRepository(db)

	// создаем EmailVerifier
	emailVerifier := email.NewRedisEmailVerifier(rdb)
	authUseCase := auth_usecase.NewAuthUseCase(userRepo, roleRepo, tokenManager, emailVerifier)

	// start clean up of unverified users
	email.StartUnverifiedCleanupJob(userRepo)
	kafkaLogger.Log("INFO", "Started background cleanup job for unverified users")

	kafkaLogger.Log("INFO", "All core services initialized successfully")
	return app, authUseCase, tokenManager, rdb, esClient, nil
}

func (a *App) Run(authUseCase *auth_usecase.AuthUseCase, tokenManager domain.TokenManager, kafkaLogger *logging.KafkaLogger, rdb *redis.Client, esClient *elasticsearch.Client) {
	kafkaLogger.Log("INFO", "Starting HTTP server on port 8080...")

	mux := http.NewServeMux()

	// Middleware to log every request
	logMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			kafkaLogger.Log("INFO", fmt.Sprintf("Accessed %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
			next.ServeHTTP(w, r)
		})
	}

	// Swagger UI route
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	//auth
	authHandler := login_routes.NewAuthHandler(authUseCase, tokenManager, kafkaLogger)
	login_routes.AuthorizationRoutes(mux, authHandler, tokenManager, kafkaLogger)
	kafkaLogger.Log("INFO", "Authentication routes registered")

	// create rbac repo once
	rbacRepo := repositories.NewRBACRepository(a.DB)

	// Candidate
	candidateHandler := candidate_routes.NewCandidateHandler(
		candidate_usecase.NewCandidateUseCase(
			repositories.NewCandidateRepository(a.DB),
			repositories.NewVoteRepository(a.DB),
			a.Blockchain,
			rdb,
			repositories.NewSearchRepository(esClient, "candidates"),
			kafkaLogger,
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	candidate_routes.RegisterCandidateRoutes(mux, candidateHandler, tokenManager, rbacRepo)
	kafkaLogger.Log("INFO", "Candidate routes registered")

	//Petitions
	petitionsHandler := petition_routes.NewPetitionHandler(
		petition_usecase.NewPetitionUseCase(
			repositories.NewPetitionRepository(a.DB),
			repositories.NewPetitionVoteRepository(a.DB),
			a.Blockchain,
			rdb,
			kafkaLogger,
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	petition_routes.RegisterPetitionRoutes(mux, petitionsHandler, tokenManager, rbacRepo)
	kafkaLogger.Log("INFO", "Petition routes registered")

	// Blockchain (Handler now shows service info)
	blockchainHandler := blockchain_routes.NewBlockchainHandler(a.Blockchain) // <-- PASSING THE INTERFACE
	blockchain_routes.RegisterBlockchainRoutes(mux, blockchainHandler)
	kafkaLogger.Log("INFO", "Blockchain routes registered")

	// Search
	searcher := search.NewElasticsearch("http://elasticsearch:9200")
	search_routes.SetupRoutes(mux, searcher)
	kafkaLogger.Log("INFO", "Search routes registered")

	// fallback for unknown routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		kafkaLogger.Log("WARN", fmt.Sprintf("Unknown route accessed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "route not found"}`))
	},
	)

	// Wrap mux with logging middleware
	handler := middleware.CORSMiddleware(logMiddleware(mux))
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Server failed to start: %v", err))
		log.Fatalf("Error starting server: %v", err)
	}

}
