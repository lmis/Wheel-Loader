package main

import "math"

// Vec is a simple 2D vector used throughout the simulation.
type Vec struct {
	X, Y float64
}

func V(x, y float64) Vec { return Vec{x, y} }

func (a Vec) Add(b Vec) Vec       { return Vec{a.X + b.X, a.Y + b.Y} }
func (a Vec) Sub(b Vec) Vec       { return Vec{a.X - b.X, a.Y - b.Y} }
func (a Vec) Scale(s float64) Vec { return Vec{a.X * s, a.Y * s} }
func (a Vec) Dot(b Vec) float64   { return a.X*b.X + a.Y*b.Y }
func (a Vec) Len() float64        { return math.Hypot(a.X, a.Y) }

func (a Vec) Norm() Vec {
	l := a.Len()
	if l < 1e-9 {
		return Vec{}
	}
	return Vec{a.X / l, a.Y / l}
}

// Rot returns the vector rotated by angle a (radians).
func (v Vec) Rot(a float64) Vec {
	s, c := math.Sincos(a)
	return Vec{v.X*c - v.Y*s, v.X*s + v.Y*c}
}

// Angle returns the direction of the vector.
func (v Vec) Angle() float64 { return math.Atan2(v.Y, v.X) }

// dir returns a unit vector pointing along angle a.
func dir(a float64) Vec {
	s, c := math.Sincos(a)
	return Vec{c, s}
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// approach moves cur toward target by at most step.
func approach(cur, target, step float64) float64 {
	if cur < target {
		return math.Min(cur+step, target)
	}
	return math.Max(cur-step, target)
}

// angleDiff returns the smallest signed difference a-b wrapped to [-pi,pi].
func angleDiff(a, b float64) float64 {
	d := math.Mod(a-b+math.Pi, 2*math.Pi)
	if d < 0 {
		d += 2 * math.Pi
	}
	return d - math.Pi
}

func sign(x float64) float64 {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}
