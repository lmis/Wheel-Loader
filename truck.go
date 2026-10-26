package main

import (
	"math"
)

// Dump truck handling. Pos is the middle of the chassis; the rear tandem sits
// at truckRearX and the front axle at +36.
const (
	truckWheelbase = 71.0
	truckRearX     = -35.0
	truckMaxSteer  = 0.55
	truckMaxSpeed  = 250.0
	truckMaxRev    = 90.0
	truckAccel     = 200.0
	truckBrake     = 620.0
	truckBedHingeX = -51.0 // rear edge of the bed, where it pivots
	truckBedFloorY = 30.0
)

func (t *Truck) Name() string          { return "Dump " + t.ID }
func (t *Truck) Center() Vec           { return t.Pos }
func (t *Truck) Facing() float64       { return t.Angle }
func (t *Truck) SpeedVal() float64     { return t.Speed }
func (t *Truck) EnterRadius() float64  { return 95 }
func (t *Truck) SetOccupied(b bool)    { t.Occupied = b }
func (t *Truck) Jolt() float64         { return t.jolt }
func (t *Truck) CamDist() float64      { return 560 }
func (t *Truck) Colliders() []obstacle { return t.collidersAt(t.Pos, t.Angle) }

func (t *Truck) collidersAt(p Vec, ang float64) []obstacle {
	f := dir(ang)
	return []obstacle{{p.Add(f.Scale(40)), 33, t}, {p, 33, t}, {p.Sub(f.Scale(40)), 33, t}}
}

func (t *Truck) Drive(c *Controls) {
	if c == nil {
		t.throttle, t.steerCmd, t.bedCmd, t.brake = 0, 0, 0, true
		return
	}
	t.throttle, t.steerCmd, t.brake, t.bedCmd = c.Throttle, c.Steer, c.Brake, c.A1
}

// local converts truck-local (x forward, z right, y up) into world space.
func (t *Truck) local(x, z, y float64) Vec3 {
	return ground(t.Pos.Add(dir(t.Angle).Scale(x)).Add(dir(t.Angle+math.Pi/2).Scale(z)), y)
}

func (t *Truck) Seat() (Vec3, float64) { return t.local(40, -11, 47), t.Angle }

func (t *Truck) ExitPoint(side float64) Vec {
	return t.Pos.Add(dir(t.Angle).Scale(36)).Add(dir(t.Angle + math.Pi/2).Scale(-side * 52))
}

func (t *Truck) Update(dt float64, w *World) {
	// steering: wheels turn at a finite rate and self-centre
	if t.steerCmd != 0 {
		t.Steer += t.steerCmd * 1.4 * dt
	} else {
		t.Steer = approach(t.Steer, 0, 1.8*dt)
	}
	t.Steer = clamp(t.Steer, -truckMaxSteer, truckMaxSteer)

	// longitudinal: heavier when loaded
	loadF := 1 + t.Total()/t.Capacity*0.8
	if t.throttle != 0 {
		t.Speed += t.throttle * truckAccel / loadF * dt
		t.rpm = approach(t.rpm, 0.45+math.Abs(t.throttle)*0.55, 2*dt)
	} else {
		t.rpm = approach(t.rpm, math.Abs(t.Speed)/truckMaxSpeed*0.5+0.12, 1.5*dt)
	}
	if t.brake {
		t.Speed = approach(t.Speed, 0, truckBrake*dt)
	}
	t.Speed -= t.Speed * 0.7 * dt
	if t.throttle == 0 && !t.brake {
		t.Speed = approach(t.Speed, 0, 50*dt)
	}
	t.Speed = clamp(t.Speed, -truckMaxRev, truckMaxSpeed)

	// kinematic bicycle about the rear tandem, so the rear tracks properly
	rear := t.Pos.Add(dir(t.Angle).Scale(truckRearX)).Add(dir(t.Angle).Scale(t.Speed * dt))
	ang := t.Angle + t.Speed*math.Tan(t.Steer)/truckWheelbase*dt
	np := rear.Sub(dir(ang).Scale(truckRearX))
	if w != nil {
		if hit, n := collideCircles(w, t.collidersAt(np, ang), t); hit {
			np, ang = t.Pos.Add(n.Scale(2)), t.Angle
			t.jolt = math.Max(t.jolt, math.Abs(t.Speed)*0.6+40)
			t.Speed *= -0.2
		}
		np.X = clamp(np.X, 60, w.Width-60)
		np.Y = clamp(np.Y, 60, w.Height-60)
	}
	t.Pos, t.Angle = np, ang
	t.jolt *= math.Pow(0.02, dt)
	t.spin += t.Speed * dt / 15

	// tipping body: material slides out of the back once it is steep enough
	t.BedTilt = clamp(t.BedTilt+t.bedCmd*0.5*dt, 0, 1)
	if w != nil && t.BedTilt > 0.3 && t.Total() > 0 {
		t.pour(w, (t.BedTilt-0.3)*320*dt)
	}
}

// pour moves amount of the truck's contents onto the ground behind it,
// keeping the mix of material types.
func (t *Truck) pour(w *World, amount float64) {
	tot := t.Total()
	if amount > tot {
		amount = tot
	}
	if amount <= 0 {
		return
	}
	tail := t.Pos.Add(dir(t.Angle).Scale(truckBedHingeX - 16))
	side := dir(t.Angle + math.Pi/2)
	for ty := 1; ty < numMat; ty++ {
		share := amount * t.Contents[ty] / tot
		if share <= 0 {
			continue
		}
		t.Contents[ty] -= share
		for k := -1; k <= 1; k++ {
			p := tail.Add(side.Scale(float64(k) * 16))
			w.Grid.Add(p.X, p.Y, share/3, ty)
		}
		w.Grid.Relax(tail.X, tail.Y, 70, 1)
		w.spawnMaterialFall(tail, t.Angle+math.Pi, share, ty)
	}
}

func (t *Truck) Effects(w *World) {
	if spd := math.Abs(t.Speed); spd > 20 {
		inten := clamp(spd/truckMaxSpeed, 0, 1) * 0.7
		side := dir(t.Angle + math.Pi/2)
		rear := t.Pos.Add(dir(t.Angle).Scale(-40))
		w.spawnDriveDust(rear.Add(side.Scale(28)), t.Angle, inten)
		w.spawnDriveDust(rear.Sub(side.Scale(28)), t.Angle, inten)
	}
	if t.Occupied || math.Abs(t.Speed) > 5 {
		stack := t.Pos.Add(dir(t.Angle).Scale(31)).Add(dir(t.Angle + math.Pi/2).Scale(32.5))
		w.spawnExhaustAt(stack, 66, t.rpm)
	}
}

func (t *Truck) ControlHelp() []string {
	return []string{
		"{T}\taccelerate / reverse",
		"{S}\tsteer",
		"{B}\tbrake",
		"{1}\traise / lower dump body",
	}
}

func (t *Truck) Gauges() []gauge {
	tot := t.Total()
	dom := 0
	best := 0.0
	for ty := 1; ty < numMat; ty++ {
		if t.Contents[ty] > best {
			best, dom = t.Contents[ty], ty
		}
	}
	load := "empty"
	if tot >= 1 {
		load = matNames[dom]
	}
	bed := "DOWN"
	if t.BedTilt > 0.3 {
		bed = "TIPPING"
	} else if t.BedTilt > 0.02 {
		bed = "RAISED"
	}
	return []gauge{
		{"LOAD", tot / t.Capacity, load},
		{"BODY", t.BedTilt, bed},
	}
}

// Render draws the truck: steering front wheels, cab, chassis and a dump
// body that pivots about its rear hinge.
func (t *Truck) Render(g *Game, s *Scene, cockpit bool) {
	g.shadow(s, t.Pos, 118, 74, t.Angle)
	if !g.cam.inView(ground(t.Pos, 30), 110) {
		return
	}
	ang := t.Angle
	at := t.local
	body := rgba(222, 140, 32, 255)

	for _, x := range []float64{36, -24, -46} {
		yaw := ang
		if x > 0 {
			yaw += t.Steer
		}
		for _, z := range []float64{-27, 27} {
			s.addWheel(at(x, z, 15), 15, 12, yaw, t.spin, colTire, colSteel)
		}
	}
	// chassis rails + fuel tank
	s.addBox(at(-4, 0, 19), 52, 5, 18, ang, colDark)
	s.addCylinder(at(14, -22, 22), 6, 12, 10, colSteel)
	// cab: lower body, glass band, roof
	o := at(0, 0, 0)
	s.addPrismCaps(o, ang, []ring{{26, 56, 30, 22, 3}, {26, 56, 30, 38, 3}}, body, true, false)
	s.addPrismCaps(o, ang, []ring{{26, 56, 30, 38, 3}, {28, 51, 28, 51, 3}}, colGlass, false, false)
	s.addPrism(o, ang, []ring{{27, 52, 29.5, 51, 4}, {28, 50, 29, 55, 6}}, body)
	s.detail(func() {
		s.addBox(at(57.5, 0, 22), 2.5, 6, 30, ang, colDark)
		for _, z := range []float64{-22, 22} {
			s.addBox(at(56.8, z, 30), 0.8, 3, 4, ang, rgba(255, 246, 210, 255))
		}
	})
	s.addCylinder(at(31, 32.5, 50), 2.2, 30, 8, colDark) // exhaust beside the cab
	if cockpit {
		// dashboard across the windscreen base, steering wheel at the left seat
		s.addBox(at(53, 0, 39), 3, 1.5, 27, ang, rgba(62, 64, 70, 255))
		s.addSteeringWheel(at(47.5, -11, 41), ang, 4, rgba(30, 30, 33, 255))
	}

	// ---- dump body, pivoting about the rear hinge ----
	bed := t.bedRect()
	bc := bed.Center.Sub(t.Pos).Rot(-ang)
	bx0, bx1 := bc.X-bed.W/2, bc.X+bed.W/2
	hw := bed.H / 2
	th := t.BedTilt * 0.95
	si, co := math.Sincos(th)
	hx, hy := bx0, 27.0
	bedAt := func(x, y, z float64) Vec3 {
		dx, dy := x-hx, y-hy
		return at(hx+dx*co-dy*si, z, hy+dx*si+dy*co)
	}
	f3 := rotY(Vec3{1, 0, 0}, ang)
	up := Vec3{0, 1, 0}
	bf := f3.Scale(co).Add(up.Scale(si))
	bu := up.Scale(co).Sub(f3.Scale(si))
	box := func(x, y, z, hl, hh, hwid float64, c [4]uint8) {
		s.addOBB2(bedAt(x, y, z), bf, bu, hl, hh, hwid, rgba(c[0], c[1], c[2], c[3]))
	}
	bodyC := [4]uint8{body.R, body.G, body.B, 255}
	box(bc.X, 27, 0, bed.W/2, 3, hw, [4]uint8{colSteel.R, colSteel.G, colSteel.B, 255}) // floor
	for _, sg := range []float64{-1, 1} {
		box(bc.X+0.75, 42, sg*(hw-1.5), bed.W/2+0.75, 12, 1.8, bodyC) // side walls
	}
	// headboard, split at cab-roof height so neither half ties with the roof
	box(bx1+0.5, 41, 0, 1.5, 14, hw, bodyC)
	box(bx1+0.5, 59, 0, 1.5, 4, hw, bodyC)
	box(bx1+2.5, 64.5, 0, 3.5, 1.5, hw, [4]uint8{uint8(float64(body.R) * 0.9), uint8(float64(body.G) * 0.9), uint8(float64(body.B) * 0.9), 255})
	// tailgate hangs from its top hinge, so it swings open as the body tips
	hinge := bedAt(bx0-1.5, 55, 0)
	s.addOBB2(hinge.Sub(Vec3{0, 13, 0}), f3, up, 2, 13, hw, shadeC(body, 0.92))
	s.detail(func() { s.addBox(at(bx0-3, 0, 14), 1, 8, hw+2, ang, colTire) }) // mud flaps
	// hoist ram, visible once the body lifts
	if t.BedTilt > 0.02 {
		a, b := at(bx1-10, 0, 22), bedAt(bx1-12, 24, 0)
		d := b.Sub(a)
		s.addOBB(a.Add(d.Scale(0.5)), d, d.Len()/2, 2.5, 2.5, colChrom)
	}
	// contents heaped above the bed floor, tilting with it
	if tot := t.Total(); tot > 1 {
		frac := clamp(tot/t.Capacity, 0, 1)
		dom, best := 0, 0.0
		for ty := 1; ty < numMat; ty++ {
			if t.Contents[ty] > best {
				best, dom = t.Contents[ty], ty
			}
		}
		r, gg, b := matColor(dom)
		l1, l2 := 4+frac*14, frac*10
		box(bc.X, truckBedFloorY+l1/2, 0, bed.W/2-3, l1/2, hw-3, [4]uint8{r, gg, b, 255})
		if l2 > 0.5 {
			box(bc.X-3, truckBedFloorY+l1+l2/2, 0, bed.W*0.3, l2/2, hw*0.55, [4]uint8{clampByte(float64(r) * 1.06), clampByte(float64(gg) * 1.06), clampByte(float64(b) * 1.06), 255})
		}
	}
}
