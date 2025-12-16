package src

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger

func Initlogger() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed to open logger file: %v", err)
	}

	multi := io.MultiWriter(os.Stdout, file)
	Logger = log.New(multi, "chacalc_log", log.Ldate|log.Ltime|log.Lshortfile)
}
