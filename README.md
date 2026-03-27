# ed-survey-tools

A web service for processing and analysing Elite Dangerous survey data
from Google Drive and Google Sheets. It serves an Angular frontend and
a REST API from a single binary.

---

## Requirements

### Google Cloud Platform — Service Account

The service authenticates to Google APIs using a GCP service account.

1. Create a project in the [Google Cloud Console](https://console.cloud.google.com/).
2. Enable the following APIs for your project:
   - [Google Drive API](https://console.cloud.google.com/apis/library/drive.googleapis.com)
   - [Google Sheets API](https://console.cloud.google.com/apis/library/sheets.googleapis.com)
3. Create a service account and download its JSON key file:
   [Creating and managing service account keys](https://cloud.google.com/iam/docs/keys-create-delete)
4. The service account requires the following OAuth2 scopes at runtime
   (no IAM roles needed — scopes are requested per-call):
   - `https://www.googleapis.com/auth/drive.readonly` — file downloads
   - `https://www.googleapis.com/auth/drive.metadata.readonly` — folder/file metadata
   - `https://www.googleapis.com/auth/spreadsheets` — spreadsheet reads

   For Drive folders shared with the service account, the account must
   have at least **Viewer** access on those folders in Google Drive.

### PostgreSQL

A PostgreSQL database is required. Once you have a running instance and
are connected to `template1` (or any maintenance database), run:

```sql
-- Create roles
CREATE ROLE edadmin WITH LOGIN PASSWORD 'changeme';
CREATE ROLE edservice WITH LOGIN PASSWORD 'changeme';
CREATE ROLE edviewer WITH NOLOGIN;

-- Create database
CREATE DATABASE edtools WITH OWNER edadmin;
```

Then import the schema. Connect to the newly created database (`psql -d
edtools`) from within the `sql/` directory and run:

```
\i _all.sql
```

### Redis

Redis is required for session storage. Any Redis instance (local or
remote) is supported. Default connection is `localhost:6379`.

### Frontier Developments OAuth2

Authentication is provided exclusively via the
[Frontier Developments](https://user.frontierstore.net/) OAuth2 provider.
This service is built for Elite Dangerous community tools and is tied to
the Frontier auth infrastructure.

Register an OAuth2 application in the
[Frontier developer zone](https://user.frontierstore.net/) to obtain a
`client_id` and `client_secret`. Set the redirect URI of your application
to `https://<your-host>/api/auth/callback`.

Refer to the [Frontier developer docs](https://user.frontierstore.net/developer/docs)
and [community OAuth2 notes](https://github.com/Athanasius/fd-api/blob/main/docs/FrontierDevelopments-oAuth2-notes.md)
for details on the authorisation flow.

---

## Build Prerequisites

| Tool | Notes |
|------|-------|
| Go 1.25+ | Backend |
| Node.js 22+ / npm | Angular frontend build |
| GNU Make (`gmake`) | Build orchestration |

---

## Building

Build everything (frontend + backend, frontend is embedded into the
binary):

```sh
gmake build
```

The resulting binary is placed at `dist/edst`.

To build only the frontend:

```sh
gmake -C frontend/ build
```

---

## Configuration

The service is configured via a YAML file. The default path is
`~/.edsda.yaml`; override with the `-c` flag.

```yaml
db:
  host: localhost
  # port: 5432        # optional, defaults to driver default
  dbname: edtools
  user: edservice
  password: changeme
  maxconns: 8         # optional, default 8
  minconns: 1         # optional, default 1
  ssl: false          # optional, default false

http:
  port: 8080          # default 80

oauth2:
  # issuer defaults to https://auth.frontierstore.net
  clientid: "your-frontier-client-id"
  clientsecret: "your-frontier-client-secret"
  # authorizeUrl, tokenUrl, userinfoUrl are derived from issuer by default

sessions:
  store: redis
  # Generate a strong random key, e.g.: openssl rand -hex 32
  key: "replace-with-a-long-random-secret"
  redis:
    host: localhost   # optional, default localhost
    port: 6379        # optional, default 6379
    # db: 0           # optional, Redis database index
    # user: ""        # optional
    # pass: ""        # optional
    maxidle: 16       # optional, default 16
    # idletimeout: 5m # optional, default 5m

logging:
  level: info         # debug, info, warn, error
  output: stdio       # stdio or syslog
  timestamp: false
  syslog:             # only used when output: syslog
    host: 127.0.0.1
    port: 514
    proto: udp
    facility: LOCAL0

vsds:
  processorInterval: 1m   # optional, default 1m

edsm:
  timeout: 5s             # optional, default 5s
  retries: 10             # optional, default 10

bundles:
  path: /var/lib/edst/bundles  # directory for .json.gz bundle files
  serve: false                 # true = serve /static/ from this backend
  baseUrl: ""                  # URL prefix for download links
                               # empty = frontend defaults to /static/
                               # e.g. https://static.example.com/
  checkInterval: 5m            # optional, default 5m
```

---

## Running

```sh
./dist/edst -c /path/to/config.yaml -s /path/to/credentials.json
```

| Flag | Default | Description |
|------|---------|-------------|
| `-c` | `~/.edsda.yaml` | Path to the YAML configuration file |
| `-s` | `credentials.json` | Path to the GCP service account credentials JSON |

The service listens on the port configured under `http.port` and serves
the frontend from `/` and the API from `/api`.

Shutdown is handled gracefully on `SIGINT` or `SIGTERM`.

---

## Deployment

### Production static serving

For production, bundle files are best served from a dedicated static host or
subdomain rather than the application backend. Point nginx at the configured
`bundles.path` directory with `gzip_static on;` so pre-compressed `.json.gz`
files are served with the correct `Content-Encoding: gzip` header:

```nginx
location / {
    root /var/lib/edst/bundles;
    gzip_static on;
    default_type application/json;
}
```

Set `bundles.baseUrl` in the application config to the base URL of that host
(e.g. `https://static.example.com/`) so the frontend constructs correct links.
Leave `bundles.serve: false` when using external static serving.
