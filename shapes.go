package main

import (
	"image/color"
	"math"
)

// Smooth, rounded shapes for the characters: ellipsoids, capsules and
// rounded-corner lofts. They use per-vertex normals for smooth shading and
// fewer segments far from the camera.

// lod is 2 up close, 1 at middle distance and 0 far away.
func (s *Scene) lod(c Vec3) int {
	switch d := c.Sub(s.eye).Len(); {
	case d < 320:
		return 2
	case d < 900:
		return 1
	}
	return 0
}

// smoothTri adds a back-face-culled, smooth-shaded triangle, fixing its
// winding so it faces along its normals.
func (s *Scene) smoothTri(p [3]Vec3, n [3]Vec3, col color.RGBA) {
	face := p[1].Sub(p[0]).Cross(p[2].Sub(p[0]))
	if face.Len() < 1e-9 {
		return // degenerate (at a pole)
	}
	if face.Dot(n[0].Add(n[1]).Add(n[2])) < 0 {
		p[1], p[2] = p[2], p[1]
		n[1], n[2] = n[2], n[1]
	}
	s.addTriV(p, [3]color.RGBA{col, col, col}, n, true)
}

// grid joins rows of points (each row a closed loop of the same length)
// into a smooth surface.
func (s *Scene) grid(pts, nrm [][]Vec3, col color.RGBA) {
	for r := 0; r+1 < len(pts); r++ {
		a, b := pts[r], pts[r+1]
		na, nb := nrm[r], nrm[r+1]
		m := len(a)
		for i := 0; i < m; i++ {
			j := (i + 1) % m
			s.smoothTri([3]Vec3{a[i], b[i], b[j]}, [3]Vec3{na[i], nb[i], nb[j]}, col)
			s.smoothTri([3]Vec3{a[i], b[j], a[j]}, [3]Vec3{na[i], nb[j], na[j]}, col)
		}
	}
}

// frame3 builds an orthonormal frame from a forward axis and an up hint.
func frame3(fwd, upHint Vec3) (f, u, r Vec3) {
	f = fwd.Norm()
	u = upHint.Sub(f.Scale(upHint.Dot(f)))
	if u.Len() < 1e-6 {
		u = Vec3{0, 1, 0}
		if math.Abs(f.Y) > 0.9 {
			u = Vec3{1, 0, 0}
		}
		u = u.Sub(f.Scale(u.Dot(f)))
	}
	u = u.Norm()
	r = f.Cross(u).Norm()
	return
}

// addEllipsoid adds an ellipsoid with radii rf along fwd, ru along up and
// rs to the side.
func (s *Scene) addEllipsoid(c, fwd, up Vec3, rf, ru, rs float64, col color.RGBA) {
	f, u, r := frame3(fwd, up)
	seg, rings := [3]int{6, 10, 14}[s.lod(c)], [3]int{4, 6, 9}[s.lod(c)]
	pts := make([][]Vec3, 0, rings+1)
	nrm := make([][]Vec3, 0, rings+1)
	for k := 0; k <= rings; k++ {
		phi := -math.Pi/2 + math.Pi*float64(k)/float64(rings)
		cp, sp := math.Cos(phi), math.Sin(phi)
		row, nrow := make([]Vec3, seg), make([]Vec3, seg)
		for i := 0; i < seg; i++ {
			th := 2 * math.Pi * float64(i) / float64(seg)
			ct, st := math.Cos(th), math.Sin(th)
			row[i] = c.Add(f.Scale(rf * cp * ct)).Add(r.Scale(rs * cp * st)).Add(u.Scale(ru * sp))
			nrow[i] = f.Scale(cp * ct / rf).Add(r.Scale(cp * st / rs)).Add(u.Scale(sp / ru)).Norm()
		}
		pts, nrm = append(pts, row), append(nrm, nrow)
	}
	s.grid(pts, nrm, col)
}

// addSphere is a round ellipsoid.
func (s *Scene) addSphere(c Vec3, r float64, col color.RGBA) {
	s.addEllipsoid(c, Vec3{1, 0, 0}, Vec3{0, 1, 0}, r, r, r, col)
}

// addCapsule adds a rounded limb from a (radius ra) to b (radius rb).
func (s *Scene) addCapsule(a, b Vec3, ra, rb float64, col color.RGBA) {
	d := b.Sub(a)
	L := d.Len()
	if L < 1e-6 {
		s.addSphere(a, math.Max(ra, rb), col)
		return
	}
	ax, u, r := frame3(d, Vec3{0, 1, 0})
	q := s.lod(a.Add(d.Scale(0.5)))
	seg, capN := [3]int{5, 7, 10}[q], [3]int{1, 2, 3}[q]
	type ringPt struct{ t, rad, ca, sa float64 } // position, radius, normal (radial, axial)
	var rings []ringPt
	for k := 0; k <= capN; k++ { // cap at a, from its tip to the equator
		al := -math.Pi/2 + math.Pi/2*float64(k)/float64(capN)
		rings = append(rings, ringPt{math.Sin(al) * ra, math.Cos(al) * ra, math.Cos(al), math.Sin(al)})
	}
	for k := 0; k <= capN; k++ { // cap at b, from the equator to its tip
		al := math.Pi / 2 * float64(k) / float64(capN)
		rings = append(rings, ringPt{L + math.Sin(al)*rb, math.Cos(al) * rb, math.Cos(al), math.Sin(al)})
	}
	pts := make([][]Vec3, len(rings))
	nrm := make([][]Vec3, len(rings))
	for k, rp := range rings {
		pts[k], nrm[k] = make([]Vec3, seg), make([]Vec3, seg)
		for i := 0; i < seg; i++ {
			th := 2 * math.Pi * float64(i) / float64(seg)
			radial := u.Scale(math.Cos(th)).Add(r.Scale(math.Sin(th)))
			pts[k][i] = a.Add(ax.Scale(rp.t)).Add(radial.Scale(rp.rad))
			nrm[k][i] = radial.Scale(rp.ca).Add(ax.Scale(rp.sa)).Norm()
		}
	}
	s.grid(pts, nrm, col)
}

// addRoundPrism lofts rings like addPrism, but each ring is a rectangle with
// properly rounded corners (radius ch), smooth-shaded. capBottom/capTop
// close the ends.
func (s *Scene) addRoundPrism(origin Vec3, yaw float64, rings []ring, col color.RGBA, capBottom, capTop bool) {
	if len(rings) == 0 {
		return
	}
	per := [3]int{1, 2, 3}[s.lod(origin)] // points per corner arc, beyond its start
	ringPts := func(rg ring) ([]Vec3, []Vec3) {
		ch := math.Min(rg.ch, math.Min(rg.hw, (rg.front-rg.back)/2)*0.98)
		corners := [4][3]float64{ // centre x, centre z, start angle
			{rg.front - ch, rg.hw - ch, 0},
			{rg.back + ch, rg.hw - ch, math.Pi / 2},
			{rg.back + ch, -rg.hw + ch, math.Pi},
			{rg.front - ch, -rg.hw + ch, 3 * math.Pi / 2},
		}
		var p, n []Vec3
		for _, cn := range corners {
			for k := 0; k <= per; k++ {
				a := cn[2] + math.Pi/2*float64(k)/float64(per)
				dx, dz := math.Cos(a), math.Sin(a)
				p = append(p, origin.Add(rotY(Vec3{cn[0] + dx*ch, rg.y, cn[1] + dz*ch}, yaw)))
				n = append(n, rotY(Vec3{dx, 0, dz}, yaw))
			}
		}
		return p, n
	}
	pts := make([][]Vec3, len(rings))
	nrm := make([][]Vec3, len(rings))
	for i, rg := range rings {
		pts[i], nrm[i] = ringPts(rg)
	}
	// tilt the normals with the taper between rings
	for i := range rings {
		lo, hi := i-1, i+1
		if lo < 0 {
			lo = 0
		}
		if hi >= len(rings) {
			hi = len(rings) - 1
		}
		dy := rings[hi].y - rings[lo].y
		if dy == 0 {
			continue
		}
		for j := range nrm[i] {
			out := pts[lo][j].Sub(pts[hi][j]) // horizontal shrink as we rise
			out.Y = 0
			nrm[i][j] = nrm[i][j].Add(Vec3{0, out.Dot(nrm[i][j]) / dy, 0}).Norm()
		}
	}
	s.grid(pts, nrm, col)
	cap := func(row []Vec3, up float64) {
		var c Vec3
		for _, p := range row {
			c = c.Add(p)
		}
		c = c.Scale(1 / float64(len(row)))
		nn := Vec3{0, up, 0}
		for j := range row {
			s.smoothTri([3]Vec3{c, row[j], row[(j+1)%len(row)]}, [3]Vec3{nn, nn, nn}, col)
		}
	}
	if capBottom {
		cap(pts[0], -1)
	}
	if capTop {
		cap(pts[len(pts)-1], 1)
	}
}

// ellPoint is the point at azimuth th (from fwd toward the side) and
// elevation ph on an ellipsoid, with its outward normal.
func ellPoint(c, f, u, r Vec3, rf, ru, rs, th, ph float64) (Vec3, Vec3) {
	cp, sp := math.Cos(ph), math.Sin(ph)
	ct, st := math.Cos(th), math.Sin(th)
	p := c.Add(f.Scale(rf * cp * ct)).Add(r.Scale(rs * cp * st)).Add(u.Scale(ru * sp))
	n := f.Scale(cp * ct / rf).Add(r.Scale(cp * st / rs)).Add(u.Scale(sp / ru)).Norm()
	return p, n
}

// addSpot adds a small flat oval marking lying on a surface with normal n,
// stretched ra along the horizontal and rb across it.
func (s *Scene) addSpot(c, n Vec3, ra, rb float64, col color.RGBA) {
	n = n.Norm()
	a := Vec3{0, 1, 0}.Cross(n)
	if a.Len() < 1e-3 {
		a = Vec3{1, 0, 0}
	}
	a = a.Norm()
	b := n.Cross(a)
	c = c.Add(n.Scale(0.05))
	const k = 6
	nn := [3]Vec3{n, n, n}
	for i := 0; i < k; i++ {
		t0, t1 := 2*math.Pi*float64(i)/k, 2*math.Pi*float64(i+1)/k
		p0 := c.Add(a.Scale(ra * math.Cos(t0))).Add(b.Scale(rb * math.Sin(t0)))
		p1 := c.Add(a.Scale(ra * math.Cos(t1))).Add(b.Scale(rb * math.Sin(t1)))
		s.smoothTri([3]Vec3{c, p0, p1}, nn, col)
	}
}

// rotateSince turns the triangles added since mark (a len(s.tris) taken
// earlier) by ang about the axis through about.
func (s *Scene) rotateSince(mark int, about, axis Vec3, ang float64) {
	k := axis.Norm()
	for i := mark; i < len(s.tris); i++ {
		t := &s.tris[i]
		for j := range t.p {
			t.p[j] = about.Add(rotAbout(t.p[j].Sub(about), k, ang))
			if t.n[j] != (Vec3{}) {
				t.n[j] = rotAbout(t.n[j], k, ang)
			}
		}
	}
}

// addSteeringWheel adds a round steering wheel tilted toward the driver,
// with a spoke and its column.
func (s *Scene) addSteeringWheel(center Vec3, yaw, radius float64, col color.RGBA) {
	axis := rotY(Vec3{-0.85, 0.55, 0}, yaw).Norm()
	a := rotY(Vec3{0, 0, 1}, yaw)
	b := axis.Cross(a).Norm()
	const n = 14
	pt := func(i int) Vec3 {
		t := 2 * math.Pi * float64(i) / n
		return center.Add(a.Scale(math.Cos(t) * radius)).Add(b.Scale(math.Sin(t) * radius))
	}
	rim := radius * 0.1
	for i := 0; i < n; i++ {
		s.addCapsule(pt(i), pt(i+1), rim, rim, col)
	}
	s.addCapsule(center.Sub(a.Scale(radius)), center.Add(a.Scale(radius)), rim*0.7, rim*0.7, col) // spoke
	s.addCapsule(center, center.Sub(axis.Scale(radius*0.8)), rim*1.2, rim*1.2, col)               // column
}

// loftStation is one elliptical cross-section of a loft: at x along the
// loft's forward axis, centred y above its origin, with half width hw and
// half height hh.
type loftStation struct{ x, y, hw, hh float64 }

// addLoft adds one smooth closed body through elliptical cross-sections, so
// a body built of several bulges has no seams where separate shapes would
// cut into each other. The first and last stations should be small to close
// the ends.
func (s *Scene) addLoft(origin, fwd, up Vec3, st []loftStation, col color.RGBA) {
	f, u, r := frame3(fwd, up)
	seg := [3]int{8, 12, 16}[s.lod(origin)]
	pts := make([][]Vec3, len(st))
	nrm := make([][]Vec3, len(st))
	for k, c := range st {
		// how fast the section grows along the axis tilts the normals
		lo, hi := st[max(k-1, 0)], st[min(k+1, len(st)-1)]
		dw, dh, dy := 0.0, 0.0, 0.0
		if dx := hi.x - lo.x; dx != 0 {
			dw, dh, dy = (hi.hw-lo.hw)/dx, (hi.hh-lo.hh)/dx, (hi.y-lo.y)/dx
		}
		pts[k], nrm[k] = make([]Vec3, seg), make([]Vec3, seg)
		for i := 0; i < seg; i++ {
			th := 2 * math.Pi * float64(i) / float64(seg)
			ct, st := math.Cos(th), math.Sin(th)
			pts[k][i] = origin.Add(f.Scale(c.x)).Add(u.Scale(c.y + c.hh*st)).Add(r.Scale(c.hw * ct))
			radial := r.Scale(ct / c.hw).Add(u.Scale(st / c.hh)).Norm()
			slope := -(dw*ct*ct + dh*st*st + dy*st)
			nrm[k][i] = radial.Add(f.Scale(slope)).Norm()
		}
	}
	s.grid(pts, nrm, col)
}
