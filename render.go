package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func rgba(r, g, b, a uint8) color.RGBA { return color.RGBA{r, g, b, a} }

// render builds the whole 3D scene for this frame and draws it, then the HUD.
func (g *Game) render(screen *ebiten.Image) {
	g.cam.setup()
	s := &g.scene
	s.reset()
	s.eye = g.cam.eye

	g.buildGround3D(s)
	g.buildMaterial3D(s)
	g.buildShadows(s)
	g.buildProps3D(s)
	g.buildTrucks3D(s)
	g.buildLoader3D(s, g.loader, true)
	g.buildParticles3D(s)

	s.Flush(screen, g.cam, g.ground, g.skirt, g.atlas, g.art.white)

	g.drawHUD(screen)
}

// ---- ground ----

func (g *Game) buildGround3D(s *Scene) {
	W, H := g.world.Width, g.world.Height
	tileQuad := func(x, y, x1, y1 float64) {
		c := Vec3{(x + x1) / 2, 0, (y + y1) / 2}
		if !g.cam.inView(c, math.Hypot(x1-x, y1-y)*0.6) {
			return
		}
		inside := x >= 0 && y >= 0 && x1 <= W && y1 <= H
		if inside {
			s.addGroundQuad(Vec3{x, 0, y}, Vec3{x1, 0, y}, Vec3{x1, 0, y1}, Vec3{x, 0, y1})
		} else {
			s.addSkirtQuad(Vec3{x, 0, y}, Vec3{x1, 0, y}, Vec3{x1, 0, y1}, Vec3{x, 0, y1})
		}
	}
	// the site itself, finely tessellated so the texture stays steady
	const tile = 110.0
	for y := 0.0; y < H; y += tile {
		for x := 0.0; x < W; x += tile {
			tileQuad(x, y, math.Min(x+tile, W), math.Min(y+tile, H))
		}
	}
	// coarse skirt beyond the site edge, textured with repeating dirt
	const big, reach = 500.0, 4000.0
	for y := -reach; y < H+reach; y += big {
		for x := -reach; x < W+reach; x += big {
			if x >= 0 && y >= 0 && x+big <= W && y+big <= H {
				continue
			}
			if x+big > 0 && y+big > 0 && x < W && y < H {
				// straddles the edge: split into the outside parts only
				for _, r := range outsideParts(x, y, x+big, y+big, W, H) {
					tileQuad(r[0], r[1], r[2], r[3])
				}
				continue
			}
			tileQuad(x, y, x+big, y+big)
		}
	}
}

// outsideParts returns the pieces of rect (x0,y0)-(x1,y1) lying outside the
// world rectangle (0,0)-(W,H).
func outsideParts(x0, y0, x1, y1, W, H float64) [][4]float64 {
	var out [][4]float64
	if y0 < 0 {
		out = append(out, [4]float64{x0, y0, x1, 0})
	}
	if y1 > H {
		out = append(out, [4]float64{x0, H, x1, y1})
	}
	cy0, cy1 := math.Max(y0, 0), math.Min(y1, H)
	if x0 < 0 {
		out = append(out, [4]float64{x0, cy0, 0, cy1})
	}
	if x1 > W {
		out = append(out, [4]float64{W, cy0, x1, cy1})
	}
	return out
}

// ---- material grid as a smooth heightfield ----

// pileHeight maps a cell's material amount to a visual height. The curve
// lifts thin layers a little so heaps get a defined, steep toe (like loose
// material at its angle of repose) and tapers off toward a cap.
func pileHeight(amount float64) float64 {
	if amount < 0.5 {
		return 0
	}
	return 66 * (1 - math.Exp(-amount/55))
}

// buildMaterial3D turns the loose-material grid into a continuous mesh: one
// vertex per cell centre, heights blurred so heaps read as rounded mounds,
// per-vertex normals for smooth shading, and alpha fading at the toe so piles
// blend into the ground instead of ending in a hard edge.
func (g *Game) buildMaterial3D(s *Scene) {
	grid := g.world.Grid
	cs := grid.CellSize
	eye := Vec{g.cam.eye.X, g.cam.eye.Z}
	r := fogEnd
	x0 := int(math.Max(0, (eye.X-r)/cs)) - 1
	y0 := int(math.Max(0, (eye.Y-r)/cs)) - 1
	x1 := int(math.Min(float64(grid.W), (eye.X+r)/cs)) + 1
	y1 := int(math.Min(float64(grid.H), (eye.Y+r)/cs)) + 1
	ww, wh := x1-x0+1, y1-y0+1
	if ww <= 2 || wh <= 2 {
		return
	}

	raw := make([]float64, ww*wh)
	amt := make([]float64, ww*wh)
	typ := make([]int, ww*wh)
	any := false
	for j := 0; j < wh; j++ {
		for i := 0; i < ww; i++ {
			cx, cy := x0+i, y0+j
			if !grid.inBounds(cx, cy) {
				continue
			}
			c := grid.at(cx, cy)
			k := j*ww + i
			amt[k] = c.Amount
			typ[k] = c.Type
			raw[k] = pileHeight(c.Amount)
			if raw[k] > 0 {
				any = true
			}
		}
	}
	if !any {
		return
	}
	at := func(a []float64, i, j int) float64 {
		if i < 0 || j < 0 || i >= ww || j >= wh {
			return 0
		}
		return a[j*ww+i]
	}
	// light 3x3 blur of heights and amount-weighted colour: rounds the heaps
	// without smearing a wide fringe past the material's edge
	h := make([]float64, ww*wh)
	col := make([][3]float64, ww*wh)
	kern := [3]float64{1, 4, 1}
	for j := 0; j < wh; j++ {
		for i := 0; i < ww; i++ {
			var hs, ws float64
			var cr, cg, cb, cw float64
			for dj := -1; dj <= 1; dj++ {
				for di := -1; di <= 1; di++ {
					w := kern[di+1] * kern[dj+1]
					hs += at(raw, i+di, j+dj) * w
					ws += w
					a := at(amt, i+di, j+dj)
					if a > 0.5 {
						t := typ[(j+dj)*ww+(i+di)]
						r, gg, b := matColor(t)
						cw += a * w
						cr += float64(r) * a * w
						cg += float64(gg) * a * w
						cb += float64(b) * a * w
					}
				}
			}
			k := j*ww + i
			h[k] = hs / ws
			if cw > 0 {
				col[k] = [3]float64{cr / cw, cg / cw, cb / cw}
			}
		}
	}

	vert := func(i, j int) (Vec3, Vec3, color.RGBA) {
		hh := at(h, i, j)
		p := Vec3{(float64(x0+i) + 0.5) * cs, hh, (float64(y0+j) + 0.5) * cs}
		// normals from central differences, slopes exaggerated a little so
		// the sun picks out the shape of each heap
		const relief = 1.5
		n := Vec3{-relief * (at(h, i+1, j) - at(h, i-1, j)) / (2 * cs), 1, -relief * (at(h, i, j+1) - at(h, i, j-1)) / (2 * cs)}.Norm()
		c := col[j*ww+i]
		// subtle granular variation + a darker, damp toe where it meets the ground
		v := (0.95 + 0.1*hashNoise(x0+i, y0+j, 41)) / grainMean // the grain texture darkens by grainMean
		v *= 0.74 + 0.26*math.Min(1, hh/10)
		a := clamp(hh/1.5, 0, 1)
		return p, n, color.RGBA{clampByte(c[0] * v), clampByte(c[1] * v), clampByte(c[2] * v), uint8(a * 255)}
	}
	for j := 0; j+1 < wh; j++ {
		for i := 0; i+1 < ww; i++ {
			h00, h10, h01, h11 := at(h, i, j), at(h, i+1, j), at(h, i, j+1), at(h, i+1, j+1)
			if math.Max(math.Max(h00, h10), math.Max(h01, h11)) < 0.6 {
				continue
			}
			center := Vec3{(float64(x0+i) + 1) * cs, (h00 + h11) / 2, (float64(y0+j) + 1) * cs}
			if !g.cam.inView(center, cs*1.2+math.Max(h00, h11)) {
				continue
			}
			p00, n00, c00 := vert(i, j)
			p10, n10, c10 := vert(i+1, j)
			p01, n01, c01 := vert(i, j+1)
			p11, n11, c11 := vert(i+1, j+1)
			// world-space grain mapping: 2 texels per unit, wrapped into the
			// periodic grain region (its extension band covers one cell)
			u0 := float32(grainX0 + math.Mod(p00.X*2, grainPeriod))
			v0 := float32(math.Mod(p00.Z*2, grainPeriod))
			u1, v1 := u0+grainSpan, v0+grainSpan
			s.addTriUV([3]Vec3{p00, p11, p10}, [3]color.RGBA{c00, c11, c10}, [3]Vec3{n00, n11, n10},
				[3][2]float32{{u0, v0}, {u1, v1}, {u1, v0}}, true)
			s.addTriUV([3]Vec3{p00, p01, p11}, [3]color.RGBA{c00, c01, c11}, [3]Vec3{n00, n01, n11},
				[3][2]float32{{u0, v0}, {u0, v1}, {u1, v1}}, true)
		}
	}
}

// ---- shadows (flat dark quads on the ground) ----

func (g *Game) buildShadows(s *Scene) {
	sh := rgba(18, 22, 20, 80)
	off := V(12, 16) // cast offset away from the light
	quad := func(center Vec, w, h, ang float64) {
		cc := center.Add(off)
		if !g.cam.inView(ground(cc, 0), math.Max(w, h)) {
			return
		}
		c := dir(ang)
		p := dir(ang + math.Pi/2)
		a := ground(cc.Add(c.Scale(-w/2)).Add(p.Scale(-h/2)), 0.4)
		b := ground(cc.Add(c.Scale(w/2)).Add(p.Scale(-h/2)), 0.4)
		d := ground(cc.Add(c.Scale(w/2)).Add(p.Scale(h/2)), 0.4)
		e := ground(cc.Add(c.Scale(-w/2)).Add(p.Scale(h/2)), 0.4)
		s.addDecalQuad(a, b, d, e, sh)
	}
	for i := range g.world.Props {
		p := &g.world.Props[i]
		if p.Kind == propSign || (p.Kind == propModel && p.Tag == "tree") {
			continue
		}
		quad(p.Rect.Center, p.Rect.W+8, p.Rect.H+8, p.Rect.Angle)
	}
	for _, t := range g.world.Trucks {
		quad(t.Pos, 118, 74, t.Angle)
	}
	l := g.loader
	quad(l.RearAxle().Sub(dir(l.RearHeading()).Scale(6)), rearLen+34, bodyWidth+16, l.RearHeading())
	quad(l.FrontAxle(), frontLen+18, bodyWidth+16, l.FrontHeading())
	quad(l.BucketPivot().Add(dir(l.FrontHeading()).Scale(16)), 34, bodyWidth+28, l.FrontHeading())
}

// ---- props ----

var (
	colTire  = rgba(34, 34, 37, 255)
	colDark  = rgba(52, 54, 58, 255)
	colSteel = rgba(78, 80, 86, 255)
	colGlass = rgba(86, 124, 152, 255)
	colRim   = rgba(214, 168, 30, 255)
	colChrom = rgba(196, 198, 204, 255)
)

func (g *Game) buildProps3D(s *Scene) {
	for i := range g.world.Props {
		p := &g.world.Props[i]
		rc := p.Rect
		if !g.cam.inView(ground(rc.Center, 30), math.Hypot(rc.W, rc.H)*0.6+40) {
			continue
		}
		c := rgba(p.Color[0], p.Color[1], p.Color[2], 255)
		o := ground(rc.Center, 0)
		fwd := dir(rc.Angle)
		side := dir(rc.Angle + math.Pi/2)
		switch p.Kind {
		case propModel:
			if m, err := loadModel(p.Model); err == nil {
				s.addMesh(m, o, rc.Angle+p.ModelYaw, p.Scale)
			}
		case propBuilding:
			// two-storey site office: chamfered block, parapet, window bands
			s.addPrismCaps(o, rc.Angle, []ring{
				{-rc.W / 2, rc.W / 2, rc.H / 2, 0, 3},
				{-rc.W / 2, rc.W / 2, rc.H / 2, 70, 3},
			}, c, false, false)
			s.addPrism(o, rc.Angle, []ring{
				{-rc.W/2 - 2, rc.W/2 + 2, rc.H/2 + 2, 70, 4},
				{-rc.W/2 - 2, rc.W/2 + 2, rc.H/2 + 2, 76, 4},
			}, shadeC(c, 0.85))
			s.detail(func() {
				for _, y := range []float64{20, 48} {
					s.addPrismCaps(o, rc.Angle, []ring{
						{-rc.W/2 - 0.8, rc.W/2 + 0.8, rc.H/2 + 0.8, y, 3},
						{-rc.W/2 - 0.8, rc.W/2 + 0.8, rc.H/2 + 0.8, y + 12, 3},
					}, colGlass, false, false)
				}
			})
			s.addBox(ground(rc.Center.Add(fwd.Scale(-rc.W*0.2)), 82), 18, 6, 13, rc.Angle, rgba(150, 152, 156, 255))
			s.addBox(ground(rc.Center.Add(fwd.Scale(rc.W*0.22)), 80), 12, 4, 10, rc.Angle, rgba(135, 137, 142, 255))
		case propCabin:
			// shipping-container cabin: body, corrugation ribs, door, window
			s.addPrismCaps(o, rc.Angle, []ring{
				{-rc.W / 2, rc.W / 2, rc.H / 2, 0, 0.5},
				{-rc.W / 2, rc.W / 2, rc.H / 2, 46, 0.5},
			}, c, false, false)
			s.detail(func() {
				ribs := int(rc.W / 14)
				for k := 0; k <= ribs; k++ {
					x := -rc.W/2 + 4 + float64(k)*(rc.W-8)/float64(ribs)
					for _, sg := range []float64{-1, 1} {
						// leave the door and window openings clear of ribs
						if sg > 0 && (math.Abs(x-rc.W*0.28) < 11 || math.Abs(x+rc.W*0.12) < 16) {
							continue
						}
						s.addBox(ground(rc.Center.Add(fwd.Scale(x)).Add(side.Scale(sg*(rc.H/2+0.8))), 23), 1.6, 21, 1.2, rc.Angle, shadeC(c, 0.88))
					}
				}
				s.addBox(ground(rc.Center.Add(side.Scale(rc.H/2+1.5)).Add(fwd.Scale(rc.W*0.28)), 20), 9, 18, 1, rc.Angle, colDark)
				s.addBox(ground(rc.Center.Add(side.Scale(rc.H/2+1.5)).Add(fwd.Scale(-rc.W*0.12)), 28), 14, 7, 1, rc.Angle, colGlass)
				s.addBox(ground(rc.Center, 47), rc.W/2+1, 1.5, rc.H/2+1, rc.Angle, shadeC(c, 1.1))
			})
		case propTank:
			s.addCylinder(ground(rc.Center, 24), rc.W/2, 48, 20, c)
			s.addCone(ground(rc.Center, 52), rc.W/2, rc.W*0.25, 8, 20, shadeC(c, 1.05))
			for _, a := range []float64{0.6, 2.2, 3.8, 5.4} {
				s.addBox(ground(rc.Center.Add(dir(a).Scale(rc.W*0.4)), 3), 4, 3, 4, a, colDark)
			}
		case propBarrier:
			// jersey barrier profile, alternating red/white segments
			n := 3
			seg := rc.W / float64(n)
			for k := 0; k < n; k++ {
				x0 := -rc.W/2 + float64(k)*seg
				cc := c
				if k%2 == 1 {
					cc = rgba(236, 236, 230, 255)
				}
				s.addPrism(o, rc.Angle, []ring{
					{x0 + 0.5, x0 + seg - 0.5, rc.H / 2, 0, 1},
					{x0 + 0.5, x0 + seg - 0.5, rc.H/2 - 1, 5, 1},
					{x0 + 0.5, x0 + seg - 0.5, rc.H / 4, 10, 1},
					{x0 + 0.5, x0 + seg - 0.5, rc.H / 5, 24, 1},
				}, cc)
			}
		case propCone:
			s.addBox(ground(rc.Center, 1.2), rc.W*0.6, 1.2, rc.W*0.6, 0.3, colDark)
			s.addCone(ground(rc.Center, 7), rc.W*0.42, rc.W*0.3, 11, 12, rgba(242, 110, 26, 255))
			s.addCone(ground(rc.Center, 14.5), rc.W*0.3, rc.W*0.22, 4, 12, rgba(240, 240, 236, 255))
			s.addCone(ground(rc.Center, 19), rc.W*0.22, rc.W*0.06, 5, 12, rgba(242, 110, 26, 255))
		case propSign:
			s.addBox(ground(rc.Center, 17), 1.5, 17, 1.5, 0, colSteel)
			s.addOBB2(ground(rc.Center, 36), Vec3{0, 0, 1}, Vec3{1, 1, 0}, 2, rc.W*0.42, rc.W*0.42, rgba(250, 206, 40, 255))
		case propExcavator:
			g.buildExcavator3D(s, p)
		case propParkedLoader:
			parked := Loader{Joint: rc.Center, Heading: rc.Angle, Artic: 0.22, Lift: 0.02, Tilt: 0.5}
			g.buildLoader3D(s, &parked, false)
		}
	}
}

func (g *Game) buildExcavator3D(s *Scene, p *Prop) {
	rc := p.Rect
	c := rgba(p.Color[0], p.Color[1], p.Color[2], 255)
	side := dir(rc.Angle + math.Pi/2)
	// crawler tracks with rounded ends
	for _, sgn := range []float64{-1, 1} {
		tc := rc.Center.Add(side.Scale(sgn * rc.H * 0.34))
		s.addPrism(ground(tc, 0), rc.Angle, []ring{
			{-rc.W/2 + 10, rc.W/2 - 10, rc.H * 0.15, 0, 2},
			{-rc.W / 2, rc.W / 2, rc.H * 0.15, 9, 2},
			{-rc.W/2 + 4, rc.W/2 - 4, rc.H * 0.15, 18, 2},
		}, colDark)
	}
	// slewing upper structure: house with counterweight and cab
	ha := rc.Angle + 0.35
	o := ground(rc.Center, 20)
	s.addPrism(o, ha, []ring{
		{-rc.W * 0.34, rc.W * 0.26, rc.H * 0.38, 0, 6},
		{-rc.W * 0.34, rc.W * 0.26, rc.H * 0.38, 18, 8},
		{-rc.W * 0.3, rc.W * 0.22, rc.H * 0.33, 24, 10},
	}, c)
	cab := o.Add(rotY(Vec3{rc.W * 0.12, 0, -rc.H * 0.2}, ha))
	s.addPrism(cab, ha, []ring{{-12, 12, 11, 24, 2}, {-12, 12, 11, 30, 2}}, shadeC(c, 0.9))
	s.addPrism(cab, ha, []ring{{-11, 11, 10, 30, 2}, {-10, 9, 9, 46, 2}}, colGlass)
	s.addPrism(cab, ha, []ring{{-12, 11, 11, 46, 3}, {-12, 11, 11, 49, 3}}, shadeC(c, 0.9))
	// boom -> dipper -> bucket
	f := rotY(Vec3{1, 0, 0}, ha)
	base := o.Add(f.Scale(rc.W * 0.24)).Add(Vec3{0, 18, 0}).Add(rotY(Vec3{0, 0, 8}, ha))
	elbow := base.Add(f.Scale(50)).Add(Vec3{0, 30, 0})
	wrist := elbow.Add(f.Scale(38)).Add(Vec3{0, -48, 0})
	beam := func(a, b Vec3, hh, hw float64, col color.RGBA) {
		s.addOBB(a.Add(b.Sub(a).Scale(0.5)), b.Sub(a), b.Sub(a).Len()/2+hh*0.6, hh, hw, col)
	}
	beam(base, elbow, 7, 6, c)
	beam(elbow, wrist, 5, 5, shadeC(c, 0.9))
	beam(base.Add(Vec3{0, -6, 0}).Add(f.Scale(8)), elbow.Sub(f.Scale(18)), 2.5, 2.5, colChrom)
	s.addOBB2(wrist.Add(f.Scale(6)), f, Vec3{0, 1, 0}, 9, 9, 14, colSteel)
}

// ---- trucks ----

func (g *Game) buildTrucks3D(s *Scene) {
	for _, t := range g.world.Trucks {
		if !g.cam.inView(ground(t.Pos, 30), 90) {
			continue
		}
		ang := t.Angle
		fwd := dir(ang)
		side := dir(ang + math.Pi/2)
		at := func(x, z, y float64) Vec3 { return ground(t.Pos.Add(fwd.Scale(x)).Add(side.Scale(z)), y) }
		body := rgba(222, 140, 32, 255)

		for _, x := range []float64{36, -24, -46} {
			for _, z := range []float64{-27, 27} {
				s.addWheel(at(x, z, 15), 15, 12, ang, 0, colTire, colSteel)
			}
		}
		// chassis rails + fuel tank
		s.addBox(at(-4, 0, 19), 52, 5, 18, ang, colDark)
		s.addCylinder(at(14, -22, 22), 6, 12, 10, colSteel)
		// cab: lower body, glass band, roof
		s.addPrismCaps(at(0, 0, 0), ang, []ring{{26, 56, 30, 22, 3}, {26, 56, 30, 38, 3}}, body, true, false)
		s.addPrismCaps(at(0, 0, 0), ang, []ring{{26, 56, 30, 38, 3}, {28, 51, 28, 51, 3}}, colGlass, false, false)
		s.addPrism(at(0, 0, 0), ang, []ring{{27, 52, 29.5, 51, 4}, {28, 50, 29, 55, 6}}, body)
		s.detail(func() {
			s.addBox(at(57.5, 0, 22), 2.5, 6, 30, ang, colDark)
			for _, z := range []float64{-22, 22} {
				s.addBox(at(56.8, z, 30), 0.8, 3, 4, ang, rgba(255, 246, 210, 255))
			}
		})
		s.addCylinder(at(31, 32.5, 50), 2.2, 30, 8, colDark) // exhaust beside the cab
		// dump body: floor, tapered side walls, headboard + canopy, tailgate
		bed := t.bedRect()
		bc := bed.Center.Sub(t.Pos).Rot(-ang)
		bx0, bx1 := bc.X-bed.W/2, bc.X+bed.W/2
		hw := bed.H / 2
		s.addBox(at(bc.X, 0, 27), bed.W/2, 3, hw, ang, colSteel)
		for _, sg := range []float64{-1, 1} {
			s.addPrism(at(0, sg*(hw-1.5), 0), ang, []ring{{bx0, bx1, 1.8, 30, 0}, {bx0, bx1 + 1.5, 2.2, 54, 0}}, body)
		}
		// headboard stays clear of the cab (which starts at x=26), topped by a
		// short cab-protector lip
		// split at cab-roof height so neither half ties with the roof in the sort
		s.addBox(at(bx1+0.5, 0, 41), 1.5, 14, hw, ang, body)
		s.addBox(at(bx1+0.5, 0, 59), 1.5, 4, hw, ang, body)
		s.addBox(at(bx1+2.5, 0, 64.5), 3.5, 1.5, hw, ang, shadeC(body, 0.9))
		s.addBox(at(bx0-1.5, 0, 42), 2, 13, hw, ang, shadeC(body, 0.92))
		s.detail(func() { s.addBox(at(bx0-3, 0, 14), 1, 8, hw+2, ang, colTire) }) // mud flaps
		// contents heaped above the bed floor
		if tot := t.Total(); tot > 1 {
			frac := clamp(tot/t.Capacity, 0, 1)
			dom, best := 0, 0.0
			for ty := 1; ty < numMat; ty++ {
				if t.Contents[ty] > best {
					best, dom = t.Contents[ty], ty
				}
			}
			r, gg, b := matColor(dom)
			mc := rgba(r, gg, b, 255)
			top := 30 + 6 + frac*22
			s.addPrism(at(bc.X, 0, 0), ang, []ring{
				{-bed.W/2 + 2, bed.W/2 - 2, hw - 3, 30, 4},
				{-bed.W/2 + 4, bed.W/2 - 4, hw - 4, 30 + (top-30)*0.7, 8},
				{-bed.W * 0.32, bed.W * 0.28, hw * 0.55, top, 10},
			}, mc)
		}
	}
}

// ---- the loader ----

const wheelR = 18.0

// buildLoader3D builds an articulated wheel loader from the given simulated
// state. The cab and engine sit on the rear frame (as on real loaders); the
// front frame carries the lift arms, linkage and bucket.
func (g *Game) buildLoader3D(s *Scene, l *Loader, live bool) {
	if !g.cam.inView(ground(l.Joint, 40), 160) {
		return
	}
	fh := l.FrontHeading()
	rh := l.RearHeading()
	J := l.Joint
	bob := 0.0
	if live {
		bob = clamp(l.bounce*0.3, -3, 3)
	}
	yellow := rgba(240, 186, 24, 255)
	if !live {
		yellow = rgba(226, 150, 30, 255)
	}
	yellowD := shadeC(yellow, 0.86)

	rearO := ground(J, bob)
	frontO := ground(J, bob)
	rAt := func(x, z, y float64) Vec3 { return rearO.Add(rotY(Vec3{x, y, z}, rh)) }
	fAt := func(x, z, y float64) Vec3 { return frontO.Add(rotY(Vec3{x, y, z}, fh)) }

	// wheels (rims match the body colour)
	spin := l.wheelSpin / 0.08 / wheelR
	lat := bodyWidth/2 + 4.0
	for _, z := range []float64{-lat, lat} {
		s.addWheel(ground(J.Sub(dir(rh).Scale(rearLen)), wheelR).Add(rotY(Vec3{0, 0, z}, rh)), wheelR, 14, rh, spin, colTire, yellow)
		s.addWheel(ground(J.Add(dir(fh).Scale(frontLen)), wheelR).Add(rotY(Vec3{0, 0, z}, fh)), wheelR, 14, fh, spin, colTire, yellow)
	}

	// ---- rear frame ----
	rl := float64(rearLen)
	s.addPrism(rearO, rh, []ring{
		{-rl - 22, -2, 17, 9, 3},
		{-rl - 22, -2, 17, 23, 3},
	}, yellowD)
	// engine hood: tapers and rounds toward the back
	s.addPrism(rearO, rh, []ring{
		{-rl - 20, -27, 16.5, 23, 3},
		{-rl - 20, -27, 16.5, 38, 4},
		{-rl - 16, -28, 13, 45, 6},
	}, yellow)
	// counterweight + grille
	s.addPrism(rearO, rh, []ring{
		{-rl - 27, -rl - 18, 19, 6, 4},
		{-rl - 27, -rl - 18, 19, 26, 5},
		{-rl - 25, -rl - 19, 17, 30, 4},
	}, yellow)
	s.detail(func() {
		s.addBox(rAt(-rl-27.3, 0, 19), 0.8, 6, 12, rh, colDark)
		for k := -2; k <= 2; k++ {
			s.addBox(rAt(-rl-27.6, float64(k)*4.4, 19), 0.5, 5.5, 0.8, rh, colSteel)
		}
	})
	// exhaust stack + air precleaner
	s.addCylinder(rAt(-36, 7, 52), 2.4, 16, 8, colDark)
	s.addCylinder(rAt(-38, -7, 49), 3.5, 8, 10, colSteel)
	// rear fenders
	s.detail(func() {
		for _, z := range []float64{-lat, lat} {
			s.addOBB2(rAt(-rl, z, wheelR*2+3), rotY(Vec3{1, 0, 0}, rh), Vec3{0, 1, 0}, 17, 1.5, 8, colDark)
		}
	})
	// cab: base, glass band, roof with overhang, beacon
	s.addPrismCaps(rearO, rh, []ring{{-26, -3, 15, 23, 2}, {-26, -3, 15, 30, 2}}, yellowD, true, false)
	s.addPrismCaps(rearO, rh, []ring{{-26, -3, 15, 30, 2}, {-25, -6, 14, 54, 2}}, colGlass, false, false)
	s.detail(func() {
		for _, c := range [][2]float64{{-25.5, -14.5}, {-25.5, 14.5}, {-3.8, -14.5}, {-3.8, 14.5}} {
			top := rAt(c[0], c[1], 54)
			if c[0] > -10 {
				top = rAt(c[0]-2.2, c[1], 54)
			}
			bot := rAt(c[0], c[1], 30)
			s.addOBB(bot.Add(top.Sub(bot).Scale(0.5)), top.Sub(bot), top.Sub(bot).Len()/2, 1.3, 1.3, colDark)
		}
	})
	s.addPrism(rearO, rh, []ring{{-27.5, -3.5, 16.5, 54, 3}, {-27, -4, 16, 58, 4}}, yellow)
	beacon := rgba(140, 60, 20, 255)
	if live && math.Sin(l.Beacon*2) > 0 {
		beacon = rgba(255, 150, 35, 255)
		s.addBillboard(rAt(-15, 0, 63), 9, rgba(255, 150, 40, 70))
	}
	s.addCylinder(rAt(-15, 0, 61), 2.6, 6, 8, beacon)

	// ---- articulation hinge ----
	s.addCylinder(ground(J, 12+bob), 5.5, 7, 10, colSteel)
	s.addCylinder(ground(J, 26+bob), 5.5, 7, 10, colSteel)
	s.addCylinder(ground(J, 19+bob), 2.5, 22, 8, colChrom)

	// ---- front frame ----
	fl := float64(frontLen)
	s.addPrism(frontO, fh, []ring{
		{4, fl + 14, 14, 8, 3},
		{4, fl + 14, 14, 22, 3},
		{4, fl + 6, 12, 27, 4},
	}, yellow)
	s.detail(func() {
		for _, z := range []float64{-lat, lat} {
			s.addOBB2(fAt(fl, z, wheelR*2+3), rotY(Vec3{1, 0, 0}, fh), Vec3{0, 1, 0}, 16, 1.5, 8, colDark)
		}
	})
	g.buildArmsAndBucket(s, l, fh, bob, yellow, fAt)
}

// buildArmsAndBucket draws the lift arms, cylinders, Z-bar linkage and a
// U-shaped bucket. Lift raises the bucket pivot; tilt pitches the bucket.
func (g *Game) buildArmsAndBucket(s *Scene, l *Loader, fh, bob float64, yellow color.RGBA, fAt func(x, z, y float64) Vec3) {
	pivot2D := l.BucketPivot()
	reach := pivot2D.Sub(l.Joint).Dot(dir(fh)) // local x of the bucket pivot
	pivotY := 16 + l.Lift*72 + bob
	armZ := bodyWidth * 0.36
	beam := func(a, b Vec3, hh, hw float64, col color.RGBA) {
		d := b.Sub(a)
		s.addOBB(a.Add(d.Scale(0.5)), d, d.Len()/2, hh, hw, col)
	}

	// arm towers on the front frame
	for _, z := range []float64{-armZ, armZ} {
		s.addPrism(fAt(0, z, 0), fh, []ring{{6, 22, 3, 20, 1}, {10, 20, 3, 44, 1}}, shadeC(yellow, 0.9))
	}
	var armMid [2]Vec3
	for k, z := range []float64{-armZ, armZ} {
		sh := fAt(15, z, 42+bob*0.5)
		pv := fAt(reach, z, pivotY)
		// a slight dog-leg: arms bow upward over the front tyres
		knee := sh.Add(pv.Sub(sh).Scale(0.45)).Add(Vec3{0, 8, 0})
		beam(sh, knee, 5, 3.5, yellow)
		beam(knee, pv, 5, 3.5, yellow)
		armMid[k] = knee
		// lift cylinder from the low frame to the arm
		base := fAt(28, z*0.7, 12+bob)
		beam(base, base.Add(knee.Sub(base).Scale(0.55)), 2.8, 2.8, colSteel)
		beam(base.Add(knee.Sub(base).Scale(0.5)), knee, 1.6, 1.6, colChrom)
		// pivot pins
		s.addOBB2(sh, rotY(Vec3{0, 0, 1}, fh), Vec3{0, 1, 0}, 4.5, 3, 3, colSteel)
	}
	// cross tube between the arms and the Z-bar tilt linkage
	beam(armMid[0], armMid[1], 3.5, 3.5, shadeC(yellow, 0.92))
	crankTop := armMid[0].Add(armMid[1]).Scale(0.5).Add(Vec3{0, 14, 0})
	tiltBase := fAt(26, 0, 30+bob)
	beam(tiltBase, tiltBase.Add(crankTop.Sub(tiltBase).Scale(0.6)), 3, 3, colSteel)
	beam(tiltBase.Add(crankTop.Sub(tiltBase).Scale(0.55)), crankTop, 1.8, 1.8, colChrom)
	beam(crankTop.Sub(Vec3{0, 18, 0}), crankTop, 3.5, 2.5, shadeC(yellow, 0.9))

	// ---- bucket frame ----
	tiltPitch := lerp(-0.85, 0.75, l.Tilt)
	f0 := rotY(Vec3{1, 0, 0}, fh)
	lat := rotY(Vec3{0, 0, 1}, fh)
	bf := rotAbout(f0, lat, tiltPitch)
	bu := rotAbout(Vec3{0, 1, 0}, lat, tiltPitch)
	P := fAt(reach, 0, pivotY)
	B := func(fwd, up, z float64) Vec3 { return P.Add(bf.Scale(fwd)).Add(bu.Scale(up)).Add(lat.Scale(z)) }
	hw := bodyWidth/2 + 14.0
	steel := rgba(82, 84, 90, 255)
	worn := rgba(132, 134, 140, 255)

	// link from the crank to the top of the bucket
	beam(crankTop, B(-4, 12, 0), 2.5, 2.5, shadeC(yellow, 0.9))

	// side plates (double-sided polygons tracing the bucket profile)
	profile := [][2]float64{{40, -12}, {0, -12}, {-6, -4}, {-6, 14}, {12, 16}}
	for _, z := range []float64{-hw, hw} {
		pts := make([]Vec3, len(profile))
		for i, pr := range profile {
			pts[i] = B(pr[0], pr[1], z)
		}
		for i := 2; i < len(pts); i++ {
			s.addTri(pts[0], pts[i-1], pts[i], steel, true, false)
		}
	}
	plate := func(a, b [2]float64, col color.RGBA) {
		pa, pb := B(a[0], a[1], 0), B(b[0], b[1], 0)
		d := pb.Sub(pa)
		s.addOBB2(pa.Add(d.Scale(0.5)), d, bu.Add(bf.Scale(-0.01)), d.Len()/2+0.8, 1.2, hw, col)
	}
	plate([2]float64{0, -12}, [2]float64{40, -12}, steel) // floor
	plate([2]float64{-6, -4}, [2]float64{0, -12}, steel)  // heel
	plate([2]float64{-6, 14}, [2]float64{-6, -4}, steel)  // back
	plate([2]float64{12, 16}, [2]float64{-6, 14}, steel)  // spill guard
	// cutting edge + teeth
	s.addOBB2(B(40, -12, 0), lat, bu, hw+1, 1.6, 2.4, worn)
	for k := 0; k < 7; k++ {
		z := -hw + 4 + float64(k)*(2*hw-8)/6
		s.addOBB2(B(44.5, -12, z), bf, bu, 4.5, 1.4, 2, worn)
	}
	// hinge ears on the back
	for _, z := range []float64{-armZ, armZ} {
		s.addOBB2(B(-6, 2, z), bf, bu, 4, 8, 2.5, steel)
	}

	// load heaped in the bucket, visible mostly when curled back
	if l.Load > 2 {
		vis := clamp(l.LoadFrac()*(0.3+l.Tilt*0.7), 0, 1)
		r, gg, b := matColor(l.LoadType)
		mc := rgba(r, gg, b, 255)
		sq := math.Sqrt(vis)
		s.addOBB2(B(16, -8+5*vis, 0), bf, bu, 18*sq+2, 4+4*vis, (hw-3)*(0.6+0.4*sq), mc)
		s.addOBB2(B(14, -2+9*vis, 0), bf, bu, 11*sq+1, 2+4*vis, (hw-6)*(0.4+0.4*sq), shadeC(mc, 1.08))
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
