FROM golang:1.13.5-alpine3.11 AS build

WORKDIR /app

COPY . .

RUN go build \
        -o /app/worker \
        /app/main/worker.go


FROM alpine:3.11

WORKDIR /app

COPY --from=build /app/worker /app/worker

CMD /app/worker
# POSTGRES_URI=postgres://root:root@localhost:5432/test
#docker run --rm -it --network host -p 5432:5432 -e POSTGRES_DB=test -e POSTGRES_PASSWORD=root postgres