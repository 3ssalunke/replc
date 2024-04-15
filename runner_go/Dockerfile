FROM golang:1.21.6

WORKDIR /code

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/runner ./cmd/main.go

EXPOSE 3000
EXPOSE 3001

CMD [ "./bin/runner" ]