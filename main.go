package main

import (
	"VoteGolang/database"
	"VoteGolang/handlers"
	"VoteGolang/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	database.GetDB()
	database.Migrate()

	r := gin.Default()
	r.POST("/vote/users_data", handlers.CreateUser) // Create user
	r.GET("/vote/users_data/:id", handlers.GetUser) // Get user by ID
	// Apply the authentication middleware to all vote routes
	voteRoutes := r.Group("/vote", middleware.AuthMiddleware())
	{

		voteRoutes.GET("/general_news", handlers.GetGeneralNews)
		voteRoutes.GET("/petition", handlers.GetPetitions)
		voteRoutes.GET("/president", handlers.GetPresidentCandidates)
		voteRoutes.GET("/session_deputy", handlers.GetSessionDeputyCandidates)
		voteRoutes.GET("/deputy", handlers.GetDeputyCandidates)
	}

	r.Run(":8080")
}
