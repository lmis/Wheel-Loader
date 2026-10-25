package main

import (
	"math"
	"testing"
)

func step(l *Loader, w *World, n int) {
	for i := 0; i < n; i++ {
		l.Update(fixedDT, w)
		if w != nil {
			w.Update(fixedDT)
		}
	}
}

func TestInertia(t *testing.T) {
	l := &Loader{Heading: 0}
	l.throttle = 1
	step(l, nil, 30) // 0.5s accel
	if l.Speed <= 0 {
		t.Fatalf("expected positive speed after throttle, got %f", l.Speed)
	}
	fast := l.Speed
	l.throttle = 0
	step(l, nil, 1)
	if l.Speed >= fast {
		// should coast but not instantly stop
	}
	// coasting should not stop instantly
	step(l, nil, 5)
	if l.Speed <= 0 {
		t.Fatalf("machine stopped too abruptly (no inertia): %f", l.Speed)
	}
}

func TestArticulatedTurning(t *testing.T) {
	l := &Loader{Heading: 0}
	h0 := l.Heading
	// bend the joint fully and drive forward
	for i := 0; i < 120; i++ {
		l.throttle = 1
		l.steer = 1
		l.Update(fixedDT, nil)
	}
	if math.Abs(angleDiff(l.Heading, h0)) < 0.3 {
		t.Fatalf("articulated steering did not change heading enough: dh=%f", angleDiff(l.Heading, h0))
	}
	if l.Artic <= 0.1 {
		t.Fatalf("articulation angle did not build up: %f", l.Artic)
	}
	// front and rear halves must splay (visible articulation)
	if math.Abs(angleDiff(l.FrontHeading(), l.RearHeading())) < 0.1 {
		t.Fatalf("front/rear not articulated: %f vs %f", l.FrontHeading(), l.RearHeading())
	}
}

func TestNoTurnWhenStationary(t *testing.T) {
	l := &Loader{Heading: 0}
	h0 := l.Heading
	for i := 0; i < 120; i++ {
		l.steer = 1 // bend joint but no throttle
		l.Update(fixedDT, nil)
	}
	if math.Abs(angleDiff(l.Heading, h0)) > 0.05 {
		t.Fatalf("machine turned while stationary (should only articulate): dh=%f", angleDiff(l.Heading, h0))
	}
	if l.Artic < 0.5 {
		t.Fatalf("expected joint to bend while stationary: %f", l.Artic)
	}
}

func TestScoopRemovesFromPile(t *testing.T) {
	w := NewWorld()
	// place loader just in front of the dirt stockpile, facing it
	zone := w.Zones[0] // dirt
	start := zone.Pos.Add(V(0, 150))
	l := &Loader{Joint: start, Heading: -math.Pi / 2, Tilt: 0.3, Lift: 0.15}

	before := totalMaterial(w)
	// drive into the pile while curling the bucket
	for i := 0; i < 180; i++ {
		l.throttle = 0.6
		l.tiltCmd = 1
		l.liftCmd = -1
		l.Update(fixedDT, w)
		w.Update(fixedDT)
	}
	if l.Load < 20 {
		t.Fatalf("bucket did not scoop meaningfully: load=%f", l.Load)
	}
	after := totalMaterial(w)
	if after >= before {
		t.Fatalf("pile did not lose material: before=%f after=%f", before, after)
	}
	if l.LoadType != MatDirt {
		t.Fatalf("expected dirt load, got %s", matNames[l.LoadType])
	}
}

func TestDumpIntoTruck(t *testing.T) {
	w := NewWorld()
	truck := w.Trucks[0]
	l := &Loader{Joint: truck.Pos.Add(dir(truck.Angle).Scale(70)), Tilt: 0.9, Lift: 0.8, Load: 80, LoadType: MatGravel}
	l.Heading = truck.Pos.Sub(l.Joint).Angle()
	// nudge so the bucket pivot sits over the bed
	for truck.Total() < 1 && l.Load > 1 {
		l.tiltCmd = -1
		l.liftCmd = 0.2
		l.Update(fixedDT, w)
		w.Update(fixedDT)
		// guard against infinite loop
		if l.Tilt <= 0 && l.Load > 1 && truck.Total() < 1 {
			break
		}
	}
	if truck.Total() < 1 {
		t.Logf("loader pos=%v tilt=%f load=%f", l.Joint, l.Tilt, l.Load)
		t.Fatalf("truck received no material")
	}
	if truck.Contents[MatGravel] < 1 {
		t.Fatalf("truck did not receive gravel")
	}
}

func TestAutopilotCompletesCycle(t *testing.T) {
	g := NewGame()
	g.auto = NewAutopilot()
	// run ~40 simulated seconds and ensure the loader both loads and delivers
	maxLoad := 0.0
	delivered := false
	for i := 0; i < 60*60; i++ {
		g.auto.Apply(g)
		g.loader.Update(fixedDT, g.world)
		g.world.Update(fixedDT)
		g.jobs.Update(fixedDT)
		if g.loader.Load > maxLoad {
			maxLoad = g.loader.Load
		}
		for _, tr := range g.world.Trucks {
			if tr.Total() > 20 {
				delivered = true
			}
		}
	}
	if maxLoad < 30 {
		t.Fatalf("autopilot never loaded the bucket (max=%f)", maxLoad)
	}
	if !delivered {
		t.Fatalf("autopilot never delivered material to a truck")
	}
}

func totalMaterial(w *World) float64 {
	var s float64
	for i := range w.Grid.cells {
		s += w.Grid.cells[i].Amount
	}
	return s
}
