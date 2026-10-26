package main

import "math"

// Camera is a perspective chase/orbit camera. It looks at a ground target
// (the loader) from a distance, at an adjustable yaw and pitch. The yaw
// smoothly swings to sit behind the loader, with a user-controlled offset.
type Camera struct {
	Target  Vec     // ground point being looked at (sim X,Y)
	Dist    float64 // current distance from target
	TgtDist float64
	Yaw     float64 // current orbit yaw (radians)
	Pitch   float64 // elevation angle (radians)
	UserYaw float64 // player orbit offset added to the chase yaw

	ScreenW, ScreenH float64
	fov              float64
	lookHeight       float64

	// derived each frame
	eye            Vec3
	fwd, right, up Vec3
	f, aspect      float64

	shake          float64
	shakeX, shakeY float64
	t              float64
	followYaw      float64 // smoothed chase yaw (before user offset)
}

func NewCamera(sw, sh float64) *Camera {
	return &Camera{
		ScreenW:    sw,
		ScreenH:    sh,
		Dist:       520,
		TgtDist:    520,
		Pitch:      0.72,
		fov:        1.05,
		lookHeight: 40,
	}
}

// Follow eases the camera target onto the loader and swings the chase yaw to
// sit behind the machine's heading.
func (c *Camera) Follow(target Vec, heading float64, dt float64) {
	k := 1 - math.Pow(0.0008, dt)
	c.Target = c.Target.Add(target.Sub(c.Target).Scale(k))
	// smoothly swing the chase yaw to follow the loader's heading, then sit
	// the camera on the opposite side so we look in the travel direction.
	dd := angleDiff(heading, c.followYaw)
	c.followYaw += dd * (1 - math.Pow(0.04, dt))
	c.Yaw = c.followYaw + math.Pi + c.UserYaw
	c.Dist = lerp(c.Dist, c.TgtDist, 1-math.Pow(0.01, dt))

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

// setup recomputes the camera basis for this frame.
func (c *Camera) setup() {
	c.aspect = c.ScreenW / c.ScreenH
	c.f = 1 / math.Tan(c.fov/2)
	if c.Pitch < 0.2 {
		c.Pitch = 0.2
	}
	if c.Pitch > 1.35 {
		c.Pitch = 1.35
	}
	tgt := Vec3{c.Target.X, c.lookHeight, c.Target.Y}
	dir := Vec3{
		math.Cos(c.Pitch) * math.Cos(c.Yaw),
		math.Sin(c.Pitch),
		math.Cos(c.Pitch) * math.Sin(c.Yaw),
	}
	c.eye = tgt.Add(dir.Scale(c.Dist))
	c.fwd = tgt.Sub(c.eye).Norm()
	c.right = c.fwd.Cross(Vec3{0, 1, 0}).Norm()
	c.up = c.right.Cross(c.fwd)
}

// view transforms a world point into camera space (Z is forward depth).
func (c *Camera) view(p Vec3) Vec3 {
	d := p.Sub(c.eye)
	return Vec3{d.Dot(c.right), d.Dot(c.up), d.Dot(c.fwd)}
}

// project maps a view-space point to screen coordinates.
func (c *Camera) project(v Vec3) (float64, float64) {
	ndcx := (v.X * c.f / c.aspect) / v.Z
	ndcy := (v.Y * c.f) / v.Z
	sx := c.ScreenW/2 + ndcx*c.ScreenW/2 + c.shakeX
	sy := c.ScreenH/2 - ndcy*c.ScreenH/2 + c.shakeY
	return sx, sy
}

// VisibleBounds returns an approximate world rectangle in view, used for the
// minimap and material culling. It is a generous box around the target.
func (c *Camera) VisibleBounds() (min, max Vec) {
	r := c.Dist * 1.8
	return Vec{c.Target.X - r, c.Target.Y - r}, Vec{c.Target.X + r, c.Target.Y + r}
}
