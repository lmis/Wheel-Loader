package main

import "github.com/hajimehoshi/ebiten/v2"

func keyDown(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

func (g *Game) handleInput() {
	l := g.loader

	// Toggles.
	if inpTapped(ebiten.KeyTab) {
		g.showHelp = !g.showHelp
	}
	if inpTapped(ebiten.KeyP) {
		g.paused = !g.paused
	}
	if inpTapped(ebiten.KeyBackspace) {
		g.resetLoader()
	}

	// Drive.
	l.throttle = 0
	if keyDown(ebiten.KeyW, ebiten.KeyArrowUp) {
		l.throttle = 1
	}
	if keyDown(ebiten.KeyS, ebiten.KeyArrowDown) {
		l.throttle = -1
	}
	l.brake = keyDown(ebiten.KeySpace)

	// Articulated steering.
	l.steer = 0
	if keyDown(ebiten.KeyA, ebiten.KeyArrowLeft) {
		l.steer = -1
	}
	if keyDown(ebiten.KeyD, ebiten.KeyArrowRight) {
		l.steer = 1
	}

	// Bucket lift (arms) and tilt (curl/dump).
	l.liftCmd = 0
	if keyDown(ebiten.KeyR) {
		l.liftCmd = 1
	}
	if keyDown(ebiten.KeyF) {
		l.liftCmd = -1
	}
	l.tiltCmd = 0
	if keyDown(ebiten.KeyT) {
		l.tiltCmd = 1 // curl back
	}
	if keyDown(ebiten.KeyG) {
		l.tiltCmd = -1 // dump forward
	}

	// Zoom (camera distance).
	if keyDown(ebiten.KeyEqual, ebiten.KeyKPAdd) {
		g.cam.TgtDist = clamp(g.cam.TgtDist*0.98, 180, 1100)
	}
	if keyDown(ebiten.KeyMinus, ebiten.KeyKPSubtract) {
		g.cam.TgtDist = clamp(g.cam.TgtDist*1.02, 180, 1100)
	}
	if _, dy := ebiten.Wheel(); dy != 0 {
		g.cam.TgtDist = clamp(g.cam.TgtDist*(1-dy*0.08), 180, 1100)
	}

	// Orbit the camera (yaw offset) and tilt (pitch).
	if keyDown(ebiten.KeyQ) {
		g.cam.UserYaw -= 1.6 * fixedDT
	}
	if keyDown(ebiten.KeyE) {
		g.cam.UserYaw += 1.6 * fixedDT
	}
	if keyDown(ebiten.KeyBracketLeft) {
		g.cam.Pitch = clamp(g.cam.Pitch-1.2*fixedDT, 0.2, 1.35)
	}
	if keyDown(ebiten.KeyBracketRight) {
		g.cam.Pitch = clamp(g.cam.Pitch+1.2*fixedDT, 0.2, 1.35)
	}
}

func (g *Game) resetLoader() {
	l := g.loader
	l.Speed = 0
	l.Artic = 0
	l.Load = 0
	l.Lift = 0.2
	l.Tilt = 0.8
}

// --- simple edge-trigger tracking for taps ---
var prevKeys = map[ebiten.Key]bool{}

func inpTapped(k ebiten.Key) bool {
	down := ebiten.IsKeyPressed(k)
	was := prevKeys[k]
	prevKeys[k] = down
	return down && !was
}
