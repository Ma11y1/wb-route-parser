package logger

import (
	"errors"
	"fmt"
	"log"
)

func Init(place string, target string) {
	log.Print("[INIT]   ", place+" >> ", "Start initialization <"+target+">")
}

func InitSuccessfully(place string, target string) {
	log.Print("[INIT]   ", place+" >> ", "Initialization <"+target+"> successfully")
}

func Log(place string, msg ...interface{}) {
	log.Print("[INFO]", place+" >> ", fmt.Sprint(msg...))
}

func LogLn(place string, msg ...interface{}) {
	log.Println("[INFO]", place+" >> ", fmt.Sprint(msg...))
}

func Warning(place string, msg ...interface{}) {
	log.Println("[WARNING] ", place+" >> ", fmt.Sprint(msg...))
}

func Fatal(place string, msg ...interface{}) {
	log.Fatalln("[FATAL] ", place+" >> ", fmt.Sprint(msg...))
}

func LogError(place string, msg ...interface{}) error {
	message := fmt.Sprintln("[ERROR] ", place+" >> ", fmt.Sprint(msg...))
	log.Println(message)
	return errors.New(message)
}

func Error(place string, msg ...interface{}) error {
	return errors.New(fmt.Sprintln("[ERROR] ", place+" >> ", fmt.Sprint(msg...)))
}
