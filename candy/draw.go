package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Storybook drawing: flat colour shapes with a soft brown ink outline, drawn
// in local coordinates under push/pop transforms so characters can squash,
// stretch, flip and spin.

type col = color.RGBA

func c3(r, g, b uint8) col { return col{r, g, b, 255} }

var (
	ink   = c3(70, 42, 48)
	white = c3(255, 252, 246)
)

const inkW = 3.0

func fade(k col, a float64) col {
	k.A = uint8(clamp(a, 0, 1) * float64(k.A))
	return k
}

func mix(a, b col, t float64) col {
	l := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*clamp(t, 0, 1)) }
	return col{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), l(a.A, b.A)}
}

func clamp(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }

func v2(x, y float64) rl.Vector2 { return rl.NewVector2(float32(x), float32(y)) }

func push(x, y, rotDeg, sx, sy float64) {
	rl.PushMatrix()
	rl.Translatef(float32(x), float32(y), 0)
	if rotDeg != 0 {
		rl.Rotatef(float32(rotDeg), 0, 0, 1)
	}
	rl.Scalef(float32(sx), float32(sy), 1)
}

func pop() { rl.PopMatrix() }

// ---- plain fills ----

func fillCircle(x, y, r float64, k col) {
	if r > 0 {
		rl.DrawCircleSector(v2(x, y), float32(r), 0, 360, 36, k)
	}
}

func fillEllipse(x, y, rx, ry float64, k col) {
	const n = 36
	pts := make([]rl.Vector2, 0, n+2)
	pts = append(pts, v2(x, y))
	for i := 0; i <= n; i++ {
		a := -float64(i) / n * 2 * math.Pi
		pts = append(pts, v2(x+rx*math.Cos(a), y+ry*math.Sin(a)))
	}
	rl.DrawTriangleFan(pts, k)
}

func fillRect(x, y, w, h float64, k col) {
	rl.DrawRectangleRec(rl.NewRectangle(float32(x), float32(y), float32(w), float32(h)), k)
}

// fillRound is a rectangle with corners of radius r pixels.
func fillRound(x, y, w, h, r float64, k col) {
	m := math.Min(w, h)
	if m <= 0 {
		return
	}
	rl.DrawRectangleRounded(rl.NewRectangle(float32(x), float32(y), float32(w), float32(h)),
		float32(clamp(2*r/m, 0, 1)), 10, k)
}

func fillTri(ax, ay, bx, by, cx, cy float64, k col) {
	rl.DrawTriangle(v2(ax, ay), v2(bx, by), v2(cx, cy), k)
}

func fillLine(x1, y1, x2, y2, w float64, k col) {
	rl.DrawLineEx(v2(x1, y1), v2(x2, y2), float32(w), k)
	fillCircle(x1, y1, w/2, k)
	fillCircle(x2, y2, w/2, k)
}

// ---- inked shapes: outline, then fill ----

func circleI(x, y, r float64, k col) {
	fillCircle(x, y, r+inkW, ink)
	fillCircle(x, y, r, k)
}

func ellipseI(x, y, rx, ry float64, k col) {
	fillEllipse(x, y, rx+inkW, ry+inkW, ink)
	fillEllipse(x, y, rx, ry, k)
}

func roundI(x, y, w, h, r float64, k col) {
	fillRound(x-inkW, y-inkW, w+2*inkW, h+2*inkW, r+inkW, ink)
	fillRound(x, y, w, h, r, k)
}

func lineI(x1, y1, x2, y2, w float64, k col) {
	fillLine(x1, y1, x2, y2, w+2*inkW, ink)
	fillLine(x1, y1, x2, y2, w, k)
}

func triI(ax, ay, bx, by, cx, cy float64, k col) {
	mx, my := (ax+bx+cx)/3, (ay+by+cy)/3
	grow := func(x, y float64) (float64, float64) {
		d := math.Hypot(x-mx, y-my)
		if d == 0 {
			return x, y
		}
		s := (d + inkW*1.8) / d
		return mx + (x-mx)*s, my + (y-my)*s
	}
	a1, a2 := grow(ax, ay)
	b1, b2 := grow(bx, by)
	c1, c2 := grow(cx, cy)
	fillTri(a1, a2, b1, b2, c1, c2, ink)
	fillTri(ax, ay, bx, by, cx, cy, k)
}

// curveI is a thick quadratic curve (a noodle arm, a tail, a string).
func curveI(x1, y1, cx, cy, x2, y2, w float64, k col) {
	curve(x1, y1, cx, cy, x2, y2, w+2*inkW, ink)
	curve(x1, y1, cx, cy, x2, y2, w, k)
}

func curve(x1, y1, cx, cy, x2, y2, w float64, k col) {
	const n = 14
	px, py := x1, y1
	for i := 1; i <= n; i++ {
		t := float64(i) / n
		u := 1 - t
		x := u*u*x1 + 2*u*t*cx + t*t*x2
		y := u*u*y1 + 2*u*t*cy + t*t*y2
		rl.DrawLineEx(v2(px, py), v2(x, y), float32(w), k)
		fillCircle(x, y, w/2, k)
		px, py = x, y
	}
	fillCircle(x1, y1, w/2, k)
}

// shine is the little white highlight that makes things look round and sweet.
func shine(x, y, r float64) {
	fillEllipse(x, y, r, r*0.6, fade(white, 0.75))
}

// star draws a puffy n-pointed star.
func star(x, y, r, rot float64, points int, k col) {
	n := points * 2
	pts := []rl.Vector2{v2(x, y)}
	for i := 0; i <= n; i++ {
		a := rot - float64(i)/float64(n)*2*math.Pi
		rr := r
		if i%2 == 1 {
			rr = r * 0.5
		}
		pts = append(pts, v2(x+rr*math.Cos(a), y+rr*math.Sin(a)))
	}
	rl.DrawTriangleFan(pts, k)
}

func starI(x, y, r, rot float64, points int, k col) {
	star(x, y, r+inkW*1.8, rot, points, ink)
	star(x, y, r, rot, points, k)
}

func heart(x, y, s float64, k col) {
	circleI(x-s*0.5, y-s*0.2, s*0.55, k)
	circleI(x+s*0.5, y-s*0.2, s*0.55, k)
	triI(x-s*1.02, y, x+s*1.02, y, x, y+s*1.15, k)
	fillCircle(x-s*0.5, y-s*0.2, s*0.55, k)
	fillCircle(x+s*0.5, y-s*0.2, s*0.55, k)
	shine(x-s*0.6, y-s*0.4, s*0.2)
}

// gear is a toothed wheel turned by rot radians.
func gear(x, y, r, rot float64, teeth int, k col) {
	for pass := 0; pass < 2; pass++ {
		kk, g := ink, inkW
		if pass == 1 {
			kk, g = k, 0
		}
		for i := 0; i < teeth; i++ {
			a := rot + float64(i)/float64(teeth)*2*math.Pi
			push(x+math.Cos(a)*r, y+math.Sin(a)*r, a*180/math.Pi, 1, 1)
			fillRound(-r*0.18-g, -r*0.16-g, r*0.36+2*g, r*0.32+2*g, 3, kk)
			pop()
		}
		fillCircle(x, y, r+g, kk)
	}
	fillCircle(x, y, r*0.62, mix(k, ink, 0.25))
	fillCircle(x, y, r*0.5, k)
	circleI(x, y, r*0.18, mix(k, white, 0.4))
}

// ---- words ----

func (g *Game) word(s string, x, y, size, rot float64, k col) {
	if size < 2 {
		return
	}
	f := g.font
	m := rl.MeasureTextEx(f, s, float32(size), 0)
	o := v2(float64(m.X)/2, float64(m.Y)/2)
	for i := 0; i < 12; i++ {
		a := float64(i) / 12 * 2 * math.Pi
		d := size * 0.09
		rl.DrawTextPro(f, s, v2(x+math.Cos(a)*d, y+math.Sin(a)*d), o, float32(rot), float32(size), 0, ink)
	}
	rl.DrawTextPro(f, s, v2(x, y), o, float32(rot), float32(size), 0, k)
}

// ---- easing ----

func ease(t float64) float64 { t = clamp(t, 0, 1); return t * t * (3 - 2*t) }

// springOut overshoots and settles: 0 at t=0, wobbles about 1.
func springOut(t float64) float64 {
	if t <= 0 {
		return 0
	}
	return 1 - math.Exp(-6*t)*math.Cos(14*t)
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
