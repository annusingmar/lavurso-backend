FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

ARG POSTGRES_DATABASE="lavurso"
ARG POSTGRES_USER="lavurso"
ARG POSTGRES_PASSWORD="postgres_pwd"

RUN sed -i "s/POSTGRES_DATABASE/${POSTGRES_DATABASE}/g" config.toml
RUN sed -i "s/POSTGRES_USER/${POSTGRES_USER}/g" config.toml
RUN sed -i "s/POSTGRES_PASSWORD/${POSTGRES_PASSWORD}/g" config.toml

CMD ["go", "run", "./cmd/api"]