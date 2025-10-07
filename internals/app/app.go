package app

import (
	"VoteGolang/conf"
	_ "VoteGolang/docs" // ← This is important for registering the generated Swagger docs
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/app/logging"
	"VoteGolang/internals/app/migrations"
	"VoteGolang/internals/controller/candidate_routes"
	"VoteGolang/internals/controller/login_routes"
	"VoteGolang/internals/controller/petition_routes"
	"VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/email"
	candidate_repo "VoteGolang/internals/infrastructure/repositories/candidate_repository"
	petition3 "VoteGolang/internals/infrastructure/repositories/petition_repository"
	"VoteGolang/internals/infrastructure/repositories/user_repository"
	votes_repositories2 "VoteGolang/internals/infrastructure/repositories/votes_repositories"
	"VoteGolang/internals/usecases/auth_usecase"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/usecases/petittion_usecase"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

type App struct {
	Config *conf.Config
	DB     *gorm.DB
}

func NewApp() (*App, *auth_usecase.AuthUseCase, domain.TokenManager, error) {
	config := conf.LoadConfig()
	db, err := connect.ConnectDB(config)

	if err != nil {
		return nil, nil, nil, err
	}

	err = migrations.MigrateAllTables(db)
	if err != nil {
		return nil, nil, nil, err
	}

	app := &App{
		Config: config,
		DB:     db,
	}

	userRepo := user_repository.NewUserRepository(db)
	tokenManager := domain.NewJwtToken(config.JWTSecret)
	// создаем Redis клиент
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%v:%v", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		//Password: "", // если есть пароль - укажи
		DB: 0,
	})

	// создаем EmailVerifier
	emailVerifier := email.NewRedisEmailVerifier(rdb)
	authUseCase := auth_usecase.NewAuthUseCase(userRepo, tokenManager, emailVerifier)

	// start clean up of unverified users
	email.StartUnverifiedCleanupJob(userRepo)

	return app, authUseCase, tokenManager, nil
}

func (a *App) Run(authUseCase *auth_usecase.AuthUseCase, tokenManager domain.TokenManager, kafkaLogger *logging.KafkaLogger) {
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

	// Candidate
	candidateHandler := candidate_routes.NewCandidateHandler(
		candidate_usecase.NewCandidateUseCase(
			candidate_repo.NewCandidateRepository(a.DB),
			votes_repositories2.NewVoteRepository(a.DB),
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	candidate_routes.RegisterCandidateRoutes(mux, candidateHandler, tokenManager)

	//Petitions
	petitionsHandler := petition_routes.NewPetitionHandler(
		petittion_usecase.NewPetitionUseCase(
			petition3.NewPetitionRepository(a.DB),
			votes_repositories2.NewPetitionVoteRepository(a.DB),
		),
		tokenManager.(*domain.JwtToken),
		kafkaLogger,
	)
	petition_routes.RegisterPetitionRoutes(mux, petitionsHandler, tokenManager)

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
