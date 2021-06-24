FROM golang:alpine
RUN mkdir /app
ADD ../cmd/sender /app/
WORKDIR /app
RUN go build -o main cmd/sender/main.go
CMD ["./main"]