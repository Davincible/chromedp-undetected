FROM golang:1.19

RUN apt-get update && apt-get -y upgrade && apt-get -y install gcc g++ ca-certificates chromium xvfb

WORKDIR /app

COPY . .

RUN go mod download

ENTRYPOINT [ "go", "test", "-v", "-count=1", "." ]
