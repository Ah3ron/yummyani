# YummyAnime Player

TUI-клиент для поиска и просмотра аниме, написанный на Go.
Использует [YummyAni API](https://api.yani.tv) для поиска и плеер [Kodik](https://kodik.info) для воспроизведения.
Видео извлекается напрямую из Kodik через нативный Go-парсер (порт [kodik-parser](https://github.com/BatogiX/kodik)).

## Требования

- **Go** 1.24+
- **mpv** — плеер для воспроизведения (`mpv --version`)

## Установка

```bash
# Из исходников
cd yummyani
make install

# Или собрать бинарник вручную
make build
sudo cp yummyani /usr/local/bin/
```

## Запуск

```bash
yummyani
```

## Управление

| Клавиша | Действие |
|---------|----------|
| `Enter` | Подтвердить выбор / начать воспроизведение |
| `↑` / `k` | Навигация вверх |
| `↓` / `j` | Навигация вниз |
| `b` | Назад к озвучкам (из списка серий) |
| `q` | Назад / выход |
| `Esc` | Назад |
| `Ctrl+C` | Выход |

## Навигация

```
Поиск → Результаты → Озвучка → Серии → Воспроизведение
         ← q/esc    ← q/esc  ← q/esc/b   ← (после mpv)
```

1. **Поиск** — введите название аниме и нажмите Enter
2. **Результаты** — выберите аниме из списка (показаны статус, год, тип)
3. **Озвучки** — выберите студию озвучки (фильтруются только Kodik-плееры)
4. **Серии** — выберите серию для воспроизведения
5. **Воспроизведение** — видео открывается в mpv, после закрытия возврат к списку серий

## Структура проекта

```
yummyani/
├── cmd/yummyani/
│   └── main.go              # Точка входа
├── internal/
│   ├── api/
│   │   └── yummyani.go      # YummyAni API клиент (поиск, видео, группы)
│   ├── kodik/
│   │   └── kodik.go         # Kodik-парсер (извлечение прямых ссылок)
│   └── tui/
│       ├── model.go         # BubbleTea модель и state machine
│       ├── views.go         # Отрисовка экранов
│       └── styles.go        # Lipgloss-стили
├── go.mod
├── Makefile
└── README.md
```

### Как работает Kodik-парсер

Алгоритм извлечения прямой ссылки на видео из Kodik-плеера:

1. GET страницы плеера — установка cookies и сессии
2. Извлечение `video info` (type, hash, id) из HTML
3. GET player JS — поиск скрипта плеера в HTML
4. Извлечение API-endpoint — base64-закодированный путь (обычно `/ftor`)
5. POST запрос с 6 полями (type, hash, id, bad_user, info, cdn_is_working)
6. Декодирование ссылок — Caesar cipher (shift 0–26) + Base64

## Сборка

```bash
make build      # Собрать бинарник
make run        # Собрать и запустить
make clean      # Удалить бинарник
make lint       # Проверить код (go vet)
make fmt        # Отформатировать код
make test       # Запустить тесты
make archive    # Создать .tar.gz архив
make install    # Установить в систему
make help       # Показать доступные цели
```

## Зависимости

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — TUI фреймворк
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) — компоненты (spinner, text input)
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — стилизация

## Лицензия

MIT
