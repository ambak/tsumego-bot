FROM golang:1.23.4 AS builder

WORKDIR /tsumego-bot
COPY main.go go.mod go.sum ./
COPY command ./command
COPY config ./config
COPY sgf ./sgf
RUN go build -o tsumego-bot


FROM ubuntu:24.04

WORKDIR /tsumego-bot
RUN apt-get update && apt-get -y install python3-pip
RUN pip install sgfmill pillow --break-system-packages
COPY --from=builder tsumego-bot .
COPY sgf2image ./sgf2image

CMD ./tsumego-bot -c /data/config.json
