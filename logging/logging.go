package logging

import (
	"log"
	"os"
)

var (
	warningLog *log.Logger
	infoLog    *log.Logger
	errorLog   *log.Logger
)

func Init() {
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Error() (log *log.Logger) {
	if errorLog == nil {
		logNotInitMessage()
	}

	return errorLog
}

func Info() (log *log.Logger) {
	if infoLog == nil {
		logNotInitMessage()
	}

	return infoLog
}

func Warn() (log *log.Logger) {
	if warningLog == nil {
		logNotInitMessage()
	}

	return warningLog
}

func logNotInitMessage() {
	log.Println("Logger must be initialized with call to logging.Init() at application startup")
}
