package main

import (
	"image/color"
	"math"
)

// Walker is a player on foot: a kid site worker in an orange vest and a
// yellow hard hat, with blonde hair and brown eyes.
type Walker struct {
	Pos   Vec
	Yaw   float64
	Pitch float64 // first-person look pitch

	move, strafe, turn, look float64
	run                      bool

	vel    Vec
	phase  float64 // walk cycle
	moving float64 // 0 idle .. ~1.6 running
	marker color.RGBA
}

const (
	kidWalk   = 42.0
	kidRun    = 100.0
	kidEyeH   = 16.6
	kidRadius = 6.0
)

// steer applies this step's on-foot controls: forward/back, turn, strafe,
// look up/down and run.
func (k *Walker) steer(c Controls) {
	k.move, k.turn, k.strafe, k.look, k.run = c.Throttle, c.Steer, c.A4, c.A1, c.Brake
}

func (k *Walker) Colliders() []obstacle { return []obstacle{{k.Pos, kidRadius, k}} }

func (k *Walker) Update(dt float64, w *World) {
	k.Yaw += k.turn * 2.6 * dt
	k.Pitch = clamp(k.Pitch+k.look*1.3*dt, -1.2, 1.1)
	want := dir(k.Yaw).Scale(k.move).Add(dir(k.Yaw + math.Pi/2).Scale(k.strafe))
	if want.Len() > 1 {
		want = want.Norm()
	}
	speed := kidWalk
	if k.run {
		speed = kidRun
	}
	want = want.Scale(speed)
	dv := want.Sub(k.vel)
	if l := dv.Len(); l > 420*dt {
		dv = dv.Scale(420 * dt / l)
	}
	k.vel = k.vel.Add(dv)

	step := k.vel.Scale(dt)
	free := func(p Vec) bool {
		if w == nil {
			return true
		}
		hit, _ := w.CollideCircle(p, kidRadius, k)
		return !hit
	}
	switch {
	case free(k.Pos.Add(step)):
		k.Pos = k.Pos.Add(step)
	case free(k.Pos.Add(V(step.X, 0))):
		k.Pos.X += step.X
		k.vel.Y = 0
	case free(k.Pos.Add(V(0, step.Y))):
		k.Pos.Y += step.Y
		k.vel.X = 0
	default:
		k.vel = Vec{}
	}
	if w != nil {
		k.Pos.X = clamp(k.Pos.X, 20, w.Width-20)
		k.Pos.Y = clamp(k.Pos.Y, 20, w.Height-20)
	}
	sp := k.vel.Len()
	k.moving = approach(k.moving, clamp(sp/kidWalk, 0, 1.6), 6*dt)
	k.phase += sp * dt * 0.45
}

// Render builds the kid: big head, small body, swinging arms and legs.
func (k *Walker) Render(g *Game, s *Scene) {
	g.shadow(s, k.Pos, 12, 10, k.Yaw)
	if !g.cam.inView(ground(k.Pos, 10), 16) {
		return
	}
	f := dir(k.Yaw)
	r := dir(k.Yaw + math.Pi/2)
	bob := math.Abs(math.Sin(k.phase)) * 0.6 * math.Min(k.moving, 1)
	at := func(x, z, y float64) Vec3 { return ground(k.Pos.Add(f.Scale(x)).Add(r.Scale(z)), y+bob) }
	fw := rotY(Vec3{1, 0, 0}, k.Yaw)
	up := Vec3{0, 1, 0}
	limb := func(a, b Vec3, hh, hw float64, col color.RGBA) {
		d := b.Sub(a)
		s.addOBB2(a.Add(d.Scale(0.5)), d, fw, d.Len()/2+hh*0.4, hh, hw, col)
	}
	var (
		skin  = rgba(242, 198, 164, 255)
		jeans = rgba(62, 84, 136, 255)
		shirt = rgba(72, 96, 128, 255)
		shoe  = rgba(58, 44, 36, 255)
		vest  = rgba(255, 116, 18, 255)
		refl  = rgba(222, 226, 230, 255)
		hair  = rgba(238, 204, 112, 255)
		hat   = rgba(252, 206, 28, 255)
		brown = rgba(104, 60, 28, 255)
	)
	swing := math.Sin(k.phase) * 0.6 * math.Min(k.moving, 1.3)
	// legs and shoes
	for i, z := range []float64{-1.4, 1.4} {
		a := swing
		if i == 1 {
			a = -swing
		}
		hip := at(0, z, 7.5)
		foot := hip.Add(fw.Scale(math.Sin(a) * 7)).Add(up.Scale(-math.Cos(a) * 7))
		limb(hip, foot, 1.4, 1.3, jeans)
		s.addOBB2(foot.Add(fw.Scale(0.8)).Add(up.Scale(-0.2)), fw, up, 2, 0.8, 1.3, shoe)
	}
	// torso: shirt under an orange hi-vis vest with reflective bands
	s.addBox(at(0, 0, 10.6), 1.8, 3, 2.9, k.Yaw, shirt)
	s.addBox(at(0, 0, 10.9), 2.05, 2.6, 3.15, k.Yaw, vest)
	s.detail(func() {
		for _, y := range []float64{9.6, 11.6} {
			s.addBox(at(0, 0, y), 2.12, 0.35, 3.22, k.Yaw, refl)
		}
	})
	// arms swing opposite the legs
	for i, z := range []float64{-3.6, 3.6} {
		a := -swing * 0.8
		if i == 1 {
			a = swing * 0.8
		}
		sh := at(0, z, 13)
		hand := sh.Add(fw.Scale(math.Sin(a) * 5.4)).Add(up.Scale(-math.Cos(a) * 5.4))
		limb(sh, hand, 0.9, 0.9, shirt)
		s.addOBB2(hand, fw, up, 0.8, 0.8, 0.8, skin)
	}
	// head (big, as kids' heads are) with a little neck
	s.addBox(at(0, 0, 13.9), 0.8, 0.6, 0.8, k.Yaw, skin)
	head := at(0, 0, 14.4)
	s.addPrism(head, k.Yaw, []ring{{-2.2, 2.2, 2.2, 0, 0.9}, {-2.4, 2.4, 2.4, 1.6, 1.2}, {-2.4, 2.4, 2.4, 3.6, 1.2}, {-2.0, 2.0, 2.0, 4.6, 1.2}}, skin)
	s.detail(func() {
		// brown eyes with a white surround, a nose and a smile
		for _, z := range []float64{-1.0, 1.0} {
			s.addBox(at(2.42, z, 16.3), 0.08, 0.5, 0.55, k.Yaw, rgba(250, 250, 248, 255))
			s.addBox(at(2.5, z, 16.25), 0.08, 0.36, 0.34, k.Yaw, brown)
			s.addBox(at(2.55, z, 16.2), 0.06, 0.14, 0.13, k.Yaw, rgba(30, 20, 14, 255))
		}
		s.addBox(at(2.5, 0, 15.6), 0.35, 0.4, 0.3, k.Yaw, shadeC(skin, 0.95))
		s.addBox(at(2.45, 0, 14.9), 0.06, 0.18, 0.75, k.Yaw, rgba(176, 82, 76, 255))
		// blonde hair: fringe under the brim, sides, and a tuft at the back
		s.addBox(at(2.2, 0, 17.95), 0.4, 0.35, 2.1, k.Yaw, hair)
		for _, z := range []float64{-2.45, 2.45} {
			s.addBox(at(0.2, z, 17.2), 1.9, 1.0, 0.12, k.Yaw, hair)
		}
		s.addBox(at(-2.45, 0, 16.9), 0.15, 1.6, 2.2, k.Yaw, hair)
		s.addBox(at(-2.3, 0, 15.2), 0.3, 0.6, 1.4, k.Yaw, shadeC(hair, 0.92))
	})
	// yellow hard hat: brim, dome, front peak and a ridge
	s.addCylinder(at(0.2, 0, 18.35), 3.2, 0.45, 14, hat)
	s.addCone(at(0, 0, 19.4), 2.7, 1.8, 1.9, 14, hat)
	s.addBox(at(0.1, 0, 20.4), 1.2, 0.3, 1.4, k.Yaw, shadeC(hat, 0.95))
	s.addBox(at(3.3, 0, 18.4), 0.9, 0.22, 1.9, k.Yaw, hat)
	s.addBox(at(0, 0, 20.1), 2.4, 0.35, 0.4, k.Yaw, shadeC(hat, 1.06))
	// a coloured flag above the other player's head in split screen
	if k.marker.A != 0 {
		s.addBillboard(at(0, 0, 25), 1.6, k.marker)
	}
}
