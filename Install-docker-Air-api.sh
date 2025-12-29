FROM golang:1.23-alpine

WORKDIR /app

# Instalar dependencias necesarias para compilar y herramientas de ayuda
RUN apk add --no-cache git

# Instalar Air para Hot Reload
RUN go install github.com/air-verse/air@latest

# Instalar golang-migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Air se encargará de ejecutar el binario
CMD ["air", "-c", ".air.toml"]
