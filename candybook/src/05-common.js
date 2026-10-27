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
