FROM golang:1.23

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o merch-api cmd/server/main.go

EXPOSE 8080

CMD ["./merch-api"]