package main

import (
	"image/color"
	"math"
)

// Loader geometry (world units ~ pixels at zoom 1).
const (
	rearLen   = 46 // joint -> rear axle
	frontLen  = 44 // joint -> front axle
	bodyWidth = 42
	armReach  = 66 // joint -> bucket pivot when arms down and extended
	bucketCap = 100.0

	maxSpeed    = 190.0 // forward units/sec
	maxRevSpeed = 120.0
	accelPower  = 240.0 // engine push
	brakePower  = 520.0
	rollDrag    = 1.6 // natural rolling resistance coefficient
	engineBrake = 90.0

	maxArtic    = 0.72 // radians the chassis can bend each side (~41deg)
	articRate   = 1.9  // rad/sec articulation speed
	articReturn = 1.1  // auto-centering speed when no steer input

	liftRate = 1.5 // per second (0..1)
	tiltRate = 1.9 // per second (0..1)
)

// Loader is the articulated wheel loader.
type Loader struct {
	// Kinematic state. Joint is the articulation pivot.
	Joint   Vec
	Heading float64 // chassis heading (direction of travel)
	Artic   float64 // articulation angle (bend at joint)
	Speed   float64 // signed along Heading

	// Implement state, 0..1.
	Lift float64 // 0 = on ground, 1 = fully raised
	Tilt float64 // 0 = dumped forward, 1 = curled back (carry)

	// Bucket contents.
	Load     float64
	LoadType int

	// Visual reaction state.
	Pitch   float64 // body pitch from accel/load (visual bob)
	Roll    float64 // body roll from steering
	bounce  float64
	bounceV float64

	// control intents, set by input each step
	throttle float64 // -1..1
	brake    bool
	steer    float64 // -1..1 (left/right)
	liftCmd  float64 // -1..1
	tiltCmd  float64 // -1..1

	Beacon    float64 // rotating warning light phase
	wheelSpin float64 // wheel rotation phase for animation
	engineRPM float64 // 0..1 visual/audio proxy

	Label    string     // display name (default "Wheel loader")
	Color    color.RGBA // body colour (zero = construction yellow)
	Occupied bool       // someone is in the cab (engine running, beacon on)
}

// FrontHeading / RearHeading splay around the chassis heading so the
// articulation is clearly visible on both halves.
func (l *Loader) FrontHeading() float64 { return l.Heading + l.Artic*0.60 }
func (l *Loader) RearHeading() float64  { return l.Heading - l.Artic*0.40 }

func (l *Loader) FrontAxle() Vec {
	return l.Joint.Add(dir(l.FrontHeading()).Scale(frontLen))
}
func (l *Loader) RearAxle() Vec {
	return l.Joint.Sub(dir(l.RearHeading()).Scale(rearLen))
}

// BucketPivot is where the loader arms meet the bucket. It rises (visually
// drawn smaller/offset) with lift and sits ahead of the front axle.
func (l *Loader) BucketPivot() Vec {
	fh := l.FrontHeading()
	// arms retract slightly toward the body as they lift
	reach := armReach - l.Lift*10
	return l.FrontAxle().Add(dir(fh).Scale(reach))
}

// BucketTip is the leading cutting edge of the bucket, where it bites piles.
func (l *Loader) BucketTip() Vec {
	fh := l.FrontHeading()
	p := l.BucketPivot()
	// when tilted forward (Tilt low) the edge reaches further ahead & down
	out := 16 + (1-l.Tilt)*14
	return p.Add(dir(fh).Scale(out))
}

func (l *Loader) LoadFrac() float64 { return l.Load / bucketCap }

// Update advances the loader by dt seconds using a stable fixed step.
func (l *Loader) Update(dt float64, w *World) {
	// --- Articulation (can bend even while stationary; hydraulics) ---
	if l.steer != 0 {
		l.Artic += l.steer * articRate * dt
	} else {
		// auto-center toward straight
		l.Artic = approach(l.Artic, 0, articReturn*dt)
	}
	l.Artic = clamp(l.Artic, -maxArtic, maxArtic)

	// --- Longitudinal dynamics with heavy inertia ---
	loadFactor := 1 + l.LoadFrac()*0.6 // heavier when loaded
	if l.throttle != 0 {
		l.Speed += l.throttle * (accelPower / loadFactor) * dt
		l.engineRPM = approach(l.engineRPM, 0.4+math.Abs(l.throttle)*0.6, 2*dt)
	} else {
		l.engineRPM = approach(l.engineRPM, math.Abs(l.Speed)/maxSpeed*0.5+0.12, 1.5*dt)
	}
	if l.brake {
		dec := brakePower * dt
		if l.Speed > 0 {
			l.Speed = math.Max(0, l.Speed-dec)
		} else {
			l.Speed = math.Min(0, l.Speed+dec)
		}
		l.engineRPM = approach(l.engineRPM, 0.15, 3*dt)
	}
	// rolling + engine resistance
	l.Speed -= l.Speed * rollDrag * dt
	if l.throttle == 0 && !l.brake {
		eb := engineBrake * dt
		if l.Speed > 0 {
			l.Speed = math.Max(0, l.Speed-eb)
		} else {
			l.Speed = math.Min(0, l.Speed+eb)
		}
	}
	l.Speed = clamp(l.Speed, -maxRevSpeed, maxSpeed)

	// --- Turning: articulated vehicles only turn while driving ---
	wheelbase := float64(frontLen + rearLen)
	omega := l.Speed * math.Sin(l.Artic) / wheelbase
	l.Heading += omega * dt

	// --- Move ---
	move := dir(l.Heading).Scale(l.Speed * dt)
	newJoint := l.Joint.Add(move)

	// collision against solid world objects: push back & bleed speed
	if w != nil {
		if hit, normal := collideCircles(w, l.collidersAt(newJoint), l); hit {
			newJoint = newJoint.Add(normal.Scale(3))
			// reflect a little & lose energy -> visible lurch
			l.Speed *= -0.25
			l.bounceV += 120
			l.Pitch += sign(l.Speed) * 0.05
		}
	}
	l.Joint = newJoint
	if w != nil {
		l.Joint.X = clamp(l.Joint.X, 40, w.Width-40)
		l.Joint.Y = clamp(l.Joint.Y, 40, w.Height-40)
	}

	// --- Implements ---
	if l.liftCmd != 0 {
		l.Lift = clamp(l.Lift+l.liftCmd*liftRate*dt, 0, 1)
	}
	if l.tiltCmd != 0 {
		l.Tilt = clamp(l.Tilt+l.tiltCmd*tiltRate*dt, 0, 1)
	}

	// --- Visual reactions ---
	// pitch from acceleration (squat/dive), scaled by load
	targetPitch := -l.throttle * 0.06 * loadFactor
	if l.brake {
		targetPitch = sign(l.Speed) * 0.08
	}
	l.Pitch = lerp(l.Pitch, targetPitch, 6*dt)
	// roll from turning
	targetRoll := -omega * 2.2
	l.Roll = lerp(l.Roll, targetRoll, 6*dt)
	// suspension bounce spring
	l.bounceV += -l.bounce * 60 * dt
	l.bounceV *= math.Pow(0.02, dt)
	l.bounce += l.bounceV * dt
	// wheels & beacon
	l.wheelSpin += l.Speed * dt * 0.08
	l.Beacon += dt * 3.2

	// --- Material interaction ---
	l.interactMaterial(dt, w)
}

// interactMaterial handles scooping, carrying, spilling and dumping.
func (l *Loader) interactMaterial(dt float64, w *World) {
	if w == nil {
		return
	}
	g := w.Grid
	tip := l.BucketTip()
	fh := l.FrontHeading()
	grounded := l.Lift < 0.28 // bucket low enough to engage ground

	// SCOOP: bucket low, curling back (tilt increasing) or driving forward
	// into material. Material flows from the pile into the bucket.
	if grounded && l.Load < bucketCap {
		// rate depends on forward motion into pile and on curling the bucket
		drive := math.Max(0, l.Speed) / maxSpeed
		curl := math.Max(0, l.tiltCmd) // actively curling scoops hard
		rate := (drive*70 + curl*90) * dt
		if rate > 0 {
			want := math.Min(rate, bucketCap-l.Load)
			got, t := g.TakeAround(tip.X, tip.Y, 26, want, l.LoadType)
			if got > 0 {
				if l.Load < 1 {
					l.LoadType = t
				} else if t != l.LoadType {
					// mixing: keep original type label but still accept
				}
				l.Load += got
				w.spawnScoopDust(tip, fh)
			}
		}
	}

	// PUSH / bulldoze: bucket low and tilted forward shoves loose material.
	if grounded && l.Tilt < 0.45 && math.Abs(l.Speed) > 20 {
		d := dir(fh)
		// grid-step direction
		gd := V(math.Round(d.X), math.Round(d.Y))
		g.Push(tip.X, tip.Y, 24, gd, 40*dt*math.Abs(l.Speed)/maxSpeed)
	}

	// SPILL: carrying material but bucket tilted forward (not actively
	// dumping) sheds load gradually — a badly held bucket loses material.
	if l.Load > 0 && l.Tilt < 0.4 && l.tiltCmd >= 0 {
		spillRisk := (0.4 - l.Tilt) * 2.2
		lost := math.Min(spillRisk*55*dt, l.Load)
		l.discharge(w, lost, 6)
	}

	// DUMP: deliberately tipping forward empties the bucket fast.
	if l.Load > 0 && l.Tilt < 0.3 && l.tiltCmd < 0 {
		dumped := math.Min(160*dt, l.Load)
		l.discharge(w, dumped, 18)
	}
}

// discharge removes amount from the bucket and deposits it ahead of the
// cutting edge, into a truck bed if one is there, otherwise onto the ground.
func (l *Loader) discharge(w *World, amount, forward float64) {
	if amount <= 0 {
		return
	}
	l.Load -= amount
	fh := l.FrontHeading()
	drop := l.BucketTip().Add(dir(fh).Scale(forward - 10))
	// check a couple of points along the discharge arc for a truck bed
	var tr *Truck
	for _, p := range []Vec{drop, l.BucketTip(), l.BucketPivot()} {
		if tr = w.TruckAt(p); tr != nil {
			break
		}
	}
	if tr != nil {
		tr.Receive(amount, l.LoadType)
	} else {
		w.Grid.Add(drop.X, drop.Y, amount, l.LoadType)
		w.Grid.Relax(drop.X, drop.Y, 60, 1)
	}
	w.spawnMaterialFall(l.BucketPivot(), fh, amount, l.LoadType)
}
