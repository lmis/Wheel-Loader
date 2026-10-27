// ---------- page: the candy factory ----------
const BELT = { x0: 60, x1: 1170, y: 512, speed: 55, pipeX: 96 };
const M = { x: 640, top: 250, bot: 480, hw: 140 };
M.domeY = M.top - 4; M.chimX = M.x + 105;
const ROBOT_BASE = BELT.y + 22;
const LOOKS = [
  { name: 'bolt', col: '#5acdc3', bw: 92, bh: 104, neck: 12, hw: 88, hh: 72 },
  { name: 'gumdrop', col: '#ffa555', bw: 124, bh: 110, neck: 6, hw: 92, hh: 80 },
  { name: 'tinny', col: '#b9a0f5', bw: 74, bh: 84, neck: 28, hw: 100, hh: 76 },
];

class Belt {
  constructor(st) {
    Object.assign(this, BELT); this.st = st; this.offset = 0; this.candies = []; this.bulge = -1; this.nextPipe = 0.5; this.basket = [];
    for (let x = 140; x < BELT.x1 - 60; x += 150) this.candies.push({ x, col: CANDY[(x / 10) % 7 | 0], kind: (x / 150 | 0) % 4, hop: 0 });
  }
  drop(x, col, kind) { this.candies.push({ x, col, kind, hop: 1 }); }
  nearest(x, reach) {
    let best = null;
    for (const c of this.candies) if (!c.taken && Math.abs(c.x - x) < reach && (!best || Math.abs(c.x - x) < Math.abs(best.x - x))) best = c;
    return best;
  }
  update(dt) {
    this.offset += this.speed * dt;
    if (this.bulge >= 0) {
      this.bulge += dt / 0.9;
      if (this.bulge >= 1) { this.bulge = -1; this.drop(this.pipeX + 30, pick(CANDY), irnd(4)); sfx('pop', 0.1, 1.8); }
    } else if ((this.nextPipe -= dt) <= 0) { this.bulge = 0; this.nextPipe = rnd(1.8, 3.2); }
    this.candies = this.candies.filter(c => {
      if (c.taken) return false;
      c.x += this.speed * dt; c.hop = Math.max(0, c.hop - dt * 3);
      if (c.x > this.x1 - 6) {
        this.st.candy(c.x, this.y - 12, 90, -120, { col: c.col, candy: c.kind, size: 15, life: 0.55, vr: 6 });
        if (this.basket.length < 14) this.basket.push({ x: rnd(-24, 24), col: c.col, kind: c.kind, rot: rnd(-1, 1) });
        else Object.assign(this.basket[irnd(14)], { col: c.col, kind: c.kind });
        return false;
      }
      return true;
    });
  }
  draw(t) {
    for (let x = this.x0 + 40; x < this.x1; x += 220) {
      line(x, this.y + 30, x - 12, 690, 10, '#9696aa'); line(x + 20, this.y + 30, x + 32, 690, 10, '#9696aa');
    }
    const py0 = 40, py1 = this.y - 70, pc = '#78c8e6';
    line(this.pipeX, py0, this.pipeX, py1, 34, pc);
    curve(this.pipeX, py1, this.pipeX, py1 + 40, this.pipeX + 34, py1 + 44, 34, pc);
    line(this.pipeX - 8, py0, this.pipeX - 8, py1, 6, 'rgba(255,255,255,0.55)', 0);
    if (this.bulge >= 0) {
      const y = lerp(py0 + 20, py1 + 30, this.bulge), x = this.pipeX + (this.bulge > 0.85 ? (this.bulge - 0.85) / 0.15 * 26 : 0);
      const w = 1 + 0.15 * Math.sin(this.bulge * 40);
      ell(x, y, 24 * w, 28 / w, pc); ell(x - 8, y - 6, 6, 9, 'rgba(255,255,255,0.6)', 0);
    }
    rrect(this.pipeX + 30, py1 + 26, 16, 36, 4, '#5aaad2');
    rrect(this.x0, this.y, this.x1 - this.x0, 34, 17, '#5a5668');
    rrect(this.x0 + 4, this.y + 2, this.x1 - this.x0 - 8, 8, 4, '#827e92', 0);
    ctx.fillStyle = 'rgba(70,40,47,0.45)';
    for (let x = this.x0 + 20 + this.offset % 40; x < this.x1 - 16; x += 40) ctx.fillRect(x, this.y + 4, 4, 26);
    for (const x of [this.x0 + 17, this.x1 - 17]) {
      circ(x, this.y + 17, 13, '#c8c8d7');
      const a = this.offset / 13;
      line(x - Math.cos(a) * 10, this.y + 17 - Math.sin(a) * 10, x + Math.cos(a) * 10, this.y + 17 + Math.sin(a) * 10, 3, INK, 0);
    }
    for (const c of this.candies) drawCandy(c.x, this.y - 12 - Math.sin(c.hop * Math.PI) * 14, Math.sin(t * 3 + c.x) * 0.05, 15, c.col, c.kind);
    const bx = 1222, by = 626;
    this.basket.forEach((c, i) => drawCandy(bx + c.x, by - Math.floor(i / 5) * 12 + 4, c.rot, 13, c.col, c.kind));
    rrect(bx - 46, by, 92, 66, 12, '#d6a05a');
    for (let y = by + 12; y < by + 60; y += 14) line(bx - 40, y, bx + 40, y, 3, '#b47d41', 0);
    rrect(bx - 52, by - 6, 104, 14, 7, '#e6b46e');
  }
}

class Robot {
  constructor(page, kind, x) {
    Object.assign(this, { page, kind, x, look: LOOKS[kind], blink: 1 + kind * 1.3, cool: 1.5 + kind, phase: 'idle', pt: 0,
      target: null, held: null, side: 1, surprise: 0, ant: 0, antV: 0, grabX: 0 });
  }
  get st() { return this.page.st; }
  kick(s) { this.antV += s * 14; }
  height() { return this.look.bh + this.look.neck + this.look.hh; }
  update(dt) {
    const crew = this.page.crew;
    this.blink -= dt; if (this.blink < -0.13) this.blink = rnd(2, 5);
    this.surprise = Math.max(0, this.surprise - dt);
    this.antV += (-this.ant * 60 - this.antV * 5) * dt; this.ant += this.antV * dt;
    if (crew.gag === 'dance') this.antV += Math.sin(this.st.t * 16) * 6 * dt * 60;
    if (crew.gag && this.phase !== 'chew') return;
    this.pt += dt;
    const belt = this.page.belt;
    switch (this.phase) {
      case 'idle':
        if ((this.cool -= dt) <= 0) {
          const c = belt.nearest(this.x + 30, 60);
          if (c) { this.target = c; this.phase = 'reach'; this.pt = 0; this.side = c.x < this.x ? -1 : 1; }
        }
        break;
      case 'reach':
        if (!this.target || this.target.taken) { this.phase = 'idle'; this.cool = 0.5; break; }
        if (this.pt > 0.35) { this.held = this.target; this.target = null; this.held.taken = true; this.grabX = this.held.x; this.phase = 'bring'; this.pt = 0; }
        break;
      case 'bring': if (this.pt > 0.4) { this.phase = 'chew'; this.pt = 0; } break;
      case 'chew':
        if (Math.floor((this.pt - dt) / 0.22) !== Math.floor(this.pt / 0.22)) {
          const [mx, my] = this.mouthPos(); this.st.crumbs(mx, my, this.held.col, 4); sfx('chomp', 0.35, rnd(1, 1.4), (this.x - 640) / 700);
        }
        if (this.pt > 1.1) {
          this.held = null; this.phase = 'happy'; this.pt = 0;
          const [hx, hy] = this.headPos(); this.st.hearts(hx, hy - this.look.hh / 2, 1);
          if (Math.random() < 0.35) { this.st.say('YUM!', hx, hy - this.look.hh - 20, 40, '#ff7896'); sfx('yum', 0.4, rnd(1, 1.3)); }
          this.kick(0.4);
        }
        break;
      case 'happy': if (this.pt > 0.6) { this.phase = 'idle'; this.cool = rnd(2, 5.5); } break;
    }
  }
  gagT() { const c = this.page.crew; return c.gag === 'heads' ? c.gt - this.kind * 0.3 : c.gag === 'burp' ? c.gt - this.kind * 0.6 : c.gt; }
  pose() {
    const t = this.st.t, i = this.kind;
    let dy = Math.sin(t * 2.6 + i * 2) * 3, lean = 0;
    if (this.phase === 'happy') dy -= Math.sin(this.pt / 0.6 * Math.PI) * 18;
    if (this.surprise > 0) dy -= Math.sin(this.surprise * Math.PI) * 26;
    if (this.page.crew.gag === 'dance') { const ph = t * 9 + i * 1.4; dy -= Math.abs(Math.sin(ph)) * 22; lean = Math.sin(ph) * 10; }
    return [dy, lean];
  }
  headLift() {
    if (this.page.crew.gag !== 'heads') return 0;
    const t = this.gagT();
    if (t < 0) return 0;
    return 95 * Math.sin(Math.PI * clamp(t / 1.6, 0, 1)) * (1 + 0.25 * Math.sin(t * 18) * Math.exp(-t * 2));
  }
  headPos() {
    const [dy, lean] = this.pose(), l = this.look;
    const h = l.bh + l.neck + l.hh / 2 + this.headLift(), a = lean * Math.PI / 180;
    return [this.x + Math.sin(a) * h, ROBOT_BASE + dy - Math.cos(a) * h];
  }
  mouthPos() { const [hx, hy] = this.headPos(); return [hx, hy + this.look.hh * 0.22]; }
  mouthOpen() {
    const crew = this.page.crew;
    if (this.phase === 'bring') return ease(this.pt / 0.3);
    if (this.phase === 'chew') return 0.5 + 0.5 * Math.cos(this.pt / 0.22 * TAU);
    if (crew.gag === 'burp') { const t = this.gagT(); if (t > 0.45 && t < 1.1) return 1.3; }
    if (crew.gag === 'dance') return 0.5;
    if (this.surprise > 0) return 1;
    return 0;
  }
  lookAt() {
    if (this.page.machine.lollipop() > 0.2) return [M.x, M.domeY - 120];
    if (this.target) return [this.target.x, BELT.y];
    if (this.held) return this.mouthPos();
    return [this.page.dino.x, this.page.dino.y - 60];
  }
  drawBody() {
    const [dy, lean] = this.pose(), l = this.look, lift = this.headLift();
    push(this.x, ROBOT_BASE + dy, lean);
    this.drawTorso();
    const ny0 = -l.bh, ny1 = -l.bh - l.neck - lift;
    if (lift > 3 || this.kind === 2) for (let i = 0; i < 8; i++) ell(0, lerp(ny0, ny1, (i + 0.5) / 8), 13, 4, '#c8c8d7', 2.5);
    else rrect(-12, ny1, 24, l.neck + 6, 4, '#aaaabe');
    pop();
    const [hx, hy] = this.headPos();
    let sq = 1;
    if (this.page.crew.gag === 'burp') { const t = this.gagT(); if (t > 0 && t < 0.45) sq = 1 + 0.18 * ease(t / 0.45); }
    push(hx, hy, lean + this.ant * 0.4, sq, 1 / Math.sqrt(sq));
    this.drawHead();
    pop();
  }
  drawTorso() {
    const l = this.look, t = this.st.t;
    if (this.kind === 0) {
      rrect(-l.bw / 2, -l.bh, l.bw, l.bh, 14, l.col);
      rrect(-28, -l.bh + 18, 56, 40, 8, '#f0f8f5');
      const fill = this.phase === 'happy' || this.phase === 'chew' ? 0.9 : 0.3 + 0.2 * Math.sin(t);
      rrect(-22, -l.bh + 24 + 28 * (1 - fill), 44, 28 * fill, 4, '#ff7896', 0);
      for (let i = 0; i < 3; i++) circ(-20 + i * 20, -l.bh + 76, 6, Math.floor(t * 3 + i) % 3 === 0 ? '#ffe664' : '#78828c');
    } else if (this.kind === 1) {
      ell(0, -l.bh / 2, l.bw / 2, l.bh / 2, l.col);
      ell(0, -l.bh / 2 + 8, l.bw * 0.32, l.bh * 0.3, shade(l.col, 0.45), 2.5);
      heart(0, -l.bh / 2 + 2, 9, '#ff6e8c', 2.5);
    } else {
      rrect(-l.bw / 2, -l.bh, l.bw, l.bh, 10, l.col);
      for (let i = 0; i < 3; i++) line(-20, -l.bh + 20 + i * 16, 20, -l.bh + 20 + i * 16, 5, shade(l.col, -0.3), 0);
      const a = t * 3, cx = l.bw / 2 + Math.cos(a) * 16, cy = -l.bh / 2 + Math.sin(a) * 16;
      line(l.bw / 2, -l.bh / 2, cx, cy, 4, '#ffd250'); circ(cx, cy, 5, '#ff7878');
    }
  }
  drawHead() {
    const l = this.look, t = this.st.t, w = l.hw, h = l.hh, col = shade(l.col, 0.12), open = this.mouthOpen();
    const [lx, ly] = this.lookAt(), [hx, hy] = this.headPos();
    let ex = lx - hx, ey = ly - hy; const d = Math.hypot(ex, ey) || 1; ex /= d; ey /= d;
    if (this.page.crew.gag === 'heads' && this.headLift() > 10) { ex = Math.cos(t * 14); ey = Math.sin(t * 14); }
    const blink = this.blink < 0 && !this.surprise, big = 1 + 0.3 * Math.min(1, this.surprise * 3);
    const eye = (x, y, r, screen) => {
      if (blink) { line(x - r, y, x + r, y, 4, INK, 0); return; }
      if (screen) { rrect(x - r * 0.5, y - r * big, r, r * 2 * big, r * 0.5, '#8cffc8', 0); return; }
      circ(x, y, r * big, WHITE); circ(x + ex * r * 0.4, y + ey * r * 0.4, r * 0.48, INK, 0);
      circ(x + ex * r * 0.4 - r * 0.18, y + ey * r * 0.4 - r * 0.18, r * 0.16, WHITE, 0);
    };
    const mouth = (y, mw) => {
      const mh = 4 + open * h * 0.22;
      rrect(-mw / 2, y - mh / 2, mw, mh, Math.min(mh, mw) / 2, '#78283c');
      if (open > 0.3) rrect(-mw / 2 + 4, y - mh / 2 + 1, mw - 8, 5, 2, WHITE, 0);
    };
    if (this.kind === 0) {
      const ax = this.ant * 18;
      curve(0, -h / 2, ax * 0.3, -h / 2 - 24, ax, -h / 2 - 44, 4, '#aaaabe');
      glow(ax, -h / 2 - 46, 22, '#ff5a6e', 0.35 + 0.25 * Math.sin(t * 5));
      circ(ax, -h / 2 - 46, 9, '#ff5a6e');
      rrect(-w / 2 - 8, -10, 10, 22, 4, '#aaaabe'); rrect(w / 2 - 2, -10, 10, 22, 4, '#aaaabe');
      rrect(-w / 2, -h / 2, w, h, 16, col);
      eye(-19, -8, 14); eye(19, -8, 14); mouth(h * 0.26, 36);
      circ(-30, 12, 6, 'rgba(255,120,150,0.5)', 0); circ(30, 12, 6, 'rgba(255,120,150,0.5)', 0);
    } else if (this.kind === 1) {
      const sp = t * (this.page.crew.gag ? 30 : 6);
      line(0, -h / 2 - 4, 0, -h / 2 - 22, 5, '#aaaabe');
      push(0, -h / 2 - 24, 0, Math.cos(sp) || 0.01, 1); ell(0, 0, 34, 6, '#ffdc5a'); pop();
      circ(0, -h / 2 - 24, 5, '#ff6e8c');
      ell(0, 0, w / 2, h / 2, col); eye(0, -10, 22); mouth(h * 0.28, 44);
    } else {
      for (const s of [-1, 1]) {
        line(s * w * 0.25, -h / 2, s * (w * 0.4 + this.ant * 10), -h / 2 - 26, 3, '#aaaabe');
        circ(s * (w * 0.4 + this.ant * 10), -h / 2 - 28, 6, '#ffd250'); circ(s * (w / 2 + 6), 0, 10, '#ffa0be');
      }
      rrect(-w / 2, -h / 2, w, h, 12, col); rrect(-w / 2 + 10, -h / 2 + 9, w - 20, h - 18, 10, '#283c46');
      rect(-w / 2 + 12, -h / 2 + 10 + (t * 30) % (h - 20), w - 24, 2, 'rgba(140,255,200,0.3)');
      eye(-18 + ex * 3, -6 + ey * 3, 10, true); eye(18 + ex * 3, -6 + ey * 3, 10, true);
      const g = '#8cffc8';
      if (open > 0.1) { const mh = 4 + open * 12; rrect(-14, 16 - mh / 2, 28, mh, 3, g, 0); }
      else { rect(-14, 12, 5, 5, g); rect(-9, 16, 18, 5, g); rect(9, 12, 5, 5, g); }
    }
  }
  drawArms() {
    const [dy, lean] = this.pose(), l = this.look, t = this.st.t, crew = this.page.crew, base = ROBOT_BASE + dy;
    let shY = base - l.bh + 26;
    for (const s of [-1, 1]) {
      let sx = this.x + s * (l.bw / 2 - 4);
      if (this.kind === 1) { sx = this.x + s * (l.bw / 2 - 10); shY = base - l.bh * 0.55; }
      sx += lean * 0.6;
      let hx = sx + s * 26, hy = shY + 48 + Math.sin(t * 2.6 + this.kind + s) * 3;
      const eating = s === this.side;
      if (eating && this.phase === 'reach' && this.target) { const u = ease(this.pt / 0.35); hx = lerp(hx, this.target.x, u); hy = lerp(hy, BELT.y - 12, u); }
      if (eating && (this.phase === 'bring' || this.phase === 'chew')) {
        const [mx, my] = this.mouthPos(), u = this.phase === 'bring' ? ease(this.pt / 0.4) : 1;
        hx = lerp(this.grabX, mx + s * 10, u); hy = lerp(BELT.y - 12, my + 8, u);
      }
      if (crew.gag === 'dance') { const ph = t * 9 + this.kind * 1.4 + (s + 1) * 1.6; hx = sx + s * (40 + Math.sin(ph) * 16); hy = shY - 40 - Math.cos(ph) * 30; }
      if (crew.gag === 'heads' && this.headLift() > 10) { hx = sx + s * 44; hy = shY - 10; }
      if (this.surprise > 0) { hx = sx + s * 40; hy = shY - 50; }
      if (this.waveT > 0 && s === 1) { hx = sx + 30 + Math.sin(t * 14) * 10; hy = shY - 60; }
      curve(sx, shY, (sx + hx) / 2 + s * 20, Math.max(shY, hy) + 10, hx, hy, 9, '#aaaabe');
      if (eating && this.held) drawCandy(hx, hy - 6, 0.3 * s, 13, this.held.col, this.held.kind);
      if (this.kind === 2) { line(hx, hy, hx + s * 8, hy - 10, 5, '#ffd250'); line(hx, hy, hx + s * 10, hy + 6, 5, '#ffd250'); }
      else circ(hx, hy, 11, shade(l.col, 0.3));
    }
  }
}

class Machine {
  constructor(page) { this.page = page; this.gag = null; this.gt = 0; this.pending = false; this.puffT = 1; this.chute = 2; this.btn = 0; this.lid = 0; this.shot = 0; this.bubble = 0; }
  get st() { return this.page.st; }
  press() {
    this.btn = 0.25; sfx('click', 1);
    if (this.gag && this.gt < 2) { this.st.sparkles(M.x, M.top + 150, 6, 30); this.pending = true; return; }
    this.start();
  }
  start() {
    const list = ['fountain', 'lollipop', 'bubble'];
    this.gag = list[this.st.next[0]++ % 3]; this.gt = 0; this.shot = 0; this.pending = false;
    sfx({ fountain: 'windup', lollipop: 'slide', bubble: 'blow' }[this.gag], 1);
  }
  lollipop() {
    if (this.gag !== 'lollipop' || this.gt < 0.4) return 0;
    let out = springOut((this.gt - 0.4) * 1.1);
    if (this.gt > 3.3) out *= 1 - ease((this.gt - 3.3) / 0.6);
    return out;
  }
  update(dt) {
    const st = this.st;
    this.btn = Math.max(0, this.btn - dt); this.lid = Math.max(0, this.lid - dt * 2.5);
    if ((this.puffT -= dt) <= 0) { this.puffT = rnd(1.8, 3); st.puff(M.chimX, M.top - 70, rnd(12, 18), '#fff0f5', 0); }
    if ((this.chute -= dt) <= 0) { this.chute = rnd(2.5, 4.5); st.candy(M.x + M.hw + 30, M.bot - 30, 40, -60, { life: 3 }); }
    if (!this.gag) return;
    this.gt += dt;
    const t = this.gt, at = k => t > k && t - dt <= k;
    if (this.gag === 'fountain') {
      if (at(0.6)) { this.lid = 1; st.shake = 0.6; sfx('boing', 1); st.say('BOING!', M.x, M.domeY - 110, 72, '#ffd250'); }
      if (t > 0.6 && t < 1.6 && this.shot < (t - 0.6) * 30) {
        this.shot++;
        const a = -Math.PI / 2 + rnd(-0.9, 0.9), s = rnd(380, 640);
        st.candy(M.x + rnd(-30, 30), M.domeY - 50, Math.cos(a) * s, Math.sin(a) * s);
        if (this.shot % 5 === 0) sfx('pop', 0.4, rnd(0.8, 1.5));
      }
    } else if (this.gag === 'lollipop') {
      if (at(0.4)) { this.lid = 1; st.say('WOW!', M.x + 190, M.domeY - 100, 64, '#ff8cc8'); st.sparkles(M.x, M.domeY - 130, 14, 80); sfx('sparkle', 1); }
    } else if (this.gag === 'bubble') {
      this.bubble = ease(t / 2.6) * 90;
      if (at(2.6)) {
        this.bubble = 0; st.shake = 0.5; sfx('pop', 1, 0.6); sfx('sparkle', 0.6);
        st.say('POP!', M.chimX + 40, M.top - 130, 80, '#ff78b4');
        st.confetti(M.chimX + 40, M.top - 160, 90, 520);
        for (let i = 0; i < 6; i++) st.candy(M.chimX + 40 + rnd(-50, 50), M.top - 160 + rnd(-30, 30), rnd(-300, 300), rnd(-300, 0));
      }
    }
    if (t > { fountain: 3, lollipop: 4, bubble: 3.6 }[this.gag]) { this.gag = null; this.bubble = 0; if (this.pending) this.start(); }
  }
  squash() {
    const t = this.st.t, s = 0.012 * Math.sin(t * 2.3);
    let sx = 1 - s, sy = 1 + s;
    if (this.gag === 'fountain') {
      if (this.gt < 0.6) { const k = ease(this.gt / 0.6); sx += 0.08 * k + Math.sin(this.gt * 80) * 0.012 * k; sy -= 0.1 * k; }
      else { const k = springOut(this.gt - 0.6) - 1; sx -= 0.06 * k; sy += 0.08 * k; }
    }
    if (this.gag === 'bubble' && this.gt < 2.6) sy += 0.03 * Math.sin(this.gt * 14);
    return [sx, sy];
  }
  drawBack() {
    const t = this.st.t, spin = this.gag ? 4 : 1;
    gear(M.x - M.hw - 8, M.top + 70, 40, t * 0.8 * spin, 10, '#fac850');
    gear(M.x - M.hw - 50, M.top + 150, 26, -t * 0.8 * spin * 46 / 30 + 0.2, 7, '#82d2c8');
    gear(M.x + M.hw + 10, M.top + 40, 30, -t * 0.9 * spin, 8, '#b496f0');
  }
  draw() {
    const t = this.st.t, [sx, sy] = this.squash(), l = this.lollipop();
    if (l > 0) {
      const h = l * 120, wob = Math.sin(this.gt * 9) * 8 * Math.exp(-this.gt);
      push(M.x + wob, M.domeY - 30 - h, wob);
      line(0, 0, 0, h + 30, 16, WHITE);
      for (let i = 0; i < 5; i++) line(-8, 30 + i * h / 5, 8, 50 + i * h / 5, 5, '#ff7896', 0);
      circ(0, 0, 70, '#ff6ea0');
      for (let i = 0; i < 28; i++) { const a = i * 0.55 + this.gt * 4, r = i * 2.4; circ(Math.cos(a) * r, Math.sin(a) * r, 9, CANDY[i % 7], 0); }
      shine(-30, -34, 14);
      pop();
    }
    // chimney, puffing, and the bubblegum bubble
    const cp = this.puffT < 0.2 ? Math.sin(this.puffT / 0.2 * Math.PI) : 0;
    push(M.chimX, M.top, 0, 1 + cp * 0.12, 1 - cp * 0.08);
    rrect(-16, -80, 32, 90, 4, '#aa6ec8'); rrect(-24, -92, 48, 18, 6, '#c88ce6');
    pop();
    if (this.bubble > 0) {
      const r = this.bubble, wb = 1 + 0.06 * Math.sin(this.gt * 11), cy = M.top - 96 - r * 0.7, bx = M.chimX + r * 0.45;
      ell(bx, cy, r * wb, r / wb, '#ff8cc8'); ell(bx - r * 0.35, cy - r * 0.4, r * 0.25, r * 0.15, 'rgba(255,255,255,0.8)', 0);
    }
    push(M.x, M.bot, 0, sx, sy); this.drawBody(); pop();
    // the dome full of gumballs
    const lid = Math.sin(this.lid * Math.PI) * 40;
    push(M.x, M.bot - (M.bot - M.domeY) * sy, 0, sx, sy);
    ctx.beginPath(); ctx.arc(0, 0, 95, Math.PI, TAU); ctx.closePath(); paint('rgba(210,240,255,0.95)', 3);
    for (let i = 0; i < 34; i++) {
      const row = Math.floor(i / 9), cx = (i % 9 - 4) * 19 * (1 - row * 0.14);
      let cy = -12 - row * 20 + Math.sin(t * 6 + i) * 1.5;
      if (this.gag) cy += Math.sin(t * 30 + i * 2) * 4;
      circ(cx, cy, 9, CANDY[i % 7], 2.5); circ(cx - 3, cy - 3, 2.5, 'rgba(255,255,255,0.8)', 0);
    }
    ell(-45, -55, 16, 26, 'rgba(255,255,255,0.5)', 0);
    push(0, -96 - lid, Math.sin(this.lid * 12) * 8); rrect(-36, -6, 72, 14, 7, '#ff6e8c'); circ(0, -12, 11, '#ffdc5a'); pop();
    rrect(-108, -4, 216, 20, 10, '#ff6e8c');
    pop();
  }
  drawBody() {
    const t = this.st.t, h = M.bot - M.top, hw = M.hw;
    line(-hw + 30, 0, -hw + 20, BELT.y - M.bot + 40, 14, '#9696aa'); line(hw - 30, 0, hw - 20, BELT.y - M.bot + 40, 14, '#9696aa');
    rrect(-hw, -h, 2 * hw, h, 34, '#ff96af');
    ctx.fillStyle = 'rgba(255,255,255,0.18)';
    for (let x = -hw + 30; x < hw - 10; x += 40) ctx.fillRect(x, -h + 30, 14, h - 60);
    [-58, 58].forEach((dx, i) => {
      const cx = dx, cy = -h + 62;
      circ(cx, cy, 34, '#ffe678'); circ(cx, cy, 26, WHITE);
      const a = this.gag ? -Math.PI / 2 + Math.sin(t * 14 + i * 2) * 1.3 : -Math.PI / 2 + Math.sin(t * 1.7 + i * 2) * 0.6 + Math.sin(t * 13 + i) * 0.05;
      line(cx, cy, cx + Math.cos(a) * 20, cy + Math.sin(a) * 20, 4, '#e64650', 0); circ(cx, cy, 4, INK, 0);
    });
    const mw = this.gag ? 44 : 30;
    rrect(-mw, -h + 112, 2 * mw, 18 + mw * 0.2, 10, INK); rrect(-mw + 8, -h + 114, 2 * mw - 16, 6, 3, '#ff7896', 0);
    for (let i = 0; i < 7; i++) {
      const x = -90 + i * 30, c = CANDY[i];
      const on = this.gag ? Math.floor(t * 20 + i) % 2 === 0 : Math.floor(t * 6) % 7 === i || Math.floor(t * 6 + 3) % 7 === i;
      if (on) glow(x, -h + 160, 22, c, 0.6);
      circ(x, -h + 160, 8, on ? shade(c, 0.3) : shade(c, -0.45));
    }
    const p = Math.sin(Math.min(1, this.btn / 0.25) * Math.PI) * 6;
    rrect(-44, -h + 188, 88, 24, 8, '#c8c8d7');
    ell(0, -h + 186 + p, 30, 14 - p * 0.8, '#f03c46'); ell(-10, -h + 182 + p, 9, 4, 'rgba(255,255,255,0.7)', 0);
    curve(hw - 20, -60, hw + 30, -60, hw + 34, -14, 26, '#c8c8d7');
  }
}

class FactoryPage {
  constructor() {
    this.name = 'factory';
    this.text = 'Little Dino spent the whole day at the candy factory with Bolt, Gumdrop and Tinny. Now the sun is getting low. Time to go home.';
    this.st = new Stage(); this.st.floor = 692;
    this.belt = new Belt(this.st);
    this.st.catcher = b => {
      if (b.x > BELT.x0 + 10 && b.x < BELT.x1 - 10 && b.y > BELT.y - 6 && b.y < BELT.y + 14 && this.belt.candies.length < 30) { this.belt.drop(b.x, b.col, b.candy); return true; }
      return false;
    };
    this.machine = new Machine(this);
    this.crew = { gag: null, gt: 0, pending: false, notes: 0 };
    this.robots = [new Robot(this, 0, 245), new Robot(this, 1, 965), new Robot(this, 2, 1112)];
    this.dino = new Dino(this.st, {
      x: 520, ground: () => 684, tricks: ['jump', 'roar', 'catch', 'ride'],
      host: { belt: this.belt, roarAt: 640, startle: k => this.robots.forEach(r => { r.surprise = k; r.kick(1.2); }),
        cheer: () => this.robots.forEach(r => { const [hx, hy] = r.headPos(); this.st.hearts(hx, hy - 40, 1); r.kick(0.5); }) },
    });
    this.clouds = Array.from({ length: 7 }, () => ({ x: rnd(0, W), y: 90 + rnd(0, 90), s: rnd(0.6, 1.2), v: rnd(6, 14) }));
    this.mouse = 6; this.bird = -1; this.birdIn = 4; this.birdY = 120; this.disco = 0;
  }
  press(k) {
    if (k === 0) this.machine.press();
    else if (k === 1) this.crewPress();
    else this.dino.press();
  }
  crewPress() {
    const c = this.crew;
    sfx('bloop', 1);
    this.robots.forEach(r => r.kick(0.6));
    if (c.gag && c.gt < 2) { c.pending = true; this.robots.forEach(r => this.st.sparkles(r.x, ROBOT_BASE - r.height() - 10, 3, 30)); return; }
    this.crewStart();
  }
  crewStart() {
    const c = this.crew;
    c.gag = ['dance', 'heads', 'burp'][this.st.next[1]++ % 3]; c.gt = 0; c.pending = false;
    if (c.gag === 'dance') { this.disco = 4.4; sfx('tune', 0.9); this.st.say('DANCE!', 640, 170, 70, '#82dcff'); }
  }
  updateCrew(dt) {
    const c = this.crew;
    if (!c.gag) return;
    c.gt += dt;
    this.robots.forEach((r, i) => {
      const t1 = r.gagT(), was = t1 - dt, [hx, hy] = r.headPos();
      if (c.gag === 'heads' && was < 0 && t1 >= 0) { sfx('boing', 0.8, 0.8 + i * 0.2); this.st.say('BOING!', hx, hy - 130, 46, r.look.col); }
      if (c.gag === 'burp') {
        if (was < 0.45 && t1 >= 0.45) { sfx('burp', 1, 0.7 + i * 0.25); this.st.say('BURP!', hx + 50, hy - 60, 44, shade(r.look.col, 0.3)); this.st.bubbles(hx, hy + 20, 7); }
        if (was < 1.2 && t1 >= 1.2) { this.st.hearts(hx, hy - 40, 2); sfx('giggle', 0.5, 0.9 + i * 0.15); }
      }
    });
    if (c.gag === 'dance' && (c.notes -= dt) <= 0) { c.notes = 0.25; const [hx, hy] = pick(this.robots).headPos(); this.st.notes(hx, hy - 40, 1); }
    if (c.gt > { dance: 4.4, heads: 2.8, burp: 4.4 }[c.gag]) { c.gag = null; if (c.pending) this.crewStart(); }
  }
  update(dt) {
    this.st.update(dt);
    this.disco = Math.max(0, this.disco - dt);
    for (const c of this.clouds) { c.x += c.v * dt; if (c.x > W + 80) { c.x = -80; c.y = 90 + rnd(0, 90); } }
    this.mouse -= dt; if (this.mouse < -4.5) this.mouse = rnd(7, 14);
    if (this.bird >= 0) { this.bird += 70 * dt; if (this.bird > 400) { this.bird = -1; this.birdIn = rnd(6, 12); } }
    else if ((this.birdIn -= dt) <= 0) { this.bird = 110; this.birdY = 110 + rnd(0, 70); }
    this.machine.update(dt); this.belt.update(dt); this.updateCrew(dt);
    this.robots.forEach(r => r.update(dt));
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t;
    this.drawRoom(t);
    this.machine.drawBack();
    this.robots.forEach(r => r.drawBody());
    this.machine.draw();
    st.draw(0);
    this.belt.draw(t);
    this.robots.forEach(r => r.drawArms());
    st.draw(1);
    this.dino.draw();
    st.draw(2);
    this.drawLights(t);
    st.drawWords();
  }
  drawRoom(t) {
    const floorY = 600;
    rect(0, 0, W, floorY, grad(0, 0, 0, floorY, ['#ffe2cd', '#fac8b9']));
    for (let x = 0; x < W; x += 64) {
      rect(x, 0, 30, floorY, 'rgba(255,190,185,0.35)');
      for (let y = 40; y < floorY - 60; y += 80) star(x + 15, y + (x / 64 % 2 ? 40 : 0), 6, 0, 4, 'rgba(255,255,255,0.55)', 0);
    }
    rect(0, floorY - 70, W, 70, '#96cdbe');
    for (let x = 10; x < W; x += 80) rrect(x, floorY - 60, 60, 48, 6, '#a5dacb', 2.5);
    rect(0, floorY - 74, W, 6, INK);
    // mouse hole and the mouse
    ell(420, floorY, 30, 38, INK, 0);
    if (this.mouse < 0) {
      const p = -this.mouse, out = ease(p / 0.4) * ease((4.5 - p) / 0.4), look = Math.sin(p * 2.2) * 6, hx = 420 + look, hy = floorY - 6 - out * 20;
      for (const s of [-1, 1]) { circ(hx + s * 12, hy - 14, 9, '#beb4b9'); circ(hx + s * 12, hy - 14, 5, '#ffafbe', 0); }
      ell(hx, hy, 15, 13, '#beb4b9'); circ(hx - 5 + look * 0.3, hy - 2, 2.6, INK, 0); circ(hx + 5 + look * 0.3, hy - 2, 2.6, INK, 0);
      circ(hx + look * 0.5, hy + 6, 3, '#ff8ca0', 2);
      const w = Math.sin(t * 30) * 2;
      line(hx + 4, hy + 6, hx + 22, hy + 3 + w, 1.2, INK, 0); line(hx - 4, hy + 6, hx - 22, hy + 3 - w, 1.2, INK, 0);
    }
    // floor
    rect(0, floorY, W, H - floorY, grad(0, floorY, 0, H, ['#d69664', '#ba7850']));
    ctx.fillStyle = 'rgba(70,40,47,0.3)';
    for (let y = floorY + 26, row = 0; y < H; y += 34, row++) { ctx.fillRect(0, y, W, 3); for (let x = (row % 2) * 90; x < W; x += 180) ctx.fillRect(x, y - 26, 3, 26); }
    this.drawWindow(t, 150, true); this.drawWindow(t, 930, false);
    this.drawClock(t, 1205, 150);
    rect(0, 0, W, 26, '#a06950'); rect(0, 26, W, 4, INK);
    line(0, 44, W, 44, 12, '#c8c8d7');
    for (let x = 120; x < W; x += 260) { rrect(x, 36, 18, 16, 3, '#aaaabe'); gear(x + 9, 66, 8, t * 0.7 + x, 6, '#f05a5a'); }
    for (let side = 0; side < 2; side++) {
      const x0 = side * W / 2, x1 = x0 + W / 2;
      ctx.beginPath(); ctx.moveTo(x0, 34); ctx.quadraticCurveTo((x0 + x1) / 2, 34 + 92, x1, 34); ctx.lineWidth = 2; ctx.strokeStyle = INK; ctx.stroke();
      for (let i = 1; i < 9; i++) {
        const u = i / 9, x = lerp(x0, x1, u), y = 34 + 46 * 4 * u * (1 - u), sw = Math.sin(t * 2.2 + i * 0.9 + side) * 7;
        poly([x - 17, y, x + 17, y, x + sw, y + 38], CANDY[(i + side * 3) % 7], 2.5);
      }
    }
  }
  drawWindow(t, x, sunny) {
    const wy = 70, ww = 210, wh = 210;
    rrect(x - 12, wy - 12, ww + 24, wh + 24, 24, '#fff8eb');
    ctx.save(); ctx.beginPath(); ctx.rect(x, wy, ww, wh); ctx.clip();
    rect(x, wy, ww, wh, grad(0, wy, 0, wy + wh, ['#78c3fa', '#c8ebff']));
    if (sunny) {
      const sx = x + 150, sy = wy + 60;
      for (let i = 0; i < 12; i++) {
        const a = t * 0.3 + i / 12 * TAU, r2 = 58 + 4 * Math.sin(t * 3 + i);
        line(sx + Math.cos(a) * 42, sy + Math.sin(a) * 42, sx + Math.cos(a) * r2, sy + Math.sin(a) * r2, 6, '#ffd246', 0);
      }
      circ(sx, sy, 34, '#ffdc5a');
      line(sx - 14, sy - 4, sx - 6, sy - 4, 3, INK, 0); line(sx + 6, sy - 4, sx + 14, sy - 4, 3, INK, 0);
      ctx.beginPath(); ctx.arc(sx, sy + 2, 11, 0.3, Math.PI - 0.3); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.stroke();
    } else {
      ell(x + 60, wy + wh + 40, 150, 90, '#82cd6e', 0); ell(x + 190, wy + wh + 60, 140, 100, '#6ebe64', 0);
      const mx = x + 150, my = wy + 120;
      poly([mx - 14, my + 60, mx + 14, my + 60, mx, my - 4], '#faf5eb');
      for (let i = 0; i < 4; i++) { const a = t * 1.2 + i * Math.PI / 2; line(mx, my, mx + Math.cos(a) * 44, my + Math.sin(a) * 44, 7, '#f07878'); }
      circ(mx, my, 5, '#faf5eb');
    }
    for (const c of this.clouds) { const cx = sunny ? c.x : (c.x + 400) % (W + 160) - 80; if (cx > x - 90 && cx < x + ww + 90) cloud(cx, c.y, c.s); }
    if (sunny && this.bird >= 0) { const f = Math.sin(t * 12) * 8, bx = this.bird + x - 110, by = this.birdY + Math.sin(t * 2) * 6; line(bx - 12, by - f, bx, by, 3, INK, 0); line(bx, by, bx + 12, by - f, 3, INK, 0); }
    ctx.restore();
    rect(x + ww / 2 - 5, wy, 10, wh, '#fff8eb'); rect(x, wy + wh / 2 - 5, ww, 10, '#fff8eb');
    rrect(x - 26, wy + wh + 8, ww + 52, 18, 6, '#fff8eb');
    if (sunny) {
      const px = x + 30, sway = Math.sin(t * 1.5) * 4, fy = wy + wh - 42;
      line(px, wy + wh + 2, px + sway, fy, 3, '#5aaa50');
      for (let i = 0; i < 5; i++) { const a = i / 5 * TAU + t * 0.2; circ(px + sway + Math.cos(a) * 9, fy + Math.sin(a) * 9, 7, '#ff78a0'); }
      circ(px + sway, fy, 5, '#ffd750'); rrect(px - 14, wy + wh - 6, 28, 22, 4, '#dc7850');
    }
  }
  drawClock(t, x, y) {
    push(x, y + 30, Math.sin(t * 2.4) * 12); line(0, 0, 0, 70, 3, '#dcaa3c'); circ(0, 74, 11, '#f0c346'); pop();
    rrect(x - 38, y - 44, 76, 84, 18, '#aa6e50'); circ(x, y, 30, WHITE);
    for (let i = 0; i < 12; i++) { const a = i / 12 * TAU; circ(x + Math.cos(a) * 24, y + Math.sin(a) * 24, 2, INK, 0); }
    const sec = Math.floor(t) / 60 * TAU, min = t / 600 * TAU;
    line(x, y, x + Math.sin(min) * 14, y - Math.cos(min) * 14, 4, INK, 0);
    line(x, y, x + Math.sin(sec) * 22, y - Math.cos(sec) * 22, 2, '#e64650', 0); circ(x, y, 4, INK, 0);
  }
  drawLights(t) {
    [420, 860].forEach((x, i) => {
      const a = Math.sin(t * 1.3 + i * 1.7) * 0.07, lx = x + Math.sin(a) * 120, ly = 30 + Math.cos(a) * 120;
      line(x, 30, lx, ly, 3, INK, 0);
      push(lx, ly, a * 57.3);
      const col = this.disco > 0 ? alpha(CANDY[(Math.floor(t * 6) + i * 3) % 7], 0.2) : 'rgba(255,240,180,0.12)';
      poly([-20, 20, 20, 20, 170, 520, -170, 520], col, 0);
      poly([-28, 22, 28, 22, 0, -12], '#f06e6e'); circ(0, 24, 9, '#fff5be');
      pop();
    });
    if (this.disco > 0) for (let i = 0; i < 9; i++) {
      const a = t * 1.5 + i * 0.7, x = 640 + Math.cos(a) * 520 * Math.sin(i * 1.3 + t * 0.4), y = 300 + Math.sin(a * 1.3) * 200;
      ctx.globalAlpha = 0.25 * Math.min(1, this.disco); circ(x, y, 40, CANDY[i % 7], 0); ctx.globalAlpha = 1;
    }
  }
}
