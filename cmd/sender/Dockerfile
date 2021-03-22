FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main cmd/sender/main.go
CMD ["./main"]