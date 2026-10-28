// ---------- 8. pick-up at sunset ----------
class PickupPage extends PapaPage {
  constructor() {
    super({ name: 'pickup',
      text: 'When the sun turns orange, Papa zooms back to daycare. “Papa!” shout Léo and Louis, and off they all roll for home.',
      tricks: [['zoom', 'ramp', 'puddle'], ['papa', 'faster', 'giggle'], ['spin', 'wheelie', 'helmet']],
      who: ['Sunset street', 'Léo & Louis', 'Papa'] });
    this.acts.durs.papa = 2.4;
    this.sky = ['#f39a6e', '#ffb98a', '#ffe0a8'];
    this.cloudCol = '#ffd6c8'; this.treeCol = '#4e8a5a'; this.litWindows = true;
    this.speed = 260;
  }
  act(a) {
    if (a === 'papa') { this.leo.do('cheer', 2.2); this.louis.do('cheer', 2.2); this.st.say('PAPA!', 560, 400, 64, '#ffd25a'); this.st.hearts(this.papa.x + 120, 520, 4); sfx('giggle', 1, 1.1); sfx('hug', 0.6); return; }
    super.act(a);
  }
  draw() {
    const st = this.st, t = st.t;
    this.drawStreet(t);
    // the sun going down behind the houses, and the daycare sign rolling away
    const painted = art('bg-08-street-sunset');
    if (!painted) glow(1000, 420, 220, '#ffe08a', 0.4);
    const sx = 900 - Math.min(this.scroll * 0.5, 1400);
    if (sx > -300 && !painted) { rrect(sx, 300, 220, 220, 10, '#ffd2e6'); word('DAYCARE', sx + 110, 330, 30, 0, '#ff6e9a'); rrect(sx + 80, 420, 60, 100, 20, '#5aa06e'); }
    const A = this.acts;
    if (A.on('ramp')) { const x = W + 100 - A.t('ramp') * 700; if (!painted) poly([x, 652, x + 160, 652, x + 160, 600], '#e6b46e'); }
    if (A.on('puddle')) { const x = W + 100 - A.t('puddle') * 760; if (!paintedPuddle(x, 660, 220)) ell(x, 660, 90, 14, '#e6a07a', 2.5); }
    st.draw(0);
    this.drawGang(t);
    st.draw(1); st.draw(2); st.drawWords();
  }
}
