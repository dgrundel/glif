package input

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

type State struct {
	Held    map[Key]bool
	Pressed map[Key]bool
}

// Action is a game-defined action identifier.
type Action string

// Key is a key identifier (e.g. "w", "key:esc").
type Key string

// ActionMap maps actions to keys.
type ActionMap map[Action]Key

type Manager struct {
	hold    float64
	keys    map[Key]float64
	pressed map[Key]bool
}

func New(holdDuration float64) *Manager {
	if holdDuration <= 0 {
		holdDuration = 0.15
	}
	return &Manager{
		hold:    holdDuration,
		keys:    map[Key]float64{},
		pressed: map[Key]bool{},
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

	held := make(map[Key]bool, len(m.keys))
	for k := range m.keys {
		held[k] = true
	}

	pressed := m.pressed
	m.pressed = map[Key]bool{}

	return State{Held: held, Pressed: pressed}
}

func KeyID(ev *tcell.EventKey) Key {
	if ev == nil {
		return ""
	}
	if ev.Key() == tcell.KeyRune {
		return Key(strings.ToLower(ev.Str()))
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		return Key("key:esc")
	case tcell.KeyCtrlC:
		return Key("key:ctrl+c")
	default:
		return Key("key:" + strings.ToLower(ev.Name()))
	}
}
