FROM golang:alpine

ADD ./src /go/src/app

WORKDIR /go/src/app

ENV PORT=8020

RUN go get -u github.com/go-sql-driver/mysql

CMD ["go", "run", "main.go"]
