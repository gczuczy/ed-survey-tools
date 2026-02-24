import { Component } from '@angular/core';

@Component({
  selector:   'app-public-sidemenu-option2',
  standalone: true,
  imports:    [],
  template: `
    <div class="container-fluid">
      <div class="row">
        <div class="col">
          <div class="card">
            <div class="card-body">
              <h2 class="card-title">Public Menu - Option 2</h2>
              <p class="card-text">
                This is a public page accessible from the public side menu.
                No authentication is required.
              </p>
              <p class="text-muted">
                This is a placeholder component. Add your content here.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: []
})
export class PublicSidemenuOption2Component {}
