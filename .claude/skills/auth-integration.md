# Auth Integration — ED Survey Tools

This documents the **actual** auth implementation. The OAuth2 provider is
Frontier Developments (FDev); there is no `.well-known/openid-configuration`.

## Backend endpoints (`pkg/http/api/auth/`)

| Route | Method | Auth | Handler |
|-------|--------|------|---------|
| `/api/auth/config` | GET | public | `configHandler` |
| `/api/auth/callback` | GET | public | `callbackHandler` |
| `/api/auth/user` | GET | required | `userinfoHandler` |
| `/api/auth/logout` | GET | required | `logoutHandler` |
| `/api/auth/me` | DELETE | required | `deleteMeHandler` |

## Config endpoint response (`Config` struct)

```json
{
  "issuer":      "...",
  "clientId":    "...",
  "redirectUri": "https://<host>/api/auth/callback",
  "authUrl":     "...",
  "tokenUrl":    "...",
  "scope":       "auth capi [extra...]"
}
```

`RedirectURI` is built dynamically from the request host. localhost/127.0.0.1
use `http://`, all others use `https://`.

## OAuth2 scopes

Always `auth` + `capi` + any `extraScopes` from config. Both are required:
`auth` for the FDev userinfo endpoint, `capi` for the CMDR profile lookup.

## Callback flow (`callbackHandler`)

1. Extract `code` + `state` from query params.
2. Build `RedirectURL` dynamically from request host.
3. `cfg.Exchange(ctx, code)` → get OAuth2 token.
4. Call FDev userinfo endpoint (`config.UserInfoURL`) → `Userinfo` struct:
   - `Userinfo.User.CustomerID` (int64) — FDev account ID; must be non-zero.
5. Call CAPI: `capi.New(token.AccessToken)` → `capicl.GetProfile()` → CMDR name.
   - Uses a custom HTTP client (`capi.NewHTTPClient()`) that injects User-Agent.
6. `db.Pool.LoginCMDR(cmdrName, customerID)` → upserts CMDR record, returns user.
7. Store `user` in session: `s.Values["user"] = user`.
8. Redirect to `/?code=<code>&state=<state>` — frontend PKCE completes here.

## Key types

```go
type Userinfo struct {
    Issuer   string        `json:"iss"`
    IssuedAt uint64        `json:"iat"`
    Expiry   uint64        `json:"exp"`
    Sub      string        `json:"sub"`
    Scope    string        `json:"scope"`
    User     *UserinfoUser `json:"usr"`
}

type UserinfoUser struct {
    CustomerID int64    `json:"customer_id,string"`
    Email      string   `json:"email"`
    Developer  bool     `json:"developer"`
    Platform   string   `json:"platform"`
    Roles      []string `json:"roles"`
}
```

## Frontend side

The frontend uses `angular-oauth2-oidc` with PKCE (`oidc: false` — no OIDC
discovery). It reads `/api/auth/config` at startup to configure itself.
After the backend callback redirects with `?code=...&state=...`, the frontend
completes the PKCE flow and obtains the access token in sessionStorage.

`AuthService` exposes: `isLoggedIn`, `isConfigured`, `user` (UserInfo with
`isadmin`, `isowner`). Guards: `authGuard` (login required, triggers PKCE
flow), `adminGuard` (login + isAdmin).

## Session / cookies

Session is Redis-backed. Cookie flags: `HttpOnly`, `SameSite=Strict`, `Secure`
(config-driven; `sessions.secure: true` in prod, off by default for dev on
localhost).

## GDPR note

Self-delete (`DELETE /api/auth/me`) nullifies `common.cmdrs.customerid` and
invalidates the Redis session. CMDR name is retained (legitimate interest).
