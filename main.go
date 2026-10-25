package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	demo := flag.Bool("demo", false, "run the autopilot (hands-free demo)")
	shotDir := flag.String("shots", "", "run headless autopilot and save screenshots to this dir, then exit")
	dumpShot := flag.String("dumpshot", "", "render a loaded loader dumping into a truck to this dir, then exit")
	tourDir := flag.String("tour", "", "render static views of site landmarks to this dir, then exit")
	frames := flag.Int("frames", 0, "exit after this many frames (0 = run forever)")
	flag.Parse()

	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Wheel Loader — Construction Site")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)

	g := NewGame()

	if *demo {
		g.auto = NewAutopilot()
	}
	if *frames > 0 {
		g.exitFrame = *frames
	}
	if *shotDir != "" {
		if err := os.MkdirAll(*shotDir, 0o755); err != nil {
			log.Fatal(err)
		}
		g.auto = NewAutopilot()
		g.shotPlan(*shotDir)
	}
	if *dumpShot != "" {
		if err := os.MkdirAll(*dumpShot, 0o755); err != nil {
			log.Fatal(err)
		}
		g.dumpScene(*dumpShot)
	}
	if *tourDir != "" {
		if err := os.MkdirAll(*tourDir, 0o755); err != nil {
			log.Fatal(err)
		}
		g.tour = []tourStop{
			{frame: 8, pos: V(640, 620), zoom: 1.2, shot: filepath.Join(*tourDir, "t1_office.png")},
			{frame: 24, pos: V(1050, 820), zoom: 1.1, shot: filepath.Join(*tourDir, "t2_barriers.png")},
			{frame: 40, pos: V(2580, 1900), zoom: 1.2, shot: filepath.Join(*tourDir, "t3_equipment.png")},
		}
		g.exitFrame = 60
	}

	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
	if g.done {
		fmt.Println("finished")
	}
}

// shotPlan schedules screenshots across a full autopilot job cycle. Shots
// are keyed by target simulation step; with simSpeed substeps per rendered
// frame, software-GL capture reaches each state in far fewer draws.
func (g *Game) shotPlan(dir string) {
	g.simSpeed = 6
	plan := map[int]string{
		60:   "01_start.png",
		260:  "02_driving.png",
		520:  "03_approach_pile.png",
		680:  "04_digging.png",
		760:  "05_loaded.png",
		1080: "06_to_truck.png",
		1380: "07_dumping.png",
		1600: "08_cycle.png",
	}
	last := 0
	for simStep, name := range plan {
		f := simStep / g.simSpeed
		g.shots[f] = filepath.Join(dir, name)
		if f > last {
			last = f
		}
	}
	g.exitFrame = last + 10
}

// dumpScene positions a loaded loader just in front of a truck and tips the
// bucket, producing screenshots of material being delivered into the bed.
func (g *Game) dumpScene(outDir string) {
	g.simSpeed = 2
	truck := g.world.Trucks[2] // Truck C
	l := g.loader
	l.Joint = truck.Pos.Add(dir(truck.Angle).Scale(118))
	l.Heading = truck.Pos.Sub(l.Joint).Angle()
	l.Lift = 0.82
	l.Tilt = 0.92
	l.Load = 95
	l.LoadType = MatSand
	g.jobs.Current.Truck = truck
	g.jobs.Current.Material = MatSand
	g.jobs.Current.startTot = truck.Contents[MatSand]
	ap := NewAutopilot()
	ap.state = apDump
	ap.truck = truck
	g.auto = ap
	g.cam.Pos = l.Joint
	g.cam.TgtZoom = 1.5
	g.cam.Zoom = 1.5
	g.shots[16] = filepath.Join(outDir, "dump_1.png")
	g.shots[40] = filepath.Join(outDir, "dump_2.png")
	g.shots[70] = filepath.Join(outDir, "dump_3.png")
	g.exitFrame = 90
}

func (g *Game) maybeCapture(screen *ebiten.Image) {
	// Draw may skip frame numbers when Update outruns it headlessly, so
	// capture the first draw at or past each scheduled frame.
	target := -1
	name := ""
	for f, n := range g.shots {
		if g.frame >= f && (target == -1 || f < target) {
			target, name = f, n
		}
	}
	if target < 0 {
		return
	}
	delete(g.shots, target)
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	buf := make([]byte, 4*w*h)
	screen.ReadPixels(buf)
	img := &image.RGBA{Pix: buf, Stride: 4 * w, Rect: image.Rect(0, 0, w, h)}
	f, err := os.Create(name)
	if err != nil {
		log.Println("shot:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Println("shot encode:", err)
		return
	}
	fmt.Printf("captured %s (frame %d)  loader spd=%.1f load=%.0f lift=%.2f tilt=%.2f pos=%.0f,%.0f\n",
		name, g.frame, g.loader.Speed, g.loader.Load, g.loader.Lift, g.loader.Tilt, g.loader.Joint.X, g.loader.Joint.Y)
}
