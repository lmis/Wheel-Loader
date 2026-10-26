package main

import (
	"image/color"
	"math"
	"math/rand"
)

// Animal kinds hanging around the starting area.
const (
	animalGolden  = iota // golden retriever
	animalBernese        // Bernese mountain dog
	animalChimp          // small chimpanzee
	animalOwl            // snowy owl on a fence post
)

// Animal is a simple wandering critter. It idles, ambles to a nearby spot,
// turns its head toward the closest kid, and trots clear of moving vehicles.
type Animal struct {
	Kind    int
	Pos     Vec
	Yaw     float64
	home    Vec
	target  Vec
	walking bool
	timer   float64
	t       float64 // idle animation clock
	phase   float64 // gait cycle
	speed   float64
	headYaw float64 // relative to the body
	blink   float64
	rng     *rand.Rand
}

func newAnimal(kind int, at Vec, yaw float64, seed int64) *Animal {
	return &Animal{Kind: kind, Pos: at, home: at, Yaw: yaw, rng: rand.New(rand.NewSource(seed)), timer: 1 + float64(seed%3)}
}

func (a *Animal) radius() float64 {
	switch a.Kind {
	case animalOwl:
		return 4
	case animalChimp:
		return 6
	}
	return 8
}

func (a *Animal) collider() obstacle { return obstacle{a.Pos, a.radius(), a} }

// Update moves the animal. lookAt is the nearest kid (if any) and threats are
// vehicles to keep away from.
func (a *Animal) Update(dt float64, w *World, lookAt *Vec, threats []Vehicle) {
	a.t += dt
	a.blink -= dt
	if a.blink < -4-a.rng.Float64()*4 {
		a.blink = 0.1
	}
	// head turns toward the nearest kid within range
	want := 0.0
	if lookAt != nil && lookAt.Sub(a.Pos).Len() < 260 {
		want = angleDiff(lookAt.Sub(a.Pos).Angle(), a.Yaw)
		lim := 1.1
		if a.Kind == animalOwl {
			lim = 2.6 // owls can turn their heads a very long way
		}
		want = clamp(want, -lim, lim)
	} else {
		want = math.Sin(a.t*0.7+float64(a.Kind)) * 0.5
	}
	a.headYaw += (want - a.headYaw) * (1 - math.Pow(0.03, dt))
	if a.Kind == animalOwl {
		return // perched
	}

	maxSpeed := 34.0
	if a.Kind == animalChimp {
		maxSpeed = 26
	}
	// trot clear of anything big that is moving nearby
	for _, v := range threats {
		d := a.Pos.Sub(v.Center())
		if d.Len() < 150 && math.Abs(v.SpeedVal()) > 25 {
			a.target = a.Pos.Add(d.Norm().Scale(140))
			a.walking, a.timer = true, 3
			maxSpeed *= 2.2
			break
		}
	}
	a.timer -= dt
	if a.timer <= 0 {
		if a.walking {
			a.walking, a.timer = false, 2+a.rng.Float64()*4
		} else {
			ang := a.rng.Float64() * 2 * math.Pi
			a.target = a.home.Add(dir(ang).Scale(30 + a.rng.Float64()*90))
			a.walking, a.timer = true, 6
		}
	}
	goal := 0.0
	if a.walking {
		to := a.target.Sub(a.Pos)
		if to.Len() < 8 {
			a.walking, a.timer = false, 2+a.rng.Float64()*4
		} else {
			a.Yaw += angleDiff(to.Angle(), a.Yaw) * (1 - math.Pow(0.05, dt))
			goal = maxSpeed
		}
	}
	a.speed = approach(a.speed, goal, 80*dt)
	np := a.Pos.Add(dir(a.Yaw).Scale(a.speed * dt))
	if w != nil {
		if hit, _ := w.CollideCircle(np, a.radius(), a); hit {
			a.walking, a.timer, a.speed = false, 1, 0
			np = a.Pos
		}
	}
	a.Pos = np
	a.phase += a.speed * dt * 0.5
}

func (a *Animal) Render(g *Game, s *Scene) {
	g.shadow(s, a.Pos, 18, 10, a.Yaw)
	if !g.cam.inView(ground(a.Pos, 8), 20) {
		return
	}
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
	at := func(x, z, y float64) Vec3 { return ground(a.Pos.Add(f.Scale(x)).Add(r.Scale(z)), y) }
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
	if !a.walking {
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
		if i == 1 && !a.walking && math.Sin(a.t*0.9) > 0.6 {
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
	at, fw, lat := a.frame()
	up := Vec3{0, 1, 0}
	white := rgba(246, 246, 242, 255)
	speck := rgba(52, 50, 46, 255)
	// fence post perch
	s.addCylinder(at(0, 0, 6), 2.2, 12, 8, rgba(120, 86, 52, 255))
	s.addOBB2(at(0.6, 0, 12.3), fw, up, 1.4, 0.4, 1.8, rgba(80, 80, 82, 255)) // talons
	// plump, rounded body
	s.addCone(at(0, 0, 13.2), 2.2, 3.2, 2.4, 14, white)
	s.addCone(at(0, 0, 15.6), 3.2, 3.3, 2.4, 14, white)
	s.addCone(at(0, 0, 17.8), 3.3, 2.7, 2, 14, white)
	// folded wings, speckled
	for _, z := range []float64{-1, 1} {
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
