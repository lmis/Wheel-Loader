package main

// The GPU renderer. The mesh builders fill a Scene with world-space
// triangles each frame exactly as before; instead of projecting and
// painter-sorting them on the CPU, they are packed into vertex buffers and
// drawn by OpenGL (raylib-go's rlgl) with a depth buffer, per-pixel lighting
// (the same sky/ground ambient and sun as lit()), a soft sun shadow map,
// distance fog and 2x2 supersampling.

import (
	"cmp"
	"image"
	"image/color"
	"log"
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const gpuVS = `#version 330
layout(location=0) in vec3 aPos;
layout(location=1) in vec3 aN;
layout(location=2) in vec4 aCol;
layout(location=3) in vec2 aExtra; // x: sort bias (world units), y: unlit
layout(location=4) in vec2 aUV;    // atlas texels
uniform mat4 uVP;
uniform vec3 uEye;
out vec3 vPos; out vec3 vN; out vec4 vCol; out float vUnlit; out vec2 vUV;
void main() {
	vPos = aPos; vN = aN; vCol = aCol; vUnlit = aExtra.y; vUV = aUV;
	// detail() parts used to sort 'bias' units nearer the camera; sliding the
	// vertex toward the eye along its view ray does the same in depth
	// without moving it on screen
	vec3 d = aPos - uEye;
	float L = length(d);
	vec3 p = uEye + d * max(0.05, L - aExtra.x) / L;
	gl_Position = uVP * vec4(p, 1.0);
}`

const gpuFS = `#version 330
in vec3 vPos; in vec3 vN; in vec4 vCol; in float vUnlit; in vec2 vUV;
uniform vec3 uLightDir, uSky, uGnd, uSun, uEye, uFwd, uFogCol;
uniform mat4 uLightVP;
uniform sampler2D uShadow;
uniform sampler2D uTex;
uniform vec2 uTexSize;
uniform int uMode; // 0 objects (atlas), 1 site ground, 2 ground skirt, 3 flat unlit, 4 soft round billboards
uniform float uShadowOn, uRim, uTexel, uFill, uDepthBias;
out vec4 frag;

float shadowAt(vec3 n) {
	vec4 lp = uLightVP * vec4(vPos + n * 0.35, 1.0); // normal offset
	vec3 p = lp.xyz / lp.w * 0.5 + 0.5;
	if (p.x < 0.0 || p.x > 1.0 || p.y < 0.0 || p.y > 1.0 || p.z > 1.0) return 1.0;
	float s = 0.0;
	for (int x = -1; x <= 1; x++)
		for (int y = -1; y <= 1; y++) {
			float d = texture(uShadow, p.xy + vec2(x, y) * 1.25 * uTexel).r;
			s += (p.z - uDepthBias > d) ? 0.0 : 1.0;
		}
	return s / 9.0;
}

void main() {
	float z = dot(vPos - uEye, uFwd);
	float t = clamp((z - 1000.0) / 2600.0, 0.0, 1.0);
	float fog = t * t * (3.0 - 2.0 * t) * 0.9;
	if (uMode == 1 || uMode == 2) {
		// the ground is unlit texture, only darkened in shadow
		vec2 uv = uMode == 1 ? clamp(vPos.xz, vec2(0.5), uTexSize - 0.5) / uTexSize : vPos.xz / uTexSize;
		vec3 g = texture(uTex, uv).rgb;
		float sh = uShadowOn > 0.5 ? shadowAt(vec3(0.0, 1.0, 0.0)) : 1.0;
		g *= mix(0.6, 1.0, sh);
		frag = vec4(mix(g, uFogCol, fog), 1.0);
		return;
	}
	vec4 base = vCol;
	if (uMode == 0) base *= texture(uTex, vUV / uTexSize);
	if (uMode == 4) { // particles: a soft round puff (vUV runs -1..1 across)
		float r = length(vUV);
		if (r > 1.0) discard;
		frag = vec4(mix(base.rgb, uFogCol, fog), base.a * (1.0 - fog) * smoothstep(1.0, 0.15, r));
		return;
	}
	if (vUnlit > 0.5 || uMode == 3) { // decals and glowing parts
		frag = vec4(mix(base.rgb, uFogCol, fog), base.a * (uMode == 3 ? 1.0 - fog : 1.0));
		return;
	}
	vec3 n = normalize(vN);
	if (!gl_FrontFacing) n = -n;
	float d = max(dot(n, uLightDir), 0.0);
	float sh = (uShadowOn > 0.5 && d > 0.0) ? shadowAt(n) : 1.0;
	float h = n.y * 0.5 + 0.5;
	vec3 k = mix(uGnd, uSky, h) + uSun * d * mix(uFill, 1.0, sh);
	vec3 c = min(base.rgb * k, vec3(1.0));
	vec3 V = normalize(uEye - vPos);
	float rim = pow(1.0 - max(dot(n, V), 0.0), 3.0) * uRim * (0.4 + 0.6 * h);
	c += rim * vec3(0.85, 0.9, 1.0);
	frag = vec4(mix(c, uFogCol, fog), base.a);
}`

const gpuDepthVS = `#version 330
layout(location=0) in vec3 aPos;
uniform mat4 uLightVP;
void main() { gl_Position = uLightVP * vec4(aPos, 1.0); }`

const gpuDepthFS = `#version 330
void main() {}`

const (
	vstride   = 14 // floats per vertex: pos3 normal3 rgba4 extra2 uv2
	shadowRes = 2048
	superS    = 2 // supersampling factor per axis
)

// vbuf is one dynamic vertex buffer drawn as a triangle list.
type vbuf struct {
	vao, vbo uint32
	cap      int // floats
	n        int // vertices
}

func (b *vbuf) upload(data []float32) {
	if len(data) == 0 {
		b.n = 0
		return
	}
	if b.vao == 0 || len(data) > b.cap {
		if b.vao != 0 {
			rl.UnloadVertexBuffer(b.vbo)
			rl.UnloadVertexArray(b.vao)
		}
		b.cap = len(data) * 3 / 2
		b.vao = rl.LoadVertexArray()
		rl.EnableVertexArray(b.vao)
		grown := make([]float32, b.cap)
		copy(grown, data)
		b.vbo = rl.LoadVertexBuffer(grown, true)
		for i, a := range [][2]int32{{3, 0}, {3, 3}, {4, 6}, {2, 10}, {2, 12}} {
			rl.SetVertexAttribute(uint32(i), a[0], rl.Float, false, vstride*4, a[1]*4)
			rl.EnableVertexAttribute(uint32(i))
		}
		rl.DisableVertexArray()
	} else {
		rl.UpdateVertexBuffer(b.vbo, data, 0)
	}
	b.n = len(data) / vstride
}

func (b *vbuf) draw() {
	if b.n == 0 {
		return
	}
	rl.EnableVertexArray(b.vao)
	rl.DrawVertexArray(0, int32(b.n))
	rl.DisableVertexArray()
}

type prog struct {
	id  uint32
	loc map[string]int32
}

func loadProg(vs, fs string) prog {
	id := rl.LoadShaderProgram(vs, fs)
	if id == 0 {
		log.Fatal("gpu: shader failed to compile")
	}
	return prog{id: id, loc: map[string]int32{}}
}

func (p prog) l(name string) int32 {
	if v, ok := p.loc[name]; ok {
		return v
	}
	v := rl.GetLocationUniform(p.id, name)
	p.loc[name] = v
	return v
}
func (p prog) vec3(name string, v Vec3) {
	rl.SetUniform(p.l(name), []float32{float32(v.X), float32(v.Y), float32(v.Z)}, int32(rl.ShaderUniformVec3), 1)
}
func (p prog) col(name string, c [3]float64) {
	rl.SetUniform(p.l(name), []float32{float32(c[0]), float32(c[1]), float32(c[2])}, int32(rl.ShaderUniformVec3), 1)
}
func (p prog) vec2(name string, x, y float64) {
	rl.SetUniform(p.l(name), []float32{float32(x), float32(y)}, int32(rl.ShaderUniformVec2), 1)
}
func (p prog) f(name string, v float64) {
	rl.SetUniform(p.l(name), []float32{float32(v)}, int32(rl.ShaderUniformFloat), 1)
}
func (p prog) i(name string, v int32) {
	rl.SetUniform(p.l(name), []int32{v}, int32(rl.ShaderUniformInt), 1)
}
func (p prog) mat(name string, m rl.Matrix) { rl.SetUniformMatrix(p.l(name), m) }

// GPU holds the renderer's GL objects. It exists only once a window is
// open; the simulation and its tests never touch it.
type GPU struct {
	main, depth          prog
	shadowTex, shadowFB  uint32
	ground, skirt, atlas rl.Texture2D

	objs, decals, ground3, skirt3, bills vbuf
	buf                                  []float32
	targets                              map[[3]int]rl.RenderTexture2D // supersampled view targets by size and slot
}

func rlTexture(img *image.RGBA, repeat bool) rl.Texture2D {
	t := rl.LoadTextureFromImage(rl.NewImageFromImage(img))
	rl.GenTextureMipmaps(&t)
	rl.SetTextureFilter(t, rl.FilterTrilinear)
	if repeat {
		rl.SetTextureWrap(t, rl.WrapRepeat)
	} else {
		rl.SetTextureWrap(t, rl.WrapClamp)
	}
	return t
}

// initGPU compiles the shaders and uploads the textures. A window must be
// open.
func (g *Game) initGPU() {
	gp := &GPU{targets: map[[3]int]rl.RenderTexture2D{}}
	gp.main = loadProg(gpuVS, gpuFS)
	gp.depth = loadProg(gpuDepthVS, gpuDepthFS)
	gp.shadowTex = rl.LoadTextureDepth(shadowRes, shadowRes, false)
	gp.shadowFB = rl.LoadFramebuffer()
	rl.FramebufferAttach(gp.shadowFB, gp.shadowTex, rl.AttachmentDepth, rl.AttachmentTexture2d, 0)
	if !rl.FramebufferComplete(gp.shadowFB) {
		log.Fatal("gpu: shadow framebuffer incomplete")
	}
	gp.ground = rlTexture(g.ground, false)
	gp.skirt = rlTexture(g.skirt, true)
	gp.atlas = rlTexture(g.atlas, false)
	rl.SetTextureFilter(gp.atlas, rl.FilterBilinear)
	g.gpu = gp
	g.art.load()
}

// target returns a supersampled render target for a w×h view; each slot
// (player view) has its own.
func (gp *GPU) target(w, h, slot int) rl.RenderTexture2D {
	k := [3]int{w, h, slot}
	if rt, ok := gp.targets[k]; ok {
		return rt
	}
	rt := rl.LoadRenderTexture(int32(w*superS), int32(h*superS))
	rl.SetTextureFilter(rt.Texture, rl.FilterBilinear)
	gp.targets[k] = rt
	return rt
}

func v3(v Vec3) rl.Vector3 { return rl.NewVector3(float32(v.X), float32(v.Y), float32(v.Z)) }

func col4(c color.RGBA) (float32, float32, float32, float32) {
	return float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A) / 255
}

// packTris converts the Scene's triangles into vertices. Triangles the old
// renderer did not cull are emitted twice, once per winding.
func packTris(tris []tri, out []float32) []float32 {
	out = out[:0]
	for i := range tris {
		t := &tris[i]
		fn := t.p[1].Sub(t.p[0]).Cross(t.p[2].Sub(t.p[0])).Norm()
		unlit := float32(0)
		if !t.shade {
			unlit = 1
		}
		bias := float32(t.bias)
		emit := func(k int, flip bool) {
			n := t.n[k]
			if n == (Vec3{}) {
				n = fn
			}
			if flip {
				n = n.Scale(-1)
			}
			r, gg, b, a := col4(t.c[k])
			u, v := float32(atlasWhiteUV), float32(atlasWhiteUV)
			if t.tex {
				u, v = t.uv[k][0], t.uv[k][1]
			}
			p := t.p[k]
			out = append(out, float32(p.X), float32(p.Y), float32(p.Z), float32(n.X), float32(n.Y), float32(n.Z),
				r, gg, b, a, bias, unlit, u, v)
		}
		emit(0, false)
		emit(1, false)
		emit(2, false)
		if !t.cull {
			emit(0, true)
			emit(2, true)
			emit(1, true)
		}
	}
	return out
}

// packFlat packs unlit, two-sided triangles (decals and billboards).
func packFlat(tris []tri, out []float32) []float32 {
	out = out[:0]
	for i := range tris {
		t := &tris[i]
		for _, ks := range [2][3]int{{0, 1, 2}, {0, 2, 1}} {
			for _, k := range ks {
				r, gg, b, a := col4(t.c[k])
				p := t.p[k]
				out = append(out, float32(p.X), float32(p.Y), float32(p.Z), 0, 1, 0, r, gg, b, a, 0, 1, 0, 0)
			}
		}
	}
	return out
}

func packGround(quads []gquad, out []float32) []float32 {
	out = out[:0]
	for i := range quads {
		q := &quads[i]
		for _, k := range []int{0, 2, 1, 0, 3, 2, 0, 1, 2, 0, 2, 3} { // both windings
			p := q.p[k]
			out = append(out, float32(p.X), float32(p.Y), float32(p.Z), 0, 1, 0, 1, 1, 1, 1, 0, 0, 0, 0)
		}
	}
	return out
}

// packBillboards turns the scene's billboards into camera-facing quads,
// sorted far to near for blending, with corner coordinates for a round puff.
func packBillboards(s *Scene, cam *Camera, out []float32) []float32 {
	out = out[:0]
	bills := s.bills
	slices.SortFunc(bills, func(a, b billboard) int {
		return cmp.Compare(b.center.Sub(cam.eye).Len(), a.center.Sub(cam.eye).Len())
	})
	corner := [4][2]float32{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	for _, b := range bills {
		r := cam.right.Scale(b.size * 1.25) // the puff fades out toward the corners
		u := cam.up.Scale(b.size * 1.25)
		p := [4]Vec3{b.center.Sub(r).Sub(u), b.center.Add(r).Sub(u), b.center.Add(r).Add(u), b.center.Sub(r).Add(u)}
		cr, cg, cb, ca := col4(b.col)
		for _, k := range []int{0, 2, 1, 0, 3, 2, 0, 1, 2, 0, 2, 3} {
			out = append(out, float32(p[k].X), float32(p[k].Y), float32(p[k].Z), 0, 1, 0, cr, cg, cb, ca, 0, 1, corner[k][0], corner[k][1])
		}
	}
	return out
}

// drawScene renders a built scene from cam into a w×h supersampled target
// and returns it; the caller draws it scaled down where it belongs.
func (g *Game) drawScene(s *Scene, cam *Camera, w, h, slot int) rl.RenderTexture2D {
	gp := g.gpu
	rt := gp.target(w, h, slot)
	cam.ScreenW, cam.ScreenH = float64(w), float64(h)
	cam.setup()

	gp.buf = packTris(s.tris, gp.buf)
	gp.objs.upload(gp.buf)
	gp.buf = packFlat(s.decals, gp.buf)
	gp.decals.upload(gp.buf)
	gp.buf = packBillboards(s, cam, gp.buf)
	gp.bills.upload(gp.buf)
	gp.buf = packGround(s.ground, gp.buf)
	gp.ground3.upload(gp.buf)
	gp.buf = packGround(s.skirt, gp.buf)
	gp.skirt3.upload(gp.buf)

	// camera matrices from camera.go: vertical fov, eye and forward
	fov := 2 * math.Atan(1/cam.f)
	view := rl.MatrixLookAt(v3(cam.eye), v3(cam.eye.Add(cam.fwd)), rl.NewVector3(0, 1, 0))
	proj := rl.MatrixPerspective(float32(fov), float32(cam.aspect), 1, 8000)
	vp := rl.MatrixMultiply(view, proj)

	// the sun's shadow box: around what the camera looks at
	ctr := Vec3{cam.Target.X, 0, cam.Target.Y}
	ext := clamp(cam.Dist*0.95, 260, 900)
	if cam.fp {
		ahead := Vec3{cam.fwd.X, 0, cam.fwd.Z}.Norm()
		ctr, ext = Vec3{cam.eye.X, 0, cam.eye.Z}.Add(ahead.Scale(260)), 420
	}
	const lightDist, lightFar = 1400.0, 2800.0
	lview := rl.MatrixLookAt(v3(ctr.Add(lightDir.Scale(lightDist))), v3(ctr), rl.NewVector3(0, 0, 1))
	lproj := rl.MatrixOrtho(-float32(ext), float32(ext), -float32(ext), float32(ext), 1, lightFar)
	lvp := rl.MatrixMultiply(lview, lproj)

	// 1. shadow map
	rl.DrawRenderBatchActive()
	rl.EnableFramebuffer(gp.shadowFB)
	rl.Viewport(0, 0, shadowRes, shadowRes)
	rl.ClearScreenBuffers()
	rl.EnableDepthTest()
	rl.DisableBackfaceCulling()
	rl.EnableShader(gp.depth.id)
	gp.depth.mat("uLightVP", lvp)
	gp.objs.draw()
	rl.DisableShader()
	rl.EnableBackfaceCulling()
	rl.DisableFramebuffer()

	// 2. the view
	rl.BeginTextureMode(rt)
	W := int32(w * superS)
	fogC := rl.NewColor(uint8(fogCol[0]), uint8(fogCol[1]), uint8(fogCol[2]), 255)
	rl.ClearBackground(fogC)
	horizon := 0.0 // sky gradient down to the horizon, fog below
	hf := Vec3{cam.fwd.X, 0, cam.fwd.Z}.Norm()
	if v := cam.view(cam.eye.Add(hf.Scale(1e6))); v.Z > 0 {
		_, sy := cam.project(v)
		horizon = clamp(sy, 0, cam.ScreenH)
	}
	rl.DrawRectangleGradientV(0, 0, W, int32(horizon*superS),
		rl.NewColor(uint8(skyTop[0]), uint8(skyTop[1]), uint8(skyTop[2]), 255), fogC)
	rl.DrawRenderBatchActive()

	p := gp.main
	rl.EnableDepthTest()
	rl.EnableShader(p.id)
	p.mat("uVP", vp)
	p.mat("uLightVP", lvp)
	p.vec3("uEye", cam.eye)
	p.vec3("uFwd", cam.fwd)
	p.vec3("uLightDir", lightDir)
	p.col("uSky", skyAmb)
	p.col("uGnd", gndAmb)
	p.col("uSun", sunCol)
	p.col("uFogCol", [3]float64{fogCol[0] / 255, fogCol[1] / 255, fogCol[2] / 255})
	p.f("uShadowOn", 1)
	p.f("uRim", 0.18)
	p.f("uTexel", 1.0/shadowRes)
	p.f("uFill", 0.4)
	p.f("uDepthBias", 2.0/lightFar)
	rl.ActiveTextureSlot(1)
	rl.EnableTexture(gp.shadowTex)
	p.i("uShadow", 1)
	p.i("uTex", 2)
	bind := func(t rl.Texture2D, size Vec) {
		rl.ActiveTextureSlot(2)
		rl.EnableTexture(t.ID)
		p.vec2("uTexSize", size.X, size.Y)
	}
	rl.SetBlendMode(rl.BlendAlpha)

	// ground: the baked site and the repeating dirt beyond it
	p.i("uMode", 1)
	bind(gp.ground, V(float64(gp.ground.Width), float64(gp.ground.Height)))
	gp.ground3.draw()
	p.i("uMode", 2)
	bind(gp.skirt, V(float64(gp.skirt.Width), float64(gp.skirt.Height)))
	gp.skirt3.draw()
	// decals (contact shadows, glows): blended, no depth writes
	p.i("uMode", 3)
	rl.DisableDepthMask()
	gp.decals.draw()
	rl.EnableDepthMask()
	// objects
	p.i("uMode", 0)
	bind(gp.atlas, V(atlasW, atlasH))
	gp.objs.draw()
	// billboards (particles, flags): blended over everything
	p.i("uMode", 4)
	rl.DisableDepthMask()
	gp.bills.draw()
	rl.EnableDepthMask()

	rl.DisableShader()
	rl.ActiveTextureSlot(0)
	rl.DisableDepthTest()
	rl.EndTextureMode()
	return rt
}

// blit draws a supersampled view target scaled down into x, y, w, h of the
// current target.
func blit(rt rl.RenderTexture2D, x, y, w, h float64) {
	src := rl.NewRectangle(0, 0, float32(rt.Texture.Width), -float32(rt.Texture.Height))
	rl.DrawTexturePro(rt.Texture, src, rl.NewRectangle(float32(x), float32(y), float32(w), float32(h)), rl.Vector2{}, 0, rl.White)
}

// readTarget copies a render target back into a Go image, top row first.
func readTarget(rt rl.RenderTexture2D) *image.RGBA {
	img := rl.LoadImageFromTexture(rt.Texture)
	defer rl.UnloadImage(img)
	rl.ImageFlipVertical(img)
	src := img.ToImage()
	out := image.NewRGBA(src.Bounds())
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			r, gg, b, _ := src.At(x, y).RGBA()
			out.Set(x, y, color.RGBA{uint8(r >> 8), uint8(gg >> 8), uint8(b >> 8), 255})
		}
	}
	return out
}
