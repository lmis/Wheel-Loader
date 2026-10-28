// ---------- 12. the train ride home ----------
// The late train comes at last: it rattles through the night in front of the
// painted city, which slides slowly past behind it.
class RidePage extends StoryPage {
  constructor() {
    super({ name: 'ride', floor: 720,
      text: 'At last, the train comes! Mama hops on, and it rattles through the sleepy city, all the way home.',
      tricks: [['faster', 'toot', 'shoot'], [], ['wave', 'kiss', 'point']],
      who: ['Night train', '', 'Mama'],
      durs: { faster: 3, toot: 1.6, shoot: 2.4, wave: 2.4, kiss: 2.4, point: 2.6 } });
    this.scroll = 0;
  }
  act(a) {
    const st = this.st;
    if (a === 'faster') { sfx('chug', 1); st.say('FASTER!', 640, 300, 56, '#ffd25a'); }
    if (a === 'toot') { sfx('choo', 1); st.say('TOOT TOOT!', 900, 330, 56, WHITE); }
    if (a === 'shoot') { sfx('shimmer', 1); st.say('MAKE A WISH!', 640, 200, 44, '#fff3a0'); }
    if (a === 'wave') { sfx('giggle', 0.6, 1.2); st.say('HELLO, CITY!', 600, 380, 46, '#d2b4ff'); }
    if (a === 'kiss') { sfx('smooch', 1); for (let i = 0; i < 5; i++) st.add({ kind: 'heart', layer: 2, x: 610, y: 470, vx: -rnd(120, 220), vy: rnd(-80, -20), rot: 0, size: 12, col: '#ff6082', life: 2 }); }
    if (a === 'point') { sfx('sparkle', 0.8); st.say('ALMOST HOME!', 640, 380, 50, '#ffe08a'); }
  }
  tick(dt, t, A) {
    const fast = A.on('faster') ? Math.sin(Math.PI * clamp(A.t('faster') / 3, 0, 1)) : 0;
    this.scroll += dt * (90 + fast * 260);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-12-night-sky');
    if (bg) drawArtScroll(bg, this.scroll); else rect(0, 0, W, H, grad(0, 0, 0, H, ['#0f1440', '#232c6a', '#3a3a7c']));
    if (A.on('shoot')) { const k = A.t('shoot') / 1.4; if (k < 1) { const x = lerp(100, 900, k), y = lerp(140, 280, k); line(x - 150, y - 26, x, y, 3, 'rgba(255,246,200,0.6)', 0); star(x, y, 12, t * 6, 5, '#fff6c8', 0); glow(x, y, 40, '#fff3a0', 0.6); } }
    st.draw(0);
    // the train, wobbling along, with Mama at a window
    const waving = A.on('wave'), kissing = A.on('kiss'), pointing = A.on('point');
    push(0, Math.sin(t * 7) * 3 + Math.sin(t * 2.3) * 2, Math.sin(t * 3.1) * 0.4);
    drawTrainArt(1240, 772, t, { moving: true, cars: 2, window: { car: 0, draw: (x, y) => { // wheels below the page edge, so no tracks
      const head = rigImg(MAMA_RIG, kissing ? 'head-kiss' : waving || pointing ? 'head-talk' : 'head-happy');
      if (!head || !rigReady(MAMA_RIG)) return;
      const arm = rigImg(MAMA_RIG, 'arm');
      if (waving || pointing) { push(x + 14, y - 2, waving ? -60 + Math.sin(t * 14) * 22 : -20, 0.3, 0.3); ctx.drawImage(arm, -arm.naturalWidth * 0.25, -arm.naturalHeight * 0.2); pop(); }
      push(x, y - 26, 0, 0.3, 0.3); ctx.drawImage(head, -head.naturalWidth / 2, -head.naturalHeight * 0.6); pop();
    } } });
    pop();
    st.draw(1); st.draw(2); st.drawWords();
  }
}
