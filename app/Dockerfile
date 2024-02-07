FROM --platform=linux/amd64 golang:1.21.6-alpine as build

WORKDIR /app
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/server

FROM --platform=linux/amd64 scratch
COPY --from=build /app/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]