FROM golang:1.22-alpine

WORKDIR /ten

COPY . .

RUN go build -o ten /ten/cmd/ten/main.go

ENTRYPOINT ["./ten"]
