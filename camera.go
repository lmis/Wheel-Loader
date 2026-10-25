package main

import "math"

// Camera follows a target smoothly and supports zoom and a little shake.
type Camera struct {
	Pos     Vec // world point at screen centre
	Zoom    float64
	TgtZoom float64
	shake   float64
	shakeX  float64
	shakeY  float64
	ScreenW float64
	ScreenH float64
	t       float64
}

func NewCamera(sw, sh float64) *Camera {
	return &Camera{Zoom: 1.0, TgtZoom: 1.0, ScreenW: sw, ScreenH: sh}
}

func (c *Camera) Follow(target Vec, lead Vec, dt float64) {
	// aim slightly ahead in the direction of travel
	aim := target.Add(lead)
	k := 1 - math.Pow(0.0015, dt)
	c.Pos = c.Pos.Add(aim.Sub(c.Pos).Scale(k))
	c.Zoom = lerp(c.Zoom, c.TgtZoom, 1-math.Pow(0.01, dt))
	c.t += dt
	if c.shake > 0.001 {
		c.shake *= math.Pow(0.02, dt)
		c.shakeX = math.Sin(c.t*53) * c.shake
		c.shakeY = math.Cos(c.t*61) * c.shake
	} else {
		c.shakeX, c.shakeY = 0, 0
	}
}

func (c *Camera) Kick(amount float64) {
	if amount > c.shake {
		c.shake = amount
	}
}

// WorldToScreen converts a world point to screen coordinates.
func (c *Camera) WorldToScreen(p Vec) Vec {
	d := p.Sub(c.Pos).Scale(c.Zoom)
	return V(d.X+c.ScreenW/2+c.shakeX, d.Y+c.ScreenH/2+c.shakeY)
}

// ScreenToWorld is the inverse transform.
func (c *Camera) ScreenToWorld(p Vec) Vec {
	d := V(p.X-c.ScreenW/2-c.shakeX, p.Y-c.ScreenH/2-c.shakeY).Scale(1 / c.Zoom)
	return d.Add(c.Pos)
}

// VisibleBounds returns the world-space rectangle currently on screen.
func (c *Camera) VisibleBounds() (min, max Vec) {
	halfW := c.ScreenW / 2 / c.Zoom
	halfH := c.ScreenH / 2 / c.Zoom
	min = V(c.Pos.X-halfW, c.Pos.Y-halfH)
	max = V(c.Pos.X+halfW, c.Pos.Y+halfH)
	return
}
