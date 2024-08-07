package main

import (
	"os"
	"wb-assistance-logistic/config"
	"wb-assistance-logistic/logger"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.LogLn("Main()", "Recover error: \n", err)
			<-make(chan os.Signal, 1)
		}
	}()

	err := config.Init(".cfg")
	if err != nil {
		_ = logger.LogError("Main()", "Error initialization configuration:\n", err)
	}

	app, err := NewApp(config.Get())
	if err != nil {
		_ = logger.LogError("Main()", "Error initializing application:\n", err)
	} else if app != nil {
		app.Start()
	}

	<-make(chan os.Signal, 1)
}
