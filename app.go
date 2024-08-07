package main

import (
	"context"
	"fmt"
	"wb-assistance-logistic/config"
	"wb-assistance-logistic/logger"
	"wb-assistance-logistic/parser"
	"wb-assistance-logistic/sheets"
	"wb-assistance-logistic/telegramClient"
	"wb-assistance-logistic/timeTicker"
	"wb-assistance-logistic/utils"
)

type App struct {
	config         *config.Config
	telegramClient *telegramClient.Client
	parser         *parser.Parser
	googleSheet    *sheets.Sheet
	timeTicker     *timeTicker.TimeTicker

	isStarted bool
}

func NewApp(cfg *config.Config) (*App, error) {
	var err error
	app := new(App)
	app.config = cfg
	app.isStarted = false

	app.timeTicker = timeTicker.NewTimeTicker(cfg.Ticker.Frequency)
	app.timeTicker.SetCallback(app.tick)

	logger.Init("App.NewApp()", "Telegram client")
	//app.telegramClient = telegramClient.NewClient(int32(cfg.TelegramClient.Id), cfg.TelegramClient.Hash)
	app.telegramClient, err = telegramClient.NewClientByParameters(&telegramClient.ClientParameters{
		ApiId:   111,
		ApiHash: "sd",
	})
	//
	err = telegramClient.SetTelegramClientLogsVerboseLevel(telegramClient.LogsVerboseLevel(cfg.TelegramClient.LogLevel))
	if err != nil {
		return nil, logger.Error("App.NewApp()", "Error set <Telegram client> logs verbose:\n", err)
	}
	logger.LogLn("App.NewApp()", "Setup logs verbose level <Telegram client>: ", cfg.TelegramClient.LogLevel)

	logger.Init("App.NewApp()", "Auth Telegram client")
	err = app.telegramClient.Auth()
	if err != nil {
		logger.LogLn("App.NewApp()", "Error auth <Telegram client>:\n", err)
	}

	isAuth := <-app.telegramClient.AuthReady()
	if !isAuth {
		return nil, logger.Error("App.NewApp()", "Failed to auth <Telegram client>")
	}
	logger.InitSuccessfully("App.NewApp()", "Auth Telegram client")
	logger.InitSuccessfully("App.NewApp()", "Telegram client")

	logger.Init("App.NewApp()", "Parser")
	app.parser, err = parser.NewParser(cfg.Parser, app.telegramClient)
	if err != nil {
		return nil, logger.Error("App.NewApp()", "Error create <Parser>:\n", err)
	}
	logger.InitSuccessfully("App.NewApp()", "Parser")

	logger.Init("App.NewApp()", "Sheet")
	logger.Init("App.NewApp()", "Sheet service")
	googleService, err := CreateGoogleSheetsService(cfg.Sheets)
	if err != nil {
		return nil, logger.Error("App.NewApp()", "Error auth <Sheet> service:\n", err)
	}
	logger.InitSuccessfully("App.NewApp()", "Sheet service")

	app.googleSheet, err = sheets.NewSheetByService(cfg.Sheets.ID, googleService)
	if err != nil {
		return nil, logger.Error("App.NewApp()", "Error initialization <Sheet> service\n", err)
	}
	logger.InitSuccessfully("App.NewApp()", "Sheet")

	app.tick()

	return app, nil
}

func (app *App) tick() {
	logger.LogLn("App", "Parsing data...")

	data, err := app.parser.Parse()
	if err != nil {
		logger.Warning("App.NewApp()", "Error parse:\n", err)
		return
	}

	logger.LogLn("App", "Data: ", data)

	err = app.googleSheet.Update(app.config.Sheets.Name, app.config.Sheets.StartIndex, utils.ArrIntToInterface(data))
	if err != nil {
		logger.Warning("App", "Error append data to Sheet:\n", err)
		return
	}

	logger.LogLn("App", "Sheet was update")
}

func (app *App) Start() {
	logger.LogLn("App", "Starting app...")
	app.timeTicker.Start()
	app.isStarted = true
}

func (app *App) Stop() {
	logger.LogLn("App", "Stopping app...")
	app.timeTicker.Stop()
	app.isStarted = false
}

func CreateGoogleSheetsService(cfg *config.Sheets) (sheets.ServiceInterface, error) {
	if cfg.IsClientAuth {
		return AuthGoogleSheetsClient(cfg.Credentials.ClientToken, cfg.Credentials.Client)
	} else {
		return AuthGoogleSheetsService(cfg.Credentials.Service)
	}
}

func AuthGoogleSheetsClient(tokenCredentials, clientCredentials string) (*sheets.Client, error) {
	client, err := AuthGoogleSheetsClientByToken(tokenCredentials, clientCredentials)

	if err != nil {
		logger.Warning("App.AuthGoogleSheetsClient()", err)

		client, err = AuthGoogleSheetClientByCredentials(clientCredentials)
		if err != nil {
			return nil, logger.Error("App.AuthGoogleSheetsClient()", "Failed to login in Google Sheet serviceCredentials using client tokenCredentials file and credentials file:\n", err)
		}

		err = client.SaveJSONToken(tokenCredentials)
		if err != nil {
			logger.Warning("App.AuthGoogleSheetsClient()", "Failed to save the received tokenCredentials to a file\n", err)
		}
	}

	return client, nil
}

func AuthGoogleSheetsClientByToken(token string, client string) (*sheets.Client, error) {
	logger.LogLn("App.AuthGoogleSheetsClientByToken()", "Start auth Google Sheet internal by file token: "+token)

	service := sheets.NewClient()

	err := service.AuthByJSONTokenAutoRefresh(context.Background(), token, client, sheets.SHEETS_ALL_SCOPE)
	if err != nil {
		return nil, logger.Error("App.AuthGoogleSheetsClientByToken()", "Error auth Google Sheet internal by file token:\n", err)
	}

	return service, nil
}

func AuthGoogleSheetClientByCredentials(clientCredentials string) (*sheets.Client, error) {
	logger.LogLn("App.AuthGoogleSheetClientByCredentials()", "Start auth Google Sheet internal by file internal credentials: "+clientCredentials)

	service := sheets.NewClient()

	err := service.AuthByCredentials(
		context.Background(),
		clientCredentials,
		sheets.SHEETS_ALL_SCOPE,
		func(authURL string) string {
			logger.LogLn("App.AuthGoogleSheetClientByCredentials()", "Auth URL: "+authURL)
			var code string

			for {
				logger.Log("App.AuthGoogleSheetClientByCredentials()", "Auth code: ")
				_, err := fmt.Scanln(&code)
				if err != nil {
					logger.Warning("App.AuthGoogleSheetClientByCredentials()", "Error reading code\n", err)
					continue
				}
				break
			}

			return code
		})

	if err != nil {
		return nil, logger.Error("App.AuthGoogleSheetClientByCredentials()", "Error auth Google Sheet internal by file credentials:\n", err)
	}

	return service, nil
}

func AuthGoogleSheetsService(serviceCredentials string) (*sheets.Service, error) {
	logger.LogLn("App.AuthGoogleSheetsService()", "Start auth GoogleSheets internal by file internal credentials: "+serviceCredentials)

	service := sheets.NewService()
	err := service.Auth(context.Background(), serviceCredentials, sheets.SHEETS_ALL_SCOPE)
	if err != nil {
		return nil, logger.Error("App.AuthGoogleSheetsService()", "Failed to login in Google Sheet internal using internal credentials file:\n", err)
	}

	return service, nil
}
