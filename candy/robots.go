package main

import (
	"math"
)

// Three silly robots stand behind the belt. They bob, blink and watch the
// dino, and when a candy rides past they grab it, munch it and look very
// pleased. Enter sets them all off together.

type robotLook struct {
	name         string
	k            col
	bodyW, bodyH float64
	neck         float64
	headW, headH float64
}

var looks = []robotLook{
	{"bolt", c3(90, 205, 195), 92, 104, 12, 88, 72},
	{"gumdrop", c3(255, 165, 85), 124, 110, 6, 92, 80},
	{"tinny", c3(185, 160, 245), 74, 84, 28, 100, 76},
}

const (
	eatIdle = iota
	eatReach
	eatBring
	eatChew
	eatHappy
)

type Robot struct {
	kind     int
	x        float64
	look     robotLook
	blink    float64
	cool     float64
	phase    int
	pt       float64
	target   *beltCandy
	held     *beltCandy
	grabX    float64
	side     float64
	surprise float64
	ant      float64 // antenna wobble (spring)
	antV     float64
}

func newRobot(kind int, x float64) *Robot {
	return &Robot{kind: kind, x: x, look: looks[kind], side: 1}
}

const robotBase = beltY + 22 // bottom of the robots, hidden behind the belt

// ---- the crew's tricks ----

const (
	gagDance = iota
	gagHeads
	gagBurp
	crewGags
)

type Crew struct {
	gag     int
	gt      float64
	pending bool
	notes   float64
}

func (g *Game) robotsPress() {
	c := &g.crew
	g.sfx.play("bloop", 1)
	for _, r := range g.robots {
		r.kick(0.6)
	}
	if c.gag >= 0 && c.gt < 2 {
		c.pending = true
		for _, r := range g.robots {
			g.sparkles(r.x, robotBase-r.height()-10, 3, 30)
		}
		return
	}
	g.crewStart()
}

func (g *Game) crewStart() {
	c := &g.crew
	c.gag, c.gt, c.pending = g.next[1], 0, false
	g.next[1] = (g.next[1] + 1) % crewGags
	if c.gag == gagDance {
		g.disco = 4.4
		g.sfx.play("tune", 1)
		g.say("DANCE!", 640, 140, 70, c3(130, 220, 255))
	}
}

func (c *Crew) duration() float64 { return [...]float64{4.4, 2.8, 3.2}[c.gag] }

// robotGagT is how far into the trick robot i is (they go one after another).
func (g *Game) robotGagT(i int) float64 {
	switch g.crew.gag {
	case gagHeads:
		return g.crew.gt - float64(i)*0.3
	case gagBurp:
		return g.crew.gt - float64(i)*0.6
	}
	return g.crew.gt
}

func (g *Game) updateCrew() {
	c := &g.crew
	if c.gag < 0 {
		return
	}
	c.gt += dt
	for i, r := range g.robots {
		t1 := g.robotGagT(i)
		was := t1 - dt
		hx, hy := r.headPos(g)
		switch c.gag {
		case gagHeads:
			if was < 0 && t1 >= 0 {
				g.sfx.play("boing", 0.8+float64(i)*0.2)
				g.say("BOING!", hx, hy-130, 46, r.look.k)
			}
		case gagBurp:
			if was < 0.45 && t1 >= 0.45 {
				g.sfx.play("burp", 0.7+float64(i)*0.25)
				g.say("BURP!", hx+50, hy-60, 44, mix(r.look.k, white, 0.3))
				for j := 0; j < 7; j++ {
					g.add(&Bit{kind: bitBubble, layer: layerFront, x: hx + g.rnd(-10, 10), y: hy + 20,
						vx: g.rnd(-50, 50), vy: g.rnd(-120, -60), rot: g.rnd(0, 6), size: g.rnd(8, 16),
						k: candyColors[(j+i)%len(candyColors)], life: g.rnd(1.4, 2.2)})
				}
			}
			if was < 1.2 && t1 >= 1.2 {
				g.hearts(hx, hy-40, 2)
			}
		}
	}
	if c.gag == gagDance {
		if c.notes -= dt; c.notes <= 0 {
			c.notes = 0.25
			r := g.robots[g.rng.Intn(len(g.robots))]
			hx, hy := r.headPos(g)
			g.add(&Bit{kind: bitNote, layer: layerFront, x: hx + g.rnd(-40, 40), y: hy - 40, vx: g.rnd(-30, 30), vy: -90,
				rot: g.rnd(0, 6), size: 8, k: g.candyColor(), life: 1.6})
		}
	}
	if c.gt > c.duration()+float64(len(g.robots)-1)*0.6*b2f(c.gag == gagBurp) {
		c.gag = -1
		if c.pending {
			g.crewStart()
		}
	}
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// ---- one robot ----

func (r *Robot) height() float64 { return r.look.bodyH + r.look.neck + r.look.headH }

func (r *Robot) kick(s float64) { r.antV += s * 14 }

func (r *Robot) update(g *Game) {
	if r == g.robots[0] {
		g.updateCrew()
	}
	r.blink -= dt
	if r.blink < -0.13 {
		r.blink = g.rnd(2, 5)
	}
	r.surprise = math.Max(0, r.surprise-dt)
	// antenna spring
	r.antV += (-r.ant*60 - r.antV*5) * dt
	r.ant += r.antV * dt
	if g.crew.gag == gagDance {
		r.antV += math.Sin(g.t*16) * 6 * dt * 60
	}

	if g.crew.gag >= 0 && r.phase != eatChew {
		return // busy with a trick (but finish your mouthful)
	}
	r.pt += dt
	switch r.phase {
	case eatIdle:
		r.cool -= dt
		if r.cool <= 0 {
			if c := g.belt.nearest(r.x+30, 60); c != nil {
				r.target, r.phase, r.pt = c, eatReach, 0
				r.side = 1
				if c.x < r.x {
					r.side = -1
				}
			}
		}
	case eatReach:
		if r.target == nil || r.target.taken {
			r.phase, r.cool = eatIdle, 0.5
			break
		}
		if r.pt > 0.35 {
			r.held, r.target = r.target, nil
			r.grabX = r.held.x
			r.held.taken = true
			r.phase, r.pt = eatBring, 0
		}
	case eatBring:
		if r.pt > 0.4 {
			r.phase, r.pt = eatChew, 0
		}
	case eatChew:
		if int((r.pt-dt)/0.22) != int(r.pt/0.22) {
			mx, my := r.mouthPos(g)
			g.crumbs(mx, my, r.held.k, 4)
			g.sfx.play("chomp", 0.9+g.rnd(0, 0.3))
		}
		if r.pt > 1.1 {
			r.held = nil
			r.phase, r.pt = eatHappy, 0
			hx, hy := r.headPos(g)
			g.hearts(hx, hy-r.look.headH/2, 1)
			if g.rng.Float64() < 0.35 {
				g.say("YUM!", hx, hy-r.look.headH-20, 40, c3(255, 120, 150))
			}
			r.kick(0.4)
		}
	case eatHappy:
		if r.pt > 0.6 {
			r.phase, r.cool = eatIdle, g.rnd(2, 5.5)
		}
	}
}

// pose is the whole-body bounce and lean for the current moment.
func (r *Robot) pose(g *Game) (dy, lean float64) {
	i := float64(r.kind)
	dy = math.Sin(g.t*2.6+i*2) * 3
	if r.phase == eatHappy {
		dy -= math.Sin(r.pt/0.6*math.Pi) * 18
	}
	if r.surprise > 0 {
		dy -= math.Sin(r.surprise*math.Pi) * 26
	}
	if g.crew.gag == gagDance {
		ph := g.t*9 + i*1.4
		dy -= math.Abs(math.Sin(ph)) * 22
		lean = math.Sin(ph) * 10
	}
	return
}

func (r *Robot) headLift(g *Game) float64 {
	if g.crew.gag != gagHeads {
		return 0
	}
	t := g.robotGagT(r.kind)
	if t < 0 {
		return 0
	}
	return 95 * math.Sin(math.Pi*clamp(t/1.6, 0, 1)) * (1 + 0.25*math.Sin(t*18)*math.Exp(-t*2))
}

func (r *Robot) headPos(g *Game) (float64, float64) {
	dy, lean := r.pose(g)
	l := r.look
	h := l.bodyH + l.neck + l.headH/2 + r.headLift(g)
	a := lean * math.Pi / 180
	return r.x + math.Sin(a)*h, robotBase + dy - math.Cos(a)*h
}

func (r *Robot) mouthPos(g *Game) (float64, float64) {
	hx, hy := r.headPos(g)
	return hx, hy + r.look.headH*0.22
}

func (r *Robot) mouthOpen(g *Game) float64 {
	switch {
	case r.phase == eatBring:
		return ease(r.pt / 0.3)
	case r.phase == eatChew:
		return 0.5 + 0.5*math.Cos(r.pt/0.22*2*math.Pi)
	case g.crew.gag == gagBurp:
		t := g.robotGagT(r.kind)
		if t > 0.45 && t < 1.1 {
			return 1.3
		}
	case g.crew.gag == gagDance:
		return 0.5
	case r.surprise > 0:
		return 1
	}
	return 0
}

// look is where the eyes point: the candy it wants, the lollipop, or the dino.
func (r *Robot) lookAt(g *Game) (float64, float64) {
	switch {
	case g.machine.lollipop() > 0.2:
		return mX, domeY - 120
	case r.target != nil:
		return r.target.x, beltY
	case r.held != nil:
		return r.mouthPos(g)
	}
	return g.dino.X, g.dino.Y - 60
}

func (r *Robot) drawBody(g *Game) {
	dy, lean := r.pose(g)
	l := r.look
	push(r.x, robotBase+dy, lean, 1, 1)
	part(l.name+"-body", -l.bodyW/2, -l.bodyH, l.bodyW, l.bodyH, func() { r.drawTorso(g) })
	lift := r.headLift(g)
	// the neck: a spring when the head pops up
	ny0, ny1 := -l.bodyH, -l.bodyH-l.neck-lift
	if lift > 3 || r.kind == 2 {
		n := 8.0
		for i := 0.0; i < n; i++ {
			y := lerp(ny0, ny1, (i+0.5)/n)
			ellipseI(0, y, 13, 4, c3(200, 200, 215))
		}
	} else {
		roundI(-12, ny1, 24, l.neck+6, 4, c3(170, 170, 190))
	}
	pop()

	hx, hy := r.headPos(g)
	squeeze := 1.0
	if g.crew.gag == gagBurp {
		t := g.robotGagT(r.kind)
		if t > 0 && t < 0.45 {
			squeeze = 1 + 0.18*ease(t/0.45)
		}
	}
	push(hx, hy, lean+r.ant*0.4, squeeze, 1/math.Sqrt(squeeze))
	part(l.name+"-head", -l.headW/2, -l.headH/2, l.headW, l.headH, func() { r.drawHead(g) })
	pop()
}

func (r *Robot) drawTorso(g *Game) {
	l := r.look
	k := l.k
	switch r.kind {
	case 0:
		roundI(-l.bodyW/2, -l.bodyH, l.bodyW, l.bodyH, 14, k)
		roundI(-28, -l.bodyH+18, 56, 40, 8, c3(240, 248, 245))
		// a belly meter that fills with candy
		fill := 0.3 + 0.2*math.Sin(g.t)
		if r.phase == eatHappy || r.phase == eatChew {
			fill = 0.9
		}
		fillRound(-22, -l.bodyH+24+28*(1-fill), 44, 28*fill, 4, c3(255, 120, 150))
		for i := 0; i < 3; i++ {
			on := int(g.t*3+float64(i))%3 == 0
			kk := c3(120, 130, 140)
			if on {
				kk = c3(255, 230, 100)
			}
			circleI(-20+float64(i)*20, -l.bodyH+76, 6, kk)
		}
	case 1:
		ellipseI(0, -l.bodyH/2, l.bodyW/2, l.bodyH/2, k)
		ellipseI(0, -l.bodyH/2+8, l.bodyW*0.32, l.bodyH*0.3, mix(k, white, 0.45))
		// a little door with a heart on the tummy
		heart(0, -l.bodyH/2+2, 9, c3(255, 110, 140))
	case 2:
		roundI(-l.bodyW/2, -l.bodyH, l.bodyW, l.bodyH, 10, k)
		for i := 0; i < 3; i++ {
			fillLine(-20, -l.bodyH+20+float64(i)*16, 20, -l.bodyH+20+float64(i)*16, 5, mix(k, ink, 0.3))
		}
		// a little crank on the side, turning
		a := g.t * 3
		lineI(l.bodyW/2, -l.bodyH/2, l.bodyW/2+math.Cos(a)*16, -l.bodyH/2+math.Sin(a)*16, 4, c3(255, 210, 80))
		circleI(l.bodyW/2+math.Cos(a)*16, -l.bodyH/2+math.Sin(a)*16, 5, c3(255, 120, 120))
	}
}

func (r *Robot) drawHead(g *Game) {
	l := r.look
	k := mix(l.k, white, 0.12)
	w, h := l.headW, l.headH
	open := r.mouthOpen(g)
	lx, ly := r.lookAt(g)
	hx, hy := r.headPos(g)
	ex, ey := lx-hx, ly-hy
	d := math.Hypot(ex, ey)
	if d > 0 {
		ex, ey = ex/d, ey/d
	}
	spin := g.crew.gag == gagHeads && r.headLift(g) > 10
	if spin {
		a := g.t * 14
		ex, ey = math.Cos(a), math.Sin(a)
	}
	blink := r.blink < 0 && r.surprise == 0
	big := 1 + 0.3*math.Min(1, r.surprise*3)

	eye := func(x, y, rad float64, screen bool) {
		if blink {
			fillLine(x-rad, y, x+rad, y, 4, ink)
			return
		}
		if screen {
			fillRound(x-rad*0.5, y-rad*big, rad, rad*2*big, rad*0.5, c3(140, 255, 200))
			return
		}
		circleI(x, y, rad*big, white)
		circleI(x+ex*rad*0.4, y+ey*rad*0.4, rad*0.48, ink)
		fillCircle(x+ex*rad*0.4-rad*0.18, y+ey*rad*0.4-rad*0.18, rad*0.16, white)
	}
	mouth := func(y, mw float64) {
		mh := 4 + open*h*0.22
		roundI(-mw/2, y-mh/2, mw, mh, math.Min(mh, mw)/2, c3(120, 40, 60))
		if open > 0.3 {
			fillRound(-mw/2+4, y-mh/2+1, mw-8, 5, 2, white) // teeth
		}
	}

	switch r.kind {
	case 0: // Bolt: square head, antenna with a bobble, two big eyes
		ax := r.ant * 18
		curveI(0, -h/2, ax*0.3, -h/2-24, ax, -h/2-44, 4, c3(170, 170, 190))
		glow := 0.5 + 0.5*math.Sin(g.t*5)
		fillCircle(ax, -h/2-46, 15, fade(c3(255, 90, 110), 0.25*glow))
		circleI(ax, -h/2-46, 9, c3(255, 90, 110))
		roundI(-w/2-8, -10, 10, 22, 4, c3(170, 170, 190))
		roundI(w/2-2, -10, 10, 22, 4, c3(170, 170, 190))
		roundI(-w/2, -h/2, w, h, 16, k)
		eye(-19, -8, 14, false)
		eye(19, -8, 14, false)
		mouth(h*0.26, 36)
		fillCircle(-30, 12, 6, fade(c3(255, 120, 150), 0.5))
		fillCircle(30, 12, 6, fade(c3(255, 120, 150), 0.5))
	case 1: // Gumdrop: round, one huge eye, a propeller
		spinp := g.t * 6
		if g.crew.gag >= 0 {
			spinp = g.t * 30
		}
		lineI(0, -h/2-4, 0, -h/2-22, 5, c3(170, 170, 190))
		push(0, -h/2-24, 0, math.Cos(spinp), 1)
		ellipseI(0, 0, 34, 6, c3(255, 220, 90))
		pop()
		circleI(0, -h/2-24, 5, c3(255, 110, 140))
		ellipseI(0, 0, w/2, h/2, k)
		eye(0, -10, 22, false)
		mouth(h*0.28, 44)
	case 2: // Tinny: a little TV with a screen face and two ears
		for _, s := range []float64{-1, 1} {
			lineI(s*w*0.25, -h/2, s*(w*0.4+r.ant*10), -h/2-26, 3, c3(170, 170, 190))
			circleI(s*(w*0.4+r.ant*10), -h/2-28, 6, c3(255, 210, 80))
			circleI(s*(w/2+6), 0, 10, c3(255, 160, 190))
		}
		roundI(-w/2, -h/2, w, h, 12, k)
		roundI(-w/2+10, -h/2+9, w-20, h-18, 10, c3(40, 60, 70))
		// scan line
		y := -h/2 + 10 + math.Mod(g.t*30, h-20)
		fillRect(-w/2+12, y, w-24, 2, fade(c3(140, 255, 200), 0.3))
		eye(-18+ex*3, -6+ey*3, 10, true)
		eye(18+ex*3, -6+ey*3, 10, true)
		// a pixel smile, or a chomping mouth
		if open > 0.1 {
			mh := 4 + open*12
			fillRound(-14, 12-mh/2+4, 28, mh, 3, c3(140, 255, 200))
		} else {
			rl := c3(140, 255, 200)
			fillRect(-14, 12, 5, 5, rl)
			fillRect(-9, 16, 18, 5, rl)
			fillRect(9, 12, 5, 5, rl)
		}
	}
}

// drawArms runs after the belt so hands can reach over it.
func (r *Robot) drawArms(g *Game) {
	dy, lean := r.pose(g)
	l := r.look
	base := robotBase + dy
	shY := base - l.bodyH + 26
	k := c3(170, 170, 190)
	for _, s := range []float64{-1, 1} {
		sx := r.x + s*(l.bodyW/2-4)
		if r.kind == 1 {
			sx = r.x + s*(l.bodyW/2-10)
			shY = base - l.bodyH*0.55
		}
		sx += lean * 0.6
		hx, hy := sx+s*26, shY+48+math.Sin(g.t*2.6+float64(r.kind)+s)*3
		eating := s == r.side
		if eating {
			switch r.phase {
			case eatReach:
				if r.target != nil {
					u := ease(r.pt / 0.35)
					hx, hy = lerp(hx, r.target.x, u), lerp(hy, beltY-12, u)
				}
			case eatBring, eatChew:
				mx, my := r.mouthPos(g)
				u := 1.0
				if r.phase == eatBring {
					u = ease(r.pt / 0.4)
				}
				hx, hy = lerp(r.grabX, mx+s*10, u), lerp(beltY-12, my+8, u)
			}
		}
		if g.crew.gag == gagDance {
			ph := g.t*9 + float64(r.kind)*1.4 + (s+1)*1.6
			hx = sx + s*(40+math.Sin(ph)*16)
			hy = shY - 40 - math.Cos(ph)*30
		}
		if g.crew.gag == gagHeads && r.headLift(g) > 10 {
			hx, hy = sx+s*44, shY-10 // arms out in surprise
		}
		if r.surprise > 0 {
			hx, hy = sx+s*40, shY-50
		}
		cx := (sx+hx)/2 + s*20
		cy := math.Max(shY, hy) + 10
		curveI(sx, shY, cx, cy, hx, hy, 9, k)
		if eating && r.held != nil {
			drawCandy(hx, hy-6, 0.3*s, 13, r.held.k, r.held.kind)
		}
		// mitten or claw
		if r.kind == 2 {
			lineI(hx, hy, hx+s*8, hy-10, 5, c3(255, 210, 80))
			lineI(hx, hy, hx+s*10, hy+6, 5, c3(255, 210, 80))
		} else {
			circleI(hx, hy, 11, mix(l.k, white, 0.3))
		}
	}
}
