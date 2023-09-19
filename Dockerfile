ARG GOLANG_IMAGE_TAG="1.19.3-alpine3.17"
FROM golang:${GOLANG_IMAGE_TAG}

WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@latest

COPY . .
RUN sh script/install-vips.sh
RUN go mod tidy