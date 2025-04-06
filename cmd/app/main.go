package main

import (
	"log"
)

func main() {
	app, err := App.NewApp()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	app.RegisterRoutes()

	app.Run()
}
