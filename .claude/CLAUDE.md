# Basic description

This project has a golang-based backend in its root directory, and its angular-based frontend can be found in the `frontend` directory.

# Common notes

The whole project can be built using the root `GNUmakefile`'s `build` target. This packs the frontend as well and embeds it into the binary. And since it's gnumake prefer using `gmake` instead of `make`.

Programming patterns need to be respected, code shouldn't be duplicated, but reused. Avoid code duplication and aim for patterns.

Do not do any modifications to the git tree without an explicit prompt for it.

Always plan first, examine your offered solution, look for errors, fix them and iterate until you do not find errors. Then build the project, handle the build errors and start over until all errors are eliminated. Never return with a failing build.

When renaming packages, instead of creating a new dir, copying over, etc, use `mv` to rename the directory.

After coming up with a plan re-read this file, and cross-check the validity of the proposed solution. Aggressively try to find mismatches and misalignments and fix them.

# API

The api has a generic structure:

```json
{"status": "status-string",
"message": "textual-message",
"data": {}}
```

On successful requests only the `.status` and `.data` are filled. The `.data` might be omitted if its empty. Success is also reflected from the HTTP status code, and the `.status` field is set to `success`.

On errors, the `.status` field is set to `error`, no `.data` is returned, and there is an extended error message in the `.message` field which should be shown to the user. The HTTP status code is also reflecting the error.

For all the requests which have a body the content type is `application/json`.

All API endpoints which are adding a new resource return the fresly added resource'sdata, except when explicitly noted otherwise.

All API endpoint implementations must be using the wrappers under `pkg/http/wrappers`.

The API code in the backend must be suitable for later adding a swagger to it. Response objects must be globally defined with reasonable naming, reflecting their purpose.

Registered URLs with patterns/variables in it are automatically decoded by the handler, and set in the internal `Request` object. This must be used to extract their parts. These parameters must always be responsibly handled, with proper error handling.

For an example when the parameter in the route is an integer by its regexp pattern, it must be checked as an integer. Care must be taken if it can be negative or not.

# backend

The backend both serves the frontend from the root(`/`), and provides an api from the `/api` path. The actual API call implementations can be found under the `pkg/http/api` package, each endpoint must be in a separate `.go` file.

To build the backend always use `gmake build`.

# Database

The database files are found under the `sql` directory, the implementation is PostgreSQL. The database is split between functional schemas, and files here are named as such using the `$schema_$part.sql` naming, where:

 - `$schema`: the name of the schema
 - `$part` is one of `struct`, `funcs`, `views`. These represent database structure, functions and views respectively.

The database is using 3 basic roles:

 - `edadmin`: responsible for the database creation, admin permissions
 - `edservices`: Only DQL and DML permissions selectively where needed
 - `edviewer`: NOLOGIN role reserved for human database viewers, selective read-only permissiosn only

The database is aimed to be 3NF, with the practicality exception. Most table have views to return application-specific joined datasets. For insertion, functions are prefered which are returning a view-record of the freshly inserted datum.

The service's database code is found under `pkg/db`. For simple operations it is simple queries, however complex DML operations are always in transactions.

Database calls more than a single query must be encapsulated in a transaction.

# frontend

To build just the frontend use the `frontend/GNUmakefile`'s `build` target. Simply use the command `gmake -C frontend/ build` to build this.

The project is angular20 based.

All components must not have any embedded CSS or HTML, all must be in separate files.

There is a API Service class, which is mainly responsible for the API communication. This class must implement highlevel functions and must not expose direct HTTP calls.

The frontend needs to mimic the communicated data types on the API. That is, for the API responses each structure returned under the `.data` field (found in the backend code under `pkg/http/api/`) has to have its own typescript/js structure.

Regarding permissions, there are 3 categories on the UI, correlating how API endpoints work:

 - Public: these are available and visibile always.
 - Authenticated: These require authentication to access, and not visible to unauthenticated users
 - Protected(permission): These are protected with a permission flag (like `isAdmin`, from the userinfo), requiring that permission to be set on the viewing user, and not visible without it.

Sections are to be lazy-loaded whenever possible.

All timestamps must be in ISO8601 format. It should be in human-readable form, so `s/T/ /`, and seconds precision.

## Navbar and sections

The UI top navbar represents sections (except when noted otherwise). These sections are complex parts of the user interface, they all have subsections. When clicking on the section name, the section's landing page must be loaded (detailed later). Next to each navbar item a dropdown must be present, opening the section's subsections.

The dashboard dispalys the available subsections using a reactive carousel. clicking on each brings that subsection up.

There must be a breadcrumb bar as part of the navbar, right below the navitems.

# AI Memory

Backend AI self-notes are located at `memory/`, parse these. Maintain these for later efficiency.

Frontend AI self-notes are located at `frontend/memory/`, parse these. Maintain tehse for alter efficienty.
