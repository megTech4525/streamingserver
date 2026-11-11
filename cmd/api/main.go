package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"main.go/internal/auth"
	"main.go/internal/database"
	"main.go/internal/stream"
	"main.go/internal/user"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userHandler := user.NewHandler(user.NewService(user.NewRepository(db)))
	streamHandler := stream.NewHandler(stream.NewService(stream.NewRepository(db)))

	r := gin.Default()

	//Public Routes

	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
	r.GET("/streams", streamHandler.ListLive)

	// connecting valid tokens

	authed := r.Group("/", auth.Middleware())
	authed.POST("/streams", streamHandler.Create)

	//Needs the shared secret (media server only)

	r.POST("/internal/streams/validate", auth.InternalOnly, streamHandler.Validate)

	log.Fatal(r.Run(":8081"))

}
