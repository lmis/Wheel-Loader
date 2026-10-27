// ---------- the little things that fly about ----------
// Each page is a Stage with its own candies, confetti, steam, sparkles,
// hearts, notes, bubbles, crumbs, dust and words (BOING! RAWR! YUM!).
class Stage {
  constructor() {
    this.t = 0; this.bits = []; this.words = []; this.shake = 0;
    this.floor = 690;          // where flying candies come to rest
    this.catcher = null;       // (bit) => true if something caught it (the belt)
    this.next = [0, 0, 0];     // which trick each key does next
  }
  add(b) {
    if (this.bits.length > 700) return b;
    b.max = b.life; b.layer = b.layer ?? 1; b.rot = b.rot ?? 0; b.vr = b.vr ?? 0;
    b.vx = b.vx ?? 0; b.vy = b.vy ?? 0; b.g = b.g ?? 0;
    this.bits.push(b); return b;
  }
  candy(x, y, vx, vy, o = {}) {
    return this.add(Object.assign({ kind: 'candy', x, y, vx, vy, vr: rnd(-8, 8), size: rnd(13, 18), col: pick(CANDY),
      candy: irnd(4), life: 9, g: 900, layer: 1 }, o));
  }
  confetti(x, y, n, speed) {
    for (let i = 0; i < n; i++) {
      const a = rnd(0, TAU), s = rnd(0.3, 1) * speed;
      this.add({ kind: 'confetti', layer: 2, x, y, vx: Math.cos(a) * s, vy: Math.sin(a) * s - speed * 0.3,
        rot: rnd(0, 6), vr: rnd(-12, 12), size: rnd(6, 11), col: pick(CANDY), life: rnd(1.6, 2.6), g: 380 });
    }
  }
  sparkles(x, y, n, r, col) {
    for (let i = 0; i < n; i++) {
      const a = rnd(0, TAU), d = r * Math.random();
      this.add({ kind: 'sparkle', layer: 2, x: x + Math.cos(a) * d, y: y + Math.sin(a) * d, vx: Math.cos(a) * 60, vy: Math.sin(a) * 60 - 30,
        rot: rnd(0, 3), vr: rnd(-3, 3), size: rnd(7, 13), col: col || shade(pick(CANDY), 0.5), life: rnd(0.5, 0.9) });
    }
  }
  hearts(x, y, n) {
    for (let i = 0; i < n; i++)
      this.add({ kind: 'heart', layer: 2, x: x + rnd(-20, 20), y, vx: rnd(-25, 25), vy: rnd(-110, -70), rot: rnd(0, 6),
        size: rnd(8, 12), col: '#ff6082', life: rnd(1.2, 1.7) });
  }
  notes(x, y, n) {
    for (let i = 0; i < n; i++)
      this.add({ kind: 'note', layer: 2, x: x + rnd(-30, 30), y, vx: rnd(-30, 30), vy: rnd(-100, -70), rot: rnd(0, 6),
        size: 8, col: pick(CANDY), life: 1.6 });
  }
  puff(x, y, size, col, layer = 0, o = {}) {
    return this.add(Object.assign({ kind: 'puff', layer, x, y, vx: rnd(-12, 12), vy: rnd(-60, -35), size, col, life: rnd(1.6, 2.4) }, o));
  }
  dust(x, y, n) {
    for (let i = 0; i < n; i++) {
      const s = i % 2 ? 1 : -1;
      this.add({ kind: 'dust', layer: 2, x: x + s * rnd(0, 30), y, vx: s * rnd(60, 160), vy: rnd(-40, -10), size: rnd(10, 18),
        col: '#ecd6be', life: rnd(0.5, 0.8) });
    }
  }
  crumbs(x, y, col, n) {
    for (let i = 0; i < n; i++)
      this.add({ kind: 'crumb', layer: 2, x, y, vx: rnd(-90, 90), vy: rnd(-160, -40), size: rnd(2.5, 4.5), col, life: rnd(0.5, 0.9), g: 700 });
  }
  bubbles(x, y, n) {
    for (let i = 0; i < n; i++)
      this.add({ kind: 'bubble', layer: 2, x: x + rnd(-10, 10), y, vx: rnd(-50, 50), vy: rnd(-120, -60), rot: rnd(0, 6),
        size: rnd(8, 16), col: CANDY[i % CANDY.length], life: rnd(1.4, 2.2) });
  }
  zzz(x, y) {
    this.add({ kind: 'z', layer: 2, x, y, vx: 20, vy: -30, rot: 0, size: rnd(22, 30), col: '#a8c8ff', life: 2.6 });
  }
  say(text, x, y, size, col) { this.words.push({ text, x, y, size, col, life: 1.6, max: 1.6, rot: rnd(-10, 10) }); }

  update(dt) {
    this.t += dt;
    this.shake = Math.max(0, this.shake - dt * 2.5);
    const live = [];
    for (const b of this.bits) {
      b.life -= dt;
      if (b.life <= 0) { if (b.kind === 'bubble') { this.sparkles(b.x, b.y, 3, b.size); sfx('pop', 0.15, 1.6); } continue; }
      b.vy += b.g * dt;
      switch (b.kind) {
        case 'confetti': b.vx *= 1 - 2 * dt; b.vy = Math.min(b.vy, 120); b.x += Math.sin(b.life * 7 + b.rot) * 40 * dt; break;
        case 'puff': b.size += 22 * dt; b.vx *= 1 - dt; break;
        case 'heart': case 'note': case 'z': b.x += Math.sin(b.life * 5 + b.rot) * 30 * dt; break;
        case 'bubble': b.x += Math.sin(b.life * 3 + b.rot) * 40 * dt; b.vy *= 1 - 0.5 * dt; break;
        case 'dust': b.vx *= 1 - 5 * dt; b.size += 25 * dt; break;
      }
      b.x += b.vx * dt; b.y += b.vy * dt; b.rot += b.vr * dt;
      if (b.kind === 'candy') this.candyPhysics(b);
      if (b.life > 0) live.push(b);
    }
    this.bits = live;
    this.words = this.words.filter(w => (w.life -= dt) > 0);
  }
  candyPhysics(b) {
    if (b.resting) { b.vx = b.vy = b.vr = 0; return; }
    if (b.vy > 0 && this.catcher && this.catcher(b)) { b.life = 0; return; }
    const floor = this.floor - b.size * 0.4;
    if (b.y > floor && b.vy > 0) {
      b.y = floor;
      if (b.vy > 140) { b.vy *= -0.42; b.vx *= 0.7; b.vr *= 0.6; if (b.size > 12) sfx('tick', 0.12, rnd(0.8, 1.4)); }
      else b.resting = true;
    }
    if (b.x < 20 || b.x > W - 20) { b.vx = -b.vx * 0.6; b.x = clamp(b.x, 20, W - 20); }
  }
  draw(layer) {
    for (const b of this.bits) {
      if (b.layer !== layer) continue;
      const a = clamp(b.life / 0.5, 0, 1);
      ctx.globalAlpha = a;
      switch (b.kind) {
        case 'candy': drawCandy(b.x, b.y, b.rot, b.size, b.col, b.candy); break;
        case 'confetti':
          push(b.x, b.y, b.rot * 57, Math.cos(b.life * 9), 1);
          rect(-b.size / 2, -b.size / 3, b.size, b.size * 0.66, b.col); pop(); break;
        case 'puff': {
          const t = b.life / b.max;
          ctx.globalAlpha = t * 0.85;
          circ(b.x, b.y, b.size, b.col, 0);
          ctx.globalAlpha = t * 0.5; circ(b.x - b.size * 0.3, b.y - b.size * 0.3, b.size * 0.4, WHITE, 0);
          break;
        }
        case 'sparkle': {
          const s = b.size * Math.sin(Math.PI * b.life / b.max);
          star(b.x, b.y, s, b.rot, 4, b.col, 0); star(b.x, b.y, s * 0.45, b.rot, 4, WHITE, 0); break;
        }
        case 'heart': {
          const s = b.size * Math.min(1, (b.max - b.life) * 6);
          push(b.x, b.y, Math.sin(b.life * 6) * 12); heart(0, 0, s, b.col, 2.5); pop(); break;
        }
        case 'note':
          push(b.x, b.y, Math.sin(b.life * 6) * 15);
          line(b.size * 0.8, 0, b.size * 0.8, -b.size * 3, 3, b.col, 2);
          line(b.size * 0.8, -b.size * 3, b.size * 2, -b.size * 2.3, 3, b.col, 2);
          ell(0, 0, b.size, b.size * 0.75, b.col, 2.5, -0.3);
          pop(); break;
        case 'bubble':
          ctx.globalAlpha = a * 0.55; circ(b.x, b.y, b.size, b.col, 2);
          ctx.globalAlpha = a; ell(b.x - b.size * 0.35, b.y - b.size * 0.4, b.size * 0.3, b.size * 0.18, WHITE, 0); break;
        case 'crumb': circ(b.x, b.y, b.size, b.col, 0); break;
        case 'dust': ctx.globalAlpha = a * 0.9; circ(b.x, b.y, b.size, b.col, 0); break;
        case 'z': ctx.globalAlpha = Math.min(1, b.life); word('z', b.x, b.y, b.size * (1.2 - b.life / b.max * 0.5), -10, b.col); break;
      }
      ctx.globalAlpha = 1;
    }
  }
  drawWords() {
    for (const w of this.words) {
      const age = w.max - w.life;
      const s = springOut(age * 1.4) * clamp(w.life / 0.25, 0, 1);
      word(w.text, w.x, Math.max(w.y - age * 18, w.size * 0.6 + 120), w.size * s, w.rot + Math.sin(age * 9) * 4, w.col);
    }
  }
}
