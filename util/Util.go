package util

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Yellow = "\033[0;33m"
	Blue   = "\033[0;34m"
	None   = "\033[0m"
)

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func AppInput() (*string, *string, *bool, *float64) {
	listenPort := flag.String("l", "", "Port to listen on")
	sendAddress := flag.String("s", "", "Address to send to")
	startAsListener := flag.Bool("i", false, "Start as initiator")
	pingLost := flag.Float64("p", 0, "Ping lost probability (0, 1)")
	flag.Parse()

	if *listenPort == "" || *sendAddress == "" {
		fmt.Println("Usage: go run main.go -l <port> -s <address:port> (-i) <initiator>")
		os.Exit(1)
	}

	return listenPort, sendAddress, startAsListener, pingLost
}

func Log(message string, color string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf(message, args...)
	fmt.Printf(color+"[%s] %s"+None+"\n", timestamp, formattedMessage)
}

func LogInfo(message string, args ...interface{}) {
	Log(message, Blue, args...)
}

func LogError(message string, args ...interface{}) {
	Log(message, Red, args...)
}

func LogWarn(message string, args ...interface{}) {
	Log(message, Yellow, args...)
}

func LogSuccess(message string, args ...interface{}) {
	Log(message, Green, args...)
}
