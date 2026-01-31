package input

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

type State struct {
	Held    map[string]bool
	Pressed map[string]bool
}

type Manager struct {
	hold    float64
	keys    map[string]float64
	pressed map[string]bool
}

func New(holdDuration float64) *Manager {
	if holdDuration <= 0 {
		holdDuration = 0.15
	}
	return &Manager{
		hold:    holdDuration,
		keys:    map[string]float64{},
		pressed: map[string]bool{},
	}
}

func (m *Manager) HandleEvent(ev tcell.Event) {
	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return
	}
	id := KeyID(key)
	if id == "" {
		return
	}
	if m.keys[id] <= 0 {
		m.pressed[id] = true
	}
	m.keys[id] = m.hold
}

func (m *Manager) Step(dt float64) State {
	for k := range m.keys {
		m.keys[k] -= dt
		if m.keys[k] <= 0 {
			delete(m.keys, k)
		}
	}

	held := make(map[string]bool, len(m.keys))
	for k := range m.keys {
		held[k] = true
	}

	pressed := m.pressed
	m.pressed = map[string]bool{}

	return State{Held: held, Pressed: pressed}
}

func KeyID(ev *tcell.EventKey) string {
	if ev == nil {
		return ""
	}
	if ev.Key() == tcell.KeyRune {
		return strings.ToLower(ev.Str())
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		return "key:esc"
	case tcell.KeyCtrlC:
		return "key:ctrl+c"
	default:
		return "key:" + strings.ToLower(ev.Name())
	}
}
