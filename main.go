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
		voteRoutes.POST("/petitions/:id/vote", handlers.VotePetition)
		voteRoutes.GET("/petitions/:id/votes", handlers.GetPetitionVotes)

		voteRoutes.GET("/president", handlers.GetPresidentCandidates)
		voteRoutes.POST("/presidents/:id/vote", handlers.VoteForPresident)
		voteRoutes.GET("/presidents/:id/votes", handlers.GetPresidentVoteCount)

		voteRoutes.GET("/session_deputy", handlers.GetSessionDeputyCandidates)
		voteRoutes.POST("/session-deputies/:id/vote", handlers.VoteForSessionDeputy)
		voteRoutes.GET("/session-deputies/:id/votes", handlers.GetSessionDeputyVoteCount)

		voteRoutes.GET("/deputy", handlers.GetDeputyCandidates)
		voteRoutes.POST("/deputies/vote", handlers.VoteForDeputy)
		voteRoutes.GET("/deputies/:id/votes", handlers.GetDeputyVoteCount)

	}

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}
