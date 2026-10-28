// ---------- Little Dino ----------
// Hops about in little runs, looks around, gobbles candy it finds on the
// floor, and does a trick when Ctrl is pressed. Each page says which tricks
// fit there and where the ground is.
const DINO_GREEN = { body: '#78cd6e', belly: '#f5eeaa', spike: '#ffa05a' };
const MAMA_COLS = { body: '#b48ce6', belly: '#fbe7f2', spike: '#ff7fae' };

class Dino {
  constructor(st, o) {
    this.st = st;
    Object.assign(this, { x: 500, face: 1, minX: 90, maxX: W - 190, ground: () => 684, tricks: ['jump', 'roar', 'wave', 'spin'],
      cols: DINO_GREEN, scale: 1, wander: true, host: {} }, o);
    this.y = this.ground(this.x);
    this.hopping = false; this.hp = 0; this.squash = 0; this.rest = 1; this.hopsLeft = 0; this.chew = 0; this.blink = 2;
    this.trick = null; this.tt = 0; this.stage = 0; this.pending = false; this.rot = 0; this.onBelt = false;
    this.sleepy = 0; this.next = 0;
  }
  press() {
    sfx('squeak', 0.8, 1);
    this.squash = Math.max(this.squash, 0.5);
    if (this.trick) { this.pending = true; this.st.hearts(this.x, this.y - 150 * this.scale, 1); return; }
    this.start();
  }
  start(name) {
    this.trick = name || this.tricks[this.next % this.tricks.length];
    if (!name) this.next++;
    this.tt = 0; this.stage = 0; this.pending = false; this.hopping = false;
    if (!this.onBelt) this.y = this.ground(this.x);
  }
  hop(x1, y1, dur, h) {
    Object.assign(this, { hopping: true, hp: 0, hdur: dur, hx0: this.x, hy0: this.y, hx1: x1, hy1: y1, hh: h });
    if (Math.abs(x1 - this.x) > 1) this.face = Math.sign(x1 - this.x);
  }
  // follow the ground while hopping along it
  stepHop(dt, follow = true) {
    if (!this.hopping) return false;
    this.hp += dt / this.hdur;
    const p = Math.min(1, this.hp);
    this.x = lerp(this.hx0, this.hx1, p);
    const base = follow ? this.ground(this.x) : lerp(this.hy0, this.hy1, p);
    this.y = base - 4 * this.hh * p * (1 - p);
    if (this.hp >= 1) { this.hopping = false; this.y = follow ? this.ground(this.x) : this.hy1; this.squash = Math.min(1, 0.35 + this.hh / 200); return true; }
    return false;
  }
  mouth() { return [this.x + this.face * 62 * this.scale, this.y - 98 * this.scale]; }
  update(dt) {
    const st = this.st;
    this.squash = Math.max(0, this.squash - dt * 3);
    this.chew = Math.max(0, this.chew - dt);
    this.blink -= dt; if (this.blink < -0.12) this.blink = rnd(2, 4.5);
    if (this.trick) { this.updateTrick(dt); return; }
    if (this.stepHop(dt)) this.eatNearby();
    if (this.hopping || !this.wander) return;
    if ((this.rest -= dt) > 0) return;
    if (this.hopsLeft <= 0) {
      this.rest = rnd(0.8, 2.8); this.hopsLeft = 2 + irnd(4);
      const c = st.bits.find(b => b.kind === 'candy' && b.resting && b.life > 1.5);
      if (c) { this.face = Math.sign(c.x - this.x) || 1; return; }
      if (Math.random() < 0.45) this.face = -this.face;
      return;
    }
    this.hopsLeft--;
    let x = this.x + this.face * rnd(50, 80) * this.scale;
    if (x < this.minX || x > this.maxX) { this.face = -this.face; x = this.x + this.face * 60; }
    this.hop(clamp(x, this.minX, this.maxX), 0, 0.42, 34);
    sfx('hop', 0.6, rnd(0.9, 1.2));
  }
  eatNearby() {
    const mx = this.x + this.face * 55;
    const b = this.st.bits.find(b => b.kind === 'candy' && b.resting && b.life > 0.3 && Math.abs(b.x - mx) < 50);
    if (b) { b.life = 0; this.gulp(b.col, false); }
  }
  gulp(col, loud) {
    const [mx, my] = this.mouth();
    this.chew = 0.9;
    this.st.crumbs(mx, my, col, 6);
    this.st.hearts(this.x, this.y - 150 * this.scale, 2);
    sfx('chomp', 0.8, 0.8);
    if (loud) { this.st.say('YUM!', this.x, this.y - 200, 60, '#ff7896'); sfx('yum', 1); }
  }
  updateTrick(dt) {
    const st = this.st, host = this.host;
    this.tt += dt;
    const t = this.tt;
    let done = false;
    switch (this.trick) {
      case 'jump':
        if (this.stage === 0) {
          this.squash = 0.7 * ease(t / 0.35);
          if (t > 0.35) {
            this.stage = 1; this.squash = 0;
            this.hop(clamp(this.x + this.face * 200, this.minX, this.maxX), 0, 1.15, 300);
            sfx('whee', 1);
          }
        } else if (this.stage === 1) {
          this.rot = this.face * 360 * ease(Math.min(1, this.hp) * 1.1);
          if (this.stepHop(dt)) {
            this.rot = 0; this.stage = 2; this.squash = 1; this.tt = 0;
            st.shake = 0.8; st.dust(this.x, this.y, 10);
            st.say('THUMP!', this.x, this.y - 210, 66, '#ffbe5a');
            sfx('thump', 1);
            for (const b of st.bits) if (b.kind === 'candy' && b.resting) { b.resting = false; b.vy = -rnd(200, 380); b.vr = rnd(-6, 6); }
            host.startle?.(0.5);
          }
        } else done = t > 0.6;
        break;
      case 'roar':
        if (t < 0.05 && host.roarAt) this.face = Math.sign(host.roarAt - this.x) || 1;
        if (t > 0.45 && t - dt <= 0.45) {
          sfx('rawr', 1);
          const [mx, my] = this.mouth();
          st.say('RAWR!', mx + this.face * 90, my - 60, 84, '#8ce678');
          st.shake = 0.5; host.startle?.(0.7);
        }
        if (t > 0.5 && t < 1.4 && Math.floor((t - dt) * 12) !== Math.floor(t * 12)) {
          const [mx, my] = this.mouth();
          st.add({ kind: 'puff', layer: 2, x: mx, y: my, vx: this.face * rnd(240, 340), vy: rnd(-50, 50), size: rnd(6, 10), col: WHITE, life: 0.6 });
        }
        if (t > 1.8 && t - dt <= 1.8) { host.cheer?.(); sfx('giggle', 0.9); }
        done = t > 2.5;
        break;
      case 'wave':
        if (t < 0.05) this.face = host.waveTo ? (Math.sign(host.waveTo - this.x) || 1) : this.face;
        if (t > 0.3 && t - dt <= 0.3) {
          st.say(host.waveWord || 'HELLO!', this.x + this.face * 40, this.y - 210, 58, '#8cd2ff');
          sfx('giggle', 0.7, 1.2); host.waved?.();
        }
        done = t > 2.2;
        break;
      case 'spin':
        if (t < 0.9) { this.rot = 0; this.spinSx = Math.cos(t / 0.9 * TAU * 2); }
        if (t - dt <= 0 || (t > 0.45 && t - dt <= 0.45)) sfx('whee', 0.6, 1.4);
        if (t > 0.9 && t - dt <= 0.9) { this.spinSx = 1; this.squash = 0.8; st.sparkles(this.x, this.y - 70, 10, 60); sfx('sparkle', 0.7); st.say('TA-DA!', this.x, this.y - 200, 56, '#ffd25a'); }
        done = t > 1.6;
        break;
      case 'yawn':
        if (t > 0.1 && t - dt <= 0.1) sfx('yawn', 1);
        if (t > 1.0 && t - dt <= 1.0) { st.zzz(this.x + 30, this.y - 150); }
        if (t > 1.6 && t - dt <= 1.6) st.zzz(this.x + 40, this.y - 160);
        done = t > 2.6;
        break;
      case 'catch':
        if (this.stage === 0) {
          this.catchX = clamp(this.x + this.face * rnd(180, 320), this.minX, this.maxX);
          if (Math.abs(this.catchX - this.x) < 100) this.catchX = clamp(this.x - this.face * 250, this.minX, this.maxX);
          const cx = this.catchX + Math.sign(this.catchX - this.x) * 62;
          this.catchBit = st.candy(cx, -30, 0, 0, { vr: 4, size: 24, candy: 1 + irnd(3), life: 6, g: 260, layer: 2 });
          sfx('whistle', 1); this.stage = 1;
        } else if (this.stage === 1) {
          if (this.stepHop(dt) || !this.hopping) {
            if (Math.abs(this.catchX - this.x) < 4) { this.stage = 2; this.face = Math.sign(this.catchBit.x - this.x) || 1; }
            else { this.hop(this.x + clamp(this.catchX - this.x, -70, 70), 0, 0.22, 18); sfx('hop', 0.4, 1.3); }
          }
        } else if (this.stage === 2) {
          const b = this.catchBit, [mx, my] = this.mouth();
          if (b.life > 0 && b.y > my - 30 && Math.abs(b.x - mx) < 60) { b.life = 0; this.gulp(b.col, true); this.stage = 3; this.tt = 0; }
          else if (b.life <= 0 || b.resting) { this.stage = 3; this.tt = 0; }
        } else done = t > 1.1;
        done = done || t > 6;
        break;
      case 'ride': { // the factory belt
        const belt = host.belt;
        if (this.stage === 0) {
          this.hop(clamp(this.x, 160, 820), belt.y, 0.55, 110); this.face = 1; this.onBelt = true; this.stage = 1; sfx('hop', 0.7, 0.8);
        } else if (this.stage === 1) {
          if (this.stepHop(dt, false)) { this.stage = 2; this.tt = 0; st.say('WHEE!', this.x, this.y - 200, 62, '#82dcff'); sfx('whee', 1, 1.2); }
        } else if (this.stage === 2) {
          this.x += belt.speed * 2.2 * dt; this.face = 1;
          const c = belt.nearest(this.x + 60, 26);
          if (c) { c.taken = true; this.gulp(c.col, false); }
          if (t > 2.6 || this.x > belt.x1 - 120) { this.stage = 3; this.hop(Math.min(this.x + 110, this.maxX), this.ground(this.x + 110), 0.5, 70); }
        } else if (this.stepHop(dt, false)) { this.onBelt = false; st.dust(this.x, this.y, 6); done = true; }
        break;
      }
      default: // a page's own trick
        done = host.trick ? host.trick(this, t, dt) : true;
    }
    if (done) {
      this.trick = null; this.rot = 0; this.spinSx = 1; this.rest = 0.6; this.hopsLeft = 2;
      if (!this.onBelt) this.y = this.ground(this.x);
      if (this.pending) this.start();
    }
  }
  draw() {
    const t = this.st.t, s = this.squash;
    let sx = 1 + s * 0.28, sy = 1 - s * 0.26;
    if (this.hopping && this.trick !== 'jump') { const k = Math.sin(this.hp * Math.PI); sx = 1 - k * 0.08; sy = 1 + k * 0.12; }
    const roaring = this.trick === 'roar' && this.tt > 0.4 && this.tt < 1.5;
    if (this.trick === 'roar' && this.tt < 0.45) { sy += 0.1 * ease(this.tt / 0.45); sx -= 0.04 * ease(this.tt / 0.45); }
    if (roaring) sx += Math.sin(t * 50) * 0.02;
    // shadow
    const gy = this.onBelt && !this.hopping ? this.y : this.ground(this.x);
    const hgt = clamp((gy - this.y) / 300, 0, 1);
    if (!(this.onBelt && !this.hopping)) ell(this.x, gy + 4, 52 * (1 - hgt * 0.5) * this.scale, 10 * (1 - hgt * 0.5), 'rgba(70,40,47,0.2)', 0);
    push(this.x, this.y, 0, this.face * sx * this.scale * (this.spinSx ?? 1), sy * this.scale);
    if (this.rot) { ctx.translate(0, -55); ctx.rotate(this.rot * this.face * Math.PI / 180); ctx.translate(0, 55); }
    const catching = this.trick === 'catch' && this.stage === 2;
    let open = 0;
    if (roaring) open = 1; else if (catching) open = 0.8;
    else if (this.chew > 0) open = 0.5 + 0.5 * Math.cos(this.chew * 20);
    else if (this.trick === 'ride' && this.stage === 2) open = 0.6;
    else if (this.trick === 'yawn' && this.tt > 0.15 && this.tt < 1.3) open = Math.sin((this.tt - 0.15) / 1.15 * Math.PI) * 1.1;
    else if (this.trick === 'wave' && this.tt > 0.3 && this.tt < 1.2) open = 0.5;
    let arm = Math.sin(t * 4) * 4;
    if ((this.trick === 'ride' && this.stage === 2) || (this.trick === 'jump' && this.stage === 1)) arm = -18;
    const waving = this.trick === 'wave' && this.tt < 2;
    drawDino({
      t, cols: this.cols, open, roaring, walk: this.hopping ? Math.sin(this.hp * Math.PI) * 8 : 0,
      blink: this.blink < 0 || (this.trick === 'yawn' && this.tt > 0.3 && this.tt < 1.4), lookUp: this.trick === 'catch' && this.stage >= 1,
      noseUp: catching ? -14 : 0, arm, wave: waving ? Math.sin(this.tt * 14) : null, sleepy: this.sleepy, bow: this.bow,
    });
    pop();
  }
}

// drawDino draws a dino in its own frame: feet at 0,0, facing right.
function drawDino(o) {
  const { t, cols } = o;
  const wag = Math.sin(t * 5) * 8, walk = o.walk || 0;
  if (o.skates) { drawSkates(t, o.skateSpin ?? t * 12); ctx.save(); ctx.translate(0, -14); }
  // the far arm, raised when cheering
  if (o.cheer) line(-8, -70, -14, -112 + Math.sin(t * 14) * 6, 7, shade(cols.body, -0.12));
  // tail
  ctx.beginPath(); ctx.moveTo(-26, -28); ctx.quadraticCurveTo(-70, -30 + wag * 0.3, -100, -62 + wag);
  ctx.quadraticCurveTo(-74, -46 + wag * 0.4, -30, -62); ctx.closePath(); paint(cols.body, 3);
  // back leg
  ell(-16, -14 + walk * 0.3, 18, 15, shade(cols.body, -0.12));
  // spikes
  for (let i = 0; i < 5; i++) {
    const a = Math.PI * 1.05 + i * 0.32;
    const bx = Math.cos(a) * 44, by = -52 + Math.sin(a) * 42, tx = Math.cos(a) * 64, ty = -52 + Math.sin(a) * 62;
    const px = -Math.sin(a) * 11, py = Math.cos(a) * 11;
    poly([bx - px, by - py, bx + px, by + py, tx, ty], cols.spike, 3);
  }
  ell(0, -52, 48, 42, cols.body);
  if (o.pj) { // striped pyjamas
    ctx.save(); ctx.beginPath(); ctx.ellipse(0, -52, 48, 42, 0, 0, TAU); ctx.clip();
    rect(-50, -96, 100, 90, o.pj);
    for (let y = -92; y < 0; y += 14) rect(-50, y, 100, 6, shade(o.pj, 0.45));
    ctx.restore(); ell(0, -52, 48, 42, null, 3);
  }
  ell(14, -44, 26, 28, o.pj ? shade(o.pj, 0.6) : cols.belly, 2.5);
  ctx.globalAlpha = 0.4;
  for (let i = 0; i < 3; i++) line(4, -58 + i * 12, 26, -58 + i * 12, 2, shade(cols.belly, -0.3), 0);
  ctx.globalAlpha = 1;
  circ(-22, -66, 6, shade(cols.body, -0.18), 0); circ(-32, -46, 4, shade(cols.body, -0.18), 0); circ(-10, -82, 4, shade(cols.body, -0.18), 0);
  ell(20, -12 - walk * 0.3, 17, 14, cols.body);
  circ(30, -7 - walk * 0.3, 3, WHITE, 0);
  // head
  const bob = o.roaring ? -6 : Math.sin(t * 3) * 2;
  push(40, -98 + bob, o.noseUp || 0);
  drawDinoHead(o);
  pop();
  if (o.bag) { // Mama's red work bag
    line(-6, -86, 18, -40, 3, '#8c3246', 0);
    rrect(6, -44, 30, 24, 6, '#e6505a'); line(12, -44, 30, -44, 3, '#b43c46', 0);
  }
  // little arm, or waving, or both up
  if (o.cheer) line(26, -66, 44, -110 + Math.cos(t * 14) * 6, 7, cols.body);
  else if (o.wave !== null && o.wave !== undefined) line(26, -66, 46, -104 + o.wave * 8, 7, cols.body);
  else if (o.reach) line(26, -66, 62, -60, 7, cols.body);
  else line(26, -66, 42, -60 + (o.arm || 0), 7, cols.body);
  if (o.skates) ctx.restore();
}

// drawSkates: Papa's rollerblades, with red wheels that glow
function drawSkates(t, spin) {
  for (const fx of [-16, 20]) {
    glow(fx, 8, 34, '#ff3c3c', 0.35 + 0.15 * Math.sin(t * 20 + fx));
    rrect(fx - 18, -14, 34, 16, 6, '#3c3c50');
    rrect(fx - 4, -26, 18, 16, 5, '#3c3c50');
    for (const wx of [-12, -1, 10]) {
      circ(fx + wx, 7, 5.5, '#ff4646', 2);
      line(fx + wx, 7, fx + wx + Math.cos(spin) * 4, 7 + Math.sin(spin) * 4, 1.5, '#ffd0d0', 0);
    }
  }
}

// drawDinoHead draws the head in its own frame (centre of the head at 0,0).
function drawDinoHead(o) {
  const cols = o.cols;
  if (!o.hair && !o.helmet) poly([-22, -24, -8, -33, -32, -42], cols.spike, 3);
  circ(0, 0, 36, cols.body);
  ell(26, 8, 30, 22, cols.body);
  circ(0, 0, 33.5, cols.body, 0); // one outline round head and snout
  ell(26, 8, 28.5, 20.5, cols.body, 0);
  const open = o.open || 0;
  if (open > 0.05) {
    const oh = 4 + open * 22;
    ell(30, 16, 22, oh / 2 + 2, '#8c3246', 3);
    ell(30, 16 + oh * 0.25, 12, oh * 0.2, '#ff788c', 0);
    poly([18, 16 - oh / 2, 24, 16 - oh / 2, 21, 16 - oh / 2 + 6], WHITE, 0);
    poly([36, 16 - oh / 2, 42, 16 - oh / 2, 39, 16 - oh / 2 + 6], WHITE, 0);
  } else curve(14, 18, 30, 28, 48, 16, 3, INK, 0);
  circ(40, 0, 3, INK, 0); circ(48, 2, 3, INK, 0);
  if (o.blink || o.sleepy > 0.9) curve(-4, -10, 6, -4, 16, -10, 3, INK, 0);
  else if (o.roaring) line(-2, -14, 16, -6, 4, INK, 0);
  else {
    circ(6, -10, 12, WHITE);
    const py = o.lookUp ? -5 : 0;
    circ(11, -10 + py, 6, INK, 0); circ(9, -12 + py, 2, WHITE, 0);
    if (o.sleepy > 0) { // heavy eyelid
      ctx.save(); ctx.beginPath(); ctx.arc(6, -10, 12, 0, TAU); ctx.clip();
      rect(-8, -24, 28, 26 * o.sleepy, cols.body); ctx.restore();
      line(-6, -22 + 26 * o.sleepy, 18, -22 + 26 * o.sleepy, 3, INK, 0);
    }
    if (o.bow) { line(-2, -24, -8, -30, 2, INK, 0); line(4, -24, 4, -31, 2, INK, 0); line(10, -22, 14, -28, 2, INK, 0); }
  }
  circ(-6, 12, 7, 'rgba(255,130,150,0.5)', 0);
  if (o.hair) { // a fluffy blonde tuft
    for (const [hx, hy, r] of [[-14, -30, 10], [-2, -37, 11], [11, -33, 10], [20, -26, 7]]) circ(hx, hy, r, '#ffe07a', 2.5);
    for (const [hx, hy, r] of [[-14, -30, 8], [-2, -37, 9], [11, -33, 8]]) circ(hx, hy, r, '#ffe07a', 0);
    curve(-4, -44, 2, -54, 10, -48, 3, '#ffe07a', 2);
  }
  if (o.helmet) { // a red helmet, safety first
    ctx.beginPath(); ctx.arc(2, -4, 41, Math.PI * 1.02, Math.PI * 1.98); ctx.closePath(); paint('#e63c46', 3);
    line(-26, -18, 30, -18, 4, '#ffffff', 0); circ(2, -36, 5, '#ffd250', 2);
    line(-30, -4, -12, 26, 3, INK, 0);
  }
  if (o.bow) { // Mama's bow
    poly([-6, -36, -30, -50, -28, -24], '#ff6f9f'); poly([-6, -36, 16, -54, 16, -26], '#ff6f9f'); circ(-6, -36, 7, '#ffa0c0');
  }
}

// ---------- puppets: Mama, Papa, Leo and Louis ----------
// A puppet stands where a page puts it, breathes, blinks, and plays little
// actions (wave, talk, cheer, kiss, spin, yawn, sleep, hop) when asked.
const PAPA_COLS = { body: '#78cd6e', belly: '#f5eeaa', spike: '#ffa05a' };
const LEO_COLS = { body: '#8ad0ec', belly: '#f2f8ff', spike: '#ffb45a' };
const LOUIS_COLS = { body: '#ffb48c', belly: '#fff3e2', spike: '#8ad0ec' };
class Puppet {
  constructor(st, o) {
    Object.assign(this, { x: 0, y: 0, face: 1, scale: 1, cols: PAPA_COLS, look: {}, name: '' }, o);
    this.st = st; this.a = null; this.at = 0; this.dur = 0; this.blinkT = rnd(1, 4); this.squash = 0;
    this.jy = 0; this.vy = 0; this.tx = null; this.walkPh = 0; this.rot = 0;
  }
  do(a, dur = 1.6) { this.a = a; this.at = 0; this.dur = dur; if (a === 'hop') this.jump(); return this; }
  is(a) { return this.a === a; }
  jump(v = 480) { if (this.jy === 0) { this.vy = -v; } }
  walkTo(x, speed = 170) { this.tx = x; this.speed = speed; }
  mouth() { return [this.x + this.face * 62 * this.scale, this.y + this.jy - 98 * this.scale]; }
  head() { return [this.x + this.face * 40 * this.scale, this.y + this.jy - 110 * this.scale]; }
  update(dt) {
    this.blinkT -= dt; if (this.blinkT < -0.13) this.blinkT = rnd(2, 5);
    if (this.a) { this.at += dt; if (this.at > this.dur) this.a = null; }
    this.squash = Math.max(0, this.squash - dt * 3);
    if (this.vy || this.jy) {
      this.vy += 1500 * dt; this.jy += this.vy * dt;
      if (this.jy >= 0) { this.jy = 0; if (this.vy > 250) this.squash = 0.6; this.vy = 0; }
    }
    if (this.tx !== null) {
      const d = this.tx - this.x;
      if (Math.abs(d) < 3) { this.x = this.tx; this.tx = null; }
      else { this.face = Math.sign(d); this.x += Math.sign(d) * Math.min(Math.abs(d), this.speed * dt); this.walkPh += dt * 9; }
    }
  }
  draw(extra = {}) {
    const t = this.st.t, a = this.a, k = this.dur ? this.at / this.dur : 0;
    let sx = 1 + this.squash * 0.25, sy = 1 - this.squash * 0.25, rot = this.rot, bob = 0;
    const o = { t, cols: this.cols, ...this.look, ...extra, blink: this.blinkT < 0, arm: Math.sin(t * 4) * 4 };
    if (this.tx !== null && !this.look.skates) bob = -Math.abs(Math.sin(this.walkPh)) * 12 * this.scale;
    switch (a) {
      case 'wave': o.wave = Math.sin(this.at * 14); o.open = this.at < 1 ? 0.4 : 0; break;
      case 'talk': o.open = 0.3 + 0.3 * Math.sin(this.at * 20); break;
      case 'cheer': o.cheer = true; o.open = 0.6; bob -= Math.abs(Math.sin(this.at * 10)) * 14 * this.scale; break;
      case 'kiss': o.open = 0.25; rot += Math.sin(Math.PI * k) * 10; break;
      case 'spin': sx *= Math.cos(k * TAU * 2) || 0.01; break;
      case 'yawn': o.open = Math.sin(Math.PI * k) * 1.1; o.blink = k > 0.2 && k < 0.8; break;
      case 'sleep': o.blink = true; break;
      case 'sad': o.sleepy = 0.45; rot -= 4; break;
      case 'laugh': o.open = 0.5 + 0.3 * Math.sin(this.at * 25); bob -= Math.abs(Math.sin(this.at * 16)) * 5; break;
      case 'shake': rot += Math.sin(this.at * 22) * 6; break;
      case 'surprise': sy *= 1.1; sx *= 0.94; o.open = 1; break;
      case 'reach': o.reach = true; o.open = 0.3; break;
      case 'eat': o.open = 0.5 + 0.5 * Math.cos(this.at * 18); break;
      case 'hop': o.open = 0.4; break;
    }
    if (this.sleeping) o.blink = true;
    const gy = this.y;
    if (!this.noShadow) ell(this.x, gy + 4, 50 * this.scale, 9 * this.scale, 'rgba(70,40,47,0.2)', 0);
    push(this.x, gy + this.jy + bob, 0, this.face * sx * this.scale, sy * this.scale);
    if (rot) { ctx.translate(0, -55); ctx.rotate(rot * Math.PI / 180); ctx.translate(0, 55); }
    drawDino(o);
    pop();
  }
}
