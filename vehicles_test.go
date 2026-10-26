package main

import (
	"math"
	"testing"
)

// run steps the full game simulation without rendering.
func run(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.step(fixedDT)
		for _, p := range g.players {
			g.updatePlayerCamera(p, fixedDT)
		}
	}
}

func TestExitAndEnter(t *testing.T) {
	g := NewGame()
	p := g.players[0]
	if p.Driving != g.loader {
		t.Fatal("player should start in the main loader")
	}
	if !g.exit(p) {
		t.Fatal("could not climb out of a stationary loader")
	}
	if p.Driving != nil || g.loader.Occupied {
		t.Fatal("still marked as driving after exit")
	}
	if hit, _ := g.world.CollideCircle(p.W.Pos, kidRadius, p.W); hit {
		t.Fatalf("kid was placed inside something at %v", p.W.Pos)
	}
	if d := p.W.Pos.Sub(g.loader.Joint).Len(); d > g.loader.EnterRadius() {
		t.Fatalf("exit point too far from the loader: %f", d)
	}
	g.rebuildObstacles()
	if v := g.nearestVehicle(p); v != Vehicle(g.loader) {
		t.Fatalf("nearest vehicle should be the loader, got %v", v)
	}
	g.toggleVehicle(p)
	if p.Driving != g.loader {
		t.Fatal("could not climb back in")
	}
}

func TestCannotExitAtSpeed(t *testing.T) {
	g := NewGame()
	g.loader.Speed = 120
	if g.exit(g.players[0]) {
		t.Fatal("climbed out of a moving loader")
	}
}

func TestWalkerBlockedBySolids(t *testing.T) {
	g := NewGame()
	p := g.players[0]
	g.exit(p)
	// walk straight at the parked dump truck A
	tr := g.world.Trucks[0]
	p.W.Pos = tr.Pos.Add(V(0, 90))
	p.W.Yaw = -math.Pi / 2
	p.scriptCtl = &Controls{Throttle: 1}
	for i := 0; i < 240; i++ {
		p.ctl = *p.scriptCtl
		g.step(fixedDT)
	}
	if d := p.W.Pos.Sub(tr.Pos).Len(); d < 30 {
		t.Fatalf("kid walked through the truck (dist %f)", d)
	}
}

func TestDriveTruckAndTip(t *testing.T) {
	g := NewGame()
	p := g.players[0]
	g.exit(p)
	tr := g.world.Trucks[1]
	g.enter(p, tr)
	start, h0 := tr.Pos, tr.Angle
	p.ctl = Controls{Throttle: 1, Steer: 1}
	for i := 0; i < 120; i++ {
		g.step(fixedDT)
	}
	if tr.Pos.Sub(start).Len() < 30 {
		t.Fatalf("truck barely moved: %v -> %v", start, tr.Pos)
	}
	if math.Abs(angleDiff(tr.Angle, h0)) < 0.2 {
		t.Fatalf("truck did not turn: %f -> %f", h0, tr.Angle)
	}
	// load it, stop, and tip the body
	tr.Contents[MatGravel] = 300
	tr.Speed = 0
	before := totalMaterial(g.world)
	p.ctl = Controls{Brake: true, A1: 1}
	for i := 0; i < 240; i++ {
		g.step(fixedDT)
	}
	if tr.BedTilt < 0.9 {
		t.Fatalf("body did not raise: %f", tr.BedTilt)
	}
	if tr.Total() > 150 {
		t.Fatalf("load did not pour out: %f left", tr.Total())
	}
	if got := totalMaterial(g.world) - before; got < 100 {
		t.Fatalf("poured material did not reach the ground: +%f", got)
	}
}

func TestExcavatorDigsAndLoadsTruck(t *testing.T) {
	g := NewGame()
	w := g.world
	var ex *Excavator
	for _, v := range g.vehicles {
		if e, ok := v.(*Excavator); ok {
			ex = e
		}
	}
	// park it facing the sand pit with the arm over the heap
	zone := w.Zones[2].Pos
	ex.Pos = zone.Add(V(-150, 0))
	ex.Heading, ex.Slew = 0, 0
	ex.Boom, ex.Stick, ex.Bucket = 0.6, 0.3, 0.2
	p := g.players[0]
	g.exit(p)
	g.enter(p, ex)
	// lower the boom, then curl and pull in
	p.ctl = Controls{A1: -1}
	for i := 0; i < 150; i++ {
		g.step(fixedDT)
	}
	if _, y := ex.Tip(); y < -2 {
		t.Fatalf("bucket teeth went below ground: %f", y)
	}
	p.ctl = Controls{A3: 1, A2: -1}
	for i := 0; i < 200; i++ {
		g.step(fixedDT)
	}
	if ex.Load < 20 {
		t.Fatalf("excavator dug too little: %f", ex.Load)
	}
	// raise and reach out, put a truck under the bucket, and open it
	p.ctl = Controls{A1: 1, A2: 1}
	for i := 0; i < 120; i++ {
		g.step(fixedDT)
	}
	tip, _ := ex.Tip()
	if reach := tip.Sub(ex.Pos).Len(); reach < 120 {
		t.Fatalf("arm reach too short to load a truck alongside: %f", reach)
	}
	tr := w.Trucks[0]
	tr.Pos, tr.Angle = tip.Add(V(14, 0)), 0
	load := ex.Load
	p.ctl = Controls{A3: -1}
	for i := 0; i < 120; i++ {
		g.step(fixedDT)
	}
	if tr.Total() < load*0.5 {
		t.Fatalf("truck got %f of the %f the bucket held", tr.Total(), load)
	}
}

func TestCraneLiftsCargoOntoTruck(t *testing.T) {
	g := NewGame()
	w := g.world
	c := g.crane
	item := w.Cargo[0]
	// put the trolley and hook right over the item, rope just above it
	off := item.Pos.Sub(c.Base)
	c.Slew, c.Trolley = off.Angle(), off.Len()
	c.hook, c.hookV = c.trolleyPos(), Vec{}
	c.HookLen = craneJibY - 6 - (item.Y + item.H) - 4
	p := g.players[0]
	g.exit(p)
	p.W.Pos = c.Base.Add(V(40, 0))
	g.enter(p, c)
	p.ctl = Controls{A3: 1}
	run(g, 2)
	if c.held != item {
		t.Fatalf("hook did not pick up the %s", item.Name())
	}
	// hoist it up
	p.ctl = Controls{A1: 1}
	run(g, 120)
	if item.Y < 60 {
		t.Fatalf("load did not rise: y=%f", item.Y)
	}
	// park a truck under the hook and lower onto its bed
	tr := w.Trucks[1]
	tr.Pos, tr.Angle = c.hook.Add(V(14, 0)), 0
	tr.Contents = [numMat]float64{}
	p.ctl = Controls{A1: -1}
	run(g, 400)
	p.ctl = Controls{A3: -1}
	run(g, 60)
	if item.truck != tr {
		t.Fatalf("the %s did not settle on the truck (y=%f)", item.Name(), item.Y)
	}
	// drive off: it rides along
	g.exit(p)
	g.enter(p, tr)
	before := item.Pos
	p.ctl = Controls{Throttle: 1}
	run(g, 90)
	if item.Pos.Sub(before).Len() < 30 || item.Pos.Sub(tr.Pos).Len() > 50 {
		t.Fatalf("load did not ride with the truck: item %v truck %v", item.Pos, tr.Pos)
	}
}

func TestAnimalsGetOutOfTheWay(t *testing.T) {
	g := NewGame()
	dog := g.animals[0]
	l := g.loader
	start := dog.Pos
	// drive the loader at the dog
	l.Joint = dog.Pos.Add(V(-200, 0))
	l.Heading, l.Artic = 0, 0
	g.players[0].ctl = Controls{Throttle: 1}
	hitDog := false
	for i := 0; i < 240; i++ {
		g.step(fixedDT)
		for _, c := range l.Colliders() {
			if c.c.Sub(dog.Pos).Len() < c.r+dog.radius()-2 {
				hitDog = true
			}
		}
	}
	if hitDog {
		t.Fatal("loader drove through the dog")
	}
	if dog.Pos.Sub(start).Len() < 20 {
		t.Fatal("dog did not move out of the way")
	}
}

func TestCoopJoinAndLeave(t *testing.T) {
	g := NewGame()
	g.setCoop(true)
	if len(g.players) != 2 {
		t.Fatal("player 2 did not join")
	}
	p2 := g.players[1]
	if hit, _ := g.world.CollideCircle(p2.W.Pos, kidRadius, p2.W); hit {
		t.Fatalf("player 2 spawned inside something at %v", p2.W.Pos)
	}
	run(g, 10)
	g.setCoop(false)
	if len(g.players) != 1 {
		t.Fatal("player 2 did not leave")
	}
}
