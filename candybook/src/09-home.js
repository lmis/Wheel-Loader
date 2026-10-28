// ---------- page: home, with Mama at the door ----------
class HomePage {
  constructor() {
    this.name = 'home';
    this.tricks = [['door', 'chimney', 'lantern'], ['mamawave', 'kiss', 'call']]; this.who = ['House', 'Mama', 'Little Dino'];
    this.text = 'And there it is. Home! The windows glow warm and yellow, and Mama is waiting at the door.';
    this.st = new Stage(); this.st.floor = 720;
    this.acts = new Acts({ door: 6, chimney: 4, lantern: 4, hug: 2.6 });
    this.stars = makeStars(60, 0, W, 110, 330);
    this.flies = makeFlies(14, 40, W - 40, 360, 640);
    this.mama = new Dino(this.st, { x: 760, ground: () => 600, face: -1, cols: MAMA_COLS, scale: 1.3, bow: true, wander: false, tricks: [],
      host: { waveWord: 'WELCOME HOME!', trick: (d, t, dt) => this.mamaTrick(d, t, dt) } });
    this.dino = new Dino(this.st, {
      x: 180, ground: () => 652, minX: 80, maxX: 640, tricks: ['hug', 'jump', 'wave', 'spin'],
      host: { waveWord: 'HI, MAMA!', waveTo: 760, waved: () => this.mama.start('wave'), trick: (d, t, dt) => this.hugTrick(d, t, dt),
        startle: () => this.mama.start('surprised'), roarAt: 760 },
    });
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    if (k === 0) {
      const a = cycle(st, 0, this.tricks[0]);
      this.acts.run(a);
      if (a === 'door') { sfx('creak', 1); }
      if (a === 'lantern') { sfx('sparkle', 0.8); }
      if (a === 'chimney') { sfx('blow', 0.5, 1.4); }
    } else {
      const m = this.mama;
      if (m.trick) { m.pending = true; return; }
      const a = cycle(st, 1, this.tricks[1]);
      m.start(a === 'mamawave' ? 'wave' : a);
    }
  }
  mamaTrick(d, t, dt) {
    const st = this.st;
    if (d.trick === 'kiss') {
      if (t > 0.4 && t - dt <= 0.4) { sfx('smooch', 1); st.say('MWAH!', d.x - 40, 330, 50, '#ff8cc8'); }
      if (t > 0.4 && t < 1.4 && Math.floor(t * 6) !== Math.floor((t - dt) * 6))
        st.add({ kind: 'heart', layer: 2, x: d.x - 60, y: 470, vx: (this.dino.x - d.x) * 0.9, vy: -40, rot: 0, size: 12, col: '#ff6082', life: 1.2 });
      return t > 2;
    }
    if (d.trick === 'call') {
      if (t > 0.2 && t - dt <= 0.2) { sfx('hug', 0.8); st.say('I MISSED YOU!', d.x - 60, 320, 46, '#d2b4ff'); st.hearts(d.x, 440, 3); }
      return t > 2;
    }
    if (d.trick === 'surprised') { d.squash = 0.6; return t > 0.6; }
    return true;
  }
  hugTrick(d, t, dt) {
    const st = this.st, m = this.mama;
    if (d.stage === 0) { // run to Mama in quick hops
      if (d.stepHop(dt) || !d.hopping) {
        if (d.x > 600) { d.stage = 1; d.tt = 0; this.acts.run('hug'); sfx('hug', 1); st.say('HUG!', 690, 330, 70, '#ff8cc8'); st.hearts(690, 470, 6); m.squash = 0.6; }
        else { d.hop(Math.min(d.x + 70, 640), 0, 0.24, 20); sfx('hop', 0.4, 1.3); }
      }
      return false;
    }
    if (t > 0.5 && t < 2 && Math.floor(t * 4) !== Math.floor((t - dt) * 4)) { st.hearts(690, 470, 1); }
    if (t > 1.2 && t - dt <= 1.2) sfx('smooch', 0.8);
    return t > 2.4;
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    if (Math.floor(t * 0.7) !== Math.floor((t - dt) * 0.7)) st.puff(1010, 250, 14, '#c8c0d8', 0);
    if (A.on('chimney') && Math.floor(A.t('chimney') * 3) !== Math.floor((A.t('chimney') - dt) * 3))
      st.add({ kind: 'heart', layer: 0, x: 1010, y: 240, vx: rnd(-20, 20), vy: -60, rot: 0, size: rnd(12, 18), col: '#ffb4cc', life: 2.4 });
    if (A.at('door', 0.8, dt)) { sfx('chime', 0.7); st.say('COME IN!', 860, 300, 50, '#ffe08a'); }
    updateFlies(this.flies, t, dt, A.on('lantern') && A.t('lantern') < 3.4 ? (f => {
      const i = this.flies.indexOf(f), a = i / this.flies.length * TAU + t * 1.4;
      return [700 + Math.cos(a) * 50, 470 + Math.sin(a) * 40];
    }) : null);
    this.mama.update(dt);
    if (!this.mama.trick) this.mama.face = -1;
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, H, grad(0, 0, 0, 560, ['#18204e', '#2e3c7a', '#4a5490']));
    drawStars(this.stars, t);
    drawMoon(170, 190, 50, t, 0);
    // garden
    rect(0, 560, W, H - 560, grad(0, 560, 0, H, ['#4e8a68', '#3c7058']));
    ctx.beginPath(); ctx.moveTo(0, 566); ctx.quadraticCurveTo(640, 540, W, 566); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.stroke();
    this.drawTree(t);
    this.drawHouse(t);
    // path stones to the door
    for (let i = 0; i < 6; i++) ell(860 - i * 90, 618 + i * 14, 34 - i * 1, 10, '#b4aab4', 2.5);
    // fence
    for (let x = 20; x < 560; x += 36) { poly([x, 620, x + 22, 620, x + 22, 560, x + 11, 546, x, 560], '#f0ebe1', 2.5); }
    rect(10, 578, 550, 8, '#f0ebe1');
    // mailbox
    line(600, 620, 600, 560, 7, '#8c5a3c'); rrect(578, 528, 46, 34, 14, '#5a82d2'); line(622, 532, 622, 512, 3, INK, 0); poly([622, 512, 640, 517, 622, 522], '#ff5c78', 2);
    // closed night flowers
    for (const x of [120, 200, 300, 1110, 1190]) { const sw = Math.sin(t * 1.3 + x) * 3; line(x, 650, x + sw, 616, 3, '#2f6450', 0); ell(x + sw, 610, 6, 10, '#ffb4d2', 2); }
    st.draw(0);
    this.mama.draw();
    st.draw(1);
    this.dino.draw();
    st.draw(2);
    drawFlies(this.flies, t, A.on('lantern') ? 0.6 : 0);
    st.drawWords();
  }
  drawTree(t) {
    line(300, 600, 310, 300, 34, '#6e4a3c');
    line(305, 360, 420, 330, 14, '#6e4a3c');
    for (const [x, y, r] of [[250, 250, 80], [340, 220, 90], [420, 280, 70], [300, 320, 70]]) circ(x, y, r, '#2e6450', 3);
    for (const [x, y, r] of [[250, 250, 76], [340, 220, 86], [420, 280, 66], [300, 320, 66]]) circ(x, y, r, '#2e6450', 0);
    // a swing, gently swaying
    const a = Math.sin(t * 1.2) * 0.25;
    push(400, 334, a * 57.3);
    line(-22, 0, -22, 180, 2, INK, 0); line(22, 0, 22, 180, 2, INK, 0);
    rrect(-34, 176, 68, 12, 4, '#c8965a');
    pop();
  }
  drawHouse(t) {
    const A = this.acts, x0 = 640, x1 = 1080, top = 330, base = 600;
    rrect(990, 200, 44, 120, 4, '#a05a50');
    rect(x0, top, x1 - x0, base - top, '#f5e6c8'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(x0, top, x1 - x0, base - top);
    poly([x0 - 40, top + 10, x1 + 40, top + 10, (x0 + x1) / 2, top - 150], '#d2505a');
    ctx.globalAlpha = 0.35;
    for (let r = 0; r < 5; r++) for (let x = x0 - 20 + r * 40; x < x1 + 20 - r * 40; x += 34) {
      ctx.beginPath(); ctx.arc(x, top - r * 30, 15, 0, Math.PI); ctx.lineWidth = 2; ctx.strokeStyle = INK; ctx.stroke();
    }
    ctx.globalAlpha = 1;
    // round attic window
    circ((x0 + x1) / 2, top - 60, 26, '#ffe08a'); glow((x0 + x1) / 2, top - 60, 50, '#ffd27a', 0.4);
    line((x0 + x1) / 2 - 26, top - 60, (x0 + x1) / 2 + 26, top - 60, 3, INK, 0); line((x0 + x1) / 2, top - 86, (x0 + x1) / 2, top - 34, 3, INK, 0);
    // windows with curtains
    for (const wx of [690, 980]) {
      glow(wx + 35, 430, 90, '#ffd27a', 0.35 + 0.05 * Math.sin(t * 3 + wx));
      rrect(wx, 390, 70, 80, 8, '#ffe08a');
      const sw = Math.sin(t * 0.8 + wx) * 3;
      poly([wx + 3, 393, wx + 26 + sw, 393, wx + 14, 467, wx + 3, 467], '#ff8ca0', 2);
      poly([wx + 67, 393, wx + 44 + sw, 393, wx + 56, 467, wx + 67, 467], '#ff8ca0', 2);
      rrect(wx - 6, 470, 82, 16, 4, '#a0643c');
      for (let i = 0; i < 4; i++) circ(wx + 6 + i * 20, 466, 7, CANDY[(i + wx) % 7], 2);
    }
    // the round green door, opening
    const open = A.on('door') ? ease(A.t('door') / 0.6) * (1 - ease((A.t('door') - 5.2) / 0.6)) : 0;
    const dx = 820, dw = 90, dh = 140;
    ctx.beginPath(); ctx.moveTo(dx, base); ctx.lineTo(dx, base - dh + dw / 2); ctx.arc(dx + dw / 2, base - dh + dw / 2, dw / 2, Math.PI, 0); ctx.lineTo(dx + dw, base); ctx.closePath();
    paint(open > 0 ? '#ffd88a' : '#3a2a30', 3);
    if (open > 0) {
      ctx.globalAlpha = 0.35 * open; poly([dx, base, dx + dw, base, dx + dw + 140, H, dx - 160, H], '#ffe9a8', 0); ctx.globalAlpha = 1;
      glow(dx + dw / 2, base - 60, 160, '#ffd27a', 0.5 * open);
    }
    push(dx, 0, 0, 1 - open * 0.85, 1);
    ctx.beginPath(); ctx.moveTo(0, base); ctx.lineTo(0, base - dh + dw / 2); ctx.arc(dw / 2, base - dh + dw / 2, dw / 2, Math.PI, 0); ctx.lineTo(dw, base); ctx.closePath();
    paint('#5aa06e', 3);
    circ(dw - 16, base - 66, 6, '#ffd250');
    circ(dw / 2, base - dh + 40, 14, '#ffe08a', 2.5);
    pop();
    // lantern by the door
    const lg = A.on('lantern') ? 0.4 : 0;
    glow(700, 470, 90 + lg * 80, '#ffd27a', 0.5 + lg);
    line(700, 430, 700, 448, 3, INK, 0); rrect(688, 448, 24, 32, 6, '#ffe08a');
    ell(860, base + 4, 70, 8, 'rgba(70,40,47,0.15)', 0);
  }
}

class BedPage {
  constructor() {
    this.name = 'bed';
    this.tricks = [['mobile', 'glowstars', 'lamp'], ['read', 'kiss', 'robot'], ['yawn', 'giggle', 'sleep']]; this.who = ['Bedroom', 'Mama & robot', 'Little Dino'];
    this.text = 'Home, snug and warm. Mama reads a story, and the candy mobile turns round and round. Goodnight, Little Dino.';
    this.st = new Stage(); this.st.floor = 720;
    this.acts = new Acts({ mobile: 3.5, glowstars: 5, lamp: 6, read: 3, kiss: 2.4, robot: 2.4, yawn: 2.4, giggle: 2.2, sleep: 5 });
    this.stars = makeStars(18, 140, 340, 160, 360);
    this.ceilingStars = makeStars(40, 20, W - 20, 120, 380);
    this.blink = 2; this.mamaBlink = 3; this.page = 0;
  }
  press(k) {
    const st = this.st;
    const a = cycle(st, k, this.tricks[k]);
    this.acts.run(a);
    if (a === 'mobile') sfx('chime', 1);
    if (a === 'glowstars') sfx('shimmer', 0.7);
    if (a === 'lamp') { sfx('click', 1, 0.6); }
    if (a === 'read') { sfx('swoosh', 0.4, 1.4); this.page++; st.say('ONCE UPON A TIME...', 980, 300, 36, '#ffe9b4'); }
    if (a === 'kiss') { sfx('smooch', 1); }
    if (a === 'robot') { sfx('bloop', 0.8, 1.2); st.say('BEEP BOOP!', 390, 380, 40, '#8cffc8'); }
    if (a === 'yawn') { sfx('yawn', 1); st.say('YAWN!', 590, 380, 48, '#cfe0ff'); }
    if (a === 'giggle') { sfx('giggle', 1); st.hearts(600, 440, 3); }
    if (a === 'sleep') { sfx('snore', 1); }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    this.blink -= dt; if (this.blink < -0.2) this.blink = rnd(2.5, 5);
    this.mamaBlink -= dt; if (this.mamaBlink < -0.15) this.mamaBlink = rnd(2, 5);
    if (A.at('kiss', 0.5, dt)) st.hearts(600, 430, 4);
    if (A.on('sleep') && Math.floor(A.t('sleep') * 1.2) !== Math.floor((A.t('sleep') - dt) * 1.2)) st.zzz(640, 420);
    if (A.on('sleep') && A.at('sleep', 2.4, dt)) sfx('snore', 0.8);
    if (!A.on('sleep') && this.sleepy() > 0.6 && Math.random() < dt * 0.25) st.zzz(640, 420);
    if (A.on('mobile') && Math.random() < dt * 6) st.sparkles(700, 270, 1, 80, '#fff3a0');
  }
  sleepy() { return clamp((this.st.t - 4) / 50, 0, 0.75); }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const dark = A.on('lamp') ? ease(A.t('lamp') / 0.4) * (1 - ease((A.t('lamp') - 5.4) / 0.5)) : 0;
    rect(0, 0, W, 600, grad(0, 0, 0, 600, ['#5a5aa0', '#7a74b4']));
    ctx.globalAlpha = 0.25; for (let x = 30; x < W; x += 90) for (let y = 150; y < 560; y += 90) star(x + (y / 90 % 2) * 45, y, 7, 0, 5, '#fff3c8', 0); ctx.globalAlpha = 1;
    // window with the moon
    rrect(140, 150, 210, 230, 18, '#1e2456');
    ctx.save(); ctx.beginPath(); ctx.roundRect(140, 150, 210, 230, 18); ctx.clip(); drawStars(this.stars, t); drawMoon(290, 220, 36, t, 0); ctx.restore();
    rect(242, 150, 6, 230, '#f5ebdc'); rect(140, 262, 210, 6, '#f5ebdc');
    for (const [x, s] of [[118, 1], [372, -1]]) { const sw = Math.sin(t * 0.7) * 4; poly([x - 22 * s, 130, x + 26 * s + sw, 130, x + 10 * s, 410, x - 22 * s, 410], '#ff9ab4', 3); }
    rect(100, 128, 290, 10, '#c8965a');
    // floor and rug
    rect(0, 600, W, H - 600, grad(0, 600, 0, H, ['#b47c5a', '#9a6448']));
    ell(700, 670, 300, 40, '#8cc8b4'); ell(700, 670, 240, 28, '#b4e0d2', 0); ell(700, 670, 170, 18, '#8cc8b4', 0);
    // lamp on the bedside table, and the robot nightlight
    rrect(330, 470, 130, 20, 6, '#c8965a'); line(350, 490, 350, 600, 8, '#a0703c'); line(440, 490, 440, 600, 8, '#a0703c');
    if (!dark) glow(440, 400, 200, '#ffd27a', 0.45);
    line(440, 470, 440, 420, 5, '#a0703c'); poly([410, 425, 470, 425, 456, 385, 424, 385], dark ? '#c8a878' : '#ffe08a');
    this.drawNightlight(t, dark);
    // mobile
    this.drawMobile(t);
    // bed
    this.drawBed(t);
    // Mama in her armchair, reading
    this.drawMama(t);
    // the lamp goes off: glow-in-the-dark stars on the ceiling
    if (dark > 0) {
      ctx.globalAlpha = dark * 0.55; rect(0, 0, W, H, '#0c0e2a'); ctx.globalAlpha = dark;
      for (const s of this.ceilingStars) { glow(s.x, s.y, 14, '#c8ffb4', 0.5); star(s.x, s.y, 6, 0.3, 5, '#d8ffc8', 0); }
      ctx.globalAlpha = 1;
    }
    if (A.on('glowstars')) {
      const k = A.t('glowstars'), a = clamp(k / 0.5, 0, 1) * clamp((5 - k) / 0.6, 0, 1);
      for (let i = 0; i < 16; i++) {
        const ang = i / 16 * TAU + k * 0.5, r = 260 + (i % 3) * 120;
        ctx.globalAlpha = a * 0.8; star(390 + Math.cos(ang) * r, 400 + Math.sin(ang) * r * 0.6, 12, ang, 5, '#fff3a0', 0); glow(390 + Math.cos(ang) * r, 400 + Math.sin(ang) * r * 0.6, 26, '#fff3a0', 0.4);
      }
      ctx.globalAlpha = 1;
    }
    st.draw(0); st.draw(1); st.draw(2);
    if (t > 14) word('The End', 1130, 670, 34 * Math.min(1, (t - 14) / 1), -4, '#ffe9b4');
    st.drawWords();
  }
  drawNightlight(t, dark) {
    const A = this.acts, x = 370, y = 470, w = A.on('robot') ? A.t('robot') : -1;
    glow(x, y - 30, 70, '#8cffc8', 0.25 + 0.1 * Math.sin(t * 2) + dark * 0.3);
    const bob = w >= 0 ? Math.abs(Math.sin(w * 8)) * 6 : 0;
    rrect(x - 16, y - 34 - bob, 32, 34, 6, '#5acdc3', 2.5);
    rrect(x - 18, y - 62 - bob, 36, 28, 7, '#78dcd2', 2.5);
    circ(x - 7, y - 49 - bob, 4, WHITE, 2); circ(x + 7, y - 49 - bob, 4, WHITE, 2);
    circ(x - 7, y - 49 - bob, 1.8, INK, 0); circ(x + 7, y - 49 - bob, 1.8, INK, 0);
    line(x, y - 62 - bob, x, y - 74 - bob, 2, '#aaaabe', 2); circ(x, y - 76 - bob, 4, w >= 0 && Math.sin(w * 20) > 0 ? '#ff5a6e' : '#ffd250', 2);
    if (w >= 0) line(x + 16, y - 26 - bob, x + 28 + Math.sin(w * 14) * 4, y - 46 - bob, 4, '#aaaabe', 2);
  }
  drawMobile(t) {
    const A = this.acts, spin = A.on('mobile') ? t * 2.4 : t * 0.35, x = 700;
    line(x, 0, x, 150, 2, INK, 0);
    rrect(x - 110, 148, 220, 8, 4, '#c8965a', 2.5);
    for (let i = 0; i < 5; i++) {
      const a = spin + i / 5 * TAU, px = x + Math.cos(a) * 100, depth = Math.sin(a);
      const len = 60 + (i % 2) * 30, sway = Math.sin(t * 2 + i) * 4;
      line(px, 152, px + sway, 152 + len, 1.5, INK, 0);
      push(px + sway, 160 + len, 0, 0.8 + depth * 0.2, 0.8 + depth * 0.2);
      drawCandy(0, 0, Math.sin(t + i) * 0.3, 18, CANDY[i], i % 4);
      pop();
    }
  }
  drawBed(t) {
    const A = this.acts, sleepy = A.on('sleep') ? 1 : A.on('yawn') ? 0.3 : this.sleepy();
    // headboard
    rrect(470, 380, 60, 230, 20, '#c8965a'); rrect(940, 470, 40, 140, 14, '#c8965a');
    // pillow and the dino's head on it
    ell(580, 482, 70, 34, '#fffaf0');
    const g = A.on('giggle') ? A.t('giggle') : -1, wig = g >= 0 ? Math.sin(g * 24) * 6 : 0;
    const yawn = A.on('yawn') ? A.t('yawn') : -1;
    const open = yawn > 0.1 && yawn < 1.4 ? Math.sin((yawn - 0.1) / 1.3 * Math.PI) * 1.1 : g >= 0 && g < 1.6 ? 0.5 + 0.3 * Math.sin(g * 20) : 0;
    push(600 + wig * 0.4, 452, -12 + wig);
    drawDinoHead({ t, cols: DINO_GREEN, open, blink: this.blink < 0, sleepy: open > 0.3 ? 0 : sleepy });
    pop();
    // the patchwork quilt
    push(0, Math.sin(t * 1.4) * 1.5 + (g >= 0 ? Math.abs(Math.sin(g * 18)) * -4 : 0));
    ctx.beginPath(); ctx.moveTo(560, 600); ctx.lineTo(560, 500); ctx.quadraticCurveTo(620, 470, 700, 486); ctx.quadraticCurveTo(820, 500, 960, 490); ctx.lineTo(960, 600); ctx.closePath();
    paint('#ff9ab4', 3);
    ctx.save(); ctx.clip();
    const cols = ['#ffb4c8', '#b4dcff', '#fff0a0', '#c8f0b4', '#e0c8ff'];
    for (let r = 0; r < 4; r++) for (let c = 0; c < 9; c++) if ((r + c) % 2) { rect(560 + c * 46, 480 + r * 32, 46, 32, cols[(r * 3 + c) % 5]); }
    ctx.restore();
    rrect(556, 494, 410, 12, 6, '#fffaf0', 2.5);
    pop();
    rect(470, 598, 520, 14, '#a0703c'); line(490, 610, 490, 640, 10, '#8c5a3c'); line(960, 610, 960, 640, 10, '#8c5a3c');
  }
  drawMama(t) {
    const A = this.acts, x = 1060;
    // armchair back
    rrect(960, 360, 200, 250, 50, '#7a8ad2');
    const lean = A.on('kiss') ? Math.sin(Math.PI * clamp(A.t('kiss') / 2.2, 0, 1)) : 0;
    push(x - lean * 120, 560, -lean * 20, -1.25, 1.25);
    drawDino({ t, cols: MAMA_COLS, open: A.on('read') && Math.sin(t * 10) > 0 ? 0.35 : 0, blink: this.mamaBlink < 0, bow: true, walk: 0, arm: 6 });
    pop();
    // the open book in her hands
    const flip = A.on('read') ? clamp(A.t('read') / 0.6, 0, 1) : 1;
    push(x - 70 - lean * 120, 470);
    poly([0, 0, -50, -10, -50, 34, 0, 44], '#fffaf0', 2.5);
    poly([0, 0, 50, -10, 50, 34, 0, 44], '#fffaf0', 2.5);
    if (flip < 1) poly([0, 0, 50 - flip * 100, -10 - Math.sin(flip * Math.PI) * 14, 50 - flip * 100, 34 - Math.sin(flip * Math.PI) * 14, 0, 44], '#fff3dc', 2);
    ctx.globalAlpha = 0.4; for (let i = 0; i < 4; i++) { line(-42, i * 9 + 2, -8, i * 9 + 6, 1.5, INK, 0); line(8, i * 9 + 6, 42, i * 9 + 2, 1.5, INK, 0); } ctx.globalAlpha = 1;
    pop();
    // the chair's arms and seat in front of her
    rrect(950, 520, 230, 90, 30, '#8c9ce0');
    rrect(940, 480, 50, 130, 24, '#7a8ad2'); rrect(1150, 480, 50, 130, 24, '#7a8ad2');
    line(970, 610, 970, 640, 10, '#5a4a6e'); line(1170, 610, 1170, 640, 10, '#5a4a6e');
  }
}
