import { Injectable, OnDestroy } from '@angular/core';
import { Subscription }         from 'rxjs';

/**
 * Listens to the OS-level `prefers-color-scheme` media query and sets
 * `data-bs-theme` on the <html> element accordingly.  Bootstrap 5.3+
 * consumes this attribute to switch its CSS custom-property palette
 * between light and dark without reloading any stylesheet.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService implements OnDestroy {

  private mediaQuery: MediaQueryList;
  private subscription: Subscription | undefined;

  constructor() {
    this.mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    this.applyTheme(this.mediaQuery.matches);

    // addEventListener / removeEventListener style works across all browsers
    this.mediaQuery.addEventListener('change', this.onchange);
  }

  /** Returns the currently active Bootstrap theme string. */
  get currentTheme(): 'light' | 'dark' {
    return document.documentElement.getAttribute('data-bs-theme') === 'dark'
      ? 'dark'
      : 'light';
  }

  private onchange = (e: MediaQueryListEvent) => {
    this.applyTheme(e.matches);
  };

  private applyTheme(dark: boolean): void {
    document.documentElement.setAttribute('data-bs-theme', dark ? 'dark' : 'light');
  }

  ngOnDestroy(): void {
    this.mediaQuery.removeEventListener('change', this.onchange);
    this.subscription?.unsubscribe();
  }
}
