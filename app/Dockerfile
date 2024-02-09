FROM golang:1.22.0-alpine as build

WORKDIR /app
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/server

FROM alpine
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates
RUN apk add --no-cache curl
COPY --from=build /app/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
