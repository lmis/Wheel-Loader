package main

import (
	"math"
	"math/rand"
)

// Material types.
const (
	MatNone = iota
	MatDirt
	MatGravel
	MatSand
	MatClay
	numMat
)

var matNames = [numMat]string{"", "Dirt", "Gravel", "Sand", "Clay"}

// matColor returns the base RGB for a material type.
func matColor(t int) (r, g, b uint8) {
	switch t {
	case MatDirt:
		return 120, 82, 45
	case MatGravel:
		return 150, 150, 158
	case MatSand:
		return 208, 180, 120
	case MatClay:
		return 170, 120, 90
	default:
		return 100, 100, 100
	}
}

// Cell holds a quantity of a single dominant material type.
type Cell struct {
	Amount float64 // units of loose material heaped in this cell
	Type   int
}

// Grid is a coarse heightfield of loose material over the world.
// It is deliberately simple: each cell stores an amount and a type.
// Scooping removes amount, dumping adds amount, pushing moves amount
// between neighbouring cells. A lightweight relaxation pass makes tall
// heaps slump outward so piles look and behave convincingly.
type Grid struct {
	W, H     int
	CellSize float64
	cells    []Cell
	// active tracks cells that changed and should be relaxed soon.
	maxHeap float64
}

func NewGrid(worldW, worldH, cellSize float64) *Grid {
	w := int(math.Ceil(worldW / cellSize))
	h := int(math.Ceil(worldH / cellSize))
	return &Grid{
		W:        w,
		H:        h,
		CellSize: cellSize,
		cells:    make([]Cell, w*h),
		maxHeap:  26,
	}
}

func (g *Grid) idx(cx, cy int) int { return cy*g.W + cx }

func (g *Grid) inBounds(cx, cy int) bool {
	return cx >= 0 && cy >= 0 && cx < g.W && cy < g.H
}

func (g *Grid) cellAt(x, y float64) (int, int, bool) {
	cx := int(x / g.CellSize)
	cy := int(y / g.CellSize)
	return cx, cy, g.inBounds(cx, cy)
}

func (g *Grid) at(cx, cy int) *Cell {
	return &g.cells[g.idx(cx, cy)]
}

// Add deposits amount of type t at world position (x,y).
func (g *Grid) Add(x, y, amount float64, t int) {
	cx, cy, ok := g.cellAt(x, y)
	if !ok {
		return
	}
	c := g.at(cx, cy)
	if c.Amount < 1 {
		c.Type = t
	}
	c.Amount += amount
}

// AddPile builds a conical heap centred at (x,y).
func (g *Grid) AddPile(x, y, radius, peak float64, t int) {
	cxMin, cyMin, _ := g.cellAt(x-radius, y-radius)
	cxMax, cyMax, _ := g.cellAt(x+radius, y+radius)
	for cy := cyMin; cy <= cyMax; cy++ {
		for cx := cxMin; cx <= cxMax; cx++ {
			if !g.inBounds(cx, cy) {
				continue
			}
			px := (float64(cx) + 0.5) * g.CellSize
			py := (float64(cy) + 0.5) * g.CellSize
			d := math.Hypot(px-x, py-y)
			if d > radius {
				continue
			}
			f := 1 - d/radius
			amt := peak * f * f
			c := g.at(cx, cy)
			c.Type = t
			c.Amount += amt
		}
	}
}

// TakeAround removes up to want units from cells within radius of (x,y),
// preferring the requested type. It returns how much was taken and the
// dominant type taken. Used by the bucket when scooping.
func (g *Grid) TakeAround(x, y, radius, want float64, preferType int) (float64, int) {
	cxMin, cyMin, _ := g.cellAt(x-radius, y-radius)
	cxMax, cyMax, _ := g.cellAt(x+radius, y+radius)
	var taken float64
	typeTally := [numMat]float64{}
	for cy := cyMin; cy <= cyMax && taken < want; cy++ {
		for cx := cxMin; cx <= cxMax && taken < want; cx++ {
			if !g.inBounds(cx, cy) {
				continue
			}
			c := g.at(cx, cy)
			if c.Amount <= 0 {
				continue
			}
			px := (float64(cx) + 0.5) * g.CellSize
			py := (float64(cy) + 0.5) * g.CellSize
			if math.Hypot(px-x, py-y) > radius {
				continue
			}
			take := math.Min(c.Amount, want-taken)
			c.Amount -= take
			taken += take
			typeTally[c.Type] += take
		}
	}
	// pick dominant type taken
	best := MatNone
	var bestAmt float64
	for t := 1; t < numMat; t++ {
		if typeTally[t] > bestAmt {
			bestAmt = typeTally[t]
			best = t
		}
	}
	return taken, best
}

// Push moves material from cells near (x,y) in direction d (bulldozing).
// Material is shoved from the leading edge to cells a little further along.
func (g *Grid) Push(x, y, radius float64, d Vec, strength float64) {
	cxMin, cyMin, _ := g.cellAt(x-radius, y-radius)
	cxMax, cyMax, _ := g.cellAt(x+radius, y+radius)
	type mv struct {
		fromX, fromY, toX, toY int
		amt                    float64
		t                      int
	}
	var moves []mv
	for cy := cyMin; cy <= cyMax; cy++ {
		for cx := cxMin; cx <= cxMax; cx++ {
			if !g.inBounds(cx, cy) {
				continue
			}
			c := g.at(cx, cy)
			if c.Amount <= 0 {
				continue
			}
			px := (float64(cx) + 0.5) * g.CellSize
			py := (float64(cy) + 0.5) * g.CellSize
			if math.Hypot(px-x, py-y) > radius {
				continue
			}
			tx := cx + int(math.Round(d.X))
			ty := cy + int(math.Round(d.Y))
			if !g.inBounds(tx, ty) {
				continue
			}
			move := math.Min(c.Amount, strength)
			moves = append(moves, mv{cx, cy, tx, ty, move, c.Type})
		}
	}
	for _, m := range moves {
		from := g.at(m.fromX, m.fromY)
		to := g.at(m.toX, m.toY)
		if from.Amount < m.amt {
			m.amt = from.Amount
		}
		from.Amount -= m.amt
		if to.Amount < 1 {
			to.Type = m.t
		}
		to.Amount += m.amt
	}
}

// Relax performs a few localised slump passes so heaps above the angle of
// repose spread to neighbours. Only a window around (x,y) is processed to
// keep the cost bounded regardless of world size.
func (g *Grid) Relax(x, y, radius float64, iterations int) {
	cxMin, cyMin, _ := g.cellAt(x-radius, y-radius)
	cxMax, cyMax, _ := g.cellAt(x+radius, y+radius)
	if cxMin < 1 {
		cxMin = 1
	}
	if cyMin < 1 {
		cyMin = 1
	}
	if cxMax > g.W-2 {
		cxMax = g.W - 2
	}
	if cyMax > g.H-2 {
		cyMax = g.H - 2
	}
	for it := 0; it < iterations; it++ {
		for cy := cyMin; cy <= cyMax; cy++ {
			for cx := cxMin; cx <= cxMax; cx++ {
				c := g.at(cx, cy)
				if c.Amount <= g.maxHeap {
					continue
				}
				excess := c.Amount - g.maxHeap
				// distribute a fraction to the 4 neighbours
				share := excess * 0.22
				for _, n := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nb := g.at(cx+n[0], cy+n[1])
					if nb.Amount < 1 {
						nb.Type = c.Type
					}
					nb.Amount += share
					c.Amount -= share
				}
			}
		}
	}
}

// scatterSeed gives each cell a stable pseudo-random value for texturing.
func (g *Grid) scatterSeed(cx, cy int) float64 {
	n := cx*73856093 ^ cy*19349663
	r := rand.New(rand.NewSource(int64(n)))
	return r.Float64()
}
