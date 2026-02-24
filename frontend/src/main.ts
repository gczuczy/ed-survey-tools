import { bootstrapApplication } from '@angular/platform-browser';
import { AppComponent }         from './app/app.component';
import { provideRouter }        from '@angular/router';
import { routes }               from './app/app.routes';
import { provideHttpClient }    from '@angular/common/http';
import { provideOAuthService }  from './app/auth/oauth.provider';

bootstrapApplication(AppComponent, {
  providers: [
    provideHttpClient(),
    provideRouter(routes),
    provideOAuthService(),
  ]
}).catch(err => console.error(err));
