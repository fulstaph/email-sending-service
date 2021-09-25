# syntax=docker/dockerfile:experimental

FROM golang:1.16 as builder

WORKDIR /app
COPY . .

RUN go mod tidy

RUN go build -o ./app ./cmd/acceptor/main.go

FROM gcr.io/distroless/static
COPY --from=builder /app/app /usr/bin/app
ENTRYPOINT ["/usr/bin/app"]
