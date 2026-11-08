package main

import (
	"math"
)

// The conveyor belt runs left to right across the room. Candies come down a
// wobbly pipe at its left end and out of the machine's chute, ride along past
// the robots, and tumble off the right end into a basket.

const (
	beltX0    = 60.0
	beltX1    = 1170.0
	beltY     = 512.0 // top of the belt
	beltSpeed = 55.0
	pipeX     = 96.0
)

type beltCandy struct {
	x     float64
	k     col
	kind  int
	rot   float64
	taken bool // a robot has it in hand
	hop   float64
}

type Belt struct {
	offset   float64
	candies  []*beltCandy
	bulge    float64 // position of a candy travelling down the pipe (0..1), <0 none
	nextPipe float64
	basket   []beltCandy // the pile in the basket
}

func (b *Belt) init() {
	b.bulge, b.nextPipe = -1, 0.5
	for x := 140.0; x < beltX1-60; x += 150 {
		b.candies = append(b.candies, &beltCandy{x: x, k: candyColors[int(x)%len(candyColors)], kind: int(x/150) % 4})
	}
}

func (b *Belt) drop(x float64, k col, kind int) {
	b.candies = append(b.candies, &beltCandy{x: x, k: k, kind: kind, hop: 1})
}

func (b *Belt) update(g *Game) {
	b.offset += beltSpeed * dt
	if b.bulge >= 0 {
		b.bulge += dt / 0.9
		if b.bulge >= 1 {
			b.bulge = -1
			b.drop(pipeX+30, g.candyColor(), g.rng.Intn(4))
		}
	} else if b.nextPipe -= dt; b.nextPipe <= 0 {
		b.bulge, b.nextPipe = 0, g.rnd(1.8, 3.2)
	}
	live := b.candies[:0]
	for _, c := range b.candies {
		if c.taken {
			continue
		}
		c.x += beltSpeed * dt
		c.hop = math.Max(0, c.hop-dt*3)
		if c.x > beltX1-6 {
			// off the end: tumble into the basket
			bit := g.add(&Bit{kind: bitCandy, layer: layerMid, x: c.x, y: beltY - 12, vx: 90, vy: -120,
				vr: 6, size: 15, k: c.k, candy: c.kind, life: 0.55})
			_ = bit
			if len(b.basket) < 14 {
				b.basket = append(b.basket, beltCandy{x: g.rnd(-24, 24), k: c.k, kind: c.kind, rot: g.rnd(-1, 1)})
			} else {
				i := g.rng.Intn(len(b.basket))
				b.basket[i].k, b.basket[i].kind = c.k, c.kind
			}
			continue
		}
		live = append(live, c)
	}
	b.candies = live
}

// nearest finds an untaken candy within reach of x.
func (b *Belt) nearest(x, reach float64) *beltCandy {
	var best *beltCandy
	for _, c := range b.candies {
		if !c.taken && math.Abs(c.x-x) < reach && (best == nil || math.Abs(c.x-x) < math.Abs(best.x-x)) {
			best = c
		}
	}
	return best
}

func (b *Belt) draw(g *Game) {
	// legs
	for x := beltX0 + 40; x < beltX1; x += 220 {
		lineI(x, beltY+30, x-12, feetY+6, 10, c3(150, 150, 170))
		lineI(x+20, beltY+30, x+32, feetY+6, 10, c3(150, 150, 170))
	}
	// the pipe from the ceiling, bulging as a candy squeezes down it
	py0, py1 := 40.0, beltY-70
	lineI(pipeX, py0, pipeX, py1, 34, c3(120, 200, 230))
	curveI(pipeX, py1, pipeX, py1+40, pipeX+34, py1+44, 34, c3(120, 200, 230))
	fillLine(pipeX-8, py0, pipeX-8, py1, 6, fade(white, 0.6))
	if b.bulge >= 0 {
		y := lerp(py0+20, py1+30, b.bulge)
		x := pipeX
		if b.bulge > 0.85 {
			x += (b.bulge - 0.85) / 0.15 * 26
		}
		w := 1 + 0.15*math.Sin(b.bulge*40)
		ellipseI(x, y, 24*w, 28/w, c3(120, 200, 230))
		fillEllipse(x-8, y-6, 6, 9, fade(white, 0.6))
	}
	roundI(pipeX+30, py1+26, 16, 36, 4, c3(90, 170, 210))

	// belt body
	roundI(beltX0, beltY, beltX1-beltX0, 34, 17, c3(90, 86, 104))
	fillRound(beltX0+4, beltY+2, beltX1-beltX0-8, 8, 4, c3(130, 126, 146))
	// moving slats
	for x := beltX0 + 20 + math.Mod(b.offset, 40); x < beltX1-16; x += 40 {
		fillRect(x, beltY+4, 4, 26, fade(ink, 0.45))
	}
	// rollers that turn
	for _, x := range []float64{beltX0 + 17, beltX1 - 17} {
		circleI(x, beltY+17, 13, c3(200, 200, 215))
		a := b.offset / 13
		fillLine(x-math.Cos(a)*10, beltY+17-math.Sin(a)*10, x+math.Cos(a)*10, beltY+17+math.Sin(a)*10, 3, ink)
	}
	// candies riding
	for _, c := range b.candies {
		h := math.Sin(c.hop*math.Pi) * 14
		drawCandy(c.x, beltY-12-h, c.rot+math.Sin(g.t*3+c.x)*0.05, 15, c.k, c.kind)
	}

	// the basket at the end
	bx, by := 1222.0, feetY-58
	for i, c := range b.basket {
		row := float64(i / 5)
		drawCandy(bx+c.x, by-row*12+4, c.rot, 13, c.k, c.kind)
	}
	roundI(bx-46, by, 92, 66, 12, c3(214, 160, 90))
	for y := by + 12; y < by+60; y += 14 {
		fillLine(bx-40, y, bx+40, y, 3, c3(180, 125, 65))
	}
	roundI(bx-52, by-6, 104, 14, 7, c3(230, 180, 110))
}
