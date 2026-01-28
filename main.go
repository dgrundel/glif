package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

const duck = `
>o)
(_>
`

const whale = `
       .
      ":"
    ___:____     |"\/"|
  ,'        '.    \  /
  |  O        \___/  |
~^~^~^~^~^~^~^~^~^~^~^~^~
`

func splitLines(s string) []string {
	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func maxLineLen(lines []string) int {
	max := 0
	for _, line := range lines {
		if len(line) > max {
			max = len(line)
		}
	}
	return max
}

func drawSprite(s tcell.Screen, x, y int, style tcell.Style, lines []string) {
	for row, line := range lines {
		for col, r := range line {
			s.SetContent(x+col, y+row, r, nil, style)
		}
	}
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	var width int
	for text != "" {
		text, width = s.Put(col, row, text, style)
		col += width
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
		if width == 0 {
			// incomplete grapheme at end of string
			break
		}
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			s.Put(col, row, " ", style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.Put(col, y1, string(tcell.RuneHLine), style)
		s.Put(col, y2, string(tcell.RuneHLine), style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.Put(x1, row, string(tcell.RuneVLine), style)
		s.Put(x2, row, string(tcell.RuneVLine), style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.Put(x1, y1, string(tcell.RuneULCorner), style)
		s.Put(x2, y1, string(tcell.RuneURCorner), style)
		s.Put(x1, y2, string(tcell.RuneLLCorner), style)
		s.Put(x2, y2, string(tcell.RuneLRCorner), style)
	}

	drawText(s, x1+1, y1+1, x2-1, y2-1, style, text)
}

func main() {
	defStyle := tcell.StyleDefault.Background(color.Reset).Foreground(color.Reset)
	boxStyle := tcell.StyleDefault.Foreground(color.Black).Background(color.LightGoldenrodYellow)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	// Draw initial boxes
	drawBox(s, 1, 1, 42, 7, boxStyle, "Click and drag to draw a box")
	drawBox(s, 5, 9, 32, 14, boxStyle, "Press C to reset")

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	// Here's how to get the screen size when you need it.
	// xmax, ymax := s.Size()

	// Here's an example of how to inject a keystroke where it will
	// be picked up by a future read of the event queue.  Note that
	// care should be used to avoid blocking writes to the queue if
	// this is done from the same thread that is responsible for reading
	// the queue, or else a single-party deadlock might occur.
	// s.EventQ() <- tcell.NewEventKey(tcell.KeyRune, rune('a'), 0)

	duckLines := splitLines(duck)
	duckW := maxLineLen(duckLines)
	duckH := len(duckLines)
	duckXf := 1.0
	duckY := 2
	duckDX := 1
	duckSpeed := 20.0 // cells per second

	events := s.EventQ()

	ticker := time.NewTicker(33 * time.Millisecond) // ~30 FPS
	defer ticker.Stop()
	lastTick := time.Now()

	// Event loop
	ox, oy := -1, -1
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(lastTick).Seconds()
			lastTick = now
			if dt > 0.1 {
				dt = 0.1
			}

			xmax, ymax := s.Size()
			maxX := xmax - duckW
			maxY := ymax - duckH - 1
			if maxX < 0 {
				maxX = 0
			}
			if maxY < 0 {
				maxY = 0
			}
			if duckY < 0 {
				duckY = 0
			}
			if duckY > maxY {
				duckY = maxY
			}

			duckXf += float64(duckDX) * duckSpeed * dt
			if duckXf <= 0 {
				duckXf = 0
				duckDX = 1
			}
			if duckXf >= float64(maxX) {
				duckXf = float64(maxX)
				duckDX = -1
			}

			duckX := int(duckXf + 0.5)
			s.Clear()
			drawSprite(s, duckX, duckY, defStyle, duckLines)
			s.Show()
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					return
				} else if ev.Key() == tcell.KeyCtrlL {
					s.Sync()
				} else if ev.Str() == "C" || ev.Str() == "c" {
					s.Clear()
				}
			case *tcell.EventMouse:
				x, y := ev.Position()

				switch ev.Buttons() {
				case tcell.Button1, tcell.Button2:
					if ox < 0 {
						ox, oy = x, y // record location when click started
					}

				case tcell.ButtonNone:
					if ox >= 0 {
						label := fmt.Sprintf("%d,%d to %d,%d", ox, oy, x, y)
						drawBox(s, ox, oy, x, y, boxStyle, label)
						ox, oy = -1, -1
					}
				}
			}
		}
	}
}
