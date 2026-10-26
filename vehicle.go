package main

import (
	"image/color"
	"math"
)

// Controls are the player's driving intents for one step. Each vehicle maps
// them onto its own machine; nil means nobody is at the controls.
type Controls struct {
	Throttle float64 // -1..1 (W/S)
	Steer    float64 // -1..1 (A/D)
	Brake    bool    // Space
	// Implement axes, each -1..1: R/F, T/G, Y/H and Z/X.
	A1, A2, A3, A4 float64
}

// gauge is one bar in the HUD status panel.
type gauge struct {
	Label string
	Value float64 // 0..1
	Text  string  // optional text after the bar
}

// Vehicle is anything the player can climb into and drive.
type Vehicle interface {
	Name() string
	Center() Vec
	Facing() float64   // heading the chase camera sits behind
	SpeedVal() float64 // signed speed in world units/s
	Drive(c *Controls) // apply controls (nil when unoccupied)
	Update(dt float64, w *World)
	Colliders() []obstacle
	Seat() (eye Vec3, yaw float64) // first-person eye point and forward yaw
	ExitPoint(side float64) Vec    // where the driver steps out (side ±1)
	EnterRadius() float64
	SetOccupied(bool)
	Render(g *Game, s *Scene, cockpit bool) // cockpit: player sits inside in first person
	Effects(w *World)                       // dust and exhaust
	Jolt() float64                          // recent impact strength, for camera shake
	ControlHelp() []string
	Gauges() []gauge
	CamDist() float64 // preferred chase-camera distance
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

func (l *Loader) Name() string {
	if l.Label != "" {
		return l.Label
	}
	return "Wheel loader"
}

func (l *Loader) Center() Vec          { return l.Joint }
func (l *Loader) Facing() float64      { return l.Heading }
func (l *Loader) SpeedVal() float64    { return l.Speed }
func (l *Loader) EnterRadius() float64 { return 95 }
func (l *Loader) SetOccupied(b bool)   { l.Occupied = b }
func (l *Loader) Jolt() float64        { return math.Abs(l.bounceV) }
func (l *Loader) CamDist() float64     { return 520 }

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
	return ground(l.Joint.Sub(dir(rh).Scale(13)), 50+clamp(l.bounce*0.3, -3, 3)), rh
}

func (l *Loader) ExitPoint(side float64) Vec {
	rh := l.RearHeading()
	return l.Joint.Sub(dir(rh).Scale(14)).Add(dir(rh - side*math.Pi/2).Scale(bodyWidth/2 + 24))
}

func (l *Loader) Render(g *Game, s *Scene, cockpit bool) {
	rh, fh := l.RearHeading(), l.FrontHeading()
	g.shadow(s, l.RearAxle().Sub(dir(rh).Scale(6)), rearLen+34, bodyWidth+16, rh)
	g.shadow(s, l.FrontAxle(), frontLen+18, bodyWidth+16, fh)
	g.shadow(s, l.BucketPivot().Add(dir(fh).Scale(16)), 34, bodyWidth+28, fh)
	g.buildLoader3D(s, l)
	if cockpit {
		// dashboard and steering wheel in front of the seat
		at := func(x, z, y float64) Vec3 { return ground(l.Joint, 0).Add(rotY(Vec3{x, y, z}, rh)) }
		s.addBox(at(-6, 0, 38), 2.5, 5, 12, rh, colDark)
		s.addSteeringWheel(at(-7, 0, 43), rh, 4, rgba(30, 30, 33, 255))
	}
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

func (l *Loader) ControlHelp() []string {
	return []string{
		"{T}\taccelerate / reverse",
		"{S}\tarticulate left / right",
		"{B}\tbrake",
		"{1}\traise / lower arms",
		"{2}\tcurl / dump bucket",
	}
}

func (l *Loader) Gauges() []gauge {
	pos := "RAISED"
	if l.Lift < 0.28 {
		pos = "GROUND"
	}
	tilt := "DUMP"
	if l.Tilt > 0.6 {
		tilt = "CARRY"
	} else if l.Tilt > 0.35 {
		tilt = "LEVEL"
	}
	load := "empty"
	if l.Load >= 1 {
		load = matNames[l.LoadType]
	}
	return []gauge{
		{"BUCKET", l.LoadFrac(), load},
		{"LIFT", l.Lift, pos},
		{"TILT", l.Tilt, tilt},
	}
}

// loaderColor returns the body colour, defaulting to construction yellow.
func (l *Loader) bodyColor() color.RGBA {
	if l.Color.A != 0 {
		return l.Color
	}
	return rgba(240, 186, 24, 255)
}
