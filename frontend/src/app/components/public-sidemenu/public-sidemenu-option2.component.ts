import { Component }  from '@angular/core';
import { CardModule } from 'primeng/card';

@Component({
  selector:   'app-public-sidemenu-option2',
  standalone: true,
  imports:    [CardModule],
  template: `
    <p-card header="Public Menu - Option 2">
      <p>
        This is a public page accessible from the public side menu.
        No authentication is required.
      </p>
      <p class="text-muted">
        This is a placeholder component. Add your content here.
      </p>
    </p-card>
  `,
})
export class PublicSidemenuOption2Component {}
