#!/bin/bash

# Script para controlar los contenedores Docker de la API de bienes raíces
# Uso: ./docker-control.sh [up|down|restart|logs]

set -e

COMPOSE_FILE="docker-compose.yml"
DB_CONTAINER="postgres_server_realstate"
NETWORK="real-state-backend_realstate_network"

case "$1" in
    up)
        echo "Verificando PostgreSQL..."
        # función para comprobar si un puerto está en uso
        is_port_in_use() {
            if command -v ss >/dev/null 2>&1; then
                ss -lnt | awk '{print $4}' | grep -q ":$1$"
            elif command -v netstat >/dev/null 2>&1; then
                netstat -lnt | awk '{print $4}' | grep -q ":$1$"
            else
                return 1
            fi
        }

        # Si ya está corriendo, continuar
        if docker ps --format '{{.Names}}' | grep -q "^$DB_CONTAINER$"; then
            echo "PostgreSQL ya está corriendo en el contenedor '$DB_CONTAINER'."
        else
            # Si existe pero detenido, intentar iniciar
            if docker ps -a --format '{{.Names}}' | grep -q "^$DB_CONTAINER$"; then
                echo "Contenedor '$DB_CONTAINER' existe pero está detenido. Iniciando..."
                if docker start "$DB_CONTAINER" >/dev/null 2>&1; then
                    echo "Contenedor '$DB_CONTAINER' iniciado."
                else
                    echo "Error al iniciar el contenedor '$DB_CONTAINER'. Revisa 'docker logs $DB_CONTAINER' para más detalles."
                    exit 1
                fi
            else
                # Verificar si el puerto 5432 está en uso en el host
                HOST_PORT=5432
                if is_port_in_use "$HOST_PORT"; then
                    echo "Puerto $HOST_PORT ya está en uso en el host. Buscando puerto disponible..."
                    for p in $(seq 5432 5442); do
                        if ! is_port_in_use "$p"; then
                            HOST_PORT=$p
                            break
                        fi
                    done
                    echo "Usando puerto $HOST_PORT para PostgreSQL en el host."
                fi

                echo "Creando e iniciando el contenedor PostgreSQL '$DB_CONTAINER' en el puerto $HOST_PORT..."
                if ! docker run --name "$DB_CONTAINER" -e POSTGRES_USER=uadmin -e POSTGRES_PASSWORD='$R3hat555' -e POSTGRES_DB=realstatedb -p "$HOST_PORT":5432 -d postgres:15 >/dev/null 2>&1; then
                    echo "Fallo al crear o iniciar '$DB_CONTAINER'. Revisa el puerto o permisos de Docker."
                    exit 1
                fi

                # Comprobar si se inició correctamente
                if docker ps --format '{{.Names}}' | grep -q "^$DB_CONTAINER$"; then
                    echo "Contenedor '$DB_CONTAINER' iniciado correctamente (host port: $HOST_PORT)."
                else
                    echo "Fallo al crear o iniciar '$DB_CONTAINER'. Revisa 'docker ps -a' y 'docker logs $DB_CONTAINER'."
                    exit 1
                fi
            fi
        fi

        echo "Levantando la API..."
        docker-compose -f "$COMPOSE_FILE" up -d --build
        sleep 2
        echo "API corriendo en http://localhost:8080"

        echo "Conectando PostgreSQL a la red..."
        docker network connect "$NETWORK" "$DB_CONTAINER" 2>/dev/null || true
        ;;

    down)
        echo "Desconectando PostgreSQL de la red..."
        docker network disconnect "$NETWORK" "$DB_CONTAINER" 2>/dev/null || true

        echo "Deteniendo la API..."
        docker-compose -f "$COMPOSE_FILE" down
        echo "API detenida."
        ;;

    restart)
        echo "Reiniciando la API..."
        docker-compose -f "$COMPOSE_FILE" restart
        echo "API reiniciada."
        ;;

    logs)
        echo "Mostrando logs de la API..."
        docker-compose -f "$COMPOSE_FILE" logs -f
        ;;

    *)
        echo "Uso: $0 {up|down|restart|logs}"
        echo "  up      - Levantar la API (asegúrate de que PostgreSQL esté corriendo)"
        echo "  down    - Detener la API"
        echo "  restart - Reiniciar la API"
        echo "  logs    - Ver logs de la API"
        exit 1
        ;;
esac