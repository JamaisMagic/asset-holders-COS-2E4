FROM golang:alpine

ADD ./src /go/src/app

WORKDIR /go/src/app

ENV PORT=8020

RUN apk --no-cache --virtual build-dependencies add \
    git \
    && go get -u github.com/go-sql-driver/mysql \
    && apk del build-dependencies

CMD ["go", "run", "main.go"]
