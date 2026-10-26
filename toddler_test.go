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
	if len(g.players) != 2 {
		t.Fatal("the game should start with two players")
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
	if !toddlerDrive(g, p, tr.Pos, false, 1500, func() bool { return tr.Total() > 30 && l.Load < 2 }) {
		t.Fatalf("no delivery: mode %d load %.1f truck %.1f", p.td.mode, l.Load, tr.Total())
	}
}

func TestToddlerButtonDumps(t *testing.T) {
	g := NewGame()
	p := g.players[1]
	l := p.Driving.(*Loader)
	tr := p.td.own[1].(*Truck)
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
	if tr.Total() > 0 {
		t.Fatal("the load should go on the ground, not in the truck")
	}
}

func TestToddlerWalksOverToPickedMachine(t *testing.T) {
	g := NewGame()
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
		if c.held != nil {
			t.Fatalf("player %d: the button should put the crate down", p.ID)
		}
		run(g, 60)
		if box.truck != tr {
			t.Fatalf("player %d: crate is not riding on the truck (y %.0f)", p.ID, box.Y)
		}
	}
}

func TestAnimalsComeAlong(t *testing.T) {
	g := NewGame()
	owners := map[int][]int{1: {animalOwl, animalChimp}, 2: {animalGolden, animalBernese}}
	for _, a := range g.animals {
		if a.ride != a.owner.Driving || a.Y < 40 {
			t.Fatalf("animal %d should start on its kid's loader (y %.0f)", a.Kind, a.Y)
		}
		found := false
		for _, k := range owners[a.owner.ID] {
			found = found || k == a.Kind
		}
		if !found {
			t.Fatalf("animal %d belongs to the wrong kid", a.Kind)
		}
	}
	want := map[*Player]Vehicle{}
	for i, p := range g.players {
		want[p] = p.td.own[1+i] // player 1 to the truck, player 2 to the crane
		g.toddlerPick(p, want[p])
	}
	walkedAlong := map[*Animal]bool{}
	for i := 0; i < 60*12; i++ {
		run(g, 1)
		for _, a := range g.animals {
			p := a.owner
			if p.Driving == nil && a.ride == nil && !a.hopping && p.W.Pos.Sub(a.Pos).Len() < 70 {
				walkedAlong[a] = true
			}
			if a.Kind == animalOwl && p.Driving == nil && !a.hopping && a.Y < 15 {
				t.Fatal("the owl should fly while its kid walks")
			}
		}
	}
	for _, a := range g.animals {
		if !walkedAlong[a] {
			t.Fatalf("animal %d did not keep up with its kid", a.Kind)
		}
		if a.ride != a.owner.Driving || a.owner.Driving != want[a.owner] {
			t.Fatalf("animal %d should be sitting on its kid's new machine", a.Kind)
		}
	}
	// and they ride along on a moving machine
	p := g.players[1]
	g.toddlerPick(p, p.td.own[1])
	for i := 0; i < 60*12 && p.Driving != p.td.own[1]; i++ {
		run(g, 1)
	}
	run(g, 120)
	tr := p.Driving.(*Truck)
	p.ctl = Controls{Throttle: 1}
	for i := 0; i < 60; i++ {
		run(g, 1)
	}
	for _, a := range g.animals[2:] {
		spot, y, _ := a.rideSpot(tr)
		if a.ride != tr || a.Pos.Sub(spot).Len() > 1 || a.Y != y {
			t.Fatalf("dog %d should ride on the truck's roof", a.Kind)
		}
	}
}
