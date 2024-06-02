# syntax=docker/dockerfile:1

ARG version=undefined

FROM golang:1.22.3

LABEL org.opencontainers.image.source=https://github.com/shortcuts/codes
LABEL org.opencontainers.image.revision=$version

WORKDIR /codes

COPY . .

RUN go mod download

RUN make build

EXPOSE 8080

CMD [".bin/cmd"]
