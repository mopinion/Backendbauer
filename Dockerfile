# https://hub.docker.com/_/golang/

FROM golang:1.5
# FROM ubuntu:trusty

MAINTAINER Mopinion.com

# Go
# RUN wget https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz
# RUN tar -C /usr/local -xzf go1.5.linux-amd64.tar.gz
# ENV PATH "$PATH:/usr/local/go/bin"

ENV BB_PATH /var/www/backendbauer

# deps
# RUN go get github.com/fzzy/radix/redis
# RUN go get github.com/ziutek/mymysql/mysql
# RUN go get github.com/ziutek/mymysql/native
RUN go get gopkg.in/mgo.v2
RUN go get github.com/fzzy/radix/redis
RUN go get github.com/ziutek/mymysql/native
RUN go get github.com/abbot/go-http-auth

# crypto
# RUN mkdir -p $GOPATH/src/golang.org/x/
# RUN cd $GOPATH/src/golang.org/x/
# RUN git clone git@github.com:golang/crypto.git

# files
ADD ./ $BB_PATH
RUN cd $BB_PATH
RUN go build -o $BB_PATH/server/server $BB_PATH/server/server.go

EXPOSE 8181
EXPOSE 8484
EXPOSE 81

RUN cd $BB_PATH

# CMD ['nohup','/var/www/backendbauer_efm/server','8181','&']
CMD $BB_PATH/server 81
