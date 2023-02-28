FROM golang:1.19-alpine

WORKDIR /app

RUN go install github.com/jackc/tern/v2@latest

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["go", "run", "./cmd/api"]