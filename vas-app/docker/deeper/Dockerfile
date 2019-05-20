FROM golang:1.10.2-stretch as builder
COPY . /go/src/qiniu.com/vas-app/
RUN GOOS=linux GOARCH=amd64 go install qiniu.com/vas-app/app/deeper

FROM ubuntu:16.04

RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update && \
    apt-get -y install vim && \
    apt-get -y install net-tools && \
    apt-get -y install curl && \
    apt-get -y install telnet && \
    apt-get -y install iputils-ping

RUN apt-get install -y tzdata && ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /go/bin/deeper /workspace/deeper
COPY --from=builder /go/src/qiniu.com/vas-app/app/deeper/deeper.conf /workspace/deeper.conf

WORKDIR /workspace
CMD /workspace/deeper -f /workspace/deeper.conf
