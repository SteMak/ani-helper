FROM golang:1.13.5-alpine3.11 AS build

WORKDIR /app

COPY . .

RUN go build \
        -o /app/bin/worker \
        /app/main/worker.go


FROM alpine:3.11

COPY --from=build /app/bin /bin

WORKDIR /bin

CMD /bin/worker 
