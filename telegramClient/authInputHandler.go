package telegramClient

type AuthInputHandler interface {
	InputPhoneNumber() (string, error)
	InputCode() (string, error)
	InputPassword() (string, error)
}
