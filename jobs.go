package main

import (
	"fmt"
	"math/rand"
)

// Job asks the player to deliver a quantity of a material to a truck.
type Job struct {
	Material  int
	Target    float64 // units required
	Delivered float64
	Truck     *Truck
	Reward    int
	Done      bool
	startTot  float64 // truck total of this material when job began
}

func (j *Job) Progress() float64 {
	if j.Target <= 0 {
		return 1
	}
	return clamp(j.Delivered/j.Target, 0, 1)
}

func (j *Job) Describe() string {
	return fmt.Sprintf("Deliver %.0f units of %s to %s", j.Target, matNames[j.Material], j.Truck.Name)
}

// JobBoard manages the current job, progression and money.
type JobBoard struct {
	Current *Job
	Money   int
	Level   int
	Jobs    int
	rng     *rand.Rand
	world   *World
	flash   string
	flashT  float64
}

func NewJobBoard(w *World) *JobBoard {
	jb := &JobBoard{Money: 0, Level: 1, rng: rand.New(rand.NewSource(7)), world: w}
	jb.next()
	return jb
}

func (jb *JobBoard) next() {
	w := jb.world
	mat := 1 + jb.rng.Intn(numMat-1)
	truck := w.Trucks[jb.rng.Intn(len(w.Trucks))]
	target := 150 + float64(jb.Level)*60 + jb.rng.Float64()*120
	reward := int(target*1.5) + jb.Level*40
	jb.Current = &Job{
		Material: mat,
		Target:   target,
		Truck:    truck,
		Reward:   reward,
		startTot: truck.Contents[mat],
	}
}

// Update tracks delivery progress based on what the truck has received.
func (jb *JobBoard) Update(dt float64) {
	if jb.flashT > 0 {
		jb.flashT -= dt
	}
	j := jb.Current
	if j == nil || j.Done {
		return
	}
	j.Delivered = j.Truck.Contents[j.Material] - j.startTot
	if j.Delivered >= j.Target {
		j.Done = true
		jb.Money += j.Reward
		jb.Jobs++
		if jb.Jobs%2 == 0 {
			jb.Level++
		}
		jb.flash = fmt.Sprintf("JOB COMPLETE  +$%d", j.Reward)
		jb.flashT = 4
		jb.next()
	}
}
