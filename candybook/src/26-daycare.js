// ---------- 6. dino daycare ----------
class DaycarePage extends StoryPage {
  constructor() {
    super({ name: 'daycare', floor: 690,
      text: 'At dino daycare, Léo and Louis build towers, paint pictures and sing songs. Then everyone has a little nap.',
      tricks: [['blocks', 'paint', 'music'], ['friends', 'teacher', 'bubbles'], ['jump', 'giggle', 'nap']],
      who: ['Daycare', 'Friends', 'Léo & Louis'],
      durs: { blocks: 3.4, paint: 2.4, music: 3, friends: 3, teacher: 2.4, bubbles: 3, jump: 1.4, giggle: 1.8, nap: 4.5 } });
    this.leo = leo(this.st, { x: 520, y: 660, face: 1 });
    this.louis = louis(this.st, { x: 640, y: 660, face: -1 });
    this.pip = new Puppet(this.st, { x: 860, y: 664, face: -1, scale: 0.58, cols: { body: '#ff9ad2', belly: '#fff0f8', spike: '#b48ce6' } });
    this.bea = new Puppet(this.st, { x: 960, y: 664, face: -1, scale: 0.58, cols: { body: '#ffd25a', belly: '#fff8dc', spike: '#5aaa6e' } });
    this.cast = [this.leo, this.louis, this.pip, this.bea];
    this.splats = [];
  }
  act(a) {
    const st = this.st;
    if (a === 'blocks') sfx('tick', 0.6, 1);
    if (a === 'paint') { sfx('splash', 0.5, 1.6); st.say('SPLAT!', 200, 270, 54, '#ff78a0'); for (let i = 0; i < 6; i++) this.splats.push({ x: 160 + rnd(0, 120), y: 320 + rnd(0, 110), r: rnd(10, 22), c: pick(CANDY) }); if (this.splats.length > 24) this.splats.splice(0, 6); }
    if (a === 'music') st.say('LA LA LA!', 1100, 420, 46, '#ffd25a');
    if (a === 'friends') { this.pip.do('cheer', 3); this.bea.do('cheer', 3); sfx('tune', 0.7, 1.2); }
    if (a === 'teacher') { sfx('tick', 0.8, 0.7); setTimeout(() => sfx('tick', 0.8, 0.7), 180); st.say('STORY TIME!', 360, 300, 50, '#8cd2ff'); }
    if (a === 'bubbles') sfx('blow', 0.4, 1.8);
    if (a === 'jump') { this.leo.do('hop'); setTimeout(() => this.louis.do('hop'), 140); sfx('boing', 0.5, 1.6); }
    if (a === 'giggle') { this.leo.do('laugh'); this.louis.do('laugh'); sfx('giggle', 1, 1.2); }
    if (a === 'nap') { this.leo.do('sleep', 4.4); this.louis.do('sleep', 4.4); sfx('yawn', 0.6, 1.5); st.say('SHHH, NAP TIME', 580, 420, 40, '#c8e6ff'); }
  }
  tick(dt, t, A) {
    const st = this.st;
    if (A.on('music')) for (let i = 0; i < 6; i++) if (this.at('music', i * 0.32)) { if (AU.ctx && AU.ctx.state === 'running') INST.marimba(mtof([72, 74, 76, 79, 81, 84][i]), AU.ctx.currentTime + 0.01, 0.3, 2, AU.fx); st.notes(1080 + i * 18, 520, 1); }
    if (this.at('blocks', 2.4)) { sfx('thump', 0.5, 1.6); st.say('CRASH!', 1080, 330, 56, '#ffbe5a'); this.leo.do('laugh'); this.louis.do('laugh'); }
    if (A.on('bubbles') && this.every('b', 0.12)) st.bubbles(760, 600, 1);
    if (A.on('nap') && this.every('z', 0.9)) { st.zzz(560, 560); st.zzz(680, 560); }
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-06-daycare');
    if (bg) drawArt(bg);
    else {
      drawRoom('#fff0c8', '#ffe2b4', '#a8dcc8', '#8cc8b4', 610);
      // alphabet bunting and a sunny window
      ['A', 'B', 'C', 'D', 'E', 'F', 'G'].forEach((ch, i) => { const x = 420 + i * 70, y = 150 + Math.sin(i * 0.9) * 6; poly([x - 26, y, x + 26, y, x, y + 46], CANDY[i], 2.5); word(ch, x, y + 16, 22, 0, WHITE); });
      drawWindow(940, 150, 220, 170, grad(0, 150, 0, 320, ['#8ccdff', '#e0f2ff']), () => { glow(1120, 190, 60, '#fff0a0', 0.7); circ(1120, 190, 26, '#ffdc5a', 2); cloud(980 + (t * 8) % 200, 240, 0.7); });
    }
    // the easel with paint
    line(150, 600, 210, 280, 8, '#c8965a'); line(310, 600, 250, 280, 8, '#c8965a');
    rrect(140, 300, 180, 150, 6, WHITE);
    for (const s of this.splats) circ(s.x, s.y, s.r, s.c, 0);
    circ(200, 360, 22, '#ffdc5a', 2); for (let i = 0; i < 8; i++) { const a = i / 8 * TAU; line(200 + Math.cos(a) * 28, 360 + Math.sin(a) * 28, 200 + Math.cos(a) * 38, 360 + Math.sin(a) * 38, 4, '#ffdc5a', 0); }
    // the teacher robot
    const tt = A.on('teacher') ? A.t('teacher') : -1;
    push(360, 600, 0);
    rrect(-40, -130, 80, 120, 16, '#ffa3c4'); rrect(-46, -200, 92, 70, 20, '#ffbcd4');
    circ(-16, -170, 9, WHITE, 2); circ(16, -170, 9, WHITE, 2); circ(-16, -170, 4, INK, 0); circ(16, -170, 4, INK, 0);
    ctx.beginPath(); ctx.arc(0, -152, 12, 0.2, Math.PI - 0.2); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.stroke();
    const clap = tt >= 0 && tt < 1 ? Math.abs(Math.sin(tt * 18)) * 14 : 30;
    line(-40, -100, -clap, -70, 7, '#aaaabe'); line(40, -100, clap, -70, 7, '#aaaabe');
    rrect(-30, -10, 24, 14, 5, '#5a5a6e'); rrect(6, -10, 24, 14, 5, '#5a5a6e');
    pop();
    // nap mats
    rrect(450, 662, 260, 18, 9, '#b4dcff');
    // the block tower: builds up, then topples
    const bk = A.on('blocks') ? A.t('blocks') : -1;
    for (let i = 0; i < 7; i++) {
      let x = 1080, y = 650 - i * 36, rot = 0;
      const shown = bk < 0 ? i < 3 : bk > i * 0.3;
      if (!shown) continue;
      if (bk > 2.4) { const k = Math.min(1, (bk - 2.4) * 2); x += k * i * 30; y += k * i * 32; rot = k * i * 18; }
      push(x, y, rot); rrect(-20, -34, 40, 34, 5, CANDY[i]); word('ABCDEFG'[i], 0, -17, 18, 0, WHITE); pop();
    }
    // xylophone
    for (let i = 0; i < 6; i++) rrect(1130 + i * 18, 560 - i * 4, 14, 50 + i * 4, 3, CANDY[i], 2);
    this.cast.forEach(c => c.draw());
    this.finish();
  }
}
