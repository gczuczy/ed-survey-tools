import { Component }                  from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector:   'app-public-sidemenu',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  template: `
    <div class="d-flex">
      <!-- ── Sidebar ──────────────────────────────────────────────────── -->
      <aside class="sidebar">
        <div class="p-3">
          <h5 class="text-muted mb-3">Public Menu</h5>
          <nav class="nav flex-column">
            <a class="nav-link" routerLink="option1" routerLinkActive="active">
              Option 1
            </a>
            <a class="nav-link" routerLink="option2" routerLinkActive="active">
              Option 2
            </a>
          </nav>
        </div>
      </aside>

      <!-- ── Main content area ────────────────────────────────────────── -->
      <main class="flex-grow-1 page-wrapper">
        <router-outlet />
      </main>
    </div>
  `,
  styles: []
})
export class PublicSidemenuComponent {}
