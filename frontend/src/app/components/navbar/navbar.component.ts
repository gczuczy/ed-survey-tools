import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild, ElementRef } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, NavigationEnd }  from '@angular/router';
import { filter, Subscription }               from 'rxjs';
import { AuthService }                        from '../../auth/auth.service';
import { ToolbarModule }                      from 'primeng/toolbar';
import { ButtonModule }                       from 'primeng/button';
import { MenuModule, Menu }                   from 'primeng/menu';
import { BreadcrumbModule }                   from 'primeng/breadcrumb';
import { MenuItem }                           from 'primeng/api';

@Component({
  selector:    'app-navbar',
  standalone:  true,
  imports: [
    RouterLink,
    RouterLinkActive,
    ToolbarModule,
    ButtonModule,
    MenuModule,
    BreadcrumbModule,
  ],
  templateUrl: './navbar.component.html',
  styleUrl:    './navbar.component.scss',
})
export class NavbarComponent implements OnInit, AfterViewInit, OnDestroy {
  isSidemenuActive   = false;
  isPublicMenuActive = false;
  isVsdsActive       = false;

  sideMenuItems: MenuItem[] = [];
  userMenuItems: MenuItem[] = [];

  breadcrumbItems: MenuItem[] = [];
  homeCrumb: MenuItem = { icon: 'pi pi-home', routerLink: '/' };
  hasBreadcrumb = false;

  @ViewChild('sideMenu') sideMenu!: Menu;
  @ViewChild('vsdsMenu') vsdsMenu!: Menu;
  @ViewChild('userMenu') userMenu!: Menu;

  private navSub?: Subscription;
  private resizeObserver?: ResizeObserver;

  constructor(
    public  authService: AuthService,
    private router:      Router,
    private el:          ElementRef,
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

    this.updateBreadcrumbs();
    this.navSub = this.router.events.pipe(
      filter(e => e instanceof NavigationEnd)
    ).subscribe(() => this.updateBreadcrumbs());
  }

  ngAfterViewInit(): void {
    this.resizeObserver = new ResizeObserver(() => this.syncNavbarHeight());
    this.resizeObserver.observe(this.el.nativeElement);
    this.syncNavbarHeight();
  }

  ngOnDestroy(): void {
    this.navSub?.unsubscribe();
    this.resizeObserver?.disconnect();
  }

  private syncNavbarHeight(): void {
    document.documentElement.style.setProperty(
      '--navbar-height',
      `${this.el.nativeElement.offsetHeight}px`,
    );
  }

  private updateBreadcrumbs(): void {
    const url = this.router.url;
    const crumbs: MenuItem[] = [];

    if (url.startsWith('/vsds/')) {
      crumbs.push({ label: 'VSDS', routerLink: '/vsds' });
      if (url.includes('/vsds/projects')) {
        crumbs.push({ label: 'Projects' });
      } else if (url.includes('/vsds/folders')) {
        crumbs.push({ label: 'Folders' });
      }
    }

    this.breadcrumbItems = crumbs;
    this.hasBreadcrumb = crumbs.length > 0;
  }

  get vsdsMenuItems(): MenuItem[] {
    const items: MenuItem[] = [];
    items.push({ label: 'Projects', icon: 'pi pi-briefcase', command: () => this.router.navigate(['/vsds/projects']) });
    if (this.authService.user?.isadmin) {
      items.push({ label: 'Folders', icon: 'pi pi-folder', command: () => this.router.navigate(['/vsds/folders']) });
    }
    return items;
  }

  login():  void { this.authService.login(); }
  logout(): void { this.authService.logout(); }
}
