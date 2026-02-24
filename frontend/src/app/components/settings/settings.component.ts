import { Component }  from '@angular/core';
import { AuthService } from '../../auth/auth.service';
import { ThemeService } from '../../auth/theme.service';

@Component({
  selector:   'app-settings',
  standalone: true,
  template: `
    <div class="row justify-content-center mt-4">
      <div class="col-md-8">
        <div class="card p-4">
          <h2 class="card-title">Settings</h2>
          <span class="badge bg-warning text-dark mb-3" style="width:fit-content">🔒 Login-protected</span>

          <!-- ── Theme indicator ─────────────────────────────────── -->
          <div class="mb-3">
            <label class="form-label fw-semibold">Current theme</label>
            <p class="text-secondary">
              <span class="badge" [class]="themeBadgeClass">{{ themeService.currentTheme }}</span>
              <small class="ms-2">(auto-detected from OS preference)</small>
            </p>
          </div>

          <!-- ── Placeholder settings slots ──────────────────────── -->
          <hr>
          <h5>General</h5>
          <p class="text-secondary"><em>Placeholder – add your settings here.</em></p>

          <h5 class="mt-3">Notifications</h5>
          <p class="text-secondary"><em>Placeholder – notification preferences.</em></p>
        </div>
      </div>
    </div>
  `
})
export class SettingsComponent {
  constructor(
    public authService: AuthService,
    public themeService: ThemeService
  ) {}

  get themeBadgeClass(): string {
    return this.themeService.currentTheme === 'dark'
      ? 'bg-secondary text-white'
      : 'bg-light text-dark border';
  }
}
