package main

import (
	"image/color"
	"math"
)

// Player is one kid at the keyboard, in a machine or on foot walking over
// to the next one, with their own camera and half of the screen.
type Player struct {
	ID      int
	W       *Walker
	Driving Vehicle
	cam     *Camera
	fp      bool // looking out of the cab instead of the chase camera

	distWalker, distVehicle float64 // chase-camera distances per mode
	ctl                     Controls
	scriptCtl               *Controls // scripted controls (screenshot scenes)
	marker                  color.RGBA
	td                      *toddlerState
	trail                   []Vec // where the kid walked, for the animals to follow
}

func newPlayer(id int, at Vec, yaw float64) *Player {
	cam := NewCamera(screenW/2, screenH)
	cam.Target, cam.followYaw = at, yaw
	p := &Player{
		ID: id, cam: cam,
		W:          &Walker{Pos: at, Yaw: yaw},
		distWalker: 150, distVehicle: 520,
		marker: rgba(60, 200, 255, 230),
	}
	if id == 2 {
		p.marker = rgba(255, 90, 200, 230)
	}
	return p
}

// driverOf returns the player driving v, if any.
func (g *Game) driverOf(v Vehicle) *Player {
	for _, p := range g.players {
		if p.Driving == v {
			return p
		}
	}
	return nil
}

func (g *Game) enter(p *Player, v Vehicle) {
	p.Driving = v
	v.SetOccupied(true)
	p.distVehicle = v.CamDist()
	p.cam.Pitch = 0.52
	p.cam.UserYaw = 0
}

// exit climbs out on whichever side is free. Vehicles must be nearly
// stopped first.
func (g *Game) exit(p *Player) bool {
	v := p.Driving
	if v == nil || math.Abs(v.SpeedVal()) > 35 {
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
		p.W.Pos, p.W.Yaw, p.W.vel = pt, v.Facing(), Vec{}
		p.trail = append(p.trail[:0], pt)
		v.SetOccupied(false)
		p.Driving = nil
		p.cam.Pitch = 0.38
		p.cam.UserYaw = 0
		return true
	}
	return false
}

// put seats p in v, taking them out of whatever they were in.
func (g *Game) put(p *Player, v Vehicle) {
	if p.Driving == v {
		return
	}
	if p.Driving != nil {
		p.Driving.SetOccupied(false)
		p.Driving = nil
	}
	if q := g.driverOf(v); q != nil {
		q.Driving = nil
	}
	g.enter(p, v)
}

// snapCamera moves p's chase camera straight to its resting position.
func (g *Game) snapCamera(p *Player) {
	c := p.cam
	if v := p.Driving; v != nil {
		tgt, h, head := v.Center(), 40.0, v.Facing()
		if ct, ok := v.(interface {
			CamTarget() (Vec, float64, float64)
		}); ok {
			tgt, h, head = ct.CamTarget()
		}
		c.Target, c.followYaw, c.lookHeight = tgt, head, h
		c.Dist, c.TgtDist = p.distVehicle, p.distVehicle
		return
	}
	c.Target, c.followYaw, c.lookHeight = p.W.Pos, p.W.Yaw, 14
	c.Dist, c.TgtDist = p.distWalker, p.distWalker
}

// updatePlayerCamera places this player's camera for the current mode.
func (g *Game) updatePlayerCamera(p *Player, dt float64) {
	c := p.cam
	st, inCab := p.Driving.(seated)
	switch {
	case p.fp && inCab:
		eye, yaw := st.Seat()
		c.SetFP(eye, yaw, st.SeatPitch())
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
		head := p.W.Yaw
		c.TgtDist = p.distWalker
		c.SetLookHeight(14)
		c.Follow(p.W.Pos, head, dt)
	}
	if v := p.Driving; v != nil {
		if j := v.Jolt(); j > 60 {
			c.Kick(math.Min(10, j*0.05))
		}
	}
	c.Tick(dt)
}
