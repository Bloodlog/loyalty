FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOARCH=amd64 go build -o gophermart ./cmd/gophermart/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gophermart /app/gophermart

RUN chmod +x /app/gophermart

EXPOSE 8080

CMD ["/app/gophermart"]