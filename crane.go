package main

import (
	"image/color"
	"math"
)

// TowerCrane is a static tower crane with a cab at the top of the mast. The
// jib slews, a trolley runs along it, and a hook on a swinging rope lifts
// cargo.
type TowerCrane struct {
	Base     Vec
	Slew     float64 // jib heading
	Trolley  float64 // trolley distance from the mast
	HookLen  float64 // rope length below the jib
	Occupied bool
	Easy     bool // toddler mode: brisk slewing and a well-damped hook

	hook, hookV Vec // swinging hook position on the ground plane
	held        *Cargo
	heldYaw     float64 // load yaw relative to the jib while held

	slewCmd, trolleyCmd, hoistCmd, grabCmd float64
	slewV                                  float64
	grabLatch                              bool
}

const (
	craneMastH   = 330.0
	craneJibY    = craneMastH + 14 // bottom chords of the jib
	craneJibLen  = 420.0
	craneCounter = 130.0
	craneGravity = 137.0 // ~9.8 m/s² at ~14 units per metre
)

func NewTowerCrane(base Vec, slew float64) *TowerCrane {
	c := &TowerCrane{Base: base, Slew: slew, Trolley: 200, HookLen: 220}
	c.hook = c.trolleyPos()
	return c
}

func (c *TowerCrane) Center() Vec           { return c.Base }
func (c *TowerCrane) Facing() float64       { return c.Slew }
func (c *TowerCrane) SpeedVal() float64     { return 0 }
func (c *TowerCrane) SetOccupied(b bool)    { c.Occupied = b }
func (c *TowerCrane) Jolt() float64         { return 0 }
func (c *TowerCrane) CamDist() float64      { return 520 }
func (c *TowerCrane) Effects(w *World)      {}
func (c *TowerCrane) Colliders() []obstacle { return []obstacle{{c.Base, 22, c}} }

func (c *TowerCrane) Drive(ctl *Controls) {
	if ctl == nil {
		c.slewCmd, c.trolleyCmd, c.hoistCmd, c.grabCmd = 0, 0, 0, 0
		return
	}
	c.slewCmd, c.trolleyCmd, c.hoistCmd, c.grabCmd = ctl.Steer, ctl.Throttle, ctl.A1, ctl.A3
}

func (c *TowerCrane) trolleyPos() Vec { return c.Base.Add(dir(c.Slew).Scale(c.Trolley)) }

// HookY is the height of the bottom of the hook block.
func (c *TowerCrane) HookY() float64 { return craneJibY - 6 - c.HookLen }

// jibAt converts slewing-frame coordinates (x along the jib, z right, y up).
func (c *TowerCrane) jibAt(x, z, y float64) Vec3 {
	return ground(c.Base.Add(dir(c.Slew).Scale(x)).Add(dir(c.Slew+math.Pi/2).Scale(z)), y)
}

// CamTarget aims the third-person camera at the hook rather than the mast.
func (c *TowerCrane) CamTarget() (Vec, float64, float64) {
	return c.hook, math.Max(c.HookY()-20, 20), c.Slew
}

// Seat: the cab hangs off the side of the slewing unit, looking along the
// jib and down at the hook.
func (c *TowerCrane) Seat() (Vec3, float64) { return c.jibAt(20.5, -17, craneMastH-9), c.Slew }
func (c *TowerCrane) SeatPitch() float64    { return -0.5 }

// RenderCockpit adds the inside of the crane cab: a framed front window and
// glass floor, a desk with two screens, and joysticks on the armrests.
func (c *TowerCrane) RenderCockpit(s *Scene) {
	const z0, H = -17.0, craneMastH
	at := func(x, z, y float64) Vec3 { return c.jibAt(x, z0+z, y) }
	trim := rgba(56, 58, 64, 255)
	fw := rotY(Vec3{1, 0, 0}, c.Slew)
	up := Vec3{0, 1, 0}
	// window frame round the front glass, and the frame of the glass floor
	for _, z := range []float64{-8.4, 8.4} {
		s.addCapsule(at(31.6, z, H-24), at(31.6, z, H-1), 0.7, 0.7, trim) // corner pillars
		s.addCapsule(at(31.6, z, H-23.6), at(14, z, H-23.6), 0.5, 0.5, trim)
	}
	s.addCapsule(at(31.6, -8.4, H-1.2), at(31.6, 8.4, H-1.2), 0.7, 0.7, trim)   // header
	s.addCapsule(at(31.6, -8.4, H-16.5), at(31.6, 8.4, H-16.5), 0.5, 0.5, trim) // sill between window and floor glass
	s.addCapsule(at(31.6, -8.4, H-23.6), at(31.6, 8.4, H-23.6), 0.5, 0.5, trim) // front edge of the floor
	s.addCapsule(at(24, -8.4, H-23.6), at(24, 8.4, H-23.6), 0.4, 0.4, trim)     // floor glass bar
	s.addOBB2(at(17, 0, H-23.8), fw, up, 5, 0.3, 8.6, rgba(70, 72, 78, 255))    // solid floor under the seat
	// desk with two glowing screens
	s.addRoundPrism(at(0, 0, 0), c.Slew, []ring{{27.5, 30.5, 6, H - 19, 1}, {27.8, 30.2, 6, H - 16.8, 1}}, rgba(70, 72, 78, 255), false, true)
	for _, z := range []float64{-3, 3} {
		s.addOBB2(at(28.6, z, H-15.6), fw.Add(up.Scale(0.6)), up, 0.15, 1.1, 1.8, colDark)
		s.detail(func() {
			s.glowing(func() {
				s.addOBB2(at(28.75, z, H-15.55), fw.Add(up.Scale(0.6)), up, 0.05, 0.9, 1.6, rgba(120, 200, 255, 255))
			})
		})
	}
	// armrests with joysticks either side of the seat
	for _, z := range []float64{-4.6, 4.6} {
		s.addRoundPrism(at(0, z, 0), c.Slew, []ring{{19, 25.5, 1.1, H - 16, 0.6}, {19, 25.5, 1.1, H - 14.6, 0.6}}, colDark, false, true)
		base := at(24.6, z, H-14.6)
		s.addCapsule(base, base.Add(Vec3{0, 1.6, 0}), 0.22, 0.2, colDark)
		s.addSphere(base.Add(Vec3{0, 1.85, 0}), 0.45, rgba(220, 40, 36, 255))
	}
}

func (c *TowerCrane) ExitPoint(side float64) Vec {
	return c.Base.Add(dir(c.Slew + math.Pi/2).Scale(-side * 42))
}

func (c *TowerCrane) Update(dt float64, w *World) {
	// heavy slewing inertia: it takes a moment to get going and to stop
	maxV, slewAcc, damp := 0.32, 0.35, 0.35
	if c.Easy {
		maxV, slewAcc, damp = 0.5, 1.4, 2.5
	}
	c.slewV = approach(c.slewV, c.slewCmd*maxV, slewAcc*dt)
	c.Slew += c.slewV * dt
	c.Trolley = clamp(c.Trolley+c.trolleyCmd*100*dt, 70, craneJibLen-14)
	c.HookLen = clamp(c.HookLen-c.hoistCmd*80*dt, 12, craneJibY-12)

	// the hook swings like a pendulum hanging under the trolley
	L := math.Max(c.HookLen, 20)
	tp := c.trolleyPos()
	acc := tp.Sub(c.hook).Scale(craneGravity / L).Sub(c.hookV.Scale(damp))
	c.hookV = c.hookV.Add(acc.Scale(dt))
	c.hook = c.hook.Add(c.hookV.Scale(dt))

	// keep the hook (and anything on it) above whatever is underneath
	bottom := c.HookY()
	if c.held != nil {
		bottom -= 6 + c.held.H
	}
	floor := 0.0
	if w != nil {
		if cx, cy, ok := w.Grid.cellAt(c.hook.X, c.hook.Y); ok {
			floor = pileHeight(w.Grid.at(cx, cy).Amount)
		}
		if t := w.TruckAt(c.hook); t != nil && c.held != nil {
			floor = math.Max(floor, truckBedFloorY)
		}
		for _, o := range w.Cargo {
			if o != c.held && !o.held && o.Pos.Sub(c.hook).Len() < o.footRadius()*0.8 {
				floor = math.Max(floor, o.Y+o.H)
			}
		}
	}
	if bottom < floor {
		c.HookLen -= floor - bottom
	}

	// hook on / release (edge-triggered)
	switch {
	case c.grabCmd > 0.5 && !c.grabLatch:
		c.grabLatch = true
		if c.held == nil && w != nil {
			c.attach(w)
		}
	case c.grabCmd < -0.5 && !c.grabLatch:
		c.grabLatch = true
		if c.held != nil {
			c.held.held = false
			c.held.vy = 0
			c.held = nil
		}
	case c.grabCmd == 0:
		c.grabLatch = false
	}
	if h := c.held; h != nil {
		h.Pos = c.hook
		h.Y = c.HookY() - 6 - h.H
		h.Yaw = c.Slew + c.heldYaw
	}
}

// attach hooks the nearest item whose top is just below the hook.
func (c *TowerCrane) attach(w *World) {
	var best *Cargo
	bestD := 1e9
	for _, o := range w.Cargo {
		if o.held {
			continue
		}
		d := o.Pos.Sub(c.hook).Len()
		top := o.Y + o.H
		if d < o.footRadius()+10 && math.Abs(top-c.HookY()) < 18 && d < bestD {
			best, bestD = o, d
		}
	}
	if best != nil {
		best.held, best.truck = true, nil
		c.held, c.heldYaw = best, best.Yaw-c.Slew
	}
}

func (c *TowerCrane) Render(g *Game, s *Scene) {
	g.blobShadow(s, c.hook, 7, 7, c.Slew) // helps to aim the hook
	yellow := rgba(246, 192, 22, 255)
	beam := func(a, b Vec3, r float64, col color.RGBA) {
		d := b.Sub(a)
		s.addOBB(a.Add(d.Scale(0.5)), d, d.Len()/2, r, r, col)
	}
	// footing and mast lattice
	if g.cam.inView(ground(c.Base, craneMastH/2), craneMastH/2+30) {
		s.addBox(ground(c.Base, 4), 34, 4, 34, 0, rgba(150, 148, 142, 255))
		const half, step = 10.0, 30.0
		corner := func(i int, y float64) Vec3 {
			sx, sz := []float64{-1, 1, 1, -1}[i], []float64{-1, -1, 1, 1}[i]
			return ground(c.Base, y).Add(Vec3{sx * half, 0, sz * half})
		}
		for i := 0; i < 4; i++ {
			beam(corner(i, 8), corner(i, craneMastH), 1.6, yellow)
		}
		for y := 8.0; y+step <= craneMastH+0.1; y += step {
			for i := 0; i < 4; i++ {
				j := (i + 1) % 4
				beam(corner(i, y+step), corner(j, y+step), 0.9, yellow)
				if int(y/step)%2 == 0 {
					beam(corner(i, y), corner(j, y+step), 0.8, yellow)
				} else {
					beam(corner(j, y), corner(i, y+step), 0.8, yellow)
				}
			}
		}
	}

	// slewing unit, cab, apex and jibs
	top := c.jibAt(0, 0, craneMastH)
	if !g.cam.inView(c.jibAt(craneJibLen/2-60, 0, craneJibY), craneJibLen/2+140) {
		return
	}
	s.addBox(top.Add(Vec3{0, 6, 0}), 14, 6, 14, c.Slew, shadeC(yellow, 0.9))
	cab := c.jibAt(20, -17, 0)
	s.addPrismCaps(cab, c.Slew, []ring{{-10, 12, 9, craneMastH - 24, 2}, {-10, 12, 9, craneMastH - 18, 2}}, rgba(235, 235, 230, 255), true, false)
	s.addPrismCaps(cab, c.Slew, []ring{{-10, 12, 9, craneMastH - 18, 2}, {-10, 10, 8.5, craneMastH - 3, 2}}, colGlass, false, false)
	s.addPrism(cab, c.Slew, []ring{{-11, 12, 10, craneMastH - 3, 3}, {-10, 11, 9, craneMastH, 3}}, rgba(235, 235, 230, 255))
	// apex and tie bars
	apex := c.jibAt(0, 0, craneJibY+72)
	for _, z := range []float64{-7, 7} {
		beam(c.jibAt(-6, z, craneJibY), apex, 1.4, yellow)
		beam(c.jibAt(6, z, craneJibY), apex, 1.4, yellow)
	}
	beam(apex, c.jibAt(craneJibLen*0.55, 0, craneJibY+22), 0.7, colSteel)
	beam(apex, c.jibAt(craneJibLen-30, 0, craneJibY+22), 0.7, colSteel)
	beam(apex, c.jibAt(-craneCounter+8, 0, craneJibY+4), 0.7, colSteel)
	// jib: triangular lattice, two bottom chords and a top chord
	const seg = 35.0
	jb := func(x, z float64) Vec3 { return c.jibAt(x, z, craneJibY) }
	jt := func(x float64) Vec3 { return c.jibAt(x, 0, craneJibY+22) }
	beam(jb(-4, -7), jb(craneJibLen, -7), 1.2, yellow)
	beam(jb(-4, 7), jb(craneJibLen, 7), 1.2, yellow)
	beam(jt(-4), jt(craneJibLen-10), 1.2, yellow)
	for x := 0.0; x+seg <= craneJibLen; x += seg {
		beam(jb(x, -7), jt(x+seg/2), 0.7, yellow)
		beam(jt(x+seg/2), jb(x+seg, -7), 0.7, yellow)
		beam(jb(x, 7), jt(x+seg/2), 0.7, yellow)
		beam(jt(x+seg/2), jb(x+seg, 7), 0.7, yellow)
		beam(jb(x+seg, -7), jb(x+seg, 7), 0.6, yellow)
	}
	// counter-jib with walkway and concrete counterweights
	s.addBox(c.jibAt(-craneCounter/2, 0, craneJibY+1), craneCounter/2, 1.5, 9, c.Slew, shadeC(yellow, 0.85))
	for k := 0; k < 3; k++ {
		x := -craneCounter + 12 + float64(k)*14
		s.addBox(c.jibAt(x, 0, craneJibY-12), 6, 12, 11, c.Slew, rgba(168, 166, 160, 255))
	}
	// trolley, rope and hook block
	tx := c.Trolley
	s.addBox(c.jibAt(tx, 0, craneJibY-3), 7, 3, 9, c.Slew, rgba(60, 62, 68, 255))
	hy := c.HookY()
	hook3 := ground(c.hook, hy)
	for _, z := range []float64{-2, 2} {
		beam(c.jibAt(tx, z, craneJibY-6), hook3.Add(rotY(Vec3{0, 12, z}, c.Slew)), 0.35, rgba(40, 40, 44, 255))
	}
	s.addBox(hook3.Add(Vec3{0, 9, 0}), 4, 3, 5, c.Slew, rgba(230, 60, 40, 255))
	s.addBox(hook3.Add(Vec3{0, 4.5, 0}), 1.2, 2, 1.2, c.Slew, colSteel)
	s.addOBB2(hook3.Add(rotY(Vec3{1.5, 2, 0}, c.Slew)), Vec3{0, 1, 0}, rotY(Vec3{1, 0, 0}, c.Slew), 1.4, 2.5, 0.7, colSteel)
	// slings to the load's corners
	if h := c.held; h != nil {
		for _, sx := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				corner := ground(h.Pos.Add(dir(h.Yaw).Scale(sx*h.W*0.4)).Add(dir(h.Yaw+math.Pi/2).Scale(sz*h.D*0.4)), h.Y+h.H)
				beam(hook3, corner, 0.3, rgba(220, 200, 60, 255))
			}
		}
	}
}
