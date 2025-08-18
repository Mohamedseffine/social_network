FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache bash gcc musl-dev

COPY . .

RUN go build -o server server.go

EXPOSE 8080

CMD ["./server"]