import { Component }              from '@angular/core';
import { RouterOutlet }           from '@angular/router';
import { NavbarComponent }        from './components/navbar/navbar.component';
import { CookieBannerComponent }  from
  './components/cookie-banner/cookie-banner.component';
import { PrivacyNoticeComponent } from
  './components/privacy-notice/privacy-notice.component';

@Component({
  selector:    'app-root',
  standalone:  true,
  imports:     [
    RouterOutlet,
    NavbarComponent,
    CookieBannerComponent,
    PrivacyNoticeComponent,
  ],
  templateUrl: './app.component.html'
})
export class AppComponent {
  privacyVisible = false;
}
