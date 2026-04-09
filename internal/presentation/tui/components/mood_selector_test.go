package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMoodSelector_Creation(t *testing.T) {
	selector := NewMoodSelector(5)

	if selector == nil {
		t.Error("Selector should not be nil")
	}

	if selector.Value() != 5 {
		t.Errorf("Initial value should be 5, got %d", selector.Value())
	}

	if selector.min != 0 {
		t.Errorf("Min should be 0, got %d", selector.min)
	}

	if selector.max != 10 {
		t.Errorf("Max should be 10, got %d", selector.max)
	}
}

func TestMoodSelector_Navigation(t *testing.T) {
	selector := NewMoodSelector(5)
	selector.Focus()

	// Увеличение
	selector.Update(tea.KeyMsg{Type: tea.KeyRight})
	if selector.Value() != 6 {
		t.Errorf("Expected 6, got %d", selector.Value())
	}

	// Уменьшение
	selector.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if selector.Value() != 5 {
		t.Errorf("Expected 5, got %d", selector.Value())
	}
}

func TestMoodSelector_KeyboardNavigation(t *testing.T) {
	selector := NewMoodSelector(5)
	selector.Focus()

	// h для уменьшения
	selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if selector.Value() != 4 {
		t.Errorf("Expected 4 after 'h', got %d", selector.Value())
	}

	// l для увеличения
	selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if selector.Value() != 5 {
		t.Errorf("Expected 5 after 'l', got %d", selector.Value())
	}
}

func TestMoodSelector_Bounds(t *testing.T) {
	selector := NewMoodSelector(0)
	selector.Focus()

	// Не должно идти ниже 0
	selector.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if selector.Value() != 0 {
		t.Errorf("Should not go below 0, got %d", selector.Value())
	}

	selector.SetValue(10)

	// Не должно идти выше 10
	selector.Update(tea.KeyMsg{Type: tea.KeyRight})
	if selector.Value() != 10 {
		t.Errorf("Should not go above 10, got %d", selector.Value())
	}
}

func TestMoodSelector_DirectInput(t *testing.T) {
	selector := NewMoodSelector(5)
	selector.Focus()

	// Прямой ввод числа
	selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	if selector.Value() != 7 {
		t.Errorf("Expected 7 after direct input, got %d", selector.Value())
	}

	// Недопустимое число должно игнорироваться
	initialValue := selector.Value()
	selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if selector.Value() != initialValue {
		t.Error("Invalid input should not change value")
	}
}

func TestMoodSelector_Focus(t *testing.T) {
	selector := NewMoodSelector(5)

	// По умолчанию не в фокусе
	if selector.focused {
		t.Error("Selector should not be focused initially")
	}

	// Установка фокуса
	selector.Focus()
	if !selector.focused {
		t.Error("Selector should be focused after Focus()")
	}

	// Снятие фокуса
	selector.Blur()
	if selector.focused {
		t.Error("Selector should not be focused after Blur()")
	}
}

func TestMoodSelector_UpdateWhenNotFocused(t *testing.T) {
	selector := NewMoodSelector(5)
	// Не вызываем Focus()

	initialValue := selector.Value()
	selector.Update(tea.KeyMsg{Type: tea.KeyRight})

	if selector.Value() != initialValue {
		t.Error("Unfocused selector should not respond to input")
	}
}

func TestMoodSelector_OnChange(t *testing.T) {
	selector := NewMoodSelector(5)

	changeCallbackCalled := false
	var callbackValue int

	selector.OnChange(func(value int) {
		changeCallbackCalled = true
		callbackValue = value
	})

	selector.SetValue(7)

	if !changeCallbackCalled {
		t.Error("OnChange callback should be called")
	}

	if callbackValue != 7 {
		t.Errorf("Callback should receive new value 7, got %d", callbackValue)
	}
}

func TestMoodSelector_SetValue(t *testing.T) {
	selector := NewMoodSelector(5)

	// Установка допустимого значения
	selector.SetValue(8)
	if selector.Value() != 8 {
		t.Errorf("Expected 8, got %d", selector.Value())
	}

	// Установка значения ниже минимума
	selector.SetValue(-1)
	if selector.Value() != 8 {
		t.Error("Value should not change when setting below minimum")
	}

	// Установка значения выше максимума
	selector.SetValue(11)
	if selector.Value() != 8 {
		t.Error("Value should not change when setting above maximum")
	}

	// Установка граничных значений
	selector.SetValue(0)
	if selector.Value() != 0 {
		t.Error("Should accept minimum value 0")
	}

	selector.SetValue(10)
	if selector.Value() != 10 {
		t.Error("Should accept maximum value 10")
	}
}

func TestMoodSelector_View(t *testing.T) {
	selector := NewMoodSelector(5)

	view := selector.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Проверяем что view содержит информацию о текущем значении
	// (это базовая проверка, полное тестирование UI сложно в unit-тестах)
	if len(view) < 10 {
		t.Error("View should contain reasonable amount of content")
	}
}
