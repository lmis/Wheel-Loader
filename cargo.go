package main

import "math"

// Cargo kinds the tower crane can lift.
const (
	cargoBricks = iota
	cargoBeams
	cargoPipe
	cargoCrate
	cargoContainer
)

var cargoNames = [...]string{"Bricks", "Steel beams", "Concrete pipe", "Crate", "Container"}

// Cargo is a loose item: it can hang from the crane hook, fall, stack on
// other cargo, and ride on a dump truck's bed.
type Cargo struct {
	Pos  Vec
	Y    float64 // height of its underside
	Yaw  float64
	Kind int
	W    float64 // length along Yaw
	D    float64 // width across
	H    float64 // height

	held  bool
	vy    float64
	truck *Truck // riding on this truck's bed
	local Vec    // offset in truck space while riding
	lyaw  float64
}

func newCargo(kind int, at Vec, yaw float64) *Cargo {
	c := &Cargo{Pos: at, Yaw: yaw, Kind: kind}
	switch kind {
	case cargoBricks:
		c.W, c.D, c.H = 34, 28, 22
	case cargoBeams:
		c.W, c.D, c.H = 92, 22, 13
	case cargoPipe:
		c.W, c.D, c.H = 40, 26, 26
	case cargoCrate:
		c.W, c.D, c.H = 30, 30, 28
	case cargoContainer:
		c.W, c.D, c.H = 86, 34, 36
	}
	return c
}

func (c *Cargo) Name() string { return cargoNames[c.Kind] }

// footRadius is a generous circle around the item's footprint.
func (c *Cargo) footRadius() float64 { return math.Max(c.W, c.D) * 0.5 }

// support finds what is directly beneath the item: the ground, a truck bed
// or the top of other cargo. It returns that height and the truck it belongs
// to, if any.
func (w *World) support(c *Cargo) (float64, *Truck) {
	h := 0.0
	var tr *Truck
	if t := w.TruckAt(c.Pos); t != nil && t.BedTilt < 0.1 {
		h, tr = truckBedFloorY, t
	}
	for _, o := range w.Cargo {
		if o == c || o.held {
			continue
		}
		if o.Pos.Sub(c.Pos).Len() < (o.footRadius()+math.Min(c.W, c.D)*0.5)*0.7 && o.Y+o.H > h && o.Y < c.Y+1 {
			h, tr = o.Y+o.H, o.truck
		}
	}
	return h, tr
}

func (c *Cargo) Update(dt float64, w *World) {
	if c.held {
		return // positioned by the crane
	}
	if t := c.truck; t != nil {
		if t.BedTilt <= 0.25 {
			c.Pos = t.Pos.Add(c.local.Rot(t.Angle))
			c.Yaw = t.Angle + c.lyaw
			return
		}
		// the body is tipping: the load slides off the back
		c.truck = nil
		c.Pos = t.Pos.Add(dir(t.Angle).Scale(truckBedHingeX - c.footRadius() - 6))
	}
	h, tr := w.support(c)
	if c.Y > h+0.01 {
		c.vy -= 400 * dt
		c.Y += c.vy * dt
		if c.Y > h {
			return
		}
	}
	landed := c.vy < -1
	c.Y, c.vy = h, 0
	if tr != nil {
		c.truck = tr
		c.local = c.Pos.Sub(tr.Pos).Rot(-tr.Angle)
		c.lyaw = c.Yaw - tr.Angle
	}
	if landed {
		w.spawnScoopDust(c.Pos, c.Yaw)
	}
}

// Collider: items on the ground are obstacles for vehicles and walkers.
func (c *Cargo) collider() (obstacle, bool) {
	if c.held || c.truck != nil || c.Y > 30 {
		return obstacle{}, false
	}
	return obstacle{c.Pos, c.footRadius() * 0.85, c}, true
}

// Render draws the item at its pose; the crane draws the slings.
func (c *Cargo) Render(g *Game, s *Scene) {
	if c.Y < 2 {
	}
	if !g.cam.inView(ground(c.Pos, c.Y+c.H/2), c.footRadius()+c.H) {
		return
	}
	at := func(x, z, y float64) Vec3 {
		return ground(c.Pos.Add(dir(c.Yaw).Scale(x)).Add(dir(c.Yaw+math.Pi/2).Scale(z)), c.Y+y)
	}
	wood := rgba(150, 110, 66, 255)
	switch c.Kind {
	case cargoBricks:
		s.addBox(at(0, 0, 2), c.W/2, 2, c.D/2, c.Yaw, wood) // pallet
		brick := rgba(170, 78, 52, 255)
		s.addBox(at(0, 0, 4+(c.H-4)/2), c.W/2-1.5, (c.H-4)/2, c.D/2-1.5, c.Yaw, brick)
		s.detail(func() {
			for k := 1; k < 5; k++ {
				y := 4 + float64(k)*(c.H-4)/5
				s.addBox(at(0, 0, y), c.W/2-1.2, 0.35, c.D/2-1.2, c.Yaw, rgba(190, 180, 170, 255)) // mortar lines
			}
			s.addBox(at(0, 0, 4+(c.H-4)/2), 0.8, (c.H-4)/2+0.4, c.D/2-1.1, c.Yaw, rgba(40, 80, 160, 255)) // strap
		})
	case cargoBeams:
		steel := rgba(96, 104, 118, 255)
		for _, z := range []float64{-7, 0, 7} {
			s.addBox(at(0, z, 1.5), c.W/2, 1.5, 3, c.Yaw, steel)   // bottom flange
			s.addBox(at(0, z, 6.5), c.W/2, 3.5, 0.8, c.Yaw, steel) // web
			s.addBox(at(0, z, 11.5), c.W/2, 1.5, 3, c.Yaw, steel)  // top flange
		}
		for _, x := range []float64{-c.W * 0.35, c.W * 0.35} {
			s.addBox(at(x, 0, c.H/2), 1.2, c.H/2+0.5, c.D/2+0.8, c.Yaw, rgba(200, 60, 40, 255)) // straps
		}
	case cargoPipe:
		s.addTube(at(0, 0, c.H/2), rotY(Vec3{1, 0, 0}, c.Yaw), c.H/2, c.W, 14, rgba(176, 172, 164, 255), rgba(70, 68, 66, 255))
	case cargoCrate:
		s.addBox(at(0, 0, c.H/2), c.W/2, c.H/2, c.D/2, c.Yaw, wood)
		s.detail(func() {
			dark := shadeC(wood, 0.72)
			for _, sx := range []float64{-1, 1} {
				for _, sz := range []float64{-1, 1} {
					s.addBox(at(sx*(c.W/2-1.5), sz*(c.D/2-1.5), c.H/2), 1.8, c.H/2+0.3, 1.8, c.Yaw, dark)
				}
			}
			s.addBox(at(0, 0, c.H-1.5), c.W/2+0.3, 1.6, c.D/2+0.3, c.Yaw, dark)
		})
	case cargoContainer:
		col := rgba(46, 110, 160, 255)
		s.addPrism(at(0, 0, 0), c.Yaw, []ring{{-c.W / 2, c.W / 2, c.D / 2, 0, 0.5}, {-c.W / 2, c.W / 2, c.D / 2, c.H, 0.5}}, col)
		s.detail(func() {
			for k := 0; k <= 10; k++ {
				x := -c.W/2 + 4 + float64(k)*(c.W-8)/10
				for _, sg := range []float64{-1, 1} {
					s.addBox(at(x, sg*(c.D/2+0.6), c.H/2), 1.4, c.H/2-2, 0.9, c.Yaw, shadeC(col, 0.86))
				}
			}
			s.addBox(at(c.W/2+0.8, 0, c.H/2), 0.8, c.H/2-1, c.D/2-1, c.Yaw, shadeC(col, 0.8)) // doors
		})
	}
}
