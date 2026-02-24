import { CanActivateFn }  from '@angular/router';
import { inject }         from '@angular/core';
import { AuthService }    from './auth.service';

/**
 * Functional route guard (Angular 15+ style).
 *
 * • If the user is logged in      → allows navigation.
 * • If the user is NOT logged in  → triggers the PKCE login flow.
 *
 * Usage in route definitions:
 *   canActivate: [authGuard]
 */
export const authGuard: CanActivateFn = (_route, state) => {
  const authService = inject(AuthService);

  if (authService.isLoggedIn) {
    return true;
  }

  // Store the target URL and trigger login
  // (authService.login() will store it in sessionStorage)
  authService.login();
  return false;
};
