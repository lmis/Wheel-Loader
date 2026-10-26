package main

import (
	"math"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Single-screen play (the default): one kid, the whole screen, the arrow
// keys, the space bar and shift. An arrow means "go that way on the
// screen": the machine turns toward it and drives there, so one key is
// enough. The camera keeps one fixed direction, so screen directions stay
// put. Space does the machine's one thing (scoop or tip the loader's bucket,
// tip the truck, dig once with the excavator, pick up or drop with the
// crane); shift climbs out and walks over to the next machine.
// -split-screen brings back two kids side by side.

// singleYaw is the fixed way the single-screen camera looks (north, the way
// the machines face at the start).
const singleYaw = -math.Pi / 2

// Companions for single-screen play.
const (
	companionsOwl  = "owl"   // the snowy owl and the chimp
	companionsDogs = "doggo" // the golden retriever and the Bernese
)

// setSingle turns a two-kid game into single-screen play for player 1, with
// the chosen animal friends.
func (g *Game) setSingle(companions string) {
	g.single = true
	p := g.players[0]
	for _, q := range g.players[1:] {
		if q.Driving != nil {
			q.Driving.SetOccupied(false) // player 2's machines stand parked
		}
	}
	g.players = g.players[:1]
	var keep []*Animal
	for _, a := range g.animals {
		dog := a.Kind == animalGolden || a.Kind == animalBernese
		if dog == (companions == companionsDogs) {
			a.owner, a.slot = p, len(keep)
			keep = append(keep, a)
		}
	}
	g.animals = keep
	for _, a := range g.animals {
		a.ride, a.hopping, a.flying = p.Driving, false, false
		a.Pos, a.Y, a.Yaw = a.rideSpot(a.ride)
	}
	p.cam.UserYaw = 0
	p.cam.followYaw = singleYaw
	g.snapCamera(p)
}

var (
	arrowUp    = []Key{rl.KeyUp}
	arrowDown  = []Key{rl.KeyDown}
	arrowLeft  = []Key{rl.KeyLeft}
	arrowRight = []Key{rl.KeyRight}
)

// singleInput reads the arrows as a direction on the screen, space as the
// machine's action and shift as "next machine".
func (g *Game) singleInput(p *Player) {
	td := p.td
	up := dir(singleYaw)
	right := dir(singleYaw + math.Pi/2)
	want := up.Scale(axis(arrowUp, arrowDown)).Add(right.Scale(axis(arrowRight, arrowLeft)))
	if want.Len() > 0 {
		want = want.Norm()
	}
	td.want = want
	down := keyDown(rl.KeySpace)
	if down && !td.button {
		td.pressed = true
	}
	td.button = down
	if inpTapped(rl.KeyLeftShift) || inpTapped(rl.KeyRightShift) {
		g.nextMachine(p)
	}
	p.ctl = Controls{}
}

// nextMachine sends the kid to the next of their machines.
func (g *Game) nextMachine(p *Player) {
	td := p.td
	cur := p.Driving
	if td.goTo != nil {
		cur = td.goTo
	}
	for i, m := range td.own {
		if m == cur {
			g.toddlerPick(p, td.own[(i+1)%len(td.own)])
			return
		}
	}
	g.toddlerPick(p, td.own[0])
}

// singleDrive turns the wanted screen direction into the machine's
// controls: turn toward it, and drive when facing roughly that way.
func (g *Game) singleDrive(p *Player, dt float64) {
	td := p.td
	want := td.want
	if want.Len() == 0 {
		td.idleT += dt
	} else {
		td.idleT = 0
	}
	if v := p.Driving; v != nil {
		td.moved += math.Abs(v.SpeedVal()) * dt
	}
	if want.Len() == 0 || td.goTo != nil {
		return
	}
	a := want.Angle()
	drive := func(heading, sharp float64, spin bool) {
		d := angleDiff(a, heading)
		p.ctl.Steer = clamp(d*2, -1, 1)
		switch {
		case math.Abs(d) < sharp:
			p.ctl.Throttle = 1
		case spin:
			p.ctl.Throttle = 0 // tracks turn on the spot
		default:
			p.ctl.Throttle = 0.45 // wheels: keep rolling while turning
		}
	}
	switch v := p.Driving.(type) {
	case *Loader:
		drive(v.FrontHeading(), 0.9, false)
	case *Truck:
		drive(v.Angle, 0.9, false)
	case *Excavator:
		drive(v.Heading, 0.5, true)
	case *TowerCrane:
		// the hook moves that way: in and out along the jib, and round it
		p.ctl.Throttle = want.Dot(dir(v.Slew))
		p.ctl.Steer = want.Dot(dir(v.Slew + math.Pi/2))
	}
}

// stageSingleScenes captures single-screen play.
func (g *Game) stageSingleScenes(outDir string) {
	g.scripted = true
	g.simSpeed = 2
	p := g.players[0]
	steps := []struct {
		name   string
		frames int
		want   Vec
		setup  func()
	}{
		{"s1_start", 10, Vec{}, nil},
		{"s2_right_arrow", 70, V(1, 0), nil},
		{"s3_up_arrow", 50, V(0, -1), nil},
		{"s4_shift_next_machine", 60, Vec{}, func() { g.nextMachine(p) }},
		{"s5_truck_down_arrow", 50, V(0, 1), nil},
		{"s6_crane", 90, Vec{}, func() { g.toddlerPick(p, p.td.own[2]) }},
		{"s7_excavator", 90, Vec{}, func() { g.toddlerPick(p, p.td.own[3]) }},
	}
	f := 5
	for _, st := range steps {
		st := st
		start := f
		g.script = append(g.script, scriptStep{at: start, fn: func(g *Game) {
			p.td.want = st.want
			if st.setup != nil {
				st.setup()
			}
			g.shots[start+st.frames] = filepath.Join(outDir, st.name+".png")
		}})
		f += st.frames + 3
	}
	g.exitFrame = f + 5
}
