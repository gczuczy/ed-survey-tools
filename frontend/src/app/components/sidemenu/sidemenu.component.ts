import { Component }                                    from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet }   from '@angular/router';
import { TagModule }                                    from 'primeng/tag';

@Component({
  selector:    'app-sidemenu',
  standalone:  true,
  imports:     [RouterLink, RouterLinkActive, RouterOutlet, TagModule],
  templateUrl: './sidemenu.component.html',
  styleUrl:    './sidemenu.component.scss'
})
export class SidemenuComponent {}
