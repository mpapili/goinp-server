package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	semaphore = make(chan struct{}, maxGoRoutines)
)

const (
	deadZoneThreshold = 0.4
	maxGoRoutines     = 8
)

func init() {}

func main() {
	display, exists := os.LookupEnv("DISPLAY")
	if !exists || display == "" {
		fmt.Println("The DISPLAY variable is not set or is nil. (Default mobox sets it to :0)")
		os.Exit(1)
	} else {
		fmt.Println("The DISPLAY variable is set to:", display)
	}

	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		// Process the message
		processMessage(string(msg))
	})

	log.Println("starting server on 8089")

	r.Run("0.0.0.0:8089")
}

func processMessage(message string) {
	log.Println(message)
	switch message {
	case "96 down":
		holdKey("space")
	case "96 up":
		releaseKey("space")
	case "97 down":
		holdKey("R")
	case "97 up":
		releaseKey("R")
	case "105 down":
		holdMouse("1")
	case "105 up":
		releaseMouse("1")
	case "104 down":
		holdMouse("3")
	case "104 up":
		releaseMouse("3")
	}
	if strings.Contains(message, "Joystick Move") {

		// Parse the X and Y values from the message for "D"
		log.Printf("joystick has moved! here is the message %s", message)
		splits := strings.Split(message, ",")
		if len(splits) == 2 {
			xPart := strings.TrimSpace(strings.Split(splits[0], "X:")[1])
			xValue, err := strconv.ParseFloat(xPart, 64)
			if err != nil {
				log.Printf("error parsing joystick X value for %s: %v", xPart, err)
				return
			}
			yPart := strings.TrimSpace(strings.Split(splits[1], "Y:")[1])
			yValue, err := strconv.ParseFloat(yPart, 64)
			if err != nil {
				log.Printf("error parsing joystick Y value for %s : %v", yPart, err)
				return
			}
			log.Println(xValue, yValue)
			// move left or right?
			switch {
			case xValue > deadZoneThreshold:
				holdKey("D")
			case xValue <= deadZoneThreshold:
				releaseKey("D")
			}
			switch {
			case xValue < -1*deadZoneThreshold:
				holdKey("A")
			case xValue >= -1*deadZoneThreshold:
				releaseKey("A")
			}
			// move up or down?
			switch {
			case yValue > deadZoneThreshold:
				holdKey("S")
			case yValue <= deadZoneThreshold:
				releaseKey("S")
			}
			switch {
			case yValue < -1*deadZoneThreshold:
				holdKey("W")
			case yValue >= -1*deadZoneThreshold:
				releaseKey("W")
			}
		}
	}
}

func runXdoTool(keyAction string, key string) {
	go func(keyAction string, key string) {
		semaphore <- struct{}{}
		out, err := exec.Command("xdotool", keyAction, key).CombinedOutput()
		if err != nil {
			log.Printf("error performing %s %s : %v : %v", keyAction, key, err, out)
		}
		<-semaphore
	}(keyAction, key)
}

func holdKey(key string) {
	runXdoTool("keydown", key)
}

func releaseKey(key string) {
	runXdoTool("keyup", key)
}

func holdMouse(key string) {
	runXdoTool("mousedown", key)
}

func releaseMouse(key string) {
	runXdoTool("mouseup", key)
}
