import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject }                               from '@angular/core';
import { catchError, throwError }               from 'rxjs';
import { AuthService }                          from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  return next(req).pipe(
    catchError((err) => {
      if (err instanceof HttpErrorResponse &&
          err.status === 401 &&
          auth.isLoggedIn) {
        auth.login();
      }
      return throwError(() => err);
    })
  );
};
