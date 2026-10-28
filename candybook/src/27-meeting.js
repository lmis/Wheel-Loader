// ---------- 7. the afternoon meeting ----------
class MeetingPage extends OfficePage {
  constructor() {
    super({ name: 'meeting',
      text: 'All afternoon, Mama helps the robots. Sometimes she looks at the picture of Léo and Louis on her desk, and smiles.',
      tricks: [['slides', 'chart', 'cuckoo'], ['snooze', 'clap', 'burp'], ['photo', 'idea', 'thumbs']],
      who: ['Meeting', 'Robots', 'Mama'],
      durs: { slides: 3, chart: 3, cuckoo: 2, snooze: 4, clap: 2, photo: 3, idea: 2.4, thumbs: 2.2 } });
    this.robots = [new Robot(this, 0, 560), new Robot(this, 1, 820), new Robot(this, 2, 1080)];
    this.mama.x = 230; this.mama.y = 680; this.mama.face = 1;
    this.slide = 0;
  }
  act(a, k) {
    const st = this.st;
    if (a === 'slides') { this.slide = (this.slide + 1) % 4; sfx('click', 1, 0.8); sfx('swoosh', 0.3, 1.6); }
    if (a === 'chart') { sfx('chime', 0.6); }
    if (a === 'cuckoo') { sfx('cuckoo', 1); st.say('CUCKOO!', 1170, 210, 46, '#ffe0b4'); }
    if (a === 'snooze') { sfx('snore', 0.8); st.say('ZZZ...', 560, 300, 46, '#c8e6ff'); }
    if (a === 'clap') { this.robots.forEach(r => { r.surprise = 0.6; r.kick(1); }); st.say('HOORAY!', 820, 280, 54, '#ffd25a'); for (let i = 0; i < 8; i++) setTimeout(() => sfx('tick', 0.5, rnd(0.8, 1.2)), i * 110); }
    if (a === 'burp') { this.crewDo('burp'); }
    if (a === 'photo') { this.mama.do('reach', 2.6); sfx('hug', 0.6); st.say('MY BOYS!', 300, 380, 50, '#ff8cc8'); st.hearts(300, 480, 4); }
    if (a === 'idea') { this.mama.do('surprise', 0.6); sfx('ding', 1, 1.5); st.say('AHA!', 230, 330, 56, '#ffd25a'); }
    if (a === 'thumbs') { this.mama.do('cheer', 1.6); st.say('GREAT JOB, ROBOTS!', 640, 300, 46, '#d2b4ff'); sfx('giggle', 0.6); this.robots.forEach(r => r.kick(1)); }
  }
  tick(dt, t, A) {
    FactoryPage.prototype.updateCrew.call(this, dt);
    this.robots.forEach(r => r.update(dt));
    if (A.on('snooze')) { this.robots[0].blink = -0.05; if (this.every('z', 0.8)) this.st.zzz(580, 330); }
    if (A.on('photo') && this.every('h', 0.5)) this.st.hearts(300, 470, 1);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    drawRoom('#e8e0f5', '#d4caec', '#a89ac8', '#9488b8', 620);
    drawWindow(60, 130, 200, 180, grad(0, 130, 0, 310, ['#ffc88c', '#ffe6c0']), () => { drawCity(t, 320, '#c8a0b4', 0.3, 3); });
    // the clock, with its cuckoo
    const ck = A.on('cuckoo') ? Math.sin(Math.PI * clamp(A.t('cuckoo') / 2, 0, 1)) : 0;
    rrect(1130, 140, 80, 100, 16, '#a06950'); circ(1170, 200, 26, WHITE, 3);
    line(1170, 200, 1170 + Math.sin(t * 0.1) * 16, 200 - Math.cos(t * 0.1) * 16, 3, INK, 0); line(1170, 200, 1170 + Math.sin(3.5) * 12, 200 - Math.cos(3.5) * 12, 4, INK, 0);
    if (ck > 0) drawBird(1170, 150 - ck * 30, 0.7, null, '#ffd25a', -1);
    // the screen, with slides
    rrect(320, 120, 420, 250, 10, '#3c3c50'); rrect(332, 132, 396, 226, 6, '#fffaf0');
    const ch = A.on('chart') ? A.t('chart') : 3;
    if (this.slide === 0) { // a bar chart going up
      for (let i = 0; i < 5; i++) { const h = (40 + i * 30) * springOut(Math.max(0, ch - i * 0.15)); rrect(372 + i * 66, 330 - h, 44, h, 6, CANDY[i]); }
      line(360, 332, 710, 332, 3, INK, 0);
    } else if (this.slide === 1) { // a picture of a robot cat
      circ(530, 250, 60, '#c8c8d7'); poly([480, 210, 490, 160, 520, 200], '#c8c8d7'); poly([580, 210, 570, 160, 540, 200], '#c8c8d7');
      rrect(500, 230, 60, 16, 6, '#8cffc8', 0); word('MEOW.EXE', 530, 340, 22, 0, '#5a5a6e');
    } else if (this.slide === 2) { // Léo and Louis!
      push(470, 320, 0, 0.7, 0.7); drawDinoHead({ t, cols: LEO_COLS, hair: true, open: 0.3 }); pop();
      push(600, 320, 0, -0.7, 0.7); drawDinoHead({ t, cols: LOUIS_COLS, hair: true, open: 0.3 }); pop();
      heart(535, 200, 18, '#ff6082');
    } else { word('LUNCH?', 530, 245, 64, 0, '#ffa03c'); }
    // the long table; robots sit behind it
    this.robots.forEach(r => r.drawBody());
    rrect(420, 500, 820, 30, 12, '#c8965a'); line(460, 530, 460, 620, 12, '#a0703c'); line(1200, 530, 1200, 620, 12, '#a0703c');
    for (const x of [560, 820, 1080]) { rrect(x - 30, 480, 34, 22, 4, '#fff6dc'); rrect(x + 10, 486, 30, 16, 3, WHITE, 1.5); }
    this.robots.forEach(r => r.drawArms());
    // Mama's desk with the photo frame
    rrect(150, 520, 200, 24, 8, '#a0703c'); line(170, 544, 170, 620, 10, '#8c5a3c'); line(330, 544, 330, 620, 10, '#8c5a3c');
    const lift = this.mama.is('reach') ? 40 : 0;
    push(300, 500 - lift, lift ? -8 : 0);
    rrect(-34, -44, 68, 50, 6, '#ffd25a'); rrect(-28, -38, 56, 38, 3, '#cdeeff');
    push(-12, -12, 0, 0.3, 0.3); drawDinoHead({ t, cols: LEO_COLS, hair: true }); pop();
    push(12, -12, 0, -0.3, 0.3); drawDinoHead({ t, cols: LOUIS_COLS, hair: true }); pop();
    pop();
    if (A.on('idea')) { const k = A.t('idea'); glow(230, 400, 70, '#fff3a0', 0.7 * Math.sin(Math.PI * clamp(k / 2.4, 0, 1))); circ(230, 400, 18, '#fff3a0', 3); rrect(222, 418, 16, 10, 3, '#c8c8d7'); }
    st.draw(0); st.draw(1);
    this.mama.draw({ reach: this.mama.is('reach') });
    st.draw(2); st.drawWords();
  }
}
