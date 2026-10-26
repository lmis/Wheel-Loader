package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func rgba(r, g, b, a uint8) color.RGBA { return color.RGBA{r, g, b, a} }

// render builds the whole 3D scene for this frame and draws it, then the HUD.
func (g *Game) render(screen *ebiten.Image) {
	screen.Fill(rgba(150, 170, 190, 255)) // sky
	g.drawSkyGradient(screen)

	s := &g.scene
	s.reset()

	g.buildGround3D(s)
	g.buildMaterial3D(s)
	g.buildShadows(s)
	g.buildProps3D(s)
	g.buildTrucks3D(s)
	g.buildLoader3D(s)
	g.buildParticles3D(s)

	s.Flush(screen, g.cam, g.ground, g.art.white)

	g.drawHUD(screen)
}

// drawSkyGradient paints a simple vertical gradient band for the sky above
// the horizon so the void does not read as flat.
func (g *Game) drawSkyGradient(screen *ebiten.Image) {
	top := rgba(120, 150, 180, 255)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(screenW, screenH*0.6)
	op.ColorScale.ScaleWithColor(top)
	screen.DrawImage(g.art.white, op)
}

// ---- ground ----

func (g *Game) buildGround3D(s *Scene) {
	const tile = 110.0
	mn, mx := g.cam.VisibleBounds()
	x0 := math.Max(0, math.Floor(mn.X/tile)*tile)
	y0 := math.Max(0, math.Floor(mn.Y/tile)*tile)
	x1 := math.Min(g.world.Width, mx.X)
	y1 := math.Min(g.world.Height, mx.Y)
	for y := y0; y < y1; y += tile {
		for x := x0; x < x1; x += tile {
			xe := math.Min(x+tile, g.world.Width)
			ye := math.Min(y+tile, g.world.Height)
			s.addGroundQuad(
				Vec3{x, 0, y}, Vec3{xe, 0, y}, Vec3{xe, 0, ye}, Vec3{x, 0, ye}, 1.0)
		}
	}
}

// ---- material grid as 3D heaps ----

func (g *Game) buildMaterial3D(s *Scene) {
	grid := g.world.Grid
	mn, mx := g.cam.VisibleBounds()
	cxMin := int(mn.X/grid.CellSize) - 1
	cyMin := int(mn.Y/grid.CellSize) - 1
	cxMax := int(mx.X/grid.CellSize) + 1
	cyMax := int(mx.Y/grid.CellSize) + 1
	cs := grid.CellSize
	for cy := cyMin; cy <= cyMax; cy++ {
		for cx := cxMin; cx <= cxMax; cx++ {
			if !grid.inBounds(cx, cy) {
				continue
			}
			c := grid.at(cx, cy)
			if c.Amount < 2 {
				continue
			}
			h := clamp(c.Amount*0.5, 3, 46)
			px := float64(cx) * cs
			py := float64(cy) * cs
			r, gg, b := matColor(c.Type)
			col := rgba(r, gg, b, 255)
			// a column: top quad + the four side skirts (culled automatically)
			cxw := px + cs/2
			cyw := py + cs/2
			s.addBox(Vec3{cxw, h / 2, cyw}, cs*0.55, h/2, cs*0.55, 0, col)
		}
	}
}

// ---- shadows (flat dark quads on the ground) ----

func (g *Game) buildShadows(s *Scene) {
	sh := rgba(18, 22, 20, 90)
	off := V(14, 18) // cast offset away from the light
	quad := func(center Vec, w, h, ang float64) {
		cc := center.Add(off)
		c := dir(ang)
		p := dir(ang + math.Pi/2)
		a := ground(cc.Add(c.Scale(-w/2)).Add(p.Scale(-h/2)), 0.4)
		b := ground(cc.Add(c.Scale(w/2)).Add(p.Scale(-h/2)), 0.4)
		d := ground(cc.Add(c.Scale(w/2)).Add(p.Scale(h/2)), 0.4)
		e := ground(cc.Add(c.Scale(-w/2)).Add(p.Scale(h/2)), 0.4)
		s.addQuad(a, b, d, e, sh, false, false)
	}
	for i := range g.world.Props {
		p := &g.world.Props[i]
		if !p.Solid && p.Kind != propCone {
			continue
		}
		quad(p.Rect.Center, p.Rect.W+10, p.Rect.H+10, p.Rect.Angle)
	}
	for _, t := range g.world.Trucks {
		quad(t.Pos, 80, 120, t.Angle)
	}
	// loader shadow (two halves)
	l := g.loader
	quad(l.RearAxle(), bodyWidth+14, rearLen+20, l.RearHeading())
	quad(l.FrontAxle(), bodyWidth+14, frontLen+20, l.FrontHeading())
}

// ---- props ----

func propHeight(k PropKind) float64 {
	switch k {
	case propBuilding:
		return 74
	case propCabin:
		return 46
	case propTank:
		return 54
	case propBarrier:
		return 24
	case propCone:
		return 20
	case propSign:
		return 34
	case propExcavator:
		return 40
	case propParkedLoader:
		return 38
	}
	return 30
}

func (g *Game) buildProps3D(s *Scene) {
	for i := range g.world.Props {
		p := &g.world.Props[i]
		c := rgba(p.Color[0], p.Color[1], p.Color[2], 255)
		rc := p.Rect
		h := propHeight(p.Kind)
		switch p.Kind {
		case propBuilding:
			s.addBox(ground(rc.Center, h/2), rc.W/2, h/2, rc.H/2, rc.Angle, c)
			// roof HVAC units
			s.addBox(ground(rc.Center.Add(dir(rc.Angle).Scale(-rc.W*0.18)), h+8), 18, 8, 12, rc.Angle, rgba(150, 152, 156, 255))
			s.addBox(ground(rc.Center.Add(dir(rc.Angle).Scale(rc.W*0.22)), h+6), 12, 6, 10, rc.Angle, rgba(135, 137, 142, 255))
		case propCabin:
			s.addBox(ground(rc.Center, h/2), rc.W/2, h/2, rc.H/2, rc.Angle, c)
			s.addBox(ground(rc.Center, h+2), rc.W/2-2, 2, rc.H/2-2, rc.Angle, shade(c, 1.15))
		case propTank:
			s.addCylinder(ground(rc.Center, h/2), rc.W/2, h, 12, c)
		case propBarrier:
			s.addBox(ground(rc.Center, h/2), rc.W/2, h/2, rc.H/2, rc.Angle, c)
			s.addBox(ground(rc.Center, h*0.5), rc.W/2+1, h*0.18, rc.H/2+1, rc.Angle, rgba(240, 240, 235, 255))
		case propCone:
			s.addCylinder(ground(rc.Center, h*0.3), rc.W*0.7, h*0.6, 8, rgba(240, 120, 30, 255))
			s.addCylinder(ground(rc.Center, h*0.72), rc.W*0.3, h*0.3, 8, rgba(240, 120, 30, 255))
			s.addBox(ground(rc.Center, 2), rc.W, 2, rc.W, 0, rgba(40, 40, 44, 255))
		case propSign:
			s.addBox(ground(rc.Center.Add(dir(0).Scale(0)), h*0.5), 2, h*0.5, 2, 0, rgba(120, 120, 124, 255))
			s.addBox(ground(rc.Center, h), rc.W*0.5, rc.H*0.5, 3, 0.78, rgba(250, 210, 40, 255))
		case propExcavator:
			g.buildExcavator3D(s, p)
		case propParkedLoader:
			s.addBox(ground(rc.Center, 22), rc.W*0.42, 12, rc.H*0.5, rc.Angle, c)
			s.addBox(ground(rc.Center.Add(dir(rc.Angle).Scale(rc.W*0.1)), 40), rc.W*0.2, 10, rc.H*0.4, rc.Angle, shade(c, 0.9))
			fc := rc.Center.Add(dir(rc.Angle).Scale(rc.W/2 + 8))
			s.addBox(ground(fc, 14), 8, 14, rc.H*0.5, rc.Angle, rgba(60, 60, 64, 255))
			for _, sx := range []float64{-1, 1} {
				for _, fx := range []float64{-1, 1} {
					wc := rc.Center.Add(dir(rc.Angle).Scale(fx * rc.W * 0.32)).Add(dir(rc.Angle + math.Pi/2).Scale(sx * rc.H * 0.5))
					s.addWheel(ground(wc, 14), 14, 10, rc.Angle, rgba(30, 30, 32, 255))
				}
			}
		}
	}
}

func (g *Game) buildExcavator3D(s *Scene, p *Prop) {
	rc := p.Rect
	c := rgba(p.Color[0], p.Color[1], p.Color[2], 255)
	side := dir(rc.Angle + math.Pi/2)
	// crawler tracks
	for _, sgn := range []float64{-1, 1} {
		tc := rc.Center.Add(side.Scale(sgn * rc.H * 0.34))
		s.addBox(ground(tc, 10), rc.W/2, 10, rc.H*0.16, rc.Angle, rgba(45, 45, 48, 255))
	}
	// slewing house
	ha := rc.Angle + 0.35
	house := rc.Center
	s.addBox(ground(house, 34), rc.W*0.3, 16, rc.H*0.36, ha, c)
	s.addBox(ground(house.Add(dir(ha).Scale(rc.W*0.18)), 40), 10, 8, rc.H*0.28, ha, rgba(150, 200, 225, 235))
	// boom -> dipper -> bucket
	base := ground(house.Add(dir(ha).Scale(rc.W*0.26)), 36)
	elbow := base.Add(Vec3{math.Cos(ha) * 46, 26, math.Sin(ha) * 46})
	wrist := elbow.Add(Vec3{math.Cos(ha) * 42, -40, math.Sin(ha) * 42})
	s.addOBB(base.Add(elbow.Sub(base).Scale(0.5)), elbow.Sub(base), base.Sub(elbow).Len()/2, 6, 6, c)
	s.addOBB(elbow.Add(wrist.Sub(elbow).Scale(0.5)), wrist.Sub(elbow), wrist.Sub(elbow).Len()/2, 5, 5, shade(c, 0.8))
	s.addBox(Vec3{wrist.X, wrist.Y, wrist.Z}, 10, 12, rc.H*0.2, ha, rgba(48, 48, 52, 255))
}

// ---- trucks ----

func (g *Game) buildTrucks3D(s *Scene) {
	for _, t := range g.world.Trucks {
		ang := t.Angle
		fwd := dir(ang)
		side := dir(ang + math.Pi/2)
		// wheels
		for _, f := range []float64{-1, -0.35, 0.28} {
			for _, sgn := range []float64{-1, 1} {
				wc := t.Pos.Add(fwd.Scale(f * 42)).Add(side.Scale(sgn * 34))
				s.addWheel(ground(wc, 15), 15, 12, ang, rgba(26, 26, 28, 255))
			}
		}
		// chassis
		s.addBox(ground(t.Pos, 26), 56, 10, 32, ang, rgba(58, 62, 70, 255))
		// cab
		cab := t.Pos.Add(fwd.Scale(42))
		s.addBox(ground(cab, 40), 18, 22, 30, ang, rgba(210, 160, 40, 255))
		s.addBox(ground(cab.Add(fwd.Scale(10)), 44), 6, 12, 26, ang, rgba(150, 200, 230, 235))
		// dump bed: floor + 4 low walls (open top)
		bed := t.bedRect()
		bc := bed.Center
		fh := 36.0
		s.addBox(ground(bc, fh-8), bed.W/2, 4, bed.H/2, ang, rgba(44, 46, 52, 255)) // floor
		s.addBox(ground(bc.Add(fwd.Scale(-bed.W/2)), fh), 3, 12, bed.H/2, ang, rgba(66, 70, 78, 255))
		s.addBox(ground(bc.Add(fwd.Scale(bed.W/2)), fh), 3, 12, bed.H/2, ang, rgba(66, 70, 78, 255))
		s.addBox(ground(bc.Add(side.Scale(-bed.H/2)), fh), bed.W/2, 12, 3, ang, rgba(66, 70, 78, 255))
		s.addBox(ground(bc.Add(side.Scale(bed.H/2)), fh), bed.W/2, 12, 3, ang, rgba(66, 70, 78, 255))
		// contents heap
		if tot := t.Total(); tot > 1 {
			frac := clamp(tot/t.Capacity, 0, 1)
			dom, best := 0, 0.0
			for ty := 1; ty < numMat; ty++ {
				if t.Contents[ty] > best {
					best, dom = t.Contents[ty], ty
				}
			}
			r, gg, b := matColor(dom)
			hh := 6 + frac*22
			s.addBox(ground(bc, fh-10+hh/2), (bed.W/2-5)*math.Sqrt(frac), hh/2, (bed.H/2-5)*math.Sqrt(frac), ang, rgba(r, gg, b, 255))
		}
	}
}

// ---- the loader ----

const (
	wheelR     = 18.0
	bodyY      = 24.0 // chassis vertical centre
	liftBaseY  = 28.0
	liftRangeY = 56.0
)

func (g *Game) buildLoader3D(s *Scene) {
	l := g.loader
	fh := l.FrontHeading()
	rh := l.RearHeading()
	front := l.FrontAxle()
	rear := l.RearAxle()
	joint := l.Joint

	yellow := rgba(245, 194, 30, 255)
	yellowD := rgba(205, 158, 22, 255)
	dark := rgba(46, 48, 52, 255)

	// subtle body pitch/roll affect the chassis height a touch
	bob := l.bounce * 0.3

	// wheels
	for _, sgn := range []float64{-1, 1} {
		rc := rear.Add(dir(rh + math.Pi/2).Scale(sgn * (bodyWidth/2 + 2)))
		fc := front.Add(dir(fh + math.Pi/2).Scale(sgn * (bodyWidth/2 + 2)))
		s.addWheel(ground(rc, wheelR), wheelR, 13, rh, rgba(28, 28, 30, 255))
		s.addWheel(ground(fc, wheelR), wheelR, 13, fh, rgba(28, 28, 30, 255))
	}

	// rear body (engine compartment)
	rearCenter := joint.Sub(dir(rh).Scale(rearLen * 0.5))
	s.addBox(ground(rearCenter, bodyY+bob), rearLen*0.5, 13, bodyWidth*0.5, rh, yellowD)
	// engine hood detail
	s.addBox(ground(rearCenter.Add(dir(rh).Scale(-rearLen*0.2)), bodyY+16+bob), rearLen*0.22, 6, bodyWidth*0.4, rh, shade(yellowD, 0.9))
	// exhaust stack
	stack := rearCenter.Add(dir(rh).Scale(-rearLen * 0.35)).Add(dir(rh + math.Pi/2).Scale(bodyWidth * 0.3))
	s.addCylinder(ground(stack, bodyY+26+bob), 3, 34, 6, dark)

	// front body (cab section)
	frontCenter := joint.Add(dir(fh).Scale(frontLen * 0.5))
	s.addBox(ground(frontCenter, bodyY+bob), frontLen*0.5, 12, bodyWidth*0.5, fh, yellow)

	// cab with glass, raised on the front body
	cabPos := joint.Add(dir(fh).Scale(frontLen * 0.34))
	s.addBox(ground(cabPos, bodyY+30+bob), 15, 18, bodyWidth*0.38, fh, shade(yellow, 0.85))
	s.addBox(ground(cabPos, bodyY+32+bob), 13, 15, bodyWidth*0.33, fh, rgba(150, 200, 225, 230))

	// warning beacon on the cab roof (blinks)
	beaconCol := rgba(150, 60, 20, 255)
	if math.Sin(l.Beacon*2) > 0 {
		beaconCol = rgba(255, 150, 35, 255)
	}
	s.addCylinder(ground(cabPos.Add(dir(fh).Scale(-8)), bodyY+50+bob), 4, 8, 7, beaconCol)

	// articulation joint hinge
	s.addCylinder(ground(joint, bodyY+bob), 7, 30, 8, rgba(70, 72, 78, 255))

	// ---- lift arms + bucket ----
	g.buildBucket3D(s, l, fh, bob)
}

func (g *Game) buildBucket3D(s *Scene, l *Loader, fh, bob float64) {
	// shoulders on the front body sides
	shL := ground(l.FrontAxle().Add(dir(fh+math.Pi/2).Scale(bodyWidth*0.4)).Add(dir(fh).Scale(4)), liftBaseY+bob)
	shR := ground(l.FrontAxle().Add(dir(fh-math.Pi/2).Scale(bodyWidth*0.4)).Add(dir(fh).Scale(4)), liftBaseY+bob)

	liftY := liftBaseY + l.Lift*liftRangeY + bob
	reach := armReach - l.Lift*8
	pivot2D := l.FrontAxle().Add(dir(fh).Scale(reach))
	pivotL := ground(pivot2D.Add(dir(fh+math.Pi/2).Scale(bodyWidth*0.36)), liftY)
	pivotR := ground(pivot2D.Add(dir(fh-math.Pi/2).Scale(bodyWidth*0.36)), liftY)
	pivotC := ground(pivot2D, liftY)

	armCol := rgba(215, 168, 25, 255)
	s.addOBB(shL.Add(pivotL.Sub(shL).Scale(0.5)), pivotL.Sub(shL), pivotL.Sub(shL).Len()/2, 4, 5, armCol)
	s.addOBB(shR.Add(pivotR.Sub(shR).Scale(0.5)), pivotR.Sub(shR), pivotR.Sub(shR).Len()/2, 4, 5, armCol)

	// tilt hydraulic ram (visual)
	ramBase := ground(l.FrontAxle().Add(dir(fh).Scale(8)), bodyY+18+bob)
	s.addOBB(ramBase.Add(pivotC.Sub(ramBase).Scale(0.5)), pivotC.Sub(ramBase), pivotC.Sub(ramBase).Len()/2, 3, 3, rgba(120, 120, 128, 255))

	// bucket: oriented so tilt curls the opening. tiltPitch < 0 points the
	// cutting edge down (dig/dump), > 0 curls it back (carry).
	tiltPitch := (l.Tilt - 0.45) * 1.25
	fwd3 := Vec3{math.Cos(fh), 0, math.Sin(fh)}
	up3 := Vec3{0, 1, 0}
	lat := fwd3.Cross(up3).Norm()
	// bf is the bucket's forward (floor) axis; bu is its "up" (opening) axis.
	bf := rotAbout(fwd3, lat, -tiltPitch)
	bu := rotAbout(up3, lat, -tiltPitch)
	steel := rgba(72, 74, 80, 255)
	bw := bodyWidth*0.5 + 14
	// scoop body: a box tilted with the bucket
	bodyC := pivotC.Add(bf.Scale(16)).Add(bu.Scale(2))
	s.addOBB(bodyC, bf, 20, 11, bw, steel)
	// cutting-edge lip at the front
	lipC := pivotC.Add(bf.Scale(36)).Add(bu.Scale(-4))
	s.addOBB(lipC, bf, 5, 4, bw, rgba(120, 122, 128, 255))

	// load heaped inside the bucket, visible mostly when curled back
	if l.Load > 2 {
		vis := l.LoadFrac() * (0.3 + l.Tilt*0.7)
		r, gg, b := matColor(l.LoadType)
		sq := math.Sqrt(clamp(vis, 0, 1))
		mc := pivotC.Add(bf.Scale(14)).Add(bu.Scale(9))
		s.addOBB(mc, bf, 17*sq, 4+5*vis, bw*0.9*sq, rgba(r, gg, b, 255))
	}
}

// rotAbout rotates vector v about unit axis k by angle a (Rodrigues).
func rotAbout(v, k Vec3, a float64) Vec3 {
	si, co := math.Sincos(a)
	return v.Scale(co).
		Add(k.Cross(v).Scale(si)).
		Add(k.Scale(k.Dot(v) * (1 - co)))
}

// ---- particles ----

func (g *Game) buildParticles3D(s *Scene) {
	for i := range g.world.Particles {
		p := &g.world.Particles[i]
		life := p.Life / p.Max
		var a float64
		switch p.Kind {
		case pfxSmoke:
			a = clamp(life*90, 0, 90)
		case pfxDust:
			a = clamp(life*95, 0, 95)
		default:
			a = clamp(life*235, 0, 235)
		}
		size := p.Size * (0.7 + life*0.6)
		if p.Kind == pfxSmoke || p.Kind == pfxDust {
			size = p.Size * (1.1 - life*0.4)
		}
		center := Vec3{p.Pos.X, p.Z + size*0.5, p.Pos.Y}
		s.addBillboard(center, size, rgba(p.Color[0], p.Color[1], p.Color[2], uint8(a)))
	}
}
