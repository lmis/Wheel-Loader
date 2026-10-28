// ---------- page: bath time ----------
const TUB = { x0: 360, x1: 920, rim: 470, water: 492 };
class BathPage {
  constructor() {
    this.name = 'bath';
    this.text = 'Splish, splash! A warm bath before bed, with lots of bubbles and a little rubber duck.';
    this.tricks = [['bubbles', 'tap', 'toyboat'], ['duck', 'sponge', 'fishy'], ['splash', 'beard', 'dive']]; this.who = ['Bath', 'Bath toys', 'Little Dino'];
    this.st = new Stage(); this.st.floor = 700;
    this.acts = new Acts({ bubbles: 3, tap: 3.5, toyboat: 6, duck: 1.6, sponge: 2, fishy: 2.4, splash: 1.8, beard: 4, dive: 2.6 });
    this.stars = makeStars(10, 1040, 1180, 120, 250);
    this.blink = 2; this.bubT = 0; this.foam = 0;
  }
  press(k) {
    const st = this.st, a = cycle(st, k, this.tricks[k]);
    this.acts.run(a);
    if (a === 'bubbles') { sfx('blow', 0.4, 1.6); }
    if (a === 'tap') { sfx('tap', 1); sfx('creak', 0.3, 2.4); }
    if (a === 'toyboat') { sfx('toot', 1); }
    if (a === 'duck') { sfx('rubberduck', 1); st.say('SQUEAK!', 770, 380, 50, '#ffe050'); }
    if (a === 'sponge') { sfx('splash', 0.5, 1.6); }
    if (a === 'fishy') { sfx('squirt', 1); }
    if (a === 'splash') { sfx('splash', 1); st.say('SPLASH!', 600, 330, 66, '#8cd2ff'); st.drops(600, TUB.water - 10, 26, 420); }
    if (a === 'beard') { sfx('giggle', 0.8); st.say('HO HO HO!', 600, 330, 50, WHITE); }
    if (a === 'dive') { sfx('blub', 0.8, 0.8); }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    this.blink -= dt; if (this.blink < -0.15) this.blink = rnd(2, 5);
    this.bubT -= dt * (A.on('bubbles') ? 14 : 1);
    if (this.bubT <= 0) { this.bubT = rnd(0.5, 1.4); st.bubbles(rnd(TUB.x0 + 60, TUB.x1 - 60), TUB.water - 10, 1); }
    if (A.on('tap')) { this.foam = Math.min(1, this.foam + dt * 0.4); if (Math.random() < dt * 16) st.drops(TUB.x1 - 46, TUB.water - 6, 1, 140); }
    else this.foam = Math.max(0, this.foam - dt * 0.03);
    if (A.at('sponge', 0.5, dt)) { for (let i = 0; i < 24; i++) st.add({ kind: 'drop', layer: 2, x: 480 + rnd(-30, 30), y: 300, vx: rnd(-40, 40), vy: rnd(0, 80), size: rnd(3, 6), col: '#bfe6ff', life: 1, g: 900 }); st.say('DRIP DROP!', 480, 250, 44, '#bfe6ff'); }
    if (A.at('fishy', 0.4, dt)) st.say('SQUIRT!', 380, 300, 48, '#ff9a6e');
    if (A.on('fishy') && A.t('fishy') > 0.4 && A.t('fishy') < 1.4 && Math.random() < dt * 40)
      st.add({ kind: 'drop', layer: 2, x: 440, y: TUB.water - 20, vx: rnd(200, 260), vy: rnd(-420, -360), size: rnd(3, 5), col: '#bfe6ff', life: 1.2, g: 700 });
    if (A.on('dive')) {
      const d = A.t('dive');
      if (d > 0.3 && d < 1.6 && Math.random() < dt * 12) st.bubbles(600 + rnd(-30, 30), TUB.water, 1);
      if (A.at('dive', 1.7, dt)) { st.drops(600, TUB.water - 20, 20, 360); sfx('splash', 0.7, 1.3); st.say('TA-DA!', 600, 300, 60, '#ffd25a'); }
    }
    if (A.on('beard') && Math.random() < dt * 3) st.bubbles(640, 440, 1);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    // tiles and floor
    rect(0, 0, W, 600, '#b4e6dc');
    ctx.fillStyle = 'rgba(255,255,255,0.45)';
    for (let y = 0; y < 600; y += 60) ctx.fillRect(0, y, W, 3);
    for (let x = 0; x < W; x += 60) ctx.fillRect(x, 0, 3, 600);
    rect(0, 400, W, 10, '#8cc8bc');
    for (let y = 600, r = 0; y < H; y += 40, r++) for (let x = (r % 2) * 40; x < W; x += 80) rect(x, y, 40, 40, '#f5ebdc');
    rect(0, 600, W, H - 600, 'rgba(255,220,200,0.35)');
    // round window with the night outside
    circ(1110, 190, 84, '#fff6ea');
    ctx.save(); ctx.beginPath(); ctx.arc(1110, 190, 70, 0, TAU); ctx.clip(); rect(1030, 110, 160, 160, '#1e2456'); drawStars(this.stars, t); drawMoon(1140, 165, 22, t, 0); ctx.restore();
    line(1040, 190, 1180, 190, 6, '#fff6ea'); line(1110, 120, 1110, 260, 6, '#fff6ea');
    // shelf with bottles, and a mirror
    rrect(120, 200, 180, 220, 70, '#dff6ff'); ell(160, 260, 20, 40, 'rgba(255,255,255,0.7)', 0); ctx.lineWidth = 6; ctx.strokeStyle = '#f5d78c'; ctx.beginPath(); ctx.roundRect(120, 200, 180, 220, 70); ctx.stroke();
    rrect(960, 330, 200, 12, 4, '#c8965a');
    [['#ff9ab4', 26, 60], ['#8cd2ff', 22, 44], ['#ffe08a', 20, 52], ['#b4e6a0', 24, 38]].forEach(([c, w, h], i) => {
      const x = 980 + i * 46; rrect(x, 330 - h, w, h, 6, c); rrect(x + w / 2 - 4, 330 - h - 10, 8, 12, 2, WHITE);
    });
    // towel on its rail
    line(30, 470, 30, 600, 8, '#c8c8d7'); rrect(10, 476, 70, 120, 10, '#ffb4d2'); line(16, 560, 74, 560, 6, '#ff8cb4', 0);
    // the tub
    this.drawTub(t);
    st.draw(0); st.draw(1); st.draw(2);
    st.drawWords();
  }
  drawTub(t) {
    const A = this.acts, { x0, x1, rim, water } = TUB;
    // feet
    for (const x of [x0 + 50, x1 - 50]) { ell(x, 630, 22, 14, '#ffd250'); }
    // tap
    curve(x1 - 20, 380, x1 - 20, 340, x1 - 60, 350, 12, '#c8c8d7');
    circ(x1 - 10, 396, 10, '#c8c8d7'); line(x1 - 20, 380, x1 - 20, 470, 12, '#c8c8d7');
    if (A.on('tap')) { const k = Math.min(1, A.t('tap') * 3) * (1 - ease((A.t('tap') - 3) / 0.4)); ctx.globalAlpha = 0.7; line(x1 - 62, 356, x1 - 62, water, 8 * k, '#bfe6ff', 0); ctx.globalAlpha = 1; }
    // the back inside wall of the tub
    ctx.beginPath(); ctx.moveTo(x0, rim); ctx.lineTo(x1, rim); ctx.lineTo(x1 - 20, rim + 20); ctx.lineTo(x0 + 20, rim + 20); ctx.closePath(); paint('#e6f0f5', 0);
    // Little Dino in the bath
    this.drawDino(t);
    // the toys floating
    this.drawDuck(t); this.drawBoat(t); this.drawFish(t);
    // water and foam
    ctx.beginPath(); ctx.moveTo(x0 + 10, water);
    for (let x = x0 + 10; x <= x1 - 10; x += 20) ctx.lineTo(x, water + Math.sin(t * 3 + x * 0.05) * 3);
    ctx.lineTo(x1 - 10, water + 20); ctx.lineTo(x0 + 10, water + 20); ctx.closePath(); paint('#9ad8f0', 0);
    const nf = 18 + Math.floor(this.foam * 10);
    for (let i = 0; i < nf; i++) { const x = x0 + 30 + (i / (nf - 1)) * (x1 - x0 - 60), r = 18 + (i * 7 % 10) + this.foam * 8; circ(x, water - 2 + Math.sin(t * 2 + i) * 2, r, WHITE, 2); }
    for (let i = 0; i < nf; i++) { const x = x0 + 30 + (i / (nf - 1)) * (x1 - x0 - 60), r = 18 + (i * 7 % 10) + this.foam * 8; circ(x, water - 2 + Math.sin(t * 2 + i) * 2, r - 2, WHITE, 0); }
    // the outside of the tub
    ctx.beginPath(); ctx.moveTo(x0 - 20, rim + 10); ctx.quadraticCurveTo(x0 - 10, 640, x0 + 80, 630); ctx.lineTo(x1 - 80, 630); ctx.quadraticCurveTo(x1 + 10, 640, x1 + 20, rim + 10); ctx.closePath();
    paint('#ffffff', 3);
    rrect(x0 - 28, rim, x1 - x0 + 56, 24, 12, '#f5f8fa');
    ctx.globalAlpha = 0.5; rect(x0 + 20, 540, x1 - x0 - 40, 8, '#b4e6dc'); ctx.globalAlpha = 1;
  }
  drawDino(t) {
    const A = this.acts;
    const dive = A.on('dive') ? A.t('dive') : -1, under = dive >= 0 ? Math.sin(Math.PI * clamp((dive - 0.2) / 1.5, 0, 1)) : 0;
    const sp = A.on('splash') ? A.t('splash') : -1, splashing = sp >= 0 && sp < 1.2;
    const bob = Math.sin(t * 1.6) * 3 + under * 110 - (dive > 1.6 && dive < 2.2 ? Math.sin((dive - 1.6) / 0.6 * Math.PI) * 30 : 0);
    ctx.save(); ctx.beginPath(); ctx.rect(0, 0, W, TUB.water + 4); ctx.clip();
    // arms resting on the rim, or splashing
    const ay = splashing ? Math.sin(sp * 30) * 16 : 0;
    if (under < 0.3) { line(540, 470 + bob * 0.3, 500, 456 - ay, 12, DINO_GREEN.body); line(660, 470 + bob * 0.3, 700, 456 + ay, 12, DINO_GREEN.body); }
    push(600, 410 + bob, -6);
    const open = splashing ? 0.8 : (A.on('beard') ? 0.4 : A.on('duck') ? 0.5 : 0);
    drawDinoHead({ t, cols: DINO_GREEN, open, blink: this.blink < 0, lookUp: A.on('sponge') });
    // a bubble beard, or a bubble crown
    if (A.on('beard')) { const k = Math.min(1, A.t('beard') * 3) * (1 - ease((A.t('beard') - 3.4) / 0.5)); for (let i = 0; i < 7; i++) circ(20 + Math.cos(i) * 14, 34 + i * 4 * k, 12 * k, WHITE, 2); }
    circ(-10, -40, 10, WHITE, 2); circ(4, -44, 8, WHITE, 2);
    pop();
    ctx.restore();
  }
  drawDuck(t) {
    const A = this.acts, d = A.on('duck') ? A.t('duck') : -1, sq = d >= 0 && d < 0.4 ? Math.sin(d / 0.4 * Math.PI) : 0;
    const x = 770 + Math.sin(t * 0.6) * 40, y = TUB.water - 6 + Math.sin(t * 2.2) * 3 - (d >= 0 ? Math.abs(Math.sin(d * 8)) * 10 * (1 - d / 1.6) : 0);
    push(x, y, Math.sin(t * 2) * 6, 1 + sq * 0.15, 1 - sq * 0.2);
    ell(0, -14, 30, 18, '#ffe050'); poly([-26, -20, -40, -30, -28, -10], '#ffe050');
    circ(16, -38, 15, '#ffe050'); poly([28, -40, 44, -36, 28, -32], '#ff9a3c', 2.5);
    circ(20, -42, 3, INK, 0); circ(19, -43, 1, WHITE, 0);
    pop();
  }
  drawBoat(t) {
    const A = this.acts, b = A.on('toyboat') ? A.t('toyboat') : -1;
    if (b < 0) return;
    const x = lerp(420, 880, b / 6), y = TUB.water - 4 + Math.sin(t * 3) * 3;
    push(x, y, Math.sin(t * 2.4) * 6);
    poly([-34, -10, 34, -10, 24, 8, -24, 8], '#ff6e6e');
    rrect(-12, -32, 26, 22, 4, WHITE); rrect(-4, -46, 10, 16, 2, '#5a82d2');
    if (Math.sin(b * 3) > 0.9) this.st.puff(x + 2, y - 50, 6, '#ffffff', 2);
    pop();
  }
  drawFish(t) {
    const A = this.acts, f = A.on('fishy') ? A.t('fishy') : -1, x = 440, y = TUB.water - 4 + Math.sin(t * 2.6 + 1) * 3;
    const up = f >= 0 ? Math.sin(Math.PI * clamp(f / 2.4, 0, 1)) * 16 : 0;
    push(x, y - up, -20 * (up / 16));
    poly([-26, -6, -40, -18, -40, 6], '#ff9a6e', 2.5);
    ell(0, -6, 26, 14, '#ffb47a');
    circ(12, -10, 4, WHITE, 2); circ(13, -10, 2, INK, 0);
    ell(24, -4, 4, 3 + up * 0.1, '#ff6e6e', 2);
    pop();
  }
}
