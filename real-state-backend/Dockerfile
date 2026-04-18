FROM golang:1.24-alpine
RUN apk add --no-cache git gcc musl-dev tzdata
RUN cp /usr/share/zoneinfo/America/Guatemala /etc/localtime
RUN echo "America/Guatemala" > /etc/timezone
WORKDIR /app
RUN go install github.com/air-verse/air@v1.52.3
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2
COPY go.mod go.sum ./
RUN go mod download
# No copiamos go.mod aun porque lo inicializaremos abajo
COPY . .
CMD ["air", "-c", ".air.toml"]
