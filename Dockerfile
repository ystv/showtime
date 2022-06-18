FROM golang:1.18 AS build

WORKDIR /workspace

# Copy the Go modules mainfests.
COPY go.mod go.sum ./

# Cache dependencies before building and copying source.
RUN go mod download

# Copy the source.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o showtime ./cmd/main

FROM scratch

COPY --from=build /workspace/showtime /

EXPOSE 8080

ENTRYPOINT ["/showtime"]
