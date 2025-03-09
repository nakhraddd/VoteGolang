package main

import (
	"VoteGolang/database"
	"VoteGolang/handlers"
	"VoteGolang/middleware"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	database.GetDB()
	database.Migrate()

	r := gin.Default()
	r.POST("/login", handlers.LoginHandler)
	r.POST("/register", handlers.RegisterHandler)

	voteRoutes := r.Group("/vote", middleware.AuthMiddleware())
	{
		voteRoutes.GET("/general_news", handlers.GetGeneralNews)
		voteRoutes.GET("/petition", handlers.GetPetitions)
		voteRoutes.GET("/president", handlers.GetPresidentCandidates)
		voteRoutes.GET("/session_deputy", handlers.GetSessionDeputyCandidates)
		voteRoutes.GET("/deputy", handlers.GetDeputyCandidates)
	}

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}
