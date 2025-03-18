# Prestress

Prestress is a realtime REST API layer for PostgreSQL database written with Go.
The name Prestress is a wordplay on PostgreSQL and REST. Note that Prestress is
not supposed to induce more stress, but have it managed preemptively for you.

## Status

Prestress is in early development. It's not ready for use in production.

## Features

- **Database as an API**
  Prestress connects to PostgreSQL database and exposes it as REST API. There's
  no additional backend code required. The database is your backend. Prestress
  leans heavily on this concept.

- **Authorization**
  You can use your existing OAuth2 compatible service to authenticate users. To
  authorize usage of data, Prestress makes properties from access token
  associated with a request accessible in the database. Use role permissions,
  table views, and row-level security policies, which are already available
  in latest PostgreSQL, to make authorization the way you need it to be.

- **Data subscriptions**
  Users can subscribe to realtime notifications on changes on tables and views.
  Notifications are sent in Server-Sent Events (SSE) format. The changes are
  authorized the same way as other operations, by utilizing native PostgreSQL
  data authorization methods.

- **Simple filtering and pagination**
  Data retrieval operations have option to filter rows by simple key-value
  combinations. Advanced filtering methods are intentionally not implemented as
  you are expected to provide fitting tables and views for users with database
  schema instead of letting users to compose those views on demand.

- **Migrations**
  Prestress implements a way to execute database migrations, so you don't have
  to worry about implementing it or finding a suitable tool for keeping your
  database schema up to date.

# Configuration

Configuration for Prestress is set using environment variables. The table below
outlines possible variables that may be defined.

| Name (* = Mandatory)             | Possible values                        |
| -------------------------------- | -------------------------------------- |
| PRESTRESS_ENVIRONMENT *          | `development` or `production`          |
| PRESTRESS_DB_CONNECTION_STRING * | [PostgreSQL connection string][1]      |
| PRESTRESS_MIGRATION_DIRECTORY    | Absolute or relative file path         |
| PRESTRESS_ALLOWED_ORIGINS        | [HTTP origins to allow][2]             |
| PRESTRESS_AUTH_DISABLE           | `1`                                    |
| PRESTRESS_AUTH_INTROSPECTION_URL | [OAuth2 introspection URL][3]          |
| PRESTRESS_AUTH_CLIENT_ID         | [OAuth2 client ID for this app][4]     |
| PRESTRESS_AUTH_CLIENT_SECRET     | [OAuth2 client secret for this app][5] |

[1]: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
[2]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Access-Control-Allow-Origin
[3]: https://www.rfc-editor.org/rfc/rfc7662#section-2
[4]: https://www.rfc-editor.org/rfc/rfc6749#section-2.2
[5]: https://www.rfc-editor.org/rfc/rfc6749#section-2.3.1

# API documentation

Generic documentation for Prestress API can be found from file
[openapi.yaml](./openapi.yaml). In GitLab it's automatically rendered as
interactive Swagger UI, so you can view it easily.

# Contributing

- Report vulnerabilities via email the project owner. Do not post them as
  issues.

- Before contributing code, discuss the issue with the project maintainers.

- Read [Developing Prestress](./docs/developing-prestress.md) to get started and
  familiar with developing Prestress.