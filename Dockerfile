FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /out/go_tmx .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/go_tmx /app/go_tmx
EXPOSE 8080
ENTRYPOINT ["/app/go_tmx", "-serve"]
