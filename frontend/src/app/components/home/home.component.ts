import { Component }    from '@angular/core';
import { AuthService }  from '../../auth/auth.service';

@Component({
  selector:   'app-home',
  standalone: true,
  imports:    [],
  template: `
    <div class="container page-wrapper">
      <div class="row justify-content-center">
        <div class="col-lg-8">

          @if (!authService.isConfigured) {
            <div class="alert alert-warning" role="alert">
              <h5 class="alert-heading">⚠️ Authentication Unavailable</h5>
              <p class="mb-0">
                The authentication service could not be initialized. 
                Protected pages will not be accessible until the OAuth configuration is restored.
                Public pages continue to work normally.
              </p>
            </div>
          }

          <div class="card">
            <div class="card-body">
              <h1 class="card-title">Welcome to ED Survey Tools</h1>
              <p class="card-text">
                This is the public home page of the ED Survey Tools.
                No authentication is required to view this page.
              </p>
              
              @if (authService.isLoggedIn) {
                <p class="text-success">✓ You are currently logged in.</p>
              } @else if (authService.isConfigured) {
                <p class="text-muted">You are not logged in. Click "Login" to access protected features.</p>
              }
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: []
})
export class HomeComponent {
  constructor(public authService: AuthService) {}
}
