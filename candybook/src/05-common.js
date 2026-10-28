// ---------- bits every page uses ----------
// Acts keeps the clocks of the little shows running on a page.
class Acts {
  constructor(durs) { this.a = {}; this.durs = durs; }
  run(n) { this.a[n] = 0; }
  t(n) { return this.a[n]; }
  on(n) { return n in this.a; }
  at(n, k, dt) { const t = this.a[n]; return t !== undefined && t >= k && t - dt < k; }
  update(dt) { for (const k in this.a) { this.a[k] += dt; if (this.a[k] > (this.durs[k] ?? 3)) delete this.a[k]; } }
}
const cycle = (st, k, list) => list[st.next[k]++ % list.length];

function makeStars(n, x0, x1, y0, y1) {
  return Array.from({ length: n }, () => ({ x: rnd(x0, x1), y: rnd(y0, y1), r: rnd(1, 2.6), ph: rnd(0, TAU), sp: rnd(1, 3) }));
}
function drawStars(stars, t, a = 1) {
  for (const s of stars) {
    const k = 0.55 + 0.45 * Math.sin(t * s.sp + s.ph);
    ctx.globalAlpha = a * k;
    if (s.r > 2.2) star(s.x, s.y, s.r * 2.4, t * 0.2 + s.ph, 4, '#fff6c8', 0);
    else circ(s.x, s.y, s.r, '#fff6dc', 0);
  }
  ctx.globalAlpha = 1;
}
// Fireflies drift and blink over an area.
function makeFlies(n, x0, x1, y0, y1) {
  return Array.from({ length: n }, () => ({ x: rnd(x0, x1), y: rnd(y0, y1), hx: 0, hy: 0, ph: rnd(0, TAU), x0, x1, y0, y1, sp: rnd(0.3, 0.8) }));
}
function updateFlies(flies, t, dt, target) {
  for (const f of flies) {
    if (target) {
      const [tx, ty] = target(f);
      f.x = lerp(f.x, tx, Math.min(1, dt * 2.2)); f.y = lerp(f.y, ty, Math.min(1, dt * 2.2));
    } else {
      f.x += Math.sin(t * f.sp + f.ph) * 22 * dt + Math.cos(t * 0.37 + f.ph * 2) * 10 * dt;
      f.y += Math.cos(t * f.sp * 1.3 + f.ph) * 18 * dt;
      f.x = clamp(f.x, f.x0, f.x1); f.y = clamp(f.y, f.y0, f.y1);
    }
  }
}
function drawFlies(flies, t, boost = 0) {
  for (const f of flies) {
    const k = clamp(0.4 + 0.6 * Math.sin(t * 2.2 + f.ph) + boost, 0, 1);
    glow(f.x, f.y, 16, '#f5ff8c', 0.55 * k);
    circ(f.x, f.y, 2.6, alpha('#fbffc8', 0.4 + 0.6 * k), 0);
  }
}
function drawMoon(x, y, r, t, awake = 0) {
  glow(x, y, r * 2.6, '#fff4c8', 0.28 + awake * 0.2);
  circ(x, y, r, '#fff4cc');
  circ(x - r * 0.35, y + r * 0.3, r * 0.14, 'rgba(230,214,160,0.7)', 0);
  circ(x + r * 0.42, y - r * 0.35, r * 0.1, 'rgba(230,214,160,0.7)', 0);
  circ(x + r * 0.2, y + r * 0.5, r * 0.08, 'rgba(230,214,160,0.7)', 0);
  // a sleepy face that wakes up
  if (awake > 0.3) {
    circ(x - r * 0.3, y - r * 0.12, r * 0.1, INK, 0); circ(x + r * 0.25, y - r * 0.12, r * 0.1, INK, 0);
    circ(x - r * 0.33, y - r * 0.16, r * 0.03, WHITE, 0); circ(x + r * 0.22, y - r * 0.16, r * 0.03, WHITE, 0);
  } else {
    ctx.beginPath(); ctx.arc(x - r * 0.3, y - r * 0.15, r * 0.12, 0.2, Math.PI - 0.2); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.stroke();
    ctx.beginPath(); ctx.arc(x + r * 0.25, y - r * 0.15, r * 0.12, 0.2, Math.PI - 0.2); ctx.stroke();
  }
  ctx.beginPath(); ctx.arc(x - r * 0.02, y + r * 0.15, r * (0.18 + awake * 0.08), 0.25, Math.PI - 0.25); ctx.lineWidth = 3; ctx.stroke();
  circ(x - r * 0.5, y + r * 0.1, r * 0.1, 'rgba(255,150,150,0.45)', 0); circ(x + r * 0.45, y + r * 0.1, r * 0.1, 'rgba(255,150,150,0.45)', 0);
}
function grassTufts(xs, groundY, t, col) {
  for (const x of xs) {
    const g = typeof groundY === 'function' ? groundY(x) : groundY;
    for (let i = -1; i <= 1; i++) {
      const sw = Math.sin(t * 1.6 + x * 0.05 + i) * 5;
      curve(x + i * 6, g + 2, x + i * 8 + sw * 0.5, g - 12, x + i * 10 + sw, g - 22 - Math.abs(i) * -4, 3, col, 0);
    }
  }
}
function drawBird(x, y, s, flap, col = '#6482d2', face = 1) {
  if (col !== '#ffd25a' && paintedBird(x, y, s, flap, col, face)) return;
  push(x, y, 0, s * face, s);
  if (flap !== null) {
    const f = Math.sin(flap) * 14;
    poly([-2, -4, -16, -14 - f, -22, -2 - f * 0.5], shade(col, -0.15), 2.5);
    poly([2, -4, 12, -18 - f, 18, -4 - f * 0.5], shade(col, -0.15), 2.5);
  }
  ell(0, 0, 13, 10, col, 2.5);
  circ(9, -8, 8, col, 2.5);
  poly([16, -9, 23, -7, 16, -5], '#ffb43c', 2);
  circ(11, -10, 2, INK, 0);
  ell(-2, 3, 7, 5, shade(col, 0.4), 0);
  pop();
}

// ---- painted pictures ----
// art(name) gives the painted picture from art/<name>.jpg once it has
// loaded, or null; pages fall back to drawing the scene themselves.
const ART_IMG = {};
function art(name) {
  if (typeof ART === 'undefined' || !ART[name]) return null;
  let im = ART_IMG[name];
  if (!im) { im = new Image(); im.src = ART[name]; ART_IMG[name] = im; }
  return im.complete && im.naturalWidth ? im : null;
}
// drawArt fills the page with a picture, cropping the edges if its shape differs
function drawArt(im) {
  const s = Math.max(W / im.naturalWidth, H / im.naturalHeight), w = im.naturalWidth * s, h = im.naturalHeight * s;
  ctx.drawImage(im, (W - w) / 2, (H - h) / 2, w, h);
}
// drawArtScroll fills the page with a picture sliding left by off pixels,
// every other copy mirrored so the edges always meet
function drawArtScroll(im, off) {
  const w = im.naturalWidth * H / im.naturalHeight; off %= 2 * w;
  for (let i = -1; i < 3; i++) {
    const x = i * w - off;
    if (x > W || x + w < 0) continue;
    ctx.save(); ctx.translate(x + (i & 1 ? w : 0), 0); ctx.scale(i & 1 ? -1 : 1, 1);
    ctx.drawImage(im, 0, 0, w, H); ctx.restore();
  }
}

// ---- painted animals and props (fall back to the drawn ones until loaded) ----
// pic draws a picture with its bottom centre at x,y, w pixels wide, mirrored if flip
function pic(im, x, y, w, flip = false, rot = 0) {
  const h = w * im.naturalHeight / im.naturalWidth;
  ctx.save(); ctx.translate(x, y); if (rot) ctx.rotate(rot); if (flip) ctx.scale(-1, 1);
  ctx.drawImage(im, -w / 2, -h, w, h); ctx.restore();
}
const allArt = (...n) => n.every(x => art(x));
// a pigeon (grey-ish colours) or a songbird, facing right unless face < 0;
// flap is null when standing, otherwise the wings beat with it
function paintedBird(x, y, s, flap, col, face = 1) {
  const kind = col === '#a0a0c8' ? 'pigeon' : 'bird';
  if (!allArt(`${kind}-body`, `${kind}-head`, `${kind}-wing`, `${kind}-wing-open`)) return false;
  const k = s * (kind === 'pigeon' ? 0.7 : 0.6);
  ctx.save(); ctx.translate(x, y + 8); ctx.scale(face * k, k);
  const body = art(`${kind}-body`), head = art(`${kind}-head`);
  ctx.drawImage(body, -body.naturalWidth / 2, -body.naturalHeight, body.naturalWidth, body.naturalHeight);
  ctx.drawImage(head, body.naturalWidth * 0.08, -body.naturalHeight * 1.32, head.naturalWidth, head.naturalHeight);
  const w = art(flap === null ? `${kind}-wing` : `${kind}-wing-open`);
  ctx.save(); ctx.translate(-body.naturalWidth * 0.05, -body.naturalHeight * 0.6);
  if (flap !== null) ctx.rotate(-0.4 + Math.sin(flap) * 0.6);
  ctx.drawImage(w, -w.naturalWidth * 0.7, -w.naturalHeight * (flap === null ? 0.4 : 0.9), w.naturalWidth, w.naturalHeight);
  ctx.restore(); ctx.restore();
  return true;
}
function paintedCat(x, y, state) { // state: sleepy, awake, meow
  if (!allArt('cat-body', 'cat-tail', 'cat-head-sleepy', 'cat-head-awake', 'cat-head-meow')) return false;
  const t = performance.now() / 1000;
  pic(art('cat-tail'), x - 46, y - 4, 50, false, Math.sin(t * 2) * 0.15);
  pic(art('cat-body'), x, y, 92);
  pic(art('cat-head-' + state), x + 38, y - 22 - (state === 'sleepy' ? 0 : 6), 62);
  return true;
}
function paintedFrog(x, y, pose) { // pose: sit, jump, croak
  if (!art('frog-' + pose)) return false;
  pic(art('frog-' + pose), x, y, pose === 'jump' ? 78 : 54);
  return true;
}
function paintedPuddle(x, y, w) { if (!art('frog-puddle')) return false; pic(art('frog-puddle'), x, y + 14, w); return true; }
function paintedDog(x, y, t, bark, face = -1) {
  if (!allArt('dog-body', 'dog-tail', 'dog-head-happy', 'dog-head-bark')) return false;
  const step = Math.abs(Math.sin(t * 10)) * 3;
  ctx.save(); ctx.translate(x, y - step); ctx.scale(face, 1);
  pic(art('dog-tail'), -44, -40, 40, false, Math.sin(t * 20) * 0.4);
  pic(art('dog-body'), 0, 0, 100);
  pic(art(bark ? 'dog-head-bark' : 'dog-head-happy'), 50, -34, 76);
  ctx.restore();
  return true;
}
