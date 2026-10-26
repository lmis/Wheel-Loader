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
	Huge  text.Face
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
		Huge:  mustFace(gobold.TTF, 44),
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

// Object-pass texture atlas layout. Flat-coloured geometry samples the white
// block; loose-material heaps sample the periodic grain region.
const (
	atlasW, atlasH = 512, 256
	atlasWhiteUV   = 4.0   // centre of the white block (0..8)
	grainX0        = 256.0 // grain region origin in the atlas
	grainPeriod    = 200.0 // the grain texture repeats every grainPeriod texels
	grainSpan      = 40.0  // texels per heap cell (cell size 20 * 2 texels/unit)
	grainMean      = 0.84  // average brightness of the grain, compensated in colour
)

// buildAtlas makes the object-pass texture: a white block plus a seamless,
// neutral-grey granular texture (fine grit + scattered pebbles with lit tops
// and dark rims). The grain is periodic with an extra grainSpan of wrapped
// texels, so any heap cell maps onto it without seams.
func buildAtlas() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, atlasW, atlasH))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			i := (y*atlasW + x) * 4
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 255, 255, 255, 255
		}
	}
	const P = int(grainPeriod)
	// periodic value noise: lattice coords wrap, cell size divides P
	pnoise := func(x, y float64, cell, seed int) float64 {
		n := P / cell
		gx, gy := x/float64(cell), y/float64(cell)
		x0, y0 := int(math.Floor(gx)), int(math.Floor(gy))
		fx, fy := gx-float64(x0), gy-float64(y0)
		h := func(i, j int) float64 { return hashNoise(((i%n)+n)%n, ((j%n)+n)%n, seed) }
		sx, sy := fx*fx*(3-2*fx), fy*fy*(3-2*fy)
		return lerp(lerp(h(x0, y0), h(x0+1, y0), sx), lerp(h(x0, y0+1), h(x0+1, y0+1), sx), sy)
	}
	// pebbles: one jittered point per 10px cell, toroidal distance
	const pc = 10
	pebble := func(x, y int) float64 {
		best, shade := 1e9, 0.0
		cx, cy := x/pc, y/pc
		n := P / pc
		for dj := -1; dj <= 1; dj++ {
			for di := -1; di <= 1; di++ {
				i, j := ((cx+di)%n+n)%n, ((cy+dj)%n+n)%n
				if hashNoise(i, j, 77) < 0.35 {
					continue // not every cell has a pebble
				}
				px := float64((cx+di)*pc) + hashNoise(i, j, 78)*pc
				py := float64((cy+dj)*pc) + hashNoise(i, j, 79)*pc
				r := 2.5 + hashNoise(i, j, 80)*3
				dx, dy := float64(x)+0.5-px, float64(y)+0.5-py
				d := math.Hypot(dx, dy) / r
				if d < 1 && d < best {
					best = d
					lightDot := (-dx - dy) / (r * 1.4) // lit from the upper left
					shade = 0.12*lightDot - 0.28*d*d*d*d + 0.05*(hashNoise(i, j, 81)-0.5)
				}
			}
		}
		return shade
	}
	for y := 0; y < P; y++ {
		for x := 0; x < P; x++ {
			fx, fy := float64(x), float64(y)
			v := grainMean +
				0.10*(pnoise(fx, fy, 2, 3)-0.5) +
				0.08*(pnoise(fx, fy, 5, 4)-0.5) +
				0.06*(pnoise(fx, fy, 20, 5)-0.5) +
				pebble(x, y)
			b := clampByte(v * 255)
			// write the texel and its wrapped copies in the extension band
			for _, q := range [][2]int{{x, y}, {x + P, y}, {x, y + P}, {x + P, y + P}} {
				if q[0] >= P+int(grainSpan) || q[1] >= P+int(grainSpan) {
					continue
				}
				i := (q[1]*atlasW + int(grainX0) + q[0]) * 4
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = b, b, b, 255
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

// buildDirtTile makes a seamless repeating dirt texture for the land around
// the site: a 256px patch in the same style as the baked ground, mirrored in
// both axes so its edges always match when repeated.
func buildDirtTile() *ebiten.Image {
	const n = 256
	img := image.NewRGBA(image.Rect(0, 0, 2*n, 2*n))
	br, bg, bb, sp := surfaceBase(surfDirt)
	for py := 0; py < n; py++ {
		for px := 0; px < n; px++ {
			fx, fy := float64(px), float64(py)
			tone := (fbm(fx/90, fy/90, surfDirt*7) - 0.5) * 2
			grain := (fbm(fx/6, fy/6, surfDirt*13) - 0.5) * 2
			c := [3]uint8{
				clampByte(br + tone*18 + grain*sp),
				clampByte(bg + tone*16 + grain*sp),
				clampByte(bb + tone*12 + grain*sp*0.8),
			}
			for _, q := range [4][2]int{{px, py}, {2*n - 1 - px, py}, {px, 2*n - 1 - py}, {2*n - 1 - px, 2*n - 1 - py}} {
				i := (q[1]*2*n + q[0]) * 4
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c[0], c[1], c[2], 255
			}
		}
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
