# Инструкция по релизу v1.0.0

## Подготовка к релизу

### 1. Проверка готовности
- [ ] Все 4 этапа разработки завершены
- [ ] Код протестирован и работает
- [ ] Документация готова (README.md, CHANGELOG.md)
- [ ] Структура проекта причесана

### 2. Финальная сборка
```bash
# Сборка для релиза (фоновый режим)
go build -ldflags "-H=windowsgui" -o tbot-controls-pc.exe ./src

# Проверка работы
./tbot-controls-pc.exe
```

## Создание релиза

### 3. Коммит изменений
```bash
# Добавить все файлы
git add .

# Финальный коммит
git commit -m "Release v1.0.0 - Complete Telegram bot for PC control

- All 4 development stages completed
- Media controls, volume, hibernation
- Windows API integration
- Background mode support
- Comprehensive documentation"
```

### 4. Создание и пуш тэга
```bash
# Создать тэг
git tag v1.0.0

# Запушить тэг
git push origin v1.0.0
```

### 5. Пуш изменений
```bash
git push origin main
```

## GitHub Release

### 6. Создание Release на GitHub
1. Перейти в [Releases](https://github.com/DiscipulusVitae/tbot-controls-pc/releases)
2. Нажать "Create a new release"
3. Выбрать тэг `v1.0.0`
4. Заголовок: `v1.0.0 - Complete PC Control Bot`
5. Описание:
```
## 🎉 Первый релиз tbot-controls-pc

Telegram-бот для управления ПК на Windows с полным набором функций.

### ✨ Что готово
- ⏯️ Управление медиаплеером (Play/Pause)
- 🔉🔊 Управление громкостью (5 делений за нажатие)
- 💤 Перевод ПК в режим гибернации
- 🔌 Автозапуск через Планировщик Windows
- 📱 Удобный Telegram интерфейс
- 🛡️ Безопасность и логирование

### 📦 Файлы для скачивания
- `tbot-controls-pc.exe` - исполняемый файл
- `README.md` - инструкция по установке
- `CHANGELOG.md` - история изменений

### 🚀 Быстрый старт
1. Скачайте `tbot-controls-pc.exe`
2. Создайте `settings.env` с токеном и ID
3. Запустите бота
4. Настройте автозапуск

### 📋 Требования
- Windows 11 x64
- Telegram Bot Token
- Telegram User ID
```

### 7. Загрузка файлов
- Загрузить `tbot-controls-pc.exe`
- Загрузить `README.md`
- Загрузить `CHANGELOG.md`

### 8. Публикация
- Нажать "Publish release"

## После релиза

### 9. Обновление документации
- Обновить `docs/project-state.md` с информацией о релизе
- Проверить все ссылки в документации

### 10. Проверка
- [ ] Release создан на GitHub
- [ ] Файлы доступны для скачивания
- [ ] Документация актуальна
- [ ] Проект готов к использованию

---

**Дата релиза:** 2024-08-25
**Версия:** v1.0.0
**Статус:** Готов к релизу
