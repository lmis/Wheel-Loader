package main

import (
	"image/color"
	"math"
)

// A little wonder around the site, just to look at: a train passing in the
// distance, a helicopter flying over, and a rollerblader with glowing red
// skates zooming round the roads.

// ---- the train ----

const (
	trackY      = 198.0 // between the north road and the houses
	trackStart  = -1500.0
	trackEnd    = 5700.0
	trainSpeed  = 130.0
	trainPause  = 18.0 // seconds between trains
	locoLen     = 84.0
	carLen      = 72.0
	carGap      = 7.0
	trainCars   = 5
	trainLength = locoLen + trainCars*(carLen+carGap)
)

// Train runs east along the track, waits off in the fog, and comes again.
// It stops for anything standing on the line.
type Train struct {
	X     float64 // front of the locomotive
	speed float64
	wait  float64
}

func (t *Train) Update(dt float64, w *World) {
	if t.wait > 0 {
		t.wait -= dt
		return
	}
	front := V(t.X+26, trackY)
	blocked := false
	if front.X > 0 && front.X < w.Width {
		blocked, _ = w.CollideCircle(front, 24, nil)
	}
	if blocked {
		t.speed = approach(t.speed, 0, 220*dt)
	} else {
		t.speed = approach(t.speed, trainSpeed, 40*dt)
	}
	t.X += t.speed * dt
	if t.X-trainLength > trackEnd {
		t.X, t.wait = trackStart, trainPause
	}
}

func (g *Game) renderTrack(s *Scene) {
	ballast := rgba(104, 96, 88, 210)
	rail := rgba(176, 178, 184, 255)
	for x := trackStart; x < trackEnd; x += 200 {
		x1 := x + 200
		if !g.cam.inView(ground(V(x+100, trackY), 0), 110) {
			continue
		}
		strip := func(z0, z1 float64, c color.RGBA) {
			s.addDecalQuad(ground(V(x, trackY+z0), 0.5), ground(V(x1, trackY+z0), 0.5),
				ground(V(x1, trackY+z1), 0.5), ground(V(x, trackY+z1), 0.5), c)
		}
		strip(-17, 17, ballast)
		strip(-8, -6.5, rail)
		strip(6.5, 8, rail)
	}
}

func (g *Game) renderTrain(s *Scene, t *Train) {
	if t.wait > 0 {
		return
	}
	body := func(x0, x1 float64, rings []ring, c color.RGBA) {
		o := ground(V(0, trackY), 0)
		for i := range rings {
			rings[i].back += x0
			rings[i].front += x1
		}
		s.addPrism(o, 0, rings, c)
	}
	east, across := Vec3{1, 0, 0}, Vec3{0, 0, 1}
	bogie := func(x float64) {
		s.addBox(ground(V(x, trackY), 7), 9, 2.5, 7, 0, colTire)
		for _, dx := range []float64{-5, 5} {
			for _, z := range []float64{-8.5, 8.5} {
				s.addRoundWheel(ground(V(x+dx, trackY+z), 4.5), east, across, 4.2, 1.6, 14, colSteel, colDark)
			}
		}
	}
	// locomotive: red, with a yellow stripe, a sloped nose and a cab
	x1 := t.X
	x0 := x1 - locoLen
	if g.cam.inView(ground(V(x1-locoLen/2, trackY), 25), 60) {
		red := rgba(206, 52, 44, 255)
		body(x0, x1-16, []ring{{0, 0, 13, 10, 2}, {0, 0, 13, 34, 3}, {4, -2, 11, 40, 4}}, red)
		body(x1-16, x1, []ring{{0, -2, 13, 10, 2}, {0, -9, 12, 30, 3}}, red) // nose
		s.detail(func() {
			s.addBox(ground(V(x1-38, trackY), 20), 22, 2, 13.1, 0, rgba(250, 206, 40, 255))
			for _, z := range []float64{-13.1, 13.1} {
				s.addBox(ground(V(x1-24, trackY+z), 30), 6, 4, 0.1, 0, colGlass)
			}
			s.addBox(ground(V(x1-10, trackY), 26), 2, 4, 9, 0, colGlass)                    // windscreen
			s.addBox(ground(V(x1-0.5, trackY), 14), 0.5, 2, 3, 0, rgba(255, 245, 200, 255)) // headlight
		})
		bogie(x0 + 16)
		bogie(x1 - 18)
	}
	// carriages
	colors := []color.RGBA{rgba(46, 140, 96, 255), rgba(236, 226, 196, 255), rgba(54, 104, 180, 255)}
	for i := 0; i < trainCars; i++ {
		c1 := x0 - carGap - float64(i)*(carLen+carGap)
		c0 := c1 - carLen
		if !g.cam.inView(ground(V(c1-carLen/2, trackY), 22), 50) {
			continue
		}
		col := colors[i%len(colors)]
		body(c0, c1, []ring{{0, 0, 12, 10, 2}, {0, 0, 12, 34, 2}, {2, -2, 10, 38, 4}}, col)
		s.detail(func() {
			for _, z := range []float64{-12.1, 12.1} {
				s.addBox(ground(V((c0+c1)/2, trackY+z), 27), carLen/2-6, 4, 0.1, 0, colGlass) // window band
			}
		})
		s.addBox(ground(V(c1+carGap/2, trackY), 16), carGap/2+1, 2, 3, 0, colDark) // coupling
		bogie(c0 + 14)
		bogie(c1 - 14)
	}
}

// ---- the helicopter ----

const (
	heliPeriod = 46.0 // seconds between passes
	heliSpeed  = 210.0
	heliAlt    = 140.0 // well under the crane jibs, low enough for the kids to see
)

var (
	// over the middle of the site, between the two crane masts
	heliFrom = V(-500, 2825)
	heliTo   = V(3700, -325)
)

// Helicopter crosses the site diagonally every so often.
type Helicopter struct {
	t float64 // time into the current pass
}

func (h *Helicopter) Update(dt float64) {
	h.t += dt
	if h.t > heliPeriod {
		h.t = 0
	}
}

// Pos is where the helicopter is, and whether it is in the air at all.
func (h *Helicopter) Pos() (Vec, float64, bool) {
	d := heliTo.Sub(heliFrom)
	along := h.t * heliSpeed
	if along > d.Len() {
		return Vec{}, 0, false
	}
	return heliFrom.Add(d.Norm().Scale(along)), d.Angle(), true
}

func (g *Game) renderHelicopter(s *Scene, h *Helicopter) {
	pos, yaw, ok := h.Pos()
	if !ok {
		return
	}
	alt := heliAlt + math.Sin(h.t*1.3)*6
	g.shadow(s, pos, 120, 50, yaw)
	if !g.cam.inView(ground(pos, alt), 110) {
		return
	}
	o := ground(pos, alt)
	at := func(x, z, y float64) Vec3 { return o.Add(rotY(Vec3{x, y, z}, yaw)) }
	fw := rotY(Vec3{1, 0, 0}, yaw)
	up := Vec3{0, 1, 0}
	red := rgba(214, 46, 40, 255)
	white := rgba(240, 240, 236, 255)
	// cabin and glass nose
	s.addPrism(o, yaw, []ring{{-22, 18, 9, -2, 4}, {-26, 24, 14, 8, 6}, {-24, 22, 14, 22, 6}, {-16, 10, 9, 30, 4}}, red)
	s.addPrism(o, yaw, []ring{{18, 26, 9, 0, 4}, {24, 34, 12, 8, 5}, {22, 30, 11, 20, 5}, {10, 18, 7, 27, 3}}, colGlass)
	s.detail(func() {
		for _, z := range []float64{-14.1, 14.1} {
			s.addOBB2(at(-2, z, 12), fw, up, 14, 2, 0.1, white) // stripe
		}
	})
	// tail boom, fin and tail rotor
	s.addOBB2(at(-56, 0, 18), fw, up, 32, 3.5, 3, red)
	s.addOBB2(at(-88, 0, 28), fw.Add(up.Scale(0.6)), up, 7, 8, 1.2, red)
	tr := h.t * 40
	for k := 0; k < 2; k++ {
		a := tr + float64(k)*math.Pi/2
		blade := fw.Scale(math.Cos(a)).Add(up.Scale(math.Sin(a)))
		s.addOBB2(at(-88, 2.6, 30), blade, rotY(Vec3{0, 0, 1}, yaw), 9, 0.3, 1.4, colDark)
	}
	// skids
	for _, z := range []float64{-13, 13} {
		s.addOBB2(at(0, z, -11), fw, up, 26, 1, 1, colDark)
		for _, x := range []float64{-12, 12} {
			s.addOBB2(at(x, z*0.8, -6), up, fw, 5, 0.8, 0.8, colDark)
		}
	}
	// mast and main rotor: four long blades, turning fast
	s.addOBB2(at(0, 0, 34), up, fw, 4, 2, 2, colDark)
	rot := h.t * 22
	for k := 0; k < 4; k++ {
		bd := rotY(Vec3{1, 0, 0}, yaw+rot+float64(k)*math.Pi/2)
		s.addOBB2(o.Add(Vec3{0, 38, 0}).Add(bd.Scale(36)), bd, up, 36, 0.4, 2.6, rgba(40, 40, 44, 255))
	}
	// blinking light on the tail
	if math.Mod(h.t, 1) < 0.15 {
		s.glowing(func() { s.addBox(at(-90, 0, 37), 1.2, 1.2, 1.2, yaw, rgba(255, 60, 60, 255)) })
	}
}

// ---- the rollerblader ----

// Skater laps the site roads at speed: down the middle road between the
// kids, along the south road, up the west road and back along the north.
type Skater struct {
	s     float64 // distance along the loop
	off   float64 // sideways offset from the road centre (to the right)
	speed float64
	phase float64 // stride cycle
	Pos   Vec
	Yaw   float64
	lean  float64
}

var skateLoop = []Vec{V(1600, 300), V(1600, 2100), V(300, 2100), V(300, 300)}

const skaterSpeed = 280.0

// loopAt returns the point and heading at distance d along the loop.
func loopAt(d float64) (Vec, float64) {
	total := 0.0
	for i := range skateLoop {
		a, b := skateLoop[i], skateLoop[(i+1)%len(skateLoop)]
		total += b.Sub(a).Len()
	}
	d = math.Mod(d, total)
	if d < 0 {
		d += total
	}
	for i := range skateLoop {
		a, b := skateLoop[i], skateLoop[(i+1)%len(skateLoop)]
		l := b.Sub(a).Len()
		if d <= l {
			return a.Add(b.Sub(a).Scale(d / l)), b.Sub(a).Angle()
		}
		d -= l
	}
	return skateLoop[0], 0
}

func (k *Skater) Update(dt float64, w *World) {
	k.speed = approach(k.speed, skaterSpeed, 120*dt)
	k.s += k.speed * dt
	// steer for a point a little ahead (rounds the corners), swerving round
	// anything in the way
	ahead, head := loopAt(k.s + 60)
	right := dir(head + math.Pi/2)
	want := 22.0
	for _, off := range []float64{22, 52, -10, -40, 80} {
		if hit, _ := w.CollideCircle(ahead.Add(right.Scale(off)), 14, nil); !hit {
			want = off
			break
		}
	}
	k.off = approach(k.off, want, 90*dt)
	center, h := loopAt(k.s)
	target := center.Add(dir(h + math.Pi/2).Scale(k.off))
	prev := k.Yaw
	if d := ahead.Add(right.Scale(k.off)).Sub(k.Pos); d.Len() > 1 {
		k.Yaw += angleDiff(d.Angle(), k.Yaw) * (1 - math.Pow(0.002, dt))
	}
	k.Pos = target
	turn := angleDiff(k.Yaw, prev) / dt
	k.lean = approach(k.lean, clamp(turn*0.35, -0.5, 0.5), 2*dt)
	k.phase += k.speed * dt / 55
	// red sparks of light trail off the wheels
	if w.rng.Float64() < 0.5 {
		foot := k.Pos.Add(dir(k.Yaw + math.Pi/2).Scale(math.Sin(k.phase) * 4))
		w.addParticle(Particle{
			Pos: foot, Vel: dir(k.Yaw).Scale(-30), Z: 1, VZ: 4,
			Life: 0.35, Size: 2.2, Kind: pfxSpark, Color: [3]uint8{255, 50, 40},
		})
	}
}

func (g *Game) renderSkater(s *Scene, k *Skater) {
	g.shadow(s, k.Pos, 16, 10, k.Yaw)
	if !g.cam.inView(ground(k.Pos, 18), 30) {
		return
	}
	f := dir(k.Yaw)
	r := dir(k.Yaw + math.Pi/2)
	fw := rotY(Vec3{1, 0, 0}, k.Yaw)
	lat := rotY(Vec3{0, 0, 1}, k.Yaw)
	up := Vec3{0, 1, 0}
	// lean forward like a speed skater, and into the turns
	tilt := up.Add(fw.Scale(0.35)).Add(lat.Scale(k.lean)).Norm()
	base := ground(k.Pos, 0)
	at := func(x, z, y float64) Vec3 {
		return base.Add(fw.Scale(x)).Add(lat.Scale(z)).Add(tilt.Scale(y))
	}
	var (
		skin    = rgba(236, 190, 156, 255)
		shirt   = rgba(28, 168, 172, 255)
		tights  = rgba(36, 36, 44, 255)
		boot    = rgba(196, 26, 30, 255)
		glowRed = rgba(255, 70, 56, 255)
		hair    = rgba(112, 66, 34, 255)
		helmet  = rgba(34, 36, 44, 255)
		lens    = rgba(40, 140, 255, 255)
	)
	// legs push out to the sides in turn, skates gliding along
	stride := math.Sin(k.phase)
	for _, side := range []float64{-1, 1} {
		push := math.Max(0, stride*side) * 7 // this leg is pushing out
		hip := at(0, side*2.2, 15)
		foot := ground(k.Pos.Add(r.Scale(side*(2.5+push))).Add(f.Scale(-push*0.4)), 2.2)
		knee := hip.Add(foot.Sub(hip).Scale(0.5)).Add(fw.Scale(2.4))
		s.addCapsule(hip, knee, 1.75, 1.45, tights)
		s.addCapsule(knee, foot.Add(up.Scale(2.2)), 1.45, 1.2, tights)
		// glowing red rollerblades: a rounded boot and a row of lit wheels
		s.addEllipsoid(foot.Add(up.Scale(1.3)).Add(fw.Scale(0.4)), fw, up, 3.3, 1.5, 1.5, boot)
		s.addCapsule(foot.Add(up.Scale(1.8)), foot.Add(up.Scale(3.4)), 1.45, 1.4, boot) // cuff
		s.addOBB2(foot.Add(up.Scale(-0.4)), fw, up, 3.3, 0.3, 0.5, colDark)             // frame
		s.glowing(func() {
			for w := -1.5; w <= 1.51; w += 1 {
				c := foot.Add(fw.Scale(w * 1.7)).Add(up.Scale(-1.15))
				s.addRoundWheel(c, fw, lat, 0.85, 0.7, 14, glowRed, rgba(255, 150, 120, 255))
			}
		})
		gp := k.Pos.Add(r.Scale(side * (2.5 + push)))
		s.addDecalQuad(ground(gp.Add(f.Scale(-8)).Add(r.Scale(-5)), 0.6), ground(gp.Add(f.Scale(8)).Add(r.Scale(-5)), 0.6),
			ground(gp.Add(f.Scale(8)).Add(r.Scale(5)), 0.6), ground(gp.Add(f.Scale(-8)).Add(r.Scale(5)), 0.6), rgba(255, 40, 30, 70))
	}
	// body: hips, chest and shoulders, rounded
	s.addEllipsoid(at(0, 0, 15.5), fw, tilt, 2.4, 1.8, 3.0, tights)
	s.addRoundPrism(at(0, 0, 0), k.Yaw, []ring{{-2.3, 2.3, 3, 15.5, 1.8}, {-2.6, 2.6, 3.6, 21, 2.2}, {-2.2, 2.2, 3.4, 25, 2}}, shirt, false, true)
	// arms swing wide, against the legs
	for _, side := range []float64{-1, 1} {
		sw := -stride * side
		sh := at(0, side*3.9, 24)
		hand := sh.Add(lat.Scale(side * (3 + math.Max(0, sw)*3))).Add(fw.Scale(sw * 6)).Add(tilt.Scale(-8))
		elbow := sh.Add(hand.Sub(sh).Scale(0.5)).Add(fw.Scale(-1.2))
		s.addSphere(sh, 1.4, shirt)
		s.addCapsule(sh, elbow, 1.25, 1.1, shirt)
		s.addCapsule(elbow, hand, 1.1, 0.95, skin)
		s.addSphere(hand, 1.05, skin)
	}
	// head with a full beard, blue glasses, long hair and a dark helmet
	s.addCapsule(at(0, 0, 25), at(0.2, 0, 27), 1.2, 1.15, skin)
	head := at(0.4, 0, 29)
	const hx, hy, hz = 2.5, 2.8, 2.4
	onFace := func(dz, dy, out float64) Vec3 {
		x := hx * math.Sqrt(math.Max(0, 1-(dz/hz)*(dz/hz)-(dy/hy)*(dy/hy)))
		return head.Add(fw.Scale(x + out)).Add(lat.Scale(dz)).Add(tilt.Scale(dy))
	}
	s.addEllipsoid(head, fw, tilt, hx, hy, hz, skin)
	// the beard: round the jaw and chin, and a moustache
	s.addEllipsoid(head.Add(fw.Scale(1.2)).Add(tilt.Scale(-1.7)), fw, tilt, 1.6, 1.5, 2.35, hair)
	s.addEllipsoid(head.Add(fw.Scale(2.0)).Add(tilt.Scale(-2.7)), fw.Add(tilt.Scale(-0.4)), tilt, 1.1, 1.2, 1.5, hair)
	s.detail(func() {
		s.addEllipsoid(onFace(0, -0.95, 0.05), fw, tilt, 0.4, 0.38, 1.25, hair)                                           // moustache
		s.addEllipsoid(head.Add(fw.Scale(2.82)).Add(tilt.Scale(-1.75)), fw, tilt, 0.15, 0.2, 0.7, rgba(150, 60, 56, 255)) // smile
		s.addEllipsoid(onFace(0, -0.3, 0.05), fw, tilt, 0.4, 0.45, 0.38, shadeC(skin, 0.95))                              // nose
		// round blue glasses on a dark frame
		for _, z := range []float64{-1, 1} {
			e := onFace(z, 0.45, 0.05)
			s.addEllipsoid(e, fw, tilt, 0.18, 0.72, 0.82, colDark)
			s.glowing(func() { s.addEllipsoid(e.Add(fw.Scale(0.08)), fw, tilt, 0.14, 0.6, 0.7, lens) })
			s.addCapsule(e.Add(lat.Scale(z*0.8)), head.Add(lat.Scale(z*2.35)).Add(tilt.Scale(0.5)), 0.12, 0.12, colDark)
		}
		s.addCapsule(onFace(-0.2, 0.5, 0.1), onFace(0.2, 0.5, 0.1), 0.12, 0.12, colDark) // bridge
	})
	// long hair streaming out behind, and round the sides under the helmet
	strand := head.Add(fw.Scale(-2)).Add(tilt.Scale(0.4))
	for i := 0; i < 5; i++ {
		wave := math.Sin(g.frameTime()*14 - float64(i)*0.9)
		next := strand.Add(fw.Scale(-3.4)).Add(tilt.Scale(-0.6 - 0.25*float64(i))).Add(lat.Scale(wave * 0.8))
		s.addCapsule(strand, next, 1.9-0.22*float64(i), 1.7-0.22*float64(i), hair)
		strand = next
	}
	s.addEllipsoid(head.Add(fw.Scale(-0.7)).Add(tilt.Scale(-0.6)), fw, tilt, 2.2, 2.2, 2.6, hair)
	// the helmet: a smooth dark dome, with a red stripe to match the skates
	s.addEllipsoid(head.Add(tilt.Scale(1.9)), fw, tilt, 2.95, 2.3, 2.8, helmet)
	s.detail(func() {
		s.addEllipsoid(head.Add(tilt.Scale(4.15)), fw, tilt, 2.2, 0.2, 0.45, glowRed)
	})
}

// ---- wiring ----

// wonders holds the things to look at.
type wonders struct {
	train  Train
	heli   Helicopter
	skater Skater
}

func newWonders() *wonders {
	wd := &wonders{train: Train{X: 1100, speed: trainSpeed}, heli: Helicopter{t: heliPeriod - 6}}
	wd.skater.s = 300
	wd.skater.Pos, wd.skater.Yaw = loopAt(wd.skater.s)
	wd.skater.speed = skaterSpeed
	return wd
}

func (g *Game) updateWonders(dt float64) {
	wd := g.wonder
	wd.train.Update(dt, g.world)
	wd.heli.Update(dt)
	wd.skater.Update(dt, g.world)
}

func (g *Game) renderWonders(s *Scene) {
	wd := g.wonder
	g.renderTrack(s)
	g.renderTrain(s, &wd.train)
	g.renderHelicopter(s, &wd.heli)
	g.renderSkater(s, &wd.skater)
}
