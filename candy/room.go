package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The room: striped wallpaper, two windows onto a sunny day, bunting, swinging
// lamps, a ticking clock, a wooden floor and a mouse who peeks out of a hole.

type cloud struct{ x, y, s, speed float64 }

type Room struct {
	clouds []cloud
	mouse  float64 // counts down to the next peek; negative while peeking
	bird   float64 // x of a bird flying past the left window, <0 when none
	birdY  float64
	birdIn float64
}

const (
	winY, winH = 70.0, 210.0
	win1X      = 150.0
	win2X      = 930.0
	winW       = 210.0
	mouseX     = 420.0
)

func (r *Room) init(rng *rand.Rand) {
	for i := 0; i < 7; i++ {
		r.clouds = append(r.clouds, cloud{x: rng.Float64() * W, y: winY + 20 + rng.Float64()*90,
			s: 0.6 + rng.Float64()*0.6, speed: 6 + rng.Float64()*8})
	}
	r.mouse = 6
	r.bird, r.birdIn = -1, 4
}

func (r *Room) update(g *Game) {
	for i := range r.clouds {
		c := &r.clouds[i]
		c.x += c.speed * dt
		if c.x > W+80 {
			c.x, c.y = -80, winY+20+g.rnd(0, 90)
		}
	}
	r.mouse -= dt
	if r.mouse < -4.5 {
		r.mouse = g.rnd(7, 14)
	}
	if r.bird >= 0 {
		r.bird += 70 * dt
		if r.bird > win1X+winW+40 {
			r.bird, r.birdIn = -1, g.rnd(6, 12)
		}
	} else if r.birdIn -= dt; r.birdIn <= 0 {
		r.bird, r.birdY = win1X-40, winY+40+g.rnd(0, 70)
	}
}

func (r *Room) drawBack(g *Game) {
	t := g.t
	// wallpaper
	rl.DrawRectangleGradientV(0, 0, W, int32(floorY), c3(255, 226, 205), c3(250, 205, 190))
	for x := 0.0; x < W; x += 64 {
		fillRect(x, 0, 30, floorY, fade(c3(255, 190, 185), 0.35))
		for y := 40.0; y < floorY-60; y += 80 {
			off := 0.0
			if int(x/64)%2 == 1 {
				off = 40
			}
			star(x+15, y+off, 6, 0, 4, fade(white, 0.55))
		}
	}
	// wainscot and skirting
	fillRect(0, floorY-70, W, 70, c3(150, 205, 190))
	for x := 10.0; x < W; x += 80 {
		roundI(x, floorY-60, 60, 48, 6, c3(165, 218, 202))
	}
	fillRect(0, floorY-74, W, 8, ink)
	fillRect(0, floorY-73, W, 5, c3(120, 180, 165))

	// the mouse hole, and the mouse
	fillEllipse(mouseX, floorY, 30, 38, ink)
	if r.mouse < 0 {
		p := -r.mouse
		out := ease(p/0.4) * ease((4.5-p)/0.4)
		look := math.Sin(p*2.2) * 6
		hx := mouseX + look
		hy := floorY - 6 - out*20
		circleI(hx-12, hy-14, 9, c3(190, 180, 185))
		circleI(hx+12, hy-14, 9, c3(190, 180, 185))
		fillCircle(hx-12, hy-14, 5, c3(255, 175, 190))
		fillCircle(hx+12, hy-14, 5, c3(255, 175, 190))
		ellipseI(hx, hy, 15, 13, c3(190, 180, 185))
		fillCircle(hx-5+look*0.3, hy-2, 2.6, ink)
		fillCircle(hx+5+look*0.3, hy-2, 2.6, ink)
		circleI(hx+look*0.5, hy+6, 3, c3(255, 140, 160))
		w := math.Sin(t*30) * 2
		fillLine(hx+4, hy+6, hx+22, hy+3+w, 1.2, ink)
		fillLine(hx-4, hy+6, hx-22, hy+3-w, 1.2, ink)
	}
	fillRect(mouseX-36, floorY, 72, 10, c3(150, 100, 70))

	drawFloor()
	r.drawWindow(g, win1X, true)
	r.drawWindow(g, win2X, false)
	r.drawClock(g, 1205, 150)

	// ceiling beam and pipes
	fillRect(0, 0, W, 26, c3(160, 105, 80))
	fillRect(0, 26, W, 4, ink)
	lineI(0, 44, W, 44, 12, c3(200, 200, 215))
	for x := 120.0; x < W; x += 260 {
		roundI(x, 36, 18, 16, 3, c3(170, 170, 190))
		// a valve wheel that turns now and then
		gear(x+9, 66, 9, t*0.7+x, 6, c3(240, 90, 90))
	}

	// bunting
	for side := 0; side < 2; side++ {
		x0 := float64(side) * W / 2
		x1 := x0 + W/2
		sag := 46.0
		n := 9
		px, py := x0, 34.0
		for i := 0; i <= n*2; i++ {
			u := float64(i) / float64(n*2)
			x := lerp(x0, x1, u)
			y := 34 + sag*4*u*(1-u)
			fillLine(px, py, x, y, 2, ink)
			px, py = x, y
		}
		for i := 1; i < n; i++ {
			u := float64(i) / float64(n)
			x := lerp(x0, x1, u)
			y := 34 + sag*4*u*(1-u)
			sw := math.Sin(t*2.2+float64(i)*0.9+float64(side)) * 7
			triI(x-17, y, x+17, y, x+sw, y+38, candyColors[(i+side*3)%len(candyColors)])
		}
	}
}

func (r *Room) drawWindow(g *Game, x float64, sunny bool) {
	t := g.t
	roundI(x-12, winY-12, winW+24, winH+24, 24, c3(255, 248, 235))
	rl.BeginScissorMode(int32(x), int32(winY), int32(winW), int32(winH))
	rl.DrawRectangleGradientV(int32(x), int32(winY), int32(winW), int32(winH), c3(120, 195, 250), c3(200, 235, 255))
	if sunny {
		sx, sy := x+150, winY+60
		for i := 0; i < 12; i++ {
			a := t*0.3 + float64(i)/12*2*math.Pi
			fillLine(sx+math.Cos(a)*42, sy+math.Sin(a)*42, sx+math.Cos(a)*(58+4*math.Sin(t*3+float64(i))), sy+math.Sin(a)*(58+4*math.Sin(t*3+float64(i))), 6, c3(255, 210, 70))
		}
		circleI(sx, sy, 34, c3(255, 220, 90))
		// a sleepy smiling sun
		fillLine(sx-14, sy-4, sx-6, sy-4, 3, ink)
		fillLine(sx+6, sy-4, sx+14, sy-4, 3, ink)
		rl.DrawRing(v2(sx, sy+4), 10, 13, 20, 160, 12, ink)
	} else {
		// hills and a windmill
		fillEllipse(x+60, winY+winH+40, 150, 90, c3(130, 205, 110))
		fillEllipse(x+190, winY+winH+60, 140, 100, c3(110, 190, 100))
		mx, my := x+150, winY+120
		triI(mx-14, my+60, mx+14, my+60, mx, my-4, c3(250, 245, 235))
		for i := 0; i < 4; i++ {
			a := t*1.2 + float64(i)*math.Pi/2
			lineI(mx, my, mx+math.Cos(a)*44, my+math.Sin(a)*44, 7, c3(240, 120, 120))
		}
		circleI(mx, my, 5, c3(250, 245, 235))
	}
	for _, c := range r.clouds {
		cx := c.x
		if !sunny {
			cx = math.Mod(c.x+400, W+160) - 80
		}
		if cx > x-90 && cx < x+winW+90 {
			drawCloud(cx, c.y, c.s)
		}
	}
	if sunny && r.bird >= 0 {
		f := math.Sin(t*12) * 8
		bx, by := r.bird, r.birdY+math.Sin(t*2)*6
		fillLine(bx-12, by-f, bx, by, 3, ink)
		fillLine(bx, by, bx+12, by-f, 3, ink)
	}
	rl.EndScissorMode()
	// window frame: crossbars and sill
	fillRect(x+winW/2-5, winY, 10, winH, c3(255, 248, 235))
	fillRect(x, winY+winH/2-5, winW, 10, c3(255, 248, 235))
	roundI(x-26, winY+winH+8, winW+52, 18, 6, c3(255, 248, 235))
	// a flower pot on the left sill
	if sunny {
		px := x + 30
		sway := math.Sin(t*1.5) * 4
		lineI(px, winY+winH+2, px+sway, winY+winH-40, 3, c3(90, 170, 80))
		for i := 0; i < 5; i++ {
			a := float64(i)/5*2*math.Pi + t*0.2
			circleI(px+sway+math.Cos(a)*9, winY+winH-42+math.Sin(a)*9, 7, c3(255, 120, 160))
		}
		circleI(px+sway, winY+winH-42, 5, c3(255, 215, 80))
		roundI(px-14, winY+winH-6, 28, 22, 4, c3(220, 120, 80))
	}
}

func drawCloud(x, y, s float64) {
	for _, p := range [][3]float64{{-26, 6, 18}, {0, 0, 24}, {26, 6, 18}, {12, 12, 16}, {-12, 12, 16}} {
		fillCircle(x+p[0]*s, y+p[1]*s, (p[2]+inkW)*s, fade(ink, 0.25))
	}
	for _, p := range [][3]float64{{-26, 6, 18}, {0, 0, 24}, {26, 6, 18}, {12, 12, 16}, {-12, 12, 16}} {
		fillCircle(x+p[0]*s, y+p[1]*s, p[2]*s, white)
	}
}

func (r *Room) drawClock(g *Game, x, y float64) {
	sway := math.Sin(g.t*2.4) * 0.4
	// pendulum
	push(x, y+30, sway*57.3/2, 1, 1)
	lineI(0, 0, 0, 70, 3, c3(220, 170, 60))
	circleI(0, 74, 11, c3(240, 195, 70))
	pop()
	roundI(x-38, y-44, 76, 84, 18, c3(170, 110, 80))
	circleI(x, y, 30, white)
	for i := 0; i < 12; i++ {
		a := float64(i) / 12 * 2 * math.Pi
		fillCircle(x+math.Cos(a)*24, y+math.Sin(a)*24, 2, ink)
	}
	sec := math.Floor(g.t) / 60 * 2 * math.Pi
	min := g.t / 600 * 2 * math.Pi
	fillLine(x, y, x+math.Sin(min)*14, y-math.Cos(min)*14, 4, ink)
	fillLine(x, y, x+math.Sin(sec)*22, y-math.Cos(sec)*22, 2, c3(230, 70, 80))
	fillCircle(x, y, 4, ink)
}

// drawFloor runs before the belt: the planks.
func drawFloor() {
	rl.DrawRectangleGradientV(0, int32(floorY), W, H-int32(floorY), c3(214, 150, 100), c3(186, 120, 80))
	for y := floorY + 26.0; y < H; y += 34 {
		fillRect(0, y, W, 3, fade(ink, 0.35))
	}
	row := 0
	for y := floorY; y < H; y += 34 {
		off := float64(row%2) * 90
		for x := off; x < W; x += 180 {
			fillRect(x, y, 3, 26, fade(ink, 0.3))
			fillCircle(x+40, y+12, 2, fade(ink, 0.25))
		}
		row++
	}
}

// drawLights are the two swinging lamps and their warm glow, drawn last.
func (r *Room) drawLights(g *Game) {
	for i, x := range []float64{420, 860} {
		a := math.Sin(g.t*1.3+float64(i)*1.7) * 0.07
		lx, ly := x+math.Sin(a)*120, 30+math.Cos(a)*120
		fillLine(x, 30, lx, ly, 3, ink)
		// glow
		push(lx, ly, a*57.3, 1, 1)
		col := fade(c3(255, 240, 180), 0.12)
		if g.disco > 0 {
			col = fade(candyColors[(int(g.t*6)+i*3)%len(candyColors)], 0.22)
		}
		fillTri(-20, 20, 20, 20, 170, 520, col)
		fillTri(-20, 20, -170, 520, 170, 520, col)
		triI(-28, 22, 28, 22, 0, -12, c3(240, 110, 110))
		circleI(0, 24, 9, c3(255, 245, 190))
		pop()
	}
	if g.disco > 0 {
		for i := 0; i < 9; i++ {
			a := g.t*1.5 + float64(i)*0.7
			x := 640 + math.Cos(a)*520*math.Sin(float64(i)*1.3+g.t*0.4)
			y := 280 + math.Sin(a*1.3)*200
			fillCircle(x, y, 40, fade(candyColors[i%len(candyColors)], 0.25*math.Min(1, g.disco)))
		}
	}
}
