package telegramClient

import "github.com/zelenin/go-tdlib/client"

type LogsVerboseLevel int32

const (
	FATAL         LogsVerboseLevel = 0
	ERRORS        LogsVerboseLevel = 1
	WARNINGS      LogsVerboseLevel = 2
	INFO          LogsVerboseLevel = 3
	DEBUG         LogsVerboseLevel = 4
	VERBOSE_DEBUG LogsVerboseLevel = 5
)

func SetTelegramClientLogsVerboseLevel(level LogsVerboseLevel) error {
	_, err := client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: int32(level),
	})
	if err != nil {
		return err
	}

	return nil
}
