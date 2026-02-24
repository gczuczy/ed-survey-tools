import { Component } from '@angular/core';

@Component({
  selector:   'app-sidemenu-beta',
  standalone: true,
  template: `
    <div class="card p-4 mt-2">
      <h3 class="card-title">Beta</h3>
      <p class="card-text text-secondary">
        This is the second placeholder entry in the side-menu section.
      </p>
      <p class="card-text">
        <em>Replace this content with your Beta feature.</em>
      </p>
    </div>
  `
})
export class SidemenuBetaComponent {}
