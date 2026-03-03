import { Component }    from '@angular/core';
import { AuthService }  from '../../auth/auth.service';
import { ThemeService } from '../../auth/theme.service';
import { CardModule }   from 'primeng/card';
import { TagModule }    from 'primeng/tag';
import { DividerModule } from 'primeng/divider';

@Component({
  selector:   'app-settings',
  standalone: true,
  imports:    [CardModule, TagModule, DividerModule],
  template: `
    <div class="content-center">
      <p-card header="Settings">
        <p-tag value="Login-protected" icon="pi pi-lock" severity="warn" />

        <!-- Theme indicator -->
        <div style="margin-top: 1rem">
          <label class="label-bold">Current theme</label>
          <p class="text-muted">
            <p-tag [value]="themeService.currentTheme" [severity]="themeSeverity" />
            <small style="margin-left: 0.5rem">(auto-detected from OS preference)</small>
          </p>
        </div>

        <p-divider />

        <h4>General</h4>
        <p class="text-muted"><em>Placeholder – add your settings here.</em></p>

        <h4 style="margin-top: 1rem">Notifications</h4>
        <p class="text-muted"><em>Placeholder – notification preferences.</em></p>
      </p-card>
    </div>
  `,
  styles: [`
    .content-center {
      max-width: 800px;
      margin: 1rem auto 0;
    }

    .label-bold {
      font-weight: 600;
      display: block;
      margin-bottom: 0.25rem;
    }
  `]
})
export class SettingsComponent {
  constructor(
    public authService: AuthService,
    public themeService: ThemeService,
  ) {}

  get themeSeverity(): 'secondary' | 'info' {
    return this.themeService.currentTheme === 'dark' ? 'secondary' : 'info';
  }
}
