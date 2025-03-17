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