FROM golang:1.23-alpine
RUN apk add --no-cache git gcc musl-dev
WORKDIR /app
RUN go install github.com/air-verse/air@v1.52.3
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2
# No copiamos go.mod aun porque lo inicializaremos abajo
COPY . .
CMD ["air", "-c", ".air.toml"]