package main

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// keyMap is one player's half of the shared keyboard.
type keyMap struct {
	fwd, back, left, right []ebiten.Key
	brake                  []ebiten.Key       // brake; run when on foot
	axes                   [4][2][]ebiten.Key // implement axes A1..A4: [+, -]
	enter, camera          ebiten.Key         // get in/out, first/third person
	zoomIn, zoomOut        []ebiten.Key
}

// Player 1 uses the left half of the keyboard, player 2 the right half.
var (
	p1Keys = keyMap{
		fwd: []ebiten.Key{ebiten.KeyW}, back: []ebiten.Key{ebiten.KeyS},
		left: []ebiten.Key{ebiten.KeyA}, right: []ebiten.Key{ebiten.KeyD},
		brake: []ebiten.Key{ebiten.KeySpace, ebiten.KeyShiftLeft},
		axes: [4][2][]ebiten.Key{
			{{ebiten.KeyR}, {ebiten.KeyF}},
			{{ebiten.KeyT}, {ebiten.KeyG}},
			{{ebiten.KeyZ}, {ebiten.KeyX}},
			{{ebiten.KeyE}, {ebiten.KeyQ}},
		},
		enter: ebiten.KeyC, camera: ebiten.KeyV,
		zoomIn: []ebiten.Key{ebiten.KeyDigit2}, zoomOut: []ebiten.Key{ebiten.KeyDigit1},
	}
	p2Keys = keyMap{
		fwd: []ebiten.Key{ebiten.KeyK, ebiten.KeyArrowUp}, back: []ebiten.Key{ebiten.KeyJ, ebiten.KeyArrowDown},
		left: []ebiten.Key{ebiten.KeyH, ebiten.KeyArrowLeft}, right: []ebiten.Key{ebiten.KeyL, ebiten.KeyArrowRight},
		brake: []ebiten.Key{ebiten.KeyShiftRight},
		axes: [4][2][]ebiten.Key{
			{{ebiten.KeyI}, {ebiten.KeyComma}},
			{{ebiten.KeyO}, {ebiten.KeyPeriod}},
			{{ebiten.KeyP}, {ebiten.KeySemicolon}},
			{{ebiten.KeyU}, {ebiten.KeyY}},
		},
		enter: ebiten.KeyN, camera: ebiten.KeyM,
		zoomIn: []ebiten.Key{ebiten.KeyDigit0}, zoomOut: []ebiten.Key{ebiten.KeyDigit9},
	}
)

// solo returns the map used when one player has the whole keyboard: the
// arrow keys and the mouse wheel join in as alternatives.
func (k keyMap) solo() keyMap {
	k.fwd = append(append([]ebiten.Key{}, k.fwd...), ebiten.KeyArrowUp)
	k.back = append(append([]ebiten.Key{}, k.back...), ebiten.KeyArrowDown)
	k.left = append(append([]ebiten.Key{}, k.left...), ebiten.KeyArrowLeft)
	k.right = append(append([]ebiten.Key{}, k.right...), ebiten.KeyArrowRight)
	k.zoomIn = append(append([]ebiten.Key{}, k.zoomIn...), ebiten.KeyEqual)
	k.zoomOut = append(append([]ebiten.Key{}, k.zoomOut...), ebiten.KeyMinus)
	return k
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

// read samples this player's driving controls from the keyboard.
func (k *keyMap) read() Controls {
	c := Controls{
		Throttle: axis(k.fwd, k.back),
		Steer:    axis(k.right, k.left),
		Brake:    keyDown(k.brake...),
	}
	c.A1 = axis(k.axes[0][0], k.axes[0][1])
	c.A2 = axis(k.axes[1][0], k.axes[1][1])
	c.A3 = axis(k.axes[2][0], k.axes[2][1])
	c.A4 = axis(k.axes[3][0], k.axes[3][1])
	return c
}

func keyName(k ebiten.Key) string {
	switch k {
	case ebiten.KeyComma:
		return ","
	case ebiten.KeyPeriod:
		return "."
	case ebiten.KeySemicolon:
		return ";"
	case ebiten.KeySpace:
		return "Space"
	case ebiten.KeyShiftLeft:
		return "LShift"
	case ebiten.KeyShiftRight:
		return "RShift"
	}
	s := k.String()
	s = strings.TrimPrefix(s, "Digit")
	return s
}

// helpText expands the key tokens in a control-help line for this player:
// {T} throttle, {S} steer, {B} brake, {1}..{4} implement axes, {E} enter,
// {V} camera.
func (k *keyMap) helpText(line string) string {
	pair := func(a, b []ebiten.Key) string { return keyName(a[0]) + " / " + keyName(b[0]) }
	r := strings.NewReplacer(
		"{T}", pair(k.fwd, k.back),
		"{S}", pair(k.left, k.right),
		"{B}", keyName(k.brake[0]),
		"{1}", pair(k.axes[0][0], k.axes[0][1]),
		"{2}", pair(k.axes[1][0], k.axes[1][1]),
		"{3}", pair(k.axes[2][0], k.axes[2][1]),
		"{4}", pair(k.axes[3][1], k.axes[3][0]), // shown left / right
		"{E}", keyName(k.enter),
		"{V}", keyName(k.camera),
	)
	return r.Replace(line)
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
