FROM golang:1.27.1-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
RUN mkdir -p /var/lib/alumnee/files && chown -R 10001:10001 /var/lib/alumnee
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/migrate /usr/local/bin/migrate
USER 10001:10001
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/api"]
