package main

import (
	"log"
	"math"
	"math/rand"
)

// Surface kinds painted onto the ground base layer.
const (
	surfDirt = iota
	surfGravel
	surfConcrete
	surfAsphalt
)

// Rect is an axis-aligned-ish rectangle with a rotation, used for props.
type Rect struct {
	Center Vec
	W, H   float64
	Angle  float64
}

// Prop is a static site object (building, cabin, barrier, equipment, cone).
type PropKind int

const (
	propCabin PropKind = iota
	propBuilding
	propBarrier
	propCone
	propExcavator
	propParkedLoader
	propTank
	propSign
	propModel // an imported Kenney model (Model/Scale/ModelYaw)
)

type Prop struct {
	Kind  PropKind
	Rect  Rect
	Solid bool
	Color [3]uint8
	Tag   string

	Model    string  // propModel: embedded model name, e.g. "city/building-garage"
	Scale    float64 // propModel: model units -> world units
	ModelYaw float64 // propModel: extra yaw so the model faces along Rect.Angle
}

// SurfacePatch is a painted ground region.
type SurfacePatch struct {
	Rect Rect
	Kind int
}

// RoadMark is a painted line/marking.
type RoadMark struct {
	A, B  Vec
	Width float64
	Color [4]uint8
	Dash  bool
}

// Truck is a dump truck whose bed is a delivery target.
type Truck struct {
	Pos      Vec
	Angle    float64
	Capacity float64
	Contents [numMat]float64
	Name     string
}

func (t *Truck) Total() float64 {
	var s float64
	for _, c := range t.Contents {
		s += c
	}
	return s
}

func (t *Truck) Receive(amt float64, typ int) {
	room := t.Capacity - t.Total()
	if room <= 0 {
		return
	}
	if amt > room {
		amt = room
	}
	t.Contents[typ] += amt
}

// bedRect returns the dump bed rectangle (the catch area) in world space.
func (t *Truck) bedRect() Rect {
	return Rect{Center: t.Pos.Add(dir(t.Angle).Scale(-14)), W: 74, H: 58, Angle: t.Angle}
}

// LoadingZone marks a place where piles of a given material are kept.
type LoadingZone struct {
	Pos  Vec
	Type int
	Name string
}

// World holds the whole construction site.
type World struct {
	Width, Height float64
	Grid          *Grid
	Surfaces      []SurfacePatch
	Marks         []RoadMark
	Props         []Prop
	Trucks        []*Truck
	Zones         []LoadingZone
	Particles     []Particle
	rng           *rand.Rand
}

func NewWorld() *World {
	w := &World{
		Width:  3200,
		Height: 2400,
		rng:    rand.New(rand.NewSource(1)),
	}
	w.Grid = NewGrid(w.Width, w.Height, 20)
	w.build()
	return w
}

func (w *World) build() {
	// Ground is dirt by default; lay gravel yards, concrete pads and roads.
	w.Surfaces = append(w.Surfaces,
		SurfacePatch{Rect{V(1600, 1200), 3200, 2400, 0}, surfDirt},
		// central gravel yard
		SurfacePatch{Rect{V(1600, 1200), 1500, 1100, 0}, surfGravel},
		// concrete pads
		SurfacePatch{Rect{V(640, 560), 520, 360, 0}, surfConcrete},
		SurfacePatch{Rect{V(2550, 520), 560, 380, 0}, surfConcrete},
		SurfacePatch{Rect{V(2560, 1900), 620, 420, 0}, surfConcrete},
	)

	// Main asphalt roads (horizontal + vertical spine).
	w.Surfaces = append(w.Surfaces,
		SurfacePatch{Rect{V(1600, 300), 3000, 150, 0}, surfAsphalt},
		SurfacePatch{Rect{V(1600, 2100), 3000, 150, 0}, surfAsphalt},
		SurfacePatch{Rect{V(300, 1200), 150, 1800, 0}, surfAsphalt},
		SurfacePatch{Rect{V(1600, 1200), 150, 1800, 0}, surfAsphalt},
	)

	// Road centre dashes.
	w.Marks = append(w.Marks,
		RoadMark{V(120, 300), V(3080, 300), 6, [4]uint8{240, 220, 90, 255}, true},
		RoadMark{V(120, 2100), V(3080, 2100), 6, [4]uint8{240, 220, 90, 255}, true},
		RoadMark{V(300, 320), V(300, 2080), 6, [4]uint8{240, 220, 90, 255}, true},
		RoadMark{V(1600, 320), V(1600, 2080), 6, [4]uint8{240, 220, 90, 255}, true},
	)
	// Loading bay chevrons / boxes near loading zones (hazard paint).
	for _, c := range []Vec{{780, 900}, {2400, 1500}} {
		w.Marks = append(w.Marks, RoadMark{c.Add(V(-80, -60)), c.Add(V(80, -60)), 8, [4]uint8{250, 200, 40, 200}, false})
		w.Marks = append(w.Marks, RoadMark{c.Add(V(-80, 60)), c.Add(V(80, 60)), 8, [4]uint8{250, 200, 40, 200}, false})
	}

	// Buildings & cabins.
	w.Props = append(w.Props,
		Prop{Kind: propBuilding, Rect: Rect{V(640, 560), 260, 180, 0}, Solid: true, Color: [3]uint8{200, 196, 188}, Tag: "Office"},
		Prop{Kind: propCabin, Rect: Rect{V(2550, 470), 150, 110, 0}, Solid: true, Color: [3]uint8{215, 180, 90}, Tag: "Site Cabin"},
		Prop{Kind: propCabin, Rect: Rect{V(2680, 600), 140, 100, 0.2}, Solid: true, Color: [3]uint8{90, 150, 190}, Tag: "Cabin"},
		Prop{Kind: propTank, Rect: Rect{V(500, 650), 90, 90, 0}, Solid: true, Color: [3]uint8{210, 120, 60}, Tag: "Fuel"},
	)

	// Parked construction equipment.
	w.Props = append(w.Props,
		Prop{Kind: propExcavator, Rect: Rect{V(2500, 1820), 150, 90, 2.4}, Solid: true, Color: [3]uint8{245, 190, 40}},
		Prop{Kind: propParkedLoader, Rect: Rect{V(2650, 1980), 150, 80, 0.6}, Solid: true, Color: [3]uint8{240, 170, 30}},
	)

	// Jersey barriers lining part of the yard.
	for i := 0; i < 8; i++ {
		x := 900 + float64(i)*110
		w.Props = append(w.Props, Prop{Kind: propBarrier, Rect: Rect{V(x, 760), 96, 30, 0}, Solid: true, Color: [3]uint8{220, 90, 60}})
	}
	// Traffic cones dotting a lane.
	for i := 0; i < 10; i++ {
		y := 1050 + float64(i)*70
		w.Props = append(w.Props, Prop{Kind: propCone, Rect: Rect{V(1180, y), 18, 18, 0}, Solid: false, Color: [3]uint8{245, 120, 30}})
	}
	// A warning sign.
	w.Props = append(w.Props, Prop{Kind: propSign, Rect: Rect{V(430, 1000), 40, 40, 0}, Solid: false, Color: [3]uint8{250, 210, 40}})

	// Kenney CC0 models: a workshop, parked pickups, a row of houses beyond
	// the north road, tree verges along the site edges and woodland outside.
	w.addModel("city/building-garage", V(2380, 600), 0, 0, 150, true, "Workshop")
	w.addModel("racing/vehicle-truck-yellow", V(2440, 410), 0.15, -math.Pi/2, 30, true, "")
	w.addModel("racing/vehicle-truck-red", V(830, 690), -0.4, -math.Pi/2, 30, true, "")
	w.addModel("racing/vehicle-truck-green", V(2820, 470), 1.7, -math.Pi/2, 30, true, "")
	decor := rand.New(rand.NewSource(5))
	houses := []string{"city/building-small-a", "city/building-small-b", "city/building-small-c", "city/building-small-d"}
	for i, x := 0, 230.0; x < 3050; i, x = i+1, x+270 {
		w.addModel(houses[i%len(houses)], V(x, 105), math.Pi/2, 0, 135, true, "")
	}
	trees := []string{"city/grass-trees", "city/grass-trees-tall"}
	tree := func(x, y float64) {
		w.addModel(trees[decor.Intn(2)], V(x, y), float64(decor.Intn(4))*math.Pi/2, 0, 112, true, "tree")
	}
	onRoad := func(y float64) bool { return math.Abs(y-300) < 140 || math.Abs(y-2100) < 140 }
	for y := 450.0; y < 2000; y += 112 {
		if !onRoad(y) {
			tree(56, y)
		}
	}
	for y := 60.0; y < 2350; y += 112 {
		tree(3144, y)
	}
	for x := 170.0; x < 3090; x += 112 {
		tree(x, 2290)
	}
	for _, p := range []Vec{{-260, 700}, {-260, 1500}, {-300, 2300}, {3460, 400}, {3460, 1300}, {3460, 2200}, {900, 2660}, {2100, 2680}, {3300, 2700}, {-200, -250}, {3400, -250}} {
		w.addModel("racing/decoration-forest", p, decor.Float64()*math.Pi*2, 0, 17, false, "tree")
	}

	// Dump trucks parked at bays.
	w.Trucks = append(w.Trucks,
		&Truck{Pos: V(1600, 560), Angle: 0, Capacity: 900, Name: "Truck A"},
		&Truck{Pos: V(1600, 1820), Angle: math.Pi, Capacity: 900, Name: "Truck B"},
		&Truck{Pos: V(2500, 300), Angle: 0, Capacity: 900, Name: "Truck C"},
	)

	// Loading zones with piles of different materials.
	w.Zones = append(w.Zones,
		LoadingZone{V(780, 900), MatDirt, "Dirt Stockpile"},
		LoadingZone{V(1050, 1500), MatGravel, "Gravel Heap"},
		LoadingZone{V(2400, 1500), MatSand, "Sand Pit"},
		LoadingZone{V(2250, 950), MatClay, "Clay Mound"},
	)
	for _, z := range w.Zones {
		// a cluster of overlapping heaps for a natural look
		w.Grid.AddPile(z.Pos.X, z.Pos.Y, 95, 70, z.Type)
		w.Grid.AddPile(z.Pos.X-55, z.Pos.Y+30, 60, 55, z.Type)
		w.Grid.AddPile(z.Pos.X+60, z.Pos.Y-20, 55, 50, z.Type)
		w.Grid.AddPile(z.Pos.X+10, z.Pos.Y+60, 48, 45, z.Type)
	}
	// Scatter some loose debris around the yard for pushing/messing about.
	for i := 0; i < 40; i++ {
		x := 600 + w.rng.Float64()*2000
		y := 1000 + w.rng.Float64()*900
		w.Grid.AddPile(x, y, 26+w.rng.Float64()*18, 14+w.rng.Float64()*12, MatGravel)
	}
}

// addModel places an embedded model as a prop, sizing its collision footprint
// from the model's bounds. modelYaw turns the model so it faces Rect.Angle.
func (w *World) addModel(name string, at Vec, angle, modelYaw, scale float64, solid bool, tag string) {
	m, err := loadModel(name)
	if err != nil {
		log.Println(err)
		return
	}
	sx := (m.Max.X - m.Min.X) * scale
	sz := (m.Max.Z - m.Min.Z) * scale
	if math.Abs(math.Abs(modelYaw)-math.Pi/2) < 0.1 {
		sx, sz = sz, sx // model's long axis is Z
	}
	w.Props = append(w.Props, Prop{
		Kind: propModel, Rect: Rect{at, sx, sz, angle}, Solid: solid, Tag: tag,
		Model: name, Scale: scale, ModelYaw: modelYaw,
	})
}

// CollideCircle tests a circle against solid rectangular props and trucks.
// Returns whether it hit and a push-out normal.
func (w *World) CollideCircle(c Vec, r float64) (bool, Vec) {
	check := func(rc Rect, pad float64) (bool, Vec) {
		// transform circle into rect local space
		d := c.Sub(rc.Center).Rot(-rc.Angle)
		hw := rc.W/2 + pad
		hh := rc.H/2 + pad
		// closest point on rect to circle centre
		cx := clamp(d.X, -hw, hw)
		cy := clamp(d.Y, -hh, hh)
		diff := V(d.X-cx, d.Y-cy)
		if diff.Len() < r {
			n := diff
			if n.Len() < 1e-6 {
				// centre inside: push along smallest penetration axis
				if hw-math.Abs(d.X) < hh-math.Abs(d.Y) {
					n = V(sign(d.X), 0)
				} else {
					n = V(0, sign(d.Y))
				}
			}
			return true, n.Norm().Rot(rc.Angle)
		}
		return false, Vec{}
	}
	for i := range w.Props {
		if !w.Props[i].Solid {
			continue
		}
		if hit, n := check(w.Props[i].Rect, r*0.2); hit {
			return true, n
		}
	}
	for _, t := range w.Trucks {
		if hit, n := check(Rect{t.Pos, 70, 110, t.Angle}, r*0.2); hit {
			return true, n
		}
	}
	return false, Vec{}
}

// TruckAt returns a truck whose bed contains point p, or nil.
func (w *World) TruckAt(p Vec) *Truck {
	for _, t := range w.Trucks {
		rc := t.bedRect()
		d := p.Sub(rc.Center).Rot(-rc.Angle)
		if math.Abs(d.X) < rc.W/2 && math.Abs(d.Y) < rc.H/2 {
			return t
		}
	}
	return nil
}

// Update advances world-owned simulation (particles, slow pile relaxation).
func (w *World) Update(dt float64) {
	// integrate particles
	n := 0
	for i := range w.Particles {
		p := &w.Particles[i]
		p.Life -= dt
		if p.Life <= 0 {
			continue
		}
		p.Vel = p.Vel.Add(V(0, 0)) // 2D top-down; z handled separately
		p.Z += p.VZ * dt
		p.VZ -= 220 * dt
		if p.Z < 0 {
			p.Z = 0
			p.VZ *= -0.35
			p.Vel = p.Vel.Scale(0.6)
		}
		p.Vel = p.Vel.Scale(math.Pow(p.Damp, dt))
		p.Pos = p.Pos.Add(p.Vel.Scale(dt))
		p.Spin += p.SpinV * dt
		w.Particles[n] = *p
		n++
	}
	w.Particles = w.Particles[:n]
}
