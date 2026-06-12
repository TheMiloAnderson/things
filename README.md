# Yet Another Task Management System

I'm not satisfied with the available task management products; I'd use a notebook (do currently, actually), but it would have to be an inconveniently small one if I'm to carry it with me everywhere. Seems like a simple fun project to roll my own e-version (even if it is a bit of a cliche). 

This is set up based loosely on the Getting Things Done system, by David Allen -- but just in the simple way I use it, with no extra bells or whistles.

Not to be confused with [Things](https://culturedcode.com/things/), the GTD app that only works on Mac & iOS.

## Running with Docker

Copy `.env.example` to `.env` and fill in the values, then:

```bash
docker compose up --build
```

The app listens on `http://127.0.0.1:8888`. MySQL runs in a container on the compose network (not published to the host by default).

### Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DBUSER` | yes | MySQL username |
| `DBPASS` | yes | MySQL password |
| `MYSQL_ROOT_PASSWORD` | yes | MySQL root password (container only) |
| `SESSION_KEY` | yes | Cookie session signing key |
| `CSRF_KEY` | yes | CSRF signing key (32+ bytes) |
| `LISTEN_ADDR` | no | HTTP bind address (default `127.0.0.1:8888`) |
| `DBADDR` | no | MySQL host:port (default `127.0.0.1:3306`) |
| `APP_BASE_URL` | no | Base URL for email links |
| `CSRF_TRUSTED_ORIGINS` | no | Comma-separated trusted hostnames |
| `RESEND_API_KEY` | no | Resend API key for email |
| `MAIL_FROM_ADDR` | no | From address for emails |
| `NEW_ACCOUNT_NOTIFICATION_EMAIL` | no | Admin notification on signup |

## Running locally (without Docker)

Requires a local MySQL server with the `tasks` database. Set `DBUSER` and `DBPASS` in a `.env` file, apply schema from `internal/db/tables.sql` and migrations, then:

```bash
go run .
```
