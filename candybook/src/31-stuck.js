// ---------- 11. stuck: the train is late ----------
class StuckPage extends StoryPage {
  constructor() {
    super({ name: 'stuck', floor: 700,
      text: 'Mama is at the station, but the dino train is late. Drip, drop goes the rain. “Oh dear,” says Mama. “Léo and Louis are waiting.”',
      tricks: [['board', 'thunder', 'downpour'], ['cat', 'announce', 'frog'], ['sigh', 'umbrella', 'phone']],
      who: ['Rainy station', 'Station', 'Mama'],
      durs: { board: 2.4, thunder: 2, downpour: 4, cat: 2, announce: 3, frog: 1.6, sigh: 2.4, umbrella: 2, phone: 3 } });
    this.mama = mama(this.st, { x: 620, y: 650, face: -1, look: { bow: true, bag: true } });
    this.cast = [this.mama];
    this.frogPos = 0;
    this.glints = 0;
  }
  act(a) {
    const st = this.st;
    if (a === 'board') sfx('flip', 1);
    if (a === 'thunder') { sfx('thunder', 1); st.shake = 0.4; this.mama.do('surprise', 0.6); }
    if (a === 'downpour') sfx('raindrops', 1, 0.8);
    if (a === 'cat') { sfx('meow', 0.8, 1.2); st.say('MEW!', 960, 560, 40, '#ffb45a'); }
    if (a === 'announce') { sfx('ding', 0.6, 0.8); setTimeout(() => sfx('ding', 0.6, 0.65), 250); st.say('THE TRAIN IS LATE. SORRY!', 640, 250, 38, '#ffe0b4'); }
    if (a === 'frog') { sfx('ribbit', 1); this.frogPos ^= 1; st.say('RIBBIT!', 300, 600, 40, '#8ce678'); }
    if (a === 'sigh') { this.mama.do('sad', 2.4); sfx('yawn', 0.4, 0.8); st.say('OH DEAR...', 620, 380, 48, '#c8d2ff'); }
    if (a === 'umbrella') { this.mama.do('spin', 1.4); sfx('whee', 0.4, 1.2); st.drops(620, 380, 26, 300, '#bfe6ff'); }
    if (a === 'phone') { this.mama.do('talk', 2.8); sfx('ring', 0.6, 1.2); st.say('I’M COMING, MY LOVES!', 620, 360, 42, '#d2b4ff'); st.hearts(620, 440, 3); }
  }
  tick(dt, t, A) {
    if (this.at('board', 1)) this.st.say('DELAYED!', 900, 230, 50, '#ff8c8c');
    if (this.at('thunder', 0)) {}
    if (this.every('drip', 0.5)) this.st.add({ kind: 'drop', layer: 2, x: rnd(500, 740), y: 290, vx: 0, vy: 100, size: 3, col: '#bfe6ff', life: 0.8, g: 600 });
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const flash = A.on('thunder') ? Math.max(0, 1 - A.t('thunder') * 3) * (Math.sin(A.t('thunder') * 40) > 0 ? 1 : 0.4) : 0;
    const bg = art('bg-11-stuck'); if (bg) drawArt(bg);
    if (!bg) {
      rect(0, 0, W, H, grad(0, 0, 0, H, ['#1a1e40', '#2c3260', '#3a3f6e']));
      drawCity(t, 520, '#262a52', 0.4, 4);
    }
    // a few twinkles in the sky... could those be robots?
    if (false) for (let i = 0; i < 3; i++) { const k = clamp((t - 8) / 20, 0, 1); glow(200 + i * 40 + k * 300, 200 - k * 40 + Math.sin(t * 3 + i) * 6, 14, '#a0ffd2', 0.5 * k); }
    // the shelter roof
    if (!bg) {
      rect(380, 270, 540, 20, '#4a5080'); line(400, 290, 400, 640, 10, '#4a5080'); line(900, 290, 900, 640, 10, '#4a5080');
    }
    // the departures board
    if (!bg) {
      const flipping = A.on('board') && A.t('board') < 1;
      rrect(780, 310, 240, 90, 8, '#202030');
      const txt = flipping ? 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('').sort(() => Math.random() - 0.5).slice(0, 7).join('') : 'DELAYED';
      ctx.font = '700 30px Fredoka, system-ui, sans-serif'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.fillStyle = '#ffd25a'; ctx.fillText('TRAIN HOME', 900, 336); ctx.fillStyle = flipping ? '#ffffff' : '#ff6e6e'; ctx.fillText(txt, 900, 374);
    }
    // rails and platform, wet and shiny
    if (!bg) {
      drawRails(600);
      rect(0, 620, W, H - 620, grad(0, 620, 0, H, ['#5a5e80', '#484c6e'])); rect(0, 616, W, 8, '#ffd25a');
      ctx.globalAlpha = 0.25; for (let i = 0; i < 6; i++) ell(100 + i * 220, 670 + (i % 2) * 20, 70, 8, '#a0b4ff', 0); ctx.globalAlpha = 1;
    }
    // the bench, with a cat sheltering underneath
    if (!bg) {
      rrect(860, 600, 180, 14, 5, '#8c6450'); line(880, 614, 880, 650, 6, '#3c3c50'); line(1020, 614, 1020, 650, 6, '#3c3c50');
    }
    const cat = A.on('cat') ? A.t('cat') : -1;
    if (!paintedCat(960, 656, cat < 0 ? 'sleepy' : cat < 1.2 ? 'meow' : 'awake')) {
    ell(960, 640, 26, 14, '#ff9a46'); circ(985, 628 - (cat >= 0 ? 6 : 0), 11, '#ff9a46');
    poly([978, 620 - (cat >= 0 ? 6 : 0), 980, 608, 986, 618], '#ff9a46', 2); poly([988, 618, 994, 608, 994, 622], '#ff9a46', 2);
    if (cat >= 0) { circ(982, 627, 1.8, INK, 0); circ(990, 627, 1.8, INK, 0); } else { line(979, 627, 984, 627, 2, INK, 0); line(987, 627, 992, 627, 2, INK, 0); }
    }
    // the frog in the puddle
    const fx = this.frogPos ? 360 : 260;
    if (!paintedPuddle(310, 676, 220)) ell(310, 676, 80, 10, '#6a7ab4', 0);
    const fj = A.on('frog') ? Math.sin(Math.PI * clamp(A.t('frog') / 0.6, 0, 1)) * 50 : 0;
    const fr = A.on('frog') ? A.t('frog') : -1;
    if (!paintedFrog(fx, 676 - fj, fr >= 0 && fr < 0.6 ? 'jump' : fr >= 0.6 && fr < 1.6 ? 'croak' : 'sit')) ell(fx, 664 - fj, 14, 10, '#8ce678'); circ(fx - 6, 654 - fj, 5, '#8ce678'); circ(fx + 6, 654 - fj, 5, '#8ce678'); circ(fx - 6, 653 - fj, 2, INK, 0); circ(fx + 6, 653 - fj, 2, INK, 0);
    // Mama under her umbrella
    const um = this.mama.is('spin') ? A.t('umbrella') * 12 : 0;
    st.draw(0); st.draw(1);
    this.mama.draw();
    const ux = this.mama.x - 20, uy = 410 - (this.mama.is('surprise') ? 20 : 0);
    if (art('mama-umbrella')) { // her painted umbrella, held up over her head
      pic(art('mama-umbrella'), ux + 10, 560, 260, true, Math.sin(um) * 0.18);
    } else {
    line(ux, uy, ux + 10, 560, 4, INK, 0);
    push(ux, uy, Math.sin(um) * 10);
    ctx.beginPath(); ctx.arc(0, 0, 110, Math.PI, TAU); for (let i = 0; i <= 6; i++) { const x = 110 - i * 220 / 6; ctx.quadraticCurveTo(x + 18, 14, x, 0); } ctx.closePath();
    paint('#ff8cb4', 3);
    for (let i = 1; i < 6; i++) line(0, -110 + 0, -110 + i * 220 / 6, 0, 2, '#e66a96', 0);
    pop();
    }
    if (!bg || A.on('downpour')) drawRain(t, bg ? 1.2 : A.on('downpour') ? 2.2 : 1.1); // the painting already rains
    if (flash) { ctx.globalAlpha = flash * 0.5; rect(0, 0, W, H, '#ffffff'); ctx.globalAlpha = 1; }
    st.draw(2); st.drawWords();
  }
}
