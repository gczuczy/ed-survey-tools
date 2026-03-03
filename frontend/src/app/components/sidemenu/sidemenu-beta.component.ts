import { Component }  from '@angular/core';
import { CardModule } from 'primeng/card';

@Component({
  selector:   'app-sidemenu-beta',
  standalone: true,
  imports:    [CardModule],
  template: `
    <p-card header="Beta" styleClass="content-card">
      <p class="text-muted">
        This is the second placeholder entry in the side-menu section.
      </p>
      <p><em>Replace this content with your Beta feature.</em></p>
    </p-card>
  `,
  styles: [`
    .content-card { margin-top: 0.5rem; }
  `]
})
export class SidemenuBetaComponent {}
