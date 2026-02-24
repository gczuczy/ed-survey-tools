import { Component } from '@angular/core';

@Component({
  selector:   'app-foobar',
  standalone: true,
  template: `
    <div class="row justify-content-center mt-4">
      <div class="col-md-8">
        <div class="card p-4">
          <h2 class="card-title">Foobar</h2>
          <span class="badge bg-warning text-dark mb-2" style="width:fit-content">🔒 Login-protected</span>
          <p class="card-text text-secondary">
            This is the <strong>Foobar</strong> page.  You can only see it after
            successfully authenticating via OAuth2 PKCE.
          </p>
          <p class="card-text">
            <em>Placeholder – replace this content with your feature.</em>
          </p>
        </div>
      </div>
    </div>
  `
})
export class FoobarComponent {}
