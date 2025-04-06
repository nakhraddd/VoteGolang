package main

import (
	"VoteGolang/internals/app/connections"
	"VoteGolang/internals/deliveries/middleware"
	"VoteGolang/internals/repositories"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	connections.GetDB()
	connections.Migrate()

	r := gin.Default()
	r.POST("/login", repositories.LoginHandler)
	r.POST("/register", repositories.RegisterHandler)

	voteRoutes := r.Group("/vote", middleware.AuthMiddleware())
	{
		voteRoutes.GET("/general_news", repositories.GetGeneralNews)

		voteRoutes.GET("/petition", repositories.GetPetitions)
		voteRoutes.POST("/petitions/:title", repositories.VotePetition)

		voteRoutes.GET("/president", repositories.GetPresidentCandidates)
		voteRoutes.POST("/presidents/:name", repositories.VoteForPresident)

		voteRoutes.GET("/session_deputy", repositories.GetSessionDeputyCandidates)
		voteRoutes.POST("/session_deputies/:name", repositories.VoteForSessionDeputy)

		voteRoutes.GET("/deputy", repositories.GetDeputyCandidates)
		voteRoutes.POST("/deputy/:name", repositories.VoteForDeputy)

	}

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}
