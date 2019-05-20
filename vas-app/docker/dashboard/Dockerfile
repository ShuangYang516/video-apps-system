FROM denismakogon/gocv-alpine:4.0.1-buildstage as builder
COPY . /go/src/qiniu.com/vas-app
RUN GOOS=linux GOARCH=amd64 go install qiniu.com/vas-app/app/dashboard

FROM denismakogon/gocv-alpine:4.0.1-runtime

RUN mkdir -p /workspace
ENV TZ=Asia/Shanghai

COPY --from=builder /go/bin/dashboard  /workspace/dashboard
COPY --from=builder /go/src/qiniu.com/vas-app/app/dashboard/dashboard.conf /workspace/conf/dashboard.conf
COPY --from=builder /go/src/qiniu.com/vas-app/deployment/export/font /workspace/font

WORKDIR /workspace
CMD /workspace/dashboard -f /workspace/conf/dashboard.conf
