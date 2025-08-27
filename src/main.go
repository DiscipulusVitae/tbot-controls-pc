package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Обработка паники для graceful завершения
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ПАНИКА в main: %v", r)
			log.Println("Бот завершает работу из-за непредвиденной ошибки")
		}
	}()

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
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось загрузить settings.env: %v", err)
		log.Println("Будут использоваться переменные окружения системы")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Println("ОШИБКА: TELEGRAM_BOT_TOKEN не найден")
		log.Println("Убедитесь, что файл settings.env существует и содержит TELEGRAM_BOT_TOKEN")
		log.Println("Бот завершает работу из-за отсутствия токена")
		return
	}

	authorizedUsersStr := os.Getenv("TELEGRAM_AUTHORIZED_USER_IDS")
	if authorizedUsersStr == "" {
		log.Println("ОШИБКА: TELEGRAM_AUTHORIZED_USER_IDS не найдены")
		log.Println("Убедитесь, что файл settings.env содержит TELEGRAM_AUTHORIZED_USER_IDS")
		log.Println("Бот завершает работу из-за отсутствия авторизованных пользователей")
		return
	}

	authorizedUsersMap := make(map[int64]bool)
	for _, s := range strings.Split(authorizedUsersStr, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Пропускаем невалидный ID пользователя '%s': %v", s, err)
			continue
		}
		authorizedUsersMap[id] = true
	}

	// Валидация: проверяем, что есть хотя бы один авторизованный пользователь
	if len(authorizedUsersMap) == 0 {
		log.Println("ОШИБКА: Не найдено ни одного валидного ID пользователя в TELEGRAM_AUTHORIZED_USER_IDS")
		log.Println("Проверьте формат: ID1,ID2,ID3 (только цифры, через запятую)")
		log.Println("Бот завершает работу из-за отсутствия авторизованных пользователей")
		return
	}

	log.Printf("Загружено %d авторизованных пользователей", len(authorizedUsersMap))

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("КРИТИЧЕСКАЯ ОШИБКА: Невозможно подключиться к Telegram API: %v", err)
		log.Println("Проверяйте TELEGRAM_BOT_TOKEN в settings.env")
		log.Println("Бот завершает работу из-за ошибки подключения")
		return
	}

	bot.Debug = false
	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Создание контекста с отменой для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	// Запуск цикла обработки обновлений Telegram в отдельной горутине
	go func() {
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
					if err := SendPlayPauseKey(); err != nil {
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
	}()

	// Ожидание сигнала отмены
	<-ctx.Done()

	// Graceful shutdown
	log.Println("Получен сигнал завершения, останавливаем бота...")
	bot.StopReceivingUpdates()
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
			msg := tgbotapi.NewMessage(userID, "Панель управления ПК")
			msg.ReplyMarkup = keyboard
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Не удалось отправить текстовое сообщение пользователю %d: %v", userID, err)
			} else {
				log.Printf("Текстовое сообщение с панелью успешно отправлено пользователю %d", userID)
			}
		} else {
			log.Printf("Фото '%s' успешно отправлено пользователю %d", imagePath, userID)
		}
	} else if os.IsNotExist(err) {
		log.Printf("Файл изображения '%s' не найден. Отправляем текстовое сообщение пользователю %d.", imagePath, userID)
		msg := tgbotapi.NewMessage(userID, "Панель управления ПК")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Не удалось отправить текстовое сообщение пользователю %d: %v", userID, err)
		} else {
			log.Printf("Текстовое сообщение с панелью успешно отправлено пользователю %d", userID)
		}
	} else {
		log.Printf("Ошибка при проверке файла изображения '%s': %v. Отправляем текстовое сообщение пользователю %d.", imagePath, err)
		msg := tgbotapi.NewMessage(userID, "Панель управления ПК")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Не удалось отправить текстовое сообщение пользователю %d: %v", userID, err)
		} else {
			log.Printf("Текстовое сообщение с панелью успешно отправлено пользователю %d", userID)
		}
	}
}

// HibernatePC выполняет команду гибернации ПК
func HibernatePC() error {
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ошибка выполнения команды гибернации: %w", err)
	}

	return nil
}
