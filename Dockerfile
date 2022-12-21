# syntax=docker/dockerfile:1
FROM golang:1.18 AS build

WORKDIR /workspace

# Copy the Go modules mainfests.
COPY go.mod go.sum ./

# Cache dependencies before building and copying source.
RUN --mount=type=cache,id=go-mod,target=/root/.cache/go-build \
    go mod download

# Copy the source.
COPY . .

WORKDIR /workspace/cmd/main
RUN GOOS=linux go build -o showtime

FROM debian:bullseye-slim
RUN apt update && apt install -y ca-certificates fonts-dejavu-core

COPY --from=mwader/static-ffmpeg:5.1.2 /ffmpeg /usr/local/bin/

COPY --from=build /workspace/cmd/main/showtime /usr/bin/

WORKDIR /opt/showtime
RUN mkdir -p assets/ch

EXPOSE 8080

HEALTHCHECK --interval=15s CMD curl --fail http://localhost:8080/api/health || exit 1

CMD ["showtime"]
