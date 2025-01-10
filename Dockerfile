FROM golang:1.23.4 AS builder

WORKDIR /tsumego_bot
COPY main.go go.mod go.sum ./
RUN go build -o tsumego_bot


FROM ubuntu:24.04

WORKDIR /tsumego_bot
RUN apt-get update && apt-get -y install python3-pip
RUN pip install sgfmill pillow --break-system-packages
COPY --from=builder tsumego_bot .
COPY sgf2image ./sgf2image

CMD ./tsumego_bot -c /data/config.json