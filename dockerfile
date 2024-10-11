FROM golang:1.22.2 AS builder

WORKDIR /src

COPY . .

RUN apt-get update && apt-get install -y libpcap-dev
RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN make
