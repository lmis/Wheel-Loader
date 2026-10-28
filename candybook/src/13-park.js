// ---------- page: the park in the golden light ----------
const PARK = { ground: 652, ladderX: 822, top: 430, swingX: 360, pivotY: 392, rope: 168 };
// the slide's chute, as a curve from the platform down to the grass
const slideAt = u => {
  const x0 = 882, y0 = PARK.top, cx = 930, cy = 640, x1 = 1060, y1 = PARK.ground;
  const v = 1 - u;
  return [v * v * x0 + 2 * v * u * cx + u * u * x1, v * v * y0 + 2 * v * u * cy + u * u * y1];
};
class ParkPage {
  constructor() {
    this.name = 'park';
    this.text = 'Goodbye, robots! Little Dino skips home through the park, where the leaves twirl down in the golden light.';
    this.tricks = [['geyser', 'leaves', 'icecream'], ['squirrel', 'pigeons', 'kite']]; this.who = ['Park', 'Park friends', 'Little Dino'];
    this.st = new Stage(); this.st.floor = 664;
    this.acts = new Acts({ geyser: 3.2, leaves: 3.5, icecream: 4, squirrel: 3.4, pigeons: 3, kite: 7 });
    this.pigeons = [520, 590, 660, 1180].map((x, i) => ({ x, ph: rnd(0, 6), face: i % 2 ? 1 : -1 }));
    this.swingA = 0; this.swingV = 0.6; this.swing2 = 0;
    this.leafT = 0;
    this.dino = new Dino(this.st, {
      x: 160, ground: () => PARK.ground, minX: 80, maxX: 1090, tricks: ['slide', 'swing', 'jump', 'spin'],
      host: { trick: (d, t, dt) => this.dinoTrick(d, t, dt), startle: () => this.acts.run('pigeons'), roarAt: 600 },
    });
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    const a = cycle(st, k, this.tricks[k]);
    this.acts.run(a);
    if (a === 'geyser') { sfx('splash', 0.6, 1.3); st.say('WHOOSH!', 640, 260, 56, '#8cd2ff'); }
    if (a === 'leaves') { sfx('gust', 1); this.swingV += 1.2; }
    if (a === 'icecream') { sfx('ding', 0.9, 1.4); setTimeout(() => sfx('ding', 0.9, 1.4), 220); st.say('ICE CREAM!', 1150, 330, 50, '#ff9ac8'); }
    if (a === 'squirrel') { sfx('chitter', 1); }
    if (a === 'pigeons') { sfx('flutter', 1); sfx('coo', 0.8); }
    if (a === 'kite') { sfx('whee', 0.5, 0.8); st.say('UP, UP, UP!', 520, 200, 48, '#ffd25a'); }
  }
  // Little Dino's park tricks: down the slide, and a go on the swing
  dinoTrick(d, t, dt) {
    const st = this.st;
    if (d.trick === 'slide') {
      if (d.stage === 0) { // hop over to the ladder
        if (d.stepHop(dt) || !d.hopping) {
          if (Math.abs(d.x - PARK.ladderX) < 4) { d.stage = 1; d.tt = 0; d.face = 1; sfx('creak', 0.4, 2); }
          else { d.hop(d.x + clamp(PARK.ladderX - d.x, -80, 80), 0, 0.26, 22); sfx('hop', 0.4, 1.2); }
        }
      } else if (d.stage === 1) { // climb
        const p = clamp(t / 1.1, 0, 1);
        d.x = lerp(PARK.ladderX, PARK.ladderX + 34, p); d.y = lerp(PARK.ground, PARK.top, p) + Math.abs(Math.sin(p * 18)) * 4;
        if (Math.floor(p * 6) !== Math.floor(clamp((t - dt) / 1.1, 0, 1) * 6)) sfx('tick', 0.4, 0.7 + p);
        if (p >= 1) { d.stage = 2; d.tt = 0; st.say('WHEE!', 900, 330, 64, '#ff8ca0'); sfx('whee', 1, 1.1); }
      } else if (d.stage === 2) { // slide down
        const u = ease(clamp(t / 0.9, 0, 1)), [x, y] = slideAt(u), [x2, y2] = slideAt(Math.min(1, u + 0.02));
        d.x = x; d.y = y; d.rot = Math.atan2(y2 - y, x2 - x) * 57.3 * 0.7;
        if (t >= 0.9) { d.stage = 3; d.tt = 0; d.rot = 0; d.y = PARK.ground; d.squash = 1; st.dust(d.x, d.y, 8); sfx('thump', 0.5, 1.4); st.sparkles(d.x, d.y - 60, 8, 50); }
      } else return t > 0.5;
      return false;
    }
    if (d.trick === 'swing') {
      if (d.stage === 0) {
        if (d.stepHop(dt) || !d.hopping) {
          if (Math.abs(d.x - PARK.swingX) < 4) { d.stage = 1; d.tt = 0; d.hop(PARK.swingX, PARK.pivotY + PARK.rope, 0.4, 40); }
          else { d.hop(d.x + clamp(PARK.swingX - d.x, -80, 80), 0, 0.26, 22); sfx('hop', 0.4, 1.2); }
        }
      } else if (d.stage === 1) {
        if (d.stepHop(dt, false)) { d.stage = 2; d.tt = 0; sfx('whee', 0.7, 0.9); }
      } else if (d.stage === 2) { // swing, higher and higher, then gently down
        const k = Math.sin(Math.PI * clamp(t / 4, 0, 1)), a = 0.75 * k * Math.sin(t * 2.6);
        this.swingA = a;
        d.face = 1; d.x = PARK.swingX + Math.sin(a) * PARK.rope; d.y = PARK.pivotY + Math.cos(a) * PARK.rope; d.rot = -a * 57.3;
        if (Math.floor(t * 2.6 / Math.PI) !== Math.floor((t - dt) * 2.6 / Math.PI) && k > 0.3) sfx('whee', 0.3, 1 + k * 0.4);
        if (t > 4) { d.rot = 0; d.stage = 3; this.swingA = 0; d.hop(PARK.swingX + 120, PARK.ground, 0.5, 50); }
      } else if (d.stepHop(dt, false)) { st.dust(d.x, d.y, 5); return true; }
      return false;
    }
    return true;
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    const t = st.t;
    // leaves drift down all the time, lots in a gust
    this.leafT -= dt * (A.on('leaves') ? 12 : 1);
    if (this.leafT <= 0) { this.leafT = rnd(0.4, 1.1); st.leaf(rnd(-40, W - 200), rnd(120, 300)); }
    if (A.on('leaves')) for (const b of st.bits) if (b.kind === 'leaf') b.vx += 140 * dt;
    // the empty swing sways
    this.swingV *= Math.pow(0.6, dt);
    this.swing2 = Math.sin(t * 2.4) * (0.06 + this.swingV * 0.25);
    if (A.on('geyser') && Math.random() < dt * 30) st.drops(640, 470 - A.t('geyser') * 0, 1, 380);
    if (A.at('squirrel', 1.5, dt)) { st.say('NUT!', 120, 400, 48, '#c8783c'); sfx('chitter', 0.8, 1.2); }
    if (A.at('pigeons', 0.2, dt)) st.say('COO COO!', 600, 480, 46, '#b4b4dc');
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, 560, grad(0, 0, 0, 560, ['#f3a36b', '#ffc98a', '#ffe7b0']));
    glow(1100, 420, 260, '#fff0b4', 0.6);
    circ(1100, 430, 66, '#ffe9a0', 0);
    // far trees in autumn colours
    const far = ['#e8975a', '#d9784a', '#efb45a'];
    for (let x = -40, i = 0; x < W + 60; x += 90, i++) { const r = 56 + (i * 29 % 30); circ(x, 500 - r * 0.5, r, far[i % 3], 0); }
    rect(0, 510, W, H - 510, grad(0, 510, 0, H, ['#a8c86e', '#86b05a']));
    ctx.beginPath(); ctx.moveTo(0, 512); ctx.lineTo(W, 512); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.stroke();
    // the path
    ctx.beginPath(); ctx.moveTo(0, 610); ctx.quadraticCurveTo(640, 590, W, 615); ctx.lineTo(W, 700); ctx.quadraticCurveTo(640, 680, 0, 702); ctx.closePath(); paint('#ecd4a8', 3);
    this.drawKite(t);
    this.drawTree(t);
    this.drawFountain(t);
    this.drawSwings(t);
    this.drawSlide(t);
    this.drawCart(t);
    this.drawPigeons(t);
    grassTufts([40, 260, 560, 760, 1240], 600, t, '#6e9a4a');
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    st.drawWords();
  }
  drawTree(t) {
    line(130, 600, 140, 250, 40, '#8c5a3c');
    line(135, 330, 230, 290, 14, '#8c5a3c');
    for (const [x, y, r, c] of [[80, 220, 80, '#e6783c'], [180, 190, 90, '#f0a03c'], [250, 250, 70, '#e65a3c'], [120, 290, 66, '#f5c446']]) circ(x, y, r, c, 3);
    for (const [x, y, r, c] of [[80, 220, 76, '#e6783c'], [180, 190, 86, '#f0a03c'], [250, 250, 66, '#e65a3c'], [120, 290, 62, '#f5c446']]) circ(x, y, r, c, 0);
    // the squirrel scampers down for a nut and back up
    const A = this.acts, s = A.on('squirrel') ? A.t('squirrel') : -1;
    let sx = 160, sy = 330, down = 0;
    if (s >= 0) { down = Math.sin(Math.PI * clamp(s / 3.2, 0, 1)); sy = lerp(330, 560, down); sx = 160 + down * 40; }
    push(sx, sy, s >= 0 && s < 1.6 ? 90 : s >= 1.6 ? -90 : 0);
    curve(-10, 4, -34, -10, -26, -34, 12, '#c8783c');
    ell(0, 0, 14, 10, '#c8783c'); circ(12, -6, 8, '#c8783c'); poly([10, -12, 12, -20, 16, -12], '#c8783c', 2);
    circ(15, -8, 1.8, INK, 0);
    if (s > 1.5) circ(20, 0, 5, '#8c5a32', 2);
    pop();
  }
  drawFountain(t) {
    const A = this.acts, x = 640, y = 560;
    const big = A.on('geyser') ? Math.sin(Math.PI * clamp(A.t('geyser') / 3, 0, 1)) : 0;
    const h = 60 + big * 230;
    // the water
    ctx.globalAlpha = 0.75;
    for (const s of [-1, 1]) curve(x, y - 40, x + s * 30, y - 40 - h * 0.9, x + s * (50 + big * 40), y - 20, 6 + big * 4, '#9ad8ff', 0);
    line(x, y - 40, x, y - 40 - h, 7 + big * 6, '#bfe6ff', 0);
    ctx.globalAlpha = 1;
    if (big > 0.3) { glow(x, y - 40 - h, 80, '#ffffff', 0.4); ctx.globalAlpha = big * 0.5; ctx.beginPath(); ctx.arc(x, y - 40, h * 0.6, Math.PI, TAU); ctx.lineWidth = 8; ctx.strokeStyle = '#ff9ab4'; ctx.stroke(); ctx.beginPath(); ctx.arc(x, y - 40, h * 0.6 - 8, Math.PI, TAU); ctx.strokeStyle = '#ffe680'; ctx.stroke(); ctx.beginPath(); ctx.arc(x, y - 40, h * 0.6 - 16, Math.PI, TAU); ctx.strokeStyle = '#9ad8ff'; ctx.stroke(); ctx.globalAlpha = 1; }
    ell(x, y, 90, 22, '#c8c8d7'); ell(x, y - 6, 78, 14, '#9ad8ff', 2);
    rrect(x - 14, y - 50, 28, 40, 6, '#c8c8d7'); ell(x, y - 50, 26, 8, '#b4b4c8');
  }
  drawSwings(t) {
    const px = PARK.swingX, py = PARK.pivotY;
    line(260, 600, 300, py - 6, 10, '#5a82d2'); line(340, 600, 300, py - 6, 10, '#5a82d2');
    line(460, 600, 500, py - 6, 10, '#5a82d2'); line(540, 600, 500, py - 6, 10, '#5a82d2');
    line(290, py, 510, py, 12, '#4a6ec0');
    const onSwing = this.dino.trick === 'swing' && this.dino.stage === 2;
    for (const [x, a] of [[px, onSwing ? this.swingA : this.swing2 * 0.7], [px + 80, -this.swing2]]) {
      const sx = x + Math.sin(a) * PARK.rope, sy = py + Math.cos(a) * PARK.rope;
      line(x - 14, py, sx - 16, sy, 2.5, INK, 0); line(x + 14, py, sx + 16, sy, 2.5, INK, 0);
      push(sx, sy, -a * 57.3); rrect(-22, -2, 44, 10, 4, '#ff8c5a'); pop();
    }
  }
  drawSlide(t) {
    const lx = PARK.ladderX, top = PARK.top;
    line(lx - 12, 640, lx - 4, top - 30, 8, '#ffcd46'); line(lx + 28, 640, lx + 36, top - 30, 8, '#ffcd46');
    for (let y = 620; y > top; y -= 30) line(lx - 9, y, lx + 31, y, 5, '#ffcd46');
    rrect(lx - 6, top - 4, 70, 14, 5, '#5a82d2');
    // the chute
    ctx.beginPath(); for (let u = 0; u <= 1.001; u += 0.05) { const [x, y] = slideAt(u); ctx.lineTo(x, y); }
    ctx.lineWidth = 30; ctx.strokeStyle = INK; ctx.stroke(); ctx.lineWidth = 24; ctx.strokeStyle = '#f05a5a'; ctx.stroke();
    ctx.lineWidth = 6; ctx.strokeStyle = 'rgba(255,255,255,0.35)'; ctx.stroke();
    line(1000, 600, 1010, 650, 8, '#ffcd46');
  }
  drawCart(t) {
    const A = this.acts, x = 1150, y = 600, ic = A.on('icecream') ? A.t('icecream') : -1;
    line(x + 6, y - 90, x + 6, y - 210, 5, '#8c8ca0');
    const sw = Math.sin(t * 1.2) * 3;
    for (let i = 0; i < 6; i++) poly([x + 6, y - 230, x + 6 + Math.cos(Math.PI + i * Math.PI / 6) * 90 + sw, y - 190, x + 6 + Math.cos(Math.PI + (i + 1) * Math.PI / 6) * 90 + sw, y - 190], i % 2 ? WHITE : '#ff9ac8', 2.5);
    rrect(x - 70, y - 100, 150, 80, 12, '#9ad8ff');
    word('ICE CREAM', x + 5, y - 60, 22, 0, '#ff8cc8');
    for (const wx of [x - 40, x + 50]) { circ(wx, y - 14, 16, '#5a5a6e'); circ(wx, y - 14, 6, '#c8c8d7'); }
    // the cone tower grows when the bell rings
    if (ic >= 0) {
      const n = Math.min(4, Math.floor(ic * 3)), pop1 = 1 - ease((ic - 3.4) / 0.5);
      push(x + 5, y - 104, 0, pop1, pop1);
      poly([-16, 0, 16, 0, 0, 40], '#e6b46e');
      for (let i = 0; i < n; i++) { circ(Math.sin(t * 4 + i) * 2, -10 - i * 22, 18, ['#ffb4d2', '#fff0c8', '#b4e6a0', '#c8a0e6'][i]); }
      if (n === 4) circ(0, -10 - 4 * 22 + 6, 7, '#ff3c50');
      pop();
    }
  }
  drawPigeons(t) {
    const A = this.acts, f = A.on('pigeons') ? A.t('pigeons') : -1;
    this.pigeons.forEach((p, i) => {
      let x = p.x, y = 646, flap = null;
      if (f >= 0) { const k = clamp(f / 2.8, 0, 1); x += Math.sin(k * TAU) * 120 * p.face; y -= Math.sin(k * Math.PI) * 180; flap = k < 1 ? f * 26 + i : null; }
      const peck = flap === null ? Math.max(0, Math.sin(t * 3 + p.ph)) * 8 : 0;
      push(x, y, peck * 3 * p.face);
      drawBird(0, 0, 0.9, flap, '#a0a0c8', p.face);
      pop();
    });
  }
  drawKite(t) {
    const A = this.acts, k = A.on('kite') ? A.t('kite') : -1;
    if (k < 0) return;
    const up = Math.sin(Math.PI * clamp(k / 7, 0, 1)), x = 300 + up * 260 + Math.sin(t * 1.3) * 30, y = 520 - up * 380 + Math.sin(t * 2) * 14;
    ctx.beginPath(); ctx.moveTo(230, 290); ctx.quadraticCurveTo((x + 230) / 2, (y + 290) / 2 + 60, x, y + 30); ctx.lineWidth = 1.5; ctx.strokeStyle = INK; ctx.stroke();
    push(x, y, Math.sin(t * 2.4) * 10);
    poly([0, -40, 28, 0, 0, 40, -28, 0], '#ff6e8c');
    line(0, -40, 0, 40, 2, INK, 0); line(-28, 0, 28, 0, 2, INK, 0);
    ctx.beginPath(); for (let i = 0; i <= 8; i++) ctx.lineTo(Math.sin(t * 5 + i) * 10, 40 + i * 12); ctx.lineWidth = 2; ctx.strokeStyle = INK; ctx.stroke();
    for (let i = 1; i < 8; i += 2) poly([Math.sin(t * 5 + i) * 10 - 6, 40 + i * 12, Math.sin(t * 5 + i) * 10 + 6, 40 + i * 12 - 6, Math.sin(t * 5 + i) * 10 + 6, 40 + i * 12 + 6], CANDY[i % 7], 1.5);
    pop();
  }
}
