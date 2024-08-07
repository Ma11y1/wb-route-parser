package telegramClient

import (
	"errors"
	"github.com/zelenin/go-tdlib/client"
	"log"
	"time"
)

type ClientParameters struct {
	ApiId                 int32  `json:"api_id"`
	ApiHash               string `json:"api_hash"`
	UseTestDc             bool   `json:"use_test_dc"`
	DatabaseDirectory     string `json:"database_directory"`
	FilesDirectory        string `json:"files_directory"`
	DatabaseEncryptionKey []byte `json:"database_encryption_key"`
	UseFileDatabase       bool   `json:"use_file_database"`
	UseChatInfoDatabase   bool   `json:"use_chat_info_database"`
	UseMessageDatabase    bool   `json:"use_message_database"`
	UseSecretChats        bool   `json:"use_secret_chats"`
	SystemLanguageCode    string `json:"system_language_code"`
	DeviceModel           string `json:"device_model"`
	SystemVersion         string `json:"system_version"`
	ApplicationVersion    string `json:"application_version"`
}

type Client struct {
	id            int32
	hash          string
	client        *client.Client
	parameters    *client.SetTdlibParametersRequest
	inputHandler  AuthInputHandler
	user          *client.User
	chanAuthClose chan struct{}
	chanAuthReady chan bool
	isAuth        bool
}

func NewClient(id int32, hash string) *Client {
	return &Client{
		id:   id,
		hash: hash,
		parameters: &client.SetTdlibParametersRequest{
			UseTestDc:           false,
			DatabaseDirectory:   "./",
			FilesDirectory:      "",
			UseFileDatabase:     false,
			UseChatInfoDatabase: false,
			UseMessageDatabase:  false,
			UseSecretChats:      false,
			ApiId:               id,
			ApiHash:             hash,
			SystemLanguageCode:  "en",
			DeviceModel:         "Server",
			SystemVersion:       "1.0.0",
			ApplicationVersion:  "1.0.0",
		},
		inputHandler:  &ConsoleAuthInputHandler{},
		user:          nil,
		chanAuthClose: make(chan struct{}),
		chanAuthReady: make(chan bool),
		isAuth:        false,
	}
}

func NewClientByParameters(parameters *ClientParameters) (*Client, error) {
	c := &Client{}
	err := c.SetParameters(parameters)
	if err != nil {
		return nil, err
	}
	c.id = parameters.ApiId
	c.hash = parameters.ApiHash
	c.inputHandler = &ConsoleAuthInputHandler{}
	c.user = nil
	c.chanAuthClose = make(chan struct{})
	c.chanAuthReady = make(chan bool)
	c.isAuth = false

	return c, nil
}

func (c *Client) SetParameters(parameters *ClientParameters) error {
	if parameters == nil {
		return errors.New("parameters can not be nil")
	}

	if parameters.ApiId <= 0 {
		return errors.New("parameters.ApiId must be greater than zero")
	}

	if parameters.ApiHash == "" {
		return errors.New("parameters.ApiHash can not be empty")
	}

	defaults := map[*string]string{
		&parameters.DatabaseDirectory:  "./",
		&parameters.FilesDirectory:     "./",
		&parameters.SystemLanguageCode: "en",
		&parameters.DeviceModel:        "Server",
		&parameters.SystemVersion:      "1.0.0",
		&parameters.ApplicationVersion: "1.0.0",
	}

	for param, value := range defaults {
		if *param == "" {
			*param = value
		}
	}

	c.parameters = &client.SetTdlibParametersRequest{
		ApiId:                 parameters.ApiId,
		ApiHash:               parameters.ApiHash,
		UseTestDc:             parameters.UseTestDc,
		DatabaseDirectory:     parameters.DatabaseDirectory,
		FilesDirectory:        parameters.FilesDirectory,
		DatabaseEncryptionKey: parameters.DatabaseEncryptionKey,
		UseFileDatabase:       parameters.UseFileDatabase,
		UseChatInfoDatabase:   parameters.UseChatInfoDatabase,
		UseMessageDatabase:    parameters.UseMessageDatabase,
		UseSecretChats:        parameters.UseSecretChats,
		SystemLanguageCode:    parameters.SystemLanguageCode,
		DeviceModel:           parameters.DeviceModel,
		SystemVersion:         parameters.SystemVersion,
		ApplicationVersion:    parameters.ApplicationVersion,
	}

	return nil
}

func (c *Client) Auth() error {
	authorizer := client.ClientAuthorizer()

	go func() {
		defer close(c.chanAuthReady)
		time.Sleep(500 * time.Millisecond)

		for {
			select {
			case state, ok := <-authorizer.State:
				if !ok {
					var err error
					c.user, err = c.GetMe()
					if err != nil {
						c.isAuth = false
					} else {
						c.isAuth = true
					}

					c.chanAuthReady <- c.isAuth
					return
				}

				switch state.AuthorizationStateType() {
				case client.TypeAuthorizationStateWaitPhoneNumber:
					phoneNumber, err := c.inputHandler.InputPhoneNumber()
					if err != nil {
						log.Println(err)
						continue
					}
					authorizer.PhoneNumber <- phoneNumber

				case client.TypeAuthorizationStateWaitCode:
					code, err := c.inputHandler.InputCode()
					if err != nil {
						log.Println(err)
						continue
					}
					authorizer.Code <- code

				case client.TypeAuthorizationStateWaitPassword:
					password, err := c.inputHandler.InputPassword()
					if err != nil {
						log.Println(err)
						continue
					}
					authorizer.Password <- password

				case client.TypeAuthorizationStateLoggingOut:
					log.Println("Authorization state TDLib is LoggingOut. Need create new TDLib session")
					return
				case client.TypeAuthorizationStateReady:
					authorizer.Close()
					c.isAuth = true
					c.chanAuthReady <- c.isAuth
					return
				}
			case <-c.chanAuthClose:
				authorizer.Close()
				c.chanAuthReady <- false
				return
			}
		}
	}()

	authorizer.TdlibParameters <- c.parameters

	var err error
	c.client, err = client.NewClient(authorizer)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AuthReady() chan bool {
	return c.chanAuthReady
}

func (c *Client) LogOut() error {
	if c.client != nil {
		_, err := c.client.LogOut()
		if err != nil {
			return err
		}
	}

	c.isAuth = false

	return nil
}

func (c *Client) GetMe() (*client.User, error) {
	return c.client.GetMe()
}

func (c *Client) GetChats(chatList client.ChatList, limit int32) (*client.Chats, error) {
	if chatList == nil {
		chatList = &client.ChatListMain{}
	}

	return c.client.GetChats(&client.GetChatsRequest{
		ChatList: chatList,
		Limit:    limit,
	})
}

func (c *Client) GetChat(id int64) (*client.Chat, error) {
	return c.client.GetChat(&client.GetChatRequest{ChatId: id})
}

func (c *Client) SearchChats(query string, limit int32) (*client.Chats, error) {
	return c.client.SearchChats(&client.SearchChatsRequest{
		Query: query,
		Limit: limit,
	})
}

func (c *Client) SearchPublicChat(username string) (*client.Chat, error) {
	return c.client.SearchPublicChat(&client.SearchPublicChatRequest{
		Username: username,
	})
}

func (c *Client) SendMessage(chatID int64, message client.InputMessageContent) (*client.Message, error) {
	return c.client.SendMessage(&client.SendMessageRequest{
		ChatId:              chatID,
		InputMessageContent: message,
	})
}

func (c *Client) SendMessageText(chatID int64, message string) (*client.Message, error) {
	return c.client.SendMessage(&client.SendMessageRequest{
		ChatId: chatID,
		InputMessageContent: &client.InputMessageText{
			Text: &client.FormattedText{
				Text: message,
			},
		},
	})
}

func (c *Client) GetChatMessages(id int64, limit int32) (*client.Messages, error) {
	return c.client.GetChatHistory(&client.GetChatHistoryRequest{
		ChatId: id,
		Limit:  limit,
	})
}

func (c *Client) GetChatMessagesText(id int64, limit int32) ([]string, error) {
	messages, err := c.GetChatMessages(id, limit)
	if err != nil {
		return nil, err
	}

	textMessages := make([]string, len(messages.Messages))

	for i := 0; i < len(messages.Messages); i++ {
		contentType := messages.Messages[i].Content.MessageContentType()

		if contentType == client.TypeMessageText {
			text := messages.Messages[i].Content.(*client.MessageText).Text.Text

			if len(text) != 0 {
				textMessages[i] = text
			}
		}
	}

	return textMessages, nil
}

func (c *Client) Version() (string, error) {
	optionValue, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		return "", err
	}

	return optionValue.(*client.OptionValueString).Value, nil
}

func (c *Client) IsAuth() bool {
	return c.isAuth
}

func (c *Client) Close() {
	close(c.chanAuthClose)

	if c.client != nil {
		_, err := c.client.Close()
		if err != nil {
			log.Println(err)
		}
	}

	c.isAuth = false
}
