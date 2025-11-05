package app

import (
	"VoteGolang/conf"
	_ "VoteGolang/docs" // ← This is important for registering the generated Swagger docs
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/app/migrations"
	"VoteGolang/internals/controller/blockchain_routes"
	"VoteGolang/internals/controller/candidate_routes"
	"VoteGolang/internals/controller/login_routes"
	"VoteGolang/internals/controller/petition_routes"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/email"
	candidate_repo "VoteGolang/internals/infrastructure/repositories"
	"VoteGolang/internals/infrastructure/search"
	"VoteGolang/internals/service" // <-- NEW IMPORT
	"VoteGolang/internals/usecases/auth_usecase"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/usecases/petittion_usecase"
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

	bc := blockchain.NewBlockchain(2, kafkaLogger)
	kafkaLogger.Log("INFO", "Blockchain initialized")

	app := &App{
		Config:     config,
		DB:         db,
		Blockchain: bc, // <-- CHANGED
	}

	userRepo := candidate_repo.NewUserRepository(db)
	tokenManager := domain.NewJwtToken(config.JWTSecret)
	// создаем Redis клиент
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%v:%v", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		//Password: "", // если есть пароль
		DB: 0,
	})

	ctx := context.Background()
	status, err := rdb.Ping(ctx).Result()
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Redis connection failed: %v", err))
		log.Fatalln("Redis connection was refused:", err)
	} else {
		kafkaLogger.Log("INFO", fmt.Sprintf("Redis connected successfully: %s", status))
	}

	roleRepo := candidate_repo.NewRoleRepository(db)

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
	rbacRepo := candidate_repo.NewRBACRepository(a.DB)

	searchRepo := search.NewSearchRepository(esClient, "candidates")

	// Candidate
	candidateHandler := candidate_routes.NewCandidateHandler(
		candidate_usecase.NewCandidateUseCase(
			candidate_repo.NewCandidateRepository(a.DB),
			candidate_repo.NewVoteRepository(a.DB),
			a.Blockchain,
			rdb,
			searchRepo,
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	candidate_routes.RegisterCandidateRoutes(mux, candidateHandler, tokenManager, rbacRepo)
	kafkaLogger.Log("INFO", "Candidate routes registered")

	//Petitions
	petitionsHandler := petition_routes.NewPetitionHandler(
		petittion_usecase.NewPetitionUseCase(
			candidate_repo.NewPetitionRepository(a.DB),
			candidate_repo.NewPetitionVoteRepository(a.DB),
			a.Blockchain,
			rdb,
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	petition_routes.RegisterPetitionRoutes(mux, petitionsHandler, tokenManager, rbacRepo)
	kafkaLogger.Log("INFO", "Petition routes registered")

	// Blockchain
	blockchainHandler := blockchain_routes.NewBlockchainHandler(a.Blockchain)
	blockchain_routes.RegisterBlockchainRoutes(mux, blockchainHandler)
	kafkaLogger.Log("INFO", "Blockchain routes registered")

	// fallback for unknown routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		kafkaLogger.Log("WARN", fmt.Sprintf("Unknown route accessed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "route not found"}`))
	})

	// Wrap mux with logging middleware
	err := http.ListenAndServe(":8080", logMiddleware(mux))
	if err != nil {
		kafkaLogger.Log("ERROR", fmt.Sprintf("Server failed to start: %v", err))
		log.Fatalf("Error starting server: %v", err)
	}

}
