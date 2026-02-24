import { Component }                  from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector:   'app-public-sidemenu',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './public-sidemenu.component.html',
  styleUrl: './public-sidemenu.component.scss'
})
export class PublicSidemenuComponent {}
