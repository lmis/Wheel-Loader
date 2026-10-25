package main

import "math"

// Autopilot plays the loader through the full job loop. It doubles as test
// instrumentation: it exercises driving, articulated steering, scooping and
// dumping without a human, and is used to produce screenshots headlessly.
type Autopilot struct {
	state   int
	timer   float64
	target  Vec
	truck   *Truck
	Verbose bool
	log     func(string)
}

const (
	apToPile = iota
	apDig
	apToTruck
	apDump
)

func NewAutopilot() *Autopilot { return &Autopilot{state: apToPile} }

func (a *Autopilot) Apply(g *Game) {
	l := g.loader
	w := g.world
	j := g.jobs.Current

	// reset intents
	l.throttle, l.steer, l.liftCmd, l.tiltCmd, l.brake = 0, 0, 0, 0, false
	a.timer += fixedDT

	switch a.state {
	case apToPile:
		// pick the pile matching the current job material
		a.target = a.pileFor(w, j.Material)
		a.driveTo(l, a.target, 70)
		dist := l.Joint.Sub(a.target).Len()
		if dist < 230 {
			// close to the pile: lower and level the bucket to dig
			l.liftCmd = -1
			if l.Tilt > 0.35 {
				l.tiltCmd = -1
			}
		} else {
			// carry the bucket up while travelling so we don't scrape debris
			if l.Lift < 0.5 {
				l.liftCmd = 1
			}
			if l.Tilt < 0.85 {
				l.tiltCmd = 1
			}
		}
		if dist < 120 && l.Lift < 0.3 {
			a.state = apDig
			a.timer = 0
		}
	case apDig:
		// creep into the pile while curling the bucket to scoop
		l.throttle = 0.5
		l.liftCmd = -1
		l.tiltCmd = 1 // curl back
		if l.Load > bucketCap*0.85 || a.timer > 3.0 {
			a.state = apToTruck
			a.timer = 0
		}
	case apToTruck:
		a.truck = j.Truck
		// raise and carry
		if l.Lift < 0.78 {
			l.liftCmd = 1
		}
		if l.Tilt < 0.9 {
			l.tiltCmd = 1
		}
		a.driveTo(l, a.dumpSpot(a.truck), 70)
		if l.Joint.Sub(a.dumpSpot(a.truck)).Len() < 150 {
			a.state = apDump
			a.timer = 0
		}
	case apDump:
		// keep the bucket high, aim squarely at the truck and creep the
		// bucket over the bed before tipping.
		if l.Lift < 0.78 {
			l.liftCmd = 1
		}
		to := a.truck.Pos.Sub(l.Joint)
		err := angleDiff(to.Angle(), l.Heading)
		l.steer = clamp(err*2.5, -1, 1)
		overBed := g.world.TruckAt(l.BucketTip()) != nil || g.world.TruckAt(l.BucketPivot()) != nil
		switch {
		case overBed || to.Len() < 95:
			l.brake = true
			l.tiltCmd = -1 // dump forward
		case math.Abs(err) < 0.5:
			l.throttle = 0.3 // creep straight in
		default:
			l.throttle = 0.25 // creep while aligning
		}
		if l.Load < 2 || a.timer > 6 {
			a.state = apToPile
			a.timer = 0
		}
	}
}

// pileFor returns the loading-zone position for a material type.
func (a *Autopilot) pileFor(w *World, mat int) Vec {
	for _, z := range w.Zones {
		if z.Type == mat {
			return z.Pos
		}
	}
	return w.Zones[0].Pos
}

// dumpSpot returns a point just in front of the truck's bed.
func (a *Autopilot) dumpSpot(t *Truck) Vec {
	return t.Pos.Add(dir(t.Angle).Scale(90))
}

// driveTo steers the articulated loader toward p and sets throttle.
func (a *Autopilot) driveTo(l *Loader, p Vec, stopDist float64) {
	to := p.Sub(l.Joint)
	d := to.Len()
	desired := to.Angle()
	err := angleDiff(desired, l.Heading)

	if d < stopDist {
		l.brake = true
		return
	}
	// steer proportional to heading error. Articulated loaders arc forward
	// to come around rather than reversing, so we always drive forward and
	// let the bend bring the nose onto the target.
	l.steer = clamp(err*2.4, -1, 1)
	sp := 1 - math.Min(0.65, math.Abs(err)*0.45)
	if d < stopDist*2.5 {
		sp *= 0.55
	}
	l.throttle = clamp(sp, 0.35, 1)
}
