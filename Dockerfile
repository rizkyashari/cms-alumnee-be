ARG GOLANG_IMAGE_TAG="1.27.1-alpine"
FROM golang:${GOLANG_IMAGE_TAG}

WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@v1.41.0

COPY . .
RUN go mod tidy