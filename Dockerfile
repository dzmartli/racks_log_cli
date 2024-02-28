FROM golang:1.18

RUN apt-get install -y git

WORKDIR /usr/src/racks-log-cli

RUN mkdir -p /home/log && \
    groupadd log && \
    useradd -m -g log log -p racks && \
    usermod -aG log log

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o racks-log-cli