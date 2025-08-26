package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Настройка логгирования
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Ошибка получения пути к исполняемому файлу: %v", err)
	}
	logPath := filepath.Join(filepath.Dir(exePath), "tbot-controls-pc.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Не удалось открыть файл лога: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Бот запускается...")

	// Загрузка .env файла
	err = godotenv.Load("settings.env")
	if err != nil {
		log.Fatalf("Ошибка загрузки файла settings.env: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не найден в .env")
	}

	authorizedUsersStr := os.Getenv("TELEGRAM_AUTHORIZED_USER_IDS")
	if authorizedUsersStr == "" {
		log.Fatal("TELEGRAM_AUTHORIZED_USER_IDS не найдены в .env")
	}

	authorizedUsersMap := make(map[int64]bool)
	for _, s := range strings.Split(authorizedUsersStr, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Fatalf("Ошибка преобразования ID пользователя '%s': %v", s, err)
		}
		authorizedUsersMap[id] = true
	}
	log.Printf("Авторизованные пользователи: %v", authorizedUsersMap)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Создание клавиатуры
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💤", "hibernate"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏯️", "play_pause"),
			tgbotapi.NewInlineKeyboardButtonData("🔉", "volume_down"),
			tgbotapi.NewInlineKeyboardButtonData("🔊", "volume_up"),
		),
	)

	// Определение пути к файлу изображения
	imagePath := filepath.Join(filepath.Dir(exePath), "tbot-picture.jpg")

	// Запуск цикла обработки обновлений Telegram
	for update := range updates {
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
			}

			// Проверяем, что callback от авторизованного пользователя
			if !authorizedUsersMap[update.CallbackQuery.From.ID] {
				log.Printf("Неавторизованный доступ от пользователя %d", update.CallbackQuery.From.ID)
				continue
			}

			log.Printf("Нажата кнопка: %s", update.CallbackQuery.Data)

			// Обработка нажатий кнопок
			switch update.CallbackQuery.Data {
			case "play_pause":
				if err := SendMediaKey(0xB3); err != nil {
					log.Printf("Ошибка отправки команды Play/Pause: %v", err)
				} else {
					log.Printf("Команда Play/Pause выполнена успешно")
				}
			case "hibernate":
				if err := HibernatePC(); err != nil {
					log.Printf("Ошибка выполнения гибернации: %v", err)
				} else {
					log.Printf("Команда гибернации выполнена успешно")
				}
			case "volume_down":
				if err := SendVolumeDownKey(); err != nil {
					log.Printf("Ошибка отправки команды Volume Down: %v", err)
				} else {
					log.Printf("Команда Volume Down выполнена успешно (5 нажатий)")
				}
			case "volume_up":
				if err := SendVolumeUpKey(); err != nil {
					log.Printf("Ошибка отправки команды Volume Up: %v", err)
				} else {
					log.Printf("Команда Volume Up выполнена успешно (5 нажатий)")
				}
			}
		}
		// Обработка текстовых сообщений
		if update.Message != nil {
			// Проверяем, что сообщение от авторизованного пользователя
			if !authorizedUsersMap[update.Message.From.ID] {
				log.Printf("Неавторизованное сообщение от пользователя %d: %s", update.Message.From.ID, update.Message.Text)
				continue
			}

			log.Printf("Получено сообщение от %s: %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					sendPanelToUser(bot, update.Message.From.ID, keyboard, imagePath)
				}
			}
		}
	}

	log.Println("Бот завершает работу.")
}

// sendPanelToUser отправляет панель управления (с фото или без) указанному пользователю
func sendPanelToUser(bot *tgbotapi.BotAPI, userID int64, keyboard tgbotapi.InlineKeyboardMarkup, imagePath string) {
	if _, err := os.Stat(imagePath); err == nil {
		// Файл изображения существует, отправляем фото
		photoMsg := tgbotapi.NewPhoto(userID, tgbotapi.FilePath(imagePath))
		photoMsg.ReplyMarkup = keyboard
		photoMsg.Caption = "Панель управления ПК" // Добавляем подпись к фото
		if _, err := bot.Send(photoMsg); err != nil {
			log.Printf("Не удалось отправить фото пользователю %d: %v. Отправляем текстовое сообщение.", userID, err)
			sendTextMessage(bot, userID, "Панель управления ПК", keyboard)
		} else {
			log.Printf("Фото '%s' успешно отправлено пользователю %d", imagePath, userID)
		}
	} else if os.IsNotExist(err) {
		log.Printf("Файл изображения '%s' не найден. Отправляем текстовое сообщение пользователю %d.", imagePath, userID)
		sendTextMessage(bot, userID, "Панель управления ПК", keyboard)
	} else {
		log.Printf("Ошибка при проверке файла изображения '%s': %v. Отправляем текстовое сообщение пользователю %d.", imagePath, err)
		sendTextMessage(bot, userID, "Панель управления ПК", keyboard)
	}
}

// sendTextMessage отправляет текстовое сообщение с клавиатурой
func sendTextMessage(bot *tgbotapi.BotAPI, userID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Не удалось отправить текстовое сообщение пользователю %d: %v", userID, err)
	} else {
		log.Printf("Текстовое сообщение с панелью успешно отправлено пользователю %d", userID)
	}
}

// HibernatePC выполняет команду гибернации ПК
func HibernatePC() error {
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	return cmd.Run()
}
