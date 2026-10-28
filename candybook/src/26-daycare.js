// ---------- 6. dino daycare ----------
// Only painted dinos here: Léo, Louis, three friends and their teacher, in
// the painted playroom. The kids run about on their own.
class DaycarePage extends StoryPage {
  constructor() {
    super({ name: 'daycare', floor: 700,
      text: 'At dino daycare, Léo and Louis play with their friends. They run and jump and giggle, and then everyone has a little nap.',
      tricks: [['chase', 'ring', 'bubbles'], ['jump', 'giggle', 'nap']],
      who: ['Daycare', 'Friends', 'Léo & Louis'],
      durs: { chase: 3.5, ring: 4, bubbles: 3, jump: 1.4, giggle: 1.8, nap: 5 } });
    this.leo = leo(this.st, { x: 520, y: 660, face: 1, scale: 0.8 });
    this.louis = louis(this.st, { x: 640, y: 672, face: -1, scale: 0.8 });
    const kid = (n, x, y, face) => new Puppet(this.st, { x, y, face, scale: 0.8, cols: LEO_COLS, look: { hair: true }, rig: kidRig('kid' + n) });
    this.kids = [this.leo, this.louis, kid(1, 360, 690, 1), kid(2, 820, 664, -1), kid(3, 940, 684, -1)];
    this.teacher = new Puppet(this.st, { x: 1100, y: 650, face: -1, scale: 1.35, cols: PAPA_COLS, rig: TEACHER_RIG });
    this.cast = [...this.kids, this.teacher];
    this.kids.forEach(k => { k.idle = rnd(0.5, 3); });
  }
  run(k, speed = 220) { k.walkTo(rnd(250, 1000), speed); }
  act(a) {
    const st = this.st, ks = this.kids;
    if (a === 'chase') { ks.forEach(k => this.run(k, 380)); sfx('giggle', 1, 1.3); st.say('CATCH ME!', 640, 330, 56, '#ffb45a'); }
    if (a === 'ring') { // everyone runs to the teacher and dances round
      ks.forEach((k, i) => k.walkTo(this.teacher.x - 140 - i * 85, 300));
      this.teacher.do('wave', 2.4); sfx('tune', 0.7, 1.2);
    }
    if (a === 'bubbles') { this.teacher.do('talk', 2.8); sfx('blow', 0.4, 1.8); }
    if (a === 'jump') { ks.forEach((k, i) => setTimeout(() => k.do('hop'), i * 110)); sfx('boing', 0.5, 1.6); }
    if (a === 'giggle') { ks.forEach(k => k.do('laugh', 1.8)); sfx('giggle', 1, 1.2); }
    if (a === 'nap') { ks.forEach(k => { k.tx = null; k.do('sleep', 4.8); }); sfx('yawn', 0.6, 1.5); st.say('SHHH, NAP TIME', 640, 380, 40, '#c8e6ff'); }
  }
  tick(dt, t, A) {
    const st = this.st;
    // the kids wander off now and then, unless it's nap time or a game
    if (!A.on('nap') && !A.on('ring')) for (const k of this.kids) {
      if (k.tx === null && (k.idle -= dt) <= 0) { k.idle = rnd(1.5, 4.5); this.run(k, rnd(150, 240)); }
    }
    if (A.on('chase') && this.every('chase', 0.9)) this.kids.forEach(k => this.run(k, 380));
    if (A.on('ring') && A.t('ring') > 1.6 && this.every('ring', 0.5)) this.kids.forEach((k, i) => { k.face = (Math.floor(t * 2) + i) % 2 ? 1 : -1; k.jump(260); });
    if (A.on('ring') && this.every('n', 0.35)) st.notes(this.teacher.x - 300, 520, 1);
    if (A.on('bubbles') && this.every('b', 0.12)) { const [mx, my] = this.teacher.mouth(); st.bubbles(mx, my, 1); }
    if (A.on('nap') && this.every('z', 0.9)) { const k = pick(this.kids); st.zzz(k.x, k.y - 90); }
    if (A.on('giggle') && this.every('h', 0.4)) { const k = pick(this.kids); st.hearts(k.x, k.y - 90, 1); }
  }
  draw() {
    const st = this.st;
    const bg = art('bg-06-daycare');
    if (bg) drawArt(bg); else drawRoom('#fff0c8', '#ffe2b4', '#a8dcc8', '#8cc8b4', 610);
    st.draw(0);
    // nearer kids drawn last
    [...this.cast].sort((a, b) => a.y - b.y).forEach(c => c.draw());
    this.finish();
  }
}
