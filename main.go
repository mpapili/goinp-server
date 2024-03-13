package main

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

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

	go monitorSignalsAndAct()

	r.Run("0.0.0.0:8089")
}

func processMessage(message string) {
	log.Println(message)
	switch message {
	case "96 down":
		holdKey("space")
	case "96 up":
		releaseKey("space")
	}
	if strings.Contains(message, "Joystick Move") {
		// Parse the X and Y values from the message for "D"
		splits := strings.Split(message, ",")
		if len(splits) == 2 {
			xPart := strings.TrimSpace(strings.Split(splits[0], "X:")[1])
			xValue, err := strconv.ParseFloat(xPart, 64)
			if err == nil && xValue > 0.1 {
				lastSignalTime.Store("D", time.Now())
				ensureKeyHeld("D")
			} else if err == nil && xValue < 0.1 {
				lastSignalTime.Store("A", time.Now())
				ensureKeyHeld("A")
			}
		}
	}
}

func monitorSignalsAndAct() {
	for {
		now := time.Now()
		lastSignalTime.Range(func(key, value interface{}) bool {
			action := key.(string)
			lastTime := value.(time.Time)
			if now.Sub(lastTime) > 300*time.Millisecond {
				if keyReleased(action) {
					exec.Command("xdotool", "keyup", action).Run()
				}
			}
			return true // continue ranging
		})
		time.Sleep(20 * time.Millisecond) // Adjust as needed
	}
}

func ensureKeyHeld(key string) {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	if !keyHeld[key] {
		exec.Command("xdotool", "keydown", key).Run()
		keyHeld[key] = true
	}
}

func holdKey(key string) {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	out, err := exec.Command("xdotool", "keydown", key).CombinedOutput()
	if err != nil {
		log.Printf("error releasing %s key : %v : %s", key, err, out)
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

func keyReleased(key string) bool {
	keyHeldMutex.Lock()
	defer keyHeldMutex.Unlock()
	if keyHeld[key] {
		delete(keyHeld, key)
		return true
	}
	return false
}
