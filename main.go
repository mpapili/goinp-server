package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	lastSignalTime sync.Map        // To track the last time signals were received for different actions
	keyHeld        map[string]bool // Tracks whether a key is currently being held down
	keyHeldMutex   sync.Mutex      // Mutex for safe access to keyHeld
)

func init() {
	keyHeld = make(map[string]bool)
}

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
			case xValue > 0.125:
				holdKey("D")
			case xValue <= 0.125:
				releaseKey("D")
			}
			switch {
			case xValue < -0.125:
				holdKey("A")
			case xValue >= -0.1:
				releaseKey("A")
			}
			// move up or down?
			switch {
			case yValue > 0.125:
				holdKey("S")
			case yValue <= 0.125:
				releaseKey("S")
			}
			switch {
			case yValue < -0.125:
				holdKey("W")
			case yValue >= -0.125:
				releaseKey("W")
			}
		}
	}
}

func holdKey(key string) {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	out, err := exec.Command("xdotool", "keydown", key).CombinedOutput()
	if err != nil {
		log.Printf("error holding %s key : %v : %s", key, err, out)
	}
	if !keyHeld[key] {
		keyHeld[key] = true
	}
}

func releaseKey(key string) {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	out, err := exec.Command("xdotool", "keyup", key).CombinedOutput()
	if err != nil {
		log.Printf("error releasing %s key : %v : %s", key, err, out)
	}
	if _, ok := keyHeld[key]; ok {
		delete(keyHeld, key)
	}
}

func holdMouse(key string) {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	out, err := exec.Command("xdotool", "mousedown", key).CombinedOutput()
	if err != nil {
		log.Printf("error releasing %s key : %v : %s", key, err, out)
	}
	if _, ok := keyHeld[key]; ok {
		delete(keyHeld, key)
	}
}

func releaseMouse(key string) {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	out, err := exec.Command("xdotool", "mouseup", key).CombinedOutput()
	if err != nil {
		log.Printf("error releasing %s key : %v : %s", key, err, out)
	}
	if _, ok := keyHeld[key]; ok {
		delete(keyHeld, key)
	}
}
