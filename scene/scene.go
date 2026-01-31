package scene

import "github.com/dgrundel/glif/render"

type Entity interface {
	Update(dt float64)
	Draw(r *render.Renderer)
}

type Resizer interface {
	Resize(w, h int)
}

type Scene struct {
	Entities []Entity
	OnResize func(w, h int)
}

func New() *Scene {
	return &Scene{Entities: []Entity{}}
}

func (s *Scene) Add(e Entity) {
	s.Entities = append(s.Entities, e)
}

func (s *Scene) Update(dt float64) {
	for _, e := range s.Entities {
		e.Update(dt)
	}
}

func (s *Scene) Draw(r *render.Renderer) {
	for _, e := range s.Entities {
		e.Draw(r)
	}
}

func (s *Scene) Resize(w, h int) {
	if s.OnResize != nil {
		s.OnResize(w, h)
	}
	for _, e := range s.Entities {
		if r, ok := e.(Resizer); ok {
			r.Resize(w, h)
		}
	}
}
