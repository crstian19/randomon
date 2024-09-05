FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o /app/randomom

FROM alpine:3.18

COPY --from=builder /app/randomom /randomom

EXPOSE 3000

CMD ["/randomom"]
