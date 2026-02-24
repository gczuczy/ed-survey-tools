import { APP_INITIALIZER, EnvironmentProviders, Provider } from '@angular/core';
import { provideOAuthClient }                               from 'angular-oauth2-oidc';
import { AuthService }                                      from './auth.service';
import { ThemeService }                                     from './theme.service';

/**
 * Call this from the root providers[] in main.ts.
 */
export function provideOAuthService(): (EnvironmentProviders | Provider)[] {
  return [
    provideOAuthClient(),

    {
      provide:    APP_INITIALIZER,
      multi:      true,
      useFactory: (auth: AuthService, _theme: ThemeService) => {
        return () => auth.initialize().toPromise();
      },
      deps: [AuthService, ThemeService]
    }
  ];
}
