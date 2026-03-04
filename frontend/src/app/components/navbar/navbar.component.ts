import { Component, OnInit, ViewChild } from '@angular/core';
import { Router, RouterLink, RouterLinkActive }  from '@angular/router';
import { AuthService }                from '../../auth/auth.service';
import { ToolbarModule }              from 'primeng/toolbar';
import { ButtonModule }               from 'primeng/button';
import { MenuModule, Menu }           from 'primeng/menu';
import { MenuItem }                   from 'primeng/api';

@Component({
  selector:    'app-navbar',
  standalone:  true,
  imports: [
    RouterLink,
    RouterLinkActive,
    ToolbarModule,
    ButtonModule,
    MenuModule,
  ],
  templateUrl: './navbar.component.html',
  styleUrl:    './navbar.component.scss',
})
export class NavbarComponent implements OnInit {
  isSidemenuActive   = false;
  isPublicMenuActive = false;
  isVsdsActive       = false;

  sideMenuItems: MenuItem[] = [];
  userMenuItems: MenuItem[] = [];

  @ViewChild('sideMenu') sideMenu!: Menu;
  @ViewChild('vsdsMenu') vsdsMenu!: Menu;
  @ViewChild('userMenu') userMenu!: Menu;

  constructor(
    public authService: AuthService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    const check = () => {
      this.isSidemenuActive   = window.location.pathname.startsWith('/sidemenu');
      this.isPublicMenuActive = window.location.pathname.startsWith('/public-menu');
      this.isVsdsActive       = window.location.pathname.startsWith('/vsds');
    };
    check();
    window.addEventListener('popstate', check);

    this.sideMenuItems = [
      { label: 'Alpha', command: () => this.router.navigate(['/sidemenu/alpha']) },
      { label: 'Beta',  command: () => this.router.navigate(['/sidemenu/beta'])  },
    ];

    this.userMenuItems = [
      { label: 'Settings', icon: 'pi pi-cog', command: () => this.router.navigate(['/settings']) },
      { separator: true },
      { label: 'Logout', icon: 'pi pi-sign-out', command: () => this.logout() },
    ];
  }

  get vsdsMenuItems(): MenuItem[] {
    const items: MenuItem[] = [];
    if (this.authService.user?.isadmin) {
      items.push({ label: 'Folders', icon: 'pi pi-folder', command: () => this.router.navigate(['/vsds/folders']) });
    }
    return items;
  }

  login():  void { this.authService.login(); }
  logout(): void { this.authService.logout(); }
}
