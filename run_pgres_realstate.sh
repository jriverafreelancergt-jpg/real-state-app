docker run -d \
  --name postgres_server_realstate \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=$R3dh@t555 \
  -e POSTGRES_DB=realstatedb \
  -p 5432:5432 \
  -v pgdata:/var/lib/postgresql/data \
  postgres:15-alpine
