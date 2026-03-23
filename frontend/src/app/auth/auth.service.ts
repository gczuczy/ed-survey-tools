import { Injectable }                       from '@angular/core';
import { OAuthService, AuthConfig }        from 'angular-oauth2-oidc';
import { Observable, of }                  from 'rxjs';
import { map, switchMap, tap, catchError } from 'rxjs/operators';
import { Router }                         from '@angular/router';
import { ApiService }                     from '../services/api.service';

/**
 * The flat key→value shape that `/api/auth/config` must return.
 *
 * Because the primary OAuth2 provider does not expose
 * `.well-known/openid-configuration`, the backend supplies the
 * authorization and token endpoint URLs directly.
 */
export interface OAuthBackendConfig {
  clientId:            string;
  redirectUri:         string;
  scope:               string;
  issuer:              string;
  authUrl:             string;
  tokenUrl:            string;
}

/**
 * Mirrors the backend `db.User` struct returned by `/api/auth/user`.
 */
export interface UserInfo {
  id:         number;
  name:       string;
  customerid: number | null;
  isowner:    boolean;
  isadmin:    boolean;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private configLoadFailed = false;
  private cachedUser: UserInfo | null = null;

  constructor(
    private oauth  : OAuthService,
    private api    : ApiService,
    private router : Router
  ) {}

  /**
   * Called once by APP_INITIALIZER.
   *
   * Loads the OAuth configuration (needed for the login redirect),
   * then queries `/api/auth/user` to determine whether the user
   * already has a valid backend session.
   *
   * If the OAuth config fails to load, we mark the service as disabled
   * and allow the app to continue loading (public pages will work).
   */
  initialize(): Observable<void> {
    return this.api.getConfig<OAuthBackendConfig>('/api/auth/config').pipe(

      tap((cfg: OAuthBackendConfig) => {
        // Validate required fields
        if (!cfg.clientId || !cfg.authUrl || !cfg.tokenUrl) {
          throw new Error('OAuth config missing required fields: clientId, authUrl, tokenUrl');
        }

        const authConfig: AuthConfig = {
          issuer:        cfg.issuer,
          clientId:      cfg.clientId,
          redirectUri:   cfg.redirectUri  || `${window.location.origin}/api/auth/callback`,
          scope:         cfg.scope        || 'auth',
          responseType:  'code',
          loginUrl:      cfg.authUrl,
          tokenEndpoint: cfg.tokenUrl,
          oidc:          false,
          requireHttps:  false,
          showDebugInformation: false,
        };
        this.oauth.configure(authConfig);
      }),

      // Clean up code/state URL params left by the backend callback redirect
      tap(() => this.cleanupCallbackParams()),

      // Determine login status from the backend session
      switchMap(() => this.fetchUserInfo()),

      tap(() => {
        const target = sessionStorage.getItem('_redirectUri');
        if (target && this.cachedUser) {
          sessionStorage.removeItem('_redirectUri');
          this.router.navigateByUrl(target);
        }
      }),

      map(() => undefined),

      catchError((err) => {
        console.error('[AuthService] OAuth configuration failed:', err);
        console.warn('[AuthService] App will continue in public-only mode');
        this.configLoadFailed = true;
        return of(undefined);
      })
    );
  }

  // ── queries ───────────────────────────────────────────────────────────────

  get isLoggedIn(): boolean {
    return this.cachedUser !== null;
  }

  get user(): UserInfo | null {
    return this.cachedUser;
  }

  get isConfigured(): boolean {
    return !this.configLoadFailed;
  }

  // ── actions ───────────────────────────────────────────────────────────────

  login(): void {
    if (this.configLoadFailed) {
      console.warn('[AuthService] Cannot login: OAuth config failed to load');
      alert('Authentication is currently unavailable. Please contact support.');
      return;
    }

    sessionStorage.setItem('_redirectUri', window.location.pathname);
    this.oauth.initCodeFlow();
  }

  logout(): void {
    this.api.getConfig<void>('/api/auth/logout').pipe(
      catchError(() => of(undefined))
    ).subscribe(() => {
      this.cachedUser = null;
      this.router.navigateByUrl('/');
    });
  }

  // ── private helpers ─────────────────────────────────────────────────────

  private fetchUserInfo(): Observable<void> {
    return this.api.getConfig<UserInfo>('/api/auth/user').pipe(
      tap(user => { this.cachedUser = user; }),
      map(() => undefined),
      catchError(() => {
        this.cachedUser = null;
        return of(undefined);
      })
    );
  }

  private cleanupCallbackParams(): void {
    const url = new URL(window.location.href);
    if (url.searchParams.has('code') || url.searchParams.has('state')) {
      url.searchParams.delete('code');
      url.searchParams.delete('state');
      const cleanUrl = url.pathname + (url.search || '') + (url.hash || '');
      window.history.replaceState({}, '', cleanUrl);
    }
  }
}
