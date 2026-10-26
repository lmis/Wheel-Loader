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
