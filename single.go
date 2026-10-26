package main

import (
	"path/filepath"
)

// Single-screen play (the default): one kid, the whole screen, and only the
// arrow keys, space, shift, 1-4 and V. Split screen plays exactly the same,
// just with two key sets (see toddlerP1 and toddlerP2). Up drives forward and down backs up;
// left and right steer while driving forward (so one key is enough to go
// round a corner). Space does the machine's one thing (lift or tip the
// loader's bucket, tip the truck, dig once with the excavator, pick up or
// drop with the crane); shift climbs out and walks over to the next machine;
// 1-4 pick a machine; V switches to the cab view. -split-screen brings back
// two kids side by side.

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
	p.td.keys = &singleKeys
	for _, q := range g.players[1:] {
		if q.Driving != nil {
			q.Driving.SetOccupied(false) // player 2's machines stand parked
		}
	}
	g.players = g.players[:1]
	// only the kid's own machines stay: nothing on the site they cannot use
	own := map[Vehicle]bool{}
	for _, m := range p.td.own {
		own[m] = true
	}
	var vs []Vehicle
	for _, v := range g.vehicles {
		if own[v] {
			vs = append(vs, v)
		}
	}
	g.vehicles = vs
	var trucks []*Truck
	for _, t := range g.world.Trucks {
		if own[t] {
			trucks = append(trucks, t)
		}
	}
	g.world.Trucks = trucks
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
	g.snapCamera(p)
}

// arrowControls maps the arrows to controls: up forward, down back, and
// left or right steer while driving forward (backing up when down is held
// too). On the crane they swing the jib and run the trolley.
func arrowControls(up, down, left, right, crane bool) Controls {
	b := func(k bool) float64 {
		if k {
			return 1
		}
		return 0
	}
	c := Controls{Throttle: b(up) - b(down), Steer: b(right) - b(left)}
	if c.Throttle == 0 && c.Steer != 0 && !crane {
		c.Throttle = 1 // a turn on its own drives forward round the corner
	}
	return c
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

// stageSingleScenes captures single-screen play.
func (g *Game) stageSingleScenes(outDir string) {
	g.scripted = true
	g.simSpeed = 2
	p := g.players[0]
	steps := []struct {
		name   string
		frames int
		ctl    Controls
		setup  func()
	}{
		{"s1_start", 10, Controls{}, nil},
		{"s2_right_arrow", 70, Controls{Throttle: 1, Steer: 1}, nil},
		{"s3_up_arrow", 50, Controls{Throttle: 1}, nil},
		{"s4_shift_next_machine", 60, Controls{}, func() { g.nextMachine(p) }},
		{"s5_truck_down_arrow", 50, Controls{Throttle: -1}, nil},
		{"s6_crane", 90, Controls{}, func() { g.toddlerPick(p, p.td.own[2]) }},
		{"s7_excavator", 90, Controls{}, func() { g.toddlerPick(p, p.td.own[3]) }},
	}
	f := 5
	for _, st := range steps {
		st := st
		start := f
		g.script = append(g.script, scriptStep{at: start, fn: func(g *Game) {
			ctl := st.ctl
			p.scriptCtl = &ctl
			if st.setup != nil {
				st.setup()
			}
			g.shots[start+st.frames] = filepath.Join(outDir, st.name+".png")
		}})
		f += st.frames + 3
	}
	g.exitFrame = f + 5
}
