import { CanActivateFn } from '@angular/router';
import { inject }        from '@angular/core';
import { AuthService }   from './auth.service';

/**
 * Functional route guard for admin-only routes (Angular 15+ style).
 *
 * • If the user is not logged in  → triggers the PKCE login flow.
 * • If the user is logged in but not admin  → blocks navigation.
 * • If the user is logged in and admin      → allows navigation.
 *
 * Usage in route definitions:
 *   canActivate: [adminGuard]
 */
export const adminGuard: CanActivateFn = (_route, _state) => {
  const authService = inject(AuthService);

  if (!authService.isLoggedIn) {
    authService.login();
    return false;
  }

  return authService.user?.isadmin ?? false;
};
