package main

import (
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
	_, _, err := keybdEvent.Call(keyCode, 0, 0, 0)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err
	}

	// Отпускание клавиши
	_, _, err = keybdEvent.Call(keyCode, 0, 2, 0)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err
	}

	return nil
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
