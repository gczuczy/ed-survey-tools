import { CanActivateFn } from '@angular/router';
import { inject }        from '@angular/core';
import { AuthService }   from './auth.service';

export const ownerGuard: CanActivateFn = (_route, _state) => {
  const authService = inject(AuthService);

  if (!authService.isLoggedIn) {
    authService.login();
    return false;
  }

  return authService.user?.isowner ?? false;
};
