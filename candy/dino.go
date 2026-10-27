package main

import (
	"math"
)

// The dino hops about the floor in little runs, stops to look around and
// gobbles any candy it finds lying there. Ctrl makes it do something.

const (
	gagJump = iota
	gagRoar
	gagCatch
	gagRide
	dinoGags
)

type Dino struct {
	X, Y float64
	face float64 // 1 right, -1 left
	// a hop in progress
	hopping  bool
	hp, hdur float64
	hx0, hx1 float64
	hy0, hy1 float64
	hh       float64
	squash   float64
	rest     float64
	hopsLeft int
	chew     float64
	blink    float64
	// a trick
	gag      int
	gt       float64
	pending  bool
	stage    int
	catchBit *Bit
	catchX   float64
	onBelt   bool
	rot      float64
}

var dinoGreen = c3(120, 205, 110)

func (d *Dino) press(g *Game) {
	g.sfx.play("squeak", 1)
	d.squash = math.Max(d.squash, 0.5)
	if d.gag > 0 {
		d.pending = true
		g.hearts(d.X, d.Y-150, 1)
		return
	}
	d.start(g)
}

func (d *Dino) start(g *Game) {
	d.gag = g.next[2] + 1 // stored +1 so the zero Dino is idle
	g.next[2] = (g.next[2] + 1) % dinoGags
	d.gt, d.stage, d.pending = 0, 0, false
	d.hopping = false
	d.Y = feetY
	if d.onBelt {
		d.Y = beltY
	}
}

func (d *Dino) trick() int { return d.gag - 1 }

func (d *Dino) hop(x1, y1, dur, h float64) {
	d.hopping, d.hp, d.hdur = true, 0, dur
	d.hx0, d.hy0, d.hx1, d.hy1, d.hh = d.X, d.Y, x1, y1, h
	if x1 != d.X {
		d.face = math.Copysign(1, x1-d.X)
	}
}

// stepHop moves along the hop; it reports true on landing.
func (d *Dino) stepHop(g *Game) bool {
	if !d.hopping {
		return false
	}
	d.hp += dt / d.hdur
	p := math.Min(1, d.hp)
	d.X = lerp(d.hx0, d.hx1, p)
	d.Y = lerp(d.hy0, d.hy1, p) - 4*d.hh*p*(1-p)
	if d.hp >= 1 {
		d.hopping = false
		d.squash = math.Min(1, 0.35+d.hh/200)
		return true
	}
	return false
}

func (d *Dino) mouth() (float64, float64) { return d.X + d.face*62, d.Y - 98 }

func (d *Dino) update(g *Game) {
	d.squash = math.Max(0, d.squash-dt*3)
	d.chew = math.Max(0, d.chew-dt)
	d.blink -= dt
	if d.blink < -0.12 {
		d.blink = g.rnd(2, 4.5)
	}
	if d.gag > 0 {
		d.updateTrick(g)
		return
	}
	if d.stepHop(g) {
		d.eatNearby(g)
	}
	if d.hopping {
		return
	}
	if d.rest -= dt; d.rest > 0 {
		return
	}
	if d.hopsLeft <= 0 {
		d.rest, d.hopsLeft = g.rnd(0.8, 2.6), 2+g.rng.Intn(4)
		// fancy a candy off the floor?
		for _, b := range g.bits {
			if b.kind == bitCandy && b.resting && b.life > 1.5 {
				d.face = math.Copysign(1, b.x-d.X)
				return
			}
		}
		if g.rng.Float64() < 0.45 {
			d.face = -d.face
		}
		return
	}
	d.hopsLeft--
	x := d.X + d.face*g.rnd(50, 80)
	if x < 90 || x > W-190 {
		d.face = -d.face
		x = d.X + d.face*60
	}
	d.hop(x, feetY, 0.42, 34)
	g.sfx.play("hop", 0.9+g.rnd(0, 0.3))
}

func (d *Dino) eatNearby(g *Game) {
	mx := d.X + d.face*55
	for _, b := range g.bits {
		if b.kind == bitCandy && b.resting && b.life > 0.3 && math.Abs(b.x-mx) < 50 {
			b.life = 0
			d.gulp(g, b.k, false)
			return
		}
	}
}

func (d *Dino) gulp(g *Game, k col, loud bool) {
	d.chew = 0.9
	mx, my := d.mouth()
	g.crumbs(mx, my, k, 6)
	g.hearts(d.X, d.Y-150, 2)
	g.sfx.play("chomp", 0.8)
	if loud {
		g.say("YUM!", d.X, d.Y-200, 60, c3(255, 120, 150))
		g.sfx.play("yum", 1)
	}
}

func (d *Dino) updateTrick(g *Game) {
	d.gt += dt
	t := d.gt
	done := false
	switch d.trick() {
	case gagJump:
		switch d.stage {
		case 0: // crouch
			d.squash = 0.7 * ease(t/0.35)
			if t > 0.35 {
				d.stage = 1
				x := clamp(d.X+d.face*200, 100, W-190)
				d.hop(x, d.Y, 1.15, 300)
				d.squash = 0
				g.sfx.play("whee", 1)
			}
		case 1:
			p := math.Min(1, d.hp)
			d.rot = d.face * 360 * ease(p*1.1)
			if d.stepHop(g) {
				d.rot, d.stage = 0, 2
				d.squash = 1
				g.shake = 0.8
				g.dust(d.X, d.Y, 10)
				g.say("THUMP!", d.X, d.Y-210, 66, c3(255, 190, 90))
				g.sfx.play("thump", 1)
				for _, b := range g.bits {
					if b.kind == bitCandy && b.resting {
						b.resting, b.vy, b.vr = false, -g.rnd(200, 380), g.rnd(-6, 6)
					}
				}
				for _, r := range g.robots {
					r.surprise = 0.5
					r.kick(1)
				}
				d.gt = 0
			}
		case 2:
			done = t > 0.6
		}
	case gagRoar:
		if t < 0.05 {
			d.face = math.Copysign(1, 640-d.X)
		}
		if t > 0.45 && t-dt <= 0.45 {
			g.sfx.play("rawr", 1)
			mx, my := d.mouth()
			g.say("RAWR!", mx+d.face*90, my-60, 84, c3(140, 230, 120))
			g.shake = 0.5
			for _, r := range g.robots {
				r.surprise = 0.7
				r.kick(1.2)
			}
		}
		if t > 0.5 && t < 1.4 && int((t-dt)*12) != int(t*12) {
			mx, my := d.mouth()
			g.add(&Bit{kind: bitPuff, layer: layerFront, x: mx, y: my, vx: d.face * g.rnd(240, 340), vy: g.rnd(-50, 50),
				size: g.rnd(6, 10), k: fade(white, 0.7), life: 0.6})
		}
		if t > 1.8 && t-dt <= 1.8 {
			for _, r := range g.robots {
				hx, hy := r.headPos(g)
				g.hearts(hx, hy-40, 1)
				r.kick(0.5)
			}
			g.sfx.play("giggle", 1)
		}
		done = t > 2.5
	case gagCatch:
		switch d.stage {
		case 0:
			d.catchX = clamp(d.X+d.face*g.rnd(180, 320), 120, W-190)
			if math.Abs(d.catchX-d.X) < 100 {
				d.catchX = clamp(d.X-d.face*250, 120, W-190)
			}
			cx := d.catchX + math.Copysign(62, d.catchX-d.X)
			d.catchBit = g.add(&Bit{kind: bitCandy, layer: layerFront, x: cx, y: -30, vr: 4, size: 24,
				k: g.candyColor(), candy: 1 + g.rng.Intn(3), life: 6, gravity: 260, bounce: true})
			g.sfx.play("whistle", 1)
			d.stage = 1
		case 1: // scurry under it in quick hops
			if d.stepHop(g) || !d.hopping {
				if math.Abs(d.catchX-d.X) < 4 {
					d.stage = 2
					d.face = math.Copysign(1, d.catchBit.x-d.X)
				} else {
					step := clamp(d.catchX-d.X, -70, 70)
					d.hop(d.X+step, feetY, 0.22, 18)
				}
			}
		case 2: // mouth open, wait for it
			b := d.catchBit
			mx, my := d.mouth()
			if b.life > 0 && b.y > my-30 && math.Abs(b.x-mx) < 60 {
				b.life = 0
				d.gulp(g, b.k, true)
				d.stage, d.gt = 3, 0
			} else if b.life <= 0 || b.resting {
				d.stage, d.gt = 3, 0
			}
		case 3:
			done = t > 1.1
		}
		done = done || t > 6
	case gagRide:
		switch d.stage {
		case 0:
			if !d.onBelt {
				x := clamp(d.X, 160, 820)
				d.hop(x, beltY, 0.55, 110)
				d.face = 1
				d.onBelt = true
				g.sfx.play("hop", 0.7)
			}
			d.stage = 1
		case 1:
			if d.stepHop(g) {
				d.stage, d.gt = 2, 0
				g.say("WHEE!", d.X, d.Y-200, 62, c3(130, 220, 255))
				g.sfx.play("whee", 1.2)
			}
		case 2: // ride along, munching what is on the belt ahead
			d.X += beltSpeed * 2.2 * dt
			d.face = 1
			if c := g.belt.nearest(d.X+60, 26); c != nil {
				c.taken = true
				d.gulp(g, c.k, false)
			}
			if t > 2.6 || d.X > beltX1-120 {
				d.stage = 3
				d.hop(math.Min(d.X+110, W-190), feetY, 0.5, 70)
			}
		case 3:
			if d.stepHop(g) {
				d.onBelt = false
				g.dust(d.X, d.Y, 6)
				done = true
			}
		}
	}
	if done {
		d.gag, d.rot = 0, 0
		d.rest, d.hopsLeft = 0.6, 2
		if d.pending {
			d.start(g)
		}
	}
}

func (d *Dino) draw(g *Game) {
	t := g.t
	s := d.squash
	sx, sy := 1+s*0.28, 1-s*0.26
	if d.hopping && d.trick() != gagJump {
		p := d.hp
		st := math.Sin(p * math.Pi)
		sx, sy = 1-st*0.08, 1+st*0.12
	}
	roaring := d.trick() == gagRoar && d.gt > 0.4 && d.gt < 1.5
	inhale := d.trick() == gagRoar && d.gt < 0.45
	if inhale {
		sy += 0.1 * ease(d.gt/0.45)
		sx -= 0.04 * ease(d.gt/0.45)
	}
	if roaring {
		sx += math.Sin(t*50) * 0.02
	}
	// the shadow on the floor
	if !d.onBelt || d.hopping {
		ground := feetY
		if d.onBelt && !d.hopping {
			ground = beltY
		}
		h := clamp((ground-d.Y)/300, 0, 1)
		fillEllipse(d.X, ground+4, 52*(1-h*0.5), 10*(1-h*0.5), fade(ink, 0.22))
	}

	push(d.X, d.Y, 0, d.face*sx, sy)
	if d.rot != 0 {
		push(0, -55, d.rot*d.face, 1, 1)
		push(0, 55, 0, 1, 1)
	}
	part("dino", -95, -150, 190, 150, func() { d.drawBody(g, roaring) })
	if d.rot != 0 {
		pop()
		pop()
	}
	pop()
}

func (d *Dino) drawBody(g *Game, roaring bool) {
	t := g.t
	light := c3(245, 238, 170)
	spike := c3(255, 160, 90)
	wag := math.Sin(t*5) * 8
	walk := 0.0
	if d.hopping {
		walk = math.Sin(d.hp*math.Pi) * 8
	}
	// tail
	curveI(-30, -38, -70, -36+wag*0.4, -96, -60+wag, 22, dinoGreen)
	curveI(-60, -44, -80, -45+wag*0.4, -96, -60+wag, 10, dinoGreen)
	// back leg
	ellipseI(-16, -14+walk*0.3, 18, 15, mix(dinoGreen, ink, 0.12))
	// spikes along the back
	for i := 0; i < 5; i++ {
		a := math.Pi*1.05 + float64(i)*0.32
		bx, by := math.Cos(a)*44, -52+math.Sin(a)*42
		tx, ty := math.Cos(a)*62, -52+math.Sin(a)*60
		px, py := -math.Sin(a)*10, math.Cos(a)*10
		triI(bx-px, by-py, bx+px, by+py, tx, ty, spike)
	}
	ellipseI(0, -52, 48, 42, dinoGreen)
	ellipseI(14, -44, 26, 28, light)
	for i := 0; i < 3; i++ {
		fillLine(2, -60+float64(i)*12, 26, -60+float64(i)*12, 2, fade(mix(light, ink, 0.3), 0.6))
	}
	// spots
	fillCircle(-22, -66, 6, mix(dinoGreen, ink, 0.18))
	fillCircle(-32, -46, 4, mix(dinoGreen, ink, 0.18))
	// front leg
	ellipseI(20, -12-walk*0.3, 17, 14, dinoGreen)
	fillCircle(30, -6-walk*0.3, 3, white)
	// head
	bob := math.Sin(t*3) * 2
	if roaring {
		bob = -6
	}
	look := 0.0
	if d.trick() == gagCatch && d.stage == 2 {
		look = -14 // nose up to catch
	}
	push(40, -98+bob, look, 1, 1)
	// spike on the head
	triI(-24, -22, -10, -32, -34, -40, spike)
	circleI(0, 0, 36, dinoGreen)
	ellipseI(26, 8, 30, 22, dinoGreen)
	open := 0.0
	switch {
	case roaring:
		open = 1
	case d.trick() == gagCatch && d.stage == 2:
		open = 0.8
	case d.chew > 0:
		open = 0.5 + 0.5*math.Cos(d.chew*20)
	case d.trick() == gagRide && d.stage == 2:
		open = 0.6
	}
	if open > 0.05 {
		oh := 4 + open*22
		ellipseI(30, 16, 22, oh/2+2, c3(140, 50, 70))
		fillEllipse(30, 16+oh*0.25, 12, oh*0.2, c3(255, 120, 140))
		fillTri(18, 16-oh/2, 24, 16-oh/2, 21, 16-oh/2+6, white)
		fillTri(36, 16-oh/2, 42, 16-oh/2, 39, 16-oh/2+6, white)
	} else {
		curve(14, 18, 30, 28, 48, 16, 3, ink)
	}
	fillCircle(40, 0, 3, ink)
	fillCircle(48, 2, 3, ink)
	// eye
	if d.blink < 0 {
		curve(-4, -10, 6, -4, 16, -10, 3, ink)
	} else if roaring {
		fillLine(-2, -14, 16, -6, 4, ink) // squinty fierce
	} else {
		circleI(6, -10, 12, white)
		px, py := 3.0, 0.0
		if d.trick() == gagCatch && d.stage >= 1 {
			py = -5
		}
		circleI(8+px, -10+py, 6, ink)
		fillCircle(6+px, -12+py, 2, white)
	}
	fillCircle(-6, 12, 7, fade(c3(255, 130, 150), 0.5))
	pop()
	// little arms
	arm := math.Sin(t*4) * 4
	if d.trick() == gagRide && d.stage == 2 || d.trick() == gagJump && d.stage == 1 {
		arm = -18
	}
	lineI(26, -66, 42, -60+arm, 7, dinoGreen)
}
