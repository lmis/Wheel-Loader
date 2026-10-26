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
	t       float64 // idle animation clock
	phase   float64 // gait cycle
	speed   float64
	headYaw float64 // relative to the body
	blink   float64
	rng     *rand.Rand
}

func newAnimal(kind int, owner *Player, slot int, seed int64) *Animal {
	return &Animal{Kind: kind, owner: owner, slot: slot, rng: rand.New(rand.NewSource(seed))}
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
		g.shadow(s, a.Pos, 18*r, 10*r, a.Yaw)
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
	at := func(x, z, y float64) Vec3 { return ground(a.Pos.Add(f.Scale(x)).Add(r.Scale(z)), a.Y+y) }
	return at, rotY(Vec3{1, 0, 0}, a.Yaw), rotY(Vec3{0, 0, 1}, a.Yaw)
}

// renderDog draws a retriever-shaped dog. Bernese: black coat with a white
// chest and muzzle and rust markings (accent) on the legs and cheeks.
func (a *Animal) renderDog(s *Scene, coat, light, accent color.RGBA, bernese bool, k float64) {
	at, fw, lat := a.frame()
	up := Vec3{0, 1, 0}
	moving := clamp(a.speed/30, 0, 1.5)
	// legs with a trot cycle
	for i, p := range [][2]float64{{4.5, -1.8}, {4.5, 1.8}, {-4.5, -1.8}, {-4.5, 1.8}} {
		sw := math.Sin(a.phase*2+float64(i%2)*math.Pi+float64(i/2)*math.Pi) * 0.45 * moving
		hip := at(p[0]*k, p[1]*k, 6.2*k)
		foot := hip.Add(fw.Scale(math.Sin(sw) * 5.8 * k)).Add(up.Scale(-math.Cos(sw) * 5.8 * k))
		d := foot.Sub(hip)
		legCol := coat
		if bernese {
			legCol = accent
		}
		s.addOBB2(hip.Add(d.Scale(0.5)), d, fw, d.Len()/2+0.6, 0.9*k, 0.9*k, legCol)
		paw := light
		if !bernese {
			paw = coat
		}
		s.addOBB2(foot, fw, up, 1.1*k, 0.5*k, 0.9*k, paw)
	}
	// body, chest
	s.addPrism(at(0, 0, 0), a.Yaw, []ring{
		{-6 * k, 5.5 * k, 2.3 * k, 5.2 * k, 1.4 * k},
		{-6.4 * k, 6.2 * k, 2.7 * k, 7.2 * k, 1.6 * k},
		{-5.8 * k, 5.6 * k, 2.3 * k, 9.2 * k, 1.6 * k},
	}, coat)
	chest := coat
	if bernese {
		chest = light
	}
	s.addOBB2(at(5.8*k, 0, 7.2*k), fw, up, 1.2*k, 2*k, 1.9*k, chest)
	// tail wags
	wag := math.Sin(a.t*11) * 0.7
	tb := at(-6*k, 0, 8.6*k)
	td := fw.Scale(-math.Cos(wag) * 0.7).Add(lat.Scale(math.Sin(wag) * 0.7)).Add(up.Scale(0.55)).Norm()
	s.addOBB2(tb.Add(td.Scale(3.2*k)), td, up, 3.4*k, 0.9*k, 1.1*k, coat)
	if bernese {
		s.addOBB2(tb.Add(td.Scale(6.4*k)), td, up, 0.9*k, 0.9*k, 1.1*k, light)
	}
	// head on a neck, turning toward the nearest kid
	hy := a.Yaw + a.headYaw
	hf := rotY(Vec3{1, 0, 0}, hy)
	hl := rotY(Vec3{0, 0, 1}, hy)
	neck := at(6.4*k, 0, 9.6*k)
	hc := neck.Add(hf.Scale(2 * k)).Add(up.Scale(2.4 * k))
	s.addOBB2(neck.Add(hc.Sub(neck).Scale(0.5)), hc.Sub(neck), hf, hc.Sub(neck).Len()/2+0.5, 1.5*k, 1.6*k, coat)
	s.addOBB2(hc, hf, up, 2.2*k, 2*k, 2*k, coat)
	muzzle := light
	if !bernese {
		muzzle = shadeC(coat, 1.08)
	}
	s.addOBB2(hc.Add(hf.Scale(3*k)).Add(up.Scale(-0.6*k)), hf, up, 1.5*k, 1.1*k, 1.2*k, muzzle)
	s.detail(func() {
		s.addOBB2(hc.Add(hf.Scale(4.6*k)).Add(up.Scale(-0.1*k)), hf, up, 0.4*k, 0.45*k, 0.6*k, rgba(20, 16, 14, 255)) // nose
		for _, z := range []float64{-1, 1} {
			eye := hc.Add(hf.Scale(2.15 * k)).Add(hl.Scale(z * 0.95 * k)).Add(up.Scale(0.6 * k))
			s.addOBB2(eye, hf, up, 0.1*k, 0.35*k, 0.35*k, rgba(40, 26, 16, 255))
			if bernese {
				s.addOBB2(eye.Add(up.Scale(0.7*k)), hf, up, 0.12*k, 0.2*k, 0.4*k, accent) // rust brows
				s.addOBB2(hc.Add(hf.Scale(2.2*k)).Add(hl.Scale(z*1.6*k)).Add(up.Scale(-0.9*k)), hf, up, 0.4*k, 0.5*k, 0.3*k, accent)
			}
			// floppy ears
			ear := hc.Add(hl.Scale(z * 2.2 * k)).Add(up.Scale(0.2 * k))
			s.addOBB2(ear, hf, up, 0.9*k, 1.6*k, 0.4*k, shadeC(coat, 0.85))
		}
		if bernese {
			s.addOBB2(hc.Add(hf.Scale(2.05*k)).Add(up.Scale(1.2*k)), hf, up, 0.1*k, 1.2*k, 0.35*k, light) // blaze
		}
	})
}

func (a *Animal) renderChimp(s *Scene) {
	at, fw, _ := a.frame()
	up := Vec3{0, 1, 0}
	fur := rgba(46, 36, 30, 255)
	face := rgba(196, 156, 124, 255)
	moving := clamp(a.speed/20, 0, 1.5)
	hop := 0.0
	if a.speed < 5 && !a.hopping {
		hop = math.Max(0, math.Sin(a.t*2.1)-0.92) * 18 // the odd excited bounce
	}
	at2 := func(x, z, y float64) Vec3 { return at(x, z, y+hop) }
	// knuckle-walking stance: short legs, long arms
	for i, z := range []float64{-1.6, 1.6} {
		sw := math.Sin(a.phase*2+float64(i)*math.Pi) * 0.4 * moving
		hip := at2(-1.5, z, 4.5)
		foot := hip.Add(fw.Scale(math.Sin(sw) * 4.2)).Add(up.Scale(-math.Cos(sw) * 4.2))
		d := foot.Sub(hip)
		s.addOBB2(hip.Add(d.Scale(0.5)), d, fw, d.Len()/2+0.4, 1, 1, fur)
		s.addOBB2(foot.Add(fw.Scale(0.6)), fw, up, 1.2, 0.4, 0.8, face)
	}
	for i, z := range []float64{-2.6, 2.6} {
		sw := -math.Sin(a.phase*2+float64(i)*math.Pi) * 0.4 * moving
		sh := at2(1.8, z, 8.6)
		var hand Vec3
		if i == 1 && a.speed < 5 && math.Sin(a.t*0.9) > 0.6 {
			// scratching its head
			hand = at2(1.6, 1.8, 12.6)
		} else {
			hand = sh.Add(fw.Scale(1.5 + math.Sin(sw)*3)).Add(up.Scale(-8.2))
		}
		d := hand.Sub(sh)
		s.addOBB2(sh.Add(d.Scale(0.5)), d, fw, d.Len()/2+0.4, 0.9, 0.9, fur)
		s.addOBB2(hand, fw, up, 0.8, 0.6, 0.8, face)
	}
	// body leaning forward, head
	s.addPrism(at2(0, 0, 0), a.Yaw, []ring{{-2.8, 1.6, 2.3, 4.2, 1}, {-2.2, 2.6, 2.8, 7.4, 1.2}, {-1.2, 2.6, 2.3, 9.6, 1.2}}, fur)
	hy := a.Yaw + a.headYaw
	hf := rotY(Vec3{1, 0, 0}, hy)
	hl := rotY(Vec3{0, 0, 1}, hy)
	hc := at2(2.4, 0, 11.6)
	s.addOBB2(hc, hf, up, 2, 2, 2, fur)
	s.detail(func() {
		s.addOBB2(hc.Add(hf.Scale(1.7)).Add(up.Scale(-0.4)), hf, up, 0.6, 1.5, 1.5, face)
		s.addOBB2(hc.Add(hf.Scale(2.2)).Add(up.Scale(-1.1)), hf, up, 0.7, 0.7, 1.1, shadeC(face, 0.92)) // muzzle
		s.addOBB2(hc.Add(hf.Scale(1.8)).Add(up.Scale(0.75)), hf, up, 0.35, 0.3, 1.6, shadeC(fur, 1.2))  // brow ridge
		for _, z := range []float64{-1, 1} {
			s.addOBB2(hc.Add(hf.Scale(2.3)).Add(hl.Scale(z*0.6)).Add(up.Scale(0.3)), hf, up, 0.1, 0.3, 0.3, rgba(36, 22, 12, 255))
			s.addOBB2(hc.Add(hl.Scale(z*2.2)).Add(up.Scale(0.2)), hf, up, 0.9, 1, 0.4, face) // big ears
		}
	})
}

func (a *Animal) renderOwl(s *Scene) {
	at0, fw, lat := a.frame()
	at := func(x, z, y float64) Vec3 { return at0(x, z, y-12) } // feet at a.Y
	up := Vec3{0, 1, 0}
	white := rgba(246, 246, 242, 255)
	speck := rgba(52, 50, 46, 255)
	s.addOBB2(at(0.6, 0, 12.3), fw, up, 1.4, 0.4, 1.8, rgba(80, 80, 82, 255)) // talons
	// plump, rounded body
	s.addCone(at(0, 0, 13.2), 2.2, 3.2, 2.4, 14, white)
	s.addCone(at(0, 0, 15.6), 3.2, 3.3, 2.4, 14, white)
	s.addCone(at(0, 0, 17.8), 3.3, 2.7, 2, 14, white)
	// wings: folded and speckled, or spread and flapping in flight
	for _, z := range []float64{-1, 1} {
		if a.flying || a.hopping {
			flap := math.Sin(a.t*9) * 0.6
			span := lat.Scale(z * math.Cos(flap)).Add(up.Scale(math.Sin(flap)))
			root := at(-0.4, z*2.8, 16.2)
			s.addOBB2(root.Add(span.Scale(5.5)), span, fw, 5.5, 3, 0.4, shadeC(white, 0.95))
			continue
		}
		wc := at(-0.6, z*3.1, 15.8)
		s.addOBB2(wc, fw.Scale(0.9).Add(up.Scale(-0.4)), lat.Scale(z), 3, 0.5, 3, shadeC(white, 0.95))
	}
	s.detail(func() {
		for i := 0; i < 16; i++ {
			ang := 1.2 + float64(i)*0.55
			r := 3.0 + 0.2*float64(i%2)
			p := at(math.Cos(ang)*r, math.Sin(ang)*r, 13.4+float64(i%4)*1.3)
			s.addOBB2(p, fw, up, 0.22, 0.2, 0.45, speck)
		}
	})
	// round head that swivels a long way, with a facial disc and big yellow eyes
	hy := a.Yaw + a.headYaw
	hf := rotY(Vec3{1, 0, 0}, hy)
	hl := rotY(Vec3{0, 0, 1}, hy)
	hc := at(0, 0, 21.2)
	s.addCone(hc.Sub(Vec3{0, 1.6, 0}), 2.6, 3.1, 1.6, 14, white)
	s.addCone(hc.Add(Vec3{0, 0.6, 0}), 3.1, 1.6, 2.8, 14, white)
	s.detail(func() {
		disc := hc.Add(hf.Scale(2.75)).Add(up.Scale(-0.3))
		s.addOBB2(disc, hf, up, 0.15, 1.9, 2.4, rgba(252, 252, 250, 255))
		for _, z := range []float64{-1.1, 1.1} {
			e := disc.Add(hf.Scale(0.2)).Add(hl.Scale(z)).Add(up.Scale(0.45))
			if a.blink > 0 {
				s.addOBB2(e, hf, up, 0.1, 0.12, 0.75, rgba(210, 210, 206, 255))
				continue
			}
			s.addOBB2(e, hf, up, 0.1, 0.72, 0.72, rgba(250, 200, 24, 255))
			s.addOBB2(e.Add(hf.Scale(0.1)), hf, up, 0.08, 0.36, 0.36, rgba(14, 12, 10, 255))
		}
		s.addOBB2(disc.Add(hf.Scale(0.35)).Add(up.Scale(-0.6)), hf, up, 0.35, 0.55, 0.32, rgba(40, 38, 36, 255)) // beak
		for i := 0; i < 5; i++ {
			ang := 2.2 + float64(i)*0.45
			p := hc.Add(hf.Scale(math.Cos(ang) * 2.9)).Add(hl.Scale(math.Sin(ang) * 2.9)).Add(up.Scale(0.8))
			s.addOBB2(p, hf, up, 0.2, 0.18, 0.4, speck)
		}
	})
}
