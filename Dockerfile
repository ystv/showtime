FROM golang:1.18 AS build

WORKDIR /workspace

# Copy the Go modules mainfests.
COPY go.mod go.sum ./

# Cache dependencies before building and copying source.
RUN go mod download

# Copy the source.
COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o showtime ./cmd/main

FROM registry.comp.ystv.co.uk/ffmpeg:latest

COPY --from=build /workspace/showtime /usr/bin/
EXPOSE 8080
ENTRYPOINT ["/usr/bin/showtime"]
