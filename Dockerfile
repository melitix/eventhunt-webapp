FROM golang:1.21.3

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . /app/

WORKDIR /app/webapp

RUN CGO_ENABLED=0 GOOS=linux go build -o /the-go-app

CMD ["/the-go-app"]

EXPOSE 8200
