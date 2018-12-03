# Dockerfile
FROM golang:1.11

ADD . /socialservice

WORKDIR /socialservice

RUN go build ./...

CMD sh -c "go run /socialservice/server/main/main.go -dsn admin:test@tcp\(db\)/nphw3 0.0.0.0 5100"