# 🏗️ Architecture Documentation

## Обзор архитектуры

Mood Diary построен на принципах **Clean Architecture** и **Domain-Driven Design (DDD)**, обеспечивая:
- Независимость от фреймворков
- Тестируемость
- Независимость от UI
- Независимость от базы данных
- Независимость от внешних сервисов

## Слои архитектуры

```
┌─────────────────────────────────────────────────────────┐
│                    Presentation Layer                    │
│                  (TUI, Styles, Views)                    │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                   Application Layer                      │
│                  (Use Cases, Services)                   │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                     Domain Layer                         │
│          (Entities, Value Objects, Interfaces)           │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                     │
│            (Database, Repositories, External)            │
└─────────────────────────────────────────────────────────┘
```

### Правила зависимостей

1. **Внешние слои зависят от внутренних**
   - Presentation → Application → Domain
   - Infrastructure → Domain (только интерфейсы)

2. **Внутренние слои не знают о внешних**
   - Domain не знает о TUI
   - Application не знает о SQLite

3. **Зависимости инвертируются через интерфейсы**
   - Domain определяет `MoodRepository` interface
   - Infrastructure реализует `SQLiteMoodRepository`

---

## Domain Layer (Доменный слой)

### Ответственность
- Бизнес-логика приложения
- Правила валидации
- Доменные модели

### Компоненты

#### Entity: MoodEntry
```go
type MoodEntry struct {
    ID        uuid.UUID
    Date      time.Time
    Level     MoodLevel
    Note      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Бизнес-правила:**
- ID должен быть валидным UUID
- Date нормализуется до начала дня
- Level должен быть от 0 до 10
- Note может быть пустым
- Даты автоматически управляются

**Методы:**
- `NewMoodEntry()` - создание с валидацией
- `Update()` - обновление с валидацией
- `IsToday()` - проверка актуальности
- `DaysAgo()` - вычисление возраста записи

#### Value Object: MoodLevel
```go
type MoodLevel int

const (
    MinMoodLevel MoodLevel = 0
    MaxMoodLevel MoodLevel = 10
)
```

**Характеристики:**
- Immutable (неизменяемый)
- Самовалидирующийся
- Имеет rich behavior (String(), Emoji())
- Сравнивается по значению

**Методы:**
- `NewMoodLevel()` - создание с валидацией
- `Int()` - получение числового значения
- `String()` - текстовое описание
- `Emoji()` - эмоджи представление

#### Repository Interface
```go
type MoodRepository interface {
    Create(ctx context.Context, entry *MoodEntry) error
    Update(ctx context.Context, entry *MoodEntry) error
    Delete(ctx context.Context, id uuid.UUID) error
    FindByID(ctx context.Context, id uuid.UUID) (*MoodEntry, error)
    FindByDate(ctx context.Context, date time.Time) (*MoodEntry, error)
    FindByDateRange(ctx context.Context, start, end time.Time) ([]*MoodEntry, error)
    FindRecent(ctx context.Context, limit int) ([]*MoodEntry, error)
    FindAll(ctx context.Context) ([]*MoodEntry, error)
    GetStatistics(ctx context.Context, start, end time.Time) (*MoodStatistics, error)
}
```

**Принципы:**
- Определяется в domain слое
- Реализуется в infrastructure слое
- Контракт между слоями
- Не зависит от конкретной БД

---

## Application Layer (Прикладной слой)

### Ответственность
- Оркестрация бизнес-логики
- Координация между domain и infrastructure
- Реализация use cases

### Компоненты

#### MoodService
```go
type MoodService struct {
    repo repository.MoodRepository
}
```

**Use Cases:**

1. **RecordMood** - Записать новое настроение
   ```go
   func (s *MoodService) RecordMood(
       ctx context.Context, 
       level int, 
       note string, 
       date *time.Time
   ) error
   ```
   - Валидирует уровень настроения
   - Проверяет на дубликаты
   - Создает новую запись
   - Сохраняет в репозиторий

2. **UpdateMood** - Обновить существующее настроение
   ```go
   func (s *MoodService) UpdateMood(
       ctx context.Context, 
       date time.Time, 
       level int, 
       note string
   ) error
   ```
   - Находит существующую запись
   - Валидирует новые данные
   - Обновляет запись
   - Сохраняет изменения

3. **GetStatistics** - Получить статистику
   ```go
   func (s *MoodService) GetStatistics(
       ctx context.Context, 
       period Period
   ) (*repository.MoodStatistics, error)
   ```
   - Определяет диапазон дат
   - Запрашивает статистику из репозитория
   - Возвращает агрегированные данные

#### Period Value Object
```go
type Period int

const (
    PeriodWeek Period = iota
    PeriodMonth
    PeriodQuarter
    PeriodYear
    PeriodAll
)
```

**Методы:**
- `DateRange()` - вычисляет start и end даты
- `String()` - локализованное название

---

## Infrastructure Layer (Инфраструктурный слой)

### Ответственность
- Взаимодействие с внешними системами
- Реализация интерфейсов репозиториев
- Конфигурация базы данных

### Компоненты

#### Database
```go
type Database struct {
    db *sql.DB
}
```

**Функции:**
- `New(dbPath string)` - создание подключения
- `Migrate()` - применение миграций
- `Close()` - закрытие подключения
- `GetDefaultDBPath()` - путь по умолчанию

**Конфигурация:**
- WAL mode для лучшей производительности
- Single connection (SQLite limitation)
- Автоматическое создание директории
- Проверка подключения при старте

#### SQLiteMoodRepository
```go
type SQLiteMoodRepository struct {
    db *sql.DB
}
```

**Реализация MoodRepository:**
- Преобразование domain entities в SQL
- Обработка ошибок базы данных
- Транзакционность операций
- Оптимизированные запросы с индексами

**Особенности:**
- Использует prepared statements
- Context для отмены операций
- Proper error wrapping
- Efficient batch operations

#### Schema
```sql
CREATE TABLE mood_entries (
    id TEXT PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    level INTEGER NOT NULL CHECK(level >= 0 AND level <= 10),
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_mood_entries_date` - быстрый поиск по дате
- `idx_mood_entries_date_range` - эффективные range queries

**Триггеры:**
- `update_mood_entries_updated_at` - автообновление timestamp

**Ограничения:**
- PRIMARY KEY на id
- UNIQUE на date (одна запись в день)
- CHECK constraint на level (0-10)
- NOT NULL на обязательных полях

---

## Presentation Layer (Слой представления)

### Ответственность
- Отображение данных пользователю
- Обработка пользовательского ввода
- Навигация между экранами

### Архитектура TUI (Bubble Tea)

#### The Elm Architecture
```
┌──────────┐
│  Model   │ ◄─┐
└────┬─────┘   │
     │         │
     ▼         │
┌──────────┐   │
│   View   │   │
└────┬─────┘   │
     │         │
     ▼         │
┌──────────┐   │
│  Update  │ ──┘
└──────────┘
```

1. **Model** - состояние приложения
2. **View** - рендеринг UI
3. **Update** - обработка событий и обновление состояния

#### Главная модель
```go
type Model struct {
    ctx           context.Context
    service       *usecase.MoodService
    currentScreen Screen
    
    // Sub-models
    menuModel    *MenuModel
    recordModel  *RecordModel
    statsModel   *StatsModel
    historyModel *HistoryModel
    editModel    *EditModel
}
```

**Навигация:**
- `NavigateMsg` - сообщение для смены экрана
- `initCurrentScreen()` - инициализация при переходе
- Делегирование обновлений к sub-models

#### Экраны (Sub-models)

**MenuModel** - главное меню
```go
type MenuModel struct {
    choices  []string
    cursor   int
    selected int
}
```

**RecordModel** - запись настроения
```go
type RecordModel struct {
    service   *usecase.MoodService
    moodLevel int
    noteInput textinput.Model
    step      int // Multi-step wizard
    success   bool
    errorMsg  string
}
```

**StatsModel** - статистика
```go
type StatsModel struct {
    service  *usecase.MoodService
    period   usecase.Period
    stats    *repository.MoodStatistics
    entries  []*entity.MoodEntry
    loading  bool
}
```

**HistoryModel** - история записей
```go
type HistoryModel struct {
    service  *usecase.MoodService
    entries  []*entity.MoodEntry
    cursor   int
    loading  bool
}
```

### Стили (Lipgloss)

#### Цветовая палитра
```go
// Пастельные цвета
PastelPink      = "#FFB3BA"
PastelPeach     = "#FFDFBA"
PastelYellow    = "#FFFFBA"
PastelMint      = "#BAFFC9"
PastelSky       = "#BAE1FF"
PastelLavender  = "#D4BAFF"
PastelRose      = "#FFBAE8"
```

#### Компоненты стилей
- `TitleStyle` - заголовки
- `BoxStyle` - контейнеры
- `ButtonStyle` - кнопки
- `ListItemStyle` - элементы списка
- `InputStyle` - поля ввода
- `MoodStyle(level)` - динамические цвета по настроению

---

## Потоки данных

### Запись нового настроения

```
User Input (TUI)
    ↓
RecordModel.Update()
    ↓
RecordModel.saveMood() → tea.Cmd
    ↓
MoodService.RecordMood(ctx, level, note, date)
    ↓
Validate & Create MoodEntry
    ↓
SQLiteMoodRepository.Create(ctx, entry)
    ↓
SQL INSERT with prepared statement
    ↓
Return SavedMsg or ErrorMsg
    ↓
RecordModel.Update() handles message
    ↓
Navigate to Menu
```

### Загрузка статистики

```
Navigate to Stats Screen
    ↓
StatsModel.Init() → loadStats()
    ↓
MoodService.GetStatistics(ctx, period)
    ↓
Period.DateRange() → start, end
    ↓
SQLiteMoodRepository.GetStatistics(ctx, start, end)
    ↓
Aggregate queries (COUNT, AVG, MIN, MAX)
    ↓
Distribution query (GROUP BY level)
    ↓
Calculate trend
    ↓
Return StatsLoadedMsg
    ↓
StatsModel.Update() handles message
    ↓
StatsModel.View() renders UI
```

---

## Паттерны проектирования

### Repository Pattern
- Абстракция доступа к данным
- Domain определяет интерфейс
- Infrastructure реализует
- Легко заменить SQLite на другую БД

### Dependency Injection
```go
// Создание зависимостей
db := database.New(dbPath)
repo := persistence.NewSQLiteMoodRepository(db.DB())
service := usecase.NewMoodService(repo)

// Внедрение в TUI
model := tui.NewModel(ctx, service)
```

### Command Pattern (Bubble Tea)
```go
type tea.Cmd func() tea.Msg

// Асинхронная операция
func loadData() tea.Cmd {
    return func() tea.Msg {
        data := fetchFromDB()
        return DataLoadedMsg{data}
    }
}
```

### Value Object Pattern
```go
type MoodLevel int

func NewMoodLevel(level int) (MoodLevel, error) {
    if level < 0 || level > 10 {
        return 0, ErrInvalidMoodLevel
    }
    return MoodLevel(level), nil
}
```

### Factory Pattern
```go
func NewMoodEntry(level MoodLevel, note string, date time.Time) (*MoodEntry, error) {
    // Validation and initialization
    return &MoodEntry{...}, nil
}
```

---

## Производительность

### Оптимизации базы данных

1. **Индексы**
   - Индекс на date для быстрого поиска
   - Composite индексы для range queries

2. **WAL Mode**
   - Concurrent reads
   - Better performance
   - Меньше блокировок

3. **Prepared Statements**
   - Защита от SQL injection
   - Лучшая производительность
   - Кеширование запросов

4. **Connection Pooling**
   ```go
   db.SetMaxOpenConns(1)  // SQLite single-writer
   db.SetMaxIdleConns(1)
   ```

### Оптимизации TUI

1. **Lazy Loading**
   - Загрузка данных только при переходе на экран
   - Кеширование результатов

2. **Pagination**
   - История показывает только последние N записей
   - LIMIT в SQL queries

3. **Efficient Rendering**
   - Минимальные перерисовки
   - Использование Lipgloss caching

---

## Расширяемость

### Добавление новых функций

**Пример: Добавление тегов**

1. **Domain Layer**
```go
// entity/tag.go
type Tag struct {
    ID   uuid.UUID
    Name string
}

// repository/tag_repository.go
type TagRepository interface {
    Create(ctx context.Context, tag *Tag) error
    FindByMoodID(ctx context.Context, moodID uuid.UUID) ([]*Tag, error)
}
```

2. **Infrastructure Layer**
```sql
-- database/schema.go
CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE mood_tags (
    mood_id TEXT REFERENCES mood_entries(id),
    tag_id TEXT REFERENCES tags(id),
    PRIMARY KEY (mood_id, tag_id)
);
```

3. **Application Layer**
```go
// usecase/tag_service.go
type TagService struct {
    tagRepo repository.TagRepository
}

func (s *TagService) AddTagToMood(ctx context.Context, moodID uuid.UUID, tagName string) error {
    // Implementation
}
```

4. **Presentation Layer**
```go
// tui/tags.go
type TagsModel struct {
    // Implementation
}
```

---

## Безопасность

### Валидация входных данных
- Все входные данные валидируются в domain layer
- Type-safe интерфейсы
- Невозможно создать невалидные entity

### SQL Injection Protection
- Prepared statements для всех запросов
- Параметризованные queries
- Нет конкатенации SQL строк

### Error Handling
```go
// Правильная обработка ошибок
if err != nil {
    return fmt.Errorf("failed to create mood entry: %w", err)
}
```

### Data Privacy
- Локальное хранение данных
- Нет сетевых запросов
- Полный контроль пользователя

---

## Будущие улучшения

### Планируемые функции
1. Теги и категории
2. Экспорт/Импорт данных
3. Графики и визуализации
4. Поиск и фильтрация
5. Напоминания
6. Backup/Restore

### Возможные архитектурные изменения
1. Event Sourcing для истории изменений
2. CQRS для разделения read/write моделей
3. Aggregate Root для комплексных операций
4. Domain Events для loose coupling

---

**Архитектура построена для долгосрочного развития и легкого тестирования!**
