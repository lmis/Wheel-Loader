// ---------- the story pages share this ----------
// A StoryPage has a stage, a cast of puppets, and three lists of tricks
// (pink, blue, green). Pages fill in act(name) for what happens at once,
// tick(dt, t, A) for what happens over time, and draw().
class StoryPage {
  constructor(o) {
    Object.assign(this, o);
    this.st = new Stage(); this.st.floor = o.floor ?? 700;
    this.acts = new Acts(o.durs || {});
    this.cast = [];
    this.dt = 1 / 60;
  }
  press(k) {
    const a = cycle(this.st, k, this.tricks[k]);
    this.acts.run(a);
    this.act && this.act(a, k);
  }
  at(n, k) { return this.acts.at(n, k, this.dt); }
  update(dt) {
    this.dt = dt;
    this.st.update(dt); this.acts.update(dt);
    for (const c of this.cast) c.update(dt);
    this.tick && this.tick(dt, this.st.t, this.acts);
  }
  every(key, period) { const t = this.st.t; return Math.floor(t / period) !== Math.floor((t - this.dt) / period); }
  finish() { this.st.draw(0); this.st.draw(1); this.st.draw(2); this.st.drawWords(); }
}
const mama = (st, o) => new Puppet(st, Object.assign({ cols: MAMA_COLS, scale: 1.25, look: { bow: true }, name: 'Mama', rig: MAMA_RIG }, o));
const papa = (st, o) => new Puppet(st, Object.assign({ cols: PAPA_COLS, scale: 1.3, name: 'Papa', rig: PAPA_RIG }, o));
const leo = (st, o) => new Puppet(st, Object.assign({ cols: LEO_COLS, scale: 0.62, look: { hair: true }, name: 'Léo', rig: LEO_RIG }, o));
const louis = (st, o) => new Puppet(st, Object.assign({ cols: LOUIS_COLS, scale: 0.62, look: { hair: true }, name: 'Louis', rig: LOUIS_RIG }, o));

// the dino train: a smiling engine and two carriages
function drawTrain(x, y, t, o = {}) {
  const bob = Math.sin(t * 9) * (o.moving ? 1.5 : 0.4), spin = o.moving ? t * 12 : 0;
  const cars = o.cars ?? 2;
  for (let i = 0; i < cars; i++) {
    const cx = x - 470 - i * 230;
    rrect(cx, y - 130 + bob, 210, 110, 18, i % 2 ? '#ffd25a' : '#8ad0ec');
    for (let w = 0; w < 3; w++) {
      const wx = cx + 18 + w * 64;
      rrect(wx, y - 112 + bob, 48, 44, 8, '#fff6e2');
      if (o.window && o.window.car === i && o.window.seat === w) o.window.draw(wx + 24, y - 68 + bob);
    }
    for (const wx of [cx + 40, cx + 170]) { circ(wx, y - 14, 18, '#3c3c50'); circ(wx, y - 14, 7, '#c8c8d7'); line(wx, y - 14, wx + Math.cos(spin) * 7, y - 14 + Math.sin(spin) * 7, 2, INK, 0); }
    line(cx + 210, y - 40, cx + 230, y - 40, 6, '#5a5a6e');
  }
  // the engine, with a friendly face
  rrect(x - 240, y - 150 + bob, 200, 130, 20, '#e6505a');
  rrect(x - 60, y - 120 + bob, 90, 100, 30, '#e6505a');
  rrect(x - 230, y - 190 + bob, 90, 50, 10, '#b43c46');
  rrect(x - 30, y - 170 + bob, 34, 54, 6, '#5a5a6e');
  rrect(x - 218, y - 178 + bob, 66, 32, 6, '#fff6e2');
  circ(x + 6, y - 88 + bob, 13, WHITE); circ(x + 9, y - 88 + bob, 6, INK, 0);
  ctx.beginPath(); ctx.arc(x + 4, y - 60 + bob, 14, 0.2, Math.PI - 0.6); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.stroke();
  circ(x + 24, y - 64 + bob, 6, 'rgba(255,140,160,0.6)', 0);
  glow(x + 30, y - 40, 50, '#fff3a0', o.lamp ? 0.6 : 0.2);
  for (const wx of [x - 200, x - 120, x - 10]) { circ(wx, y - 14, 20, '#3c3c50'); circ(wx, y - 14, 8, '#ffd250'); line(wx, y - 14, wx + Math.cos(spin) * 8, y - 14 + Math.sin(spin) * 8, 2, INK, 0); }
  line(x - 200, y - 14, x - 10, y - 14, 5, '#8c8ca0');
}
function drawRails(y, x0 = 0, x1 = W) {
  for (let x = x0; x < x1; x += 40) rect(x, y + 2, 26, 10, '#8c6450');
  rect(x0, y - 2, x1 - x0, 5, '#9a9ab4');
}
// Papa's trolley, a wire basket on glowing wheels; the twins ride inside
function drawTrolley(x, y, t, kids, spin = t * 14) {
  const tl = 150, top = y - 110;
  line(x + tl - 10, top + 10, x + tl + 30, top - 30, 6, '#9a9ab4');
  for (const wx of [x + 20, x + tl - 20]) { glow(wx, y - 8, 30, '#ff3c3c', 0.4); circ(wx, y - 8, 9, '#ff4646', 2.5); line(wx, y - 8, wx + Math.cos(spin) * 6, y - 8 + Math.sin(spin) * 6, 2, '#ffd0d0', 0); }
  line(x + 6, y - 20, x + tl - 6, y - 20, 5, '#9a9ab4');
  for (const k of kids) k();
  // basket front
  ctx.beginPath(); ctx.moveTo(x, top); ctx.lineTo(x + tl, top); ctx.lineTo(x + tl - 14, y - 22); ctx.lineTo(x + 14, y - 22); ctx.closePath();
  ctx.fillStyle = 'rgba(200,210,230,0.35)'; ctx.fill(); ctx.lineWidth = 4; ctx.strokeStyle = INK; ctx.stroke();
  ctx.lineWidth = 2; ctx.strokeStyle = '#9a9ab4';
  for (let i = 1; i < 6; i++) { ctx.beginPath(); ctx.moveTo(x + i * tl / 6, top); ctx.lineTo(x + 14 + i * (tl - 28) / 6, y - 22); ctx.stroke(); }
  for (let j = 1; j < 3; j++) { const yy = top + j * (y - 22 - top) / 3; ctx.beginPath(); ctx.moveTo(x + j * 5, yy); ctx.lineTo(x + tl - j * 5, yy); ctx.stroke(); }
  rrect(x - 4, top - 6, tl + 8, 10, 5, '#e6505a');
}
// rain streaks over the whole page
function drawRain(t, amount, col = 'rgba(200,220,255,0.55)') {
  ctx.strokeStyle = col; ctx.lineWidth = 2;
  const n = Math.floor(80 * amount);
  ctx.beginPath();
  for (let i = 0; i < n; i++) {
    const x = (i * 137.5 + t * 120) % (W + 100) - 50, y = (i * 61.8 * 7 + t * 900) % (H + 60) - 30;
    ctx.moveTo(x, y); ctx.lineTo(x - 8, y + 26);
  }
  ctx.stroke();
}
// a city skyline with windows that light up
function drawCity(t, y, col, lit, seed = 1) {
  for (let x = -20, i = 0; x < W + 40; i++) {
    const w = 60 + (i * 37 * seed) % 50, h = 90 + (i * 53 * seed) % 140;
    rect(x, y - h, w, h, col);
    for (let r = 0; r < Math.floor(h / 26) - 1; r++) for (let c = 0; c < Math.floor(w / 20); c++) {
      const on = ((i * 7 + r * 3 + c * 5) % 5) < lit * 5;
      if (on) rect(x + 6 + c * 20, y - h + 12 + r * 26, 9, 12, 'rgba(255,224,140,0.85)');
    }
    x += w + 6;
  }
}
// a cosy wall with a skirting board and a floor
function drawRoom(wallTop, wallBot, floorA, floorB, floorY = 600) {
  rect(0, 0, W, floorY, grad(0, 0, 0, floorY, [wallTop, wallBot]));
  rect(0, floorY, W, H - floorY, grad(0, floorY, 0, H, [floorA, floorB]));
  rect(0, floorY - 8, W, 10, shade(floorA, -0.2));
}
function drawWindow(x, y, w, h, sky, inside) {
  rrect(x - 10, y - 10, w + 20, h + 20, 14, '#fff6ea');
  ctx.save(); ctx.beginPath(); ctx.roundRect(x, y, w, h, 8); ctx.clip();
  rect(x, y, w, h, sky); inside && inside();
  ctx.restore();
  rect(x + w / 2 - 4, y, 8, h, '#fff6ea'); rect(x, y + h / 2 - 4, w, 8, '#fff6ea');
}
