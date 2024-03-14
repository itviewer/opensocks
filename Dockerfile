FROM golang:alpine

WORKDIR /app
COPY . /app

RUN go build -o ./bin/opensocks ./main.go

ENTRYPOINT ["./bin/opensocks"]

