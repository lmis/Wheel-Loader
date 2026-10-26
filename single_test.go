package main

import (
	"math"
	"testing"
)

func TestSingleScreenArrowGoesThatWay(t *testing.T) {
	g := NewGame()
	g.setSingle(companionsOwl)
	if len(g.players) != 1 || len(g.animals) != 2 {
		t.Fatalf("single screen: one kid and two animals, got %d and %d", len(g.players), len(g.animals))
	}
	for _, a := range g.animals {
		if a.Kind != animalOwl && a.Kind != animalChimp {
			t.Fatal("the owl and the chimp come along by default")
		}
	}
	p := g.players[0]
	l := p.Driving.(*Loader)
	l.Joint, l.Heading = V(1250, 1750), singleYaw // open ground, facing up the screen
	start := l.Joint
	p.td.want = dir(singleYaw + math.Pi/2) // the right arrow: screen right
	run(g, 60*4)
	moved := l.Joint.Sub(start)
	if moved.Len() < 150 || moved.Norm().Dot(dir(singleYaw+math.Pi/2)) < 0.6 {
		t.Fatalf("holding right should drive to the right of the screen, moved %v", moved)
	}
	if math.Abs(angleDiff(l.FrontHeading(), singleYaw+math.Pi/2)) > 0.5 {
		t.Fatalf("it should end up facing right, heading %.2f", l.FrontHeading())
	}
}

func TestSingleScreenSpaceNextMachineAndDogs(t *testing.T) {
	g := NewGame()
	g.setSingle(companionsDogs)
	for _, a := range g.animals {
		if a.Kind != animalGolden && a.Kind != animalBernese {
			t.Fatal("-doggo brings the dogs")
		}
	}
	p := g.players[0]
	for k := 1; k <= 4; k++ {
		g.nextMachine(p)
		want := p.td.own[k%4]
		for i := 0; i < 60*12 && p.Driving != want; i++ {
			run(g, 1)
		}
		if p.Driving != want {
			t.Fatalf("space %d should take the kid to %T", k, want)
		}
	}
}

func TestSingleScreenSpaceTipsTheTruck(t *testing.T) {
	g := NewGame()
	g.setSingle(companionsOwl)
	p := g.players[0]
	tr := p.td.own[1].(*Truck)
	g.put(p, tr)
	p.td.mode = tdCarry
	tr.Contents[MatSand] = 300
	run(g, 60)
	if tr.BedTilt > 0 {
		t.Fatal("the truck should not tip by itself")
	}
	p.td.pressed = true // space
	run(g, 60*8)
	if tr.Total() > 30 {
		t.Fatalf("space should tip the truck out, still has %.0f", tr.Total())
	}
}

func TestSingleScreenSpaceDigsOnce(t *testing.T) {
	g := NewGame()
	g.setSingle(companionsOwl)
	p := g.players[0]
	e := p.td.own[3].(*Excavator)
	g.put(p, e)
	e.Pos, e.Heading = V(1450, 900), math.Pi/2 // facing the sand heap
	p.td.pressed = true                        // space
	got := 0.0
	for i := 0; i < 60*8; i++ {
		run(g, 1)
		got = math.Max(got, e.Load)
	}
	if got < 30 {
		t.Fatalf("space in front of a heap should dig, got %.0f", got)
	}
}

func TestSingleScreenSpaceScoopsAndTips(t *testing.T) {
	g := NewGame()
	g.setSingle(companionsOwl)
	p := g.players[0]
	l := p.Driving.(*Loader)
	l.Load, l.LoadType = 40, MatSand
	p.td.pressed = true // space: up with the bucket
	run(g, 60)
	if p.td.mode != tdCarry || l.Lift < 0.4 {
		t.Fatalf("space should lift the bucket (mode %d, lift %.2f)", p.td.mode, l.Lift)
	}
	p.td.pressed = true // space again: tip it out
	least := l.Load
	for i := 0; i < 60*3; i++ {
		run(g, 1)
		least = math.Min(least, l.Load)
	}
	if least > 2 {
		t.Fatalf("space again should tip the bucket out, load %.1f", least)
	}
}

func TestSingleScreenCraneSpacePicksUp(t *testing.T) {
	g := NewGame()
	g.setSingle(companionsOwl)
	p := g.players[0]
	c := p.td.own[2].(*TowerCrane)
	g.put(p, c)
	p.td.mode = tdCarry
	near := V(1150, 1260).Sub(c.Base) // near a pile of loads
	c.Slew, c.slewV, c.Trolley = near.Angle(), 0, near.Len()
	c.hook = c.trolleyPos()
	run(g, 60*2)
	if c.held != nil {
		t.Fatal("in single screen the crane waits for space")
	}
	p.td.pressed = true
	for i := 0; i < 60*5 && c.held == nil; i++ {
		run(g, 1)
	}
	if c.held == nil {
		t.Fatal("space near a load should pick it up")
	}
	run(g, 60)
	p.td.pressed = true
	for i := 0; i < 60*5 && c.held != nil; i++ {
		run(g, 1)
	}
	if c.held != nil {
		t.Fatal("space again should put it down")
	}
}
