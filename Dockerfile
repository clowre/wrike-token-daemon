FROM golang:1.17-alpine3.15 AS build

WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd/wriked

FROM alpine:3.15
COPY --from=build /app/server /server

ENTRYPOINT ["/server"]