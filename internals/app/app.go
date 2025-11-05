package app

import (
	"VoteGolang/conf"
	_ "VoteGolang/docs" // ← This is important for registering the generated Swagger docs
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/app/migrations"
	"VoteGolang/internals/blockchain"
	"VoteGolang/internals/controller/blockchain_routes"
	"VoteGolang/internals/controller/candidate_routes"
	"VoteGolang/internals/controller/login_routes"
	"VoteGolang/internals/controller/petition_routes"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/email"
	candidate_repo "VoteGolang/internals/infrastructure/repositories"
	"VoteGolang/internals/usecases/auth_usecase"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/usecases/petittion_usecase"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

type App struct {
	Config     *conf.Config
	DB         *gorm.DB
	Blockchain *blockchain.Blockchain
}

func NewApp() (*App, *auth_usecase.AuthUseCase, domain.TokenManager, *redis.Client, error) {
	config := conf.LoadConfig()
	db, err := connect.ConnectDB(config)

	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = migrations.MigrateAllTables(db)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bc := blockchain.NewBlockchain(2)

	app := &App{
		Config:     config,
		DB:         db,
		Blockchain: bc,
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
		log.Fatalln("Redis connection was refused:", err)
	}
	log.Println("Redis ping:", status)

	roleRepo := candidate_repo.NewRoleRepository(db)

	// создаем EmailVerifier
	emailVerifier := email.NewRedisEmailVerifier(rdb)
	authUseCase := auth_usecase.NewAuthUseCase(userRepo, roleRepo, tokenManager, emailVerifier)

	// start clean up of unverified users
	email.StartUnverifiedCleanupJob(userRepo)

	return app, authUseCase, tokenManager, rdb, nil
}

func (a *App) Run(authUseCase *auth_usecase.AuthUseCase, tokenManager domain.TokenManager, kafkaLogger *logging.KafkaLogger, rdb *redis.Client) {
	log.Println("Starting server on port 8080...")

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

	// create rbac repo once
	rbacRepo := candidate_repo.NewRBACRepository(a.DB)

	// Candidate
	candidateHandler := candidate_routes.NewCandidateHandler(
		candidate_usecase.NewCandidateUseCase(
			candidate_repo.NewCandidateRepository(a.DB),
			candidate_repo.NewVoteRepository(a.DB),
			a.Blockchain,
			rdb,
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	candidate_routes.RegisterCandidateRoutes(mux, candidateHandler, tokenManager, rbacRepo)

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

	// Blockchain
	blockchainHandler := blockchain_routes.NewBlockchainHandler(a.Blockchain)
	blockchain_routes.RegisterBlockchainRoutes(mux, blockchainHandler)

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
		log.Fatalf("Error starting server: %v", err)
	}

}
