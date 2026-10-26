package main

import (
	"math"
	"testing"
)

// toddlerDrive plays like a toddler holding keys: it steers toward target
// with the left/right keys and holds forward (or back), with no other input.
func toddlerDrive(g *Game, p *Player, target Vec, back bool, maxSteps int, done func() bool) bool {
	l := p.Driving.(*Loader)
	for i := 0; i < maxSteps; i++ {
		if done() {
			return true
		}
		want := target.Sub(l.BucketPivot()).Angle()
		d := angleDiff(want, l.FrontHeading())
		c := Controls{Throttle: 1}
		if back {
			c.Throttle = -1
		}
		if math.Abs(d) > 0.05 && !back {
			c.Steer = math.Copysign(1, d)
		}
		p.ctl = c
		run(g, 1)
	}
	return done()
}

func TestToddlerModeLoadsTruck(t *testing.T) {
	g := NewGame()
	g.setToddler(true)
	if !g.coop() || len(g.players) != 2 {
		t.Fatal("toddler mode should start two players")
	}
	for _, p := range g.players {
		if _, ok := p.Driving.(*Loader); !ok {
			t.Fatalf("player %d is not in a loader", p.ID)
		}
	}
	p := g.players[0]
	l := p.Driving.(*Loader)
	// drive into the sand pile: the bucket fills and lifts by itself
	if !toddlerDrive(g, p, V(1450, 1040), false, 900, func() bool { return p.td.mode == tdCarry }) {
		t.Fatalf("bucket never filled: load %.1f", l.Load)
	}
	run(g, 90)
	if l.Lift < 0.45 {
		t.Fatalf("bucket should lift for carrying, lift %.2f", l.Lift)
	}
	// back out, then drive to the truck: it tips in by itself
	toddlerDrive(g, p, Vec{}, true, 70, func() bool { return false })
	tr := g.world.Trucks[0]
	if !toddlerDrive(g, p, tr.Pos, false, 1500, func() bool { return p.td.stars > 0 }) {
		t.Fatalf("no delivery: mode %d load %.1f truck %.1f", p.td.mode, l.Load, tr.Total())
	}
	if tr.Total() < 30 {
		t.Fatalf("truck got only %.1f", tr.Total())
	}
}

func TestToddlerButtonDumps(t *testing.T) {
	g := NewGame()
	g.setToddler(true)
	p := g.players[1]
	l := p.Driving.(*Loader)
	l.Load, l.LoadType, p.td.mode = 60, MatDirt, tdCarry
	p.td.button = true
	run(g, 1)
	p.td.button = false
	least := l.Load
	for i := 0; i < 240; i++ {
		run(g, 1)
		least = math.Min(least, l.Load)
	}
	if least > 2 || p.td.mode != tdDig {
		t.Fatalf("button should dump the bucket: load %.1f mode %d", least, p.td.mode)
	}
	if p.td.stars != 0 {
		t.Fatal("dumping on the ground is not a delivery")
	}
}

func TestToddlerNoExit(t *testing.T) {
	g := NewGame()
	g.setToddler(true)
	for _, p := range g.players {
		v := p.Driving
		g.toggleVehicle(p)
		if p.Driving != v {
			t.Fatalf("player %d climbed out in toddler mode", p.ID)
		}
	}
	g.setToddler(false)
	if g.exit(g.players[0]) == false {
		t.Fatal("grown-up mode should allow climbing out again")
	}
}

func TestToddlerWalksOverToPickedMachine(t *testing.T) {
	g := NewGame()
	g.setToddler(true)
	p := g.players[0]
	tr := p.td.own[1].(*Truck)
	g.toddlerPick(p, tr)
	walked := false
	for i := 0; i < 60*12 && p.Driving != tr; i++ {
		run(g, 1)
		walked = walked || (p.Driving == nil && p.W.moving > 0.5)
	}
	if p.Driving != tr {
		t.Fatalf("kid never got into the truck; at %v, truck at %v", p.W.Pos, tr.Pos)
	}
	if !walked {
		t.Fatal("kid should climb out and walk over")
	}
	if g.loader.Occupied {
		t.Fatal("the loader should be empty now")
	}
}

func TestToddlerTruckTipsWithOneButton(t *testing.T) {
	g := NewGame()
	g.setToddler(true)
	p := g.players[0]
	tr := p.td.own[1].(*Truck)
	g.put(p, tr)
	p.td.mode = tdCarry
	tr.Contents[MatSand] = 200
	p.td.pressed = true
	run(g, 60*8)
	if tr.Total() > 20 {
		t.Fatalf("truck still holds %.1f", tr.Total())
	}
	if p.td.stars != 1 {
		t.Fatalf("emptying the truck should give a star, got %d", p.td.stars)
	}
	if tr.BedTilt > 0 {
		t.Fatalf("body should come back down, tilt %.2f", tr.BedTilt)
	}
}

// aimHook puts the crane's hook straight over a spot.
func aimHook(c *TowerCrane, at Vec) {
	d := at.Sub(c.Base)
	c.Slew, c.slewV, c.Trolley = d.Angle(), 0, d.Len()
	c.hook, c.hookV = c.trolleyPos(), Vec{}
}

func TestToddlerCraneLiftsOntoTruck(t *testing.T) {
	for _, pi := range []int{0, 1} {
		g := NewGame()
		g.setToddler(true)
		p := g.players[pi]
		c := p.td.own[2].(*TowerCrane)
		tr := p.td.own[1].(*Truck)
		g.put(p, c)
		p.td.mode = tdCarry
		var box *Cargo
		for _, o := range g.world.Cargo {
			if o.Kind == cargoCrate && o.Pos.Sub(c.Base).Len() < craneJibLen {
				box = o
			}
		}
		run(g, 60) // the hook goes up to travel height
		aimHook(c, box.Pos)
		p.td.pressed = true
		run(g, 60*5)
		if c.held != box {
			t.Fatalf("player %d: the button should hook the crate (hook y %.0f)", p.ID, c.HookY())
		}
		if box.Y < 60 {
			t.Fatalf("player %d: crate should go up, y %.0f", p.ID, box.Y)
		}
		aimHook(c, tr.Pos)
		p.td.pressed = true
		run(g, 60*5)
		if c.held != nil || p.td.stars != 1 {
			t.Fatalf("player %d: crate should be put on the truck: held %v stars %d", p.ID, c.held != nil, p.td.stars)
		}
		run(g, 60)
		if box.truck != tr {
			t.Fatalf("player %d: crate is not riding on the truck (y %.0f)", p.ID, box.Y)
		}
	}
}
