.PHONY: build run clean test coverage install help

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=mood-diary
BINARY_PATH=./cmd/mood-diary

LDFLAGS=-ldflags "-s -w"

BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

help:
	@echo "$(BLUE)Mood Diary - Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

build:
	@echo "$(BLUE)Сборка приложения...$(NC)"
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)
	@echo "$(GREEN)✓ Сборка завершена: $(BINARY_NAME)$(NC)"

build-release:
	@echo "$(BLUE)Сборка release версии...$(NC)"
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BINARY_NAME) $(BINARY_PATH)
	@echo "$(GREEN)✓ Release сборка завершена$(NC)"

run:
	@echo "$(BLUE)Запуск приложения...$(NC)"
	$(GOCMD) run $(BINARY_PATH)

clean:
	@echo "$(BLUE)Очистка...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Очистка завершена$(NC)"

test:
	@echo "$(BLUE)Запуск тестов...$(NC)"
	$(GOTEST) -v ./...

test-coverage:
	@echo "$(BLUE)Запуск тестов с покрытием...$(NC)"
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Отчёт о покрытии создан: coverage.html$(NC)"

test-race:
	@echo "$(BLUE)Проверка гонок данных...$(NC)"
	$(GOTEST) -race ./...

deps:
	@echo "$(BLUE)Загрузка зависимостей...$(NC)"
	$(GOMOD) download
	@echo "$(GREEN)✓ Зависимости загружены$(NC)"

deps-update:
	@echo "$(BLUE)Обновление зависимостей...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Зависимости обновлены$(NC)"

tidy:
	@echo "$(BLUE)Очистка go.mod...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✓ go.mod очищен$(NC)"

fmt:
	@echo "$(BLUE)Форматирование кода...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✓ Код отформатирован$(NC)"

vet:
	@echo "$(BLUE)Проверка кода...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)✓ Проверка завершена$(NC)"

lint:
	@echo "$(BLUE)Запуск линтера...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✓ Линтинг завершён$(NC)"; \
	else \
		echo "$(RED)✗ golangci-lint не установлен$(NC)"; \
		echo "  Установите: https://golangci-lint.run/usage/install/"; \
	fi

install: build
	@echo "$(BLUE)Установка приложения...$(NC)"
	@if [ -w /usr/local/bin ]; then \
		cp $(BINARY_NAME) /usr/local/bin/; \
		echo "$(GREEN)✓ Приложение установлено в /usr/local/bin/$(BINARY_NAME)$(NC)"; \
	else \
		echo "$(RED)✗ Нет прав на запись в /usr/local/bin$(NC)"; \
		echo "  Выполните: sudo make install"; \
	fi

uninstall:
	@echo "$(BLUE)Удаление приложения...$(NC)"
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/$(BINARY_NAME); \
		echo "$(GREEN)✓ Приложение удалено$(NC)"; \
	else \
		echo "$(RED)✗ Нет прав на запись в /usr/local/bin$(NC)"; \
		echo "  Выполните: sudo make uninstall"; \
	fi

	@echo "$(BLUE)Очистка базы данных...$(NC)"
	rm -f ~/.mood-diary/*.db*
	@echo "$(GREEN)✓ База данных очищена$(NC)"

all: clean deps build test
	@echo "$(GREEN)✓ Полная сборка завершена$(NC)"

dev: fmt vet run

check: fmt vet lint test
	@echo "$(GREEN)✓ Все проверки пройдены$(NC)"