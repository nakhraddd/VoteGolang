package main

import (
	"VoteGolang/internals/app"
	"VoteGolang/internals/app/start"
	"log"
)

func main() {
	appInstance, authUseCase, tokenManager, err := app.NewApp()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	start.RegisterRoutes(authUseCase, tokenManager)

	appInstance.Run()
}
