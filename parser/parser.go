package parser

import (
	"strconv"
	"strings"
	"time"
	"wb-assistance-logistic/config"
	"wb-assistance-logistic/logger"
	"wb-assistance-logistic/telegramClient"
)

type Parser struct {
	client *telegramClient.Client

	chatID            int64
	chatUsername      string
	countReadMessages int32
	skipLines         int
	sortIndex         int
	isSort            bool
	isInvertSort      bool

	warehouseID string
	mainKeyword string
	sortKeyword string
	keywords    []string
	keyValues   []int

	data [][]int

	commandRequestWarehouseRoutes string
	requestTimeSleep              time.Duration
}

func NewParser(cfg *config.Parser, client *telegramClient.Client) (*Parser, error) {
	if !client.IsAuth() {
		return nil, logger.Error("Parser.NewParser()", "Telegram client is not auth")
	}

	validationErrors := map[string]bool{
		"Invalid chat username: ":                       cfg.ChatUsername == "",
		"Invalid count read messages: ":                 cfg.CountReadMessages == 0,
		"Invalid count skip lines: ":                    cfg.SkipLines < 0,
		"Invalid warehouse id: ":                        len(cfg.WarehouseID) <= 4,
		"Invalid warehouse main keyword: ":              cfg.MainKeyword == "",
		"Invalid warehouse command request warehouse: ": cfg.CommandRequestData == "",
		"Invalid warehouse keywords: ":                  cfg.Keywords == nil,
		"Invalid warehouse key values: ":                cfg.KeyValues == nil,
		"Invalid sort keyword: ":                        cfg.IsSort && cfg.SortKeyword == "",
		"Invalid request time sleep: ":                  cfg.RequestTimeSleep < 500,
	}

	for msg, invalid := range validationErrors {
		if invalid {
			return nil, logger.Error("Parser.NewParser()", msg, invalid)
		}
	}

	parser := &Parser{
		client:                        client,
		data:                          nil,
		chatID:                        -1,
		chatUsername:                  cfg.ChatUsername,
		countReadMessages:             int32(cfg.CountReadMessages),
		skipLines:                     cfg.SkipLines,
		sortIndex:                     -1,
		isSort:                        cfg.IsSort,
		isInvertSort:                  cfg.IsSortInvert,
		warehouseID:                   cfg.WarehouseID,
		mainKeyword:                   cfg.MainKeyword,
		sortKeyword:                   cfg.SortKeyword,
		keywords:                      cfg.Keywords,
		keyValues:                     cfg.KeyValues,
		commandRequestWarehouseRoutes: cfg.CommandRequestData,
		requestTimeSleep:              time.Duration(cfg.RequestTimeSleep) * time.Millisecond,
	}

	// The data after parsing is located in the same way as in the keywords array, so the sorting index will coincide with the sorting keyword
	for i, keyword := range parser.keywords {
		if keyword == parser.sortKeyword {
			parser.sortIndex = i
		}
	}

	if parser.sortIndex < 0 {
		return nil, logger.Error("Parser.NewParser()", "Invalid sorting keyword: ", cfg.SortKeyword)
	}

	chat, err := client.SearchPublicChat(parser.chatUsername)
	if err != nil {
		return nil, logger.Error("Parser.NewParser()", "Error search chat by chat username:\n", err)
	}
	parser.chatID = chat.Id

	return parser, nil
}

func (p *Parser) Parse() ([][]int, error) {
	err := p.sendRequestWarehouseRoutes()
	if err != nil {
		return nil, logger.Error("Parser.Parse()", "Error request warehouse routes:\n", err)
	}

	time.Sleep(p.requestTimeSleep)

	p.data, err = p.getDataWarehouseRoutes()
	if err != nil {
		return nil, logger.Error("Parser.Parse()", "Error parse data routes:\n", err)
	}

	if p.isSort {
		err = p.sortData()
		if err != nil {
			return nil, logger.Error("Parser.Parse()", "Error sort data routes:\n", err)
		}
	}

	return p.data, nil
}

func (p *Parser) sendRequestWarehouseRoutes() error {
	_, err := p.client.SendMessageText(p.chatID, p.commandRequestWarehouseRoutes)
	if err != nil {
		return logger.Error("Parser.sendRequestWarehouseRoutes()", "Error sending request warehouse routes:\n", err)
	}

	time.Sleep(p.requestTimeSleep)

	_, err = p.client.SendMessageText(p.chatID, p.warehouseID)
	if err != nil {
		return logger.Error("Parser.sendRequestWarehouseRoutes()", "Error sending request warehouse routes:\n", err)
	}

	return nil
}

func (p *Parser) getDataWarehouseRoutes() ([][]int, error) {
	messages, err := p.getMessages()
	if err != nil {
		return nil, logger.Error("Parser.getDataWarehouseRoutes()", "Error getting messages:\n", err)
	}

	if len(messages) == 0 {
		return nil, logger.Error("Parser.getDataWarehouseRoutes()", "No messages")
	}

	var routesData [][]int

	for _, message := range messages {
		lines := strings.Split(message, "\n")
		skipLines := 0

		for i := 0; i < len(lines); i++ {
			if skipLines > p.skipLines {
				return nil, logger.Error("Parser.getDataWarehouseRoutes()", "The permissible skip line value has been exceeded: ", skipLines)
			}
			// Retrieve the number of the main keyword, checking for the presence of the main keyword
			// If there are more missing lines than can be skipped, we exit the function
			number, err := extractNumberAfterKeyword(lines[i], p.mainKeyword)
			if err != nil {
				skipLines++
				logger.Warning("Parser.getDataWarehouseRoutes()", "Error extracting main keyword number from line: "+lines[i]+"\n", err)
				continue
			}

			// Checking whether a string should be added to the array
			if p.isContainsValue(number) {
				numbers, err := p.extractNumbers(lines[i])
				if err != nil {
					skipLines++
					logger.Warning("Parser.getDataWarehouseRoutes()", "Error extracting numbers from line: "+lines[i]+"\n", err)
					continue
				}

				routesData = append(routesData, numbers)
			}
		}
	}

	return routesData, nil
}

func (p *Parser) getMessages() ([]string, error) {
	messages, err := p.client.GetChatMessagesText(p.chatID, p.countReadMessages)
	if err != nil {
		return nil, logger.Error("Parser.getMessages()", "Error getting messages:\n", err)
	}

	return messages, nil
}

func (p *Parser) sortData() error {
	if p.isInvertSort {
		for i := 0; i < len(p.data); i++ {
			for j := i + 1; j < len(p.data); j++ {
				if p.data[j][p.sortIndex] > p.data[i][p.sortIndex] {
					p.data[i], p.data[j] = p.data[j], p.data[i]
				}
			}
		}
	} else {
		for i := 0; i < len(p.data); i++ {
			for j := i + 1; j < len(p.data); j++ {
				if p.data[j][p.sortIndex] < p.data[i][p.sortIndex] {
					p.data[i], p.data[j] = p.data[j], p.data[i]
				}
			}
		}
	}

	return nil
}

func (p *Parser) isContainsValue(value int) bool {
	for i := 0; i < len(p.keyValues); i++ {
		if p.keyValues[i] == value {
			return true
		}
	}

	return false
}

func (p *Parser) extractNumbers(text string) ([]int, error) {
	numbers := make([]int, len(p.keywords))

	var err error

	for i, keyword := range p.keywords {
		if len(keyword) == 0 {
			continue
		}

		numbers[i], err = extractNumberAfterKeyword(text, keyword)
		if err != nil {
			return nil, err
		}
	}
	return numbers, nil
}

func extractNumberAfterKeyword(text, keyword string) (int, error) {
	start := strings.Index(text, keyword)
	if start == -1 {
		return 0, logger.Error("Parser.extractNumberAfterKeyword()", "Keyword not found in text: "+keyword)
	}

	start += len(keyword)

	for start < len(text) && (text[start] < '0' || text[start] > '9') {
		start++
	}

	end := start
	for end < len(text) && text[end] >= '0' && text[end] <= '9' {
		end++
	}

	number, err := strconv.Atoi(text[start:end])
	if err != nil {
		return 0, logger.Error("Parser.extractNumberAfterKeyword()", "Unable to convert "+text[start:end]+" to number:\n", err)
	}

	return number, nil
}
