package main

import (
	"log"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

func main() {
	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		// Echo the message back to the client
		m.Broadcast(msg)
		log.Println(string(msg))

		// Check if the message is "96"
		if string(msg) == "96" {
			// Use xdotool to simulate pressing the space key
			err := exec.Command("xdotool", "key", "space").Run()
			if err != nil {
				log.Printf("Error executing xdotool: %v", err)
			}
		}
	})
	log.Println("starting server on 8089")
	r.Run("0.0.0.0:8089")
}
