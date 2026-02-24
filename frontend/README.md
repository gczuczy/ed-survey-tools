# ED Survey Tools

Angular 20 Single-Page Application with OAuth2 PKCE authentication, Bootstrap 5 UI, and a reusable API service architecture.

## Features

- **Angular 20** (LTS) — latest stable framework
- **OAuth2 PKCE Authentication** — secure authorization code flow with Proof Key for Code Exchange
- **Bootstrap 5.3** — responsive UI with automatic dark/light mode detection
- **ng-bootstrap 19** — native Angular components for Bootstrap
- **Reusable API Service** — extendable base service for all HTTP operations
- **Protected Routes** — automatic redirect to login with post-auth navigation back to original destination
- **Standalone Components** — modern Angular architecture without NgModules

## Prerequisites

- **Node.js** v20 or higher (Angular 20 requirement)
- **npm** v10 or higher

Verify your versions:
```bash
node -v   # Should be >= v20
npm -v    # Should be >= v10
```

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

This will install all required packages including:
- Angular 20 framework and CLI
- angular-oauth2-oidc for OAuth2/PKCE
- Bootstrap 5 and ng-bootstrap
- TypeScript 5.8

### 2. Configure Backend

The application requires a backend to handle OAuth2 authentication callbacks and maintain server-side sessions.

**Required Backend Endpoints:**
- `GET /api/oauth/config` - Provides OAuth configuration to frontend
- `GET /api/auth/callback` - Handles OAuth callback from IdP, creates session, redirects to frontend

**📖 See [BACKEND_INTEGRATION.md](BACKEND_INTEGRATION.md) for complete implementation guide with code examples**

#### Quick Backend Config

The `/api/oauth/config` endpoint should return:

```json
{
  "clientId": "your-spa-client-id",
  "issuer": "https://your-idp.com/",
  "redirectUri": "http://localhost:4200/api/auth/callback",
  "scope": "openid profile email"
}
```

The `/api/auth/callback` endpoint must:
1. Receive authorization code from IdP
2. Exchange code for tokens
3. Extract user identity and create server-side session
4. Redirect back to `/?code={code}&state={state}` so frontend can complete PKCE flow

The development server (port 4200) proxies all `/api/*` requests to `http://localhost:8081` — configure this in `proxy.conf.json` if your backend runs elsewhere.

### 3. Run Development Server

```bash
npm start
# or
make devserver
```

The app will open at `http://localhost:4200`.

### 4. Build for Production

```bash
npm run build
# or
make build
```

Output lands in `dist/ed-survey-tools/browser/`.

## Makefile Targets

A `GNUmakefile` is provided for convenience:

- **`make devserver`** — Install deps (if needed) and start dev server
- **`make build`** — Production AOT build
- **`make clean`** — Remove node_modules, dist, and Angular caches

## Project Structure

```
ed-survey-tools/
├── src/
│   ├── app/
│   │   ├── auth/
│   │   │   ├── auth.service.ts       # OAuth2 PKCE service
│   │   │   ├── auth.guard.ts         # Route guard for protected pages
│   │   │   ├── oauth.provider.ts     # OAuth DI configuration
│   │   │   └── theme.service.ts      # Dark/light mode detector
│   │   ├── services/
│   │   │   └── api.service.ts        # Generic HTTP client (extendable)
│   │   ├── components/
│   │   │   ├── navbar/               # Top navigation bar
│   │   │   ├── home/                 # Public landing page
│   │   │   ├── barfoo/               # Public menu item
│   │   │   ├── foobar/               # Login-protected menu item
│   │   │   ├── sidemenu/             # Login-protected sidebar layout
│   │   │   └── settings/             # Login-protected settings page
│   │   ├── app.component.ts          # Root component
│   │   └── app.routes.ts             # Route definitions
│   ├── styles.scss                   # Global styles + Bootstrap import
│   ├── main.ts                       # Application entry point
│   └── index.html                    # HTML shell
├── angular.json                      # Angular CLI configuration
├── package.json                      # Dependencies
├── proxy.conf.json                   # Dev server proxy (/api → :8081)
├── tsconfig.json                     # TypeScript configuration
└── Makefile                          # Build automation
```

## Menu Structure

- **Barfoo** (public) — No authentication required
- **Foobar** (protected) — Requires login
- **Side Menu** (protected dropdown) — Nested routes with sidebar:
  - Alpha — Placeholder page
  - Beta — Placeholder page
- **Settings** (protected dropdown, right-aligned) — User settings

## Authentication Flow

1. User visits a protected route (e.g. `/foobar`)
2. `authGuard` intercepts, stores the target URL in sessionStorage
3. Browser redirects to IdP authorization endpoint (PKCE code challenge sent)
4. User authenticates at the IdP
5. IdP redirects back to the SPA with an authorization code
6. `AuthService.initialize()` exchanges the code for tokens (PKCE code verifier)
7. User is navigated to the originally-requested URL

## Extending the API Service

The `ApiService` provides generic typed HTTP methods. To add a new resource:

**Option 1: Direct use**
```typescript
constructor(private api: ApiService) {}

getUsers(): Observable<User[]> {
  return this.api.get<User[]>('/api/users');
}
```

**Option 2: Domain service**
```typescript
@Injectable({ providedIn: 'root' })
export class UserService {
  constructor(private api: ApiService) {}
  
  list()   = this.api.get<User[]>('/api/users');
  create() = this.api.post<User>('/api/users', userData);
}
```

## Dark/Light Mode

The `ThemeService` listens to `prefers-color-scheme` and sets `data-bs-theme` on the `<html>` element. Bootstrap 5.3+ automatically switches its CSS variables in response. No manual theme-switching code is needed — it just works.

## Troubleshooting

### `ERESOLVE unable to resolve dependency tree`

This usually means peer dependency conflicts. The versions in `package.json` have been carefully selected to work together:
- Angular 20
- TypeScript 5.8
- ng-bootstrap 19 (targets Angular 20)
- angular-oauth2-oidc 20
- Bootstrap 5.3
- zone.js 0.15

If you see this error, run:
```bash
npm install --legacy-peer-deps
```

### Build fails with "Cannot find module"

Ensure you're using Node.js 20+ and TypeScript 5.8:
```bash
node -v
npx tsc -v
```

If TypeScript is outdated, upgrade:
```bash
npm install -D typescript@~5.8.0
```

### OAuth errors in browser console

Check that:
1. Your backend serves `/api/oauth/config` with valid OAuth2 parameters
2. The `redirectUri` in the config matches the URL Angular is running on
3. The OAuth2 client is configured at your IdP to allow PKCE (no client secret required)

## License

MIT

## Built With

- [Angular](https://angular.dev/) — The web framework
- [Bootstrap](https://getbootstrap.com/) — UI toolkit
- [ng-bootstrap](https://ng-bootstrap.github.io/) — Angular Bootstrap components
- [angular-oauth2-oidc](https://github.com/manfredsteyer/angular-oauth2-oidc) — OAuth2/OIDC library
