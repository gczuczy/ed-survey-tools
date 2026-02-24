import { Routes } from '@angular/router';
import { authGuard } from './auth/auth.guard';

import { HomeComponent }     from './components/home/home.component';
import { BarfooComponent }   from './components/barfoo/barfoo.component';
import { FoobarComponent }   from './components/foobar/foobar.component';
import { SettingsComponent } from './components/settings/settings.component';

import { SidemenuComponent }      from './components/sidemenu/sidemenu.component';
import { SidemenuAlphaComponent } from './components/sidemenu/sidemenu-alpha.component';
import { SidemenuBetaComponent }  from './components/sidemenu/sidemenu-beta.component';

import { PublicSidemenuComponent }        from './components/public-sidemenu/public-sidemenu.component';
import { PublicSidemenuOption1Component } from './components/public-sidemenu/public-sidemenu-option1.component';
import { PublicSidemenuOption2Component } from './components/public-sidemenu/public-sidemenu-option2.component';

export const routes: Routes = [
  // ── Public routes ─────────────────────────────────────────────────────────
  { path: '',       component: HomeComponent },
  { path: 'barfoo', component: BarfooComponent },

  // ── Public side menu (no auth required) ───────────────────────────────────
  {
    path: 'public-menu',
    component: PublicSidemenuComponent,
    children: [
      { path: '',        redirectTo: 'option1', pathMatch: 'full' },
      { path: 'option1', component: PublicSidemenuOption1Component },
      { path: 'option2', component: PublicSidemenuOption2Component },
    ]
  },

  // ── Protected routes ──────────────────────────────────────────────────────
  { path: 'foobar',   component: FoobarComponent,   canActivate: [authGuard] },
  { path: 'settings', component: SettingsComponent, canActivate: [authGuard] },

  // ── Protected side menu ───────────────────────────────────────────────────
  {
    path: 'sidemenu',
    component: SidemenuComponent,
    canActivate: [authGuard],
    children: [
      { path: '',      redirectTo: 'alpha', pathMatch: 'full' },
      { path: 'alpha', component: SidemenuAlphaComponent },
      { path: 'beta',  component: SidemenuBetaComponent },
    ]
  },

  // ── Catch-all ─────────────────────────────────────────────────────────────
  { path: '**', redirectTo: '' }
];
