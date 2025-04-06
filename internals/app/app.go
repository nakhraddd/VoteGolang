package app

import (
	"VoteGolang/internals/app/conf"
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/auth"
	"VoteGolang/internals/repositories"
	"VoteGolang/internals/usecases"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type App struct {
	Config *conf.Config
	DB     *gorm.DB
}

func NewApp() (*App, error) {
	config := conf.LoadConfig()
	db, err := connect.ConnectDB(config)
	if err != nil {
		return nil, err
	}

	app := &App{
		Config: config,
		DB:     db,
	}

	userRepo := repositories.NewUserRepository(db)
	tokenManager := auth.NewJwtToken(config.JWTSecret)
	authUseCase := usecases.NewAuthUseCase(userRepo, tokenManager)

	http.HandleFunc("/login", usecases.NewAuthHandler(authUseCase).Login)

	return app, nil
}

func (a *App) Run() {
	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
