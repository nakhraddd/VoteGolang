package app

import (
	"VoteGolang/internals/app/conf"
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/data"
	"VoteGolang/internals/deliveries/candidate_routes"
	"VoteGolang/internals/deliveries/general_news_routes"
	"VoteGolang/internals/deliveries/login_routes"
	"VoteGolang/internals/deliveries/petition_routes"
	candidate_repo "VoteGolang/internals/repositories/candidate_repository"
	general_news3 "VoteGolang/internals/repositories/general_news_repository"
	petition3 "VoteGolang/internals/repositories/petition_repository"
	"VoteGolang/internals/repositories/user_repository"
	"VoteGolang/internals/repositories/votes_repositories"
	"VoteGolang/internals/usecases/auth_usecase"
	"VoteGolang/internals/usecases/candidate_usecase"
	"VoteGolang/internals/usecases/general_news_usecase"
	"VoteGolang/internals/usecases/petittion_usecase"
	"VoteGolang/pkg/domain"
	"gorm.io/gorm"
	"log"
	"net/http"
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
	//
	//err = MigrateAllTables(db)
	//if err != nil {
	//	return nil, nil, nil, err
	//}
	app := &App{
		Config: config,
		DB:     db,
	}

	userRepo := user_repository.NewUserRepository(db)
	tokenManager := domain.NewJwtToken(config.JWTSecret)
	authUseCase := auth_usecase.NewAuthUseCase(userRepo, tokenManager)

	return app, authUseCase, tokenManager, nil
}

func (a *App) Run(authUseCase *auth_usecase.AuthUseCase, tokenManager domain.TokenManager) {
	log.Println("Starting server on port 8080...")

	mux := http.NewServeMux()
	//auth
	authHandler := login_routes.NewAuthHandler(authUseCase, tokenManager)
	login_routes.AuthorizationRoutes(mux, authHandler, tokenManager)

	_, ok := tokenManager.(domain.TokenManager)
	if !ok {
		log.Fatal("Token manager is not of type *auth.JwtTokenManager")
	}

	// Candidate
	candidateHandler := candidate_routes.NewCandidateHandler(
		candidate_usecase.NewCandidateUseCase(
			candidate_repo.NewCandidateRepository(a.DB),
			votes_repositories.NewVoteRepository(a.DB),
		),
		tokenManager.(*domain.JwtToken),
	)
	candidate_routes.RegisterCandidateRoutes(mux, candidateHandler, tokenManager)

	// General News
	generalNewsHandler := general_news_routes.NewGeneralNewsHandler(
		general_news_usecase.NewGeneralNewsUseCase(
			general_news3.NewGeneralNewsRepository(a.DB),
		),
		tokenManager.(*domain.JwtToken),
	)
	general_news_routes.RegisterGeneralNewsRoutes(mux, generalNewsHandler, tokenManager)

	//Petitions
	petitionsHandler := petition_routes.NewPetitionHandler(
		petittion_usecase.NewPetitionUseCase(
			petition3.NewPetitionRepository(a.DB),
			votes_repositories.NewPetitionVoteRepository(a.DB),
		),
		tokenManager.(*domain.JwtToken),
	)
	petition_routes.RegisterPetitionRoutes(mux, petitionsHandler, tokenManager)

	// fallback for unknown routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "route not found"}`))
	})

	//start
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func MigrateAllTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		&data.User{},
		&data.Candidate{},
		&data.GeneralNews{},
		&data.Petition{},
		&data.PetitionVote{},
		&data.Vote{},
	)
	if err != nil {
		return err
	}
	log.Println("Database tables migrated successfully")
	return nil
}
