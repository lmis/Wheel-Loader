// ---------- page: the bridge over the river at dusk ----------
const BR = { x0: 250, x1: 1030, top: 430, bank: 566 };
function riverGround(x) {
  if (x < BR.x0 + 40 || x > BR.x1 - 40) return BR.bank;
  const u = (x - BR.x0 - 40) / (BR.x1 - BR.x0 - 80);
  return BR.bank - (BR.bank - BR.top) * Math.sin(Math.PI * u);
}
class RiverPage {
  constructor() {
    this.name = 'river';
    this.text = 'Over the little wooden bridge, the ducks are tucking in for the night. “Goodnight, ducks,” says Little Dino.';
    this.st = new Stage(); this.st.floor = 720;
    this.acts = new Acts({ fish: 2.2, flies: 4.5, boat: 9, ducks: 1.6, frog: 1.2, turtle: 3 });
    this.stars = makeStars(26, 0, W, 110, 260);
    this.flies = makeFlies(16, 40, W - 40, 300, 640);
    this.ducks = { x: 520, dir: 1, startle: 0 };
    this.frogPad = 0; this.fishX = 600;
    this.ambFish = 6;
    this.dino = new Dino(this.st, {
      x: 140, ground: riverGround, minX: 70, maxX: W - 90, tricks: ['wave', 'jump', 'spin', 'roar'],
      host: {
        waveWord: 'GOODNIGHT, DUCKS!', waveTo: 600, waved: () => { this.acts.run('ducks'); sfx('quack', 0.7, 1.1); },
        roarAt: 600, startle: () => { this.ducks.startle = 1; sfx('quack', 1, 1.2); this.st.say('QUACK!', this.ducks.x, 560, 46, '#ffd25a'); },
        cheer: () => this.st.hearts(this.ducks.x, 590, 2),
      },
    });
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    const a = k === 0 ? cycle(st, 0, ['fish', 'flies', 'boat']) : cycle(st, 1, ['ducks', 'frog', 'turtle']);
    this.acts.run(a);
    if (a === 'fish') { this.fishX = rnd(380, 760); sfx('splash', 0.7); }
    if (a === 'flies') { sfx('shimmer', 0.8); }
    if (a === 'boat') { sfx('bloop', 0.7, 0.7); }
    if (a === 'ducks') { sfx('quack', 1); st.say('QUACK QUACK!', this.ducks.x, 560, 46, '#ffd25a'); }
    if (a === 'frog') { sfx('ribbit', 1); st.say('RIBBIT!', 420, 600, 46, '#8ce678'); }
    if (a === 'turtle') { sfx('blub', 1); }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const d = this.ducks;
    d.x += d.dir * 18 * dt * (1 + d.startle * 4);
    if (d.x > 860) d.dir = -1; if (d.x < 380) d.dir = 1;
    d.startle = Math.max(0, d.startle - dt * 0.8);
    // a fish jumps now and then by itself
    if ((this.ambFish -= dt) <= 0 && !A.on('fish')) { this.ambFish = rnd(9, 16); this.fishX = rnd(380, 760); A.run('fish'); sfx('splash', 0.35); }
    if (A.on('fish')) {
      const f = A.t('fish');
      if (A.at('fish', 0.05, dt) || A.at('fish', 1.25, dt)) {
        const x = this.fishX + (f > 1 ? 180 : 0);
        for (let i = 0; i < 12; i++) st.add({ kind: 'crumb', layer: 2, x, y: 640, vx: rnd(-120, 120), vy: rnd(-300, -120), size: rnd(3, 6), col: '#bfe6ff', life: 0.8, g: 700 });
        if (f > 1) { sfx('splash', 0.8, 1.2); st.say('SPLASH!', x, 520, 54, '#8cd2ff'); }
      }
    }
    if (A.at('frog', 0.5, dt)) this.frogPad ^= 1;
    if (A.on('turtle') && Math.floor(A.t('turtle') * 3) !== Math.floor((A.t('turtle') - dt) * 3)) st.bubbles(820, 650, 2);
    if (A.at('turtle', 0.6, dt)) st.say('BLUB BLUB!', 820, 580, 44, '#8ce6b4');
    const t = st.t;
    updateFlies(this.flies, t, dt, A.on('flies') && A.t('flies') < 3.6 ? (f => {
      const i = this.flies.indexOf(f), a = i / this.flies.length * TAU + t * 0.6;
      const hx = 16 * Math.pow(Math.sin(a), 3), hy = -(13 * Math.cos(a) - 5 * Math.cos(2 * a) - 2 * Math.cos(3 * a) - Math.cos(4 * a));
      return [640 + hx * 7, 330 + hy * 7];
    }) : null);
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, 440, grad(0, 0, 0, 440, ['#2f2f6e', '#6a5aa8', '#c48cb4', '#f5b09c']));
    drawStars(this.stars, t, 0.9);
    glow(900, 420, 200, '#ffb48c', 0.35);
    // far bank: trees
    ctx.fillStyle = '#4a4f86';
    for (let x = -30, i = 0; x < W + 40; x += 64, i++) { const r = 40 + (i * 23 % 30); ctx.beginPath(); ctx.arc(x, 420 - r * 0.4, r, 0, TAU); ctx.fill(); }
    rect(0, 410, W, 30, '#4a4f86');
    // water
    rect(0, 436, W, H - 436, grad(0, 436, 0, H, ['#7a6fb8', '#4f5aa0', '#3a4486']));
    ctx.globalAlpha = 0.35;
    for (let i = 0; i < 26; i++) {
      const y = 450 + (i * 37) % 260, x = (i * 173 + t * (14 + i % 5 * 4)) % (W + 200) - 100;
      line(x, y, x + 40 + (i % 3) * 20, y, 3, '#e6d2ff', 0);
    }
    ctx.globalAlpha = 1;
    glow(900, 470, 140, '#ffb48c', 0.18);
    // ducks (behind the bridge's front, in the water beyond)
    this.drawDucks(t);
    // the bridge
    this.drawBridge(t);
    // things in the water in front
    if (A.on('boat')) this.drawBoat(-80 + A.t('boat') * 160, 640, t);
    if (A.on('turtle')) {
      const u = A.t('turtle'), up = ease(u / 0.4) * (1 - ease((u - 2.4) / 0.5));
      ctx.save(); ctx.beginPath(); ctx.rect(700, 560, 240, 96); ctx.clip();
      ell(820, 676 - up * 26, 34, 22, '#5aa064'); for (let i = 0; i < 3; i++) circ(804 + i * 16, 664 - up * 26, 5, '#468c50', 0);
      circ(860, 662 - up * 30, 11, '#8cd28c'); circ(864, 658 - up * 30, 2.2, INK, 0);
      ctx.restore();
    }
    this.drawFish(t);
    // lily pads and the frog
    for (const [x, y] of [[400, 690], [500, 700]]) { ell(x, y, 40, 12, '#5ab46e'); line(x, y, x + 30, y - 6, 3, '#3c8c50', 0); }
    {
      const pads = [[400, 684], [500, 694]], f = A.on('frog') ? A.t('frog') : -1;
      let fx = pads[this.frogPad][0], fy = pads[this.frogPad][1];
      if (f >= 0 && f < 0.5) { const p = f / 0.5, from = pads[this.frogPad], to = pads[this.frogPad ^ 1]; fx = lerp(from[0], to[0], p); fy = lerp(from[1], to[1], p) - Math.sin(p * Math.PI) * 60; }
      const puff = f >= 0 && f < 1.2 ? Math.abs(Math.sin(f * 14)) * 4 : 0;
      ell(fx, fy - 10, 18, 12, '#8ce678'); ell(fx, fy - 4 + puff * 0.5, 10, 4 + puff, '#e6f5a0', 2);
      circ(fx - 8, fy - 22, 6, '#8ce678'); circ(fx + 8, fy - 22, 6, '#8ce678');
      circ(fx - 8, fy - 23, 2.4, INK, 0); circ(fx + 8, fy - 23, 2.4, INK, 0);
    }
    // the banks with reeds
    for (const [edge, inner, bulge] of [[0, 300, 360], [W, 980, 920]]) {
      ctx.beginPath(); ctx.moveTo(edge, H); ctx.lineTo(edge, 566); ctx.lineTo(inner, 566); ctx.quadraticCurveTo(bulge, 640, inner, H);
      ctx.closePath(); paint('#5f9a6e', 3);
      rect(Math.min(edge, inner), 566, Math.abs(inner - edge), 6, '#79b67e');
    }
    for (const x of [30, 70, 250, 280, 1010, 1060, 1230]) {
      const sw = Math.sin(t * 1.4 + x) * 6;
      curve(x, 600, x + sw * 0.5, 540, x + sw, 480, 4, '#4f8a5c', 0);
      ell(x + sw, 492, 5, 16, '#9a6046', 2);
    }
    grassTufts([120, 180, 1100, 1160], 568, t, '#4f8a5c');
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    drawFlies(this.flies, t, A.on('flies') ? 0.5 : 0);
    st.drawWords();
  }
  drawBridge(t) {
    // arch underneath and its reflection
    ctx.beginPath(); ctx.moveTo(BR.x0 + 60, 566); ctx.quadraticCurveTo(640, 360, BR.x1 - 60, 566);
    ctx.lineTo(BR.x1 - 120, 566); ctx.quadraticCurveTo(640, 450, BR.x0 + 120, 566); ctx.closePath(); paint('#b48250', 3);
    ctx.globalAlpha = 0.25; ctx.beginPath(); ctx.moveTo(BR.x0 + 120, 570); ctx.quadraticCurveTo(640, 680, BR.x1 - 120, 570); ctx.lineWidth = 16; ctx.strokeStyle = '#b48250'; ctx.stroke(); ctx.globalAlpha = 1;
    // deck
    ctx.beginPath();
    for (let x = BR.x0; x <= BR.x1; x += 10) ctx.lineTo(x, riverGround(x) + 2);
    for (let x = BR.x1; x >= BR.x0; x -= 10) ctx.lineTo(x, riverGround(x) + 20);
    ctx.closePath(); paint('#c8965a', 3);
    // railing posts and the rail
    for (let x = BR.x0 + 40; x <= BR.x1 - 40; x += 60) line(x, riverGround(x) + 2, x, riverGround(x) - 44, 7, '#a0703c');
    ctx.beginPath();
    for (let x = BR.x0 + 40; x <= BR.x1 - 40; x += 10) ctx.lineTo(x, riverGround(x) - 44);
    ctx.lineWidth = 12; ctx.strokeStyle = INK; ctx.stroke(); ctx.lineWidth = 7; ctx.strokeStyle = '#c8965a'; ctx.stroke();
    // a little lantern in the middle
    glow(640, 360, 70, '#ffd27a', 0.5 + 0.1 * Math.sin(t * 3));
    rrect(630, 350, 20, 26, 4, '#ffe08a'); line(640, 386, 640, 376, 3, INK, 0);
  }
  drawDucks(t) {
    const d = this.ducks, A = this.acts;
    const q = A.on('ducks') ? A.t('ducks') : -1;
    [[0, 1.1], [-46, 0.6], [-80, 0.6], [-114, 0.6]].forEach(([off, s], i) => {
      const x = d.x + off * d.dir, y = 600 + Math.sin(t * 2 + i) * 2 - (q >= 0 ? Math.abs(Math.sin(q * 10 + i)) * 8 : 0);
      ctx.globalAlpha = 0.3; ell(x, y + 14 * s, 26 * s, 5 * s, '#e6d2ff', 0); ctx.globalAlpha = 1;
      push(x, y, 0, s * d.dir, s);
      if (d.startle > 0) { const f = Math.sin(t * 30) * 10; poly([-6, -6, -30, -24 - f, -26, 0], '#e6e0c8'); }
      ell(0, 0, 26, 14, i ? '#ffe678' : '#f0ead2');
      circ(18, -16, 12, i ? '#ffe678' : '#3c8c64');
      const open = q >= 0 && Math.sin(q * 20) > 0 ? 4 : 0;
      poly([28, -18, 40, -16 - open, 28, -12], '#ff9a3c', 2); if (open) poly([28, -14, 38, -10 + open, 28, -11], '#ff9a3c', 2);
      circ(21, -19, 2.4, INK, 0);
      pop();
    });
  }
  drawFish(t) {
    const A = this.acts;
    if (!A.on('fish')) return;
    const f = A.t('fish'), p = clamp((f - 0.05) / 1.2, 0, 1);
    if (p <= 0 || p >= 1) return;
    const x = this.fishX + p * 180, y = 640 - Math.sin(p * Math.PI) * 200, rot = (p - 0.5) * 140;
    push(x, y, rot);
    poly([-30, 0, -48, -14, -48, 14], '#ff9a6e');
    ell(0, 0, 32, 16, '#ffb47a');
    ell(4, 4, 18, 8, '#ffe0b4', 0);
    poly([-4, -14, 10, -26, 14, -12], '#ff9a6e', 2.5);
    circ(18, -4, 4, WHITE, 2); circ(19, -4, 2, INK, 0);
    pop();
    if (Math.random() < 0.3) this.st.sparkles(x, y, 1, 10, '#bfe6ff');
  }
  drawBoat(x, y, t) {
    push(x, y + Math.sin(t * 2) * 3, Math.sin(t * 1.6) * 4);
    poly([-44, -10, 44, -10, 30, 10, -30, 10], '#f5f0e6');
    poly([-10, -10, 0, -50, 12, -10], '#fff8ec', 2.5);
    drawCandy(22, -22, 0.3, 11, '#ff5c78', 0);
    pop();
  }
}
