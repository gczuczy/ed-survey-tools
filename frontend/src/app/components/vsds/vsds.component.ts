import { Component }                                  from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { AuthService }                                from '../../auth/auth.service';

@Component({
  selector:    'app-vsds',
  standalone:  true,
  imports:     [RouterLink, RouterLinkActive, RouterOutlet],
  templateUrl: './vsds.component.html',
  styleUrl:    './vsds.component.scss',
})
export class VsdsComponent {
  constructor(public authService: AuthService) {}
}
