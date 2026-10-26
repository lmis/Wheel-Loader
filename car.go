package main

import "math"

// Car is a site pickup built from a Kenney model: quick and light, for getting
// around the site rather than for work.
type Car struct {
	Pos      Vec
	Angle    float64
	Speed    float64
	Steer    float64
	Model    string // embedded model name
	Occupied bool

	throttle, steerCmd float64
	brake              bool
	rpm, jolt          float64
}

const (
	carScale     = 30.0 // Kenney model units -> world
	carWheelbase = 40.0
	carRearX     = -20.0
	carMaxSteer  = 0.6
	carMaxSpeed  = 340.0
	carMaxRev    = 120.0
)

func (c *Car) Center() Vec           { return c.Pos }
func (c *Car) Facing() float64       { return c.Angle }
func (c *Car) SpeedVal() float64     { return c.Speed }
func (c *Car) SetOccupied(b bool)    { c.Occupied = b }
func (c *Car) Jolt() float64         { return c.jolt }
func (c *Car) CamDist() float64      { return 380 }
func (c *Car) Colliders() []obstacle { return c.collidersAt(c.Pos, c.Angle) }

func (c *Car) collidersAt(p Vec, ang float64) []obstacle {
	f := dir(ang)
	return []obstacle{{p.Add(f.Scale(20)), 23, c}, {p.Sub(f.Scale(20)), 23, c}}
}

func (c *Car) Drive(ctl *Controls) {
	if ctl == nil {
		c.throttle, c.steerCmd, c.brake = 0, 0, true
		return
	}
	c.throttle, c.steerCmd, c.brake = ctl.Throttle, ctl.Steer, ctl.Brake
}

func (c *Car) local(x, z, y float64) Vec3 {
	return ground(c.Pos.Add(dir(c.Angle).Scale(x)).Add(dir(c.Angle+math.Pi/2).Scale(z)), y)
}

func (c *Car) ExitPoint(side float64) Vec {
	return c.Pos.Add(dir(c.Angle + math.Pi/2).Scale(-side * 40))
}

func (c *Car) Update(dt float64, w *World) {
	if c.steerCmd != 0 {
		c.Steer += c.steerCmd * 2.2 * dt
	} else {
		c.Steer = approach(c.Steer, 0, 2.8*dt)
	}
	c.Steer = clamp(c.Steer, -carMaxSteer, carMaxSteer)
	if c.throttle != 0 {
		c.Speed += c.throttle * 340 * dt
		c.rpm = approach(c.rpm, 0.5+math.Abs(c.throttle)*0.5, 3*dt)
	} else {
		c.rpm = approach(c.rpm, 0.15, 2*dt)
	}
	if c.brake {
		c.Speed = approach(c.Speed, 0, 800*dt)
	}
	c.Speed -= c.Speed * 0.6 * dt
	if c.throttle == 0 && !c.brake {
		c.Speed = approach(c.Speed, 0, 60*dt)
	}
	c.Speed = clamp(c.Speed, -carMaxRev, carMaxSpeed)

	rear := c.Pos.Add(dir(c.Angle).Scale(carRearX + c.Speed*dt))
	ang := c.Angle + c.Speed*math.Tan(c.Steer)/carWheelbase*dt
	np := rear.Sub(dir(ang).Scale(carRearX))
	if w != nil {
		if hit, n := collideCircles(w, c.collidersAt(np, ang), c); hit {
			np, ang = c.Pos.Add(n.Scale(2)), c.Angle
			c.jolt = math.Max(c.jolt, math.Abs(c.Speed)*0.5+30)
			c.Speed *= -0.25
		}
		np.X = clamp(np.X, 40, w.Width-40)
		np.Y = clamp(np.Y, 40, w.Height-40)
	}
	c.Pos, c.Angle = np, ang
	c.jolt *= math.Pow(0.02, dt)
}

func (c *Car) Effects(w *World) {
	if spd := math.Abs(c.Speed); spd > 40 {
		w.spawnDriveDust(c.Pos.Sub(dir(c.Angle).Scale(22)), c.Angle, clamp(spd/carMaxSpeed, 0, 1)*0.6)
	}
}

func (c *Car) Render(g *Game, s *Scene) {
	if !g.cam.inView(ground(c.Pos, 20), 70) {
		return
	}
	if m, err := loadModel(c.Model); err == nil {
		// Kenney vehicles face +Z
		s.addMesh(m, ground(c.Pos, 0), c.Angle-math.Pi/2, carScale)
	}
}
