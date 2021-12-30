package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func NewUserHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func main() {
	var u User
	u.Username = "test"
	u.Password = "test"
	fmt.Println(u.Register())
	fmt.Println(u.NewRepo("test"))
	// r := gin.Default()
	// v1 := r.Group("/v1")
	// {
	// 	v1.POST(NewUser, NewUserHandler)
	// }
	// log.Fatal(r.Run(":8080"))
}
