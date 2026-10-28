// ---------- page: up the hill under the moon ----------
const hillGround = x => 668 - 150 * ease(x / W);
const DINO_STARS = [[560, 210], [610, 170], [660, 160], [700, 180], [690, 225], [640, 240], [590, 250], [540, 270], [500, 300], [700, 120], [745, 135]];
const DINO_LINES = [[8, 7], [7, 6], [6, 5], [5, 4], [4, 3], [3, 2], [2, 1], [1, 0], [0, 6], [3, 9], [9, 10], [10, 3]];
class HillPage {
  constructor() {
    this.name = 'hill';
    this.tricks = [['shoot', 'stars', 'moon'], ['owl', 'sheep', 'bunny']]; this.who = ['Night sky', 'Hill friends', 'Little Dino'];
    this.text = 'Up the hill, under the big round moon. The stars come out, one by one, to light the way.';
    this.st = new Stage(); this.st.floor = 720;
    this.acts = new Acts({ shoot: 2.4, stars: 5, moon: 3.5, owl: 2.2, sheep: 4.8, bunny: 3 });
    this.stars = makeStars(110, 0, W, 110, 420);
    this.flies = makeFlies(12, 60, W - 60, 380, 640);
    this.owlBlink = 2;
    this.clouds = Array.from({ length: 3 }, () => ({ x: rnd(0, W), y: rnd(150, 300), v: rnd(5, 9) }));
    this.ambShoot = 8;
    this.sheep = [[770, 940], [700, 1010], [630, 1080]].map(([a, b]) => ({ a, b, side: 0 }));
    this.dino = new Dino(this.st, {
      x: 160, ground: hillGround, minX: 80, maxX: W - 120, tricks: ['yawn', 'wave', 'jump', 'roar'],
      host: {
        waveWord: 'HELLO, MOON!', waveTo: 1030, waved: () => this.acts.run('moon'),
        roarAt: 220, startle: () => { this.acts.run('owl'); sfx('hoo', 0.7, 1.2); },
        cheer: () => this.st.hearts(240, 300, 2),
      },
    });
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    const a = cycle(st, k, this.tricks[k]);
    this.acts.run(a);
    if (a === 'shoot') { sfx('shimmer', 1); st.say('MAKE A WISH!', 640, 200, 48, '#fff3a0'); }
    if (a === 'stars') { sfx('chime', 0.8); }
    if (a === 'moon') { sfx('sparkle', 0.8); }
    if (a === 'owl') { sfx('hoo', 1); st.say('HOO HOO!', 260, 250, 48, '#e6c8a0'); }
    if (a === 'bunny') { sfx('squeak', 0.6, 1.3); }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    for (const c of this.clouds) { c.x += c.v * dt; if (c.x > W + 120) c.x = -120; }
    this.owlBlink -= dt; if (this.owlBlink < -0.2) this.owlBlink = rnd(2, 5);
    if ((this.ambShoot -= dt) <= 0 && !A.on('shoot')) { this.ambShoot = rnd(10, 18); A.run('shoot'); sfx('shimmer', 0.3); }
    if (A.on('sheep')) for (let i = 0; i < 3; i++) {
      if (A.at('sheep', 0.3 + i * 1.4, dt)) { sfx('baa', 0.8, 1 + i * 0.12); st.say(String(i + 1), 860, 380, 64, '#ffffff'); }
      if (A.at('sheep', 1.4 + i * 1.4, dt)) this.sheep[i].side ^= 1;
    }
    if (A.at('moon', 0.4, dt)) st.say('HELLO, DINO!', 1030, 300, 42, '#fff3a0');
    if (A.at('bunny', 0.4, dt)) st.say('BOING!', 470, 480, 40, '#ffd2e6');
    updateFlies(this.flies, t, dt);
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, H, grad(0, 0, 0, 520, ['#121842', '#26346e', '#45589a']));
    drawStars(this.stars, t);
    // the dino constellation
    if (A.on('stars')) {
      const k = A.t('stars'), a = clamp(k / 0.5, 0, 1) * clamp((5 - k) / 0.6, 0, 1);
      ctx.globalAlpha = a * 0.7;
      DINO_LINES.forEach(([i, j], n) => {
        const p = clamp(k * 3 - n * 0.25, 0, 1);
        if (p > 0) { const [x1, y1] = DINO_STARS[i], [x2, y2] = DINO_STARS[j]; line(x1, y1, lerp(x1, x2, p), lerp(y1, y2, p), 2, '#cfe0ff', 0); }
      });
      ctx.globalAlpha = a;
      for (const [x, y] of DINO_STARS) { glow(x, y, 18, '#fff3a0', 0.6); star(x, y, 7, t, 4, '#fff6c8', 0); }
      ctx.globalAlpha = 1;
    }
    if (A.on('shoot')) {
      const k = A.t('shoot') / 1.4, x = lerp(120, 900, k), y = lerp(140, 300, k);
      if (k < 1) {
        ctx.globalAlpha = 1 - k * 0.5;
        line(x - 150, y - 30, x, y, 3, 'rgba(255,246,200,0.6)', 0);
        star(x, y, 12, t * 6, 5, '#fff6c8', 0); glow(x, y, 40, '#fff3a0', 0.6);
        ctx.globalAlpha = 1;
        if (Math.random() < 0.5) st.sparkles(x, y, 1, 6, '#fff6c8');
      }
    }
    const awake = A.on('moon') ? Math.sin(Math.PI * clamp(A.t('moon') / 3.5, 0, 1)) : 0;
    drawMoon(1030, 200, 72 + awake * 8, t, awake);
    for (const c of this.clouds) cloud(c.x, c.y, 1.2, '#7884c0', 0.5);
    // far hill with home on it: a tiny warm window
    ctx.beginPath(); ctx.moveTo(0, 520); ctx.quadraticCurveTo(400, 380, 820, 440); ctx.quadraticCurveTo(1100, 470, W, 360); ctx.lineTo(W, H); ctx.lineTo(0, H); ctx.closePath(); paint('#2c4072', 0);
    rect(1150, 344, 40, 30, '#3a2a40'); poly([1144, 346, 1196, 346, 1170, 322], '#5a3048', 0);
    rect(1162, 352, 10, 10, '#ffd27a'); glow(1167, 357, 30, '#ffd27a', 0.6 + 0.2 * Math.sin(t * 2));
    // near hill
    ctx.beginPath(); ctx.moveTo(0, H);
    for (let x = 0; x <= W; x += 20) ctx.lineTo(x, hillGround(x) + 6);
    ctx.lineTo(W, H); ctx.closePath(); paint('#3f7a74', 3);
    // the winding path
    ctx.beginPath(); for (let x = 0; x <= W; x += 20) ctx.lineTo(x, hillGround(x) + 16 + Math.sin(x / 140) * 8);
    ctx.lineWidth = 18; ctx.strokeStyle = 'rgba(200,190,160,0.35)'; ctx.stroke();
    this.drawTree(t);
    this.drawFence(t);
    this.drawBunny(t);
    grassTufts([60, 330, 520, 640, 760, 1100, 1220], x => hillGround(x) + 6, t, '#2f6460');
    for (const x of [380, 600, 1160]) { const g = hillGround(x); line(x, g + 4, x, g - 18, 3, '#2f6460', 0); circ(x, g - 22, 6, '#ffd2e6', 2); }
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    drawFlies(this.flies, t);
    st.drawWords();
  }
  drawTree(t) {
    line(200, hillGround(200) + 8, 210, 300, 30, '#5a3c3c');
    line(205, 380, 270, 350, 12, '#5a3c3c');
    for (const [x, y, r] of [[160, 250, 70], [240, 220, 80], [300, 270, 60], [200, 300, 60]]) circ(x, y, r, '#28504e', 3);
    for (const [x, y, r] of [[160, 250, 66], [240, 220, 76], [300, 270, 56], [200, 300, 56]]) circ(x, y, r, '#28504e', 0);
    // the owl on its branch
    const A = this.acts, o = A.on('owl') ? A.t('owl') : -1;
    const turn = o >= 0 ? Math.sin(Math.min(1, o / 1.2) * TAU) * 14 : 0;
    const flap = o >= 0 && o < 1.5 ? Math.sin(o * 20) * 12 : 0;
    const x = 262, y = 324;
    poly([x - 18, y - 10, x - 34, y + 10 - flap, x - 16, y + 22], '#a07850');
    poly([x + 18, y - 10, x + 34, y + 10 - flap, x + 16, y + 22], '#a07850');
    ell(x, y + 4, 22, 28, '#c89664'); ell(x, y + 10, 13, 16, '#f0dcb4', 2);
    push(x + turn * 0.4, y - 26, turn);
    poly([-18, -8, -14, -26, -6, -12], '#c89664'); poly([18, -8, 14, -26, 6, -12], '#c89664');
    circ(0, 0, 20, '#c89664');
    for (const s of [-1, 1]) {
      circ(s * 9, -2, 8, '#fff6dc', 2);
      if (this.owlBlink < 0) line(s * 9 - 5, -2, s * 9 + 5, -2, 2.5, INK, 0);
      else { circ(s * 9, -2, 4, INK, 0); circ(s * 9 - 1.5, -3.5, 1.4, WHITE, 0); }
    }
    poly([-4, 6, 4, 6, 0, 13], '#ffb43c', 2);
    pop();
  }
  drawFence(t) {
    const fx = x => hillGround(x);
    // a stile the sheep hop over when they are counted
    for (let x = 820; x <= 900; x += 40) line(x, fx(x) + 4, x, fx(x) - 54, 8, '#c8aa82');
    for (const dy of [-44, -24]) line(816, fx(816) + dy, 904, fx(904) + dy, 6, '#c8aa82');
    const A = this.acts;
    this.sheep.forEach((s, i) => {
      const from = s.side ? s.b : s.a, to = s.side ? s.a : s.b;
      const k = A.on('sheep') ? clamp((A.t('sheep') - 0.3 - i * 1.4) / 1.1, 0, 1) : 0;
      const x = lerp(from, to, ease(k)), lift = Math.sin(k * Math.PI) * 100;
      const g = fx(x) + 22, y = g - lift, moving = k > 0 && k < 1;
      push(x, y, moving ? (k - 0.5) * 30 * Math.sign(to - from) : 0, Math.sign(to - from) || 1, 1);
      for (const lx of [-14, 10]) line(lx, -6, lx, 6, 5, '#504650', 2);
      for (const [bx, by, r] of [[-14, -18, 14], [0, -24, 16], [14, -18, 14], [-6, -10, 13], [8, -10, 13]]) circ(bx, by, r, WHITE, 3);
      for (const [bx, by, r] of [[-14, -18, 12], [0, -24, 14], [14, -18, 12], [-6, -10, 11], [8, -10, 11]]) circ(bx, by, r, WHITE, 0);
      ell(26, -16, 11, 9, '#504650');
      if (!moving) curve(21, -18, 25, -14, 29, -18, 2, WHITE, 0); else circ(28, -18, 2.2, WHITE, 0);
      pop();
      if (!moving && Math.random() < 0.004) this.st.zzz(x + 10, g - 50);
    });
  }
  drawBunny(t) {
    const A = this.acts;
    const x = 470, g = hillGround(x) + 10;
    ell(x, g, 26, 8, '#2a4a46', 0);
    if (!A.on('bunny')) return;
    const b = A.t('bunny'), up = ease(b / 0.3) * (1 - ease((b - 2.5) / 0.3)), hop = Math.abs(Math.sin(b * 6)) * 10 * (b > 0.4 && b < 2.4 ? 1 : 0);
    ctx.save(); ctx.beginPath(); ctx.rect(x - 60, g - 120, 120, 120); ctx.clip();
    const y = g + 30 - up * 44 - hop;
    ell(x - 8, y - 40, 6, 18, '#f0ebf5'); ell(x + 8, y - 40, 6, 18, '#f0ebf5');
    ell(x, y - 10, 22, 20, '#f0ebf5');
    circ(x - 7, y - 14, 2.6, INK, 0); circ(x + 7, y - 14, 2.6, INK, 0);
    circ(x, y - 6 + Math.sin(t * 30) * 0.8, 3, '#ff9ab4', 0);
    ctx.restore();
  }
}
