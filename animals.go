package main

import (
	"image/color"
	"math"
	"math/rand"
)

// Animal kinds. Each kid has two companions: player 1 the owl and the
// chimp, player 2 the two dogs.
const (
	animalGolden  = iota // golden retriever
	animalBernese        // Bernese mountain dog
	animalChimp          // small chimpanzee
	animalOwl            // snowy owl
)

// Animal is a kid's companion. It rides along on the kid's machine; when the
// kid climbs out to switch machines it hops down, runs (or, the owl, flies)
// alongside, and hops up onto the next machine with them.
type Animal struct {
	Kind  int
	Pos   Vec
	Y     float64 // height of its feet above the ground
	Yaw   float64
	owner *Player
	slot  int     // 0 or 1: which riding spot and place in the line
	ride  Vehicle // machine it is sitting on, if any

	// a hop (or the owl's glide) from one spot to another
	hopping            bool
	hopFrom, hopTo     Vec
	hopY0, hopY1, hopT float64
	hopLen             float64
	hopRide            Vehicle // machine it lands on, if any

	flying  bool    // the owl, in the air beside its kid
	bounce  float64 // height of a little happy hop on the spot
	jumpT   float64 // time into that hop (<0: waiting for the next one)
	t       float64 // idle animation clock
	phase   float64 // gait cycle
	speed   float64
	headYaw float64 // relative to the body
	blink   float64
	rng     *rand.Rand
}

func newAnimal(kind int, owner *Player, slot int, seed int64) *Animal {
	return &Animal{Kind: kind, owner: owner, slot: slot, rng: rand.New(rand.NewSource(seed)), jumpT: -float64(seed)}
}

const jumpLen = 0.42 // seconds a happy hop takes

// hopAbout makes the animal bounce up and down every couple of seconds while
// it sits or stands still.
func (a *Animal) hopAbout(dt float64) {
	if a.hopping || a.speed > 20 {
		a.bounce = 0
		return
	}
	a.jumpT += dt
	if a.jumpT < 0 {
		a.bounce = 0
		return
	}
	k := a.jumpT / jumpLen
	if k >= 1 {
		a.bounce, a.jumpT = 0, -(1.2 + a.rng.Float64()*2.5)
		return
	}
	a.bounce = 9 * 4 * k * (1 - k)
}

// rideSpot is where the animal sits on machine v: its ground position,
// height and facing.
func (a *Animal) rideSpot(v Vehicle) (Vec, float64, float64) {
	on := func(o Vec, yaw, x, z, y float64) (Vec, float64, float64) {
		return o.Add(dir(yaw).Scale(x)).Add(dir(yaw + math.Pi/2).Scale(z)), y, yaw
	}
	switch m := v.(type) {
	case *Loader:
		rh := m.RearHeading()
		if a.slot == 0 {
			return on(m.Joint, rh, -15, 0, 58) // cab roof
		}
		return on(m.Joint, rh, -52, 0, 45) // engine hood
	case *Truck:
		return on(m.Pos, m.Angle, 39, []float64{-12, 12}[a.slot], 55) // cab roof
	case *TowerCrane:
		// on the concrete footing, either side of the mast
		return on(m.Base, math.Pi/2, 22, []float64{-22, 22}[a.slot], 8)
	}
	return v.Center(), 0, v.Facing()
}

// hop starts a jump (or glide) to a spot, landing on machine onto if set.
func (a *Animal) hop(to Vec, y float64, onto Vehicle) {
	a.hopping, a.hopFrom, a.hopTo, a.hopY0, a.hopY1, a.hopT = true, a.Pos, to, a.Y, y, 0
	a.hopLen = 0.35 + to.Sub(a.Pos).Len()/320
	a.hopRide, a.ride = onto, nil
}

// Update moves the companion along with its kid.
func (a *Animal) Update(dt float64) {
	a.t += dt
	a.blink -= dt
	if a.blink < -4-a.rng.Float64()*4 {
		a.blink = 0.1
	}
	p := a.owner
	kid := p.W.Pos
	if p.Driving != nil {
		kid = p.Driving.Center()
	}
	// head turns toward its kid, or looks about while riding
	want := math.Sin(a.t*0.7+float64(a.Kind)) * 0.5
	if a.ride == nil && kid.Sub(a.Pos).Len() > 4 {
		want = angleDiff(kid.Sub(a.Pos).Angle(), a.Yaw)
	}
	lim := 1.1
	if a.Kind == animalOwl {
		lim = 2.6 // owls can turn their heads a very long way
	}
	want = clamp(want, -lim, lim)
	a.headYaw += (want - a.headYaw) * (1 - math.Pow(0.03, dt))

	if a.hopping {
		a.hopT += dt
		k := math.Min(1, a.hopT/a.hopLen)
		to := a.hopTo
		if a.hopRide != nil {
			to, a.hopY1, _ = a.rideSpot(a.hopRide) // the machine may be moving
		}
		a.Pos = a.hopFrom.Add(to.Sub(a.hopFrom).Scale(k))
		arc := 10 + to.Sub(a.hopFrom).Len()*0.15
		a.Y = lerp(a.hopY0, a.hopY1, k) + arc*4*k*(1-k)
		if d := to.Sub(a.hopFrom); d.Len() > 1 {
			a.Yaw += angleDiff(d.Angle(), a.Yaw) * (1 - math.Pow(0.001, dt))
		}
		a.speed = 0
		if k >= 1 {
			a.hopping, a.ride = false, a.hopRide
			a.flying = a.Kind == animalOwl && a.ride == nil
		}
		return
	}

	defer a.hopAbout(dt)
	switch {
	case a.ride != nil && a.ride == p.Driving:
		// sitting on the kid's machine
		a.Pos, a.Y, a.Yaw = a.rideSpot(a.ride)
		a.speed = 0
	case p.Driving != nil:
		// the kid got in another machine: catch up, then hop on
		spot, y, _ := a.rideSpot(p.Driving)
		if d := spot.Sub(a.Pos).Len(); d < 80 || d > 400 || a.ride != nil {
			a.hop(spot, y, p.Driving)
			return
		}
		a.runTo(spot, dt)
	case a.ride != nil:
		// the kid climbed out: hop down beside them
		land := p.W.Pos.Sub(dir(p.W.Yaw).Scale(24 + 24*float64(a.slot)))
		y := 0.0
		if a.Kind == animalOwl {
			y = 28
		}
		a.hop(land, y, nil)
	default:
		a.runTo(a.followPoint(), dt)
	}
}

// followPoint is where the companion wants to be while its kid walks: just
// behind the kid, one on each side (the owl flies on the right).
func (a *Animal) followPoint() Vec {
	k := a.owner.W
	side := []float64{1, -1}[a.slot]
	if a.Kind == animalOwl {
		return k.Pos.Add(dir(k.Yaw + math.Pi/2).Scale(28)).Sub(dir(k.Yaw).Scale(4))
	}
	// a point a little way back along the kid's path, stepped to the side
	back := 22.0
	trail := a.owner.trail
	at := k.Pos
	for i := len(trail) - 1; i >= 0 && back > 0; i-- {
		d := at.Sub(trail[i]).Len()
		if d >= back {
			at = at.Add(trail[i].Sub(at).Scale(back / d))
			break
		}
		back -= d
		at = trail[i]
	}
	return at.Add(dir(k.Yaw + math.Pi/2).Scale(side * 22))
}

// runTo moves toward a point at up to a sprint, easing off as it arrives.
func (a *Animal) runTo(to Vec, dt float64) {
	d := to.Sub(a.Pos)
	goal := clamp(d.Len()*5, 0, 160)
	if d.Len() > 2 {
		a.Yaw += angleDiff(d.Angle(), a.Yaw) * (1 - math.Pow(0.002, dt))
	}
	a.speed = approach(a.speed, goal, 400*dt)
	step := math.Min(a.speed*dt, d.Len())
	if d.Len() > 0.01 {
		a.Pos = a.Pos.Add(d.Norm().Scale(step))
	}
	a.phase += a.speed * dt * 0.5
	if a.flying {
		a.Y = 28 + math.Sin(a.t*3)*2
	} else {
		a.Y = 0
	}
}

// animalScale enlarges the animals a little beyond life size, so they are
// easy to spot on the machines from the chase cameras.
const animalScale = 2.0

func (a *Animal) Render(g *Game, s *Scene) {
	if a.ride == nil {
		r := animalScale / (1 + a.Y/30) // smaller, softer shadow when up in the air
		g.blobShadow(s, a.Pos, 9*r, 5.5*r, a.Yaw)
	}
	if !g.cam.inView(ground(a.Pos, a.Y+12), 36) {
		return
	}
	mark := len(s.tris)
	defer s.scaleSince(mark, ground(a.Pos, a.Y), animalScale)
	switch a.Kind {
	case animalGolden:
		a.renderDog(s, rgba(222, 168, 82, 255), rgba(240, 200, 128, 255), rgba(196, 140, 64, 255), false, 1)
	case animalBernese:
		a.renderDog(s, rgba(26, 24, 26, 255), rgba(246, 244, 240, 255), rgba(170, 84, 40, 255), true, 1.12)
	case animalChimp:
		a.renderChimp(s)
	case animalOwl:
		a.renderOwl(s)
	}
}

func (a *Animal) frame() (func(x, z, y float64) Vec3, Vec3, Vec3) {
	f, r := dir(a.Yaw), dir(a.Yaw+math.Pi/2)
	at := func(x, z, y float64) Vec3 {
		return ground(a.Pos.Add(f.Scale(x)).Add(r.Scale(z)), a.Y+a.bounce/animalScale+y)
	}
	return at, rotY(Vec3{1, 0, 0}, a.Yaw), rotY(Vec3{0, 0, 1}, a.Yaw)
}

// renderDog draws a retriever-shaped dog from rounded parts. Bernese: black
// coat with a white chest and muzzle and rust markings (accent) on the legs
// and cheeks.
func (a *Animal) renderDog(s *Scene, coat, light, accent color.RGBA, bernese bool, k float64) {
	at, fw, lat := a.frame()
	up := Vec3{0, 1, 0}
	moving := clamp(a.speed/30, 0, 1.5)
	legCol, paw, chest, muzzle := coat, coat, coat, shadeC(coat, 1.08)
	if bernese {
		legCol, paw, chest, muzzle = accent, light, light, light
	}
	// legs with a trot cycle (diagonal pairs together): front legs bend at
	// the knee, back legs at the hock, and end in round paws
	for i, p := range [][2]float64{{4.4, -1.9}, {4.4, 1.9}, {-4.3, -1.9}, {-4.3, 1.9}} {
		front := i < 2
		sw := math.Sin(a.phase*2+float64(i%2)*math.Pi+float64(i/2)*math.Pi) * 0.45 * moving
		top := at(p[0]*k, p[1]*k, 7.0*k)
		lift := math.Max(0, -math.Sin(a.phase*2+float64(i%2)*math.Pi+float64(i/2)*math.Pi)) * moving // leg swinging forward
		bend := 0.9 * k
		if !front {
			bend = -1.3 * k // the hock points backward
		}
		mid := top.Add(fw.Scale(math.Sin(sw)*3.0*k + bend*(0.4+lift*0.6))).Add(up.Scale(-3.0 * k))
		foot := top.Add(fw.Scale(math.Sin(sw) * 5.2 * k)).Add(up.Scale(-6.0*k + lift*1.2*k))
		thigh := 1.3 * k
		if !front {
			thigh = 1.55 * k
		}
		s.addCapsule(top, mid, thigh, 1.0*k, coat)
		s.addCapsule(mid, foot, 0.95*k, 0.8*k, legCol)
		s.addEllipsoid(foot.Add(fw.Scale(0.5*k)).Add(up.Scale(-0.2*k)), fw, up, 1.2*k, 0.6*k, 0.95*k, paw)
	}
	// body: one smooth loft from the rump past rounded hips and a slight
	// waist to a deep chest
	body := []loftStation{
		{-7.5, 8.6, 0.5, 0.5}, {-6.8, 8.4, 2.1, 2.1}, {-5.0, 8.1, 2.8, 2.6}, {-2.2, 7.9, 2.6, 2.5},
		{1.2, 7.8, 2.9, 2.9}, {4.0, 7.5, 2.8, 3.1}, {5.9, 8.0, 2.2, 2.4}, {6.9, 8.5, 0.6, 0.7},
	}
	for i := range body {
		body[i].x *= k
		body[i].y *= k
		body[i].hw *= k
		body[i].hh *= k
	}
	s.addLoft(at(0, 0, 0), fw, up, body, coat)
	if bernese {
		s.addEllipsoid(at(4.9*k, 0, 7.3*k), fw, up, 1.9*k, 2.6*k, 2.2*k, chest) // white chest
	}
	// tail: thick at the base, curving up, wagging
	wag := math.Sin(a.t*11) * 0.7
	p := at(-6.2*k, 0, 9.0*k)
	for i, r := range []float64{1.15, 0.95, 0.75, 0.55} {
		lift := 0.25 + 0.3*float64(i)
		d := fw.Scale(-math.Cos(wag) * 0.8).Add(lat.Scale(math.Sin(wag) * 0.8)).Add(up.Scale(lift)).Norm()
		next := p.Add(d.Scale(1.25 * k))
		col := coat
		if bernese && i == 3 {
			col = light // white tip
		}
		s.addCapsule(p, next, r*k, (r-0.18)*k, col)
		p = next
	}
	// head on a neck, turning toward its kid
	hy := a.Yaw + a.headYaw
	hf := rotY(Vec3{1, 0, 0}, hy)
	hl := rotY(Vec3{0, 0, 1}, hy)
	neck := at(5.8*k, 0, 9.2*k)
	hc := neck.Add(hf.Scale(2.2 * k)).Add(up.Scale(2.6 * k))
	s.addCapsule(neck, hc, 2.0*k, 1.7*k, coat)
	s.addEllipsoid(hc, hf, up, 2.3*k, 2.0*k, 2.05*k, coat)
	s.addEllipsoid(hc.Add(hf.Scale(2.7*k)).Add(up.Scale(-0.7*k)), hf, up, 1.7*k, 1.05*k, 1.15*k, muzzle)
	s.detail(func() {
		s.addSphere(hc.Add(hf.Scale(4.35*k)).Add(up.Scale(-0.25*k)), 0.48*k, rgba(20, 16, 14, 255)) // nose
		for _, z := range []float64{-1, 1} {
			eye := hc.Add(hf.Scale(1.85 * k)).Add(hl.Scale(z * 0.95 * k)).Add(up.Scale(0.6 * k))
			s.addSphere(eye, 0.36*k, rgba(40, 26, 16, 255))
			if bernese {
				s.addEllipsoid(eye.Add(up.Scale(0.7*k)), hf, up, 0.3*k, 0.22*k, 0.42*k, accent) // rust brows
				s.addEllipsoid(hc.Add(hf.Scale(2.0*k)).Add(hl.Scale(z*1.55*k)).Add(up.Scale(-0.9*k)), hf, up, 0.5*k, 0.55*k, 0.35*k, accent)
			}
			// floppy ears hanging beside the head
			ear := hc.Add(hl.Scale(z * 1.95 * k)).Add(up.Scale(-0.95 * k)).Add(hf.Scale(-0.5 * k))
			s.addEllipsoid(ear, hf, up.Add(hl.Scale(z*0.3)), 1.0*k, 1.65*k, 0.42*k, shadeC(coat, 0.85))
		}
		if bernese {
			s.addEllipsoid(hc.Add(hf.Scale(1.9*k)).Add(up.Scale(1.2*k)), hf, up, 0.35*k, 1.1*k, 0.4*k, light) // blaze
		}
	})
}

func (a *Animal) renderChimp(s *Scene) {
	at, fw, _ := a.frame()
	up := Vec3{0, 1, 0}
	fur := rgba(46, 36, 30, 255)
	face := rgba(214, 176, 140, 255)
	moving := clamp(a.speed/20, 0, 1.5)
	hop := 0.0
	if a.speed < 5 && !a.hopping {
		hop = math.Max(0, math.Sin(a.t*2.1)-0.92) * 18 // the odd excited bounce
	}
	at2 := func(x, z, y float64) Vec3 { return at(x, z, y+hop) }
	// knuckle-walking stance: short legs, long arms
	for i, z := range []float64{-1.6, 1.6} {
		sw := math.Sin(a.phase*2+float64(i)*math.Pi) * 0.4 * moving
		hip := at2(-1.5, z, 4.6)
		foot := hip.Add(fw.Scale(math.Sin(sw) * 4.2)).Add(up.Scale(-math.Cos(sw) * 4.2))
		s.addCapsule(hip, foot, 1.15, 0.9, fur)
		s.addEllipsoid(foot.Add(fw.Scale(0.6)), fw, up, 1.25, 0.45, 0.8, face)
	}
	for i, z := range []float64{-2.6, 2.6} {
		sw := -math.Sin(a.phase*2+float64(i)*math.Pi) * 0.4 * moving
		sh := at2(1.8, z, 8.6)
		var hand Vec3
		if i == 1 && a.speed < 5 && math.Sin(a.t*0.9) > 0.6 {
			hand = at2(1.6, 1.8, 12.6) // scratching its head
		} else {
			hand = sh.Add(fw.Scale(1.5 + math.Sin(sw)*3)).Add(up.Scale(-8.2))
		}
		elbow := sh.Add(hand.Sub(sh).Scale(0.5)).Add(fw.Scale(-0.6))
		s.addCapsule(sh, elbow, 1.05, 0.95, fur)
		s.addCapsule(elbow, hand, 0.95, 0.8, fur)
		s.addEllipsoid(hand, fw, up, 0.85, 0.7, 0.8, face)
	}
	// body leaning forward, and a round head
	s.addEllipsoid(at2(0.1, 0, 7.0), fw.Add(up.Scale(0.35)), up, 2.7, 3.2, 2.8, fur)
	hy := a.Yaw + a.headYaw
	hf := rotY(Vec3{1, 0, 0}, hy)
	hl := rotY(Vec3{0, 0, 1}, hy)
	hc := at2(2.4, 0, 11.7)
	s.addEllipsoid(hc, hf, up, 2.1, 2.1, 2.15, fur)
	light := rgba(232, 196, 160, 255) // a light, friendly face
	s.detail(func() {
		// heart-shaped face: two rounds round the eyes and a big soft muzzle
		for _, z := range []float64{-1, 1} {
			s.addEllipsoid(hc.Add(hf.Scale(1.45)).Add(hl.Scale(z*0.62)).Add(up.Scale(0.3)), hf, up, 0.75, 0.95, 0.85, light)
		}
		s.addEllipsoid(hc.Add(hf.Scale(1.7)).Add(up.Scale(-0.85)), hf, up, 0.95, 1.0, 1.3, light)
		// a wide, upturned smile across the muzzle, with rosy cheeks
		var prev Vec3
		for i := 0; i <= 6; i++ {
			t := -1 + float64(i)/3
			p := hc.Add(hf.Scale(2.62 - 0.18*t*t)).Add(hl.Scale(t * 0.72)).Add(up.Scale(-1.2 + 0.32*t*t))
			if i > 0 {
				s.addCapsule(prev, p, 0.09, 0.09, rgba(84, 46, 38, 255))
			}
			prev = p
		}
		for _, z := range []float64{-1, 1} {
			s.addSpot(hc.Add(hf.Scale(2.1)).Add(hl.Scale(z*1.15)).Add(up.Scale(-0.55)), hf.Add(hl.Scale(z*0.5)), 0.4, 0.28, rgba(250, 150, 150, 255))
			// big sparkly eyes under raised brows
			e := hc.Add(hf.Scale(2.1)).Add(hl.Scale(z * 0.62)).Add(up.Scale(0.35))
			s.addEllipsoid(e, hf, up, 0.24, 0.55, 0.47, rgba(252, 250, 246, 255))
			s.addEllipsoid(e.Add(hf.Scale(0.12)), hf, up, 0.17, 0.42, 0.36, rgba(120, 72, 32, 255))
			s.addSphere(e.Add(hf.Scale(0.22)), 0.21, rgba(16, 10, 8, 255))
			s.glowing(func() {
				s.addSphere(e.Add(hf.Scale(0.36)).Add(up.Scale(0.16)).Add(hl.Scale(0.1)), 0.1, rgba(255, 255, 255, 255))
			})
			s.addEllipsoid(e.Add(up.Scale(0.85)).Add(hf.Scale(-0.05)), hf, up.Add(hl.Scale(z*0.3)), 0.2, 0.14, 0.5, shadeC(fur, 1.3)) // brow
			s.addEllipsoid(hc.Add(hl.Scale(z*2.15)).Add(up.Scale(0.25)), hf, up, 0.8, 0.9, 0.4, light)                                // ears
			s.addEllipsoid(hc.Add(hl.Scale(z*2.3)).Add(up.Scale(0.25)), hf, up, 0.5, 0.6, 0.2, shadeC(light, 0.88))
		}
	})
}

// renderOwl draws a snowy owl: one soft egg shape from feathered feet to
// a round head, a flat facial disc with big yellow eyes, folded wings
// wrapping the sides, and dark bars on the back and wings. In flight it
// leans forward on broad wings. It looks happy: raised brows, smiling eyes
// and rosy cheeks.
func (a *Animal) renderOwl(s *Scene) {
	at, fw, lat := a.frame()
	up := Vec3{0, 1, 0}
	white := rgba(248, 248, 244, 255)
	cream := rgba(236, 234, 226, 255) // wings, a touch darker than the body
	bar := rgba(70, 66, 60, 255)
	flying := a.flying
	lifting := !flying && (a.hopping || a.bounce > 0)

	mark := len(s.tris)
	// feathered feet with dark talons
	for _, z := range []float64{-1, 1} {
		foot := at(0.9, z*1.1, 0.55)
		s.addEllipsoid(foot, fw, up, 1.1, 0.6, 0.8, white)
		for _, t := range []float64{-0.45, 0, 0.45} {
			s.addEllipsoid(foot.Add(fw.Scale(1.0)).Add(lat.Scale(t)).Add(up.Scale(-0.35)), fw, up, 0.3, 0.15, 0.12, bar)
		}
	}
	// body: a plump egg, a ruff blending into the head, and a short tail
	const bx, by, bz = 3.3, 4.2, 3.5
	bc := at(0, 0, 4.7)
	s.addEllipsoid(bc, fw, up, bx, by, bz, white)
	s.addEllipsoid(at(0.25, 0, 7.7), fw, up, 3.0, 2.1, 3.3, white)
	s.addEllipsoid(at(-2.9, 0, 2.2), fw.Scale(-1).Add(up.Scale(-0.5)), up, 1.8, 0.5, 1.4, cream)
	// wings: folded along the sides, half lifted for a hop, spread to fly
	for _, z := range []float64{-1, 1} {
		side := lat.Scale(z)
		shoulder := at(0.3, z*2.9, 7.3)
		switch {
		case flying:
			flap := math.Sin(a.t*9) * 0.55
			span := side.Scale(math.Cos(flap)).Add(up.Scale(math.Sin(flap)))
			// broad wings, about two and a half times the owl's height across
			owlWing(s, shoulder.Add(span.Scale(1.2)), span, fw, cream, bar)
		default:
			tip := at(-2.8, z*1.6, 1.7)
			if lifting {
				tip = at(-1.9, z*5.2, 3.2)
			}
			d := tip.Sub(shoulder)
			mid := shoulder.Add(d.Scale(0.5)).Add(side.Scale(0.45))
			s.addEllipsoid(mid, d, side, d.Len()/2+0.6, 0.85, 2.5, cream)
			s.detail(func() {
				for k := 0; k < 5; k++ {
					p := shoulder.Add(d.Scale(0.25 + float64(k)*0.15)).Add(side.Scale(1.3))
					s.addSpot(p, side.Add(fw.Scale(-0.3)), 0.75, 0.3, bar)
				}
			})
		}
	}
	// a few bars across the back and sides
	s.detail(func() {
		for i := 0; i < 9; i++ {
			th := math.Pi*0.62 + float64(i%3)*math.Pi*0.38
			ph := -0.35 + float64(i/3)*0.4
			p, n := ellPoint(bc, fw, up, lat, bx, by, bz, th, ph)
			s.addSpot(p, n, 0.8, 0.28, bar)
		}
	})
	// in flight the body leans forward, level with the ground
	head := at(0.2, 0, 9.7)
	if flying {
		axis := up.Cross(fw)
		s.rotateSince(mark, bc, axis, 1.05)
		head = bc.Add(fw.Scale(4.4)).Add(up.Scale(1.3)) // at the front, looking ahead
	}

	// round head that swivels a long way
	hy := a.Yaw + a.headYaw
	hf := rotY(Vec3{1, 0, 0}, hy)
	hl := rotY(Vec3{0, 0, 1}, hy)
	s.addEllipsoid(head, hf, up, 2.9, 2.5, 3.1, white)
	s.detail(func() {
		disc := head.Add(hf.Scale(2.3)).Add(up.Scale(-0.2))
		s.addEllipsoid(disc, hf, up, 0.55, 1.95, 2.4, rgba(255, 255, 252, 255))
		for _, z := range []float64{-1, 1} {
			e := disc.Add(hf.Scale(0.32)).Add(hl.Scale(z * 1.05)).Add(up.Scale(0.35))
			// raised brows and smiling eyes: the lower lids push up into a
			// happy squint
			s.addEllipsoid(e.Add(up.Scale(1.05)).Add(hl.Scale(z*0.15)), hf, up.Add(hl.Scale(z*0.3)), 0.45, 0.3, 0.9, white)
			s.addEllipsoid(e.Add(hf.Scale(-0.06)), hf, up, 0.25, 0.92, 0.92, rgba(40, 36, 30, 255)) // dark rim
			if a.blink > 0 {
				s.addEllipsoid(e.Add(hf.Scale(0.05)), hf, up, 0.25, 0.88, 0.88, white) // eyelid
				continue
			}
			s.addEllipsoid(e, hf, up, 0.32, 0.8, 0.8, rgba(250, 202, 24, 255))
			s.addSphere(e.Add(hf.Scale(0.22)).Add(up.Scale(0.1)), 0.5, rgba(14, 12, 10, 255))
			s.glowing(func() {
				s.addSphere(e.Add(hf.Scale(0.62)).Add(up.Scale(0.35)).Add(hl.Scale(0.18)), 0.17, rgba(255, 255, 255, 255))
			})
			s.addEllipsoid(e.Add(hf.Scale(0.2)).Add(up.Scale(-0.72)), hf, up.Add(hl.Scale(-z*0.15)), 0.4, 0.48, 1.0, rgba(255, 255, 252, 255))
			// rosy cheeks
			s.addSpot(e.Add(hf.Scale(0.1)).Add(up.Scale(-1.1)).Add(hl.Scale(z*0.55)), hf, 0.55, 0.32, rgba(255, 168, 176, 255))
		}
		s.addEllipsoid(disc.Add(hf.Scale(0.5)).Add(up.Scale(-0.6)), hf.Add(up.Scale(-0.6)), up, 0.45, 0.5, 0.28, rgba(46, 44, 42, 255)) // beak
	})
}

// owlWing adds a spread wing from root along span: a thin, tapering slab
// with a straight leading edge, a scalloped trailing edge of flight feathers
// and dark bars across them.
func owlWing(s *Scene, root, span, fwd Vec3, col, bar color.RGBA) {
	span = span.Norm()
	chord := fwd.Sub(span.Scale(fwd.Dot(span))).Norm() // toward the leading edge
	n := chord.Cross(span).Norm()
	if n.Y < 0 {
		n = n.Scale(-1)
	}
	// (along the span, across it) for matching leading and trailing points
	lead := [][2]float64{{0, 1.7}, {2.2, 2.0}, {4.4, 2.0}, {6.6, 1.8}, {8.8, 1.4}, {11, 0.9}, {12.8, 0.3}, {14, -0.4}}
	trail := [][2]float64{{0, -2.7}, {2.2, -3.2}, {4.4, -2.7}, {6.6, -3.3}, {8.8, -2.6}, {11, -2.8}, {12.8, -1.7}, {14, -0.9}}
	pt := func(q [2]float64, side float64) Vec3 {
		t := 0.42 * (1 - q[0]/16) // thinner toward the tip
		return root.Add(span.Scale(q[0])).Add(chord.Scale(q[1])).Add(n.Scale(side * t))
	}
	nn := func(v Vec3) [3]Vec3 { return [3]Vec3{v, v, v} }
	for i := 0; i+1 < len(lead); i++ {
		for _, side := range []float64{1, -1} {
			a, b := pt(lead[i], side), pt(lead[i+1], side)
			c, d := pt(trail[i+1], side), pt(trail[i], side)
			s.smoothTri([3]Vec3{a, b, c}, nn(n.Scale(side)), col)
			s.smoothTri([3]Vec3{a, c, d}, nn(n.Scale(side)), col)
		}
		// leading and trailing edges
		for _, e := range []struct {
			p   [][2]float64
			out Vec3
		}{{lead, chord}, {trail, chord.Scale(-1)}} {
			a, b := pt(e.p[i], 1), pt(e.p[i+1], 1)
			c, d := pt(e.p[i+1], -1), pt(e.p[i], -1)
			s.smoothTri([3]Vec3{a, b, c}, nn(e.out), col)
			s.smoothTri([3]Vec3{a, c, d}, nn(e.out), col)
		}
	}
	s.detail(func() {
		for i := 2; i+1 < len(trail); i++ {
			q := [2]float64{(trail[i][0] + trail[i+1][0]) / 2, (trail[i][1]+trail[i+1][1])/2 + 0.9}
			s.addSpot(pt(q, 1), n, 0.32, 0.75, bar)
		}
	})
}
