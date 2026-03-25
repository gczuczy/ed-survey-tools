import { Routes }    from '@angular/router';
import { authGuard }  from './auth/auth.guard';
import { adminGuard } from './auth/admin.guard';
import { ownerGuard } from './auth/owner.guard';

import { HomeComponent } from './components/home/home.component';

export const routes: Routes = [
  // ── Public routes ─────────────────────────────────────────────────────────
  { path: '', component: HomeComponent },

  // ── Authenticated top-level routes ────────────────────────────────────────
  {
    path: 'settings',
    loadComponent: () =>
      import('./components/settings/settings.component')
        .then(m => m.SettingsComponent),
    canActivate: [authGuard],
  },

  // ── VSDS section (public section, subsections permission-gated) ───────────
  {
    path: 'vsds',
    loadComponent: () =>
      import('./components/vsds/vsds.component')
        .then(m => m.VsdsComponent),
    children: [
      {
        path: '',
        loadComponent: () =>
          import('./components/vsds/vsds-dashboard.component')
            .then(m => m.VsdsDashboardComponent),
      },
      {
        path: 'projects',
        loadComponent: () =>
          import('./components/vsds/vsds-projects.component')
            .then(m => m.VsdsProjectsComponent),
      },
      {
        path: 'projects/:id',
        loadComponent: () =>
          import('./components/vsds/vsds-projects.component')
            .then(m => m.VsdsProjectsComponent),
      },
      {
        path: 'folders',
        loadComponent: () =>
          import('./components/vsds/vsds-folders.component')
            .then(m => m.VsdsFoldersComponent),
        canActivate: [adminGuard],
      },
      {
        path: 'folders/:id',
        loadComponent: () =>
          import('./components/vsds/vsds-folder-extraction.component')
            .then(m => m.VsdsFolderExtractionComponent),
        canActivate: [adminGuard],
      },
    ]
  },

  // ── Admin section (owner-only) ────────────────────────────────────────────
  {
    path: 'admin',
    loadComponent: () =>
      import('./components/admin/admin.component')
        .then(m => m.AdminComponent),
    canActivate: [ownerGuard],
    children: [
      {
        path: '',
        loadComponent: () =>
          import('./components/admin/admin-dashboard.component')
            .then(m => m.AdminDashboardComponent),
      },
      {
        path: 'cmdrs',
        loadComponent: () =>
          import('./components/admin/admin-cmdrs.component')
            .then(m => m.AdminCmdrsComponent),
      },
    ]
  },

  // ── Catch-all ─────────────────────────────────────────────────────────────
  { path: '**', redirectTo: '' }
];
