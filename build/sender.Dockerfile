FROM golang:1.16 as builder

WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o ./sender ./cmd/sender/main.go

FROM gcr.io/distroless/static
COPY --from=builder /app/sender /usr/bin/app
ENTRYPOINT ["/usr/bin/sender"]
