package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Toddler mode, the whole game: two players, each with their own wheel
// loader, dump truck and tower crane, four direction keys and one button. The machines do the
// fiddly parts themselves:
//
//   - loader: the bucket scoops while driving into a pile, lifts when full
//     and tips out over a truck bed; the button dumps it anywhere
//   - dump truck: the button tips the body and lowers it again
//   - crane: left/right slews, up/down runs the trolley, and the button
//     lowers the hook to pick up or put down a load
//
// Number keys pick a machine: the kid climbs out, walks over and gets in.
// Nobody can climb out otherwise. Trucks fill up until a kid tips them out.

// toddlerKeys is one player's tiny control set.
type toddlerKeys struct {
	fwd, back, left, right, button []Key
	next                           []Key    // walk over to the next machine
	view                           []Key    // cab view / chase camera
	pick                           [4][]Key // loader, truck, crane, excavator
	label                          string
	pickLabels                     [4]string
}

var (
	toddlerP1 = toddlerKeys{
		fwd: []Key{rl.KeyW}, back: []Key{rl.KeyS},
		left: []Key{rl.KeyA}, right: []Key{rl.KeyD},
		button:     []Key{rl.KeySpace},
		next:       []Key{rl.KeyLeftShift},
		view:       []Key{rl.KeyV},
		pick:       [4][]Key{{rl.KeyOne}, {rl.KeyTwo}, {rl.KeyThree}, {rl.KeyFour}},
		label:      "W A S D: drive    Space: do    Left Shift: next machine    V: view",
		pickLabels: [4]string{"1", "2", "3", "4"},
	}
	toddlerP2 = toddlerKeys{
		fwd: []Key{rl.KeyUp}, back: []Key{rl.KeyDown},
		left: []Key{rl.KeyLeft}, right: []Key{rl.KeyRight},
		button: []Key{rl.KeyEnter},
		next:   []Key{rl.KeyRightShift},
		view:   []Key{rl.KeyM},
		pick: [4][]Key{{rl.KeyEight}, {rl.KeyNine},
			{rl.KeyZero}, {rl.KeyMinus}},
		label:      "Arrows: drive    Enter: do    Right Shift: next machine    M: view",
		pickLabels: [4]string{"8", "9", "0", "-"},
	}
	// single-screen play: the arrows, space, either shift, 1-4 and V
	singleKeys = toddlerKeys{
		fwd: []Key{rl.KeyUp}, back: []Key{rl.KeyDown},
		left: []Key{rl.KeyLeft}, right: []Key{rl.KeyRight},
		button:     []Key{rl.KeySpace},
		next:       []Key{rl.KeyLeftShift, rl.KeyRightShift},
		view:       []Key{rl.KeyV},
		pick:       [4][]Key{{rl.KeyOne}, {rl.KeyTwo}, {rl.KeyThree}, {rl.KeyFour}},
		label:      "Arrows: drive    Space: do    Shift: next machine    V: view",
		pickLabels: [4]string{"1", "2", "3", "4"},
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
	own       [4]Vehicle // loader, truck, crane, excavator
	goTo      Vehicle    // machine the kid is walking over to
	walkT     float64
	side      float64 // which way the kid walks around obstacles (±1)
	sideT     float64 // how long to keep to that side
	mode      int
	button    bool    // held
	pressed   bool    // pressed since the assistant last looked
	dumpTruck *Truck  // truck under the bucket when the dump started
	modeT     float64 // time in the current mode
	phase     int     // excavator: step of the dig cycle
	avoidT    float64 // driving: keep steering round an obstacle for this long
	cool      float64 // crane: time before the hook homes in again
	armed     bool    // crane: space was pressed near a load; lower once over it
	avoidSide float64 // which way (+1 right, -1 left)
	phaseT    float64 // time in that step
	hookLen   float64 // crane rope length last step, to see the hook land
}

// setupLanes puts each kid in their loader, with a heap and their truck
// right by the start and their crane and its loads behind.
func (g *Game) setupLanes() {
	w := g.world
	l2 := g.secondLoader()
	type lane struct {
		loader      *Loader
		at, pile    Vec
		truck       *Truck
		truckAt     Vec
		crane       *TowerCrane
		exc         *Excavator
		excAt       Vec
		keys        *toddlerKeys
		pileMat     int
		heading     float64
		truckFacing float64
	}
	lanes := []lane{
		{g.loader, V(1490, 1300), V(1450, 1040), w.Trucks[0], V(1300, 1120), g.crane2, g.excs[0], V(1230, 930), &toddlerP1, MatSand, -math.Pi / 2, -math.Pi / 2},
		{l2, V(1710, 1300), V(1750, 1040), w.Trucks[1], V(1895, 1110), g.crane, g.excs[1], V(1970, 930), &toddlerP2, MatDirt, -math.Pi / 2, -math.Pi / 2},
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
		p.distVehicle = l.CamDist()
		p.cam.Pitch, p.cam.UserYaw = 0.52, 0
		g.snapCamera(p)
		ln.truck.Pos, ln.truck.Angle, ln.truck.Speed, ln.truck.BedTilt = ln.truckAt, ln.truckFacing, 0, 0
		ln.truck.Contents = [numMat]float64{}
		// a big heap right in front of each player
		w.Grid.AddPile(ln.pile.X, ln.pile.Y, 90, 75, ln.pileMat)
		w.Grid.AddPile(ln.pile.X+30, ln.pile.Y-40, 60, 55, ln.pileMat)
		e := ln.exc
		e.Pos, e.Heading, e.Slew, e.Speed, e.Load = ln.excAt, ln.pile.Sub(ln.excAt).Angle(), 0, 0, 0 // beside the heap, facing it
		e.Boom, e.Stick, e.Bucket = excCarry[0], excCarry[1], excCarry[2]
		p.td = &toddlerState{keys: ln.keys, own: [4]Vehicle{l, ln.truck, ln.crane, e}}
	}
	for _, t := range w.Trucks {
		t.Capacity = toddlerTruckSize
	}
	// each kid's animals sit on their loader
	for _, a := range g.animals {
		a.ride, a.hopping, a.flying = a.owner.Driving, false, false
		a.Pos, a.Y, a.Yaw = a.rideSpot(a.ride)
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

// toddlerInput reads the four direction keys, the button and the machine
// picker keys.
func (g *Game) toddlerInput(p *Player) {
	td := p.td
	k := td.keys
	tapped := func(keys []Key) bool {
		hit := false
		for _, key := range keys {
			if inpTapped(key) {
				hit = true
			}
		}
		return hit
	}
	for i, keys := range k.pick {
		if tapped(keys) {
			g.toddlerPick(p, td.own[i])
		}
	}
	if tapped(k.next) {
		g.nextMachine(p)
	}
	if tapped(k.view) {
		p.fp = !p.fp // cab view / chase camera
	}
	_, crane := p.Driving.(*TowerCrane)
	p.ctl = arrowControls(keyDown(k.fwd...), keyDown(k.back...), keyDown(k.left...), keyDown(k.right...), crane)
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
			// a camera close enough to see the kid (and the animals) walk over
			td.walkT = 0
			p.distWalker, p.cam.Pitch, p.cam.UserYaw = 165, 0.6, 0
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
			td.phase, td.phaseT = exIdle, 0
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
	case *Excavator:
		g.toddlerDigger(p, v, dt)
	}
	td.pressed = false
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
	tip := l.BucketTip()
	inHeap := w.Grid.AmountNear(tip.X, tip.Y, 24) > 3
	l.ScoopAssist = td.mode == tdDig && inHeap && (p.ctl.Throttle > 0 || p.ctl.Steer != 0)
	switch td.mode {
	case tdDig:
		l.liftCmd, l.tiltCmd = toward(l.Lift, 0), toward(l.Tilt, 0.45)
		if td.pressed { // only the button lifts the bucket
			td.mode = tdCarry
		}
	case tdCarry:
		l.liftCmd, l.tiltCmd = toward(l.Lift, 0.62), toward(l.Tilt, 1)
		switch {
		case over != nil && l.Lift > 0.45 && l.Load > 2, td.pressed:
			td.mode, td.dumpTruck = tdDump, over
		}
	case tdDump:
		l.tiltCmd = -1
		if td.dumpTruck != nil {
			l.liftCmd = toward(l.Lift, 0.75)
		}
		if l.Load < 2 {
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
			td.mode, td.modeT = tdDump, 0
		}
	}
}

// toddlerHook makes the crane a toy. The hook travels at a safe height;
// swing it near a load and it homes in, drops, hooks on and lifts. Carrying a
// load, swing it near your truck and it homes in over the bed and sets the
// load down. The button lowers the hook (to pick up or put down) anywhere.
func (g *Game) toddlerHook(p *Player, c *TowerCrane, dt float64) {
	td := p.td
	w := g.world
	td.cool -= dt
	// what the hook is drawn to: the nearest free load, or the truck bed
	var target *Vec
	reach := 85.0
	if c.held == nil {
		best := reach
		for _, o := range w.Cargo {
			if o.held || o.truck != nil || o.Pos.Sub(c.Base).Len() > craneJibLen-10 {
				continue
			}
			if d := o.Pos.Sub(c.hook).Len(); d < best {
				best, target = d, &o.Pos
			}
		}
	} else if t, ok := td.own[1].(*Truck); ok {
		if bed := t.bedRect().Center; bed.Sub(c.hook).Len() < reach+20 {
			target = &bed
		}
	}
	if td.cool > 0 {
		target = nil // just let go: no grabbing it straight back
	}
	switch td.mode {
	case tdLower:
		c.hoistCmd = -1
		c.slewCmd, c.trolleyCmd = 0, 0
		// the hook has landed once it stops paying out rope
		landed := td.modeT > 0.15 && c.HookLen-td.hookLen < 240*dt*0.3
		if landed || td.modeT > 4 {
			if h := c.held; h != nil {
				h.held, h.vy = false, 0
				c.held = nil
				td.cool = 1.5
			} else {
				c.attach(w)
				if c.held == nil {
					td.cool = 1 // missed: don't keep trying the same spot
				}
			}
			td.mode, td.modeT = tdCarry, 0
		}
	default:
		want := toddlerHookY
		if c.held != nil {
			want += c.held.H
		}
		c.hoistCmd = clamp((want-c.HookY())/20, -1, 1)
		if target != nil {
			// home in: swing and run the trolley to put the hook right over it
			d := target.Sub(c.Base)
			ds := angleDiff(d.Angle(), c.Slew)
			dt2 := d.Len() - c.Trolley
			if p.ctl.Steer == 0 {
				c.slewCmd = clamp(ds*3-c.slewV*0.6, -1, 1)
			}
			if p.ctl.Throttle == 0 {
				c.trolleyCmd = clamp(dt2/25, -1, 1)
			}
			if td.pressed {
				td.armed = true // space near a load: home in, then pick it up
			}
			over := c.hook.Sub(*target).Len() < 9 && math.Abs(c.slewV) < 0.15 && td.modeT > 0.2
			if over && td.armed {
				td.mode, td.modeT, td.armed = tdLower, 0, false
			}
		}
		if td.pressed && td.modeT > 0.2 {
			td.mode, td.modeT = tdLower, 0
		}
	}
	c.grabCmd = 0
	td.hookLen = c.HookLen
}

// ---- the excavator ----

// excCarry is the excavator's resting pose: boom up, stick in, bucket curled.
var excCarry = [3]float64{0.62, 0.55, 0.65}

// Excavator assistant phases (td.phase), one dig-and-dump cycle per press.
const (
	exIdle  = iota
	exReach // swing to the front, boom down, stick out, bucket open, until the teeth bite
	exDig   // curl the bucket and pull the stick in through the material
	exLift  // boom up with the load
	exSwing // swing to the kid's truck (or out to the side)
	exDump  // open the bucket
	exBack  // back to the resting pose, facing forward
)

// toddlerDigger works the excavator: the direction keys drive the tracks,
// and the button runs a whole dig: reach in front, fill the bucket, lift,
// swing to the truck if it is in reach (or to the side), tip it out and
// swing back.
func (g *Game) toddlerDigger(p *Player, e *Excavator, dt float64) {
	td := p.td
	if p.ctl.Throttle == 0 && p.ctl.Steer == 0 {
		e.brake = true
	}
	pose := func(boom, stick, bucket float64) {
		e.boomCmd = clamp((boom-e.Boom)*5, -1, 1)
		e.stickCmd = clamp((stick-e.Stick)*5, -1, 1)
		e.bucketCmd = clamp((bucket-e.Bucket)*5, -1, 1)
	}
	slewTo := func(a float64) bool {
		d := angleDiff(a, e.Slew)
		e.slewCmd = clamp(d*2.2-e.slewV*0.8, -1, 1)
		return math.Abs(d) < 0.06 && math.Abs(e.slewV) < 0.12
	}
	td.phaseT += dt
	next := func(ph int) { td.phase, td.phaseT = ph, 0 }
	tip, tipY := e.Tip()
	surface := 0.0
	if cx, cy, ok := g.world.Grid.cellAt(tip.X, tip.Y); ok {
		surface = pileHeight(g.world.Grid.at(cx, cy).Amount)
	}
	switch td.phase {
	case exIdle:
		pose(excCarry[0], excCarry[1], excCarry[2])
		slewTo(0)
		if td.pressed {
			next(exReach)
		}
	case exReach:
		slewTo(0)
		e.boomCmd, e.stickCmd = -1, -0.6
		e.bucketCmd = clamp((0.25-e.Bucket)*5, -1, 1)
		if tipY < surface+5 || e.Boom <= 0.01 || td.phaseT > 3 {
			next(exDig)
		}
	case exDig:
		e.boomCmd, e.stickCmd, e.bucketCmd = 0.1, 1, 1
		if e.Load >= excCapacity*0.95 || e.Bucket >= 0.99 || td.phaseT > 2.5 {
			next(exLift)
		}
	case exLift:
		pose(0.85, 0.45, 1)
		if e.Boom > 0.8 {
			next(exSwing)
		}
	case exSwing:
		pose(0.85, 0.45, 1)
		target := math.Pi / 2 // out to the right, onto the ground
		if t, ok := td.own[1].(*Truck); ok {
			if d := t.bedRect().Center.Sub(e.Pos); d.Len() < 200 {
				target = angleDiff(d.Angle(), e.Heading)
			}
		}
		if slewTo(target) || td.phaseT > 5 {
			next(exDump)
		}
	case exDump:
		e.slewCmd = clamp(-e.slewV*2, -1, 1)
		// reach out over the truck, or just open the bucket
		if t, ok := td.own[1].(*Truck); ok {
			if d := t.bedRect().Center.Sub(e.Pos).Len(); d < 200 {
				e.stickCmd = clamp((tip.Sub(e.Pos).Len()-d)/40, -1, 1)
			}
		}
		e.boomCmd = clamp((0.85-e.Boom)*5, -1, 1)
		e.bucketCmd = -1
		if (e.Load < 1 && e.Bucket < 0.05) || td.phaseT > 3.5 {
			next(exBack)
		}
	case exBack:
		pose(excCarry[0], excCarry[1], excCarry[2])
		if slewTo(0) || td.phaseT > 5 {
			next(exIdle)
		}
	}
}

// ---- toddler HUD ----

// drawToddlerHUD: the player colour strip, the machine picker and the keys.
func (g *Game) drawToddlerHUD(dst canvas, p *Player, vw, vh float64) {
	td := p.td
	// player colour strip along the top
	panel(dst, 0, 0, vw, 8, p.marker)

	// machine picker: three tiles with the key and a little picture, the
	// current machine lit up
	for i, m := range td.own {
		x, y := vw-4*62+float64(i)*62-12, 20.0
		bg := hudPanel
		if p.Driving == m || td.goTo == m {
			bg = rgba(p.marker.R, p.marker.G, p.marker.B, 200)
		}
		panel(dst, x, y, 54, 64, bg)
		drawMachineIcon(dst, i, x+27, y+40)
		if !g.single {
			g.drawText(dst, td.keys.pickLabels[i], g.art.Bold, x+5, y+3, hudWhite)
		}
	}

	g.drawText(dst, td.keys.label, g.art.Small, 16, vh-24, rgba(200, 204, 210, 160))
}

// drawMachineIcon draws a tiny loader (0), dump truck (1), crane (2) or
// excavator (3).
func drawMachineIcon(dst canvas, kind int, cx, cy float64) {
	dark := rgba(40, 40, 44, 255)
	rect := func(x, y, w, h float64, c color.RGBA) {
		panel(dst, cx+x, cy+y, w, h, c)
	}
	wheel := func(x, y, r float64) {
		circle(dst, cx+x, cy+y, r, dark)
	}
	switch kind {
	case 0: // loader: body, cab, arm and bucket
		rect(-12, -6, 20, 8, colLoader)
		rect(-10, -14, 9, 8, colLoader)
		rect(6, -8, 8, 3, colLoader)
		rect(12, -8, 6, 10, dark)
		wheel(-8, 4, 5)
		wheel(5, 4, 5)
	case 1: // dump truck: body and cab
		rect(-17, -12, 22, 12, colTruck)
		rect(7, -10, 9, 10, colTruck)
		wheel(-11, 3, 4)
		wheel(-3, 3, 4)
		wheel(11, 3, 4)
	case 2: // crane: mast, jib, rope and hook
		rect(-10, -20, 4, 28, colCrane)
		rect(-16, -22, 34, 3, colCrane)
		rect(11, -19, 1, 12, dark)
		rect(9, -8, 5, 3, rgba(230, 60, 40, 255))
		rect(-14, 6, 12, 3, dark)
	case 3: // excavator: tracks, house, boom and stick down to a bucket
		rect(-16, 2, 22, 6, dark)
		rect(-14, -8, 16, 9, colExc)
		rect(-12, -14, 7, 6, colExc)
		rect(1, -16, 12, 3, colExc)
		rect(11, -16, 3, 14, colExc)
		rect(9, -4, 7, 4, dark)
	}
}
