FROM golang:alpine
RUN mkdir -p /go/src/projects/email-sending-service
ADD ../cmd/acceptor /go/src/projects/email-sending-service
WORKDIR /go/src/projects/email-sending-service
RUN ls -la
RUN go build -o main cmd/acceptor/main.go
EXPOSE 8080
CMD ["./main"]