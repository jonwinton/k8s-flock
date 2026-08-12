package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/jonwinton/k8s-flock/internal/context"
	"github.com/jonwinton/k8s-flock/pkg/types"
)

// --- CommandInputModel tests ---

func TestTUI_CommandInput_InitialView(t *testing.T) {
	m := NewCommandInputModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View()
	if !strings.Contains(view, "Command Mode") {
		t.Errorf("expected initial view to contain 'Command Mode', got:\n%s", view)
	}
}

func TestTUI_CommandInput_TypeAndEnter(t *testing.T) {
	m := NewCommandInputModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for _, r := range "pods" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.IsDone() {
		t.Fatal("expected model to be done after pressing Enter")
	}

	cmd, ok := m.GetCommand()
	if !ok {
		t.Fatal("expected GetCommand to return ok=true")
	}
	if !cmd.Valid {
		t.Fatalf("expected command to be valid, got error: %s", cmd.Error)
	}
	if cmd.Resource != "pods" {
		t.Errorf("expected resource 'pods', got %q", cmd.Resource)
	}
}

func TestTUI_CommandInput_EscCancels(t *testing.T) {
	m := NewCommandInputModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !m.IsDone() {
		t.Fatal("expected model to be done after pressing Esc")
	}
}

// --- DeleteViewModel tests ---

func TestTUI_DeleteView_ShowsResource(t *testing.T) {
	m := NewDeleteViewModel()
	m.SetResource(types.SelectedResource{
		Context:      "prod",
		ResourceType: "pods",
		Namespace:    "default",
		Name:         "nginx-abc123",
	})
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "nginx-abc123") {
		t.Errorf("expected view to contain resource name 'nginx-abc123', got:\n%s", view)
	}
	if !strings.Contains(view, "pods") {
		t.Errorf("expected view to contain resource type 'pods', got:\n%s", view)
	}
}

func TestTUI_DeleteView_ToggleForce(t *testing.T) {
	m := NewDeleteViewModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if m.forceDelete {
		t.Fatal("expected forceDelete to be false initially")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if !m.forceDelete {
		t.Fatal("expected forceDelete to be true after pressing 'f'")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if m.forceDelete {
		t.Fatal("expected forceDelete to be false after pressing 'f' again")
	}
}

func TestTUI_DeleteView_EnterConfirms(t *testing.T) {
	m := NewDeleteViewModel()
	m.SetResource(types.SelectedResource{
		Context:      "prod",
		ResourceType: "pods",
		Name:         "nginx-abc123",
	})
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command after pressing Enter")
	}

	msg := cmd()
	if _, ok := msg.(deleteViewConfirmMsg); !ok {
		t.Fatalf("expected deleteViewConfirmMsg, got %T", msg)
	}
}

func TestTUI_DeleteView_EscCancels(t *testing.T) {
	m := NewDeleteViewModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a command after pressing Esc")
	}

	msg := cmd()
	if _, ok := msg.(deleteViewCancelMsg); !ok {
		t.Fatalf("expected deleteViewCancelMsg, got %T", msg)
	}
}

// --- YAMLViewModel tests ---

func TestTUI_YAMLView_Scrolling(t *testing.T) {
	m := NewYAMLViewModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})

	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, "line: "+strings.Repeat("x", i))
	}
	m.SetContent(strings.Join(lines, "\n"))

	if m.scrollOffset != 0 {
		t.Fatalf("expected initial scrollOffset=0, got %d", m.scrollOffset)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.scrollOffset != 1 {
		t.Errorf("expected scrollOffset=1 after down, got %d", m.scrollOffset)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset=0 after up, got %d", m.scrollOffset)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if m.scrollOffset == 0 {
		t.Error("expected scrollOffset > 0 after pgdown")
	}

	saved := m.scrollOffset
	m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if m.scrollOffset >= saved {
		t.Error("expected scrollOffset to decrease after pgup")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	endOffset := m.scrollOffset
	if endOffset == 0 {
		t.Error("expected scrollOffset > 0 after end")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyHome})
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset=0 after home, got %d", m.scrollOffset)
	}
}

func TestTUI_YAMLView_ToggleManagedFields(t *testing.T) {
	m := NewYAMLViewModel()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	yamlContent := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test\nmanagedFields:\n- manager: kubectl\nstatus:\n  phase: Running"
	m.SetContent(yamlContent)

	if strings.Contains(m.content, "managedFields:") {
		t.Error("expected managed fields to be hidden by default")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if !strings.Contains(m.content, "managedFields:") {
		t.Error("expected managed fields to be visible after pressing 'm'")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if strings.Contains(m.content, "managedFields:") {
		t.Error("expected managed fields to be hidden after pressing 'm' again")
	}
}

// --- ContextSelectorModel tests ---

func newTestContextManager(names ...string) *context.Manager {
	mgr := context.NewManager()
	ctxs := make([]types.KubeContext, len(names))
	for i, name := range names {
		ctxs[i] = types.KubeContext{Name: name, Active: i == 0}
	}
	mgr.SetAvailableContexts(ctxs)
	mgr.SetSelectedContexts([]string{names[0]})
	return mgr
}

func TestTUI_ContextSelector_Navigation(t *testing.T) {
	mgr := newTestContextManager("ctx-a", "ctx-b", "ctx-c")
	m := NewContextSelectorModel(mgr)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if m.cursor != 0 {
		t.Fatalf("expected initial cursor=0, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after down, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("expected cursor=2 after second down, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("expected cursor=2 (clamped) after third down, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after up, got %d", m.cursor)
	}
}

func TestTUI_ContextSelector_ToggleSelection(t *testing.T) {
	mgr := newTestContextManager("ctx-a", "ctx-b")
	m := NewContextSelectorModel(mgr)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	selected := mgr.GetSelectedContexts()
	if len(selected) != 1 || selected[0] != "ctx-a" {
		t.Fatalf("expected only 'ctx-a' selected initially, got %v", selected)
	}

	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	selected = mgr.GetSelectedContexts()
	hasA := false
	for _, s := range selected {
		if s == "ctx-a" {
			hasA = true
		}
	}
	if hasA {
		t.Error("expected 'ctx-a' to be deselected after toggle")
	}
}

func TestTUI_ContextSelector_SelectAll(t *testing.T) {
	mgr := newTestContextManager("ctx-a", "ctx-b", "ctx-c")
	m := NewContextSelectorModel(mgr)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	selected := mgr.GetSelectedContexts()
	if len(selected) != 3 {
		t.Errorf("expected 3 selected after 'a', got %d: %v", len(selected), selected)
	}
}

func TestTUI_ContextSelector_SelectNone(t *testing.T) {
	mgr := newTestContextManager("ctx-a", "ctx-b", "ctx-c")
	m := NewContextSelectorModel(mgr)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	selected := mgr.GetSelectedContexts()
	if len(selected) != 0 {
		t.Errorf("expected 0 selected after 'n', got %d: %v", len(selected), selected)
	}
}

func TestTUI_ContextSelector_Filter(t *testing.T) {
	mgr := newTestContextManager("prod-west", "prod-east", "staging", "dev")
	m := NewContextSelectorModel(mgr)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.filterActive {
		t.Fatal("expected filterActive=true after pressing '/'")
	}

	for _, r := range "prod" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if m.filterText != "prod" {
		t.Errorf("expected filterText='prod', got %q", m.filterText)
	}

	filtered := m.filteredContexts()
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered contexts for 'prod', got %d", len(filtered))
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.filterActive {
		t.Fatal("expected filterActive=false after Enter")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	selected := mgr.GetSelectedContexts()
	hasProdWest := false
	hasProdEast := false
	for _, s := range selected {
		if s == "prod-west" {
			hasProdWest = true
		}
		if s == "prod-east" {
			hasProdEast = true
		}
	}
	if !hasProdWest || !hasProdEast {
		t.Errorf("expected both prod contexts selected, got %v", selected)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.filterText != "" {
		t.Error("expected filter cleared after Esc")
	}
}

// --- teatest integration test with a wrapper ---

type commandInputWrapper struct {
	inner *CommandInputModel
}

func (w commandInputWrapper) Init() tea.Cmd { return nil }

func (w commandInputWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		return w, tea.Quit
	}
	w.inner.Update(msg)
	return w, nil
}

func (w commandInputWrapper) View() string { return w.inner.View() }

func TestTUI_Teatest_CommandInputWrapper(t *testing.T) {
	inner := NewCommandInputModel()
	wrapper := commandInputWrapper{inner: inner}

	tm := teatest.NewTestModel(t, wrapper, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Command Mode")
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}
