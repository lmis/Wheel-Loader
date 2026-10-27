'use strict';
// ---------- storybook drawing kit ----------
// Everything is drawn in a 1280x720 page with flat colours and a soft brown
// ink line, under save/translate/rotate/scale so characters can squash,
// stretch, flip and spin.
const W = 1280, H = 720, TAU = Math.PI * 2;
const INK = '#46282f', WHITE = '#fffcf6';
let ctx = null; // the canvas being drawn on right now

const rnd = (a, b) => a + Math.random() * (b - a);
const irnd = n => Math.floor(Math.random() * n);
const clamp = (x, a, b) => Math.max(a, Math.min(b, x));
const lerp = (a, b, t) => a + (b - a) * t;
const ease = t => { t = clamp(t, 0, 1); return t * t * (3 - 2 * t); };
const springOut = t => t <= 0 ? 0 : 1 - Math.exp(-6 * t) * Math.cos(14 * t);
const pick = a => a[irnd(a.length)];
const CANDY = ['#ff5c78', '#ffaa3c', '#fade50', '#6ed278', '#5ab4f5', '#b978eb', '#ff8cc8'];

function push(x, y, rot = 0, sx = 1, sy = 1) {
  ctx.save(); ctx.translate(x, y);
  if (rot) ctx.rotate(rot * Math.PI / 180);
  if (sx !== 1 || sy !== 1) ctx.scale(sx, sy);
}
function pop() { ctx.restore(); }

function paint(fill, lw) {
  if (fill) { ctx.fillStyle = fill; ctx.fill(); }
  if (lw) { ctx.lineWidth = lw; ctx.strokeStyle = INK; ctx.stroke(); }
}
function circ(x, y, r, fill, lw = 3) { ctx.beginPath(); ctx.arc(x, y, Math.max(0.01, r), 0, TAU); paint(fill, lw); }
function ell(x, y, rx, ry, fill, lw = 3, rot = 0) {
  ctx.beginPath(); ctx.ellipse(x, y, Math.max(0.01, Math.abs(rx)), Math.max(0.01, Math.abs(ry)), rot, 0, TAU); paint(fill, lw);
}
function rrect(x, y, w, h, r, fill, lw = 3) {
  ctx.beginPath(); ctx.roundRect(x, y, w, h, Math.max(0, Math.min(r, Math.abs(w) / 2, Math.abs(h) / 2))); paint(fill, lw);
}
function rect(x, y, w, h, fill) { ctx.fillStyle = fill; ctx.fillRect(x, y, w, h); }
function poly(pts, fill, lw = 3) {
  ctx.beginPath(); ctx.moveTo(pts[0], pts[1]);
  for (let i = 2; i < pts.length; i += 2) ctx.lineTo(pts[i], pts[i + 1]);
  ctx.closePath(); paint(fill, lw);
}
// line: an inked noodle (a stroke with a darker stroke behind it)
function line(x1, y1, x2, y2, w, col, lw = 3) {
  ctx.beginPath(); ctx.moveTo(x1, y1); ctx.lineTo(x2, y2);
  if (lw) { ctx.lineWidth = w + lw * 2; ctx.strokeStyle = INK; ctx.stroke(); }
  ctx.lineWidth = w; ctx.strokeStyle = col; ctx.stroke();
}
function curve(x1, y1, cx, cy, x2, y2, w, col, lw = 3) {
  ctx.beginPath(); ctx.moveTo(x1, y1); ctx.quadraticCurveTo(cx, cy, x2, y2);
  if (lw) { ctx.lineWidth = w + lw * 2; ctx.strokeStyle = INK; ctx.stroke(); }
  ctx.lineWidth = w; ctx.strokeStyle = col; ctx.stroke();
}
function shine(x, y, r) { ell(x, y, r, r * 0.6, 'rgba(255,255,255,0.75)', 0); }
function starPath(x, y, r, rot, n, inner = 0.5) {
  ctx.beginPath();
  for (let i = 0; i < n * 2; i++) {
    const a = rot + i / (n * 2) * TAU, rr = i % 2 ? r * inner : r;
    ctx.lineTo(x + Math.cos(a) * rr, y + Math.sin(a) * rr);
  }
  ctx.closePath();
}
function star(x, y, r, rot, n, fill, lw = 3) { starPath(x, y, r, rot, n); paint(fill, lw); }
function heart(x, y, s, fill, lw = 3) {
  ctx.beginPath();
  ctx.moveTo(x, y + s * 1.1);
  ctx.bezierCurveTo(x - s * 1.6, y - s * 0.1, x - s * 0.9, y - s * 1.3, x, y - s * 0.45);
  ctx.bezierCurveTo(x + s * 0.9, y - s * 1.3, x + s * 1.6, y - s * 0.1, x, y + s * 1.1);
  paint(fill, lw);
  shine(x - s * 0.55, y - s * 0.45, s * 0.22);
}
function gear(x, y, r, rot, teeth, fill) {
  ctx.beginPath();
  const st = TAU / teeth;
  for (let i = 0; i < teeth; i++) {
    const a0 = rot + i * st;
    for (const [f, rr] of [[0, r], [0.15, r * 1.22], [0.45, r * 1.22], [0.6, r]])
      ctx.lineTo(x + Math.cos(a0 + f * st) * rr, y + Math.sin(a0 + f * st) * rr);
  }
  ctx.closePath(); paint(fill, 3);
  circ(x, y, r * 0.55, shade(fill, -0.15), 0);
  circ(x, y, r * 0.18, shade(fill, 0.35));
}
// shade lightens (k>0) or darkens (k<0) a #rrggbb colour
function shade(hex, k) {
  const n = parseInt(hex.slice(1), 16);
  let r = n >> 16, g = (n >> 8) & 255, b = n & 255;
  const t = k > 0 ? 255 : 0, a = Math.abs(k);
  r = Math.round(r + (t - r) * a); g = Math.round(g + (t - g) * a); b = Math.round(b + (t - b) * a);
  return '#' + ((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1);
}
function alpha(hex, a) {
  const n = parseInt(hex.slice(1), 16);
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${clamp(a, 0, 1)})`;
}
function grad(x0, y0, x1, y1, stops) {
  const g = ctx.createLinearGradient(x0, y0, x1, y1);
  stops.forEach((c, i) => g.addColorStop(i / (stops.length - 1), c));
  return g;
}
function glow(x, y, r, col, a) {
  const g = ctx.createRadialGradient(x, y, 0, x, y, r);
  g.addColorStop(0, alpha(col, a)); g.addColorStop(1, alpha(col, 0));
  ctx.fillStyle = g; ctx.fillRect(x - r, y - r, r * 2, r * 2);
}
// word: a bouncy onomatopoeia in the display face with a fat ink outline
function word(text, x, y, size, rot, fill) {
  if (size < 2) return;
  push(x, y, rot);
  ctx.font = `700 ${size}px Fredoka, "Baloo 2", system-ui, sans-serif`;
  ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
  ctx.lineWidth = size * 0.2; ctx.strokeStyle = INK; ctx.strokeText(text, 0, 0);
  ctx.fillStyle = fill; ctx.fillText(text, 0, 0);
  pop();
}
function cloud(x, y, s, col = WHITE, a = 1) {
  ctx.globalAlpha = a;
  const bumps = [[-26, 6, 18], [0, 0, 24], [26, 6, 18], [12, 12, 16], [-12, 12, 16]];
  ctx.beginPath();
  for (const [bx, by, r] of bumps) { ctx.moveTo(x + (bx + r) * s, y + by * s); ctx.arc(x + bx * s, y + by * s, r * s, 0, TAU); }
  ctx.lineWidth = 6 * s; ctx.strokeStyle = 'rgba(70,40,47,0.25)'; ctx.stroke();
  ctx.fillStyle = col; ctx.fill();
  ctx.globalAlpha = 1;
}

// drawCandy: a wrapped sweet, a lollipop, a gumball or a star
function drawCandy(x, y, rot, s, col, kind) {
  push(x, y, rot * 57.3);
  const light = shade(col, 0.4);
  switch (kind) {
    case 0:
      poly([-s * 0.7, 0, -s * 1.4, -s * 0.6, -s * 1.4, s * 0.6], light, 2.5);
      poly([s * 0.7, 0, s * 1.4, -s * 0.6, s * 1.4, s * 0.6], light, 2.5);
      ell(0, 0, s * 0.85, s * 0.62, col, 2.5);
      line(-s * 0.3, -s * 0.45, s * 0.1, s * 0.45, s * 0.18, light, 0);
      shine(-s * 0.35, -s * 0.25, s * 0.22);
      break;
    case 1:
      line(0, s * 0.6, 0, s * 1.8, s * 0.22, WHITE, 2.5);
      circ(0, 0, s * 0.8, col, 2.5);
      ctx.beginPath();
      for (let a = 0; a < 9; a += 0.3) ctx.lineTo(Math.cos(a) * a * s * 0.08, Math.sin(a) * a * s * 0.08);
      ctx.lineWidth = s * 0.16; ctx.strokeStyle = light; ctx.stroke();
      shine(-s * 0.3, -s * 0.35, s * 0.2);
      break;
    case 2:
      circ(0, 0, s * 0.72, col, 2.5);
      shine(-s * 0.25, -s * 0.3, s * 0.24);
      break;
    default:
      star(0, 0, s * 0.95, -Math.PI / 2, 5, col, 2.5);
      shine(-s * 0.2, -s * 0.25, s * 0.2);
  }
  pop();
}

// wrapText splits a story line into lines no wider than maxW
function wrapText(text, maxW) {
  const words = text.split(' '), lines = [];
  let cur = '';
  for (const w of words) {
    const t = cur ? cur + ' ' + w : w;
    if (ctx.measureText(t).width > maxW && cur) { lines.push(cur); cur = w; } else cur = t;
  }
  if (cur) lines.push(cur);
  return lines;
}
