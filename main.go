package main

import (
	"bufio"
	"io"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const filePath = "credentials.txt"

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Login")

	// Виджеты для ввода логина и пароля
	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	// Виджеты для отображения сообщений
	message := widget.NewLabel("")
	passwordMessage := widget.NewLabel("Password must be at least 8 characters long")

	// Кнопка для регистрации нового логина и пароля
	registerButton := widget.NewButton("Register", func() {
		username := loginEntry.Text
		password := passwordEntry.Text

		if len(password) < 8 {
			passwordMessage.SetText("Password must be at least 8 characters long")
			return
		}

		if err := appendCredentials(username, password, ""); err != nil {
			message.SetText("Failed to save credentials")
			log.Println(err)
			return
		}

		// Открываем окно для ввода секретного слова
		newWindow := myApp.NewWindow("Set Secret Word")

		secretWordEntry := widget.NewEntry()
		secretWordEntry.SetPlaceHolder("Secret Word")

		secretWordMessage := widget.NewLabel("")

		saveSecretButton := widget.NewButton("Save Secret Word", func() {
			secretWord := secretWordEntry.Text

			// Обновляем запись с секретным словом
			if err := updateSecretWord(username, password, secretWord); err != nil {
				secretWordMessage.SetText("Failed to save secret word")
				log.Println(err)
				return
			}
			secretWordMessage.SetText("Secret word saved successfully")
			newWindow.Close()
		})

		content := container.NewVBox(
			widget.NewLabel("Enter Secret Word"),
			secretWordEntry,
			widget.NewLabel("Слово для восстановления пароля"),
			saveSecretButton,
			secretWordMessage,
		)

		newWindow.SetContent(content)
		newWindow.Resize(fyne.NewSize(300, 200))
		newWindow.Show()
	})

	// Кнопка для входа
	loginButton := widget.NewButton("Login", func() {
		username := loginEntry.Text
		password := passwordEntry.Text

		users, err := loadCredentials()
		if err != nil {
			message.SetText("No credentials found")
			log.Println(err)
			return
		}

		for _, user := range users {
			if username == user.username && password == user.password {
				message.SetText("Login successful")
				return
			}
		}
		message.SetText("Invalid login or password")
	})

	// Кнопка для восстановления пароля
	forgetPasswordButton := widget.NewButton("Forgot Password", func() {
		newWindow := myApp.NewWindow("Password Recovery")

		// Виджеты для ввода секретного слова
		recoverSecretWordEntry := widget.NewEntry()
		recoverSecretWordEntry.SetPlaceHolder("Enter Secret Word")

		recoverMessage := widget.NewLabel("")

		// Кнопка для восстановления пароля
		recoverButton := widget.NewButton("Recover Password", func() {
			secretWord := recoverSecretWordEntry.Text

			users, err := loadCredentials()
			if err != nil {
				recoverMessage.SetText("No credentials found")
				log.Println(err)
				return
			}

			for _, user := range users {
				if secretWord == user.secretWord {
					recoverMessage.SetText("Password: " + user.password)
					return
				}
			}
			recoverMessage.SetText("Incorrect secret word")
		})

		content := container.NewVBox(
			widget.NewLabel("Enter Secret Word for Password Recovery"),
			recoverSecretWordEntry,
			widget.NewLabel("Слово для восстановления пароля"),
			recoverButton,
			recoverMessage,
		)

		newWindow.SetContent(content)
		newWindow.Resize(fyne.NewSize(300, 200))
		newWindow.Show()
	})

	// Создание контейнера с виджетами
	content := container.NewVBox(
		widget.NewLabel("Login"),
		loginEntry,
		widget.NewLabel("Password"),
		passwordEntry,
		passwordMessage, // Сообщение о длине пароля
		loginButton,     // Перемещаем кнопку Login на место кнопки Register
		registerButton,  // Перемещаем кнопку Register на место кнопки Restore
		forgetPasswordButton,
		message,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(300, 300))
	myWindow.ShowAndRun()
}

// appendCredentials добавляет логин и пароль в файл
func appendCredentials(username, password, secretWord string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, username+"\n"+password+"\n"+secretWord+"\n")
	return err
}

// updateSecretWord обновляет запись с секретным словом в файле
func updateSecretWord(username, password, secretWord string) error {
	users, err := loadCredentials()
	if err != nil {
		return err
	}

	var updated []struct{ username, password, secretWord string }
	found := false
	for _, user := range users {
		if user.username == username && user.password == password {
			updated = append(updated, struct{ username, password, secretWord string }{username, password, secretWord})
			found = true
		} else {
			updated = append(updated, user)
		}
	}

	if !found {
		return nil // Не обновляем, если пользователь не найден
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, user := range updated {
		_, err := io.WriteString(file, user.username+"\n"+user.password+"\n"+user.secretWord+"\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// loadCredentials загружает все логины, пароли и секретные слова из файла
func loadCredentials() ([]struct{ username, password, secretWord string }, error) {
	var users []struct{ username, password, secretWord string }

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		username := scanner.Text()
		if !scanner.Scan() {
			return nil, io.EOF
		}
		password := scanner.Text()
		if !scanner.Scan() {
			return nil, io.EOF
		}
		secretWord := scanner.Text()
		users = append(users, struct{ username, password, secretWord string }{username, password, secretWord})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
