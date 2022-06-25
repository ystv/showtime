FROM golang:1.18 AS build

WORKDIR /workspace

# Copy the Go modules mainfests.
COPY go.mod go.sum ./

# Cache dependencies before building and copying source.
RUN go mod download

# Copy the source.
COPY . .

WORKDIR /workspace/cmd/main
RUN GOOS=linux GOARCH=amd64 go build -o showtime

FROM registry.comp.ystv.co.uk/ffmpeg:latest

COPY --from=build /workspace/cmd/main/showtime /usr/bin/

WORKDIR /opt/showtime
RUN mkdir -p assets/ch

EXPOSE 8080

HEALTHCHECK --interval=15s CMD curl --fail http://localhost:8080/api/health || exit 1

CMD ["showtime"]
