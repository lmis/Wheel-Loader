package main

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---- 3D vector math ----
//
// The simulation lives on the ground plane in (X, Y) world coordinates. For
// rendering we lift that into 3D with Y as the up axis, mapping a ground
// point (x, y) with height h to the 3D point (x, h, y).

type Vec3 struct{ X, Y, Z float64 }

func V3(x, y, z float64) Vec3 { return Vec3{x, y, z} }

// ground maps a 2D sim point at height h to a 3D world point.
func ground(p Vec, h float64) Vec3 { return Vec3{p.X, h, p.Y} }

func (a Vec3) Add(b Vec3) Vec3      { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3      { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Dot(b Vec3) float64   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Len() float64         { return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z) }

func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}

func (a Vec3) Norm() Vec3 {
	l := a.Len()
	if l < 1e-9 {
		return Vec3{}
	}
	return Vec3{a.X / l, a.Y / l, a.Z / l}
}

// lightDir is the directional "sun" used for flat shading (points toward the
// surface from the light).
var lightDir = Vec3{-0.45, 1.0, -0.35}.Norm()

const ambient = 0.52

// ---- scene primitives ----

type tri struct {
	a, b, c Vec3
	col     color.RGBA
	shade   bool // apply directional lighting
	cull    bool // back-face cull (closed solids only)
}

type gquad struct {
	p    [4]Vec3 // world positions (on ground plane)
	tint float64 // brightness multiplier
}

type billboard struct {
	center Vec3
	size   float64
	col    color.RGBA
}

// Scene accumulates geometry for one frame, then flushes it with depth
// sorting in two passes: textured ground, then shaded/translucent objects.
type Scene struct {
	tris   []tri
	ground []gquad
	bills  []billboard
}

func (s *Scene) reset() {
	s.tris = s.tris[:0]
	s.ground = s.ground[:0]
	s.bills = s.bills[:0]
}

func (s *Scene) addTri(a, b, c Vec3, col color.RGBA, shade, cull bool) {
	s.tris = append(s.tris, tri{a, b, c, col, shade, cull})
}

// addQuad adds a flat quad (two triangles) with a shared color.
func (s *Scene) addQuad(a, b, c, d Vec3, col color.RGBA, shade, cull bool) {
	s.addTri(a, b, c, col, shade, cull)
	s.addTri(a, c, d, col, shade, cull)
}

// addBox adds an axis-aligned-in-local box centred at c, with half-extents
// (hx,hy,hz), rotated by yaw about the vertical axis. Outward winding so
// back-face culling and lighting work.
func (s *Scene) addBox(c Vec3, hx, hy, hz, yaw float64, col color.RGBA) {
	si, co := math.Sincos(yaw)
	// local -> world for the horizontal axes
	rx := func(x, z float64) (float64, float64) { return x*co - z*si, x*si + z*co }
	corner := func(sx, sy, sz float64) Vec3 {
		lx, lz := rx(sx*hx, sz*hz)
		return Vec3{c.X + lx, c.Y + sy*hy, c.Z + lz}
	}
	// 8 corners
	ppp := corner(1, 1, 1)
	ppn := corner(1, 1, -1)
	pnp := corner(1, -1, 1)
	pnn := corner(1, -1, -1)
	npp := corner(-1, 1, 1)
	npn := corner(-1, 1, -1)
	nnp := corner(-1, -1, 1)
	nnn := corner(-1, -1, -1)
	top := shade(col, 1.0)
	side := shade(col, 0.82)
	s.addQuad(npp, ppp, ppn, npn, top, true, true)             // +Y top
	s.addQuad(nnn, pnn, pnp, nnp, shade(col, 0.5), true, true) // -Y bottom
	s.addQuad(pnp, ppp, npp, nnp, side, true, true)            // +Z
	s.addQuad(nnn, npn, ppn, pnn, side, true, true)            // -Z
	s.addQuad(pnn, ppn, ppp, pnp, side, true, true)            // +X
	s.addQuad(nnp, npp, npn, nnn, side, true, true)            // -X
}

// addOBB adds an oriented box whose local forward axis is `fwd` (any 3D
// direction, so it can pitch). hl/hh/hw are half length/height/width.
func (s *Scene) addOBB(center, fwd Vec3, hl, hh, hw float64, col color.RGBA) {
	f := fwd.Norm()
	up := Vec3{0, 1, 0}
	if math.Abs(f.Dot(up)) > 0.95 {
		up = Vec3{1, 0, 0}
	}
	right := f.Cross(up).Norm()
	up = right.Cross(f).Norm()
	c := func(sl, sh, sw float64) Vec3 {
		return center.Add(f.Scale(sl * hl)).Add(up.Scale(sh * hh)).Add(right.Scale(sw * hw))
	}
	ppp := c(1, 1, 1)
	ppn := c(1, 1, -1)
	pnp := c(1, -1, 1)
	pnn := c(1, -1, -1)
	npp := c(-1, 1, 1)
	npn := c(-1, 1, -1)
	nnp := c(-1, -1, 1)
	nnn := c(-1, -1, -1)
	top := shade(col, 1.0)
	sd := shade(col, 0.82)
	dk := shade(col, 0.58)
	s.addQuad(npp, ppp, ppn, npn, top, true, true)
	s.addQuad(nnn, pnn, pnp, nnp, dk, true, true)
	s.addQuad(pnp, ppp, npp, nnp, sd, true, true)
	s.addQuad(nnn, npn, ppn, pnn, sd, true, true)
	s.addQuad(pnn, ppn, ppp, pnp, sd, true, true)
	s.addQuad(nnp, npp, npn, nnn, sd, true, true)
}

// addCylinder adds a vertical low-poly cylinder (sides + caps).
func (s *Scene) addCylinder(c Vec3, radius, h float64, sides int, col color.RGBA) {
	if sides < 5 {
		sides = 8
	}
	top := c.Y + h/2
	bot := c.Y - h/2
	prev := 0.0
	for i := 1; i <= sides; i++ {
		a0 := prev
		a1 := 2 * math.Pi * float64(i) / float64(sides)
		prev = a1
		x0, z0 := math.Cos(a0)*radius, math.Sin(a0)*radius
		x1, z1 := math.Cos(a1)*radius, math.Sin(a1)*radius
		p0t := Vec3{c.X + x0, top, c.Z + z0}
		p1t := Vec3{c.X + x1, top, c.Z + z1}
		p0b := Vec3{c.X + x0, bot, c.Z + z0}
		p1b := Vec3{c.X + x1, bot, c.Z + z1}
		s.addQuad(p0b, p1b, p1t, p0t, shade(col, 0.8), true, true) // side
		s.addTri(Vec3{c.X, top, c.Z}, p0t, p1t, shade(col, 1.0), true, true)
	}
}

// addWheel adds a wheel as a cylinder lying on its side (axle along `axis`).
func (s *Scene) addWheel(center Vec3, radius, width, yaw float64, col color.RGBA) {
	// axle is horizontal, perpendicular to facing. Build a short cylinder
	// around the axle by sampling a disc in the vertical plane.
	sides := 10
	si, co := math.Sincos(yaw)
	axle := Vec3{-si, 0, co} // perpendicular to facing (yaw)
	// disc basis: up and forward(=facing)
	fwd := Vec3{co, 0, si}
	up := Vec3{0, 1, 0}
	off := axle.Scale(width / 2)
	cin := center.Sub(off)
	cout := center.Add(off)
	prev := 0.0
	for i := 1; i <= sides; i++ {
		a0, a1 := prev, 2*math.Pi*float64(i)/float64(sides)
		prev = a1
		d0 := up.Scale(math.Sin(a0) * radius).Add(fwd.Scale(math.Cos(a0) * radius))
		d1 := up.Scale(math.Sin(a1) * radius).Add(fwd.Scale(math.Cos(a1) * radius))
		o0 := cout.Add(d0)
		o1 := cout.Add(d1)
		i0 := cin.Add(d0)
		i1 := cin.Add(d1)
		tread := shade(col, 0.7+0.25*math.Abs(math.Cos(a0)))
		s.addQuad(i0, o0, o1, i1, tread, true, true) // tread
		s.addTri(cout, o0, o1, shade(col, 0.9), true, true)
		s.addTri(cin, i1, i0, shade(col, 0.55), true, true)
	}
}

func (s *Scene) addBillboard(center Vec3, size float64, col color.RGBA) {
	s.bills = append(s.bills, billboard{center, size, col})
}

// addGroundQuad queues a textured ground tile.
func (s *Scene) addGroundQuad(p0, p1, p2, p3 Vec3, tint float64) {
	s.ground = append(s.ground, gquad{[4]Vec3{p0, p1, p2, p3}, tint})
}

func shade(c color.RGBA, m float64) color.RGBA {
	return color.RGBA{clampByte(float64(c.R) * m), clampByte(float64(c.G) * m), clampByte(float64(c.B) * m), c.A}
}

// ---- flush: project, clip, sort, draw ----

type screenTri struct {
	x     [3]float64
	y     [3]float64
	depth float64
	col   color.RGBA
	uv    bool       // textured ground
	u     [3]float64 // ground UV (pixels)
	v     [3]float64
}

func (s *Scene) Flush(dst *ebiten.Image, cam *Camera, groundImg, white *ebiten.Image) {
	cam.setup()

	// --- ground pass (textured) ---
	var gverts []ebiten.Vertex
	var gidx []uint16
	emitGround := func(a, b, c Vec3, tint float64) {
		// view-space + near clip
		va, vb, vc := cam.view(a), cam.view(b), cam.view(c)
		poly := clipNear([]viewV{{va, a}, {vb, b}, {vc, c}})
		for i := 2; i < len(poly); i++ {
			tri3 := [3]viewV{poly[0], poly[i-1], poly[i]}
			base := uint16(len(gverts))
			for _, vv := range tri3 {
				sx, sy := cam.project(vv.v)
				t := float32(tint)
				gverts = append(gverts, ebiten.Vertex{
					DstX: float32(sx), DstY: float32(sy),
					SrcX: float32(vv.w.X), SrcY: float32(vv.w.Z),
					ColorR: t, ColorG: t, ColorB: t, ColorA: 1,
				})
			}
			gidx = append(gidx, base, base+1, base+2)
		}
	}
	for i := range s.ground {
		q := &s.ground[i]
		emitGround(q.p[0], q.p[1], q.p[2], q.tint)
		emitGround(q.p[0], q.p[2], q.p[3], q.tint)
	}
	if len(gidx) > 0 {
		op := &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear, AntiAlias: false}
		dst.DrawTriangles(gverts, gidx, groundImg, op)
	}

	// --- object pass (shaded solids + billboards), depth sorted ---
	var sts []screenTri
	addShaded := func(t tri) {
		va, vb, vc := cam.view(t.a), cam.view(t.b), cam.view(t.c)
		// back-face cull in view space (normal facing away from camera/origin)
		if t.cull {
			n := vb.Sub(va).Cross(vc.Sub(va))
			// view space: camera at origin looking +Z; visible if normal.z<0
			if n.Z >= 0 {
				return
			}
		}
		col := t.col
		if t.shade {
			n := t.b.Sub(t.a).Cross(t.c.Sub(t.a)).Norm()
			d := n.Dot(lightDir)
			if d < 0 {
				d = -d * 0.3 // weak back-light instead of black
			}
			m := ambient + (1-ambient)*d
			col = shade(t.col, m)
			col.A = t.col.A
		}
		poly := clipNear([]viewV{{va, t.a}, {vb, t.b}, {vc, t.c}})
		for i := 2; i < len(poly); i++ {
			var st screenTri
			st.col = col
			var dsum float64
			for k, vv := range [3]viewV{poly[0], poly[i-1], poly[i]} {
				sx, sy := cam.project(vv.v)
				st.x[k], st.y[k] = sx, sy
				dsum += vv.v.Z
			}
			st.depth = dsum / 3
			sts = append(sts, st)
		}
	}
	for _, t := range s.tris {
		addShaded(t)
	}
	// billboards -> camera-facing quads
	for _, b := range s.bills {
		r := cam.right.Scale(b.size)
		u := cam.up.Scale(b.size)
		p0 := b.center.Sub(r).Sub(u)
		p1 := b.center.Add(r).Sub(u)
		p2 := b.center.Add(r).Add(u)
		p3 := b.center.Sub(r).Add(u)
		for _, triPts := range [2][3]Vec3{{p0, p1, p2}, {p0, p2, p3}} {
			va, vb, vc := cam.view(triPts[0]), cam.view(triPts[1]), cam.view(triPts[2])
			poly := clipNear([]viewV{{va, triPts[0]}, {vb, triPts[1]}, {vc, triPts[2]}})
			for i := 2; i < len(poly); i++ {
				var st screenTri
				st.col = b.col
				var dsum float64
				for k, vv := range [3]viewV{poly[0], poly[i-1], poly[i]} {
					sx, sy := cam.project(vv.v)
					st.x[k], st.y[k] = sx, sy
					dsum += vv.v.Z
				}
				st.depth = dsum / 3
				sts = append(sts, st)
			}
		}
	}

	sort.Slice(sts, func(i, j int) bool { return sts[i].depth > sts[j].depth })

	verts := make([]ebiten.Vertex, 0, len(sts)*3)
	idx := make([]uint16, 0, len(sts)*3)
	for _, st := range sts {
		base := uint16(len(verts))
		cr := float32(st.col.R) / 255
		cg := float32(st.col.G) / 255
		cb := float32(st.col.B) / 255
		ca := float32(st.col.A) / 255
		for k := 0; k < 3; k++ {
			verts = append(verts, ebiten.Vertex{
				DstX: float32(st.x[k]), DstY: float32(st.y[k]),
				SrcX: 0.5, SrcY: 0.5,
				ColorR: cr * ca, ColorG: cg * ca, ColorB: cb * ca, ColorA: ca,
			})
		}
		idx = append(idx, base, base+1, base+2)
	}
	if len(idx) > 0 {
		op := &ebiten.DrawTrianglesOptions{AntiAlias: true}
		dst.DrawTriangles(verts, idx, white, op)
	}
}

// viewV pairs a view-space position with its original world position (for
// texture UVs that must survive near-plane clipping).
type viewV struct {
	v viewPos
	w Vec3
}
type viewPos = Vec3

// clipNear clips a convex polygon against the near plane z >= nearZ in view
// space, interpolating world positions too.
func clipNear(poly []viewV) []viewV {
	const nearZ = 1.0
	out := poly[:0:0]
	n := len(poly)
	for i := 0; i < n; i++ {
		cur := poly[i]
		nxt := poly[(i+1)%n]
		curIn := cur.v.Z >= nearZ
		nxtIn := nxt.v.Z >= nearZ
		if curIn {
			out = append(out, cur)
		}
		if curIn != nxtIn {
			t := (nearZ - cur.v.Z) / (nxt.v.Z - cur.v.Z)
			iv := Vec3{
				cur.v.X + (nxt.v.X-cur.v.X)*t,
				cur.v.Y + (nxt.v.Y-cur.v.Y)*t,
				nearZ,
			}
			iw := Vec3{
				cur.w.X + (nxt.w.X-cur.w.X)*t,
				cur.w.Y + (nxt.w.Y-cur.w.Y)*t,
				cur.w.Z + (nxt.w.Z-cur.w.Z)*t,
			}
			out = append(out, viewV{iv, iw})
		}
	}
	return out
}
