FROM golang:1.23-alpine3.20 as builder

WORKDIR /app

RUN apk add --no-cache git openssl tzdata

COPY . . 

RUN go mod download -x

RUN GOOS=linux go build -o koinan -x

FROM scratch

WORKDIR /app

COPY --from=builder /app/koinan .

CMD ["./koinan"]