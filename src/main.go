package main

import (
	"context"

	"example.com/main/src/initializer"
	"example.com/main/src/services"
	"github.com/gin-gonic/gin"
)

func init() {
	err := initializer.Db()
	if err != nil {
		panic("cant initialize db")
	}
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies([]string{"192.168.1.2"})

	defer func() {
		if err := initializer.Mongos.Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	auth := r.Group("/auth")
	{
		auth.GET("", services.Get)
		auth.POST("/signup", services.Signup)
		auth.POST("/signin", services.Singin)
		//need auth
		auth.POST("/signout", services.Signout)
		auth.POST("/refresh", services.RefreshToken)
	}

	r.Run()
}
