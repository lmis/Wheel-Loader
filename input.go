package main

import "github.com/hajimehoshi/ebiten/v2"

// handleInput reads the keyboard once per frame: each kid's four direction
// keys, button and machine keys (see toddlerKeys), and F2 to pause.
func (g *Game) handleInput() {
	if inpTapped(ebiten.KeyF2) {
		g.paused = !g.paused
	}
	for _, p := range g.players {
		g.toddlerInput(p)
	}
}

func axis(pos, neg []ebiten.Key) float64 {
	v := 0.0
	if keyDown(pos...) {
		v++
	}
	if keyDown(neg...) {
		v--
	}
	return v
}

func keyDown(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

// --- simple edge-trigger tracking for taps (poll each key once per frame) ---
var prevKeys = map[ebiten.Key]bool{}

func inpTapped(k ebiten.Key) bool {
	down := ebiten.IsKeyPressed(k)
	was := prevKeys[k]
	prevKeys[k] = down
	return down && !was
}
