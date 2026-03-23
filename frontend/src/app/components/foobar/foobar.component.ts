import { Component }  from '@angular/core';
import { CardModule } from 'primeng/card';
import { TagModule }  from 'primeng/tag';

@Component({
  selector:   'app-foobar',
  standalone: true,
  imports:    [CardModule, TagModule],
  template: `
    <div class="content-center">
      <p-card header="Foobar">
        <p-tag value="Login-protected" icon="pi pi-lock" severity="warn" />
        <p class="text-muted" style="margin-top: 0.75rem">
          This is the <strong>Foobar</strong> page.  You can only see it after
          successfully authenticating via OAuth2 PKCE.
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
export class FoobarComponent {}
