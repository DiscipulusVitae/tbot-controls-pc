package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
	userIDStr := os.Getenv("TELEGRAM_USER_ID")
	if userIDStr == "" {
		log.Fatal("TELEGRAM_USER_ID не найден в .env")
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Ошибка преобразования TELEGRAM_USER_ID в число: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Отправка клавиатуры при старте
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
	msg := tgbotapi.NewMessage(userID, "Панель управления ПК")
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Не удалось отправить клавиатуру: %v", err)
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
			}

			// Проверяем, что callback от авторизованного пользователя
			if update.CallbackQuery.From.ID != userID {
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
	}

	log.Println("Бот завершает работу.")
}

// HibernatePC выполняет команду гибернации ПК
func HibernatePC() error {
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	return cmd.Run()
}
