package forms

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Мок-шаг для тестирования
type mockStep struct {
	enterCalled    bool
	exitCalled     bool
	validated      bool
	validationErr  error
	renderCalled   bool
	updateCalled   bool
}

func (m *mockStep) Render(w, h int) string {
	m.renderCalled = true
	return "mock step"
}

func (m *mockStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	m.updateCalled = true
	return m, nil
}

func (m *mockStep) Validate() error {
	m.validated = true
	return m.validationErr
}

func (m *mockStep) OnEnter() tea.Cmd {
	m.enterCalled = true
	return nil
}

func (m *mockStep) OnExit() tea.Cmd {
	m.exitCalled = true
	return nil
}

func TestWizard_Creation(t *testing.T) {
	step1 := &mockStep{}
	step2 := &mockStep{}

	wizard := NewWizard([]Step{step1, step2})

	if wizard == nil {
		t.Error("Wizard should not be nil")
	}

	if wizard.currentStep != 0 {
		t.Errorf("Initial step should be 0, got %d", wizard.currentStep)
	}

	if wizard.IsComplete() {
		t.Error("Wizard should not be complete initially")
	}

	if wizard.IsCancelled() {
		t.Error("Wizard should not be cancelled initially")
	}
}

func TestWizard_Navigation(t *testing.T) {
	step1 := &mockStep{}
	step2 := &mockStep{}

	wizard := NewWizard([]Step{step1, step2})

	// Переход на следующий шаг
	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !step1.validated {
		t.Error("Step 1 should be validated")
	}

	if !step1.exitCalled {
		t.Error("Step 1 OnExit should be called")
	}

	if !step2.enterCalled {
		t.Error("Step 2 OnEnter should be called")
	}

	if wizard.currentStep != 1 {
		t.Errorf("Current step should be 1, got %d", wizard.currentStep)
	}
}

func TestWizard_BackNavigation(t *testing.T) {
	step1 := &mockStep{}
	step2 := &mockStep{}

	wizard := NewWizard([]Step{step1, step2})

	// Переход вперед
	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Сброс флагов
	step1.enterCalled = false
	step2.exitCalled = false

	// Переход назад
	wizard.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	if !step2.exitCalled {
		t.Error("Step 2 OnExit should be called")
	}

	if !step1.enterCalled {
		t.Error("Step 1 OnEnter should be called")
	}

	if wizard.currentStep != 0 {
		t.Errorf("Current step should be 0, got %d", wizard.currentStep)
	}
}

func TestWizard_Completion(t *testing.T) {
	step1 := &mockStep{}
	wizard := NewWizard([]Step{step1})

	// Завершаем единственный шаг
	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})

	if !wizard.IsComplete() {
		t.Error("Wizard should be complete after last step")
	}

	if !step1.validated {
		t.Error("Step should be validated before completion")
	}

	if !step1.exitCalled {
		t.Error("Step OnExit should be called on completion")
	}
}

func TestWizard_Cancellation(t *testing.T) {
	step1 := &mockStep{}
	wizard := NewWizard([]Step{step1})

	wizard.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !wizard.IsCancelled() {
		t.Error("Wizard should be cancelled on Esc")
	}

	if wizard.IsComplete() {
		t.Error("Wizard should not be complete if cancelled")
	}
}

func TestWizard_ValidationError(t *testing.T) {
	step1 := &mockStep{
		validationErr: &validationError{message: "test error"},
	}
	step2 := &mockStep{}

	wizard := NewWizard([]Step{step1, step2})

	// Попытка перехода с ошибкой валидации
	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})

	if wizard.currentStep != 0 {
		t.Error("Should stay on current step if validation fails")
	}

	if wizard.errorMessage == "" {
		t.Error("Error message should be set")
	}

	if step2.enterCalled {
		t.Error("Next step should not be entered if validation fails")
	}
}

func TestWizard_Progress(t *testing.T) {
	step1 := &mockStep{}
	step2 := &mockStep{}
	step3 := &mockStep{}

	wizard := NewWizard([]Step{step1, step2, step3})

	current, total := wizard.Progress()
	if current != 1 || total != 3 {
		t.Errorf("Expected progress 1/3, got %d/%d", current, total)
	}

	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})
	current, total = wizard.Progress()
	if current != 2 || total != 3 {
		t.Errorf("Expected progress 2/3, got %d/%d", current, total)
	}
}

func TestWizard_Reset(t *testing.T) {
	step1 := &mockStep{}
	step2 := &mockStep{}

	wizard := NewWizard([]Step{step1, step2})

	// Переход и отмена
	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})
	wizard.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Сброс
	wizard.Reset()

	if wizard.currentStep != 0 {
		t.Error("Current step should be reset to 0")
	}

	if wizard.complete {
		t.Error("Complete flag should be reset")
	}

	if wizard.cancelled {
		t.Error("Cancelled flag should be reset")
	}

	if wizard.errorMessage != "" {
		t.Error("Error message should be cleared")
	}
}

func TestWizard_View(t *testing.T) {
	step1 := &mockStep{}
	wizard := NewWizard([]Step{step1})

	view := wizard.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !step1.renderCalled {
		t.Error("Step Render should be called")
	}
}

func TestWizard_SetSize(t *testing.T) {
	step1 := &mockStep{}
	wizard := NewWizard([]Step{step1})

	wizard.SetSize(100, 50)

	if wizard.width != 100 {
		t.Errorf("Width should be 100, got %d", wizard.width)
	}

	if wizard.height != 50 {
		t.Errorf("Height should be 50, got %d", wizard.height)
	}
}

// Вспомогательная структура для тестирования ошибок
type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
