package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

// Art holds cached fonts and a 1x1 white image for solid fills.
type Art struct {
	Small text.Face
	Mid   text.Face
	Big   text.Face
	Bold  text.Face
	white *ebiten.Image
}

func mustFace(ttf []byte, size float64) text.Face {
	tt, err := opentype.Parse(ttf)
	if err != nil {
		panic(err)
	}
	f, err := opentype.NewFace(tt, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: 0})
	if err != nil {
		panic(err)
	}
	return text.NewGoXFace(f)
}

func NewArt() *Art {
	wi := ebiten.NewImage(1, 1)
	wi.Fill(color.White)
	return &Art{
		Small: mustFace(goregular.TTF, 13),
		Mid:   mustFace(goregular.TTF, 17),
		Big:   mustFace(gobold.TTF, 26),
		Bold:  mustFace(gobold.TTF, 15),
		white: wi,
	}
}

// --- value noise for ground texturing ---

func hashNoise(x, y, seed int) float64 {
	n := x*374761393 + y*668265263 + seed*362437
	n = (n ^ (n >> 13)) * 1274126177
	n = n ^ (n >> 16)
	return float64(n&0x7fffffff) / float64(0x7fffffff)
}

func valueNoise(x, y float64, seed int) float64 {
	xi, yi := math.Floor(x), math.Floor(y)
	xf, yf := x-xi, y-yi
	v00 := hashNoise(int(xi), int(yi), seed)
	v10 := hashNoise(int(xi)+1, int(yi), seed)
	v01 := hashNoise(int(xi), int(yi)+1, seed)
	v11 := hashNoise(int(xi)+1, int(yi)+1, seed)
	sx := xf * xf * (3 - 2*xf)
	sy := yf * yf * (3 - 2*yf)
	return lerp(lerp(v00, v10, sx), lerp(v01, v11, sx), sy)
}

func fbm(x, y float64, seed int) float64 {
	var sum, amp, freq float64 = 0, 0.5, 1
	for i := 0; i < 4; i++ {
		sum += valueNoise(x*freq, y*freq, seed+i*17) * amp
		freq *= 2
		amp *= 0.5
	}
	return sum
}

func surfaceBase(kind int) (r, g, b float64, speckle float64) {
	switch kind {
	case surfGravel:
		return 128, 126, 120, 46
	case surfConcrete:
		return 176, 174, 170, 16
	case surfAsphalt:
		return 70, 72, 78, 22
	default: // dirt
		return 104, 80, 54, 34
	}
}

// pointInRect reports whether world point (x,y) lies in rect rc.
func pointInRect(x, y float64, rc Rect) bool {
	d := V(x, y).Sub(rc.Center).Rot(-rc.Angle)
	return math.Abs(d.X) <= rc.W/2 && math.Abs(d.Y) <= rc.H/2
}

// buildGround bakes the entire site ground (surfaces + markings) into one
// image, built once at startup and drawn with a single transform per frame.
func buildGround(w *World) *ebiten.Image {
	iw, ih := int(w.Width), int(w.Height)
	img := image.NewRGBA(image.Rect(0, 0, iw, ih))
	rng := rand.New(rand.NewSource(99))
	_ = rng

	// Determine surface kind per pixel by testing patches top-down; patches
	// are iterated in reverse so later-defined ones win.
	for py := 0; py < ih; py++ {
		fy := float64(py)
		for px := 0; px < iw; px++ {
			fx := float64(px)
			kind := surfDirt
			for i := len(w.Surfaces) - 1; i >= 0; i-- {
				if pointInRect(fx, fy, w.Surfaces[i].Rect) {
					kind = w.Surfaces[i].Kind
					break
				}
			}
			br, bg, bb, sp := surfaceBase(kind)
			// large-scale tone variation + fine speckle
			n := fbm(fx/90, fy/90, kind*7)
			f := fbm(fx/6, fy/6, kind*13)
			tone := (n - 0.5) * 2
			grain := (f - 0.5) * 2
			r := br + tone*18 + grain*sp
			g := bg + tone*16 + grain*sp
			b := bb + tone*12 + grain*sp*0.8
			// gravel gets brighter chips
			if kind == surfGravel && f > 0.82 {
				r += 40
				g += 40
				b += 42
			}
			// asphalt subtle cracks
			if kind == surfAsphalt && fbm(fx/30, fy/30, 5) > 0.78 {
				r -= 14
				g -= 14
				b -= 14
			}
			i := (py*iw + px) * 4
			img.Pix[i+0] = clampByte(r)
			img.Pix[i+1] = clampByte(g)
			img.Pix[i+2] = clampByte(b)
			img.Pix[i+3] = 255
		}
	}

	// Bake road markings as lines.
	for _, m := range w.Marks {
		drawLineRGBA(img, m)
	}

	return ebiten.NewImageFromImage(img)
}

func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// drawLineRGBA rasterizes a (possibly dashed) thick line onto the ground.
func drawLineRGBA(img *image.RGBA, m RoadMark) {
	d := m.B.Sub(m.A)
	length := d.Len()
	if length < 1 {
		return
	}
	n := d.Norm()
	perp := V(-n.Y, n.X)
	hw := m.Width / 2
	col := color.RGBA{m.Color[0], m.Color[1], m.Color[2], m.Color[3]}
	for t := 0.0; t < length; t += 1 {
		if m.Dash && int(t/22)%2 == 1 {
			continue
		}
		c := m.A.Add(n.Scale(t))
		for o := -hw; o <= hw; o += 1 {
			p := c.Add(perp.Scale(o))
			px, py := int(p.X), int(p.Y)
			if px < 0 || py < 0 || px >= img.Rect.Dx() || py >= img.Rect.Dy() {
				continue
			}
			i := (py*img.Rect.Dx() + px) * 4
			a := float64(col.A) / 255
			img.Pix[i+0] = clampByte(float64(img.Pix[i+0])*(1-a) + float64(col.R)*a)
			img.Pix[i+1] = clampByte(float64(img.Pix[i+1])*(1-a) + float64(col.G)*a)
			img.Pix[i+2] = clampByte(float64(img.Pix[i+2])*(1-a) + float64(col.B)*a)
		}
	}
}
