import { Routes } from '@angular/router';
import { authGuard }  from './auth/auth.guard';
import { adminGuard } from './auth/admin.guard';

import { HomeComponent } from './components/home/home.component';

export const routes: Routes = [
  // ── Public routes ─────────────────────────────────────────────────────────
  { path: '',       component: HomeComponent },
  {
    path: 'barfoo',
    loadComponent: () => import('./components/barfoo/barfoo.component').then(m => m.BarfooComponent),
  },

  // ── Public side menu (no auth required) ───────────────────────────────────
  {
    path: 'public-menu',
    loadComponent: () => import('./components/public-sidemenu/public-sidemenu.component').then(m => m.PublicSidemenuComponent),
    children: [
      { path: '', redirectTo: 'option1', pathMatch: 'full' },
      {
        path: 'option1',
        loadComponent: () => import('./components/public-sidemenu/public-sidemenu-option1.component').then(m => m.PublicSidemenuOption1Component),
      },
      {
        path: 'option2',
        loadComponent: () => import('./components/public-sidemenu/public-sidemenu-option2.component').then(m => m.PublicSidemenuOption2Component),
      },
    ]
  },

  // ── VSDS section (public section, subsections permission-gated) ───────────
  {
    path: 'vsds',
    loadComponent: () => import('./components/vsds/vsds.component').then(m => m.VsdsComponent),
    children: [
      {
        path: '',
        loadComponent: () => import('./components/vsds/vsds-dashboard.component').then(m => m.VsdsDashboardComponent),
      },
      {
        path: 'folders',
        loadComponent: () => import('./components/vsds/vsds-folders.component').then(m => m.VsdsFoldersComponent),
        canActivate: [adminGuard],
      },
    ]
  },

  // ── Protected routes ──────────────────────────────────────────────────────
  {
    path: 'foobar',
    loadComponent: () => import('./components/foobar/foobar.component').then(m => m.FoobarComponent),
    canActivate: [authGuard],
  },
  {
    path: 'settings',
    loadComponent: () => import('./components/settings/settings.component').then(m => m.SettingsComponent),
    canActivate: [authGuard],
  },

  // ── Protected side menu ───────────────────────────────────────────────────
  {
    path: 'sidemenu',
    loadComponent: () => import('./components/sidemenu/sidemenu.component').then(m => m.SidemenuComponent),
    canActivate: [authGuard],
    children: [
      { path: '', redirectTo: 'alpha', pathMatch: 'full' },
      {
        path: 'alpha',
        loadComponent: () => import('./components/sidemenu/sidemenu-alpha.component').then(m => m.SidemenuAlphaComponent),
      },
      {
        path: 'beta',
        loadComponent: () => import('./components/sidemenu/sidemenu-beta.component').then(m => m.SidemenuBetaComponent),
      },
    ]
  },

  // ── Catch-all ─────────────────────────────────────────────────────────────
  { path: '**', redirectTo: '' }
];
