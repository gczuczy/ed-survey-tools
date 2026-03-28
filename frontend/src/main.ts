import { bootstrapApplication }    from '@angular/platform-browser';
import { provideAnimationsAsync }  from '@angular/platform-browser/animations/async';
import { provideRouter }           from '@angular/router';
import { provideHttpClient,
         withInterceptors }        from '@angular/common/http';
import { providePrimeNG }          from 'primeng/config';
import { definePreset }            from '@primeuix/themes';
import Aura                        from '@primeuix/themes/aura';

import { AppComponent }            from './app/app.component';
import { authInterceptor }         from './app/auth/auth.interceptor';
import { routes }                  from './app/app.routes';
import { provideOAuthService }     from './app/auth/oauth.provider';

const EdstPreset = definePreset(Aura, {
  primitive: {
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", "Noto Sans", "Liberation Sans", Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"',
  },
});

bootstrapApplication(AppComponent, {
  providers: [
    provideHttpClient(withInterceptors([authInterceptor])),
    provideAnimationsAsync(),
    provideRouter(routes),
    provideOAuthService(),
    providePrimeNG({
      theme: {
        preset: EdstPreset,
        options: {
          darkModeSelector: 'system',
        }
      }
    }),
  ]
}).catch(err => console.error(err));
