package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// PanelState хранит состояние панели управления
type PanelState struct {
	MessageID int   `json:"message_id"`
	ChatID    int64 `json:"chat_id"`
}

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

	// Попытка загрузить состояние панели
	panelState := loadPanelState(exePath)

	if panelState.MessageID != 0 && panelState.ChatID == userID {
		log.Println("Обнаружено сохраненное состояние. Попытка обновить панель управления...")
		// Check if the image file exists to determine if we are dealing with a photo message or a text message.
		_, imageErr := os.Stat(imagePath)
		isImagePresent := imageErr == nil

		var err error
		if isImagePresent {
			// Attempt to edit the reply markup of the existing photo message.
			editMarkup := tgbotapi.NewEditMessageReplyMarkup(userID, panelState.MessageID, keyboard)
			_, err = bot.Request(editMarkup)
		} else {
			// Attempt to edit the reply markup of the existing text message.
			// The existing sendOrEditMessage function does this for messageID != 0 and will send a new one if editing fails.
			sendOrEditMessage(bot, userID, panelState.MessageID, "Панель управления ПК", keyboard, exePath)
			// No need to continue the outer if/else, as sendOrEditMessage handles the success/failure and new message sending.
			goto EndOfPanelUpdate // Use goto to skip the rest of the block
		}

		if err != nil {
			if strings.Contains(err.Error(), "message is not modified") {
				log.Println("Панель уже актуальна, обновление не требуется.")
			} else {
				log.Printf("Не удалось отредактировать панель (возможно, она была удалена): %v", err)
				log.Println("Отправляем новую панель управления.")
				sendInitialMessage(bot, userID, keyboard, imagePath, exePath)
			}
		} else {
			log.Println("Существующая панель успешно обновлена")
		}
	} else {
		log.Println("Отправляем новую панель управления")
		sendInitialMessage(bot, userID, keyboard, imagePath, exePath)
	}
EndOfPanelUpdate:

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

// PanelState хранит состояние панели управления
// type PanelState struct {
// 	MessageID int   `json:"message_id"`
// 	ChatID    int64 `json:"chat_id"`
// }

// loadPanelState загружает состояние панели из файла
func loadPanelState(exePath string) PanelState {
	statePath := filepath.Join(filepath.Dir(exePath), "panel-state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Не удалось прочитать файл состояния '%s': %v", statePath, err)
		}
		return PanelState{}
	}

	var state PanelState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Ошибка чтения JSON из файла состояния '%s': %v", statePath, err)
		return PanelState{}
	}

	log.Printf("Состояние панели успешно загружено из '%s': MessageID=%d", statePath, state.MessageID)
	return state
}

// savePanelState сохраняет состояние панели в файл
func savePanelState(exePath string, messageID int, chatID int64) {
	state := PanelState{
		MessageID: messageID,
		ChatID:    chatID,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Ошибка сериализации состояния панели: %v", err)
		return
	}

	statePath := filepath.Join(filepath.Dir(exePath), "panel-state.json")
	if err := os.WriteFile(statePath, data, 0666); err != nil {
		log.Printf("Ошибка сохранения состояния панели в '%s': %v", statePath, err)
	} else {
		log.Printf("Состояние панели успешно сохранено в '%s'", statePath)
	}
}

// sendInitialMessage отправляет начальное сообщение (с фото или без) и клавиатуру
func sendInitialMessage(bot *tgbotapi.BotAPI, userID int64, keyboard tgbotapi.InlineKeyboardMarkup, imagePath, exePath string) {
	if _, err := os.Stat(imagePath); err == nil {
		// Файл изображения существует, отправляем фото
		photoMsg := tgbotapi.NewPhoto(userID, tgbotapi.FilePath(imagePath))
		photoMsg.ReplyMarkup = keyboard
		photoMsg.Caption = "Панель управления ПК" // Добавляем подпись к фото
		if sentMsg, err := bot.Send(photoMsg); err != nil {
			log.Printf("Не удалось отправить фото: %v", err)
			// Если отправка фото не удалась, пробуем отправить текстовое сообщение
			sendOrEditMessage(bot, userID, 0, "Панель управления ПК", keyboard, exePath)
		} else {
			log.Printf("Фото '%s' успешно отправлено", imagePath)
			savePanelState(exePath, sentMsg.MessageID, sentMsg.Chat.ID)
		}
	} else if os.IsNotExist(err) {
		log.Printf("Файл изображения '%s' не найден. Отправляем текстовое сообщение.", imagePath)
		sendOrEditMessage(bot, userID, 0, "Панель управления ПК", keyboard, exePath)
	} else {
		log.Printf("Ошибка при проверке файла изображения '%s': %v. Отправляем текстовое сообщение.", imagePath, err)
		sendOrEditMessage(bot, userID, 0, "Панель управления ПК", keyboard, exePath)
	}
}

// sendOrEditMessage отправляет или редактирует текстовое сообщение с клавиатурой
func sendOrEditMessage(bot *tgbotapi.BotAPI, userID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup, exePath string) {
	if messageID != 0 {
		editMsg := tgbotapi.NewEditMessageText(userID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Request(editMsg); err != nil {
			if strings.Contains(err.Error(), "message is not modified") {
				log.Println("Текстовое сообщение уже актуально, обновление не требуется.")
			} else {
				log.Printf("Не удалось отредактировать текстовое сообщение: %v", err)
				// Если редактирование не удалось, отправляем новое текстовое сообщение
				log.Println("Отправляем новое текстовое сообщение с панелью.")
				msg := tgbotapi.NewMessage(userID, text)
				msg.ReplyMarkup = keyboard
				if sentMsg, err := bot.Send(msg); err != nil {
					log.Printf("Не удалось отправить текстовое сообщение: %v", err)
				} else {
					log.Printf("Новое текстовое сообщение с панелью успешно отправлено")
					savePanelState(exePath, sentMsg.MessageID, sentMsg.Chat.ID)
				}
			}
		} else {
			log.Println("Существующее текстовое сообщение успешно обновлено")
		}
	} else {
		msg := tgbotapi.NewMessage(userID, text)
		msg.ReplyMarkup = keyboard
		if sentMsg, err := bot.Send(msg); err != nil {
			log.Printf("Не удалось отправить текстовое сообщение: %v", err)
		} else {
			log.Printf("Текстовое сообщение с панелью успешно отправлено")
			savePanelState(exePath, sentMsg.MessageID, sentMsg.Chat.ID)
		}
	}
}

// HibernatePC выполняет команду гибернации ПК
func HibernatePC() error {
	cmd := exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	return cmd.Run()
}
