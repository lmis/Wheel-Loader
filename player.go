package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Player is one person at the keyboard: a kid on foot, or the driver of a
// vehicle, with their own camera and (in co-op) their own half of the screen.
type Player struct {
	ID      int
	keys    keyMap
	W       *Walker
	Driving Vehicle
	cam     *Camera
	fp      bool // first-person view

	headYaw, headPitch      float64 // first-person head look inside a vehicle
	distWalker, distVehicle float64 // chase-camera distances per mode
	near                    Vehicle // vehicle close enough to climb into
	ctl                     Controls
	scriptCtl               *Controls // scripted controls (screenshot scenes)
	view                    *ebiten.Image
	marker                  color.RGBA
	msg                     string
	msgT                    float64
	td                      *toddlerState // toddler-mode state, nil otherwise
}

func newPlayer(id int, keys keyMap, at Vec, yaw float64) *Player {
	cam := NewCamera(screenW, screenH)
	cam.Target, cam.followYaw = at, yaw
	p := &Player{
		ID: id, keys: keys, cam: cam,
		W:          &Walker{Pos: at, Yaw: yaw},
		distWalker: 150, distVehicle: 520,
		marker: rgba(60, 200, 255, 230),
	}
	if id == 2 {
		p.marker = rgba(255, 90, 200, 230)
	}
	return p
}

func (p *Player) say(s string) { p.msg, p.msgT = s, 2.5 }

// driverOf returns the player driving v, if any.
func (g *Game) driverOf(v Vehicle) *Player {
	for _, p := range g.players {
		if p.Driving == v {
			return p
		}
	}
	return nil
}

// nearestVehicle finds the closest free vehicle within climbing-in range.
func (g *Game) nearestVehicle(p *Player) Vehicle {
	var best Vehicle
	bestD := math.Inf(1)
	for _, v := range g.vehicles {
		if g.driverOf(v) != nil {
			continue
		}
		d := v.Center().Sub(p.W.Pos).Len()
		if d < v.EnterRadius() && d < bestD {
			best, bestD = v, d
		}
	}
	return best
}

func (g *Game) enter(p *Player, v Vehicle) {
	p.Driving = v
	v.SetOccupied(true)
	p.distVehicle = v.CamDist()
	p.headYaw, p.headPitch = 0, 0
	p.cam.Pitch = 0.62
	p.cam.UserYaw = 0
}

// exit climbs out on whichever side is free. Vehicles must be nearly
// stopped first.
func (g *Game) exit(p *Player) bool {
	v := p.Driving
	if v == nil {
		return false
	}
	if math.Abs(v.SpeedVal()) > 35 {
		p.say("Slow down to get out")
		return false
	}
	for _, side := range []float64{1, -1} {
		pt := v.ExitPoint(side)
		if pt.X < 20 || pt.Y < 20 || pt.X > g.world.Width-20 || pt.Y > g.world.Height-20 {
			continue
		}
		if hit, _ := g.world.CollideCircle(pt, kidRadius+1, nil); hit {
			continue
		}
		p.W.Pos, p.W.Yaw, p.W.vel, p.W.Pitch = pt, v.Facing(), Vec{}, 0
		v.SetOccupied(false)
		p.Driving = nil
		p.cam.Pitch = 0.38
		p.cam.UserYaw = 0
		return true
	}
	p.say("No room to climb out here")
	return false
}

func (g *Game) toggleVehicle(p *Player) {
	if g.toddler {
		return // toddlers stay in their machine
	}
	if p.Driving != nil {
		g.exit(p)
		return
	}
	if v := g.nearestVehicle(p); v != nil {
		g.enter(p, v)
	}
}

// updatePlayerCamera places this player's camera for the current mode.
func (g *Game) updatePlayerCamera(p *Player, dt float64) {
	c := p.cam
	switch {
	case p.fp && p.Driving != nil:
		eye, yaw := p.Driving.Seat()
		pitch := -0.1
		if sp, ok := p.Driving.(interface{ SeatPitch() float64 }); ok {
			pitch = sp.SeatPitch()
		}
		c.SetFP(eye, yaw+p.headYaw, pitch+p.headPitch)
	case p.fp:
		bob := math.Abs(math.Sin(p.W.phase)) * 0.5 * math.Min(p.W.moving, 1)
		c.SetFP(ground(p.W.Pos, kidEyeH+bob), p.W.Yaw, p.W.Pitch)
	case p.Driving != nil:
		v := p.Driving
		tgt, h, head := v.Center(), 40.0, v.Facing()
		if ct, ok := v.(interface {
			CamTarget() (Vec, float64, float64)
		}); ok {
			tgt, h, head = ct.CamTarget()
		}
		c.TgtDist = p.distVehicle
		c.SetLookHeight(h)
		c.Follow(tgt, head, dt)
	default:
		c.TgtDist = p.distWalker
		c.SetLookHeight(14)
		c.Follow(p.W.Pos, p.W.Yaw, dt)
	}
	if v := p.Driving; v != nil {
		if j := v.Jolt(); j > 60 {
			c.Kick(math.Min(10, j*0.05))
		}
	}
	c.Tick(dt)
	if p.msgT > 0 {
		p.msgT -= dt
	}
}

// setCoop switches between one full-screen player and two side by side.
func (g *Game) setCoop(on bool) {
	if on == g.coop() {
		return
	}
	p1 := g.players[0]
	if on {
		at := p1.W.Pos.Add(V(30, 20))
		if p1.Driving != nil {
			at = p1.Driving.ExitPoint(-1)
		}
		p2 := newPlayer(2, p2Keys, at, p1.W.Yaw)
		g.players = append(g.players, p2)
		p1.keys = p1Keys
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
		g.mouseCaptured = false
		p2.say("Player 2 joined!")
	} else {
		p2 := g.players[1]
		if p2.Driving != nil {
			p2.Driving.SetOccupied(false)
		}
		g.players = g.players[:1]
		p1.keys = p1Keys.solo()
	}
}

func (g *Game) coop() bool { return len(g.players) > 1 }
