FROM golang:1.16 as builder

WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o ./acceptor ./cmd/acceptor/main.go

FROM gcr.io/distroless/static
COPY --from=builder /app/acceptor /usr/bin/app
ENTRYPOINT ["/usr/bin/acceptor"]
