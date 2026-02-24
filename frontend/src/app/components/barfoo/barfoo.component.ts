import { Component } from '@angular/core';

@Component({
  selector:   'app-barfoo',
  standalone: true,
  template: `
    <div class="row justify-content-center mt-4">
      <div class="col-md-8">
        <div class="card p-4">
          <h2 class="card-title">Barfoo</h2>
          <p class="card-text text-secondary">
            This is the <strong>Barfoo</strong> page.  It is publicly accessible —
            no login is required.
          </p>
          <p class="card-text">
            <em>Placeholder – replace this content with your feature.</em>
          </p>
        </div>
      </div>
    </div>
  `
})
export class BarfooComponent {}
