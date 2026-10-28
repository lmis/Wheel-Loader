// ---------- page: into the quiet woods ----------
class WoodsPage {
  constructor() {
    this.name = 'woods';
    this.text = 'Into the quiet woods. The mushrooms glow softly, and the woodland friends are curling up for the night.';
    this.tricks = [['mushrooms', 'moonbeam', 'web'], ['hedgehog', 'fox', 'deer']]; this.who = ['Woods', 'Woodland friends', 'Little Dino'];
    this.st = new Stage(); this.st.floor = 690;
    this.acts = new Acts({ mushrooms: 3.6, moonbeam: 4.5, web: 4, hedgehog: 4, fox: 3.4, deer: 4.6 });
    this.flies = makeFlies(18, 40, W - 40, 220, 640);
    // mushrooms: each one sings a note of the scale when it lights up
    const notes = [72, 74, 76, 79, 81, 84, 86, 88];
    this.shrooms = [[190, 650, 1.1], [228, 656, 0.7], [262, 648, 0.9], [300, 658, 0.6], [930, 650, 1], [968, 658, 0.65], [1004, 646, 0.85], [1046, 656, 0.6]]
      .map(([x, y, s], i) => ({ x, y, s, note: notes[i], col: i % 2 ? '#ff9ad2' : '#8ce6ff', lit: 0 }));
    this.foxBlink = 0; this.hx = 420; this.hdir = 1;
    this.dino = new Dino(this.st, {
      x: 120, ground: () => 668, minX: 80, maxX: W - 110, tricks: ['tiptoe', 'wave', 'jump', 'spin'],
      host: {
        waveWord: 'GOODNIGHT, FOX!', waveTo: 1120, waved: () => this.acts.run('fox'),
        trick: (d, t, dt) => this.tiptoe(d, t, dt), startle: () => this.acts.run('hedgehog'), roarAt: 420,
      },
    });
  }
  // a quiet trick for a quiet wood: little slow steps, and SHHH!
  tiptoe(d, t, dt) {
    if (d.stage === 0) { d.stage = 1; this.st.say('SHHH!', d.x + d.face * 40, d.y - 200, 56, '#c8e6ff'); sfx('shh', 1); }
    if (d.stepHop(dt) || !d.hopping) {
      if (t > 3) return true;
      const x = clamp(d.x + d.face * 34, d.minX, d.maxX);
      if (x === d.x) d.face = -d.face;
      d.hop(x, 0, 0.5, 10); sfx('tick', 0.15, 0.6);
    }
    return false;
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    const a = cycle(st, k, this.tricks[k]);
    this.acts.run(a);
    if (a === 'moonbeam') sfx('shimmer', 0.6, 0.8);
    if (a === 'web') sfx('sparkle', 0.8, 1.2);
    if (a === 'hedgehog') { sfx('squeak', 0.5, 0.8); }
    if (a === 'fox') { sfx('yawn', 0.6, 1.5); st.say('YAAAWN!', 1120, 520, 42, '#ffb46e'); }
    if (a === 'deer') sfx('leaves', 0.8);
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    this.shrooms.forEach((m, i) => {
      m.lit = Math.max(0, m.lit - dt * 0.7);
      if (A.at('mushrooms', i * 0.32, dt)) { m.lit = 1; INST && AU.ctx && AU.ctx.state === 'running' && INST.box(mtof(m.note), AU.ctx.currentTime + 0.01, 0.4, 2.2, AU.fx); st.sparkles(m.x, m.y - 30 * m.s, 3, 20, m.col); }
    });
    if (A.at('hedgehog', 3.2, dt)) st.say('PEEP!', this.hx + this.hdir * 340, 560, 40, '#e6c8a0');
    if (A.at('hedgehog', 3.95, dt)) { this.hx += this.hdir * 340; this.hdir = -this.hdir; }
    if (A.at('deer', 1.2, dt)) st.say('HELLO, DEER!', 520, 330, 42, '#ffe0b4');
    if (A.on('moonbeam') && Math.random() < dt * 20) st.sparkles(rnd(600, 1100), rnd(150, 600), 1, 4, '#fff6c8');
    updateFlies(this.flies, t, dt);
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, H, grad(0, 0, 0, H, ['#0f2a3a', '#1d4650', '#2a5a58']));
    // moonbeams through the trees
    const mb = 0.07 + (A.on('moonbeam') ? 0.18 * Math.sin(Math.PI * clamp(A.t('moonbeam') / 4.5, 0, 1)) : 0) + 0.02 * Math.sin(t * 0.8);
    ctx.globalAlpha = mb;
    for (const [x, w] of [[760, 70], [920, 110], [1100, 60]]) poly([x, 0, x + w, 0, x + w - 260, H, x - 200, H], '#fff6c8', 0);
    ctx.globalAlpha = 1;
    // three layers of trunks, darker behind
    const layer = (xs, w, col) => { for (const x of xs) rect(x - w / 2, 0, w, 640, col); };
    layer([60, 330, 580, 860, 1180], 30, '#183a44');
    layer([150, 480, 760, 1050], 46, '#1f4a4c');
    for (const x of [40, 400, 690, 1230]) { rect(x - 34, 0, 68, 660, '#2c5850'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(x - 34, -5, 68, 665); for (let y = 60; y < 640; y += 90) line(x - 14, y, x - 6, y + 40, 3, '#244a44', 0); }
    // leafy canopy at the top
    for (let x = -40, i = 0; x < W + 60; x += 110, i++) circ(x, 40 + (i % 2) * 30, 90, i % 2 ? '#1e4a40' : '#24564a', 0);
    // the spider web between two trunks
    this.drawWeb(t);
    // the forest floor
    ctx.beginPath(); ctx.moveTo(0, 620); ctx.quadraticCurveTo(640, 600, W, 625); ctx.lineTo(W, H); ctx.lineTo(0, H); ctx.closePath(); paint('#2e5a46', 3);
    // ferns
    for (const x of [90, 360, 620, 820, 1150]) for (let i = -2; i <= 2; i++) { const sw = Math.sin(t * 1.2 + x + i) * 4; curve(x, 640, x + i * 18 + sw, 600, x + i * 34 + sw, 590 + Math.abs(i) * 14, 4, '#3f7a54', 0); }
    this.drawDeer(t);
    // log with the fox curled on it
    rrect(1000, 590, 200, 50, 24, '#7a5040'); ell(1000, 615, 14, 25, '#a0704a'); ell(1000, 615, 7, 13, '#7a5040', 0);
    this.drawFox(t);
    for (const m of this.shrooms) this.drawShroom(m, t);
    this.drawHedgehog(t);
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    drawFlies(this.flies, t);
    st.drawWords();
  }
  drawShroom(m, t) {
    const g = 0.35 + 0.15 * Math.sin(t * 1.5 + m.x) + m.lit * 0.6, s = m.s * (1 + m.lit * 0.15);
    glow(m.x, m.y - 30 * s, 70 * s, m.col, g * 0.6);
    rrect(m.x - 7 * s, m.y - 30 * s, 14 * s, 32 * s, 6 * s, '#f5ebdc', 2.5);
    ctx.beginPath(); ctx.ellipse(m.x, m.y - 30 * s, 30 * s, 22 * s, 0, Math.PI, TAU); ctx.closePath(); paint(m.lit > 0.2 ? shade(m.col, 0.3) : m.col, 2.5);
    for (const [dx, dy] of [[-12, -40], [6, -46], [16, -36]]) circ(m.x + dx * s, m.y + dy * s, 3.5 * s, 'rgba(255,255,255,0.8)', 0);
  }
  drawWeb(t) {
    const cx = 220, cy = 260, r = 90, A = this.acts, dew = A.on('web') ? Math.sin(Math.PI * clamp(A.t('web') / 4, 0, 1)) : 0;
    ctx.globalAlpha = 0.5 + dew * 0.4; ctx.strokeStyle = '#e6f0ff'; ctx.lineWidth = 1.2;
    for (let i = 0; i < 10; i++) { const a = i / 10 * TAU; ctx.beginPath(); ctx.moveTo(cx, cy); ctx.lineTo(cx + Math.cos(a) * r, cy + Math.sin(a) * r); ctx.stroke(); }
    for (let k = 1; k <= 5; k++) { ctx.beginPath(); for (let i = 0; i <= 10; i++) { const a = i / 10 * TAU, rr = r * k / 5.5; ctx.lineTo(cx + Math.cos(a) * rr, cy + Math.sin(a) * rr); } ctx.stroke(); }
    ctx.globalAlpha = 1;
    for (let i = 0; i < 14; i++) { const a = i * 2.3, rr = r * (0.3 + (i % 5) / 6); const k = 0.3 + dew * (0.5 + 0.5 * Math.sin(t * 6 + i)); glow(cx + Math.cos(a) * rr, cy + Math.sin(a) * rr, 8 + dew * 8, '#ffffff', k); }
    // the spider on its thread
    const sy = cy + 10 + (A.on('web') ? Math.sin(Math.PI * clamp(A.t('web') / 4, 0, 1)) * 90 : 0) + Math.sin(t * 2) * 3;
    line(cx, cy, cx, sy, 1, '#e6f0ff', 0);
    for (let i = 0; i < 4; i++) for (const s of [-1, 1]) line(cx, sy, cx + s * 12, sy - 6 + i * 4, 1.5, INK, 0);
    circ(cx, sy, 7, '#3c3c4e', 2); circ(cx - 2, sy - 1, 1.5, WHITE, 0); circ(cx + 2, sy - 1, 1.5, WHITE, 0);
  }
  drawHedgehog(t) {
    const A = this.acts, h = A.on('hedgehog') ? A.t('hedgehog') : -1;
    let x = this.hx, rolled = 0, rot = 0;
    if (h >= 0 && h < 3.95) { const k = clamp((h - 0.4) / 2.4, 0, 1); x = this.hx + ease(k) * 340 * this.hdir; rolled = h > 0.3 && h < 3 ? 1 : 0; rot = ease(k) * 340 / 26 * 57.3 * this.hdir; }
    const y = 646;
    if (rolled) {
      push(x, y - 24, rot);
      for (let i = 0; i < 14; i++) { const a = i / 14 * TAU; poly([Math.cos(a) * 18, Math.sin(a) * 18, Math.cos(a + 0.22) * 34, Math.sin(a + 0.22) * 34, Math.cos(a + 0.44) * 18, Math.sin(a + 0.44) * 18], '#8c6450', 2); }
      circ(0, 0, 22, '#a07858');
      pop();
      return;
    }
    const sniff = Math.sin(t * 8) * 1.5;
    for (let i = 0; i < 9; i++) { const a = Math.PI + 0.2 + i * 0.32; poly([x - 6 + Math.cos(a) * 20, y - 12 + Math.sin(a) * 14, x - 6 + Math.cos(a + 0.16) * 38, y - 12 + Math.sin(a + 0.16) * 30, x - 6 + Math.cos(a + 0.32) * 20, y - 12 + Math.sin(a + 0.32) * 14], '#8c6450', 2); }
    ell(x - 4, y - 12, 30, 16, '#a07858');
    ell(x + 22 + sniff, y - 10, 14, 10, '#e6c8a0'); circ(x + 36 + sniff, y - 10, 3.5, INK, 0); circ(x + 22 + sniff, y - 14, 2.2, INK, 0);
  }
  drawFox(t) {
    const A = this.acts, f = A.on('fox') ? A.t('fox') : -1, x = 1110, y = 588;
    const awake = f >= 0 ? Math.sin(Math.PI * clamp(f / 3.4, 0, 1)) : 0, br = Math.sin(t * 1.4) * 1.5;
    // curled body with the bushy tail wrapped round
    ell(x, y - 18 + br * 0.3, 50, 26 + br, '#e6783c');
    ctx.beginPath(); ctx.ellipse(x + 14, y - 8, 46, 16, 0.1, 0, Math.PI); paint('#f08c46', 3);
    ell(x + 50, y - 10, 14, 10, '#fff6ea', 2.5);
    push(x - 40, y - 26 - awake * 22, -awake * 20);
    poly([-14, -10, -18, -30, -4, -16], '#e6783c'); poly([6, -14, 12, -32, 16, -12], '#e6783c');
    circ(0, 0, 18, '#e6783c'); poly([-6, 4, -30, 12, -6, 14], '#fff6ea', 2.5); circ(-30, 11, 3, INK, 0);
    if (awake > 0.4) { circ(-4, -4, 3, INK, 0); circ(8, -4, 3, INK, 0); if (f < 1.3) ell(-14, 12, 5, 6 * awake, '#8c3246', 2); }
    else { curve(-8, -3, -4, 0, 0, -3, 2, INK, 0); curve(5, -3, 9, 0, 13, -3, 2, INK, 0); }
    pop();
    if (f < 0 && Math.random() < 0.003) this.st.zzz(x - 20, y - 70);
  }
  drawDeer(t) {
    const A = this.acts, d = A.on('deer') ? A.t('deer') : -1;
    if (d < 0) return;
    const out = ease(d / 0.8) * (1 - ease((d - 3.8) / 0.8)), x = 680 + out * 90;
    const ear = Math.sin(t * 7) * (d > 1.5 && d < 2.5 ? 12 : 2), nod = Math.sin(d * 2) * 6;
    ctx.save(); ctx.beginPath(); ctx.rect(690 + 34, 0, 400, H); ctx.clip();
    for (const lx of [-30, -10, 30, 48]) line(x + 40 + lx, 560, x + 42 + lx, 636, 9, '#b47850');
    ell(x + 50, 548, 62, 32, '#c88c5a');
    ell(x + 50, 562, 40, 14, '#e6c8a0', 0);
    for (const [sx, sy] of [[34, 534], [58, 528], [76, 540], [46, 546]]) circ(x + sx, sy, 3.5, '#fff6ea', 0);
    line(x + 2, 540, x - 8, 500, 22, '#c88c5a');
    push(x - 12, 480 + nod, nod * 2);
    ell(-24, -18, 18, 8, '#c88c5a', 3, -0.5 - ear * 0.03); ell(20, -20, 18, 8, '#c88c5a', 3, 0.5 + ear * 0.03);
    ell(0, 0, 24, 26, '#c88c5a');
    ell(-6, 16, 14, 11, '#e6c8a0', 2.5); circ(-12, 14, 3.5, INK, 0);
    circ(-8, -4, 4, INK, 0); circ(10, -4, 4, INK, 0); circ(-9, -6, 1.4, WHITE, 0); circ(9, -6, 1.4, WHITE, 0);
    pop();
    ctx.restore();
  }
}
