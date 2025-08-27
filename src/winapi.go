package main

import (
	"fmt"
	"syscall"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	keybdEvent          = user32.NewProc("keybd_event")
	VK_MEDIA_PLAY_PAUSE = uintptr(0xB3)
	VK_VOLUME_DOWN      = uintptr(0xAE)
	VK_VOLUME_UP        = uintptr(0xAF)
)

// SendMediaKey эмулирует нажатие мультимедийной клавиши через Windows API
func SendMediaKey(keyCode uintptr) error {
	// Нажатие клавиши
	ret, _, err := keybdEvent.Call(keyCode, 0, 0, 0)
	if ret == 0 { // Функция keybd_event возвращает 0 при ошибке
		return fmt.Errorf("ошибка эмуляции нажатия клавиши (код %d): %w", keyCode, err)
	}

	// Отпускание клавиши
	ret, _, err = keybdEvent.Call(keyCode, 0, 2, 0)
	if ret == 0 { // Функция keybd_event возвращает 0 при ошибке
		return fmt.Errorf("ошибка эмуляции отпускания клавиши (код %d): %w", keyCode, err)
	}

	return nil
}

// SendPlayPauseKey отправляет команду Play/Pause для управления воспроизведением медиа
func SendPlayPauseKey() error {
	return SendMediaKey(VK_MEDIA_PLAY_PAUSE)
}

// SendVolumeDownKey отправляет 5 нажатий клавиши Volume Down для заметного изменения громкости
func SendVolumeDownKey() error {
	for i := 0; i < 5; i++ {
		if err := SendMediaKey(VK_VOLUME_DOWN); err != nil {
			return err
		}
	}
	return nil
}

// SendVolumeUpKey отправляет 5 нажатий клавиши Volume Up для заметного изменения громкости
func SendVolumeUpKey() error {
	for i := 0; i < 5; i++ {
		if err := SendMediaKey(VK_VOLUME_UP); err != nil {
			return err
		}
	}
	return nil
}
