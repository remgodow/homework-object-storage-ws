FROM golang:1.15-alpine3.13
WORKDIR /mnt/homework
COPY . .
RUN go build -o homework

# Docker is used as a base image so you can easily start playing around in the container using the Docker command line client.
FROM docker
COPY --from=0 /mnt/homework/homework /usr/local/bin/homework
RUN apk add bash curl
