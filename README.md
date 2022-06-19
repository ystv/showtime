# ShowTime!

Stream integration and safety maker.

## Setup

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
