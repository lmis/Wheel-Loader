package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// handleInput reads the keyboard (and, for a solo player, the mouse) once
// per frame.
func (g *Game) handleInput() {
	// global keys
	if inpTapped(ebiten.KeyF1) {
		g.showHelp = !g.showHelp
	}
	if inpTapped(ebiten.KeyF2) {
		g.paused = !g.paused
	}
	if inpTapped(ebiten.KeyF3) {
		g.resetVehicles()
	}
	if inpTapped(ebiten.KeyF4) {
		g.setCoop(!g.coop())
	}
	// player 2 joins by pressing their get-in key
	if !g.coop() && inpTapped(p2Keys.enter) {
		g.setCoop(true)
	}
	if inpTapped(ebiten.KeyEscape) && g.mouseCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
		g.mouseCaptured = false
	}

	for _, p := range g.players {
		g.playerInput(p)
	}
	if !g.coop() {
		g.mouseLook(g.players[0])
	}
}

func (g *Game) playerInput(p *Player) {
	k := &p.keys
	if inpTapped(k.enter) {
		g.toggleVehicle(p)
	}
	if inpTapped(k.camera) {
		p.fp = !p.fp
		p.headYaw, p.headPitch = 0, 0
	}
	p.ctl = k.read()

	zoom := 0.0
	if keyDown(k.zoomIn...) {
		zoom = -1
	}
	if keyDown(k.zoomOut...) {
		zoom = 1
	}
	if p.ID == 1 && !g.coop() {
		if _, dy := ebiten.Wheel(); dy != 0 {
			zoom = -dy * 4
		}
	}
	if zoom != 0 {
		f := 1 + zoom*0.02
		if p.Driving != nil {
			p.distVehicle = clamp(p.distVehicle*f, 160, 1300)
		} else {
			p.distWalker = clamp(p.distWalker*f, 60, 500)
		}
	}
	p.near = nil
	if p.Driving == nil {
		p.near = g.nearestVehicle(p)
	}
}

// mouseLook: click to capture the mouse (Esc releases it). It then turns the
// kid, the cab view, or orbits the chase camera.
func (g *Game) mouseLook(p *Player) {
	mx, my := ebiten.CursorPosition()
	dx, dy := float64(mx-g.lastMX), float64(my-g.lastMY)
	g.lastMX, g.lastMY = mx, my
	if !g.mouseCaptured {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			ebiten.SetCursorMode(ebiten.CursorModeCaptured)
			g.mouseCaptured = true
		}
		return
	}
	if math.Abs(dx) > 200 || math.Abs(dy) > 200 {
		return // capture just started; ignore the jump
	}
	const sens = 0.0035
	switch {
	case p.Driving != nil && p.fp:
		p.headYaw = clamp(p.headYaw+dx*sens, -2.4, 2.4)
		p.headPitch = clamp(p.headPitch-dy*sens, -0.9, 0.9)
	case p.Driving != nil:
		p.cam.UserYaw += dx * sens
		p.cam.Pitch = clamp(p.cam.Pitch+dy*0.003, 0.2, 1.35)
	default:
		p.W.Yaw += dx * sens
		if p.fp {
			p.W.Pitch = clamp(p.W.Pitch-dy*sens, -1.2, 1.1)
		} else {
			p.cam.Pitch = clamp(p.cam.Pitch+dy*0.003, 0.15, 1.35)
		}
	}
}

// resetVehicles stops every machine and settles the main loader (F3).
func (g *Game) resetVehicles() {
	l := g.loader
	l.Speed, l.Artic, l.Load, l.Lift, l.Tilt = 0, 0, 0, 0.2, 0.8
	for _, v := range g.vehicles {
		switch m := v.(type) {
		case *Truck:
			m.Speed, m.Steer, m.BedTilt = 0, 0, 0
		case *Car:
			m.Speed, m.Steer = 0, 0
		case *Excavator:
			m.Speed = 0
		}
	}
}
