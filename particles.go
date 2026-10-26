package main

import "math"

// Particle kinds change how they are drawn.
const (
	pfxDust = iota
	pfxMaterial
	pfxSmoke
	pfxSpark
	pfxConfetti
)

// Particle is a cheap billboard with a fake height (Z) for a 2.5D feel.
type Particle struct {
	Pos   Vec
	Vel   Vec
	Z     float64
	VZ    float64
	Life  float64
	Max   float64
	Size  float64
	Damp  float64
	Spin  float64
	SpinV float64
	Kind  int
	Color [3]uint8
}

func (w *World) addParticle(p Particle) {
	if len(w.Particles) > 1400 {
		return // cap
	}
	if p.Max == 0 {
		p.Max = p.Life
	}
	if p.Damp == 0 {
		p.Damp = 0.1
	}
	w.Particles = append(w.Particles, p)
}

// spawnScoopDust puffs dust where the bucket bites a pile.
func (w *World) spawnScoopDust(at Vec, heading float64) {
	for i := 0; i < 2; i++ {
		a := heading + math.Pi + (w.rng.Float64()-0.5)*1.2
		sp := 15 + w.rng.Float64()*35
		w.addParticle(Particle{
			Pos:   at,
			Vel:   dir(a).Scale(sp),
			Z:     3,
			VZ:    18 + w.rng.Float64()*22,
			Life:  0.4 + w.rng.Float64()*0.4,
			Size:  3 + w.rng.Float64()*4,
			Damp:  0.2,
			Kind:  pfxDust,
			Color: [3]uint8{180, 164, 136},
		})
	}
}

// spawnMaterialFall drops chunks of material from the bucket.
func (w *World) spawnMaterialFall(at Vec, heading float64, amount float64, typ int) {
	r, g, b := matColor(typ)
	n := int(clamp(amount*0.4, 1, 6))
	for i := 0; i < n; i++ {
		a := heading + (w.rng.Float64()-0.5)*0.9
		sp := 20 + w.rng.Float64()*40
		w.addParticle(Particle{
			Pos:   at,
			Vel:   dir(a).Scale(sp),
			Z:     22 + w.rng.Float64()*8,
			VZ:    -10,
			Life:  0.7 + w.rng.Float64()*0.4,
			Size:  4 + w.rng.Float64()*4,
			Damp:  0.5,
			SpinV: (w.rng.Float64() - 0.5) * 10,
			Kind:  pfxMaterial,
			Color: [3]uint8{r, g, b},
		})
	}
}

// spawnDriveDust kicks dust from the wheels when driving on loose ground.
func (w *World) spawnDriveDust(at Vec, heading, intensity float64) {
	if w.rng.Float64() > intensity*0.5 {
		return
	}
	a := heading + math.Pi + (w.rng.Float64()-0.5)*0.8
	w.addParticle(Particle{
		Pos:   at,
		Vel:   dir(a).Scale(8 + w.rng.Float64()*18),
		Z:     1,
		VZ:    6 + w.rng.Float64()*10,
		Life:  0.4 + w.rng.Float64()*0.4,
		Size:  3 + w.rng.Float64()*3,
		Damp:  0.3,
		Kind:  pfxDust,
		Color: [3]uint8{156, 142, 120},
	})
}

// spawnExhaust puffs engine smoke from the loader's stack.
func (w *World) spawnExhaust(at Vec, rpm float64) { w.spawnExhaustAt(at, 60, rpm) }

// spawnExhaustAt puffs engine smoke from a stack whose top is at height z.
func (w *World) spawnExhaustAt(at Vec, z, rpm float64) {
	if w.rng.Float64() > rpm*0.5+0.05 {
		return
	}
	w.addParticle(Particle{
		Pos:   at.Add(V((w.rng.Float64()-0.5)*4, (w.rng.Float64()-0.5)*4)),
		Vel:   V((w.rng.Float64()-0.5)*8, (w.rng.Float64()-0.5)*8),
		Z:     z,
		VZ:    18 + w.rng.Float64()*12,
		Life:  0.9 + w.rng.Float64()*0.8,
		Size:  5 + w.rng.Float64()*4,
		Damp:  0.4,
		Kind:  pfxSmoke,
		Color: [3]uint8{60, 60, 64},
	})
}

// confettiColors are the bright party colours for celebrations.
var confettiColors = [][3]uint8{
	{255, 70, 70}, {255, 200, 40}, {70, 200, 255}, {120, 230, 90}, {240, 110, 230}, {255, 140, 40},
}

// spawnConfetti throws a burst of colourful confetti up from a point.
func (w *World) spawnConfetti(at Vec, height float64, n int) {
	for i := 0; i < n; i++ {
		a := w.rng.Float64() * 2 * math.Pi
		w.addParticle(Particle{
			Pos:   at.Add(dir(a).Scale(w.rng.Float64() * 10)),
			Vel:   dir(a).Scale(20 + w.rng.Float64()*70),
			Z:     height,
			VZ:    120 + w.rng.Float64()*140,
			Life:  2 + w.rng.Float64()*1.2,
			Size:  1.4 + w.rng.Float64()*1.4,
			Damp:  0.5,
			SpinV: (w.rng.Float64() - 0.5) * 20,
			Kind:  pfxConfetti,
			Color: confettiColors[w.rng.Intn(len(confettiColors))],
		})
	}
}
