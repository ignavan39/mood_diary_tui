# Mood Diary - Дневник Настроения

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Platform](https://img.shields.io/badge/Platform-Linux-orange.svg)

Красивый терминальный дневник настроения с интерактивным интерфейсом. Отслеживайте свои эмоции, анализируйте тренды и заботьтесь о своём ментальном здоровье.


## Возможности

- **Запись настроения** - Отмечайте своё настроение по шкале от 0 до 10 с эмоджи
- **Интерактивная статистика** - Визуализация данных с графиками и трендами
- **История записей** - Просмотр всех предыдущих записей
- **Интерфейс** - Приятный глазу интерфейс в мягких тонах
- **Удобная навигация** - Управление клавиатурой и мышью
- **Локальное хранение** - Все данные хранятся локально в SQLite
- **Clean Architecture** - Чистая архитектура с DDD принципами

## Технологии

- **Язык**: Go 1.21+
- **UI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The Elm Architecture для терминала
- **Стилизация**: [Lipgloss](https://github.com/charmbracelet/lipgloss) - Стили и цвета
- **Компоненты**: [Bubbles](https://github.com/charmbracelet/bubbles) - Готовые UI компоненты
- **База данных**: SQLite с драйвером [go-sqlite3](https://github.com/mattn/go-sqlite3)
- **UUID**: [Google UUID](https://github.com/google/uuid)

## Установка

### Предварительные требования

- Go 1.21 или выше
- GCC (для компиляции SQLite)
- Linux/macOS (или WSL на Windows)

### Клонирование репозитория

```bash
git clone https://github.com/ignavan39/mood-diary.git
cd mood-diary
```

### Установка зависимостей

```bash
go mod download
```

### Сборка приложения

```bash
go build -o mood-diary ./cmd/mood-diary
```

### Запуск

```bash
./mood-diary
```

Или напрямую через Go:

```bash
go run ./cmd/mood-diary
```

## Использование

### Главное меню

При запуске вы увидите главное меню с опциями:

```
→ 📝 Записать настроение
  📊 Посмотреть статистику
  📅 История записей
  ❌ Выход
```

**Управление:**
- `↑/↓` или `j/k` - навигация
- `Enter` - выбор
- `q` или `Ctrl+C` - выход

### Запись настроения

1. Выберите "Записать настроение" в главном меню
2. Используйте `←/→` для выбора уровня настроения (0-10)
3. Нажмите `Enter` для перехода к заметке
4. Введите заметку (необязательно) и нажмите `Enter`
5. Подтвердите запись нажатием `y` или `Enter`

**Шкала настроений:**
```
😢 😞 😔 😕 😐 😶 🙂 😊 😄 😁 🤩
0  1  2  3  4  5  6  7  8  9  10
```

### Просмотр статистики

Статистика показывает:
- **Всего записей** за выбранный период
- **Средний уровень** настроения
- **Тренд** - улучшается, ухудшается или стабильно
- **Распределение** - гистограмма по уровням настроения
- **Динамика** - график изменения настроения за период

**Периоды:**
- Неделя (7 дней)
- Месяц (30 дней)
- Квартал (90 дней)
- Год (365 дней)
- Всё время

**Управление:**
- `←/→` - переключение периодов
- `r` - обновить данные
- `Esc` - вернуться в меню

### История записей

Просмотр всех предыдущих записей в табличном формате:

```
Дата          Настроение       Заметка
──────────────────────────────────────────
01.04.2026    😊 8/10         Отличный день!
31.03.2026    😐 5/10         Обычный день
30.03.2026    😄 9/10         Закончил проект
```

**Управление:**
- `↑/↓` или `j/k` - навигация по записям
- `Enter` - редактировать запись (в разработке)
- `r` - обновить список
- `Esc` - вернуться в меню

## Архитектура

Проект следует принципам **Domain-Driven Design (DDD)** и **Clean Architecture**:

```
mood-diary/
├── cmd/
│   └── mood-diary/          # Точка входа приложения
│       └── main.go
├── internal/
│   ├── domain/              # Бизнес-логика (ядро)
│   │   ├── entity/          # Сущности
│   │   ├── repository/      # Интерфейсы репозиториев
│   │   └── service/         # Доменные сервисы
│   ├── application/         # Прикладной слой
│   │   └── usecase/         # Use Cases (бизнес-сценарии)
│   ├── infrastructure/      # Инфраструктурный слой
│   │   ├── persistence/     # Реализации репозиториев
│   │   └── database/        # Конфигурация БД
│   └── presentation/        # Слой представления
│       ├── tui/             # TUI компоненты
│       └── styles/          # Стили и цвета
├── go.mod
├── go.sum
└── README.md
```

### Слои архитектуры

#### 1. Domain Layer (Доменный слой)
- **Entity**: `MoodEntry` - основная сущность с валидацией
- **Value Objects**: `MoodLevel` - уровень настроения (0-10)
- **Repository Interfaces**: Абстракции для работы с данными

#### 2. Application Layer (Прикладной слой)
- **Use Cases**: `MoodService` - бизнес-сценарии (запись, обновление, статистика)

#### 3. Infrastructure Layer (Инфраструктурный слой)
- **Database**: Конфигурация SQLite с миграциями
- **Repository**: `SQLiteMoodRepository` - реализация репозитория

#### 4. Presentation Layer (Слой представления)
- **TUI**: Интерактивные экраны на Bubble Tea
- **Styles**: Пастельная цветовая схема

## База данных

### Схема

```sql
CREATE TABLE mood_entries (
    id TEXT PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    level INTEGER NOT NULL CHECK(level >= 0 AND level <= 10),
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX idx_mood_entries_date ON mood_entries(date DESC);
CREATE INDEX idx_mood_entries_date_range ON mood_entries(date);
```

### Расположение

База данных создаётся автоматически в:
```
~/.mood-diary/mood_diary.db
```

### Особенности

- **WAL Mode** - Write-Ahead Logging для лучшей производительности
- **Unique constraint** на дату - только одна запись в день
- **Автоматические триггеры** для обновления `updated_at`
- **Soft delete** поддержка (в разработке)

## 🎨 Цветовая схема

Приложение использует следующую цветовую палитру для комфортного просмотра:

```go
// Основные цвета
PastelPink      #FFB3BA
PastelPeach     #FFDFBA
PastelYellow    #FFFFBA  
PastelMint      #BAFFC9  
PastelSky       #BAE1FF
PastelLavender  #D4BAFF
PastelRose      #FFBAE8
```

### Градиент настроений

Цвета меняются от **пастельно-красного** (грустно) до **пастельно-голубого** (счастливо):

```
😢 → 😞 → 😔 → 😕 → 😐 → 😶 → 🙂 → 😊 → 😄 → 😁 → 🤩
🔴 → 🟠 → 🟡 → ⚪ → 🔵 → 💙
```


## 🔧 Разработка

### Добавление новых функций

1. **Доменный слой**: Добавьте новые entity или value objects
2. **Репозиторий**: Расширьте интерфейс репозитория
3. **Use Case**: Создайте новый use case в application layer
4. **TUI**: Добавьте новый экран или компонент

### Пример: Добавление тегов к записям

```go
// 1. Domain Entity
type Tag struct {
    ID   uuid.UUID
    Name string
}

// 2. Repository
type TagRepository interface {
    Create(ctx context.Context, tag *Tag) error
    FindByMoodID(ctx context.Context, moodID uuid.UUID) ([]*Tag, error)
}

// 3. Use Case
func (s *MoodService) AddTag(ctx context.Context, moodID uuid.UUID, tag string) error {
    // Implementation
}

// 4. TUI Screen
type TagsScreen struct {
    // Implementation
}
```

## 📊 Статистика и метрики

Приложение вычисляет:

- **Средний уровень** настроения за период
- **Тренд** - линейная регрессия за последние записи
- **Распределение** - количество записей по каждому уровню
- **Динамика** - sparkline график изменений

### Формула тренда

```
Тренд = СреднееВторойПоловины - СреднееПервойПоловины

> 0.5  : Улучшается ↑
< -0.5 : Ухудшается ↓
else   : Стабильно ─
```

## 🤝 Участие в разработке

Приветствуются contributions! 

### Как внести вклад

1. Fork репозиторий
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### Code Style

- Следуйте стандартам Go (`gofmt`, `golint`)
- Документируйте публичные функции
- Придерживайтесь Clean Architecture

## 📝 TODO

- [ ] Экспорт данных (CSV, JSON)
- [ ] Импорт данных
- [ ] Теги и категории
- [ ] Поиск по заметкам
- [ ] Более сложная визуализация (календарь)
- [ ] Напоминания о записи
- [ ] Backup/Restore
- [ ] Темы оформления
- [ ] Мультиязычность

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE)

## 👨‍💻 Автор

**Ivan Ignatenko**
- GitHub: [@ignavan39](https://github.com/ignavan39)


**Заботьтесь о своём настроении! 🌸**
