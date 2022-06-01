# ShowTime!

Stream integration and safety maker.

## Setup

You will need to download the OAuth client from the Google Cloud Console
(APIs and services / Credentials) and save it as `credentials.json` in
the project directory.

Create a `.env` file with the following parameters:

```
# Toggles authentication and displaying of errors
ST_DEBUG=true or false

# Nginx RTMP live application that is configured with a hook to ShowTime!
ST_INGEST_ADDR=rtmp://stream.example.com/ingest

# Address of where ShowTime! is being run, allows Brave to pull assets
ST_BASE_SERVE_ADDR=example.com

# Brave instance to support redundancy and channels
ST_BRAVE_ADDR=brave.example.com

# Note: these two don't need to be filled if debug is enabled.

# Verify web-auth tokens
ST_SIGNING_KEY=something-long-and-random

# Used for cookies
ST_DOMAIN_NAME=example.com
```

Create the database with the `init` program.

```sh
go run cmd/init/init.go
```

A new sqlite database "showtime.db" will appear in the working directory.

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
