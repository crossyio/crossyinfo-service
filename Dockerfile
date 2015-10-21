FROM golang:1.5
MAINTAINER Crossy.IO <docker@crossy.io>
EXPOSE 80

ADD https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm /go/bin/
RUN chmod +x /go/bin/gpm

COPY Godeps /go/
RUN gpm install

COPY . /go/
RUN go build -o crossyinfo-service
ENTRYPOINT ./crossyinfo-service
