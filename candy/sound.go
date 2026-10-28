package main

import (
	"encoding/binary"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Little cartoon noises, made up on the spot from sine waves and noise so
// there are no sound files: boings, pops, chomps and a very small RAWR.

type Sounds struct {
	ok bool
	s  map[string]rl.Sound
}

const rate = 22050

func synth(dur float64, f func(t, u float64) float64) rl.Sound {
	n := int(dur * rate)
	data := make([]byte, n*2)
	for i := 0; i < n; i++ {
		t := float64(i) / rate
		v := f(t, t/dur)
		// soften the very start and end so nothing clicks
		v *= clamp(t/0.005, 0, 1) * clamp((dur-t)/0.02, 0, 1)
		binary.LittleEndian.PutUint16(data[i*2:], uint16(int16(clamp(v, -1, 1)*26000)))
	}
	return rl.LoadSoundFromWave(rl.NewWave(uint32(n), rate, 16, 1, data))
}

// tone is a sine whose frequency follows fr(u), with a phase accumulator so
// sweeps stay smooth.
func tone(dur float64, fr func(u float64) float64, env func(u float64) float64) rl.Sound {
	ph := 0.0
	return synth(dur, func(t, u float64) float64 {
		ph += 2 * math.Pi * fr(u) / rate
		return math.Sin(ph) * env(u)
	})
}

func decay(k float64) func(float64) float64 {
	return func(u float64) float64 { return math.Exp(-k * u) }
}

func loadSounds() *Sounds {
	rl.InitAudioDevice()
	s := &Sounds{ok: rl.IsAudioDeviceReady(), s: map[string]rl.Sound{}}
	if !s.ok {
		return s
	}
	rng := rand.New(rand.NewSource(1))
	noise := func() float64 { return rng.Float64()*2 - 1 }
	s.s["boing"] = tone(0.6, func(u float64) float64 { return 180 + 260*u + 60*math.Sin(u*60)*(1-u) }, decay(3))
	s.s["pop"] = tone(0.12, func(u float64) float64 { return 900 - 600*u }, decay(6))
	s.s["click"] = tone(0.06, func(u float64) float64 { return 1400 }, decay(10))
	s.s["bloop"] = tone(0.18, func(u float64) float64 { return 300 + 500*u }, decay(4))
	s.s["squeak"] = tone(0.15, func(u float64) float64 { return 1200 + 600*math.Sin(u*math.Pi) }, decay(3))
	s.s["hop"] = tone(0.09, func(u float64) float64 { return 260 + 300*u }, decay(5))
	s.s["slide"] = tone(1.0, func(u float64) float64 { return 300 + 900*u*u + 30*math.Sin(u*80) }, func(u float64) float64 { return 0.8 })
	s.s["whistle"] = tone(0.9, func(u float64) float64 { return 1500 - 900*u }, func(u float64) float64 { return 0.5 })
	s.s["whee"] = tone(0.7, func(u float64) float64 { return 500 + 700*math.Sin(u*math.Pi) }, decay(1.5))
	s.s["wind"] = tone(0.6, func(u float64) float64 { return 120 + 300*u + 40*math.Sin(u*120) }, func(u float64) float64 { return 0.6 * u })
	s.s["blow"] = tone(2.5, func(u float64) float64 { return 140 + 160*u + 15*math.Sin(u*90) }, func(u float64) float64 { return 0.35 })
	s.s["chomp"] = synth(0.09, func(t, u float64) float64 { return noise() * math.Exp(-u*5) * 0.5 })
	s.s["thump"] = synth(0.4, func(t, u float64) float64 {
		return (math.Sin(2*math.Pi*(70-40*u)*t) + noise()*0.3*math.Exp(-u*12)) * math.Exp(-u*5)
	})
	s.s["burp"] = synth(0.5, func(t, u float64) float64 {
		f := 90 - 30*u + 12*math.Sin(t*55)
		return (math.Sin(2*math.Pi*f*t) + 0.4*math.Sin(4*math.Pi*f*t) + 0.15*noise()) * (1 - u) * 0.8
	})
	s.s["rawr"] = synth(0.8, func(t, u float64) float64 {
		f := 330 - 120*u + 25*math.Sin(t*70)
		saw := 2*math.Mod(f*t, 1) - 1
		return (0.5*saw + 0.3*noise()) * math.Sin(u*math.Pi) * 0.8
	})
	// melodies: a little tune for the dance, a giggle, a yum
	notes := func(freqs []float64, each float64) rl.Sound {
		return synth(each*float64(len(freqs)), func(t, u float64) float64 {
			i := int(t / each)
			if i >= len(freqs) {
				return 0
			}
			lt := t - float64(i)*each
			f := freqs[i]
			return (math.Sin(2*math.Pi*f*t) + 0.3*math.Sin(4*math.Pi*f*t)) * math.Exp(-lt*6) * 0.6
		})
	}
	s.s["tune"] = notes([]float64{523, 659, 784, 659, 523, 659, 784, 1047, 784, 659, 784, 523, 587, 659, 523, 0}, 0.25)
	s.s["giggle"] = notes([]float64{880, 988, 880, 1047, 880, 1175}, 0.09)
	s.s["yum"] = notes([]float64{659, 880, 1319}, 0.12)
	return s
}

func (s *Sounds) play(name string, pitch float64) {
	if s == nil || !s.ok {
		return
	}
	snd, ok := s.s[name]
	if !ok {
		return
	}
	rl.SetSoundPitch(snd, float32(pitch))
	rl.PlaySound(snd)
}

func (s *Sounds) close() {
	if s == nil || !s.ok {
		return
	}
	for _, snd := range s.s {
		rl.UnloadSound(snd)
	}
	rl.CloseAudioDevice()
}
