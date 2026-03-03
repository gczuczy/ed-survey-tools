import { Component }    from '@angular/core';
import { AuthService }  from '../../auth/auth.service';
import { CardModule }   from 'primeng/card';
import { MessageModule } from 'primeng/message';

@Component({
  selector:   'app-home',
  standalone: true,
  imports:    [CardModule, MessageModule],
  template: `
    <div class="home-content">
      @if (!authService.isConfigured) {
        <p-message severity="warn" styleClass="auth-warn-msg">
          <strong>Authentication Unavailable</strong><br>
          The authentication service could not be initialized.
          Protected pages will not be accessible until the OAuth configuration is restored.
          Public pages continue to work normally.
        </p-message>
      }

      <p-card header="Welcome to ED Survey Tools">
        <p>
          This is the public home page of the ED Survey Tools.
          No authentication is required to view this page.
        </p>

        @if (authService.isLoggedIn) {
          <p class="text-success">&#x2713; You are currently logged in.</p>
        } @else if (authService.isConfigured) {
          <p class="text-muted">You are not logged in. Click "Login" to access protected features.</p>
        }
      </p-card>
    </div>
  `,
  styles: [`
    .home-content {
      max-width: 800px;
      margin: 0 auto;
    }

    .auth-warn-msg {
      margin-bottom: 1rem;
    }
  `]
})
export class HomeComponent {
  constructor(public authService: AuthService) {}
}
