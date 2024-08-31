FROM golang:1-bookworm as build

ADD . /usr/local/src
WORKDIR /usr/local/src
RUN go build -o port-jump -ldflags="-w -s"

FROM debian:bookworm-slim

COPY --from=build /usr/local/src/port-jump /port-jump

VOLUME [ "/root/.config/port-jump" ]
ENTRYPOINT [ "/port-jump" ]
