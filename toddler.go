package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Toddler mode: two players, each with their own wheel loader, dump truck
// and tower crane, four direction keys and one button. The machines do the
// fiddly parts themselves:
//
//   - loader: the bucket scoops while driving into a pile, lifts when full
//     and tips out over a truck bed; the button dumps it anywhere
//   - dump truck: the button tips the body and lowers it again
//   - crane: left/right slews, up/down runs the trolley, and the button
//     lowers the hook to pick up or put down a load
//
// Number keys pick a machine: the kid climbs out, walks over and gets in.
// Nobody can climb out otherwise.

// toddlerKeys is one player's tiny control set.
type toddlerKeys struct {
	fwd, back, left, right, button []ebiten.Key
	pick                           [3][]ebiten.Key // loader, truck, crane
	label                          string
	pickLabels                     [3]string
}

var (
	toddlerP1 = toddlerKeys{
		fwd: []ebiten.Key{ebiten.KeyW}, back: []ebiten.Key{ebiten.KeyS},
		left: []ebiten.Key{ebiten.KeyA}, right: []ebiten.Key{ebiten.KeyD},
		button: []ebiten.Key{ebiten.KeySpace, ebiten.KeyShiftLeft},
		pick:   [3][]ebiten.Key{{ebiten.KeyDigit1}, {ebiten.KeyDigit2}, {ebiten.KeyDigit3}},
		label:  "W A S D  +  Space", pickLabels: [3]string{"1", "2", "3"},
	}
	toddlerP2 = toddlerKeys{
		fwd: []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyK}, back: []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyJ},
		left: []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyH}, right: []ebiten.Key{ebiten.KeyArrowRight, ebiten.KeyL},
		button: []ebiten.Key{ebiten.KeyEnter, ebiten.KeyShiftRight},
		pick: [3][]ebiten.Key{{ebiten.KeyDigit8, ebiten.KeyNumpad1}, {ebiten.KeyDigit9, ebiten.KeyNumpad2},
			{ebiten.KeyDigit0, ebiten.KeyNumpad3}},
		label: "Arrows  +  Enter", pickLabels: [3]string{"8", "9", "0"},
	}
)

// Assistant states. The loader uses dig/carry/dump, the truck carry (idle)/
// dump/lower and the crane carry (travel height)/lower.
const (
	tdDig   = iota // bucket down and level: drive into a pile to fill it
	tdCarry        // bucket up and curled; crane hook at travel height
	tdDump         // tipping it out
	tdLower        // truck body coming back down; crane hook going down
)

const (
	toddlerSpeedCap  = 135.0
	toddlerTruckCap  = 150.0
	toddlerTruckSize = 500.0
	toddlerHookY     = 150.0 // crane travel height of the hook
)

// toddlerState is each player's toddler-mode bookkeeping.
type toddlerState struct {
	keys      *toddlerKeys
	own       [3]Vehicle // loader, truck, crane
	goTo      Vehicle    // machine the kid is walking over to
	walkT     float64
	side      float64 // which way the kid walks around obstacles (±1)
	sideT     float64 // how long to keep to that side
	mode      int
	button    bool // held
	pressed   bool // pressed since the assistant last looked
	stars     int
	cheerT    float64 // a big star pops on screen after a delivery
	dumpTruck *Truck  // truck under the bucket when the dump started
	before    float64 // its load at that moment
	modeT     float64 // time in the current mode
	hookLen   float64 // crane rope length last step, to see the hook land
}

// setToddler switches toddler mode on (two players in loaders, simple
// controls, trucks and piles right by the start) or back to the full game.
func (g *Game) setToddler(on bool) {
	g.toddler = on
	w := g.world
	if !on {
		for _, p := range g.players {
			p.td = nil
		}
		for _, v := range g.vehicles {
			switch v := v.(type) {
			case *Loader:
				v.SpeedCap = 0
			case *Truck:
				v.SpeedCap = 0
			case *TowerCrane:
				v.Easy = false
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
		crane       *TowerCrane
		keys        *toddlerKeys
		pileMat     int
		heading     float64
		truckFacing float64
	}
	lanes := []lane{
		{g.loader, V(1490, 1300), V(1450, 1040), w.Trucks[0], V(1300, 1120), g.toddlerCrane(), &toddlerP1, MatSand, -math.Pi / 2, -math.Pi / 2},
		{l2, V(1710, 1300), V(1750, 1040), w.Trucks[1], V(1895, 1110), g.crane, &toddlerP2, MatDirt, -math.Pi / 2, -math.Pi / 2},
	}
	for i, ln := range lanes {
		p := g.players[i]
		l := ln.loader
		l.Joint, l.Heading, l.Artic, l.Speed = ln.at, ln.heading, 0, 0
		l.Lift, l.Tilt, l.Load = 0.05, 0.45, 0
		l.SpeedCap = toddlerSpeedCap
		ln.truck.SpeedCap = toddlerTruckCap
		ln.crane.Easy = true
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
		p.td = &toddlerState{keys: ln.keys, own: [3]Vehicle{l, ln.truck, ln.crane}}
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

// toddlerCrane returns player 1's tower crane, mirroring the site crane on
// the other side of the road, with its own loads. It is built the first time
// toddler mode starts.
func (g *Game) toddlerCrane() *TowerCrane {
	if g.crane2 != nil {
		return g.crane2
	}
	const mid = 1600.0 // the site road, the mirror line
	b := g.crane.Base
	g.crane2 = NewTowerCrane(V(2*mid-b.X, b.Y), math.Pi-g.crane.Slew)
	g.vehicles = append(g.vehicles, g.crane2)
	w := g.world
	for _, c := range w.Cargo[:len(w.Cargo)] {
		w.Cargo = append(w.Cargo, newCargo(c.Kind, V(2*mid-c.Pos.X, c.Pos.Y), math.Pi-c.Yaw))
	}
	return g.crane2
}

func (g *Game) secondLoader() *Loader {
	for _, v := range g.vehicles {
		if l, ok := v.(*Loader); ok && l != g.loader {
			return l
		}
	}
	return nil
}

// toddlerInput reads the four direction keys, the button and the machine
// picker keys.
func (g *Game) toddlerInput(p *Player) {
	td := p.td
	k := td.keys
	for i, keys := range k.pick {
		for _, key := range keys {
			if inpTapped(key) {
				g.toddlerPick(p, td.own[i])
			}
		}
	}
	p.ctl = Controls{Throttle: axis(k.fwd, k.back), Steer: axis(k.right, k.left)}
	down := keyDown(k.button...)
	if down && !td.button {
		td.pressed = true
	}
	td.button = down
}

// toddlerPick sends the kid over to machine v.
func (g *Game) toddlerPick(p *Player, v Vehicle) {
	td := p.td
	if v == nil || p.Driving == v || td.goTo == v {
		return
	}
	td.goTo, td.walkT = v, 0
}

// toddlerTravel walks the kid to the machine they picked: stop, climb out,
// walk over (running) and climb in. A kid that gets stuck is helped in.
func (g *Game) toddlerTravel(p *Player, dt float64) {
	td := p.td
	v := td.goTo
	if v == nil {
		return
	}
	td.walkT += dt
	if p.Driving != nil {
		if p.Driving == v {
			td.goTo = nil
			return
		}
		p.ctl = Controls{Brake: true}
		if math.Abs(p.Driving.SpeedVal()) < 30 && !g.exit(p) && td.walkT > 2 {
			// nowhere to climb out: hop down right next to it
			p.W.Pos, p.W.vel = p.Driving.Center(), Vec{}
			p.Driving.SetOccupied(false)
			p.Driving = nil
		}
		if p.Driving == nil {
			// a high camera that keeps the kid and the machines in view
			td.walkT = 0
			p.distWalker, p.cam.Pitch, p.cam.UserYaw = 300, 0.9, 0
		}
		return
	}
	near := false
	for _, o := range v.Colliders() {
		near = near || p.W.Pos.Sub(o.c).Len() < o.r+kidRadius+22
	}
	if near || td.walkT > 10 {
		if g.driverOf(v) == nil {
			g.enter(p, v)
			td.mode, td.modeT, td.dumpTruck, td.hookLen = tdCarry, 0, nil, 0
			if _, ok := v.(*Loader); ok {
				td.mode = tdDig
			}
		}
		td.goTo = nil
		p.ctl = Controls{}
		return
	}
	// head for the machine; when something is in the way, pick the side
	// with the shorter way round and stick to it for a while, following
	// the obstacle's edge
	want := v.Center().Sub(p.W.Pos).Angle()
	free := func(a float64) bool {
		probe := p.W.Pos.Add(dir(a).Scale(18))
		hit, _ := g.world.CollideCircle(probe, kidRadius, p.W)
		return !hit
	}
	td.sideT -= dt
	if !free(want) {
		if td.sideT <= 0 {
			td.side = 1
			for off := 0.3; off < 3.2; off += 0.3 {
				if free(want + off) {
					break
				}
				if free(want - off) {
					td.side = -1
					break
				}
			}
			td.sideT = 2.5
		}
		for off := 0.3; off < 3.2; off += 0.3 {
			if free(want + off*td.side) {
				want += off * td.side
				break
			}
		}
	}
	d := angleDiff(want, p.W.Yaw)
	c := Controls{Steer: clamp(d*3, -1, 1), Brake: true} // Brake = run
	if math.Abs(d) < 1.2 {
		c.Throttle = 1
	}
	p.ctl = c
}

// toddlerStep runs the toddler helpers for every player, after their
// controls were read and before the machines use them.
func (g *Game) toddlerStep(dt float64) {
	for _, p := range g.players {
		td := p.td
		if td == nil {
			continue
		}
		g.toddlerTravel(p, dt)
		if td.cheerT > 0 {
			td.cheerT -= dt
		}
	}
}

// toddlerAssist drives the fiddly parts of p's machine, after its controls
// were applied for this step.
func (g *Game) toddlerAssist(p *Player, dt float64) {
	td := p.td
	if td == nil || td.goTo != nil {
		return
	}
	td.modeT += dt
	switch v := p.Driving.(type) {
	case *Loader:
		g.toddlerLoader(p, v)
	case *Truck:
		g.toddlerDumpTruck(p, v)
	case *TowerCrane:
		g.toddlerHook(p, v, dt)
	}
	td.pressed = false
}

// cheer awards a star with confetti at a spot.
func (g *Game) cheer(td *toddlerState, at Vec, height float64) {
	td.stars++
	td.cheerT = 1.6
	g.world.spawnConfetti(at, height, 70)
}

// toward returns a command (-1..1) that moves cur toward target.
func toward(cur, target float64) float64 {
	d := target - cur
	if math.Abs(d) < 0.02 {
		return 0
	}
	return clamp(d*6, -1, 1)
}

// toddlerLoader runs the bucket of a toddler's loader.
func (g *Game) toddlerLoader(p *Player, l *Loader) {
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
				g.cheer(td, t.Pos, 60)
			}
			td.mode, td.dumpTruck = tdDig, nil
		}
	}
}

// toddlerDumpTruck: the button tips the body all the way, waits for the load
// to pour out, and brings it back down.
func (g *Game) toddlerDumpTruck(p *Player, t *Truck) {
	td := p.td
	if p.ctl.Throttle == 0 {
		t.brake = true
	}
	switch td.mode {
	case tdDump:
		t.bedCmd = 1
		if t.BedTilt >= 1 && (t.Total() < 1 || td.modeT > 4) {
			if td.before-t.Total() > 10 {
				g.cheer(td, t.Pos.Sub(dir(t.Angle).Scale(80)), 40)
			}
			td.mode, td.modeT = tdLower, 0
		}
	case tdLower:
		t.bedCmd = -1
		if t.BedTilt <= 0 {
			td.mode = tdCarry
		}
	default:
		t.bedCmd = 0
		if td.pressed {
			td.mode, td.modeT, td.before = tdDump, 0, t.Total()
		}
	}
}

// toddlerHook keeps the crane hook at a safe travel height. The button
// lowers it until it lands, then hooks on whatever is below or lets go of
// the load, and the hook goes back up.
func (g *Game) toddlerHook(p *Player, c *TowerCrane, dt float64) {
	td := p.td
	w := g.world
	switch td.mode {
	case tdLower:
		c.hoistCmd = -1
		// the hook has landed once it stops paying out rope
		landed := td.modeT > 0.2 && c.HookLen-td.hookLen < 80*dt*0.3
		if landed || td.modeT > 6 {
			if h := c.held; h != nil {
				h.held, h.vy = false, 0
				c.held = nil
				if w.TruckAt(h.Pos) != nil {
					g.cheer(td, h.Pos, h.H+40)
				}
			} else {
				c.attach(w)
			}
			td.mode, td.modeT = tdCarry, 0
		}
	default:
		want := toddlerHookY
		if c.held != nil {
			want += c.held.H
		}
		c.hoistCmd = clamp((want-c.HookY())/30, -1, 1)
		if td.pressed && td.modeT > 0.3 {
			td.mode, td.modeT = tdLower, 0
		}
	}
	c.grabCmd = 0
	td.hookLen = c.HookLen
}

// toddlerTrucks empties full, parked trucks with a big confetti shower so they can
// be filled again.
func (g *Game) toddlerTrucks(dt float64) {
	for _, t := range g.world.Trucks {
		if t.Total() >= t.Capacity*0.9 && !t.Occupied {
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

	// load meter: how full the bucket or the truck is, in the material's
	// colour
	meter := func(frac float64, mat int, empty bool) {
		bw, bh := vw*0.5, 30.0
		bx, by := vw/2-bw/2, vh-bh-24
		panel(dst, bx-8, by-8, bw+16, bh+16, hudPanel)
		r, gg, b := matColor(mat)
		if empty {
			r, gg, b = 120, 120, 120
		}
		bar(dst, bx, by, bw, bh, frac, rgba(r, gg, b, 255), hudTrack)
	}
	switch v := p.Driving.(type) {
	case *Loader:
		meter(v.LoadFrac(), v.LoadType, v.Load < 1)
	case *Truck:
		mat, most := 0, 0.0
		for m, a := range v.Contents {
			if a > most {
				mat, most = m, a
			}
		}
		meter(v.Total()/v.Capacity, mat, most < 1)
	}

	// machine picker: three tiles with the key and a little picture, the
	// current machine lit up
	for i, m := range td.own {
		x, y := vw-3*62+float64(i)*62-12, 20.0
		bg := hudPanel
		if p.Driving == m || td.goTo == m {
			bg = rgba(p.marker.R, p.marker.G, p.marker.B, 200)
		}
		panel(dst, x, y, 54, 64, bg)
		drawMachineIcon(dst, i, x+27, y+40)
		g.drawText(dst, td.keys.pickLabels[i], g.art.Bold, x+5, y+3, hudWhite)
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

// drawMachineIcon draws a tiny loader (0), dump truck (1) or crane (2).
func drawMachineIcon(dst *ebiten.Image, kind int, cx, cy float64) {
	yellow := rgba(246, 192, 22, 255)
	orange := rgba(232, 140, 30, 255)
	dark := rgba(40, 40, 44, 255)
	rect := func(x, y, w, h float64, c color.RGBA) {
		vector.DrawFilledRect(dst, float32(cx+x), float32(cy+y), float32(w), float32(h), c, true)
	}
	wheel := func(x, y, r float64) {
		vector.DrawFilledCircle(dst, float32(cx+x), float32(cy+y), float32(r), dark, true)
	}
	switch kind {
	case 0: // loader: body, cab, arm and bucket
		rect(-12, -6, 20, 8, yellow)
		rect(-10, -14, 9, 8, yellow)
		rect(6, -8, 8, 3, yellow)
		rect(12, -8, 6, 10, dark)
		wheel(-8, 4, 5)
		wheel(5, 4, 5)
	case 1: // dump truck: body and cab
		rect(-17, -12, 22, 12, orange)
		rect(7, -10, 9, 10, orange)
		wheel(-11, 3, 4)
		wheel(-3, 3, 4)
		wheel(11, 3, 4)
	case 2: // crane: mast, jib, rope and hook
		rect(-10, -20, 4, 28, yellow)
		rect(-16, -22, 34, 3, yellow)
		rect(11, -19, 1, 12, dark)
		rect(9, -8, 5, 3, rgba(230, 60, 40, 255))
		rect(-14, 6, 12, 3, dark)
	}
}
