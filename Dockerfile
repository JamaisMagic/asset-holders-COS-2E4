FROM golang:alpine as development

ADD . ./

ENV PORT=8020
ENV GO111MODULE=on

WORKDIR /

RUN apk --no-cache --virtual build-dependencies add \
    git \
    && go get -u github.com/go-sql-driver/mysql \
    && go get -u github.com/go-redis/redis \
    && apk del build-dependencies

CMD ["go", "run", "./src/main.go"]
