package config

import (
	"encoding/json"
	"os"
	"wb-assistance-logistic/logger"
)

type TimeTicker struct {
	Frequency int `json:"frequency"`
}

type TelegramClient struct {
	Id       int    `json:"id"`
	Hash     string `json:"hash"`
	LogLevel int    `json:"log_level"`
}

type Sheets struct {
	Credentials struct {
		Client      string `json:"client"`
		ClientToken string `json:"client_token"`
		Service     string `json:"service"`
	} `json:"credentials"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	StartIndex   string `json:"start_index"`
	IsClientAuth bool   `json:"client_auth"`
}

type Parser struct {
	ChatUsername       string   `json:"chat_username"`
	CommandRequestData string   `json:"command_request_data"`
	RequestTimeSleep   int      `json:"request_time_sleep"`
	WarehouseID        string   `json:"warehouse_id"`
	MainKeyword        string   `json:"main_keyword"`
	Keywords           []string `json:"keywords"`
	KeyValues          []int    `json:"key_values"`
	SkipLines          int      `json:"skip_lines"`
	CountReadMessages  int      `json:"count_read_msg"`
	SortKeyword        string   `json:"sort_keyword"`
	IsSort             bool     `json:"sort"`
	IsSortInvert       bool     `json:"sort_invert"`
}

type Control struct {
	Token         string `json:"token"`
	AdminUsername string `json:"admin_username"`
	IsDebugLog    bool   `json:"debug_log"`
}

type Config struct {
	Ticker         *TimeTicker     `json:"ticker"`
	TelegramClient *TelegramClient `json:"telegram_client"`
	Sheets         *Sheets         `json:"sheets"`
	Parser         *Parser         `json:"parser"`
	Control        *Control        `json:"control"`
}

var config *Config = new(Config)

func Init(path string) error {
	logger.Init("Config.Init()", "Configuration")

	file, err := os.ReadFile(path)
	if err != nil {
		return logger.LogError("Config.Init()", "Error reading file configuration: ", err)
	}

	err = json.Unmarshal(file, config)
	if err != nil {
		return logger.LogError("Config.Init()", "Error unmarshalling file configuration: ", err)
	}

	logger.InitSuccessfully("Config.Init()", "Configuration")
	return nil
}

func Get() *Config {
	return config
}
