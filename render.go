package main

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func rgba(r, g, b, a uint8) color.RGBA { return color.RGBA{r, g, b, a} }

// ---- world-space primitive helpers (transform via camera) ----

func (g *Game) fillPoly(dst *ebiten.Image, ptsWorld []Vec, c color.RGBA) {
	var p vector.Path
	for i, w := range ptsWorld {
		s := g.cam.WorldToScreen(w)
		if i == 0 {
			p.MoveTo(float32(s.X), float32(s.Y))
		} else {
			p.LineTo(float32(s.X), float32(s.Y))
		}
	}
	p.Close()
	vs, is := p.AppendVerticesAndIndicesForFilling(nil, nil)
	cr, cg, cb, ca := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, float32(c.A)/255
	for i := range vs {
		vs[i].SrcX, vs[i].SrcY = 0.5, 0.5
		vs[i].ColorR, vs[i].ColorG, vs[i].ColorB, vs[i].ColorA = cr*ca, cg*ca, cb*ca, ca
	}
	op := &ebiten.DrawTrianglesOptions{AntiAlias: true}
	dst.DrawTriangles(vs, is, g.art.white, op)
}

// rectPoly returns the 4 world corners of a rotated rectangle.
func rectPoly(center Vec, w, h, ang float64) []Vec {
	hx, hy := w/2, h/2
	return []Vec{
		center.Add(V(-hx, -hy).Rot(ang)),
		center.Add(V(hx, -hy).Rot(ang)),
		center.Add(V(hx, hy).Rot(ang)),
		center.Add(V(-hx, hy).Rot(ang)),
	}
}

func (g *Game) fillRect(dst *ebiten.Image, center Vec, w, h, ang float64, c color.RGBA) {
	g.fillPoly(dst, rectPoly(center, w, h, ang), c)
}

func (g *Game) fillCircle(dst *ebiten.Image, center Vec, r float64, c color.RGBA) {
	s := g.cam.WorldToScreen(center)
	vector.DrawFilledCircle(dst, float32(s.X), float32(s.Y), float32(r*g.cam.Zoom), c, true)
}

func (g *Game) strokeWorld(dst *ebiten.Image, a, b Vec, w float64, c color.RGBA) {
	sa := g.cam.WorldToScreen(a)
	sb := g.cam.WorldToScreen(b)
	vector.StrokeLine(dst, float32(sa.X), float32(sa.Y), float32(sb.X), float32(sb.Y), float32(w*g.cam.Zoom), c, true)
}

// ---- top-level render ----

func (g *Game) render(screen *ebiten.Image) {
	screen.Fill(rgba(60, 66, 58, 255))

	// Ground image with camera transform.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(g.cam.Zoom, g.cam.Zoom)
	topLeft := g.cam.WorldToScreen(V(0, 0))
	op.GeoM.Translate(topLeft.X, topLeft.Y)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(g.ground, op)

	g.drawMaterial(screen)
	g.drawShadows(screen)
	g.drawProps(screen)
	g.drawTrucks(screen)
	g.drawLoader(screen)
	g.drawParticles(screen)
	g.drawHUD(screen)
}

// ---- material grid ----

func (g *Game) drawMaterial(screen *ebiten.Image) {
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
			if c.Amount < 1.5 {
				continue
			}
			px := (float64(cx) + 0.5) * cs
			py := (float64(cy) + 0.5) * cs
			h := clamp(c.Amount/55, 0, 1.4)
			r, gg, b := matColor(c.Type)
			// shaded base (darker, sits on ground)
			base := rgba(uint8(float64(r)*0.62), uint8(float64(gg)*0.62), uint8(float64(b)*0.62), 255)
			rad := cs*0.62 + h*cs*0.5
			g.fillCircle(screen, V(px, py+h*2), rad, base)
			// lit top offset toward the light (up-left), size by height
			top := rgba(clampByte(float64(r)*1.08+h*18), clampByte(float64(gg)*1.08+h*18), clampByte(float64(b)*1.08+h*16), 255)
			seed := grid.scatterSeed(cx, cy)
			off := V(-2-h*3, -2-h*3+(seed-0.5)*2)
			g.fillCircle(screen, V(px+off.X, py+off.Y), rad*0.66, top)
		}
	}
}

// ---- shadows ----

func (g *Game) drawShadows(screen *ebiten.Image) {
	sh := rgba(20, 24, 20, 90)
	for i := range g.world.Props {
		p := &g.world.Props[i]
		if p.Kind == propCone || p.Kind == propSign {
			g.fillCircle(screen, p.Rect.Center.Add(V(6, 8)), math.Max(p.Rect.W, p.Rect.H)*0.6, sh)
			continue
		}
		g.fillRect(screen, p.Rect.Center.Add(V(7, 9)), p.Rect.W+8, p.Rect.H+8, p.Rect.Angle, sh)
	}
	for _, t := range g.world.Trucks {
		g.fillRect(screen, t.Pos.Add(V(8, 10)), 80, 120, t.Angle, sh)
	}
}

// ---- props ----

func (g *Game) drawProps(screen *ebiten.Image) {
	props := make([]*Prop, len(g.world.Props))
	for i := range g.world.Props {
		props[i] = &g.world.Props[i]
	}
	sort.Slice(props, func(a, b int) bool { return props[a].Rect.Center.Y < props[b].Rect.Center.Y })
	for _, p := range props {
		g.drawProp(screen, p)
	}
}

func (g *Game) drawProp(screen *ebiten.Image, p *Prop) {
	c := rgba(p.Color[0], p.Color[1], p.Color[2], 255)
	rc := p.Rect
	switch p.Kind {
	case propBuilding, propCabin:
		fwd := dir(rc.Angle)
		side := dir(rc.Angle + math.Pi/2)
		// dark parapet / outline, then the roof surface inset within it
		g.fillRect(screen, rc.Center, rc.W, rc.H, rc.Angle, rgba(c.R*2/5, c.G*2/5, c.B*2/5, 255))
		g.fillRect(screen, rc.Center, rc.W-8, rc.H-8, rc.Angle, c)
		// lit edge toward the light (up-left) for a sense of a raised roof
		g.fillRect(screen, rc.Center.Add(side.Scale(-rc.H/2+7)), rc.W-14, 6, rc.Angle, rgba(clampByte(float64(c.R)*1.18), clampByte(float64(c.G)*1.18), clampByte(float64(c.B)*1.18), 255))
		if p.Kind == propCabin {
			// shipping-container style seams + a door
			for i := -1; i <= 1; i++ {
				g.strokeWorld(screen, rc.Center.Add(fwd.Scale(float64(i)*rc.W*0.26)).Add(side.Scale(-rc.H/2+5)),
					rc.Center.Add(fwd.Scale(float64(i)*rc.W*0.26)).Add(side.Scale(rc.H/2-5)), 2, rgba(c.R*3/5, c.G*3/5, c.B*3/5, 255))
			}
			g.fillRect(screen, rc.Center.Add(fwd.Scale(rc.W*0.32)), 10, rc.H*0.5, rc.Angle, rgba(70, 72, 78, 255))
		} else {
			// flat roof with HVAC units + seams
			g.fillRect(screen, rc.Center.Add(fwd.Scale(-rc.W*0.18)).Add(side.Scale(rc.H*0.12)), 34, 24, rc.Angle, rgba(150, 152, 156, 255))
			g.fillRect(screen, rc.Center.Add(fwd.Scale(rc.W*0.22)).Add(side.Scale(-rc.H*0.14)), 24, 20, rc.Angle, rgba(135, 137, 142, 255))
			g.strokeWorld(screen, rc.Center.Add(side.Scale(-rc.H*0.02)).Add(fwd.Scale(-rc.W/2+6)), rc.Center.Add(side.Scale(-rc.H*0.02)).Add(fwd.Scale(rc.W/2-6)), 2, rgba(c.R*4/5, c.G*4/5, c.B*4/5, 255))
		}
	case propTank:
		g.fillCircle(screen, rc.Center, rc.W/2, rgba(c.R/2, c.G/2, c.B/2, 255))
		g.fillCircle(screen, rc.Center.Add(V(-3, -3)), rc.W/2-5, c)
	case propBarrier:
		// jersey barrier: dark base + colored body + reflective stripes
		g.fillRect(screen, rc.Center, rc.W, rc.H, rc.Angle, rgba(60, 60, 64, 255))
		g.fillRect(screen, rc.Center.Add(V(-2, -3)), rc.W-6, rc.H-8, rc.Angle, c)
		for s := -1; s <= 1; s++ {
			g.fillRect(screen, rc.Center.Add(V(float64(s)*rc.W*0.3, -3)), 10, rc.H-8, rc.Angle, rgba(240, 240, 235, 220))
		}
	case propCone:
		g.fillCircle(screen, rc.Center, rc.W*0.7, rgba(200, 90, 20, 255))
		g.fillCircle(screen, rc.Center, rc.W*0.45, rgba(245, 140, 40, 255))
		g.fillCircle(screen, rc.Center, rc.W*0.22, rgba(255, 255, 255, 230))
	case propSign:
		g.fillRect(screen, rc.Center, rc.W, rc.H, 0.78, rgba(250, 210, 40, 255))
		g.fillRect(screen, rc.Center, rc.W*0.7, rc.H*0.7, 0.78, rgba(30, 30, 30, 255))
	case propExcavator:
		g.drawExcavator(screen, p)
	case propParkedLoader:
		g.fillRect(screen, rc.Center, rc.W, rc.H*0.7, rc.Angle, rgba(c.R/2, c.G/2, c.B/2, 255))
		g.fillRect(screen, rc.Center.Add(V(-2, -3)), rc.W-10, rc.H*0.7-6, rc.Angle, c)
		// a little bucket out front
		fc := rc.Center.Add(dir(rc.Angle).Scale(rc.W/2 + 10))
		g.fillRect(screen, fc, 16, rc.H*0.8, rc.Angle, rgba(50, 50, 52, 255))
		// wheels
		for _, s := range []float64{-1, 1} {
			for _, f := range []float64{-1, 1} {
				wc := rc.Center.Add(dir(rc.Angle).Scale(f * rc.W * 0.3)).Add(dir(rc.Angle + math.Pi/2).Scale(s * rc.H * 0.45))
				g.fillCircle(screen, wc, 10, rgba(30, 30, 32, 255))
			}
		}
	}
}

func (g *Game) drawExcavator(screen *ebiten.Image, p *Prop) {
	rc := p.Rect
	c := rgba(p.Color[0], p.Color[1], p.Color[2], 255)
	cDark := rgba(c.R*2/3, c.G*2/3, c.B*2/3, 255)
	// two crawler tracks
	side := dir(rc.Angle + math.Pi/2)
	for _, s := range []float64{-1, 1} {
		tc := rc.Center.Add(side.Scale(s * rc.H * 0.34))
		g.fillRect(screen, tc, rc.W, rc.H*0.3, rc.Angle, rgba(38, 38, 42, 255))
		for i := -2; i <= 2; i++ {
			g.strokeWorld(screen, tc.Add(dir(rc.Angle).Scale(float64(i)*rc.W*0.18)).Add(side.Scale(s*rc.H*0.15)),
				tc.Add(dir(rc.Angle).Scale(float64(i)*rc.W*0.18)).Sub(side.Scale(s*rc.H*0.15)), 2, rgba(60, 60, 64, 255))
		}
	}
	// slewing house (upper body) rotated relative to the tracks
	ha := rc.Angle + 0.35
	house := rc.Center.Sub(dir(ha).Scale(10))
	g.fillRect(screen, house, rc.W*0.62, rc.H*0.72, ha, cDark)
	g.fillRect(screen, house, rc.W*0.58-6, rc.H*0.72-6, ha, c)
	// counterweight at the back + cab glass at the front
	g.fillRect(screen, house.Sub(dir(ha).Scale(rc.W*0.3)), 12, rc.H*0.6, ha, rgba(50, 50, 54, 255))
	g.fillRect(screen, house.Add(dir(ha).Scale(rc.W*0.18)).Add(side.Scale(rc.H*0.18)), 16, 16, ha, rgba(150, 200, 225, 230))
	// boom -> dipper -> bucket reaching out front of the house
	base := house.Add(dir(ha).Scale(rc.W * 0.28))
	elbow := base.Add(dir(ha + 0.15).Scale(64))
	wrist := elbow.Add(dir(ha - 0.55).Scale(50))
	g.strokeWorld(screen, base, elbow, 13, c)
	g.strokeWorld(screen, elbow, wrist, 9, cDark)
	// bucket at the wrist
	ba := ha - 0.55
	g.fillPoly(screen, []Vec{
		wrist,
		wrist.Add(dir(ba + 1.4).Scale(16)),
		wrist.Add(dir(ba + 1.4).Scale(16)).Add(dir(ba + 2.3).Scale(18)),
		wrist.Add(dir(ba + 2.6).Scale(20)),
	}, rgba(48, 48, 52, 255))
}

// ---- trucks ----

func (g *Game) drawTrucks(screen *ebiten.Image) {
	for _, t := range g.world.Trucks {
		ang := t.Angle
		fwd := dir(ang)
		side := dir(ang + math.Pi/2)
		// wheels first (under the body)
		for _, f := range []float64{-1, -0.35, 0.28} {
			for _, s := range []float64{-1, 1} {
				wc := t.Pos.Add(fwd.Scale(f * 40)).Add(side.Scale(s * 34))
				g.fillRect(screen, wc, 22, 13, ang, rgba(26, 26, 28, 255))
				g.fillCircle(screen, wc, 3, rgba(120, 120, 124, 255))
			}
		}
		// chassis frame
		g.fillRect(screen, t.Pos, 66, 112, ang, rgba(58, 62, 70, 255))
		// cab at front with a safety-stripe bumper
		cab := t.Pos.Add(fwd.Scale(44))
		g.fillRect(screen, cab, 64, 36, ang, rgba(210, 160, 40, 255))
		g.fillRect(screen, cab.Add(fwd.Scale(2)), 52, 18, ang, rgba(150, 200, 230, 235)) // windscreen
		g.fillRect(screen, cab.Add(fwd.Scale(18)), 60, 6, ang, rgba(40, 42, 46, 255))    // bumper
		// dump bed walls + darker interior
		bed := t.bedRect()
		g.fillRect(screen, bed.Center, bed.W, bed.H, ang, rgba(70, 74, 82, 255))
		g.fillRect(screen, bed.Center, bed.W-9, bed.H-9, ang, rgba(40, 42, 48, 255))
		// contents: rounded heap coloured by dominant material, sized by fill
		if tot := t.Total(); tot > 1 {
			frac := clamp(tot/t.Capacity, 0, 1)
			dom, best := 0, 0.0
			for ty := 1; ty < numMat; ty++ {
				if t.Contents[ty] > best {
					best, dom = t.Contents[ty], ty
				}
			}
			r, gg, b := matColor(dom)
			hw := (bed.W/2 - 6) * math.Sqrt(frac)
			hh := (bed.H/2 - 6) * math.Sqrt(frac)
			g.fillRect(screen, bed.Center, hw*2, hh*2, ang, rgba(uint8(float64(r)*0.8), uint8(float64(gg)*0.8), uint8(float64(b)*0.8), 255))
			g.fillRect(screen, bed.Center.Add(side.Scale(-hh*0.3)), hw*1.5, hh*1.1, ang, rgba(r, gg, b, 255))
		}
		// mud flaps at the rear
		rear := t.Pos.Add(fwd.Scale(-56))
		g.fillRect(screen, rear, 60, 8, ang, rgba(30, 30, 32, 255))
	}
}

// ---- the loader ----

func (g *Game) drawLoader(screen *ebiten.Image) {
	l := g.loader
	fh := l.FrontHeading()
	rh := l.RearHeading()
	front := l.FrontAxle()
	rear := l.RearAxle()
	joint := l.Joint

	yellow := rgba(240, 190, 30, 255)
	yellowD := rgba(190, 148, 20, 255)
	darkGrey := rgba(44, 46, 50, 255)

	// ground shadow for the whole machine
	shadow := rgba(16, 20, 16, 95)
	g.fillRect(screen, rear.Add(V(6, 8)), bodyWidth+10, rearLen+16, rh, shadow)
	g.fillRect(screen, front.Add(V(6, 8)), bodyWidth+10, frontLen+16, fh, shadow)

	// wheels (big) — rear pair then front pair
	g.drawWheel(screen, rear, rh, 1)
	g.drawWheel(screen, rear, rh, -1)
	g.drawWheel(screen, front, fh, 1)
	g.drawWheel(screen, front, fh, -1)

	// REAR body (engine compartment)
	rearCenter := joint.Sub(dir(rh).Scale(rearLen * 0.5))
	g.fillRect(screen, rearCenter, bodyWidth, rearLen, rh, yellowD)
	// engine grille lines at the very back
	backEnd := joint.Sub(dir(rh).Scale(rearLen))
	for i := -2; i <= 2; i++ {
		off := dir(rh + math.Pi/2).Scale(float64(i) * 7)
		g.strokeWorld(screen, backEnd.Add(off).Add(dir(rh).Scale(2)), backEnd.Add(off).Add(dir(rh).Scale(14)), 2.5, darkGrey)
	}
	// exhaust stack
	stack := backEnd.Add(dir(rh).Scale(16)).Add(dir(rh + math.Pi/2).Scale(12))
	g.fillCircle(screen, stack, 4, darkGrey)

	// FRONT body (cab section)
	frontCenter := joint.Add(dir(fh).Scale(frontLen * 0.5))
	g.fillRect(screen, frontCenter, bodyWidth, frontLen, fh, yellow)
	// cab + glass
	cabPos := joint.Add(dir(fh).Scale(frontLen * 0.28))
	g.fillRect(screen, cabPos, bodyWidth-8, 26, fh, rgba(60, 64, 72, 255))
	g.fillRect(screen, cabPos, bodyWidth-16, 18, fh, rgba(150, 200, 225, 235))

	// ARTICULATION joint — hinge plate bridging the halves + pivot pin
	g.fillRect(screen, joint, bodyWidth-6, 16, (fh+rh)/2, darkGrey)
	g.fillCircle(screen, joint, 6, rgba(210, 210, 215, 255))
	g.fillCircle(screen, joint, 3, rgba(90, 90, 95, 255))

	// warning beacon on cab roof (blinks)
	beaconOn := math.Sin(l.Beacon*2) > 0
	bc := rgba(140, 60, 20, 255)
	if beaconOn {
		bc = rgba(255, 140, 30, 255)
		// glow
		g.fillCircle(screen, cabPos.Add(dir(fh).Scale(-6)), 10, rgba(255, 150, 40, 70))
	}
	g.fillCircle(screen, cabPos.Add(dir(fh).Scale(-6)), 4, bc)

	// ---- loader arms + bucket (with fake lift elevation) ----
	g.drawBucket(screen, l, fh)
}

func (g *Game) drawWheel(screen *ebiten.Image, axle Vec, heading, side float64) {
	wc := axle.Add(dir(heading + math.Pi/2).Scale(side * (bodyWidth/2 + 4)))
	// tire
	g.fillRect(screen, wc, 30, 18, heading, rgba(28, 28, 30, 255))
	// tread hint
	g.fillRect(screen, wc, 30, 18, heading, rgba(40, 40, 44, 90))
	// hub
	g.fillCircle(screen, wc, 5, rgba(150, 150, 155, 255))
}

// drawBucket draws the lift arms and the animated bucket. Lift is faked in
// top-down by elevating the assembly toward the viewer with a growing gap
// to its ground shadow; tilt reshapes the scoop and reveals its load.
func (g *Game) drawBucket(screen *ebiten.Image, l *Loader, fh float64) {
	elev := l.Lift * 20 // screen-space lift offset (toward viewer / up)
	up := V(0, -elev)   // elevate on screen (drawn after transform via world? use world offset)

	// shoulders at the front corners of the front body
	shoulderL := l.FrontAxle().Add(dir(fh + math.Pi/2).Scale(bodyWidth * 0.42)).Add(dir(fh).Scale(6))
	shoulderR := l.FrontAxle().Add(dir(fh - math.Pi/2).Scale(bodyWidth * 0.42)).Add(dir(fh).Scale(6))

	pivot := l.BucketPivot()
	// bucket ground-shadow (stays near ground projection, no elevation)
	shcol := rgba(16, 20, 16, 80)
	g.bucketShape(screen, pivot, fh, l.Tilt, shcol, shcol, 0, V(6, 8))

	// elevate arms+bucket on screen by adjusting world target with camera-up.
	// Convert screen-up offset into world by dividing by zoom (pure vertical).
	worldUp := up.Scale(1 / g.cam.Zoom)
	pivotE := pivot.Add(worldUp)
	shLE := shoulderL.Add(worldUp)
	shRE := shoulderR.Add(worldUp)

	// arms
	armCol := rgba(210, 165, 25, 255)
	bucketPivotL := pivotE.Add(dir(fh + math.Pi/2).Scale(bodyWidth * 0.38))
	bucketPivotR := pivotE.Add(dir(fh - math.Pi/2).Scale(bodyWidth * 0.38))
	g.strokeWorld(screen, shLE, bucketPivotL, 8, armCol)
	g.strokeWorld(screen, shRE, bucketPivotR, 8, armCol)
	// tilt hydraulic ram (visual): from mid front body to bucket back
	ramBase := l.FrontAxle().Add(dir(fh).Scale(10)).Add(worldUp)
	g.strokeWorld(screen, ramBase, pivotE, 4, rgba(120, 120, 128, 255))

	// bucket body (lit) + load
	steel := rgba(60, 62, 66, 255)
	lip := rgba(90, 92, 98, 255)
	g.bucketShape(screen, pivotE, fh, l.Tilt, steel, lip, l.LoadFrac(), V(0, 0))
}

// bucketShape draws the scoop as a filled quad that reshapes with tilt and,
// optionally, a load mound inside. off is a screen-ish world offset for
// drawing the shadow copy.
func (g *Game) bucketShape(screen *ebiten.Image, pivot Vec, fh, tilt float64, body, edge color.RGBA, loadFrac float64, off Vec) {
	w := bodyWidth*0.5 + 14
	// forward reach of the cutting edge grows as the bucket tilts forward
	reach := 14 + (1-tilt)*18
	back := 4 + tilt*6
	left := dir(fh + math.Pi/2)
	fwd := dir(fh)
	pivot = pivot.Add(off)
	// quad: back-left, back-right, edge-right, edge-left
	bl := pivot.Add(left.Scale(w)).Sub(fwd.Scale(back))
	br := pivot.Sub(left.Scale(w)).Sub(fwd.Scale(back))
	er := pivot.Sub(left.Scale(w * 0.92)).Add(fwd.Scale(reach))
	el := pivot.Add(left.Scale(w * 0.92)).Add(fwd.Scale(reach))
	g.fillPoly(screen, []Vec{bl, br, er, el}, body)

	// load mound inside the bucket, visible mostly when curled (tilt high)
	if loadFrac > 0.02 {
		vis := loadFrac * (0.3 + tilt*0.7)
		r, gg, b := matColor(g.loader.LoadType)
		mc := rgba(r, gg, b, 255)
		iw := w * 0.8 * math.Sqrt(clamp(vis, 0, 1))
		ic := pivot.Add(fwd.Scale(reach * 0.35))
		g.fillPoly(screen, []Vec{
			ic.Add(left.Scale(iw)).Sub(fwd.Scale(2)),
			ic.Sub(left.Scale(iw)).Sub(fwd.Scale(2)),
			ic.Sub(left.Scale(iw * 0.9)).Add(fwd.Scale(8 * vis)),
			ic.Add(left.Scale(iw * 0.9)).Add(fwd.Scale(8 * vis)),
		}, mc)
	}
	// cutting edge bar
	g.strokeWorld(screen, el, er, 5, edge)
}

// ---- particles ----

func (g *Game) drawParticles(screen *ebiten.Image) {
	for i := range g.world.Particles {
		p := &g.world.Particles[i]
		life := p.Life / p.Max
		// draw a ground shadow for elevated particles
		var a uint8
		var col color.RGBA
		switch p.Kind {
		case pfxSmoke:
			a = uint8(clamp(life*90, 0, 90))
			col = rgba(p.Color[0], p.Color[1], p.Color[2], a)
		case pfxDust:
			a = uint8(clamp(life*95, 0, 95))
			col = rgba(p.Color[0], p.Color[1], p.Color[2], a)
		default:
			a = uint8(clamp(life*255, 0, 255))
			col = rgba(p.Color[0], p.Color[1], p.Color[2], a)
		}
		// fake elevation: shift up by Z and grow slightly
		screenPos := p.Pos.Add(V(0, -p.Z))
		size := p.Size * (0.6 + life*0.6)
		if p.Kind == pfxSmoke || p.Kind == pfxDust {
			size = p.Size * (1.15 - life*0.4)
		}
		g.fillCircle(screen, screenPos, size, col)
	}
}
