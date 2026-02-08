package input

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

type State struct {
	Held    map[Key]bool
	Pressed map[Key]bool
	Typed   []rune
}

// Action is a game-defined action identifier.
type Action string

// Key is a key identifier (e.g. "w", "key:esc").
type Key string

// ActionMap maps actions to keys.
type ActionMap map[Action]Key

type ActionState struct {
	Held    map[Action]bool
	Pressed map[Action]bool
}

type Mapper struct {
	Map ActionMap
}

func (m Mapper) MapState(state State) ActionState {
	held := map[Action]bool{}
	pressed := map[Action]bool{}
	for action, key := range m.Map {
		if state.Held[key] {
			held[action] = true
		}
		if state.Pressed[key] {
			pressed[action] = true
		}
	}
	return ActionState{Held: held, Pressed: pressed}
}

type Manager struct {
	hold    float64
	keys    map[Key]float64
	pressed map[Key]bool
	typed   []rune
}

func New(holdDuration float64) *Manager {
	if holdDuration <= 0 {
		holdDuration = 0.15
	}
	return &Manager{
		hold:    holdDuration,
		keys:    map[Key]float64{},
		pressed: map[Key]bool{},
		typed:   nil,
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

	if key.Key() == tcell.KeyRune {
		str := key.Str()
		for _, r := range []rune(str) {
			m.typed = append(m.typed, r)
		}
	}
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
	typed := m.typed
	m.typed = nil

	return State{Held: held, Pressed: pressed, Typed: typed}
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
