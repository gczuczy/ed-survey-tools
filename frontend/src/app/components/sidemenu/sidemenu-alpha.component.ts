import { Component }  from '@angular/core';
import { CardModule } from 'primeng/card';

@Component({
  selector:   'app-sidemenu-alpha',
  standalone: true,
  imports:    [CardModule],
  template: `
    <p-card header="Alpha" styleClass="content-card">
      <p class="text-muted">
        This is the first placeholder entry in the side-menu section.
      </p>
      <p><em>Replace this content with your Alpha feature.</em></p>
    </p-card>
  `,
  styles: [`
    .content-card { margin-top: 0.5rem; }
  `]
})
export class SidemenuAlphaComponent {}
