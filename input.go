package main

import rl "github.com/gen2brain/raylib-go/raylib"

// Key is a raylib key code.
type Key = int32

// handleInput reads the keyboard once per frame: each kid's four direction
// keys, button, view and machine keys (see toddlerKeys), and F2 to pause.
func (g *Game) handleInput() {
	if inpTapped(rl.KeyF2) {
		g.paused = !g.paused
	}
	for _, p := range g.players {
		if g.single {
			g.singleInput(p)
		} else {
			g.toddlerInput(p)
		}
	}
}

func axis(pos, neg []Key) float64 {
	v := 0.0
	if keyDown(pos...) {
		v++
	}
	if keyDown(neg...) {
		v--
	}
	return v
}

func keyDown(keys ...Key) bool {
	for _, k := range keys {
		if rl.IsKeyDown(k) {
			return true
		}
	}
	return false
}

// --- simple edge-trigger tracking for taps (poll each key once per frame) ---
var prevKeys = map[Key]bool{}

func inpTapped(k Key) bool {
	down := rl.IsKeyDown(k)
	was := prevKeys[k]
	prevKeys[k] = down
	return down && !was
}
