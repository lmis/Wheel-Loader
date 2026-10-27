package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// coverTitle is the name on the cover art.
const coverTitle = "Léo & Louis Onsite"

// stageCover sets up one still for the cover art: both kids out front with
// their animal friends, the loaders behind them, a skater gliding past on
// each side, a tower crane off to one side, the trucks and excavators at
// work in the back, the train in the distance and the helicopter overhead.
// The simulation is then held still so everything stays exactly where it
// was put.
func (g *Game) stageCover(path string) {
	g.scripted = true
	g.script = append(g.script, scriptStep{at: 3, fn: func(g *Game) {
		g.frozen, g.title = true, coverTitle
		g.world.Particles = nil
		// just the one crane, off to the right
		var vs []Vehicle
		for _, v := range g.vehicles {
			if v != g.crane2 {
				vs = append(vs, v)
			}
		}
		g.vehicles = vs
		g.crane.Slew = math.Pi * 0.78 // jib reaching out over the site

		// the kids stand out front, facing the camera, their loaders behind
		facing := math.Pi / 2
		for i, p := range g.players {
			if p.Driving != nil {
				p.Driving.SetOccupied(false)
			}
			p.Driving, p.td.goTo = nil, nil
			x := 1600 + float64(2*i-1)*34
			p.W.Pos, p.W.Yaw = V(x, 1450), facing-float64(2*i-1)*0.25
			p.W.moving, p.W.phase = 0, 0
		}
		// the animals gather round their kid: the owl flying between them, the chimp at
		// player 1's side, the dogs flanking player 2
		for _, a := range g.animals {
			k := g.players[0].W.Pos
			if a.owner == g.players[1] {
				k = g.players[1].W.Pos
			}
			a.ride, a.hopping, a.speed, a.bounce = nil, false, 0, 0
			switch a.Kind {
			case animalOwl:
				a.flying, a.Pos, a.Y, a.Yaw, a.t = true, V(1600, 1440), 40, facing, 0.17
			case animalChimp:
				a.Pos, a.Y, a.Yaw, a.t = k.Add(V(-24, 14)), 0, facing+0.4, 3
			case animalGolden:
				a.Pos, a.Y, a.Yaw = k.Add(V(28, 12)), 0, facing-0.5
			case animalBernese:
				a.Pos, a.Y, a.Yaw = k.Add(V(70, -26)), 0, facing-1.1
			}
		}
		// the skaters glide past on either side
		sk := &g.wonder.skaters
		sk[0].Pos, sk[0].Yaw, sk[0].phase, sk[0].lean, sk[0].look = V(1500, 1478), math.Pi-0.25, math.Pi/2, 0.15, 0
		sk[1].Pos, sk[1].Yaw, sk[1].phase, sk[1].lean, sk[1].look = V(1705, 1484), 0.3, 3*math.Pi/2, -0.2, 0
		for i := range sk {
			sk[i].speed = skaterSpeed
		}
		// the helicopter overhead, the train in the distance
		g.wonder.heli.t = 12.9
		g.wonder.train.X, g.wonder.train.wait = 1950, 0

		g.tour = []tourStop{{frame: g.frame, pos: V(1600, 1370), zoom: 250, yaw: math.Pi/2 + 0.06, pitch: 0.2, look: 34, shot: path}}
	}})
	g.exitFrame = 8
}

// drawTitle writes the cover title across the bottom of the frame (the
// helicopter has the sky) in chunky loader yellow with a dark outline. Each
// lowercase i is a traffic cone and each lowercase o a wheel.
func (g *Game) drawTitle(title string) {
	f := g.art.Title
	size := float64(f.size)
	gap := size * 0.03
	type piece struct {
		text string
		kind rune // 0: text, 'i': a cone, 'o': a wheel
		x, w float64
	}
	var pieces []piece
	run := ""
	flush := func() {
		if run != "" {
			pieces = append(pieces, piece{text: run, w: textWidth(run, f)})
			run = ""
		}
	}
	for _, r := range title {
		switch r {
		case 'i':
			flush()
			pieces = append(pieces, piece{kind: 'i', w: size*0.5 + 2*gap})
		case 'o':
			flush()
			pieces = append(pieces, piece{kind: 'o', w: textWidth("o", f) + 2*gap})
		default:
			run += string(r)
		}
	}
	flush()
	total := 0.0
	for _, p := range pieces {
		total += p.w
	}
	x := (screenW - total) / 2
	for i := range pieces {
		pieces[i].x = x
		x += pieces[i].w
	}
	y := screenH - size - 40
	ink := rgba(46, 30, 16, 255)
	text := func(dx, dy float64, c color.RGBA) {
		for _, p := range pieces {
			if p.kind == 0 {
				g.drawText(canvas{}, p.text, f, p.x+dx, y+dy, c)
			}
		}
	}
	shapes := func(grow, drop float64) {
		for _, p := range pieces {
			switch p.kind {
			case 'i':
				drawCone(p.x+gap, y, p.w-2*gap, size, grow, drop, ink)
			case 'o':
				drawWheel(p.x+gap, y, p.w-2*gap, size, grow, drop, ink)
			}
		}
	}
	for a := 0.0; a < 2*math.Pi; a += math.Pi / 8 {
		text(math.Cos(a)*5, math.Sin(a)*5+3, ink)
	}
	shapes(5, 3)
	text(0, 0, colLoader)
	// a lighter band across the top of the letters
	rl.BeginScissorMode(0, int32(y+10), screenW, int32(size*0.32))
	text(0, 0, rgba(255, 222, 96, 255))
	rl.EndScissorMode()
	shapes(0, 0)
}

// drawWheel draws a chunky tyre the size of a lowercase o in a line of text
// of the given size whose glyph box starts at (x, y) and is w wide. With
// grow > 0 it draws just the outline, as drawCone does.
func drawWheel(x, y, w, size, grow, drop float64, ink color.RGBA) {
	base := y + size*0.80 // the text baseline
	r := w * 0.46
	c := rl.NewVector2(float32(x+w/2), float32(base-r+drop))
	if grow > 0 {
		rl.DrawCircleV(c, float32(r+grow), rlc(ink))
		return
	}
	tyre := rgba(42, 42, 48, 255)
	rl.DrawCircleV(c, float32(r), rlc(tyre))
	// chunky tread blocks round the rim
	for k := 0; k < 12; k++ {
		a := float64(k) * math.Pi / 6
		p := rl.NewVector2(c.X+float32(math.Cos(a)*r*0.9), c.Y+float32(math.Sin(a)*r*0.9))
		rl.DrawPoly(p, 4, float32(r*0.2), float32(a*180/math.Pi+45), rlc(tyre))
	}
	rl.DrawCircleV(c, float32(r*0.74), rlc(rgba(66, 66, 74, 255))) // sidewall
	rl.DrawCircleV(c, float32(r*0.52), rlc(colLoader))             // the rim
	rl.DrawCircleV(c, float32(r*0.4), rlc(rgba(214, 160, 18, 255)))
	for k := 0; k < 5; k++ { // wheel nuts
		a := float64(k)*2*math.Pi/5 - math.Pi/2
		rl.DrawCircleV(rl.NewVector2(c.X+float32(math.Cos(a)*r*0.26), c.Y+float32(math.Sin(a)*r*0.26)), float32(r*0.06), rlc(rgba(90, 70, 20, 255)))
	}
	rl.DrawCircleV(c, float32(r*0.12), rlc(rgba(90, 70, 20, 255))) // hub cap
	// a highlight across the top of the tyre
	rl.DrawRing(c, float32(r*0.78), float32(r*0.92), 200, 250, 16, rlc(rgba(110, 110, 120, 200)))
}

// drawCone draws a traffic cone as tall as a t in a line of
// text of the given size whose glyph box starts at (x, y) and is w wide. With
// grow > 0 it draws just the outline: the shape grown by that much in ink,
// dropped by drop.
func drawCone(x, y, w, size, grow, drop float64, ink color.RGBA) {
	base := y + size*0.80 // the text baseline
	top := y + size*0.15  // level with the t
	foot := size * 0.08
	cx := x + w/2
	// trap fills a trapezoid between heights yt and yb, wt and wb wide there,
	// centred on cx+dx
	trap := func(yt, yb, wt, wb, dx float64, c color.RGBA) {
		tl := rl.NewVector2(float32(cx+dx-wt/2-grow), float32(yt-grow+drop))
		tr := rl.NewVector2(float32(cx+dx+wt/2+grow), float32(yt-grow+drop))
		bl := rl.NewVector2(float32(cx+dx-wb/2-grow), float32(yb+grow+drop))
		br := rl.NewVector2(float32(cx+dx+wb/2+grow), float32(yb+grow+drop))
		rl.DrawTriangle(tl, bl, br, rlc(c))
		rl.DrawTriangle(tl, br, tr, rlc(c))
	}
	wb, wt := w*0.78, w*0.2 // the body's width at its foot and its tip
	yb := base - foot
	if grow > 0 {
		trap(yb, base, w, w, 0, ink)
		trap(top, yb, wt, wb, 0, ink)
		return
	}
	trap(yb, base, w, w, 0, rgba(60, 60, 66, 255)) // the square foot
	trap(top, yb, wt, wb, 0, rgba(255, 112, 24, 255))
	// two white reflective bands
	width := func(t float64) float64 { return wb + (wt-wb)*t } // t: 0 foot, 1 tip
	height := func(t float64) float64 { return yb - (yb-top)*t }
	for _, b := range [][2]float64{{0.2, 0.36}, {0.54, 0.68}} {
		trap(height(b[1]), height(b[0]), width(b[1]), width(b[0]), 0, rgba(250, 248, 240, 255))
	}
	// sunlight down the left side
	trap(top, yb, wt*0.3, wb*0.3, -(wt+wb)/4*0.55, rgba(255, 160, 90, 140))
}
