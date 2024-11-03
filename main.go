package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

var (
	semaphore  = make(chan struct{}, maxGoRoutines)
	mouseUp    = false
	mouseDown  = false
	mouseLeft  = false
	mouseRight = false
	resetMouse = false
)

const (
	deadZoneThreshold = 0.4
	maxGoRoutines     = 4 
	mousePollingRate  = 50 // milliseconds
)

func main() {
	display, exists := os.LookupEnv("DISPLAY")
	if !exists || display == "" {
		panic("the $DISPLAY variable is not set (default mobox sets to :0")
	}
	log.Printf("The DISPLAY variable is set to: %s", display)
	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(_ *melody.Session, msg []byte) {
		// Process the message
		processMessage(string(msg))
	})

	// Start mouse-handler thread
	go func() {
		counter := 0
		for {
			time.Sleep(mousePollingRate * time.Millisecond)
			// vertical
			if mouseUp {
				moveMouse("0", "-30")
			} else if mouseDown {
				moveMouse("0", "30")
			}
			// horizontal
			if mouseLeft {
				moveMouse("-30", "0")
			} else if mouseRight {
				moveMouse("30", "0")
			}
			counter += 1
			if counter > 20 && resetMouse == true {
				counter = 0
				resetMouseCenter()
			}
		}
	}()

	log.Println("starting server on 8089")
	err := r.Run("0.0.0.0:8089")
	if err != nil {
		panic(fmt.Errorf("error occurred starting websocket server : %w", err))
	}
}

func processMessage(message string) {
	log.Println(message)
	// Clickable buttons
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
	case "102 down":
		holdKey("F")
	case "102 up":
		releaseKey("F")
	case "107 down": // Right-Trigger-Click / R3
		if resetMouse == true {
			resetMouse = false
		} else {
			resetMouse = true
		}
	}

	// Joystick Handling Logic
	if strings.Contains(message, "Left Joystick Move") {
		log.Printf("left joystick has moved! here is the message %s", message)
		splits := strings.Split(message, ",")
		if len(splits) == 2 {
			// Handle Left Joystick Movement
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
			// Test a hard-coded WASD handling
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
	} else if strings.Contains(message, "Right Joystick Move") {
		// Handle Right Joystick Movement
		log.Printf("right joystick has moved! here is the message %s", message)
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
			// move left or right?
			switch {
			case xValue > deadZoneThreshold:
				mouseRight = true
				mouseLeft = false
			case xValue < (-1 * deadZoneThreshold):
				mouseLeft = true
				mouseRight = false
			default:
				mouseLeft = false
				mouseRight = false
			}
			// move up or down?
			switch {
			case yValue < -1*deadZoneThreshold:
				mouseUp = true
				mouseDown = false
			case yValue >= deadZoneThreshold:
				mouseDown = true
				mouseUp = false
			default:
				mouseDown = false
				mouseUp = false
			}
		}
	}
}

func runXdoTool(keyAction string, key string) {
	go func(keyAction string, key string) {
		semaphore <- struct{}{}
		out, err := exec.Command("xdotool", keyAction, key).CombinedOutput()
		if err != nil {
			log.Printf("error performing %s %s : %v : %s", keyAction, key, err, out)
		}
		<-semaphore
	}(keyAction, key)
}

// TODO - nearly identical; find a better way to split args.. []... this is being lazy
func runXdoToolMouse(x string, y string) {
	go func(x string, y string) {
		semaphore <- struct{}{}
		out, err := exec.Command("xdotool", "mousemove_relative", "--", x, y).CombinedOutput()
		if err != nil {
			log.Printf("error performing mouse move %s %s : %v : %s", x, y, err, out)
		}
		<-semaphore
	}(x, y)
}

// TODO - nearly identical; find a better way to split args.. []... this is being lazy
func resetMouseCenter() {
	go func() {
		semaphore <- struct{}{}
		out, err := exec.Command("xdotool", "mousemove", "150", "150").CombinedOutput()
		if err != nil {
			log.Printf("error reseting mouse position : %v : %s", err, out)
		}
		<-semaphore
	}()

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

func moveMouse(x string, y string) {
	runXdoToolMouse(x, y)
}
