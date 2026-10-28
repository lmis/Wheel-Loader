// ---------- page: past the sleepy farm ----------
class FarmPage {
  constructor() {
    this.name = 'farm';
    this.text = 'Past the sleepy farm. The cows are munching, the chickens are roosting, and the old red barn says goodnight.';
    this.tricks = [['barn', 'tractor', 'crows'], ['cow', 'pig', 'chickens']]; this.who = ['Farm', 'Farm animals', 'Little Dino'];
    this.st = new Stage(); this.st.floor = 680;
    this.acts = new Acts({ barn: 4.2, tractor: 6, crows: 3.6, cow: 2.2, pig: 2.4, chickens: 2.4 });
    this.stars = makeStars(30, 0, W, 110, 280);
    this.flies = makeFlies(10, 40, W - 40, 420, 620);
    this.cowBlink = 2; this.eggs = [];
    this.dino = new Dino(this.st, {
      x: 140, ground: () => 668, minX: 80, maxX: W - 110, tricks: ['wave', 'jump', 'roar', 'spin'],
      host: {
        waveWord: 'GOODNIGHT, COW!', waveTo: 380, waved: () => { this.acts.run('cow'); sfx('moo', 0.8); },
        roarAt: 220, startle: () => { this.acts.run('chickens'); sfx('bawk', 1, 1.2); },
        cheer: () => this.st.hearts(380, 430, 2),
      },
    });
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    const a = cycle(st, k, this.tricks[k]);
    this.acts.run(a);
    if (a === 'barn') sfx('creak', 1, 0.8);
    if (a === 'tractor') sfx('tractor', 1);
    if (a === 'crows') { sfx('caw', 1); st.say('CAW CAW!', 1180, 300, 48, '#c8c8dc'); }
    if (a === 'cow') { sfx('moo', 1); st.say('MOO!', 380, 360, 64, WHITE); }
    if (a === 'pig') { sfx('oink', 1); st.say('OINK OINK!', 600, 470, 50, '#ffb4c8'); }
    if (a === 'chickens') { sfx('bawk', 1); st.say('BAWK!', 200, 420, 52, '#fff0b4'); }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    this.cowBlink -= dt; if (this.cowBlink < -0.2) this.cowBlink = rnd(2, 5);
    if (A.at('barn', 1.0, dt)) { sfx('neigh', 1); st.say('NEIGH!', 850, 330, 56, '#ffe0b4'); }
    if (A.at('pig', 0.9, dt)) { st.drops(600, 600, 10, 220, '#8c5a3c'); sfx('splash', 0.4, 0.7); }
    if (A.at('chickens', 1.1, dt)) { this.eggs.push({ x: 200 + rnd(-30, 30), y: 540, vy: 0, t: 0 }); sfx('pop', 0.5, 1.4); st.say('AN EGG!', 230, 470, 40, '#fff6dc'); }
    for (const e of this.eggs) { e.t += dt; e.vy += 900 * dt; e.y = Math.min(640, e.y + e.vy * dt); }
    this.eggs = this.eggs.filter(e => e.t < 6);
    if (A.on('tractor') && Math.floor(A.t('tractor') * 4) !== Math.floor((A.t('tractor') - dt) * 4)) st.puff(-260 + A.t('tractor') * 300 + 30, 470, 10, '#c8c0d2', 0);
    updateFlies(this.flies, t, dt);
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, 520, grad(0, 0, 0, 520, ['#3c4a8c', '#8c78c0', '#e8a6a6', '#f8c79a']));
    drawStars(this.stars, t, 0.6);
    glow(260, 470, 220, '#ffc88c', 0.4);
    // rolling fields
    ctx.beginPath(); ctx.moveTo(0, 470); ctx.quadraticCurveTo(300, 420, 640, 460); ctx.quadraticCurveTo(980, 500, W, 440); ctx.lineTo(W, H); ctx.lineTo(0, H); ctx.closePath(); paint('#7aa86a', 0);
    ctx.globalAlpha = 0.25; for (let i = 0; i < 6; i++) { ctx.beginPath(); ctx.moveTo(0, 480 + i * 14); ctx.quadraticCurveTo(640, 470 + i * 16, W, 470 + i * 12); ctx.lineWidth = 3; ctx.strokeStyle = '#4e7a4a'; ctx.stroke(); } ctx.globalAlpha = 1;
    rect(0, 540, W, H - 540, grad(0, 540, 0, H, ['#8cb874', '#6e9a5a']));
    this.drawScarecrow(t);
    this.drawBarn(t);
    this.drawCow(t);
    // the lane where the tractor drives
    rect(0, 600, W, 34, '#c8aa82'); rect(0, 598, W, 3, 'rgba(70,40,47,0.35)');
    if (A.on('tractor')) this.drawTractor(-260 + A.t('tractor') * 300, 604, t);
    this.drawFence(t);
    this.drawPig(t);
    for (const e of this.eggs) { ell(e.x, e.y - 9, 8, 10, '#fff6dc', 2.5); }
    // hay bales
    for (const [x, y] of [[1080, 580], [1150, 586], [1115, 548]]) { rrect(x - 34, y - 26, 68, 46, 10, '#f0cd6e'); line(x - 26, y - 4, x + 26, y - 4, 2, '#c8a046', 0); }
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    drawFlies(this.flies, t);
    st.drawWords();
  }
  drawBarn(t) {
    const A = this.acts, x0 = 700, x1 = 980, base = 560;
    rrect(1000, 300, 70, 260, 30, '#c8c8d7'); ell(1035, 302, 35, 18, '#a0a0b4');
    ctx.beginPath(); ctx.moveTo(x0, base); ctx.lineTo(x0, 400); ctx.lineTo(x0 + 40, 330); ctx.lineTo(840, 290); ctx.lineTo(x1 - 40, 330); ctx.lineTo(x1, 400); ctx.lineTo(x1, base); ctx.closePath();
    paint('#d2505a', 3);
    ctx.beginPath(); ctx.moveTo(x0 - 10, 404); ctx.lineTo(x0 + 38, 326); ctx.lineTo(840, 284); ctx.lineTo(x1 - 38, 326); ctx.lineTo(x1 + 10, 404); ctx.lineWidth = 12; ctx.strokeStyle = INK; ctx.stroke(); ctx.lineWidth = 7; ctx.strokeStyle = '#fff6ea'; ctx.stroke();
    // loft window, lit
    glow(840, 360, 60, '#ffd27a', 0.5); rrect(816, 340, 48, 40, 6, '#ffe08a'); line(840, 340, 840, 380, 3, INK, 0);
    // doors that swing open, with a horse inside
    const open = A.on('barn') ? ease(A.t('barn') / 0.6) * (1 - ease((A.t('barn') - 3.6) / 0.5)) : 0;
    rrect(770, 420, 140, 140, 4, '#3a2a30');
    if (open > 0.1) {
      glow(840, 490, 80, '#ffd27a', 0.4 * open);
      const nod = Math.sin(t * 6) * 3, hx = 840, hy = 480 + (1 - open) * 40;
      ell(hx, hy + 30, 26, 40, '#a0643c');
      ell(hx + 4, hy + nod, 22, 34, '#b4783c', 3);
      ell(hx + 6, hy + 26 + nod, 15, 12, '#dcb48c', 2.5);
      poly([hx - 14, hy - 28 + nod, hx - 6, hy - 46 + nod, hx, hy - 28 + nod], '#b4783c');
      poly([hx + 8, hy - 28 + nod, hx + 16, hy - 46 + nod, hx + 20, hy - 28 + nod], '#b4783c');
      for (let i = 0; i < 4; i++) line(hx - 14 + i * 6, hy - 30 + nod, hx - 18 + i * 6, hy - 12 + nod, 4, '#5a3c28', 0);
      circ(hx - 6, hy - 6 + nod, 3, INK, 0); circ(hx + 16, hy - 6 + nod, 3, INK, 0);
    }
    for (const [dx, s] of [[770, 1], [910, -1]]) {
      push(dx, 0, 0, s * (1 - open * 0.85), 1);
      rect(0, 420, 70, 140, '#e8606a'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(0, 420, 70, 140);
      line(6, 426, 64, 554, 6, '#fff6ea', 0); line(64, 426, 6, 554, 6, '#fff6ea', 0);
      ctx.strokeStyle = '#fff6ea'; ctx.lineWidth = 6; ctx.strokeRect(5, 425, 60, 130);
      pop();
    }
    // a weathervane rooster turning in the breeze
    push(840, 284, 0, Math.cos(t * 0.5), 1); line(0, 0, 0, -40, 3, INK, 0); poly([-14, -44, 14, -44, 8, -58, -6, -52], '#3a2a30', 0); pop();
  }
  drawCow(t) {
    const A = this.acts, c = A.on('cow') ? A.t('cow') : -1, x = 380, y = 520;
    const chew = Math.sin(t * 5) * 2, moo = c >= 0 && c < 1.4 ? Math.sin(Math.PI * c / 1.4) : 0;
    const sw = Math.sin(t * 2.2) * 14 + (c >= 0 ? Math.sin(c * 12) * 12 : 0);
    curve(x + 54, y - 34, x + 72, y - 10, x + 66 + sw * 0.4, y + 18, 4, WHITE);
    circ(x + 66 + sw * 0.4, y + 22, 6, INK, 0);
    for (const lx of [-34, -16, 26, 44]) line(x + lx, y - 6, x + lx, y + 34, 9, WHITE);
    ell(x + 6, y - 32, 62, 36, WHITE);
    for (const [sx, sy, r] of [[-20, -44, 14], [20, -22, 12], [36, -46, 10]]) ell(x + sx, y + sy, r * 1.3, r, '#3a3a46', 0);
    ell(x + 14, y + 2, 12, 8, '#ffb4c8', 2);
    push(x - 58, y - 52 - moo * 8, -moo * 15);
    poly([-14, -22, -24, -34, -6, -28], '#f5ebdc', 2.5); poly([14, -22, 24, -34, 6, -28], '#f5ebdc', 2.5);
    ell(-26, -14, 12, 6, WHITE); ell(26, -14, 12, 6, WHITE);
    ell(0, 0, 24, 28, WHITE);
    ell(0, 18 + chew * 0.3, 20, 13 + moo * 6, '#ffb4c8', 2.5);
    circ(-7, 18, 3, '#b4647a', 0); circ(7, 18, 3, '#b4647a', 0);
    if (this.cowBlink < 0) { line(-14, -6, -6, -6, 2.5, INK, 0); line(6, -6, 14, -6, 2.5, INK, 0); }
    else { circ(-10, -6, 4, INK, 0); circ(10, -6, 4, INK, 0); }
    // a little bell
    circ(0, 36, 6, '#ffd250', 2);
    pop();
  }
  drawFence(t) {
    for (let x = 20; x < W; x += 80) line(x, 600, x, 530, 9, '#b48c64');
    line(0, 548, W, 548, 7, '#c8a078'); line(0, 576, W, 576, 7, '#c8a078');
    // chickens on the top rail
    const A = this.acts, f = A.on('chickens') ? A.t('chickens') : -1;
    [[150, 0], [205, 1], [260, 2]].forEach(([x, i]) => {
      const hop = f >= 0 && f < 1.6 ? Math.abs(Math.sin(f * 10 + i)) * 26 : 0, flap = f >= 0 && f < 1.6 ? Math.sin(f * 30 + i) * 10 : 0;
      const sleepy = f < 0, y = 540 - hop;
      push(x, y, 0, i === 1 ? -1 : 1, 1);
      ell(0, -12, 18, 15, WHITE);
      poly([-4, -16, -20, -26 - flap, -16, -8], '#f0ebe1', 2.5);
      circ(12, -28, 10, WHITE);
      poly([8, -38, 12, -46, 16, -38, 20, -42, 18, -34], '#f05050', 2);
      poly([21, -29, 29, -26, 21, -23], '#ffb43c', 2);
      if (sleepy && Math.sin(t * 0.7 + i) > -0.6) curve(11, -30, 14, -27, 17, -30, 2, INK, 0); else circ(14, -30, 2, INK, 0);
      line(-4, 0, -4, 6, 2, '#ffb43c', 0); line(4, 0, 4, 6, 2, '#ffb43c', 0);
      pop();
    });
  }
  drawPig(t) {
    const A = this.acts, p = A.on('pig') ? A.t('pig') : -1, x = 600, y = 640;
    ell(x, y, 80, 18, '#8c5a3c', 2.5); ell(x - 20, y - 4, 30, 6, '#a0704a', 0);
    const roll = p >= 0 ? Math.sin(Math.PI * clamp((p - 0.5) / 1.2, 0, 1)) : 0;
    push(x, y - 26, roll * 180);
    curve(-34, -4, -46, -14, -40, -24, 3, '#ffb4c8');
    ell(0, 0, 38, 26, '#ffb4c8');
    if (roll > 0.2) { circ(-10, 6, 6, '#8c5a3c', 0); circ(14, -8, 5, '#8c5a3c', 0); }
    circ(30, -10, 18, '#ffb4c8');
    poly([22, -26, 26, -38, 34, -26], '#ffa0b8', 2.5);
    ell(44, -6, 9, 8, '#ff96b0', 2.5); circ(42, -6, 1.8, INK, 0); circ(47, -6, 1.8, INK, 0);
    circ(30, -16, 2.6, INK, 0);
    for (const lx of [-20, 18]) line(lx, 20, lx, 28, 7, '#ffb4c8');
    pop();
  }
  drawScarecrow(t) {
    const A = this.acts, c = A.on('crows') ? A.t('crows') : -1, x = 1190, y = 520;
    const wig = c >= 0 && c < 1.2 ? Math.sin(c * 20) * 8 : 0;
    line(x, y + 40, x, y - 120, 7, '#8c5a3c');
    push(x, y - 80, wig);
    line(-70, 0, 70, 0, 6, '#8c5a3c');
    poly([-30, -10, 30, -10, 36, 60, -36, 60], '#6e8cd2');
    circ(0, -34, 22, '#f0cd6e');
    poly([-34, -44, 34, -44, 0, -84], '#a0643c');
    circ(-8, -36, 3, INK, 0); circ(8, -36, 3, INK, 0);
    ctx.beginPath(); ctx.arc(0, -28, 8, 0.3, Math.PI - 0.3); ctx.lineWidth = 2.5; ctx.strokeStyle = INK; ctx.stroke();
    pop();
    // crows on its arms fly up and come back
    [[-56, 0], [52, 1]].forEach(([dx, i]) => {
      let bx = x + dx, by = y - 92, flap = null;
      if (c >= 0) { const k = clamp(c / 3.2, 0, 1); bx += Math.sin(k * TAU) * 160 * (i ? 1 : -1); by -= Math.sin(k * Math.PI) * 160; flap = k < 1 ? c * 24 + i : null; }
      drawBird(bx, by, 0.85, flap, '#3c3c4e', i ? -1 : 1);
    });
  }
  drawTractor(x, y, t) {
    const bob = Math.sin(t * 20) * 1.5;
    circ(x + 20, y + 44, 30, '#3c3c4a'); circ(x + 20, y + 44, 12, '#ffd250');
    circ(x + 110, y + 54, 20, '#3c3c4a'); circ(x + 110, y + 54, 8, '#ffd250');
    rrect(x + 40, y + 4 + bob, 90, 40, 8, '#5ab46e');
    rrect(x - 4, y - 40 + bob, 56, 64, 8, '#5ab46e'); rrect(x + 4, y - 32 + bob, 40, 30, 5, '#cdeeff');
    line(x + 100, y + 4 + bob, x + 100, y - 24 + bob, 7, '#3c3c4a');
    glow(x + 136, y + 20, 60, '#fff3a0', 0.6); circ(x + 132, y + 20 + bob, 7, '#fff6be');
    // a robot is driving!
    rrect(x + 14, y - 26 + bob, 20, 18, 5, '#5acdc3', 2); circ(x + 20, y - 18 + bob, 2, INK, 0); circ(x + 28, y - 18 + bob, 2, INK, 0);
  }
}
