// ---------- 5. meanwhile, Papa on his rollerblades ----------
class PapaPage extends StoryPage {
  constructor(o = {}) {
    super(Object.assign({ name: 'papa', floor: 700,
      text: 'Meanwhile, Papa puts on his helmet and his rollerblades with the glowing red wheels. Into the trolley go Léo and Louis. Wheee!',
      tricks: [['ramp', 'zoom', 'puddle'], ['faster', 'hello', 'giggle'], ['spin', 'wheelie', 'helmet']],
      who: ['Street', 'Léo & Louis', 'Papa'],
      durs: { ramp: 2.6, zoom: 2.6, puddle: 2.4, faster: 2, hello: 2.4, giggle: 1.8, spin: 1.4, wheelie: 2, helmet: 2.2 } }, o));
    this.papa = papa(this.st, { x: 430, y: 650, face: 1, look: { helmet: true, skates: true } });
    this.leo = leo(this.st, { x: 0, y: 0, face: 1, noShadow: true }); this.louis = louis(this.st, { x: 0, y: 0, face: 1, noShadow: true });
    this.cast = [this.papa, this.leo, this.louis];
    this.scroll = 0; this.speed = 320; this.sky = ['#9ad4ff', '#d8f0ff', '#fff4dc'];
  }
  act(a) {
    const st = this.st;
    if (a === 'ramp') sfx('whee', 0.3, 0.8);
    if (a === 'zoom') { sfx('zoom', 1); st.say('ZOOOOM!', 700, 330, 64, '#ff6e6e'); }
    if (a === 'puddle') {}
    if (a === 'faster') { this.leo.do('cheer', 2); this.louis.do('cheer', 2); st.say('FASTER, PAPA!', 640, 380, 52, '#ffd25a'); sfx('giggle', 1, 1.2); }
    if (a === 'hello') { this.leo.do('wave', 2.2); this.louis.do('wave', 2.2); sfx('woof', 0.8, 1.2); }
    if (a === 'giggle') { this.leo.do('laugh'); this.louis.do('laugh'); sfx('giggle', 1); }
    if (a === 'spin') { this.papa.do('spin', 1.2); sfx('whee', 0.7, 1.3); st.say('TA-DA!', 430, 380, 54, '#8ce678'); }
    if (a === 'wheelie') { sfx('whee', 0.6, 0.9); st.say('WHEELIE!', 500, 390, 52, '#ffbe5a'); }
    if (a === 'helmet') { this.papa.do('talk', 1.6); sfx('tick', 0.8, 0.8); setTimeout(() => sfx('tick', 0.8, 0.8), 160); st.say('SAFETY FIRST!', 430, 370, 50, '#ff6e6e'); }
  }
  jumpY() {
    const A = this.acts;
    if (!A.on('ramp')) return 0;
    const k = (A.t('ramp') - 1.1) / 1;
    return k > 0 && k < 1 ? -Math.sin(k * Math.PI) * 150 : 0;
  }
  tick(dt, t, A) {
    const st = this.st;
    const sp = this.speed * (A.on('zoom') ? 1 + Math.sin(Math.PI * A.t('zoom') / 2.6) * 1.5 : 1);
    this.scroll += sp * dt;
    this.papa.x = 430 + Math.sin(t * 0.7) * 50;
    const jy = this.jumpY();
    if (jy === 0 && this.every('tr', 0.03)) for (const x of [this.papa.x - 26, this.papa.x + 10, this.papa.x + 120, this.papa.x + 230]) {
      const b = st.add({ kind: 'trail', layer: 0, x, y: 648, vx: -sp, vy: 0, size: rnd(5, 8), col: '#ff4646', life: 0.5 });
    }
    if (this.at('ramp', 1.1)) { st.say('WHEEEE!', 560, 330, 62, '#8cd2ff'); sfx('whee', 1, 1.1); this.leo.do('cheer'); this.louis.do('cheer'); }
    if (this.at('ramp', 2.1)) { st.dust(this.papa.x + 100, 650, 10); sfx('thump', 0.6, 1.3); this.papa.squash = 0.8; }
    if (this.at('puddle', 0.9)) { st.drops(this.papa.x + 120, 640, 30, 420, '#9ad8ff'); sfx('splash', 1); st.say('SPLASH!', this.papa.x + 120, 430, 58, '#8cd2ff'); this.leo.do('laugh'); this.louis.do('laugh'); }
    if (this.at('hello', 0.2)) st.say('HI, DOGGY!', 900, 420, 46, '#ffd25a');
  }
  drawStreet(t) {
    const s = this.scroll;
    const bg = this.name === 'papa' && art('bg-05-street-morning');
    if (bg) { // the painting scrolls past, every other copy mirrored so the edges meet
      const sc = H / bg.naturalHeight, w = bg.naturalWidth * sc, off = (s * 0.6) % (2 * w);
      for (let i = -1; i < 3; i++) {
        const x = i * w - off;
        if (x > W || x + w < 0) continue;
        ctx.save(); ctx.translate(x + (i & 1 ? w : 0), 0); ctx.scale(i & 1 ? -1 : 1, 1);
        ctx.drawImage(bg, 0, 0, w, H); ctx.restore();
      }
      return;
    }
    rect(0, 0, W, H, grad(0, 0, 0, 560, this.sky));
    for (let i = 0; i < 4; i++) cloud(((i * 360 - s * 0.05) % (W + 300) + W + 300) % (W + 300) - 150, 160 + (i % 2) * 60, 1.1, this.cloudCol || WHITE);
    // houses roll past, slower than the road
    const cols = ['#8cb4e6', '#ffe08a', '#a8dc96', '#f5a08c', '#d2b4f0'];
    for (let i = -1; i < 8; i++) {
      const n = Math.floor(s * 0.5 / 220) + i, x = n * 220 - s * 0.5, h = 160 + (n * 37 % 80);
      const c = cols[((n % 5) + 5) % 5];
      rect(x, 520 - h, 190, h, c); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(x, 520 - h, 190, h);
      poly([x - 10, 520 - h, x + 200, 520 - h, x + 95, 450 - h], shade(c, -0.3));
      for (let w = 0; w < 2; w++) rrect(x + 30 + w * 90, 560 - h, 40, 44, 6, this.litWindows ? '#ffe08a' : '#ddeeff');
      if (this.litWindows) glow(x + 95, 590 - h, 80, '#ffd27a', 0.2);
    }
    // trees and lamp posts nearer, faster
    for (let i = -1; i < 6; i++) {
      const n = Math.floor(s * 0.8 / 300) + i, x = n * 300 - s * 0.8 + 120;
      line(x, 560, x, 470, 14, '#8c5a3c'); circ(x, 440, 50, this.treeCol || '#5aa064', 3); circ(x, 440, 46, this.treeCol || '#5aa064', 0);
    }
    rect(0, 520, W, 60, '#bfd88c');
    // the road, with its dashes rushing by
    rect(0, 570, W, H - 570, '#77748a'); rect(0, 566, W, 8, '#c8bfb3');
    ctx.fillStyle = '#f5e6b4'; for (let x = -((s) % 140); x < W; x += 140) ctx.fillRect(x, 690, 70, 8);
  }
  drawGang(t) {
    const A = this.acts, jy = this.jumpY(), p = this.papa;
    const wheelie = A.on('wheelie') ? Math.sin(Math.PI * clamp(A.t('wheelie') / 2, 0, 1)) : 0;
    push(0, jy);
    // the ramp, rolling in under them
    const tx = p.x + 60;
    drawTrolley(tx, 650, t, [
      () => { this.leo.x = tx + 48; this.leo.y = 616; this.leo.draw(); },
      () => { this.louis.x = tx + 104; this.louis.y = 616; this.louis.draw(); },
    ], this.scroll / 9);
    p.rot = -wheelie * 14;
    p.draw({ reach: !p.is('spin') && !p.is('talk'), skateSpin: this.scroll / 6 });
    if (p.is('talk') && !rigReady(PAPA_RIG)) line(p.x + 30, p.y - 160, p.x + 46, p.y - 196, 7, PAPA_COLS.body);
    pop();
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    this.drawStreet(t);
    // a ramp or a puddle comes along the road
    if (A.on('ramp')) { const x = W + 100 - A.t('ramp') * 700; poly([x, 652, x + 160, 652, x + 160, 600], '#e6b46e'); }
    if (A.on('puddle')) { const x = W + 100 - A.t('puddle') * 760; ell(x, 660, 90, 14, '#7ab8e6', 2.5); }
    if (A.on('hello')) { const x = W + 80 - A.t('hello') * 380; this.drawDog(x, 600, t); }
    st.draw(0);
    this.drawGang(t);
    st.draw(1); st.draw(2); st.drawWords();
  }
  drawDog(x, y, t) {
    const wag = Math.sin(t * 20) * 10;
    line(x + 30, y - 30, x + 50, y - 50 + wag, 6, '#c88c50');
    ell(x, y - 30, 36, 20, '#c88c50');
    for (const lx of [-24, -10, 14, 26]) line(x + lx, y - 14, x + lx, y, 6, '#c88c50');
    circ(x - 36, y - 46, 16, '#c88c50'); ell(x - 40, y - 34, 6, 12, '#8c5a32', 2); circ(x - 44, y - 50, 2.4, INK, 0); circ(x - 52, y - 42, 3, INK, 0);
  }
}
