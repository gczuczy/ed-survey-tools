import { Injectable, OnDestroy } from '@angular/core';

/**
 * Tracks the OS-level `prefers-color-scheme` media query and exposes
 * the current theme string for display purposes.
 *
 * PrimeNG handles the actual dark/light CSS switching automatically
 * via `darkModeSelector: 'system'` (generates @media rules).
 * This service only provides a readable `currentTheme` property.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService implements OnDestroy {

  private mediaQuery: MediaQueryList;
  private _currentTheme: 'light' | 'dark';

  constructor() {
    this.mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    this._currentTheme = this.mediaQuery.matches ? 'dark' : 'light';
    this.mediaQuery.addEventListener('change', this.onChange);
  }

  /** Returns the currently active theme string. */
  get currentTheme(): 'light' | 'dark' {
    return this._currentTheme;
  }

  private onChange = (e: MediaQueryListEvent) => {
    this._currentTheme = e.matches ? 'dark' : 'light';
  };

  ngOnDestroy(): void {
    this.mediaQuery.removeEventListener('change', this.onChange);
  }
}
