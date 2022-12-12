# ShowTime!

Stream integration and safety maker.

## Setup

### Docker

> **Note**
> The Docker setup uses some images that are only available from YSTV's internal Docker registry.
> If you are a not a YSTV member you will need to build [Brave](https://github.com/ystv/brave), nginx-rtmp, and ffmpeg images yourself.
> We hope to make these images available publicly in the future.

Sign in to the YSTV Docker registry using your edit PC (AD) credentials - if you're not sure how to do this, message #computing on YSTV Slack:

```sh
$ docker login registry.comp.ystv.co.uk
```

Download the OAuth client from the Google Cloud Console (APIs and services /
Credentials) and save it as `youtube.json` in the project directory in a
new folder called credentials.

Then, run

```shell
$ docker compose up -d
```

ShowTime! will now be listening on http://localhost:8080.
You can now also stream to the nginx over RTMP on `rtmp://localhost/ingest/<stream-key>`.

### Manual

Download the OAuth client from the Google Cloud Console (APIs and services /
Credentials) and save it as `youtube.json` in the project directory in a
new folder called credentials.

Create an empty postgres database.

Create a `.env` file with the following parameters:

```
# Toggles authentication and displaying of errors
ST_DEBUG=true or false

# Nginx RTMP live application that is configured with a on_publish hook to ShowTime!
ST_INGEST_ADDR=rtmp://stream.example.com/ingest

# Nginx RTMP live application for Brave to output too
ST_OUTPUT_ADDR=rtmp://stream.example.com/output

# Address of where ShowTime! is being run, allows Brave to pull assets
ST_BASE_SERVE_ADDR=example.com

# Brave instance to support redundancy and channels
ST_BRAVE_ADDR=brave.example.com

# Database connection details
ST_DB_HOST=db.example.com
ST_DB_PORT=5432
ST_DB_SSLMODE=require
ST_DB_DBNAME=showtime
ST_DB_USERNAME=showtime_user
ST_DB_PASSWORD=something-long-and-random

# Note: these two don't need to be filled if debug is enabled.

# Verify web-auth tokens
ST_SIGNING_KEY=something-long-and-random

# Used for cookies
ST_DOMAIN_NAME=example.com

# Note: optional

# Path to folder with oauth2 credentials
ST_CRED_PATH=/etc/showtime/credentials
```

Initialise the postgres database with the `init` program.

```sh
go run cmd/init/init.go
```

Schemas and tables should now be present in the given database.

Create the folder structure to serve assets.

```sh
mkdir -p assets/ch
```

Add an image in the last directory that will be used for channel backgrounds
and save it as:

```
0-card-bg.jpg
```

## Running

After completing setup, run the main program.

```sh
go run cmd/main/main.go
```

ShowTime! will now be listening on `:8080`. See
[handlers.go](handlers/handlers.go) for possible paths.

The current authentication system only covers the `/api` path, it's best to use
a proxy which implements it's own authentication to prevent unauthorised access
to the other paths.

## Developing against

ShowTime! exposes a API which has JWT bearer token security that is compatible
with a [web-auth](https://github.com/ystv/web-auth) generated access token.

We use [goose](https://github.com/pressly/goose) to manage database migrations (read: upgrades/downgrades).
If you need to make changes to the database, install goose, then run

```shell
$ cd db/migrations
$ goose create describe_what_its_doing sql
# write the migration
$ cd ../..
$ go run cmd/init/init.go
```

If you need to undo the migration you just wrote, run

```shell
$ go run cmd/init/init.go -down_one
```

Before you push your changes and open a PR, don't forget to run `goose fix` (see [versioning](https://github.com/pressly/goose#hybrid-versioning)):

```shell
$ cd db/migrations
$ goose fix
```
