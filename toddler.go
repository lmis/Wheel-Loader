package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Toddler mode: two players, each in a wheel loader, with only four
// direction keys and one button. The bucket works itself: it scoops while
// driving into a pile, lifts when full, and tips out over a truck bed. The
// button dumps the bucket anywhere. Nobody can climb out.

// toddlerKeys is one player's tiny control set.
type toddlerKeys struct {
	fwd, back, left, right, button []ebiten.Key
	label                          string
}

var (
	toddlerP1 = toddlerKeys{
		fwd: []ebiten.Key{ebiten.KeyW}, back: []ebiten.Key{ebiten.KeyS},
		left: []ebiten.Key{ebiten.KeyA}, right: []ebiten.Key{ebiten.KeyD},
		button: []ebiten.Key{ebiten.KeySpace, ebiten.KeyShiftLeft},
		label:  "W A S D  +  Space",
	}
	toddlerP2 = toddlerKeys{
		fwd: []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyK}, back: []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyJ},
		left: []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyH}, right: []ebiten.Key{ebiten.KeyArrowRight, ebiten.KeyL},
		button: []ebiten.Key{ebiten.KeyEnter, ebiten.KeyShiftRight},
		label:  "Arrows  +  Enter",
	}
)

// Bucket assistant states.
const (
	tdDig   = iota // bucket down and level: drive into a pile to fill it
	tdCarry        // bucket up and curled
	tdDump         // tipping it out
)

const (
	toddlerSpeedCap  = 135.0
	toddlerTruckSize = 500.0
)

// toddlerState is each player's toddler-mode bookkeeping.
type toddlerState struct {
	keys      *toddlerKeys
	mode      int
	button    bool
	stars     int
	cheerT    float64 // a big star pops on screen after a delivery
	dumpTruck *Truck  // truck under the bucket when the dump started
	before    float64 // its load at that moment
}

// setToddler switches toddler mode on (two players in loaders, simple
// controls, trucks and piles right by the start) or back to the full game.
func (g *Game) setToddler(on bool) {
	g.toddler = on
	w := g.world
	if !on {
		for _, p := range g.players {
			p.td = nil
			if l, ok := p.Driving.(*Loader); ok {
				l.SpeedCap = 0
			}
		}
		for _, t := range w.Trucks {
			t.Capacity = 900
		}
		if g.coop() {
			g.players[0].keys = p1Keys
		}
		return
	}
	g.setCoop(true)
	l2 := g.secondLoader()
	type lane struct {
		loader      *Loader
		at, pile    Vec
		truck       *Truck
		truckAt     Vec
		keys        *toddlerKeys
		pileMat     int
		heading     float64
		truckFacing float64
	}
	lanes := []lane{
		{g.loader, V(1490, 1300), V(1450, 1040), w.Trucks[0], V(1300, 1120), &toddlerP1, MatSand, -math.Pi / 2, -math.Pi / 2},
		{l2, V(1710, 1300), V(1750, 1040), w.Trucks[1], V(1895, 1110), &toddlerP2, MatDirt, -math.Pi / 2, -math.Pi / 2},
	}
	for i, ln := range lanes {
		p := g.players[i]
		l := ln.loader
		l.Joint, l.Heading, l.Artic, l.Speed = ln.at, ln.heading, 0, 0
		l.Lift, l.Tilt, l.Load = 0.05, 0.45, 0
		l.SpeedCap = toddlerSpeedCap
		g.put(p, l)
		p.fp = false
		p.distVehicle = 430
		p.cam.Pitch, p.cam.UserYaw = 0.6, 0
		g.snapCamera(p)
		ln.truck.Pos, ln.truck.Angle, ln.truck.Speed, ln.truck.BedTilt = ln.truckAt, ln.truckFacing, 0, 0
		ln.truck.Contents = [numMat]float64{}
		// a big heap right in front of each player
		w.Grid.AddPile(ln.pile.X, ln.pile.Y, 90, 75, ln.pileMat)
		w.Grid.AddPile(ln.pile.X+30, ln.pile.Y-40, 60, 55, ln.pileMat)
		p.td = &toddlerState{keys: ln.keys}
	}
	for _, t := range w.Trucks {
		t.Capacity = toddlerTruckSize
	}
	// the animals watch from behind the start, out of the driving lanes
	homes := []Vec{V(1600, 1420), V(1530, 1470), V(1670, 1460), V(1600, 1520)}
	for i, a := range g.animals {
		a.home, a.Pos, a.walking, a.timer = homes[i], homes[i], false, 2
	}
}

func (g *Game) secondLoader() *Loader {
	for _, v := range g.vehicles {
		if l, ok := v.(*Loader); ok && l != g.loader {
			return l
		}
	}
	return nil
}

// toddlerInput reads the four direction keys and the button.
func (g *Game) toddlerInput(p *Player) {
	k := p.td.keys
	p.ctl = Controls{Throttle: axis(k.fwd, k.back), Steer: axis(k.right, k.left)}
	p.td.button = keyDown(k.button...)
}

// toward returns a command (-1..1) that moves cur toward target.
func toward(cur, target float64) float64 {
	d := target - cur
	if math.Abs(d) < 0.02 {
		return 0
	}
	return clamp(d*6, -1, 1)
}

// toddlerAssist runs the bucket for a toddler's loader, after its controls
// were applied for this step.
func (g *Game) toddlerAssist(p *Player, l *Loader, dt float64) {
	td := p.td
	w := g.world
	if p.ctl.Throttle == 0 {
		l.brake = true // stop as soon as the key is let go
	}
	over := w.TruckAt(l.BucketTip())
	if over == nil {
		over = w.TruckAt(l.BucketPivot())
	}
	switch td.mode {
	case tdDig:
		l.liftCmd, l.tiltCmd = toward(l.Lift, 0), toward(l.Tilt, 0.45)
		switch {
		case l.Load > bucketCap*0.8, l.Load > 12 && l.Speed < -5:
			td.mode = tdCarry
		case l.Load > 5 && td.button:
			td.mode, td.dumpTruck = tdDump, nil
		}
	case tdCarry:
		l.liftCmd, l.tiltCmd = toward(l.Lift, 0.62), toward(l.Tilt, 1)
		switch {
		case l.Load < 2:
			td.mode = tdDig
		case over != nil && l.Lift > 0.45, td.button:
			td.mode, td.dumpTruck = tdDump, over
			if over != nil {
				td.before = over.Total()
			}
		}
	case tdDump:
		l.tiltCmd = -1
		if td.dumpTruck != nil {
			l.liftCmd = toward(l.Lift, 0.75)
		}
		if l.Load < 2 {
			if t := td.dumpTruck; t != nil && t.Total()-td.before > 10 {
				td.stars++
				td.cheerT = 1.6
				w.spawnConfetti(t.Pos, 60, 70)
			}
			td.mode, td.dumpTruck = tdDig, nil
		}
	}
	if td.cheerT > 0 {
		td.cheerT -= dt
	}
}

// toddlerTrucks empties full trucks with a big confetti shower so they can
// be filled again.
func (g *Game) toddlerTrucks(dt float64) {
	for _, t := range g.world.Trucks {
		if t.Total() >= t.Capacity*0.9 {
			t.fullT += dt
			if t.fullT > 1.2 {
				g.world.spawnConfetti(t.Pos, 70, 180)
				t.Contents = [numMat]float64{}
				t.fullT = 0
			}
		} else {
			t.fullT = 0
		}
	}
}

// ---- toddler HUD ----

// drawStar fills a five-pointed star.
func drawStar(dst *ebiten.Image, cx, cy, r float64, c [4]uint8) {
	var path vector.Path
	for i := 0; i < 10; i++ {
		rr := r
		if i%2 == 1 {
			rr = r * 0.45
		}
		a := -math.Pi/2 + float64(i)*math.Pi/5
		x, y := float32(cx+math.Cos(a)*rr), float32(cy+math.Sin(a)*rr)
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(rgba(c[0], c[1], c[2], c[3]))
	vector.FillPath(dst, &path, nil, op)
}

// drawToddlerHUD: big star counter, a big bucket meter and the player colour.
func (g *Game) drawToddlerHUD(dst *ebiten.Image, p *Player, vw, vh float64) {
	td := p.td
	// player colour strip along the top
	vector.DrawFilledRect(dst, 0, 0, float32(vw), 8, p.marker, false)

	// stars
	panel(dst, 16, 20, 170, 64, hudPanel)
	drawStar(dst, 50, 52, 24, [4]uint8{255, 205, 40, 255})
	g.drawText(dst, fmt.Sprintf("%d", td.stars), g.art.Huge, 86, 26, hudWhite)

	// bucket meter: how full the bucket is, in the material's colour
	if l, ok := p.Driving.(*Loader); ok {
		bw, bh := vw*0.5, 30.0
		bx, by := vw/2-bw/2, vh-bh-24
		panel(dst, bx-8, by-8, bw+16, bh+16, hudPanel)
		r, gg, b := matColor(l.LoadType)
		if l.Load < 1 {
			r, gg, b = 120, 120, 120
		}
		bar(dst, bx, by, bw, bh, l.LoadFrac(), rgba(r, gg, b, 255), hudTrack)
	}

	// a big star pops up after a delivery
	if td.cheerT > 0 {
		t := 1.6 - td.cheerT
		size := 40 + 140*math.Min(1, t*3)
		a := uint8(clamp(td.cheerT/1.6*255, 0, 255))
		drawStar(dst, vw/2, vh*0.42, size, [4]uint8{255, 214, 50, a})
	}
	g.drawText(dst, td.keys.label, g.art.Small, 16, vh-24, rgba(200, 204, 210, 160))
}
