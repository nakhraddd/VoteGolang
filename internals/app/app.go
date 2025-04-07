package app

import (
	"VoteGolang/internals/app/conf"
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/auth"
	"VoteGolang/internals/repositories"
	"VoteGolang/internals/usecases"
	"VoteGolang/pkg/domain"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type App struct {
	Config *conf.Config
	DB     *gorm.DB
}

func NewApp() (*App, *usecases.AuthUseCase, domain.TokenManager, error) {
	config := conf.LoadConfig()
	db, err := connect.ConnectDB(config)
	if err != nil {
		return nil, nil, nil, err
	}

	app := &App{
		Config: config,
		DB:     db,
	}

	userRepo := repositories.NewUserRepository(db)
	tokenManager := auth.NewJwtToken(config.JWTSecret)
	authUseCase := usecases.NewAuthUseCase(userRepo, tokenManager)

	return app, authUseCase, tokenManager, nil
}

func (a *App) Run() {
	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
