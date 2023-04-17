FROM golang:1.20-alpine as build

WORKDIR /app

COPY main.go .

RUN CGO_ENABLED=0 go build -o main -ldflags="-s -w" main.go

FROM scratch

COPY --from=build /app/main /app/main

WORKDIR /app

EXPOSE 8080

CMD [ "./main" ]
