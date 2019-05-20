FROM denismakogon/gocv-alpine:4.0.1-buildstage as builder
COPY . /go/src/qiniu.com/vas-app/
RUN GOOS=linux GOARCH=amd64 go install qiniu.com/vas-app/app/export

FROM denismakogon/gocv-alpine:4.0.1-runtime

RUN apk update && apk add --no-cache vim net-tools curl bash tree tzdata \
    && mkdir -p /workspace
ENV TZ=Asia/Shanghai

COPY --from=builder /go/bin/export  /workspace/export
COPY --from=builder /go/src/qiniu.com/vas-app/deployment/export/font /workspace/font

WORKDIR /workspace
CMD echo export
