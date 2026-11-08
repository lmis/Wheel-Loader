package main

import (
	"math"
)

// Bits are the little things that fly about: candies, confetti, steam,
// sparkles, hearts, notes, bubbles, crumbs and dust.

type bitKind int

const (
	bitCandy bitKind = iota
	bitConfetti
	bitPuff
	bitSparkle
	bitHeart
	bitNote
	bitBubble
	bitCrumb
	bitDust
)

type layer int

const (
	layerBack layer = iota
	layerMid
	layerFront
)

type Bit struct {
	kind         bitKind
	layer        layer
	x, y, vx, vy float64
	rot, vr      float64
	size         float64
	k            col
	candy        int
	life, max    float64
	gravity      float64
	bounce       bool
	resting      bool
	onBelt       bool
}

var candyColors = []col{
	c3(255, 92, 120), c3(255, 170, 60), c3(250, 222, 80), c3(110, 210, 120),
	c3(90, 180, 245), c3(185, 120, 235), c3(255, 140, 200),
}

func (g *Game) candyColor() col { return candyColors[g.rng.Intn(len(candyColors))] }

func (g *Game) add(b *Bit) *Bit {
	if len(g.bits) > 900 {
		return b
	}
	if b.max == 0 {
		b.max = b.life
	}
	g.bits = append(g.bits, b)
	return b
}

func (g *Game) spawnCandy(x, y, vx, vy float64) *Bit {
	return g.add(&Bit{kind: bitCandy, layer: layerMid, x: x, y: y, vx: vx, vy: vy, vr: g.rnd(-8, 8),
		size: g.rnd(13, 18), k: g.candyColor(), candy: g.rng.Intn(4), life: 9, gravity: 900, bounce: true})
}

func (g *Game) confetti(x, y float64, n int, speed float64) {
	for i := 0; i < n; i++ {
		a := g.rnd(0, 2*math.Pi)
		s := g.rnd(0.3, 1) * speed
		g.add(&Bit{kind: bitConfetti, layer: layerFront, x: x, y: y, vx: math.Cos(a) * s, vy: math.Sin(a)*s - speed*0.3,
			rot: g.rnd(0, 6), vr: g.rnd(-12, 12), size: g.rnd(6, 11), k: g.candyColor(), life: g.rnd(1.6, 2.6), gravity: 380})
	}
}

func (g *Game) sparkles(x, y float64, n int, r float64) {
	for i := 0; i < n; i++ {
		a := g.rnd(0, 2*math.Pi)
		g.add(&Bit{kind: bitSparkle, layer: layerFront, x: x + math.Cos(a)*r*g.rnd(0, 1), y: y + math.Sin(a)*r*g.rnd(0, 1),
			vx: math.Cos(a) * 60, vy: math.Sin(a)*60 - 30, rot: g.rnd(0, 3), vr: g.rnd(-3, 3), size: g.rnd(7, 13),
			k: mix(g.candyColor(), white, 0.5), life: g.rnd(0.5, 0.9)})
	}
}

func (g *Game) hearts(x, y float64, n int) {
	for i := 0; i < n; i++ {
		g.add(&Bit{kind: bitHeart, layer: layerFront, x: x + g.rnd(-20, 20), y: y, vx: g.rnd(-25, 25), vy: g.rnd(-110, -70),
			rot: g.rnd(0, 6), size: g.rnd(8, 12), k: c3(255, 96, 130), life: g.rnd(1.2, 1.7)})
	}
}

func (g *Game) puff(x, y, size float64, k col, l layer) {
	g.add(&Bit{kind: bitPuff, layer: l, x: x, y: y, vx: g.rnd(-12, 12), vy: g.rnd(-60, -35), size: size,
		k: k, life: g.rnd(1.6, 2.4)})
}

func (g *Game) dust(x, y float64, n int) {
	for i := 0; i < n; i++ {
		s := float64(i%2*2 - 1)
		g.add(&Bit{kind: bitDust, layer: layerFront, x: x + s*g.rnd(0, 30), y: y, vx: s * g.rnd(60, 160), vy: g.rnd(-40, -10),
			size: g.rnd(10, 18), k: c3(236, 214, 190), life: g.rnd(0.5, 0.8)})
	}
}

func (g *Game) crumbs(x, y float64, k col, n int) {
	for i := 0; i < n; i++ {
		g.add(&Bit{kind: bitCrumb, layer: layerFront, x: x, y: y, vx: g.rnd(-90, 90), vy: g.rnd(-160, -40),
			size: g.rnd(2.5, 4.5), k: k, life: g.rnd(0.5, 0.9), gravity: 700})
	}
}

func (g *Game) updateBits() {
	live := g.bits[:0]
	for _, b := range g.bits {
		b.life -= dt
		if b.life <= 0 {
			if b.kind == bitBubble {
				g.sparkles(b.x, b.y, 3, b.size)
			}
			continue
		}
		if b.onBelt {
			continue // the belt carries it; dropped by Belt when it lands
		}
		b.vy += b.gravity * dt
		switch b.kind {
		case bitConfetti:
			b.vx *= 1 - 2*dt
			b.vy = math.Min(b.vy, 120)
			b.x += math.Sin(b.life*7+b.rot) * 40 * dt
		case bitPuff:
			b.size += 22 * dt
			b.vx *= 1 - dt
		case bitHeart, bitNote:
			b.x += math.Sin(b.life*5+b.rot) * 30 * dt
		case bitBubble:
			b.x += math.Sin(b.life*3+b.rot) * 40 * dt
			b.vy *= 1 - 0.5*dt
		case bitDust:
			b.vx *= 1 - 5*dt
			b.size += 25 * dt
		}
		b.x += b.vx * dt
		b.y += b.vy * dt
		b.rot += b.vr * dt
		if b.kind == bitCandy {
			g.candyPhysics(b)
		}
		live = append(live, b)
	}
	g.bits = live
}

// candyPhysics lets a flying candy land on the belt, bounce on the floor and
// come to rest there (for the dino to find).
func (g *Game) candyPhysics(b *Bit) {
	if b.resting {
		b.vx, b.vy, b.vr = 0, 0, 0
		return
	}
	if b.vy > 0 && b.x > beltX0+10 && b.x < beltX1-10 && b.y > beltY-6 && b.y < beltY+14 && len(g.belt.candies) < 30 {
		b.life = 0
		g.belt.drop(b.x, b.k, b.candy)
		return
	}
	floor := feetY + 8 - b.size*0.4
	if b.y > floor {
		b.y = floor
		if b.vy > 140 {
			b.vy *= -0.42
			b.vx *= 0.7
			b.vr *= 0.6
		} else {
			b.resting = true
			b.layer = layerMid
		}
	}
	if b.x < 20 || b.x > W-20 {
		b.vx = -b.vx * 0.6
		b.x = clamp(b.x, 20, W-20)
	}
}

func (g *Game) drawBits(l layer) {
	for _, b := range g.bits {
		if b.layer != l || b.onBelt {
			continue
		}
		a := clamp(b.life/0.5, 0, 1)
		switch b.kind {
		case bitCandy:
			if a < 1 && b.life < 0.5 {
				push(b.x, b.y, 0, a, a)
				drawCandy(0, 0, b.rot, b.size, b.k, b.candy)
				pop()
			} else {
				drawCandy(b.x, b.y, b.rot, b.size, b.k, b.candy)
			}
		case bitConfetti:
			push(b.x, b.y, b.rot*57, math.Cos(b.life*9), 1)
			fillRect(-b.size/2, -b.size/3, b.size, b.size*0.66, fade(b.k, a))
			pop()
		case bitPuff:
			t := b.life / b.max
			fillCircle(b.x, b.y, b.size, fade(b.k, t*0.85))
			fillCircle(b.x-b.size*0.3, b.y-b.size*0.3, b.size*0.4, fade(white, t*0.5))
		case bitSparkle:
			s := b.size * math.Sin(math.Pi*b.life/b.max)
			star(b.x, b.y, s, b.rot, 4, fade(b.k, 1))
			star(b.x, b.y, s*0.45, b.rot, 4, white)
		case bitHeart:
			s := b.size * math.Min(1, (b.max-b.life)*6)
			push(b.x, b.y, math.Sin(b.life*6)*12, a, a)
			heart(0, 0, s, b.k)
			pop()
		case bitNote:
			push(b.x, b.y, math.Sin(b.life*6)*15, a, a)
			ellipseI(0, 0, b.size, b.size*0.75, b.k)
			lineI(b.size*0.8, 0, b.size*0.8, -b.size*3, 3, b.k)
			lineI(b.size*0.8, -b.size*3, b.size*2, -b.size*2.3, 3, b.k)
			pop()
		case bitBubble:
			fillCircle(b.x, b.y, b.size+2, fade(ink, 0.5))
			fillCircle(b.x, b.y, b.size, fade(b.k, 0.45))
			fillEllipse(b.x-b.size*0.35, b.y-b.size*0.4, b.size*0.3, b.size*0.18, fade(white, 0.9))
		case bitCrumb:
			fillCircle(b.x, b.y, b.size, fade(b.k, a))
		case bitDust:
			fillCircle(b.x, b.y, b.size, fade(b.k, a*0.9))
		}
	}
}

// drawCandy draws one sweet: a wrapped toffee, a lollipop, a gumball or a star.
func drawCandy(x, y, rot, s float64, k col, kind int) {
	push(x, y, rot*57.3, 1, 1)
	switch kind {
	case 0: // wrapped sweet
		triI(-s*0.7, 0, -s*1.4, -s*0.6, -s*1.4, s*0.6, mix(k, white, 0.35))
		triI(s*0.7, 0, s*1.4, -s*0.6, s*1.4, s*0.6, mix(k, white, 0.35))
		ellipseI(0, 0, s*0.85, s*0.62, k)
		fillLine(-s*0.3, -s*0.5, s*0.1, s*0.5, s*0.18, mix(k, white, 0.5))
		shine(-s*0.35, -s*0.25, s*0.22)
	case 1: // lollipop
		lineI(0, s*0.6, 0, s*1.8, s*0.22, white)
		circleI(0, 0, s*0.8, k)
		for i := 0; i < 3; i++ {
			fillCircle(0, 0, s*(0.62-float64(i)*0.2), mix(k, white, float64(i%2)*0.6))
		}
		shine(-s*0.3, -s*0.35, s*0.2)
	case 2: // gumball
		circleI(0, 0, s*0.72, k)
		shine(-s*0.25, -s*0.3, s*0.24)
	case 3: // star candy
		starI(0, 0, s*0.95, -math.Pi/2, 5, k)
		shine(-s*0.2, -s*0.25, s*0.2)
	}
	pop()
}

// ---- words: BOING! RAWR! YUM! ----

type Word struct {
	s    string
	x, y float64
	size float64
	k    col
	life float64
	rot  float64
}

func (g *Game) say(s string, x, y, size float64, k col) {
	g.words = append(g.words, &Word{s: s, x: x, y: y, size: size, k: k, life: 1.4, rot: g.rnd(-12, 12)})
}

func (g *Game) updateWords() {
	live := g.words[:0]
	for _, w := range g.words {
		w.life -= dt
		w.y -= 18 * dt
		if w.life > 0 {
			live = append(live, w)
		}
	}
	g.words = live
}

func (g *Game) drawWords() {
	for _, w := range g.words {
		age := 1.4 - w.life
		s := springOut(age*1.4) * clamp(w.life/0.25, 0, 1)
		y := math.Max(w.y, w.size*0.6)
		g.word(w.s, w.x, y, w.size*s, w.rot+math.Sin(age*9)*4, w.k)
	}
}
