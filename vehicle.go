package main

import (
	"image/color"
	"math"
)

// Controls are the driving intents for one step. Each vehicle maps them
// onto its own machine; nil means nobody is at the controls. The kids only
// steer and drive; the toddler assistants (toddler.go) work the rest.
type Controls struct {
	Throttle float64 // -1..1
	Steer    float64 // -1..1
	Brake    bool    // brake; run when on foot
	// Implement axes, each -1..1 (see each machine's Drive).
	A1, A2, A3, A4 float64
}

// Vehicle is a machine on the site; the kids climb into the loaders, trucks
// and cranes, the rest stand parked.
type Vehicle interface {
	Center() Vec
	Facing() float64   // heading the chase camera sits behind
	SpeedVal() float64 // signed speed in world units/s
	Drive(c *Controls) // apply controls (nil when unoccupied)
	Update(dt float64, w *World)
	Colliders() []obstacle
	ExitPoint(side float64) Vec // where the driver steps out (side ±1)
	SetOccupied(bool)
	Render(g *Game, s *Scene)
	Effects(w *World) // dust and exhaust
	Jolt() float64    // recent impact strength, for camera shake
	CamDist() float64 // preferred chase-camera distance
}

// seated is a vehicle with a driver's seat to look out of in the cab view.
type seated interface {
	Seat() (eye Vec3, yaw float64) // eye point and forward yaw
	SeatPitch() float64            // how far down the driver looks
	RenderCockpit(s *Scene)        // what the driver sees inside the cab
}

// collideCircles tests a set of circles against the world (ignoring self)
// and returns the push-out normal of the first hit.
func collideCircles(w *World, circles []obstacle, self any) (bool, Vec) {
	for _, c := range circles {
		if hit, n := w.CollideCircle(c.c, c.r, self); hit {
			return true, n
		}
	}
	return false, Vec{}
}

// ---- the wheel loader as a Vehicle ----

func (l *Loader) Center() Vec        { return l.Joint }
func (l *Loader) Facing() float64    { return l.Heading }
func (l *Loader) SpeedVal() float64  { return l.Speed }
func (l *Loader) SetOccupied(b bool) { l.Occupied = b }
func (l *Loader) Jolt() float64      { return math.Abs(l.bounceV) }
func (l *Loader) CamDist() float64   { return 290 }

func (l *Loader) Drive(c *Controls) {
	if c == nil {
		l.throttle, l.steer, l.liftCmd, l.tiltCmd = 0, 0, 0, 0
		l.brake = true
		return
	}
	l.throttle, l.steer, l.brake = c.Throttle, c.Steer, c.Brake
	l.liftCmd, l.tiltCmd = c.A1, c.A2
}

// collidersAt returns the loader's body circles for a given joint position:
// front axle, joint and the rear axle/counterweight. The bucket is left out
// so it can reach over truck beds.
func (l *Loader) collidersAt(j Vec) []obstacle {
	fh, rh := l.FrontHeading(), l.RearHeading()
	return []obstacle{
		{j.Add(dir(fh).Scale(frontLen)), 27, l},
		{j, 24, l},
		{j.Sub(dir(rh).Scale(rearLen + 8)), 28, l},
	}
}

func (l *Loader) Colliders() []obstacle { return l.collidersAt(l.Joint) }

// Seat is the operator's eye: in the cab on the rear frame, looking along it.
func (l *Loader) Seat() (Vec3, float64) {
	rh := l.RearHeading()
	return ground(l.Joint.Sub(dir(rh).Scale(16)), 51+clamp(l.bounce*0.3, -3, 3)), rh
}

func (l *Loader) SeatPitch() float64 { return -0.34 } // down over the hood to the bucket

// RenderCockpit adds the inside of the cab: dashboard and steering wheel,
// the windscreen frame and pillars, the roof lining and a mirror.
func (l *Loader) RenderCockpit(s *Scene) {
	rh := l.RearHeading()
	at := func(x, z, y float64) Vec3 { return ground(l.Joint, 0).Add(rotY(Vec3{x, y, z}, rh)) }
	trim := rgba(44, 46, 52, 255)
	fw := rotY(Vec3{1, 0, 0}, rh)
	up := Vec3{0, 1, 0}
	// dashboard under the windscreen, with two gauges
	s.addRoundPrism(at(0, 0, 0), rh, []ring{{-7.5, -4.6, 12.5, 31, 1.2}, {-7, -4.8, 12.5, 37, 1.2}}, colDark, false, true)
	s.detail(func() {
		for _, z := range []float64{-3, 3} {
			s.addEllipsoid(at(-7.05, z, 35.6), fw, up, 0.12, 0.75, 0.75, rgba(20, 20, 24, 255))
			s.glowing(func() { s.addEllipsoid(at(-7.15, z, 35.6), fw, up, 0.08, 0.5, 0.5, rgba(90, 170, 210, 255)) })
		}
	})
	s.addSteeringWheel(at(-9.5, 0, 40), rh, 2.6, l.Artic*2.2, rgba(30, 30, 33, 255))
	// windscreen frame: pillars, header and sill
	for _, z := range []float64{-11.5, 11.5} {
		s.addCapsule(at(-4.6, z, 37), at(-5.2, z, 54), 0.9, 0.9, trim)
	}
	s.addCapsule(at(-5.2, -11.5, 53.6), at(-5.2, 11.5, 53.6), 0.9, 0.9, trim)
	s.addCapsule(at(-4.6, -11.5, 37.2), at(-4.6, 11.5, 37.2), 0.6, 0.6, trim)
	// roof lining and a small mirror
	s.addOBB2(at(-15, 0, 53.3), fw, up, 10.5, 0.4, 13.5, rgba(88, 90, 96, 255))
	s.addEllipsoid(at(-6.2, 0, 52.4), fw, up, 0.3, 0.7, 1.8, trim)
}

func (l *Loader) ExitPoint(side float64) Vec {
	rh := l.RearHeading()
	return l.Joint.Sub(dir(rh).Scale(14)).Add(dir(rh - side*math.Pi/2).Scale(bodyWidth/2 + 24))
}

func (l *Loader) Render(g *Game, s *Scene) {
	g.buildLoader3D(s, l)
}

func (l *Loader) Effects(w *World) {
	spd := math.Abs(l.Speed)
	if spd > 15 {
		inten := clamp(spd/maxSpeed, 0, 1) * 0.8
		ra := l.RearAxle()
		rh := l.RearHeading()
		w.spawnDriveDust(ra.Add(dir(rh+math.Pi/2).Scale(bodyWidth/2)), l.Heading, inten)
		w.spawnDriveDust(ra.Add(dir(rh-math.Pi/2).Scale(bodyWidth/2)), l.Heading, inten)
	}
	if l.Occupied || spd > 5 {
		// exhaust stack on the engine hood (matches buildLoader3D)
		stack := l.Joint.Sub(dir(l.RearHeading()).Scale(36)).Add(dir(l.RearHeading() + math.Pi/2).Scale(7))
		w.spawnExhaust(stack, l.engineRPM)
	}
}

// Each kind of machine has its own colour, used on the machine and on its
// tile in the HUD: yellow loaders, red dump trucks, green cranes.
var (
	colLoader = rgba(240, 186, 24, 255)
	colTruck  = rgba(206, 46, 38, 255)
	colCrane  = rgba(48, 158, 74, 255)
	colExc    = rgba(236, 120, 28, 255) // excavators: orange
)

// bodyColor returns the loader's colour, defaulting to loader yellow.
func (l *Loader) bodyColor() color.RGBA {
	if l.Color.A != 0 {
		return l.Color
	}
	return colLoader
}
