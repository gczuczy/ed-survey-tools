import { Component }  from '@angular/core';
import { CardModule } from 'primeng/card';

@Component({
  selector:   'app-barfoo',
  standalone: true,
  imports:    [CardModule],
  template: `
    <div class="content-center">
      <p-card header="Barfoo">
        <p class="text-muted">
          This is the <strong>Barfoo</strong> page.  It is publicly accessible —
          no login is required.
        </p>
        <p><em>Placeholder – replace this content with your feature.</em></p>
      </p-card>
    </div>
  `,
  styles: [`
    .content-center {
      max-width: 800px;
      margin: 1rem auto 0;
    }
  `]
})
export class BarfooComponent {}
