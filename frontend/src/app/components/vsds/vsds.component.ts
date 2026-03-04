import { Component, OnInit, OnDestroy }                                  from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet, NavigationEnd } from '@angular/router';
import { filter, Subscription }                                          from 'rxjs';
import { AuthService }                                                   from '../../auth/auth.service';
import { BreadcrumbModule }                                              from 'primeng/breadcrumb';
import { MenuItem }                                                      from 'primeng/api';

@Component({
  selector:    'app-vsds',
  standalone:  true,
  imports:     [RouterLink, RouterLinkActive, RouterOutlet, BreadcrumbModule],
  templateUrl: './vsds.component.html',
  styleUrl:    './vsds.component.scss',
})
export class VsdsComponent implements OnInit, OnDestroy {
  breadcrumbItems: MenuItem[] = [];
  home: MenuItem = { icon: 'pi pi-home', routerLink: '/' };
  hasSubsection = false;

  private navSub?: Subscription;

  constructor(
    public  authService: AuthService,
    private router:      Router,
  ) {}

  ngOnInit(): void {
    this.updateBreadcrumbs();
    this.navSub = this.router.events.pipe(
      filter(e => e instanceof NavigationEnd)
    ).subscribe(() => this.updateBreadcrumbs());
  }

  ngOnDestroy(): void {
    this.navSub?.unsubscribe();
  }

  private updateBreadcrumbs(): void {
    const url = this.router.url;
    const crumbs: MenuItem[] = [{ label: 'VSDS', routerLink: '/vsds' }];
    this.hasSubsection = false;

    if (url.includes('/vsds/folders')) {
      crumbs.push({ label: 'Folders' });
      this.hasSubsection = true;
    }

    this.breadcrumbItems = crumbs;
  }
}
