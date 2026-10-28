package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The candy machine in the middle of the room. It hums, its gears turn, its
// dials wobble, its chimney puffs, and now and then it drops a candy down its
// chute onto the belt. Space makes it do something big.

const (
	mX     = 640.0 // centre
	mTop   = 250.0 // top of the body
	mBot   = 480.0
	mHalfW = 140.0
	domeY  = mTop - 4
	chimX  = mX + 105
)

const (
	gagFountain = iota
	gagLollipop
	gagBubble
	machineGags
)

type Machine struct {
	gag     int // -1 when idle
	gt      float64
	pending bool
	puff    float64
	chute   float64
	btn     float64 // the big button going down
	lid     float64 // dome lid hop
	shot    int     // candies fired so far in the fountain
	bubble  float64
}

func (m *Machine) press(g *Game) {
	m.btn = 0.25
	g.sfx.play("click", 1)
	if m.gag >= 0 && m.gt < 2 {
		// already busy: a little extra burst now, the next trick after
		g.sparkles(mX, mTop+150, 6, 30)
		m.pending = true
		return
	}
	m.start(g)
}

func (m *Machine) start(g *Game) {
	m.gag = g.next[0]
	g.next[0] = (g.next[0] + 1) % machineGags
	m.gt, m.shot, m.pending = 0, 0, false
	switch m.gag {
	case gagFountain:
		g.sfx.play("wind", 1)
	case gagLollipop:
		g.sfx.play("slide", 1)
	case gagBubble:
		g.sfx.play("blow", 1)
	}
}

func (m *Machine) duration() float64 {
	return [...]float64{3.0, 4.0, 3.6}[m.gag]
}

func (m *Machine) update(g *Game) {
	m.btn = math.Max(0, m.btn-dt)
	m.lid = math.Max(0, m.lid-dt*2.5)
	// chimney puffs
	if m.puff -= dt; m.puff <= 0 {
		m.puff = g.rnd(1.8, 3)
		g.puff(chimX, mTop-70, g.rnd(12, 18), c3(255, 236, 245), layerBack)
	}
	// a candy down the chute every so often
	if m.chute -= dt; m.chute <= 0 {
		m.chute = g.rnd(2.5, 4.5)
		g.spawnCandy(mX+mHalfW+30, mBot-30, 40, -60).life = 3
	}
	if m.gag < 0 {
		return
	}
	m.gt += dt
	t := m.gt
	switch m.gag {
	case gagFountain:
		if t > 0.6 && t-dt <= 0.6 {
			m.lid = 1
			g.shake = 0.6
			g.sfx.play("boing", 1)
			g.say("BOING!", mX, domeY-150, 72, c3(255, 210, 80))
		}
		if t > 0.6 && t < 1.6 && m.shot < int((t-0.6)*30) {
			m.shot++
			a := -math.Pi/2 + g.rnd(-0.9, 0.9)
			s := g.rnd(380, 640)
			g.spawnCandy(mX+g.rnd(-30, 30), domeY-50, math.Cos(a)*s, math.Sin(a)*s)
			if m.shot%6 == 0 {
				g.sfx.play("pop", 0.6+g.rnd(0, 0.6))
			}
		}
	case gagLollipop:
		if t > 0.4 && t-dt <= 0.4 {
			m.lid = 1
			g.say("WOW!", mX+190, domeY-120, 64, c3(255, 140, 200))
			g.sparkles(mX, domeY-130, 14, 80)
		}
	case gagBubble:
		m.bubble = ease(t/2.6) * 90
		if t > 2.6 && t-dt <= 2.6 {
			m.bubble = 0
			g.shake = 0.5
			g.sfx.play("pop", 0.5)
			g.say("POP!", chimX+40, mTop-150, 80, c3(255, 120, 180))
			g.confetti(chimX+40, mTop-160, 90, 520)
			for i := 0; i < 6; i++ {
				g.spawnCandy(chimX+40+g.rnd(-50, 50), mTop-160+g.rnd(-30, 30), g.rnd(-300, 300), g.rnd(-300, 0))
			}
		}
	}
	if t > m.duration() {
		m.gag, m.bubble = -1, 0
		if m.pending {
			m.start(g)
		}
	}
}

// lollipop is how far the giant lollipop has sprung out of the dome (0..1).
func (m *Machine) lollipop() float64 {
	if m.gag != gagLollipop {
		return 0
	}
	t := m.gt
	if t < 0.4 {
		return 0
	}
	out := springOut((t - 0.4) * 1.1)
	if t > 3.3 {
		out *= 1 - ease((t-3.3)/0.6)
	}
	return out
}

// squash is how the machine breathes and winds up.
func (m *Machine) squash(g *Game) (sx, sy float64) {
	s := 0.012 * math.Sin(g.t*2.3)
	sx, sy = 1-s, 1+s
	if m.gag == gagFountain {
		t := m.gt
		if t < 0.6 {
			k := ease(t / 0.6)
			sx += 0.08 * k
			sy -= 0.1 * k
			sx += math.Sin(t*80) * 0.012 * k
		} else {
			k := springOut(t-0.6) - 1
			sx -= 0.06 * k
			sy += 0.08 * k
		}
	}
	if m.gag == gagBubble && m.gt < 2.6 {
		sy += 0.03 * math.Sin(m.gt*14)
	}
	return
}

// drawBack is what sits behind the robots: the gears at the machine's sides
// and the giant lollipop.
func (m *Machine) drawBack(g *Game) {
	t := g.t
	spin := 1.0
	if m.gag >= 0 {
		spin = 4
	}
	gear(mX-mHalfW-8, mTop+70, 46, t*0.8*spin, 10, c3(250, 200, 80))
	gear(mX-mHalfW-50, mTop+150, 30, -t*0.8*spin*46/30+0.2, 7, c3(130, 210, 200))
	gear(mX+mHalfW+10, mTop+40, 34, -t*0.9*spin, 8, c3(180, 150, 240))
}

func (m *Machine) draw(g *Game) {
	t := g.t
	sx, sy := m.squash(g)

	// giant lollipop, behind the dome
	if l := m.lollipop(); l > 0 {
		h := l * 120
		wob := math.Sin(m.gt*9) * 8 * math.Exp(-m.gt)
		x, y := mX+wob, domeY-30-h
		push(x, y, wob, 1, 1)
		part("lollipop", -72, -72, 144, 264, func() {
			lineI(0, 0, 0, h+30, 16, white)
			for i := 0; i < 5; i++ {
				fillLine(-8, 30+float64(i)*h/5, 8, 50+float64(i)*h/5, 5, c3(255, 120, 150))
			}
			circleI(0, 0, 70, c3(255, 110, 160))
			for i := 0; i < 28; i++ {
				a := float64(i)*0.55 + m.gt*4
				r := float64(i) * 2.4
				fillCircle(math.Cos(a)*r, math.Sin(a)*r, 9, candyColors[i%len(candyColors)])
			}
			shine(-30, -34, 14)
		})
		pop()
	}

	// chimney (and the bubblegum bubble)
	chimH := 80.0
	cp := 0.0
	if m.puff > 0 && m.puff < 0.2 {
		cp = math.Sin(m.puff / 0.2 * math.Pi)
	}
	push(chimX, mTop, 0, 1+cp*0.12, 1-cp*0.08)
	roundI(-16, -chimH, 32, chimH+10, 4, c3(170, 110, 200))
	roundI(-24, -chimH-12, 48, 18, 6, c3(200, 140, 230))
	pop()
	if m.bubble > 0 {
		r := m.bubble
		wb := 1 + 0.06*math.Sin(m.gt*11)
		cy := mTop - chimH - 16 - r*0.7
		bx := chimX + r*0.45
		fillEllipse(bx, cy, r*wb+inkW, r/wb+inkW, ink)
		fillEllipse(bx, cy, r*wb, r/wb, c3(255, 140, 200))
		fillEllipse(bx-r*0.35, cy-r*0.4, r*0.25, r*0.15, fade(white, 0.8))
		fillCircle(bx+r*0.4, cy+r*0.3, r*0.08, fade(white, 0.6))
	}

	push(mX, mBot, 0, sx, sy)
	part("machine", -mHalfW, mTop-mBot, 2*mHalfW, mBot-mTop, func() { m.drawBody(g) })
	pop()

	// the glass dome full of gumballs, on top
	lid := math.Sin(m.lid*math.Pi) * 40
	push(mX, mBot-(mBot-domeY)*sy, 0, sx, sy)
	rl.DrawCircleSector(v2(0, 0), 98, 180, 360, 36, ink)
	rl.DrawCircleSector(v2(0, 0), 95, 180, 360, 36, fade(c3(210, 240, 255), 0.95))
	for i := 0; i < 34; i++ {
		row := float64(i / 9)
		cx := (float64(i%9) - 4) * 19 * (1 - row*0.14)
		cy := -12 - row*20 + math.Sin(t*6+float64(i))*1.5
		if m.gag >= 0 {
			cy += math.Sin(t*30+float64(i)*2) * 4
		}
		circleI(cx, cy, 9, candyColors[i%len(candyColors)])
		fillCircle(cx-3, cy-3, 2.5, fade(white, 0.8))
	}
	fillEllipse(-45, -55, 16, 26, fade(white, 0.55))
	// the lid with a knob
	push(0, -96-lid, math.Sin(m.lid*12)*8, 1, 1)
	roundI(-36, -6, 72, 14, 7, c3(255, 110, 140))
	circleI(0, -12, 11, c3(255, 220, 90))
	pop()
	// the base ring hides the bottom of the dome
	roundI(-108, -4, 216, 20, 10, c3(255, 110, 140))
	pop()
}

// drawBody is the machine's face-like front, in a frame with its base at 0.
func (m *Machine) drawBody(g *Game) {
	t := g.t
	h := mBot - mTop
	// legs
	lineI(-mHalfW+30, 0, -mHalfW+20, beltY-mBot+40, 14, c3(150, 150, 170))
	lineI(mHalfW-30, 0, mHalfW-20, beltY-mBot+40, 14, c3(150, 150, 170))
	roundI(-mHalfW, -h, 2*mHalfW, h, 34, c3(255, 150, 175))
	for x := -mHalfW + 30; x < mHalfW-10; x += 40 {
		fillRect(x, -h+30, 14, h-60, fade(white, 0.18))
	}
	// two dials, like eyes, needles wobbling
	for i, dx := range []float64{-58, 58} {
		cx, cy := dx, -h+62
		circleI(cx, cy, 34, c3(255, 230, 120))
		circleI(cx, cy, 26, white)
		a := -math.Pi/2 + math.Sin(t*1.7+float64(i)*2)*0.6 + math.Sin(t*13+float64(i))*0.05
		if m.gag >= 0 {
			a = -math.Pi/2 + math.Sin(t*14+float64(i)*2)*1.3
		}
		fillLine(cx, cy, cx+math.Cos(a)*20, cy+math.Sin(a)*20, 4, c3(230, 70, 80))
		fillCircle(cx, cy, 4, ink)
	}
	// a smiling slot
	mw := 30.0
	if m.gag >= 0 {
		mw = 44
	}
	roundI(-mw, -h+112, 2*mw, 18+mw*0.2, 10, ink)
	fillRound(-mw+8, -h+114, 2*mw-16, 6, 3, c3(255, 120, 150))
	// chasing lights
	for i := 0; i < 7; i++ {
		x := -90 + float64(i)*30
		on := int(t*6)%7 == i || int(t*6+3)%7 == i
		if m.gag >= 0 {
			on = int(t*20+float64(i))%2 == 0
		}
		k := mix(candyColors[i%len(candyColors)], ink, 0.45)
		if on {
			fillCircle(x, -h+160, 14, fade(candyColors[i%len(candyColors)], 0.35))
			k = mix(candyColors[i%len(candyColors)], white, 0.3)
		}
		circleI(x, -h+160, 8, k)
	}
	// the big red button
	p := math.Sin(math.Min(1, m.btn/0.25)*math.Pi) * 6
	roundI(-44, -h+188, 88, 24, 8, c3(200, 200, 215))
	ellipseI(0, -h+186+p, 30, 14-p*0.8, c3(240, 60, 70))
	fillEllipse(-10, -h+182+p, 9, 4, fade(white, 0.7))
	// the chute on the right
	curveI(mHalfW-20, -60, mHalfW+30, -60, mHalfW+34, -14, 26, c3(200, 200, 215))
}
