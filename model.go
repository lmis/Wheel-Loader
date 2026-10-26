package main

import (
	"bytes"
	"embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"path"
)

// Kenney CC0 models (see assets/kenney/LICENSE.txt), embedded so the game
// runs from a single binary.
//
//go:embed assets/kenney
var assetFS embed.FS

// Mesh is a static triangle soup with baked per-vertex colors, in model
// units (Kenney: 1 unit = 1 tile), Y up.
type Mesh struct {
	Pos      []Vec3 // 3 entries per triangle
	Norm     []Vec3
	Col      []color.RGBA
	Min, Max Vec3
}

var meshCache = map[string]*Mesh{}

// loadModel returns the named embedded model (e.g. "city/building-garage"),
// loading and caching it on first use.
func loadModel(name string) (*Mesh, error) {
	if m, ok := meshCache[name]; ok {
		return m, nil
	}
	m, err := loadGLB("assets/kenney/" + name + ".glb")
	if err != nil {
		return nil, fmt.Errorf("model %s: %w", name, err)
	}
	meshCache[name] = m
	return m, nil
}

type gltfDoc struct {
	Scene  int `json:"scene"`
	Scenes []struct {
		Nodes []int `json:"nodes"`
	} `json:"scenes"`
	Nodes []struct {
		Mesh        *int      `json:"mesh"`
		Children    []int     `json:"children"`
		Translation []float64 `json:"translation"`
		Rotation    []float64 `json:"rotation"`
		Scale       []float64 `json:"scale"`
		Matrix      []float64 `json:"matrix"`
	} `json:"nodes"`
	Meshes []struct {
		Primitives []struct {
			Attributes map[string]int `json:"attributes"`
			Indices    *int           `json:"indices"`
			Material   *int           `json:"material"`
			Mode       *int           `json:"mode"`
		} `json:"primitives"`
	} `json:"meshes"`
	Materials []struct {
		PBR struct {
			BaseColorFactor  []float64 `json:"baseColorFactor"`
			BaseColorTexture *struct {
				Index int `json:"index"`
			} `json:"baseColorTexture"`
		} `json:"pbrMetallicRoughness"`
	} `json:"materials"`
	Textures []struct {
		Source int `json:"source"`
	} `json:"textures"`
	Images []struct {
		URI string `json:"uri"`
	} `json:"images"`
	Accessors []struct {
		BufferView    *int   `json:"bufferView"`
		ByteOffset    int    `json:"byteOffset"`
		ComponentType int    `json:"componentType"`
		Count         int    `json:"count"`
		Type          string `json:"type"`
		Normalized    bool   `json:"normalized"`
	} `json:"accessors"`
	BufferViews []struct {
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
		ByteStride int `json:"byteStride"`
	} `json:"bufferViews"`
}

type glb struct {
	doc  gltfDoc
	bin  []byte
	dir  string
	imgs map[int]image.Image
}

func loadGLB(p string) (*Mesh, error) {
	data, err := assetFS.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if len(data) < 20 || string(data[:4]) != "glTF" {
		return nil, errors.New("not a GLB file")
	}
	g := &glb{dir: path.Dir(p), imgs: map[int]image.Image{}}
	off := 12
	for off+8 <= len(data) {
		clen := int(binary.LittleEndian.Uint32(data[off:]))
		ctype := string(data[off+4 : off+8])
		if off+8+clen > len(data) {
			return nil, errors.New("truncated chunk")
		}
		chunk := data[off+8 : off+8+clen]
		switch ctype {
		case "JSON":
			if err := json.Unmarshal(chunk, &g.doc); err != nil {
				return nil, err
			}
		case "BIN\x00":
			g.bin = chunk
		}
		off += 8 + clen
	}

	m := &Mesh{Min: Vec3{math.Inf(1), math.Inf(1), math.Inf(1)}, Max: Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}}
	var roots []int
	if len(g.doc.Scenes) > 0 {
		roots = g.doc.Scenes[g.doc.Scene].Nodes
	} else {
		for i := range g.doc.Nodes {
			roots = append(roots, i)
		}
	}
	for _, r := range roots {
		if err := g.walk(r, mat4Identity(), m); err != nil {
			return nil, err
		}
	}
	if len(m.Pos) == 0 {
		return nil, errors.New("no triangles")
	}
	return m, nil
}

func (g *glb) walk(ni int, parent mat4, m *Mesh) error {
	if ni < 0 || ni >= len(g.doc.Nodes) {
		return errors.New("bad node index")
	}
	n := g.doc.Nodes[ni]
	local := mat4Identity()
	if len(n.Matrix) == 16 {
		copy(local[:], n.Matrix)
	} else {
		t, r, s := [3]float64{}, [4]float64{0, 0, 0, 1}, [3]float64{1, 1, 1}
		copy(t[:], n.Translation)
		if len(n.Rotation) == 4 {
			copy(r[:], n.Rotation)
		}
		if len(n.Scale) == 3 {
			copy(s[:], n.Scale)
		}
		local = mat4TRS(t, r, s)
	}
	world := parent.mul(local)
	if n.Mesh != nil {
		if err := g.addMesh(*n.Mesh, world, m); err != nil {
			return err
		}
	}
	for _, c := range n.Children {
		if err := g.walk(c, world, m); err != nil {
			return err
		}
	}
	return nil
}

func (g *glb) addMesh(mi int, xf mat4, m *Mesh) error {
	if mi < 0 || mi >= len(g.doc.Meshes) {
		return errors.New("bad mesh index")
	}
	for _, prim := range g.doc.Meshes[mi].Primitives {
		if prim.Mode != nil && *prim.Mode != 4 {
			continue // triangles only
		}
		pi, ok := prim.Attributes["POSITION"]
		if !ok {
			continue
		}
		pos, err := g.accessor(pi)
		if err != nil {
			return err
		}
		var nrm, uv [][]float64
		if ni, ok := prim.Attributes["NORMAL"]; ok {
			if nrm, err = g.accessor(ni); err != nil {
				return err
			}
		}
		if ti, ok := prim.Attributes["TEXCOORD_0"]; ok {
			if uv, err = g.accessor(ti); err != nil {
				return err
			}
		}
		var idx []int
		if prim.Indices != nil {
			raw, err := g.accessor(*prim.Indices)
			if err != nil {
				return err
			}
			for _, v := range raw {
				idx = append(idx, int(v[0]))
			}
		} else {
			for i := range pos {
				idx = append(idx, i)
			}
		}
		factor := [4]float64{1, 1, 1, 1}
		var tex image.Image
		if prim.Material != nil && *prim.Material < len(g.doc.Materials) {
			mat := g.doc.Materials[*prim.Material]
			copy(factor[:], mat.PBR.BaseColorFactor)
			if mat.PBR.BaseColorTexture != nil {
				tex = g.texture(mat.PBR.BaseColorTexture.Index)
			}
		}
		for t := 0; t+2 < len(idx); t += 3 {
			var tp [3]Vec3
			for k := 0; k < 3; k++ {
				vi := idx[t+k]
				if vi < 0 || vi >= len(pos) {
					return errors.New("index out of range")
				}
				p := xf.point(Vec3{pos[vi][0], pos[vi][1], pos[vi][2]})
				tp[k] = p
				m.Pos = append(m.Pos, p)
				m.Min = Vec3{math.Min(m.Min.X, p.X), math.Min(m.Min.Y, p.Y), math.Min(m.Min.Z, p.Z)}
				m.Max = Vec3{math.Max(m.Max.X, p.X), math.Max(m.Max.Y, p.Y), math.Max(m.Max.Z, p.Z)}
				c := [4]float64{factor[0], factor[1], factor[2], factor[3]}
				if tex != nil && vi < len(uv) {
					tc := samplePalette(tex, uv[vi][0], uv[vi][1])
					c[0] *= tc[0]
					c[1] *= tc[1]
					c[2] *= tc[2]
				}
				m.Col = append(m.Col, color.RGBA{clampByte(c[0] * 255), clampByte(c[1] * 255), clampByte(c[2] * 255), 255})
				if vi < len(nrm) {
					m.Norm = append(m.Norm, xf.dir(Vec3{nrm[vi][0], nrm[vi][1], nrm[vi][2]}).Norm())
				} else {
					m.Norm = append(m.Norm, Vec3{})
				}
			}
			if len(nrm) == 0 {
				fn := tp[1].Sub(tp[0]).Cross(tp[2].Sub(tp[0])).Norm()
				for k := 1; k <= 3; k++ {
					m.Norm[len(m.Norm)-k] = fn
				}
			}
		}
	}
	return nil
}

// accessor decodes an accessor into count rows of float components.
func (g *glb) accessor(ai int) ([][]float64, error) {
	if ai < 0 || ai >= len(g.doc.Accessors) {
		return nil, errors.New("bad accessor index")
	}
	a := g.doc.Accessors[ai]
	if a.BufferView == nil || *a.BufferView >= len(g.doc.BufferViews) {
		return nil, errors.New("accessor without buffer view")
	}
	bv := g.doc.BufferViews[*a.BufferView]
	ncomp := map[string]int{"SCALAR": 1, "VEC2": 2, "VEC3": 3, "VEC4": 4}[a.Type]
	csize := map[int]int{5120: 1, 5121: 1, 5122: 2, 5123: 2, 5125: 4, 5126: 4}[a.ComponentType]
	if ncomp == 0 || csize == 0 {
		return nil, fmt.Errorf("unsupported accessor %s/%d", a.Type, a.ComponentType)
	}
	stride := bv.ByteStride
	if stride == 0 {
		stride = ncomp * csize
	}
	base := bv.ByteOffset + a.ByteOffset
	if a.Count > 0 && base+(a.Count-1)*stride+ncomp*csize > len(g.bin) {
		return nil, errors.New("accessor out of buffer range")
	}
	out := make([][]float64, a.Count)
	for i := 0; i < a.Count; i++ {
		row := make([]float64, ncomp)
		for k := 0; k < ncomp; k++ {
			o := base + i*stride + k*csize
			var v float64
			switch a.ComponentType {
			case 5126:
				v = float64(math.Float32frombits(binary.LittleEndian.Uint32(g.bin[o:])))
			case 5125:
				v = float64(binary.LittleEndian.Uint32(g.bin[o:]))
			case 5123:
				v = float64(binary.LittleEndian.Uint16(g.bin[o:]))
				if a.Normalized {
					v /= 65535
				}
			case 5122:
				v = float64(int16(binary.LittleEndian.Uint16(g.bin[o:])))
				if a.Normalized {
					v = math.Max(v/32767, -1)
				}
			case 5121:
				v = float64(g.bin[o])
				if a.Normalized {
					v /= 255
				}
			case 5120:
				v = float64(int8(g.bin[o]))
				if a.Normalized {
					v = math.Max(v/127, -1)
				}
			}
			row[k] = v
		}
		out[i] = row
	}
	return out, nil
}

// texture loads the image behind a glTF texture index. Kenney GLBs reference
// "Textures/colormap.png"; we keep that palette next to the model instead.
func (g *glb) texture(ti int) image.Image {
	if ti < 0 || ti >= len(g.doc.Textures) {
		return nil
	}
	src := g.doc.Textures[ti].Source
	if img, ok := g.imgs[src]; ok {
		return img
	}
	if src < 0 || src >= len(g.doc.Images) {
		return nil
	}
	data, err := assetFS.ReadFile(path.Join(g.dir, path.Base(g.doc.Images[src].URI)))
	if err != nil {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	g.imgs[src] = img
	return img
}

// samplePalette reads the texel under (u,v), wrapping like glTF REPEAT.
func samplePalette(img image.Image, u, v float64) [3]float64 {
	b := img.Bounds()
	u -= math.Floor(u)
	v -= math.Floor(v)
	x := b.Min.X + int(u*float64(b.Dx()))
	y := b.Min.Y + int(v*float64(b.Dy()))
	if x >= b.Max.X {
		x = b.Max.X - 1
	}
	if y >= b.Max.Y {
		y = b.Max.Y - 1
	}
	r, g, bb, _ := img.At(x, y).RGBA()
	return [3]float64{float64(r) / 65535, float64(g) / 65535, float64(bb) / 65535}
}

// ---- 4x4 column-major matrices (glTF convention) ----

type mat4 [16]float64

func mat4Identity() mat4 { return mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1} }

func (a mat4) mul(b mat4) mat4 {
	var r mat4
	for c := 0; c < 4; c++ {
		for rr := 0; rr < 4; rr++ {
			var s float64
			for k := 0; k < 4; k++ {
				s += a[k*4+rr] * b[c*4+k]
			}
			r[c*4+rr] = s
		}
	}
	return r
}

func mat4TRS(t [3]float64, q [4]float64, s [3]float64) mat4 {
	x, y, z, w := q[0], q[1], q[2], q[3]
	r := mat4{
		1 - 2*(y*y+z*z), 2 * (x*y + w*z), 2 * (x*z - w*y), 0,
		2 * (x*y - w*z), 1 - 2*(x*x+z*z), 2 * (y*z + w*x), 0,
		2 * (x*z + w*y), 2 * (y*z - w*x), 1 - 2*(x*x+y*y), 0,
		t[0], t[1], t[2], 1,
	}
	for c := 0; c < 3; c++ {
		for rr := 0; rr < 3; rr++ {
			r[c*4+rr] *= s[c]
		}
	}
	return r
}

func (a mat4) point(p Vec3) Vec3 {
	return Vec3{
		a[0]*p.X + a[4]*p.Y + a[8]*p.Z + a[12],
		a[1]*p.X + a[5]*p.Y + a[9]*p.Z + a[13],
		a[2]*p.X + a[6]*p.Y + a[10]*p.Z + a[14],
	}
}

func (a mat4) dir(p Vec3) Vec3 {
	return Vec3{
		a[0]*p.X + a[4]*p.Y + a[8]*p.Z,
		a[1]*p.X + a[5]*p.Y + a[9]*p.Z,
		a[2]*p.X + a[6]*p.Y + a[10]*p.Z,
	}
}

// addMesh places a model at pos with the given yaw and uniform scale. The
// model's footprint is centred on pos and its base sits on pos.Y.
func (s *Scene) addMesh(m *Mesh, pos Vec3, yaw, scale float64) {
	cx := (m.Min.X + m.Max.X) / 2
	cz := (m.Min.Z + m.Max.Z) / 2
	for t := 0; t+2 < len(m.Pos); t += 3 {
		var p, n [3]Vec3
		var c [3]color.RGBA
		for k := 0; k < 3; k++ {
			lp := m.Pos[t+k]
			lp = Vec3{(lp.X - cx) * scale, (lp.Y - m.Min.Y) * scale, (lp.Z - cz) * scale}
			p[k] = pos.Add(rotY(lp, yaw))
			n[k] = rotY(m.Norm[t+k], yaw)
			c[k] = m.Col[t+k]
		}
		s.addTriV(p, c, n, true)
	}
}
