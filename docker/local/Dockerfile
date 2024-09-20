FROM debian:latest

ENV GO_VERSION 1.22.0

RUN echo "deb http://deb.debian.org/debian bookworm main" > /etc/apt/sources.list

RUN cat /etc/resolv.conf

RUN apt-get update && apt-get upgrade -y
RUN apt-get install wget make -y

RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
RUN rm go${GO_VERSION}.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /app

RUN go version

RUN apt install ffmpeg libavcodec-dev -y

RUN ffmpeg -version

COPY go.mod go.sum ./
RUN go mod download

CMD ["go", "run", "cmd/api/main.go"]
