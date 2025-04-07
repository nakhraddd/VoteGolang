package app

import (
	"VoteGolang/internals/app/conf"
	"VoteGolang/internals/app/connect"
	"VoteGolang/internals/app/start"
	"VoteGolang/internals/auth"
	"VoteGolang/internals/data"
	"VoteGolang/internals/repositories"
	"VoteGolang/internals/usecases"
	"VoteGolang/pkg/domain"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
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
	//
	//err = MigrateAllTables(db)
	//if err != nil {
	//	return nil, nil, nil, err
	//}
	app := &App{
		Config: config,
		DB:     db,
	}

	userRepo := repositories.NewUserRepository(db)
	tokenManager := auth.NewJwtToken(config.JWTSecret)
	authUseCase := usecases.NewAuthUseCase(userRepo, tokenManager)

	return app, authUseCase, tokenManager, nil
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		log.Printf("New request from %s, Method: %s, URL: %s", r.RemoteAddr, r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		duration := time.Since(startTime)
		log.Printf("Request from %s processed in %s", r.RemoteAddr, duration)
	})
}

func logLoginRegister(w http.ResponseWriter, r *http.Request, route string) {
	log.Printf("Attempting to access %s route, Method: %s, URL: %s", route, r.Method, r.URL.Path)
}

func (a *App) Run(authUseCase *usecases.AuthUseCase, tokenManager domain.TokenManager) {
	log.Println("Starting server on port 8080...")

	mux := http.NewServeMux()

	authHandler := usecases.NewAuthHandler(authUseCase, tokenManager)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		logLoginRegister(w, r, "/login")
		authHandler.Login(w, r)
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		logLoginRegister(w, r, "/register")
		authHandler.Register(w, r)
	})
	mux.Handle("/protected", start.JWTMiddleware(tokenManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You have access to the protected route"))
	})))

	mux.Handle("/", logRequest(mux))

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
