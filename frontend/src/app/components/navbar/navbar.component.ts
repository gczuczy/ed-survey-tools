import {
  Component, OnInit, OnDestroy, AfterViewInit,
  ViewChild, ElementRef,
} from '@angular/core';
import { Router, RouterLink, NavigationEnd } from '@angular/router';
import { filter, Subscription }    from 'rxjs';
import { BreadcrumbService }       from '../../services/breadcrumb.service';
import { AuthService }             from '../../auth/auth.service';
import { ToolbarModule }           from 'primeng/toolbar';
import { ButtonModule }            from 'primeng/button';
import { MenuModule, Menu }        from 'primeng/menu';
import { BreadcrumbModule }        from 'primeng/breadcrumb';
import { MenuItem }                from 'primeng/api';

@Component({
  selector:    'app-navbar',
  standalone:  true,
  imports: [
    RouterLink,
    ToolbarModule,
    ButtonModule,
    MenuModule,
    BreadcrumbModule,
  ],
  templateUrl: './navbar.component.html',
  styleUrl:    './navbar.component.scss',
})
export class NavbarComponent implements OnInit, AfterViewInit, OnDestroy {
  isVsdsActive  = false;
  isAdminActive = false;

  userMenuItems:  MenuItem[] = [];
  vsdsMenuItems:  MenuItem[] = [];
  adminMenuItems: MenuItem[] = [];

  breadcrumbItems: MenuItem[] = [];
  homeCrumb: MenuItem = { icon: 'pi pi-home', routerLink: '/' };
  hasBreadcrumb = false;

  @ViewChild('vsdsMenu')  vsdsMenu!:  Menu;
  @ViewChild('adminMenu') adminMenu!: Menu;
  @ViewChild('userMenu')  userMenu!:  Menu;

  private dynamicLabel: string | null = null;

  private navSub?:   Subscription;
  private crumbSub?: Subscription;
  private resizeObserver?: ResizeObserver;

  constructor(
    public  authService:       AuthService,
    private router:            Router,
    private el:                ElementRef,
    private breadcrumbService: BreadcrumbService,
  ) {}

  ngOnInit(): void {
    const check = () => {
      this.isVsdsActive  = window.location.pathname.startsWith('/vsds');
      this.isAdminActive = window.location.pathname.startsWith('/admin');
    };
    check();
    window.addEventListener('popstate', check);

    this.userMenuItems = [
      {
        label:   'Settings',
        icon:    'pi pi-cog',
        command: () => this.router.navigate(['/settings']),
      },
      { separator: true },
      {
        label:   'Logout',
        icon:    'pi pi-sign-out',
        command: () => this.logout(),
      },
    ];

    this.buildVsdsMenuItems();
    this.buildAdminMenuItems();
    this.updateBreadcrumbs();

    this.navSub = this.router.events.pipe(
      filter(e => e instanceof NavigationEnd)
    ).subscribe(() => {
      this.isVsdsActive  = this.router.url.startsWith('/vsds');
      this.isAdminActive = this.router.url.startsWith('/admin');
      this.buildVsdsMenuItems();
      this.buildAdminMenuItems();
      this.updateBreadcrumbs();
    });

    this.crumbSub = this.breadcrumbService.dynamicLabel$.subscribe(label => {
      this.dynamicLabel = label;
      this.updateBreadcrumbs();
    });
  }

  ngAfterViewInit(): void {
    this.resizeObserver = new ResizeObserver(() => this.syncNavbarHeight());
    this.resizeObserver.observe(this.el.nativeElement);
    this.syncNavbarHeight();
  }

  ngOnDestroy(): void {
    this.navSub?.unsubscribe();
    this.crumbSub?.unsubscribe();
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
        if (this.dynamicLabel) {
          crumbs.push({ label: 'Projects', routerLink: '/vsds/projects' });
          crumbs.push({ label: this.dynamicLabel });
        } else {
          crumbs.push({ label: 'Projects' });
        }
      } else if (/\/vsds\/folders\/\d+/.test(url)) {
        crumbs.push({ label: 'Folders', routerLink: '/vsds/folders' });
        crumbs.push({ label: this.dynamicLabel ?? '…' });
      } else if (url.includes('/vsds/folders')) {
        crumbs.push({ label: 'Folders' });
      }
    } else if (url.startsWith('/admin/')) {
      crumbs.push({ label: 'Admin', routerLink: '/admin' });
      if (url.includes('/admin/cmdrs')) {
        crumbs.push({ label: 'Commanders' });
      }
    } else if (url.startsWith('/settings')) {
      crumbs.push({ label: 'Settings' });
    }

    this.breadcrumbItems = crumbs;
    this.hasBreadcrumb   = crumbs.length > 0;
  }

  private buildVsdsMenuItems(): void {
    const items: MenuItem[] = [];
    items.push({
      label:   'Projects',
      icon:    'pi pi-briefcase',
      command: () => this.router.navigate(['/vsds/projects']),
    });
    if (this.authService.user?.isadmin) {
      items.push({
        label:   'Folders',
        icon:    'pi pi-folder',
        command: () => this.router.navigate(['/vsds/folders']),
      });
    }
    this.vsdsMenuItems = items;
  }

  private buildAdminMenuItems(): void {
    this.adminMenuItems = [
      {
        label:   'Commanders',
        icon:    'pi pi-users',
        command: () => this.router.navigate(['/admin/cmdrs']),
      },
    ];
  }

  login():  void { this.authService.login(); }
  logout(): void { this.authService.logout(); }
}
