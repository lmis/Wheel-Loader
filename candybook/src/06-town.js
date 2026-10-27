// ---------- page: the street at sunset ----------
class TownPage {
  constructor() {
    this.name = 'town';
    this.text = 'Goodbye, robots! Goodbye, factory! Little Dino walks down the street as the sun goes down and the lamps begin to glow.';
    this.st = new Stage(); this.st.floor = 700;
    this.acts = new Acts({ lamps: 2, bus: 3.4, bakery: 4.5, birds: 3, cat: 2.4, dog: 2.4 });
    this.houses = [
      { x: 230, w: 170, h: 200, col: '#8cb4e6', roof: '#5a6eaa' },
      { x: 420, w: 220, h: 190, col: '#ffe08a', roof: '#ff8ca0', bakery: true },
      { x: 660, w: 160, h: 240, col: '#a8dc96', roof: '#5a9a6e' },
      { x: 840, w: 170, h: 210, col: '#f5a08c', roof: '#b4505a' },
    ];
    this.windows = [];
    for (const h of this.houses) if (!h.bakery)
      for (let r = 0; r < 2; r++) for (let c = 0; c < 2; c++)
        this.windows.push({ x: h.x + 26 + c * (h.w - 92), y: 470 - h.h + 50 + r * 70, on: false, at: rnd(3, 30), house: h });
    this.lamps = [300, 620, 940].map(x => ({ x, on: 0 }));
    this.birds = [0.18, 0.3, 0.44, 0.58, 0.72].map((u, i) => ({ u, ph: rnd(0, 6), dir: i % 2 ? 1 : -1 }));
    this.cat = { startle: 0, blink: 2 };
    this.robotWave = 0;
    this.clouds = Array.from({ length: 5 }, () => ({ x: rnd(0, W), y: rnd(130, 260), s: rnd(0.8, 1.4), v: rnd(4, 9) }));
    this.dino = new Dino(this.st, {
      x: 200, ground: () => 664, minX: 90, maxX: 1080, tricks: ['wave', 'jump', 'roar', 'spin'],
      host: {
        waveTo: 100, waveWord: 'BYE-BYE!', waved: () => { this.robotWave = 2.5; },
        roarAt: 1150, startle: () => { this.cat.startle = 1; sfx('meow', 0.8, 1.3); if (!this.acts.on('birds')) this.flyBirds(); },
        cheer: () => this.st.hearts(1150, 330, 2),
      },
    });
  }
  wireY(x) { const u = (x - 210) / 850; return 200 + 60 * 4 * u * (1 - u); }
  flyBirds() { this.acts.run('birds'); sfx('flutter', 1); sfx('tweet', 0.6, 1.1); }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    if (k === 0) {
      const a = cycle(st, 0, ['lamps', 'bus', 'bakery']);
      this.acts.run(a);
      if (a === 'bus') { sfx('engine', 1); }
      if (a === 'bakery') { sfx('creak', 1); }
    } else {
      const a = cycle(st, 1, ['birds', 'cat', 'dog']);
      if (a === 'birds') this.flyBirds(); else this.acts.run(a);
      if (a === 'cat') { sfx('meow', 1); st.say('MEOW!', 1150, 300, 54, '#ffb45a'); }
      if (a === 'dog') { this.windows[2].on = true; }
    }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    for (const c of this.clouds) { c.x += c.v * dt; if (c.x > W + 100) c.x = -100; }
    for (const w of this.windows) if (!w.on && t > w.at) w.on = true;
    this.cat.startle = Math.max(0, this.cat.startle - dt);
    this.cat.blink -= dt; if (this.cat.blink < -0.15) this.cat.blink = rnd(2, 5);
    this.robotWave = Math.max(0, this.robotWave - dt);
    // the lamps come on by themselves after a while, or all at once with Space
    this.lamps.forEach((l, i) => {
      const lit = (A.on('lamps') && A.t('lamps') > i * 0.35) || t > 16 + i * 1.2;
      if (lit && !l.on) { l.on = 0.001; sfx('lampOn', 0.8, 1 + i * 0.25); st.sparkles(l.x + 30, 410, 6, 30, '#fff3a0'); if (A.on('lamps')) this.windows.forEach(w => w.on = true); }
      if (l.on) l.on = Math.min(1, l.on + dt * 3);
    });
    if (A.on('bus')) {
      if (A.at('bus', 0.7, dt) || A.at('bus', 2.0, dt)) { sfx('beep', 1); st.say('BEEP BEEP!', -320 + A.t('bus') * 560 + 150, 340, 50, '#ffc83c'); }
      if (Math.floor(A.t('bus') * 6) !== Math.floor((A.t('bus') - dt) * 6)) st.puff(-320 + A.t('bus') * 560 - 10, 530, 10, '#d2c8dc', 0);
    }
    if (A.on('bakery')) {
      const b = A.t('bakery');
      if (A.at('bakery', 0.6, dt)) { st.say('MMM, PIE!', 530, 230, 50, '#ffd25a'); sfx('chime', 0.7); }
      if (b > 0.6 && b < 4 && Math.floor(b * 4) !== Math.floor((b - dt) * 4)) st.puff(530 + rnd(-20, 20), 300, 8, '#ffffff', 0, { vy: -40 });
    }
    if (A.at('dog', 0.3, dt) || A.at('dog', 1.0, dt)) { sfx('woof', 1); }
    if (A.at('dog', 0.3, dt)) st.say('WOOF!', 300, 260, 54, '#c88c50');
    if (A.at('birds', 0.1, dt)) st.say('TWEET!', 640, 160, 46, '#82aaff');
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, 470, grad(0, 0, 0, 470, ['#7d74c8', '#c88cc0', '#ff9fa8', '#ffc77a']));
    const sunY = 330 + Math.min(70, t * 1.4);
    glow(980, sunY, 220, '#ffd27a', 0.6);
    circ(980, sunY, 76, '#ffcf5a', 0);
    circ(980, sunY, 60, '#ffe08a', 0);
    for (const c of this.clouds) cloud(c.x, c.y, c.s, '#ffd6dc', 0.9);
    // far city
    ctx.fillStyle = 'rgba(150,110,180,0.55)';
    for (let x = -20, i = 0; x < W; x += 70, i++) { const h = 60 + (i * 37 % 70); ctx.fillRect(x, 470 - h, 58, h); }
    // telephone wire and poles
    for (const px of [210, 1060]) line(px, 470, px, 186, 8, '#7a5a5a');
    ctx.beginPath(); ctx.moveTo(210, 200); ctx.quadraticCurveTo(635, 320, 1060, 200); ctx.lineWidth = 2; ctx.strokeStyle = INK; ctx.stroke();
    this.drawFactory(t);
    for (const h of this.houses) this.drawHouse(h, t);
    this.drawWall(t);
    // birds on the wire
    this.birds.forEach((b, i) => {
      const x = 210 + b.u * 850, y = this.wireY(x) - 9;
      let dx = 0, dy = 0, flap = null;
      if (A.on('birds')) {
        const k = clamp(A.t('birds') / 2.8, 0, 1);
        dx = Math.sin(k * TAU) * 200 * b.dir; dy = -Math.sin(k * Math.PI) * 140 - Math.sin(k * TAU * 2) * 20;
        flap = k < 1 ? A.t('birds') * 25 + i : null;
      }
      drawBird(x + dx, y + dy + (flap === null ? Math.abs(Math.sin(t * 2 + b.ph)) * -3 : 0), 0.8, flap, '#6482d2', dx >= 0 ? 1 : -1);
    });
    // road
    rect(0, 470, W, 92, '#6d6a7d');
    ctx.fillStyle = '#f5e6b4';
    for (let x = 20; x < W; x += 110) ctx.fillRect(x, 512, 60, 7);
    if (A.on('bus')) this.drawBus(-320 + A.t('bus') * 560, t);
    rect(0, 560, W, 14, '#bfb3a8'); rect(0, 572, W, 3, 'rgba(70,40,47,0.4)');
    rect(0, 575, W, H - 575, grad(0, 575, 0, H, ['#e2d6c8', '#cdbfb0']));
    ctx.fillStyle = 'rgba(70,40,47,0.18)';
    for (let x = 0; x < W; x += 120) ctx.fillRect(x, 575, 3, H);
    ctx.fillRect(0, 650, W, 3);
    for (const l of this.lamps) this.drawLamp(l, t);
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    st.drawWords();
  }
  drawFactory(t) {
    rect(0, 240, 200, 230, '#ff9bb4'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(-3, 240, 203, 230);
    for (let i = 0; i < 4; i++) poly([i * 50, 240, i * 50 + 50, 240, i * 50 + 50, 200], '#ff6e96');
    rrect(30, 150, 30, 92, 4, '#aa6ec8');
    if (Math.floor(t * 0.6) !== Math.floor((t - 1 / 60) * 0.6)) this.st.puff(45, 150, 12, '#fff0f5', 0);
    line(160, 240, 160, 170, 6, WHITE); circ(160, 160, 24, '#ff5c78');
    ctx.beginPath(); for (let a = 0; a < 9; a += 0.3) ctx.lineTo(160 + Math.cos(a) * a * 2.2, 160 + Math.sin(a) * a * 2.2); ctx.lineWidth = 4; ctx.strokeStyle = '#ffc0cc'; ctx.stroke();
    rrect(20, 290, 40, 40, 6, '#ffe8a0'); rrect(90, 290, 40, 40, 6, '#ffe8a0');
    rrect(100, 380, 80, 90, 10, '#5a3c46');
    glow(140, 420, 60, '#ffd27a', 0.35);
    // the robots wave from the door
    const k = this.robotWave > 0 ? 18 : 6;
    [['#5acdc3', 116], ['#ffa555', 140], ['#b9a0f5', 164]].forEach(([c, x], i) => {
      const y = 430 + Math.sin(t * 3 + i) * 2;
      line(x + 8, y + 6, x + 16 + Math.sin(t * k + i) * 5, y - 16, 3, '#aaaabe', 2);
      rrect(x - 9, y - 9, 18, 18, 5, c, 2.5);
      circ(x - 3, y - 2, 2, INK, 0); circ(x + 3, y - 2, 2, INK, 0);
    });
  }
  drawHouse(h, t) {
    const top = 470 - h.h;
    rect(h.x, top, h.w, h.h, h.col); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(h.x, top, h.w, h.h);
    if (h.bakery) {
      // awning, sign, a display window full of cakes, and the window upstairs
      for (let i = 0; i < 8; i++) poly([h.x - 8 + i * 29.5, top + 80, h.x - 8 + (i + 1) * 29.5, top + 80, h.x - 8 + (i + 1) * 29.5, top + 104, h.x - 8 + i * 29.5, top + 104], i % 2 ? WHITE : '#ff8ca0', 2);
      rrect(h.x + 20, top + 116, h.w - 40, 60, 6, '#fff3c8');
      glow(h.x + h.w / 2, top + 146, 90, '#ffd27a', 0.4);
      for (let i = 0; i < 4; i++) { const cx = h.x + 44 + i * 44; rrect(cx - 12, top + 150, 24, 18, 4, CANDY[i]); circ(cx, top + 146, 8, '#fff', 2); circ(cx, top + 139, 3, '#ff4060', 0); }
      rrect(h.x + h.w / 2 - 60, top + 14, 120, 40, 12, '#fff3c8');
      word('BAKERY', h.x + h.w / 2, top + 35, 26, 0, '#ff8ca0');
      // upstairs window with shutters, and a pie when they open
      const wx = h.x + h.w / 2 - 50, wy = top + 58 - 30;
      const open = this.acts.on('bakery') ? ease(this.acts.t('bakery') / 0.5) * (1 - ease((this.acts.t('bakery') - 4) / 0.5)) : 0;
      if (open > 0.05) {
        rrect(wx + 10, wy + 2, 80, 20, 4, '#ffe9a8', 0);
        ell(wx + 50, wy + 18, 34, 10, '#d68a46'); ell(wx + 50, wy + 14, 28, 7, '#f0b464', 2);
      }
      rrect(wx, wy, 50 * (1 - open) + 2, 24, 3, '#e8708a'); rrect(wx + 100 - 50 * (1 - open) - 2, wy, 50 * (1 - open) + 2, 24, 3, '#e8708a');
      return;
    }
    poly([h.x - 14, top, h.x + h.w + 14, top, h.x + h.w / 2, top - 70], h.roof);
    for (const w of this.windows) if (w.house === h) {
      rrect(w.x, w.y, 40, 46, 6, w.on ? '#ffe28a' : '#5a5a8a');
      if (w.on) glow(w.x + 20, w.y + 23, 40, '#ffd27a', 0.3);
      rect(w.x + 18, w.y, 4, 46, INK); rect(w.x, w.y + 21, 40, 4, INK);
    }
    // the dog in the blue house
    if (h === this.houses[0] && this.acts.on('dog')) {
      const w = this.windows[2], d = this.acts.t('dog'), up = ease(d / 0.25) * (1 - ease((d - 2) / 0.3));
      const bob = Math.abs(Math.sin(d * 9)) * 4;
      ctx.save(); ctx.beginPath(); ctx.rect(w.x, w.y, 40, 46); ctx.clip();
      const y = w.y + 46 - up * 30 - bob;
      circ(w.x + 20, y + 4, 15, '#c88c50');
      ell(w.x + 7, y + 6, 5, 11, '#8c5a32', 2); ell(w.x + 33, y + 6, 5, 11, '#8c5a32', 2);
      circ(w.x + 15, y, 2.2, INK, 0); circ(w.x + 25, y, 2.2, INK, 0); circ(w.x + 20, y + 7, 3, INK, 0);
      ctx.restore();
    }
  }
  drawWall(t) {
    const c = this.cat, A = this.acts;
    circ(1200, 300, 70, '#5aa064'); circ(1150, 330, 50, '#64aa6e'); circ(1250, 340, 54, '#64aa6e'); line(1200, 470, 1200, 360, 18, '#8c5a3c');
    rect(1030, 400, 250, 70, '#c87864'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(1030, 400, 253, 70);
    ctx.fillStyle = 'rgba(70,40,47,0.3)';
    for (let r = 0; r < 3; r++) for (let x = 1030 + (r % 2) * 25; x < W; x += 50) ctx.fillRect(x, 400 + r * 23, 3, 23);
    for (let r = 1; r < 3; r++) ctx.fillRect(1030, 400 + r * 23, 250, 3);
    // the cat
    const st = c.startle, s = A.on('cat') ? A.t('cat') : -1;
    const stretch = s >= 0 ? Math.sin(Math.PI * clamp(s / 2, 0, 1)) : 0;
    const x = 1150, y = 400 - st * 30;
    const sw = Math.sin(t * 2) * 12;
    curve(x - 26, y - 10, x - 50, y + 10, x - 44 + sw, y + 60, 8, '#f0a050');
    push(x, y, 0, 1 + stretch * 0.35 + st * 0.2, 1 - stretch * 0.15 + st * 0.25);
    ell(0, -22, 30, 24, '#f0a050');
    for (let i = 0; i < 3; i++) line(-12 + i * 10, -42, -16 + i * 10, -30, 3, '#c8783c', 0);
    pop();
    const hx = x + 22 + stretch * 16, hy = y - 52 + stretch * 10 - st * 10;
    poly([hx - 16, hy - 8, hx - 12, hy - 30, hx - 2, hy - 14], '#f0a050');
    poly([hx + 16, hy - 8, hx + 12, hy - 30, hx + 2, hy - 14], '#f0a050');
    circ(hx, hy, 18, '#f0a050');
    if (c.blink < 0 || stretch > 0.3) { line(hx - 10, hy - 2, hx - 3, hy - 2, 2.5, INK, 0); line(hx + 3, hy - 2, hx + 10, hy - 2, 2.5, INK, 0); }
    else { circ(hx - 7, hy - 3, 4 + st * 2, '#9be05a', 2); circ(hx + 7, hy - 3, 4 + st * 2, '#9be05a', 2); circ(hx - 7, hy - 3, 1.6, INK, 0); circ(hx + 7, hy - 3, 1.6, INK, 0); }
    poly([hx - 3, hy + 4, hx + 3, hy + 4, hx, hy + 8], '#ff8ca0', 0);
    if (stretch > 0.3 || st > 0) ell(hx, hy + 11, 4, 3 + stretch * 3, '#8c3246', 2);
    for (const s2 of [-1, 1]) { line(hx + s2 * 6, hy + 6, hx + s2 * 24, hy + 3, 1.2, INK, 0); line(hx + s2 * 6, hy + 8, hx + s2 * 24, hy + 10, 1.2, INK, 0); }
  }
  drawLamp(l, t) {
    const x = l.x;
    if (l.on) {
      glow(x + 30, 418, 150, '#ffe08a', 0.45 * l.on);
      ctx.globalAlpha = 0.18 * l.on; poly([x + 14, 425, x + 46, 425, x + 120, 700, x - 60, 700], '#ffe9a8', 0); ctx.globalAlpha = 1;
      for (let i = 0; i < 3; i++) { const a = t * 3 + i * 2.1; circ(x + 30 + Math.cos(a) * 26, 414 + Math.sin(a * 1.3) * 16, 2, '#fff8c8', 0); }
    }
    line(x, 600, x, 410, 8, '#4a4a6a');
    curve(x, 410, x, 392, x + 24, 394, 6, '#4a4a6a');
    poly([x + 12, 398, x + 48, 398, x + 42, 418, x + 18, 418], '#4a4a6a');
    circ(x + 30, 422, 8, l.on ? '#fff6be' : '#9a9ab4');
    ell(x, 602, 14, 5, '#4a4a6a');
  }
  drawBus(x, t) {
    const y = 395;
    rrect(x, y, 300, 140, 26, '#ffb43c');
    rect(x + 6, y + 100, 288, 10, '#ff8c3c');
    for (let i = 0; i < 4; i++) {
      const wx = x + 20 + i * 64;
      rrect(wx, y + 22, 50, 50, 8, '#cdeeff');
      const c = CANDY[(i * 2) % 7];
      rrect(wx + 15, y + 44 + Math.sin(t * 8 + i) * 2, 20, 20, 5, c, 2);
      circ(wx + 21, y + 52 + Math.sin(t * 8 + i) * 2, 2, INK, 0); circ(wx + 29, y + 52 + Math.sin(t * 8 + i) * 2, 2, INK, 0);
    }
    glow(x + 300, y + 110, 50, '#fff3a0', 0.6); circ(x + 292, y + 110, 9, '#fff6be');
    for (const wx of [x + 64, x + 236]) { circ(wx, y + 140, 22, '#3c3c50'); circ(wx, y + 140, 9, '#c8c8d7'); line(wx, y + 140, wx + Math.cos(t * 10) * 9, y + 140 + Math.sin(t * 10) * 9, 2, INK, 0); }
  }
}
