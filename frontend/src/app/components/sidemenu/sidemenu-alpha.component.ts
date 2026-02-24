import { Component } from '@angular/core';

@Component({
  selector:   'app-sidemenu-alpha',
  standalone: true,
  template: `
    <div class="card p-4 mt-2">
      <h3 class="card-title">Alpha</h3>
      <p class="card-text text-secondary">
        This is the first placeholder entry in the side-menu section.
      </p>
      <p class="card-text">
        <em>Replace this content with your Alpha feature.</em>
      </p>
    </div>
  `
})
export class SidemenuAlphaComponent {}
