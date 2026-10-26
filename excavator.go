package main

import (
	"image/color"
	"math"
)

// Excavator is a tracked digger: the tracks drive and pivot, the house slews
// on top, and the boom / stick / bucket chain reaches out to dig material
// from the grid and drop it into trucks or onto the ground.
type Excavator struct {
	Pos     Vec
	Heading float64 // tracks
	Slew    float64 // house angle relative to the tracks
	Boom    float64 // 0 = lowered, 1 = raised
	Stick   float64 // 0 = reached out, 1 = tucked in
	Bucket  float64 // 0 = open (dump), 1 = curled
	Load    float64
	LoadTyp int

	Speed    float64
	Occupied bool

	throttle, steerCmd, boomCmd, stickCmd, bucketCmd, slewCmd float64
	brake                                                     bool
	slewV, rpm, jolt, trackPhase                              float64
}

const (
	excTrackLen  = 150.0
	excTrackW    = 90.0
	excHouseY    = 20.0
	excBoomLen   = 72.0
	excStickLen  = 58.0
	excBucketLen = 20.0
	excCapacity  = 90.0
)

func (e *Excavator) Name() string          { return "Excavator" }
func (e *Excavator) Center() Vec           { return e.Pos }
func (e *Excavator) Facing() float64       { return e.Heading + e.Slew }
func (e *Excavator) SpeedVal() float64     { return e.Speed }
func (e *Excavator) EnterRadius() float64  { return 105 }
func (e *Excavator) SetOccupied(b bool)    { e.Occupied = b }
func (e *Excavator) Jolt() float64         { return e.jolt }
func (e *Excavator) CamDist() float64      { return 600 }
func (e *Excavator) Colliders() []obstacle { return e.collidersAt(e.Pos, e.Heading) }

func (e *Excavator) collidersAt(p Vec, h float64) []obstacle {
	f := dir(h)
	return []obstacle{{p.Add(f.Scale(38)), 44, e}, {p.Sub(f.Scale(38)), 44, e}}
}

func (e *Excavator) Drive(c *Controls) {
	if c == nil {
		e.throttle, e.steerCmd, e.boomCmd, e.stickCmd, e.bucketCmd, e.slewCmd = 0, 0, 0, 0, 0, 0
		e.brake = true
		return
	}
	e.throttle, e.steerCmd, e.brake = c.Throttle, c.Steer, c.Brake
	e.boomCmd, e.stickCmd, e.bucketCmd, e.slewCmd = c.A1, -c.A2, c.A3, c.A4
}

// houseYaw is the world heading of the upper structure.
func (e *Excavator) houseYaw() float64 { return e.Heading + e.Slew }

// houseAt converts house-local coordinates (x forward, z right, y up).
func (e *Excavator) houseAt(x, z, y float64) Vec3 {
	h := e.houseYaw()
	return ground(e.Pos.Add(dir(h).Scale(x)).Add(dir(h+math.Pi/2).Scale(z)), y)
}

// armAngles returns the absolute pitch of the boom, stick and bucket.
func (e *Excavator) armAngles(boom, stick, bucket float64) (float64, float64, float64) {
	tb := lerp(-0.55, 0.95, boom)
	ts := tb - lerp(0.35, 2.5, stick)
	tk := ts - lerp(0.0, 2.3, bucket)
	return tb, ts, tk
}

// armPoints solves the digging chain in the house's vertical plane, returning
// the boom foot, elbow, wrist and bucket teeth as (reach, height) pairs.
func (e *Excavator) armPoints(boom, stick, bucket float64) (p0, p1, p2, p3 [2]float64) {
	tb, ts, tk := e.armAngles(boom, stick, bucket)
	p0 = [2]float64{38, 38}
	p1 = [2]float64{p0[0] + math.Cos(tb)*excBoomLen, p0[1] + math.Sin(tb)*excBoomLen}
	p2 = [2]float64{p1[0] + math.Cos(ts)*excStickLen, p1[1] + math.Sin(ts)*excStickLen}
	p3 = [2]float64{p2[0] + math.Cos(tk)*excBucketLen, p2[1] + math.Sin(tk)*excBucketLen}
	return
}

const excArmZ = 8.0 // the boom sits right of centre, beside the cab

// Tip returns the bucket teeth position on the ground plane and their height.
func (e *Excavator) Tip() (Vec, float64) {
	_, _, _, p3 := e.armPoints(e.Boom, e.Stick, e.Bucket)
	h := e.houseYaw()
	return e.Pos.Add(dir(h).Scale(p3[0])).Add(dir(h + math.Pi/2).Scale(excArmZ)), p3[1]
}

func (e *Excavator) Seat() (Vec3, float64) { return e.houseAt(16, -18, 61), e.houseYaw() }

func (e *Excavator) ExitPoint(side float64) Vec {
	return e.Pos.Add(dir(e.Heading + math.Pi/2).Scale(-side * 78))
}

func (e *Excavator) Update(dt float64, w *World) {
	// tracks: tank steering, can pivot on the spot
	target := e.throttle * 70
	if e.brake {
		target = 0
	}
	e.Speed = approach(e.Speed, target, 90*dt)
	turn := e.steerCmd * 0.8
	nh := e.Heading + turn*dt
	np := e.Pos.Add(dir(nh).Scale(e.Speed * dt))
	if w != nil {
		if hit, n := collideCircles(w, e.collidersAt(np, nh), e); hit {
			np, nh = e.Pos.Add(n.Scale(1.5)), e.Heading
			e.jolt = math.Max(e.jolt, math.Abs(e.Speed)*0.8+30)
			e.Speed = 0
		}
		np.X = clamp(np.X, 80, w.Width-80)
		np.Y = clamp(np.Y, 80, w.Height-80)
	}
	e.trackPhase += (math.Abs(e.Speed) + math.Abs(turn)*40) * dt
	e.Pos, e.Heading = np, nh
	e.jolt *= math.Pow(0.02, dt)

	// house slew has inertia, like the real thing
	e.slewV = approach(e.slewV, e.slewCmd*0.9, 1.6*dt)
	e.Slew += e.slewV * dt

	// hydraulics: the teeth may not go below the ground surface
	nb := clamp(e.Boom+e.boomCmd*0.55*dt, 0, 1)
	ns := clamp(e.Stick+e.stickCmd*0.7*dt, 0, 1)
	nk := clamp(e.Bucket+e.bucketCmd*0.9*dt, 0, 1)
	floor := -1.0
	if w != nil {
		tip, _ := e.Tip()
		if cx, cy, ok := w.Grid.cellAt(tip.X, tip.Y); ok {
			// never below the ground; may bite a little way into a heap
			floor = math.Max(-1, pileHeight(w.Grid.at(cx, cy).Amount)-14)
		}
	}
	if _, _, _, p3 := e.armPoints(nb, ns, nk); p3[1] >= floor {
		e.Boom, e.Stick, e.Bucket = nb, ns, nk
	} else {
		// keep whatever motion lifts the teeth, refuse what digs into the ground
		if e.boomCmd > 0 {
			e.Boom = nb
		}
		if e.bucketCmd > 0 {
			e.Bucket = nk
		}
		if e.stickCmd > 0 {
			e.Stick = ns
		}
	}
	e.rpm = approach(e.rpm, 0.2+0.8*math.Min(1, math.Abs(e.boomCmd)+math.Abs(e.stickCmd)+math.Abs(e.bucketCmd)+math.Abs(e.throttle)), 2*dt)
	if w == nil {
		return
	}

	tip, tipY := e.Tip()
	g := w.Grid
	surface := 0.0
	if cx, cy, ok := g.cellAt(tip.X, tip.Y); ok {
		surface = pileHeight(g.at(cx, cy).Amount)
	}
	// dig: teeth in the material while curling or pulling the stick in
	if tipY < surface+6 && e.Load < excCapacity {
		effort := math.Max(0, e.bucketCmd) + math.Max(0, e.stickCmd)*0.7
		if effort > 0 {
			got, t := g.TakeAround(tip.X, tip.Y, 24, math.Min(effort*55*dt, excCapacity-e.Load), e.LoadTyp)
			if got > 0 {
				if e.Load < 1 {
					e.LoadTyp = t
				}
				e.Load += got
				w.spawnScoopDust(tip, e.houseYaw())
			}
		}
	}
	// dump: an opened bucket held up off the ground lets the load go
	if e.Load > 0 && e.Bucket < 0.3 && tipY > surface+8 {
		amt := math.Min((0.3-e.Bucket)*500*dt, e.Load)
		e.Load -= amt
		if tr := w.TruckAt(tip); tr != nil {
			tr.Receive(amt, e.LoadTyp)
		} else {
			g.Add(tip.X, tip.Y, amt, e.LoadTyp)
			g.Relax(tip.X, tip.Y, 60, 1)
		}
		w.spawnMaterialFall(tip, e.houseYaw(), amt, e.LoadTyp)
	}
}

func (e *Excavator) Effects(w *World) {
	if e.Occupied || math.Abs(e.Speed) > 3 {
		stack := e.houseAt(-26, 14, 0)
		w.spawnExhaustAt(Vec{stack.X, stack.Z}, 72, e.rpm)
	}
	if math.Abs(e.Speed) > 20 {
		w.spawnDriveDust(e.Pos.Sub(dir(e.Heading).Scale(70)), e.Heading, 0.5)
	}
}

func (e *Excavator) ControlHelp() []string {
	return []string{
		"{T}\ttracks forward / back",
		"{S}\tturn tracks",
		"{B}\tbrake",
		"{1}\tboom up / down",
		"{2}\tstick out / in",
		"{3}\tcurl / open bucket",
		"{4}\tswing house left / right",
	}
}

func (e *Excavator) Gauges() []gauge {
	load := "empty"
	if e.Load >= 1 {
		load = matNames[e.LoadTyp]
	}
	return []gauge{
		{"BUCKET", e.Load / excCapacity, load},
		{"BOOM", e.Boom, ""},
		{"STICK", 1 - e.Stick, ""},
		{"CURL", e.Bucket, ""},
	}
}

func (e *Excavator) Render(g *Game, s *Scene, cockpit bool) {
	g.shadow(s, e.Pos, excTrackLen+8, excTrackW+8, e.Heading)
	if !g.cam.inView(ground(e.Pos, 50), 200) {
		return
	}
	c := rgba(245, 190, 40, 255)
	side := dir(e.Heading + math.Pi/2)
	// crawler tracks with moving cleats
	for _, sgn := range []float64{-1, 1} {
		tc := e.Pos.Add(side.Scale(sgn * excTrackW * 0.34))
		s.addPrism(ground(tc, 0), e.Heading, []ring{
			{-excTrackLen/2 + 10, excTrackLen/2 - 10, excTrackW * 0.15, 0, 2},
			{-excTrackLen / 2, excTrackLen / 2, excTrackW * 0.15, 9, 2},
			{-excTrackLen/2 + 4, excTrackLen/2 - 4, excTrackW * 0.15, 18, 2},
		}, colDark)
		s.detail(func() {
			const n = 10
			off := math.Mod(e.trackPhase, excTrackLen/n)
			for k := 0; k < n; k++ {
				x := -excTrackLen/2 + 8 + float64(k)*(excTrackLen-16)/n + off*0.9
				if x > excTrackLen/2-8 {
					continue
				}
				p := tc.Add(dir(e.Heading).Scale(x))
				s.addBox(ground(p, 18.6), 2, 0.8, excTrackW*0.15+0.5, e.Heading, rgba(70, 70, 74, 255))
			}
		})
		s.addCylinder(ground(tc.Add(dir(e.Heading).Scale(excTrackLen/2-12)), 9), 3, 2, 8, colSteel)
	}
	// slewing ring + upper structure
	ha := e.houseYaw()
	o := ground(e.Pos, excHouseY)
	s.addCylinder(ground(e.Pos, excHouseY-1), 26, 3, 16, colSteel)
	s.addPrism(o, ha, []ring{
		{-48, 30, 34, 0, 6},
		{-48, 30, 34, 18, 8},
		{-44, 26, 30, 24, 10},
	}, c)
	// counterweight at the back, engine cover, exhaust
	s.addPrism(o, ha, []ring{{-58, -44, 30, 2, 6}, {-58, -44, 30, 22, 7}, {-56, -46, 28, 26, 7}}, shadeC(c, 0.85))
	s.addCylinder(e.houseAt(-26, 14, excHouseY+32), 2.2, 16, 8, colDark)
	// cab on the left, with glass, pillars and roof
	cab := e.houseAt(16, -18, excHouseY)
	s.addPrismCaps(cab, ha, []ring{{-12, 12, 11, 24, 2}, {-12, 12, 11, 30, 2}}, shadeC(c, 0.9), true, false)
	s.addPrismCaps(cab, ha, []ring{{-12, 12, 11, 30, 2}, {-11, 9, 10, 46, 2}}, colGlass, false, false)
	s.addPrism(cab, ha, []ring{{-13, 11, 12, 46, 3}, {-12, 10, 11, 49, 3}}, shadeC(c, 0.9))
	beacon := rgba(140, 60, 20, 255)
	if e.Occupied && math.Sin(g.frameTime()*6) > 0 {
		beacon = rgba(255, 150, 35, 255)
	}
	s.addCylinder(e.houseAt(8, -18, excHouseY+52), 2.4, 5, 8, beacon)
	if cockpit {
		s.addBox(e.houseAt(27, -18, excHouseY+31), 1.5, 1.5, 7, ha, colDark)
	}

	// arm chain
	p0, p1, p2, p3 := e.armPoints(e.Boom, e.Stick, e.Bucket)
	at := func(p [2]float64) Vec3 { return e.houseAt(p[0], excArmZ, p[1]) }
	j0, j1, j2, j3 := at(p0), at(p1), at(p2), at(p3)
	beam := func(a, b Vec3, hh, hw float64, col color.RGBA) {
		d := b.Sub(a)
		s.addOBB(a.Add(d.Scale(0.5)), d, d.Len()/2+hh*0.5, hh, hw, col)
	}
	beam(j0, j1, 7, 6, c)
	beam(j1, j2, 5, 4.5, shadeC(c, 0.9))
	// boom and stick cylinders
	cylBase := e.houseAt(34, excArmZ, excHouseY+20)
	mid := j0.Add(j1.Sub(j0).Scale(0.55)).Add(Vec3{0, -5, 0})
	beam(cylBase, cylBase.Add(mid.Sub(cylBase).Scale(0.55)), 3, 3, colSteel)
	beam(cylBase.Add(mid.Sub(cylBase).Scale(0.5)), mid, 1.8, 1.8, colChrom)
	top := j0.Add(j1.Sub(j0).Scale(0.6)).Add(Vec3{0, 7, 0})
	beam(top, j1.Add(Vec3{0, 6, 0}), 2.2, 2.2, colChrom)

	// bucket: frame along the bucket link, a back shell and teeth
	_, _, tk := e.armAngles(e.Boom, e.Stick, e.Bucket)
	hf := rotY(Vec3{1, 0, 0}, ha)
	lat := rotY(Vec3{0, 0, 1}, ha)
	bf := hf.Scale(math.Cos(tk)).Add(Vec3{0, math.Sin(tk), 0}) // pivot -> teeth
	bn := lat.Cross(bf).Norm()                                 // the bucket's opening side
	if bn.Dot(Vec3{0, 1, 0}) < 0 && math.Sin(tk) > 0 {
		bn = bn.Scale(-1)
	}
	steel := rgba(82, 84, 90, 255)
	shell := j2.Add(bf.Scale(excBucketLen * 0.5)).Sub(bn.Scale(6))
	s.addOBB2(shell, bf, bn, excBucketLen*0.55, 6, 15, steel)
	for _, z := range []float64{-14, 14} {
		s.addOBB2(j2.Add(bf.Scale(excBucketLen*0.5)).Add(lat.Scale(z)), bf, bn, excBucketLen*0.5, 6, 1.2, shadeC(steel, 0.9))
	}
	for k := 0; k < 5; k++ {
		z := -12 + float64(k)*6
		s.addOBB2(j3.Add(bf.Scale(2.5)).Add(lat.Scale(z)), bf, bn, 3, 1.2, 1.6, rgba(132, 134, 140, 255))
	}
	if e.Load > 2 {
		r, gg, b := matColor(e.LoadTyp)
		f := clamp(e.Load/excCapacity, 0, 1)
		s.addOBB2(shell.Add(bn.Scale(5+3*f)), bf, bn, excBucketLen*0.45*math.Sqrt(f)+2, 2+3*f, 12, rgba(r, gg, b, 255))
	}
}
