FROM golang:alpine as development

ADD ./src /go/src/app

WORKDIR /go/src/app

ENV PORT=8020

RUN apk --no-cache --virtual build-dependencies add \
    git \
    && go get -u github.com/go-sql-driver/mysql \
    && go get -u github.com/go-redis/redis \
    && go get -u github.com/JamaisMagic/asset-holders-COS-2E4 \
    && apk del build-dependencies

CMD ["go", "run", "main.go"]
