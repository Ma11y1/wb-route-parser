package telegramClient

import "fmt"

type ConsoleAuthInputHandler struct{}

func (a *ConsoleAuthInputHandler) InputPhoneNumber() (string, error) {
	fmt.Print("Phone number: ")
	var phone string
	_, err := fmt.Scanln(&phone)
	return phone, err
}

func (a *ConsoleAuthInputHandler) InputCode() (string, error) {
	fmt.Print("Security code: ")
	var code string
	_, err := fmt.Scanln(&code)
	return code, err
}

func (a *ConsoleAuthInputHandler) InputPassword() (string, error) {
	fmt.Print("Password: ")
	var password string
	_, err := fmt.Scanln(&password)
	return password, err
}
