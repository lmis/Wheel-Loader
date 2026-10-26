package main

import (
	"image/color"
	"math"
)

// ---- 3D vector math ----
//
// The simulation lives on the ground plane in (X, Y) world coordinates. For
// rendering we lift that into a right-handed 3D space with Y up, mapping a
// ground point (x, y) at height h to (x, h, y).

type Vec3 struct{ X, Y, Z float64 }

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

// rotY rotates v about the vertical axis by yaw, using the same convention as
// the 2D sim (yaw 0 faces +X, increasing yaw turns toward +Z).
func rotY(v Vec3, yaw float64) Vec3 {
	si, co := math.Sincos(yaw)
	return Vec3{v.X*co - v.Z*si, v.Y, v.X*si + v.Z*co}
}

// ---- lighting & atmosphere ----

var (
	lightDir = Vec3{-0.45, 1.0, -0.35}.Norm() // direction toward the sun
	skyAmb   = [3]float64{0.50, 0.56, 0.66}   // light from above (bluish)
	gndAmb   = [3]float64{0.40, 0.35, 0.30}   // bounce from the ground (warm)
	sunCol   = [3]float64{0.72, 0.65, 0.54}
	fogCol   = [3]float64{172, 190, 206}
	skyTop   = [3]float64{92, 132, 184}
)

const (
	fogStart = 1000.0
	fogEnd   = 3600.0
	fogMax   = 0.9
)

// ---- scene primitives ----

type tri struct {
	p     [3]Vec3
	c     [3]color.RGBA
	n     [3]Vec3 // per-vertex normals; zero means use the face normal
	shade bool    // apply lighting
	cull  bool    // back-face cull (closed solids only)
	bias  float64 // sort bias toward the camera (surface-mounted details)
	tex   bool    // sample the atlas at uv (heaps); otherwise flat colour
	uv    [3][2]float32
}

type gquad struct {
	p [4]Vec3 // world positions on the ground plane
}

type billboard struct {
	center Vec3
	size   float64
	col    color.RGBA
}

// Scene accumulates the world-space geometry for one frame; the GPU renderer
// (gpu.go) draws it.
type Scene struct {
	tris   []tri
	ground []gquad
	skirt  []gquad // ground beyond the site, drawn with a repeating texture
	decals []tri   // flat ground decals (shadows), drawn before objects
	bills  []billboard

	eye  Vec3    // camera position this frame (for level of detail)
	bias float64 // sort bias applied to geometry added while it is set
	glow bool    // geometry added while set is unlit: it glows
}

func (s *Scene) reset() {
	s.tris = s.tris[:0]
	s.ground = s.ground[:0]
	s.skirt = s.skirt[:0]
	s.decals = s.decals[:0]
	s.bills = s.bills[:0]
}

// addTri adds a flat-colored triangle.
func (s *Scene) addTri(a, b, c Vec3, col color.RGBA, shade, cull bool) {
	s.push(tri{p: [3]Vec3{a, b, c}, c: [3]color.RGBA{col, col, col}, shade: shade, cull: cull})
}

// addTriV adds a triangle with per-vertex colors and normals (smooth shading).
func (s *Scene) addTriV(p [3]Vec3, c [3]color.RGBA, n [3]Vec3, cull bool) {
	s.push(tri{p: p, c: c, n: n, shade: true, cull: cull})
}

// addTriUV adds a smooth-shaded triangle textured from the object atlas.
func (s *Scene) addTriUV(p [3]Vec3, c [3]color.RGBA, n [3]Vec3, uv [3][2]float32, cull bool) {
	s.push(tri{p: p, c: c, n: n, shade: true, cull: cull, tex: true, uv: uv})
}

// maxTriEdge bounds triangle size. Objects are depth-sorted per triangle
// (painter's algorithm), so a wall made of two huge triangles can sort in
// front of a door or window sitting on it. Splitting big triangles keeps the
// sort local and fixes those overlaps.
const maxTriEdge = 22.0

// push appends t, bisecting its longest edge until every edge is short.
// Bisection keeps the winding and interpolates colors and normals.
func (s *Scene) push(t tri) {
	l0 := t.p[0].Sub(t.p[1]).Len() // edge 0-1
	l1 := t.p[1].Sub(t.p[2]).Len() // edge 1-2
	l2 := t.p[2].Sub(t.p[0]).Len() // edge 2-0
	// far away, triangles are small on screen and sort errors are invisible,
	// so the size limit grows with distance from the camera
	limit := maxTriEdge * math.Max(1, t.p[0].Sub(s.eye).Len()/450)
	if l0 <= limit && l1 <= limit && l2 <= limit {
		if t.bias == 0 {
			t.bias = s.bias
		}
		if s.glow {
			t.shade = false
		}
		s.tris = append(s.tris, t)
		return
	}
	// rotate so the longest edge is 0-1, then split it at m
	r := 0
	if l1 >= l0 && l1 >= l2 {
		r = 1
	} else if l2 >= l0 && l2 >= l1 {
		r = 2
	}
	a, b, c := r, (r+1)%3, (r+2)%3
	mp := t.p[a].Add(t.p[b]).Scale(0.5)
	var mc color.RGBA
	mc.R = uint8((int(t.c[a].R) + int(t.c[b].R)) / 2)
	mc.G = uint8((int(t.c[a].G) + int(t.c[b].G)) / 2)
	mc.B = uint8((int(t.c[a].B) + int(t.c[b].B)) / 2)
	mc.A = uint8((int(t.c[a].A) + int(t.c[b].A)) / 2)
	var mn Vec3
	if t.n[a] != (Vec3{}) {
		mn = t.n[a].Add(t.n[b]).Norm()
	}
	muv := [2]float32{(t.uv[a][0] + t.uv[b][0]) / 2, (t.uv[a][1] + t.uv[b][1]) / 2}
	t1, t2 := t, t
	t1.p, t1.c, t1.n = [3]Vec3{t.p[a], mp, t.p[c]}, [3]color.RGBA{t.c[a], mc, t.c[c]}, [3]Vec3{t.n[a], mn, t.n[c]}
	t2.p, t2.c, t2.n = [3]Vec3{mp, t.p[b], t.p[c]}, [3]color.RGBA{mc, t.c[b], t.c[c]}, [3]Vec3{mn, t.n[b], t.n[c]}
	t1.uv = [3][2]float32{t.uv[a], muv, t.uv[c]}
	t2.uv = [3][2]float32{muv, t.uv[b], t.uv[c]}
	s.push(t1)
	s.push(t2)
}

// addQuad adds a flat quad (two triangles) with a shared color.
func (s *Scene) addQuad(a, b, c, d Vec3, col color.RGBA, shade, cull bool) {
	s.addTri(a, b, c, col, shade, cull)
	s.addTri(a, c, d, col, shade, cull)
}

// newell returns the (unnormalized) Newell normal of a polygon; robust to
// repeated vertices.
func newell(pts []Vec3) Vec3 {
	var n Vec3
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		n.X += (a.Y - b.Y) * (a.Z + b.Z)
		n.Y += (a.Z - b.Z) * (a.X + b.X)
		n.Z += (a.X - b.X) * (a.Y + b.Y)
	}
	return n
}

// addFaceOutward adds a convex planar polygon as a triangle fan, flipping
// its winding if needed so its normal points away from `inside`.
func (s *Scene) addFaceOutward(pts []Vec3, inside Vec3, col color.RGBA) {
	n := newell(pts)
	var fc Vec3
	for _, p := range pts {
		fc = fc.Add(p)
	}
	fc = fc.Scale(1 / float64(len(pts)))
	if n.Dot(fc.Sub(inside)) < 0 {
		for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
			pts[i], pts[j] = pts[j], pts[i]
		}
	}
	for i := 2; i < len(pts); i++ {
		s.addTri(pts[0], pts[i-1], pts[i], col, true, true)
	}
}

// addOBB2 adds an oriented box with local forward axis fwd and an "up" hint;
// hl/hh/hw are half length (along fwd), height (along up) and width.
func (s *Scene) addOBB2(center, fwd, upHint Vec3, hl, hh, hw float64, col color.RGBA) {
	f := fwd.Norm()
	up := upHint.Sub(f.Scale(upHint.Dot(f))).Norm()
	if up.Len() < 0.5 {
		up = Vec3{0, 1, 0}
		if math.Abs(f.Dot(up)) > 0.95 {
			up = Vec3{1, 0, 0}
		}
		up = up.Sub(f.Scale(up.Dot(f))).Norm()
	}
	right := f.Cross(up).Norm()
	c := func(sl, sh, sw float64) Vec3 {
		return center.Add(f.Scale(sl * hl)).Add(up.Scale(sh * hh)).Add(right.Scale(sw * hw))
	}
	faces := [6][4]Vec3{
		{c(-1, 1, -1), c(1, 1, -1), c(1, 1, 1), c(-1, 1, 1)},     // +up
		{c(-1, -1, -1), c(-1, -1, 1), c(1, -1, 1), c(1, -1, -1)}, // -up
		{c(1, -1, -1), c(1, -1, 1), c(1, 1, 1), c(1, 1, -1)},     // +fwd
		{c(-1, -1, -1), c(-1, 1, -1), c(-1, 1, 1), c(-1, -1, 1)}, // -fwd
		{c(-1, -1, 1), c(-1, 1, 1), c(1, 1, 1), c(1, -1, 1)},     // +right
		{c(-1, -1, -1), c(1, -1, -1), c(1, 1, -1), c(-1, 1, -1)}, // -right
	}
	for _, fc := range faces {
		pts := []Vec3{fc[0], fc[1], fc[2], fc[3]}
		s.addFaceOutward(pts, center, col)
	}
}

// addOBB is addOBB2 with world-up as the up hint.
func (s *Scene) addOBB(center, fwd Vec3, hl, hh, hw float64, col color.RGBA) {
	s.addOBB2(center, fwd, Vec3{0, 1, 0}, hl, hh, hw, col)
}

// addBox adds an upright box centred at c with half-extents (hx,hy,hz),
// rotated by yaw about the vertical axis (hx runs along the yaw direction).
func (s *Scene) addBox(c Vec3, hx, hy, hz, yaw float64, col color.RGBA) {
	s.addOBB2(c, rotY(Vec3{1, 0, 0}, yaw), Vec3{0, 1, 0}, hx, hy, hz, col)
}

// ring is one horizontal cross-section of a prism: a rectangle spanning
// [back, front] along the local forward axis and ±hw laterally, with its
// four vertical edges chamfered by ch, at height y.
type ring struct {
	back, front, hw, y, ch float64
}

func (r ring) points() [8][2]float64 {
	ch := math.Min(r.ch, math.Min(r.hw, (r.front-r.back)/2)*0.95)
	return [8][2]float64{
		{r.front, -r.hw + ch}, {r.front, r.hw - ch},
		{r.front - ch, r.hw}, {r.back + ch, r.hw},
		{r.back, r.hw - ch}, {r.back, -r.hw + ch},
		{r.back + ch, -r.hw}, {r.front - ch, -r.hw},
	}
}

// addPrism lofts a solid through a stack of rings (bottom to top). origin is
// the ground-level local origin, yaw the heading. Stacking rings of different
// sizes gives tapered hoods, chamfered roofs and bevelled edges.
func (s *Scene) addPrism(origin Vec3, yaw float64, rings []ring, col color.RGBA) {
	s.addPrismCaps(origin, yaw, rings, col, true, true)
}

// addPrismCaps is addPrism with control over the end caps. When solids are
// stacked (a glass band on a cab body, a roof on the glass), the hidden caps
// between them would be nearly coplanar and fight in the depth sort, so they
// are left off.
func (s *Scene) addPrismCaps(origin Vec3, yaw float64, rings []ring, col color.RGBA, capBottom, capTop bool) {
	world := make([][8]Vec3, len(rings))
	var centroid Vec3
	n := 0
	for i, r := range rings {
		for k, p := range r.points() {
			w := origin.Add(rotY(Vec3{p[0], r.y, p[1]}, yaw))
			world[i][k] = w
			centroid = centroid.Add(w)
			n++
		}
	}
	centroid = centroid.Scale(1 / float64(n))
	for i := 0; i+1 < len(world); i++ {
		a, b := world[i], world[i+1]
		for k := 0; k < 8; k++ {
			k2 := (k + 1) % 8
			s.addFaceOutward([]Vec3{a[k], a[k2], b[k2], b[k]}, centroid, col)
		}
	}
	if capBottom {
		bot := world[0]
		s.addFaceOutward(bot[:], centroid, col)
	}
	if capTop {
		top := world[len(world)-1]
		s.addFaceOutward(top[:], centroid, col)
	}
}

// addCylinder adds an upright low-poly cylinder with smooth sides.
func (s *Scene) addCylinder(c Vec3, radius, h float64, sides int, col color.RGBA) {
	s.addCone(c, radius, radius, h, sides, col)
}

// addCone adds an upright truncated cone (r0 at the bottom, r1 at the top)
// with smooth side normals and flat caps.
func (s *Scene) addCone(c Vec3, r0, r1, h float64, sides int, col color.RGBA) {
	if sides < 5 {
		sides = 8
	}
	top := c.Y + h/2
	bot := c.Y - h/2
	slope := (r0 - r1) / h
	cols := [3]color.RGBA{col, col, col}
	for i := 0; i < sides; i++ {
		a0 := 2 * math.Pi * float64(i) / float64(sides)
		a1 := 2 * math.Pi * float64(i+1) / float64(sides)
		d0 := Vec3{math.Cos(a0), 0, math.Sin(a0)}
		d1 := Vec3{math.Cos(a1), 0, math.Sin(a1)}
		n0 := Vec3{d0.X, slope, d0.Z}.Norm()
		n1 := Vec3{d1.X, slope, d1.Z}.Norm()
		b0 := Vec3{c.X + d0.X*r0, bot, c.Z + d0.Z*r0}
		b1 := Vec3{c.X + d1.X*r0, bot, c.Z + d1.Z*r0}
		t0 := Vec3{c.X + d0.X*r1, top, c.Z + d0.Z*r1}
		t1 := Vec3{c.X + d1.X*r1, top, c.Z + d1.Z*r1}
		// outward winding for a right-handed space seen from outside
		s.addTriV([3]Vec3{b0, t0, t1}, cols, [3]Vec3{n0, n0, n1}, true)
		s.addTriV([3]Vec3{b0, t1, b1}, cols, [3]Vec3{n0, n1, n1}, true)
		if r1 > 0.01 {
			s.addTri(Vec3{c.X, top, c.Z}, t1, t0, col, true, true)
		}
	}
}

// addTube adds a hollow cylinder along axis: smooth outside, ring ends and a
// darker bore (concrete pipes, drums).
func (s *Scene) addTube(center, axis Vec3, radius, length float64, sides int, col, bore color.RGBA) {
	a := axis.Norm()
	ref := Vec3{0, 1, 0}
	if math.Abs(a.Dot(ref)) > 0.9 {
		ref = Vec3{1, 0, 0}
	}
	u := a.Cross(ref).Norm()
	v := a.Cross(u)
	half := a.Scale(length / 2)
	c0, c1 := center.Sub(half), center.Add(half)
	cols := [3]color.RGBA{col, col, col}
	for i := 0; i < sides; i++ {
		a0 := 2 * math.Pi * float64(i) / float64(sides)
		a1 := 2 * math.Pi * float64(i+1) / float64(sides)
		d0 := u.Scale(math.Cos(a0)).Add(v.Scale(math.Sin(a0)))
		d1 := u.Scale(math.Cos(a1)).Add(v.Scale(math.Sin(a1)))
		p00, p01 := c0.Add(d0.Scale(radius)), c1.Add(d0.Scale(radius))
		p10, p11 := c0.Add(d1.Scale(radius)), c1.Add(d1.Scale(radius))
		t1 := [3]Vec3{p00, p11, p01}
		t2 := [3]Vec3{p00, p10, p11}
		n1 := [3]Vec3{d0, d1, d0}
		n2 := [3]Vec3{d0, d1, d1}
		if t1[1].Sub(t1[0]).Cross(t1[2].Sub(t1[0])).Dot(d0.Add(d1)) < 0 {
			t1[1], t1[2], n1[1], n1[2] = t1[2], t1[1], n1[2], n1[1]
			t2[1], t2[2], n2[1], n2[2] = t2[2], t2[1], n2[2], n2[1]
		}
		s.addTriV(t1, cols, n1, true)
		s.addTriV(t2, cols, n2, true)
		for _, end := range []struct {
			c   Vec3
			out Vec3
		}{{c0, a.Scale(-1)}, {c1, a}} {
			o0, o1 := end.c.Add(d0.Scale(radius)), end.c.Add(d1.Scale(radius))
			i0, i1 := end.c.Add(d0.Scale(radius*0.78)), end.c.Add(d1.Scale(radius*0.78))
			s.addFaceOutward([]Vec3{o0, o1, i1, i0}, end.c.Sub(end.out), shadeC(col, 0.92))
			// the bore: a dark disc set slightly back into the pipe
			dc := end.c.Sub(end.out.Scale(1.5))
			s.addFaceOutward([]Vec3{dc, dc.Add(d0.Scale(radius * 0.78)), dc.Add(d1.Scale(radius * 0.78))}, end.c.Sub(end.out.Scale(10)), bore)
		}
	}
}

// addWheel adds a tyre lying on its side: smooth tread, rim, hub and tread
// lugs. yaw is the rolling direction; spin rotates the lugs with travel.
func (s *Scene) addWheel(center Vec3, radius, width, yaw, spin float64, tire, rim color.RGBA) {
	const sides = 18
	fwd := rotY(Vec3{1, 0, 0}, yaw)
	axle := rotY(Vec3{0, 0, 1}, yaw)
	up := Vec3{0, 1, 0}
	half := axle.Scale(width / 2)
	cin := center.Sub(half)
	cout := center.Add(half)
	radial := func(a float64) Vec3 { return up.Scale(math.Sin(a)).Add(fwd.Scale(math.Cos(a))) }
	tc := [3]color.RGBA{tire, tire, tire}
	for i := 0; i < sides; i++ {
		a0 := 2 * math.Pi * float64(i) / sides
		a1 := 2 * math.Pi * float64(i+1) / sides
		d0, d1 := radial(a0), radial(a1)
		o0, o1 := cout.Add(d0.Scale(radius)), cout.Add(d1.Scale(radius))
		i0, i1 := cin.Add(d0.Scale(radius)), cin.Add(d1.Scale(radius))
		s.addTriV([3]Vec3{i0, o0, o1}, tc, [3]Vec3{d0, d0, d1}, false)
		s.addTriV([3]Vec3{i0, o1, i1}, tc, [3]Vec3{d0, d1, d1}, false)
		// sidewalls: tyre annulus + rim disc on both faces
		for _, face := range []struct {
			c   Vec3
			out Vec3
		}{{cout, axle}, {cin, axle.Scale(-1)}} {
			r0 := face.c.Add(d0.Scale(radius * 0.97))
			r1 := face.c.Add(d1.Scale(radius * 0.97))
			m0 := face.c.Add(d0.Scale(radius * 0.6))
			m1 := face.c.Add(d1.Scale(radius * 0.6))
			inner := face.c.Sub(face.out.Scale(width * 0.12))
			s.addFaceOutward([]Vec3{m0, r0, r1, m1}, face.c.Sub(face.out), shadeC(tire, 0.9))
			s.addFaceOutward([]Vec3{inner, m0, m1}, face.c.Sub(face.out.Scale(2)), rim)
		}
	}
	// hub cap
	s.addOBB2(cout.Add(axle.Scale(0.5)), axle, up, 1.5, radius*0.22, radius*0.22, shadeC(rim, 0.7))
	// tread lugs (only up close; they are invisible at a distance)
	const lugs = 9
	if center.Sub(s.eye).Len() > 800 {
		return
	}
	for i := 0; i < lugs; i++ {
		a := 2*math.Pi*float64(i)/lugs - spin
		d := radial(a)
		lc := center.Add(d.Scale(radius + 1.2))
		s.addOBB2(lc, axle, d, width*0.42, 1.4, 2.6, shadeC(tire, 1.15))
	}
}

// addRoundWheel adds a plain round wheel: a tread band and two flat faces,
// turning about axle, with a hub disc on each face.
func (s *Scene) addRoundWheel(center, fwd, axle Vec3, radius, width float64, sides int, col, hub color.RGBA) {
	fwd, axle = fwd.Norm(), axle.Norm()
	up := axle.Cross(fwd).Norm()
	half := axle.Scale(width / 2)
	radial := func(i int) Vec3 {
		a := 2 * math.Pi * float64(i) / float64(sides)
		return up.Scale(math.Sin(a)).Add(fwd.Scale(math.Cos(a)))
	}
	cols := [3]color.RGBA{col, col, col}
	for i := 0; i < sides; i++ {
		d0, d1 := radial(i), radial(i+1)
		o0, o1 := center.Add(half).Add(d0.Scale(radius)), center.Add(half).Add(d1.Scale(radius))
		i0, i1 := center.Sub(half).Add(d0.Scale(radius)), center.Sub(half).Add(d1.Scale(radius))
		s.addTriV([3]Vec3{i0, o0, o1}, cols, [3]Vec3{d0, d0, d1}, false)
		s.addTriV([3]Vec3{i0, o1, i1}, cols, [3]Vec3{d0, d1, d1}, false)
		for _, side := range []float64{1, -1} {
			c := center.Add(half.Scale(side))
			r0, r1 := c.Add(d0.Scale(radius)), c.Add(d1.Scale(radius))
			h0, h1 := c.Add(d0.Scale(radius*0.45)), c.Add(d1.Scale(radius*0.45))
			s.addFaceOutward([]Vec3{h0, r0, r1, h1}, center, col)
			s.addFaceOutward([]Vec3{c, h0, h1}, center, hub)
		}
	}
}

func (s *Scene) addBillboard(center Vec3, size float64, col color.RGBA) {
	s.bills = append(s.bills, billboard{center, size, col})
}

// addGroundQuad queues a textured ground tile.
func (s *Scene) addGroundQuad(p0, p1, p2, p3 Vec3) {
	s.ground = append(s.ground, gquad{[4]Vec3{p0, p1, p2, p3}})
}

// addSkirtQuad queues a ground tile outside the site (repeating texture).
func (s *Scene) addSkirtQuad(p0, p1, p2, p3 Vec3) {
	s.skirt = append(s.skirt, gquad{[4]Vec3{p0, p1, p2, p3}})
}

// addDecalQuad queues a flat translucent quad lying on the ground. Decals are
// drawn right after the ground so depth sorting can never put them on top of
// the objects standing on them.
func (s *Scene) addDecalQuad(a, b, c, d Vec3, col color.RGBA) {
	cols := [3]color.RGBA{col, col, col}
	s.decals = append(s.decals, tri{p: [3]Vec3{a, b, c}, c: cols}, tri{p: [3]Vec3{a, c, d}, c: cols})
}

// detail adds geometry built by fn with a sort bias, for parts mounted on a
// larger surface (doors, windows, ribs, grilles) so they draw on top of it.
// addBlob adds a soft oval ground decal centred at c (on the ground), dark
// in the middle and fading to nothing at the rim: a contact shadow or a glow.
func (s *Scene) addBlob(c Vec3, rx, rz, yaw float64, col color.RGBA) {
	const k = 16
	mid := col
	mid.A = uint8(float64(col.A) * 0.7)
	rim := col
	rim.A = 0
	f, r := rotY(Vec3{1, 0, 0}, yaw), rotY(Vec3{0, 0, 1}, yaw)
	pt := func(i int, sc float64) Vec3 {
		a := 2 * math.Pi * float64(i) / k
		return c.Add(f.Scale(math.Cos(a) * rx * sc)).Add(r.Scale(math.Sin(a) * rz * sc))
	}
	for i := 0; i < k; i++ {
		s.decals = append(s.decals,
			tri{p: [3]Vec3{c, pt(i, 0.55), pt(i+1, 0.55)}, c: [3]color.RGBA{col, mid, mid}},
			tri{p: [3]Vec3{pt(i, 0.55), pt(i, 1), pt(i+1, 1)}, c: [3]color.RGBA{mid, rim, rim}},
			tri{p: [3]Vec3{pt(i, 0.55), pt(i+1, 1), pt(i+1, 0.55)}, c: [3]color.RGBA{mid, rim, mid}})
	}
}

// scaleSince scales the triangles added since mark (a len(s.tris) taken
// earlier) about a point. Uniform scaling keeps normals and culling valid.
func (s *Scene) scaleSince(mark int, about Vec3, k float64) {
	for i := mark; i < len(s.tris); i++ {
		p := &s.tris[i].p
		for j := range p {
			p[j] = about.Add(p[j].Sub(about).Scale(k))
		}
	}
}

// glowing adds geometry built by fn unlit, so it shines at full colour.
func (s *Scene) glowing(fn func()) {
	prev := s.glow
	s.glow = true
	fn()
	s.glow = prev
}

func (s *Scene) detail(fn func()) {
	prev := s.bias
	s.bias = 4
	fn()
	s.bias = prev
}

// shadeC scales a color's RGB, keeping alpha.
func shadeC(c color.RGBA, m float64) color.RGBA {
	return color.RGBA{clampByte(float64(c.R) * m), clampByte(float64(c.G) * m), clampByte(float64(c.B) * m), c.A}
}
