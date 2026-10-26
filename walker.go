package main

import (
	"image/color"
	"math"
)

// Walker is a player on foot: a kid site worker in an orange vest and a
// yellow hard hat, with blonde hair and brown eyes.
type Walker struct {
	Pos Vec
	Yaw float64

	move, turn float64
	run        bool

	vel    Vec
	phase  float64 // walk cycle
	moving float64 // 0 idle .. ~1.6 running
	marker color.RGBA
}

const (
	kidWalk   = 42.0
	kidRun    = 100.0
	kidRadius = 6.0
)

// steer applies this step's on-foot controls: forward/back, turn and run.
func (k *Walker) steer(c Controls) {
	k.move, k.turn, k.run = c.Throttle, c.Steer, c.Brake
}

func (k *Walker) Colliders() []obstacle { return []obstacle{{k.Pos, kidRadius, k}} }

func (k *Walker) Update(dt float64, w *World) {
	k.Yaw += k.turn * 2.6 * dt
	want := dir(k.Yaw).Scale(k.move)
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

// Render builds the kid from smooth, rounded parts: big head, small body,
// swinging arms and legs.
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
	var (
		skin  = rgba(242, 198, 164, 255)
		jeans = rgba(62, 84, 136, 255)
		shirt = rgba(72, 96, 128, 255)
		shoe  = rgba(58, 44, 36, 255)
		vest  = rgba(255, 116, 18, 255)
		refl  = rgba(204, 210, 216, 255)
		hair  = rgba(238, 204, 112, 255)
		hat   = rgba(252, 206, 28, 255)
		brown = rgba(104, 60, 28, 255)
	)
	swing := math.Sin(k.phase) * 0.6 * math.Min(k.moving, 1.3)
	// legs: rounded thigh and shin with a bent knee, and round-toed shoes
	for i, z := range []float64{-1.35, 1.35} {
		a := swing
		if i == 1 {
			a = -swing
		}
		hip := at(0, z, 7.6)
		knee := hip.Add(fw.Scale(math.Sin(a)*3.6 + 0.3)).Add(up.Scale(-math.Cos(a) * 3.6))
		foot := knee.Add(fw.Scale(math.Sin(a*0.6) * 3.4)).Add(up.Scale(-math.Cos(a*0.6) * 3.4))
		s.addCapsule(hip, knee, 1.45, 1.3, jeans)
		s.addCapsule(knee, foot, 1.3, 1.15, jeans)
		s.addEllipsoid(foot.Add(fw.Scale(0.8)).Add(up.Scale(-0.2)), fw, up, 2.1, 0.9, 1.35, shoe)
	}
	// trunk: the hi-vis vest is the torso's outer shape, rounded at the
	// corners, tapering to the shoulders, with the shirt at the collar
	o := at(0, 0, 0)
	torso := []ring{
		{-1.75, 1.75, 2.6, 7.6, 1.4},
		{-1.95, 1.95, 2.85, 9.5, 1.5},
		{-1.95, 1.95, 2.9, 11.8, 1.5},
		{-1.6, 1.6, 2.5, 13.3, 1.4},
	}
	s.addRoundPrism(o, k.Yaw, torso, vest, true, true)
	s.addCapsule(at(0, 0, 13.2), at(0, 0, 14.6), 1.05, 0.95, skin) // neck
	s.addEllipsoid(at(0, 0, 13.45), fw, up, 1.35, 0.4, 1.5, shirt)  // collar
	s.detail(func() {
		// V-neck: the shirt showing between the vest's front panels
		s.addFaceOutward([]Vec3{at(1.62, -0.85, 13.2), at(1.62, 0.85, 13.2), at(1.96, 0, 11.9)}, at(0, 0, 10.5), shirt)
		// two reflective tape bands hugging the vest
		for _, y := range []float64{9.1, 10.5} {
			s.addRoundPrism(o, k.Yaw, []ring{
				{-1.97, 1.97, 2.9, y, 1.5},
				{-1.97, 1.97, 2.92, y + 0.42, 1.5},
			}, refl, false, false)
		}
		// braces: tape running up over the shoulders, front and back
		for _, z := range []float64{-1.45, 1.45} {
			for _, sx := range []float64{1, -1} {
				a, b := at(sx*1.97, z, 10.9), at(sx*1.63, z, 13.3)
				d := b.Sub(a)
				s.addOBB2(a.Add(d.Scale(0.5)), d, fw.Scale(sx), d.Len()/2, 0.04, 0.28, refl)
			}
			s.addBox(at(0, z, 13.34), 1.6, 0.04, 0.28, k.Yaw, refl)
		}
	})
	// arms swing opposite the legs, elbows a little bent
	for i, z := range []float64{-3.25, 3.25} {
		a := -swing * 0.8
		if i == 1 {
			a = swing * 0.8
		}
		sh := at(0, z, 12.7)
		elbow := sh.Add(fw.Scale(math.Sin(a) * 2.8)).Add(up.Scale(-math.Cos(a) * 2.8))
		hand := elbow.Add(fw.Scale(math.Sin(a+0.35) * 2.6)).Add(up.Scale(-math.Cos(a+0.35) * 2.6))
		s.addSphere(sh, 1.15, vest) // shoulder of the vest
		s.addCapsule(sh, elbow, 0.95, 0.85, shirt)
		s.addCapsule(elbow, hand, 0.85, 0.8, shirt)
		s.addSphere(hand, 0.85, skin)
	}
	// head: big and round, as kids' heads are
	const hx, hy, hz = 2.45, 2.55, 2.4 // radii: forward, up, side
	hc := at(0, 0, 16.6)
	s.addEllipsoid(hc, fw, up, hx, hy, hz, skin)
	// onFace is a point on the head's surface (dz to the side, dy up from the
	// centre), pushed out by out
	onFace := func(dz, dy, out float64) Vec3 {
		x := hx * math.Sqrt(math.Max(0, 1-(dz/hz)*(dz/hz)-(dy/hy)*(dy/hy)))
		return at(x+out, dz, 16.6+dy)
	}
	for _, z := range []float64{-2.35, 2.35} {
		s.addEllipsoid(at(0.1, z, 16.4), fw, up, 0.55, 0.75, 0.35, shadeC(skin, 0.96)) // ears
	}
	// blonde hair: a cap over the back and sides, and a fringe
	s.addEllipsoid(at(-0.5, 0, 17.2), fw, up, 2.6, 2.35, 2.62, hair)
	s.addEllipsoid(at(1.5, 0, 18.1), fw, up, 1.0, 0.65, 2.0, hair)
	s.detail(func() {
		// brown eyes with a white surround, a nose and a smile
		for _, z := range []float64{-0.95, 0.95} {
			e := onFace(z, -0.15, 0)
			s.addEllipsoid(e, fw, up, 0.22, 0.52, 0.46, rgba(250, 250, 248, 255))
			s.addEllipsoid(e.Add(fw.Scale(0.12)), fw, up, 0.16, 0.36, 0.33, brown)
			s.addEllipsoid(e.Add(fw.Scale(0.22)), fw, up, 0.1, 0.15, 0.14, rgba(30, 20, 14, 255))
		}
		s.addEllipsoid(onFace(0, -0.9, 0.05), fw, up, 0.45, 0.42, 0.38, shadeC(skin, 0.95)) // nose
		s.addEllipsoid(onFace(0, -1.65, -0.05), fw, up, 0.18, 0.16, 0.8, rgba(176, 82, 76, 255))
	})
	// yellow hard hat: a round dome on a wide brim, a front peak and a ridge
	s.addCone(at(0.2, 0, 18.55), 3.25, 3.15, 0.4, 24, hat)
	s.addEllipsoid(at(0, 0, 18.7), fw, up, 2.75, 2.0, 2.75, hat)
	s.addEllipsoid(at(2.9, 0, 18.6), fw, up, 1.0, 0.25, 2.0, hat)
	s.addEllipsoid(at(0.1, 0, 20.55), fw, up, 2.3, 0.35, 0.45, shadeC(hat, 1.06))
	// a coloured flag above the other player's head in split screen
	if k.marker.A != 0 {
		s.addBillboard(at(0, 0, 25), 1.6, k.marker)
	}
}
