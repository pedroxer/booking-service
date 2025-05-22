FROM golang:1.24 AS builder

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o booking-service ./cmd/main.go


FROM alpine:3.18

WORKDIR /

COPY . .

COPY --from=builder /booking-service .

RUN apk --update --no-cache add curl
EXPOSE 8082
EXPOSE 2121

CMD ["./booking-service"]