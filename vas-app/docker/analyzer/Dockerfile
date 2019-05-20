FROM golang:1.10.2-stretch as builder
RUN apt-get update && \
    apt-get  -y install wget

RUN wget -O /go/font.ttc  http://pq19tgs5k.bkt.clouddn.com/fonts/STHeitiLight.ttc
RUN wget -O /go/didianwei_non_motor_oc_model.tronmodel http://pq18sxmq9.bkt.clouddn.com/tronmodel/waimai_v31_merged.tronmodel
RUN wget -O /go/didianwei_non_motor_det_model.tronmodel  http://pq19tgs5k.bkt.clouddn.com/models/refinedet_resnet18_xuhui_v2_merged.tronmodel
## 人脸模型
# RUN wget -O /go/non_motor_det_local_model.tronmodel http://p0my22fmi.bkt.clouddn.com/refinedet_v0.0.3_merged.tronmodel
RUN wget -O /go/gaodianwei_non_motor_oc_model.tronmodel  http://pq19tgs5k.bkt.clouddn.com/models/waimaiv2_merged.tronmodel
RUN wget -O /go/gaodianwei_non_motor_det_model.tronmodel  http://pq19tgs5k.bkt.clouddn.com/models/refinedet_resnet18_waimai_v6_merged.tronmodel
RUN wget -O /go/waimai_cls_local_model.tronmodel http://pq19tgs5k.bkt.clouddn.com/models/traffic_sign_merged.tronmodel
RUN wget -O /go/waimai_east_local_model.tronmodel http://pq19tgs5k.bkt.clouddn.com/models/east_pedestrian_256X256_P4_v2.tronmodel
RUN wget -O /go/waimai_crnn_local_model.tronmodel http://pq19tgs5k.bkt.clouddn.com/models/crnn_48_10_merged.tronmodel
# RUN wget -O /go/vehicle_oc_local_model.tronmodel  http://pq19tgs5k.bkt.clouddn.com/models/direction_res18_tiny_merged.tronmodel
RUN wget -O /go/vehicle_plate.tronmodel  http://pq18sxmq9.bkt.clouddn.com/tronmodel/vehicle_plate_v3_merged_P4.tronmodel
RUN wget -O /go/yolov3-tiny_car.tronmodel http://pq19tgs5k.bkt.clouddn.com/models/yolov3-tiny_car.tronmodel
 
COPY . /go/src/qiniu.com/vas-app
RUN GOOS=linux GOARCH=amd64 go install qiniu.com/vas-app/app/analyzer

FROM reg-xs.qiniu.io/atlab/flow:v4.23.13_t5

RUN mkdir -p /workspace/data && \
    mkdir /workspace/conf && \
    mkdir /workspace/data/tmp 

COPY --from=builder /go/font.ttc /workspace/font.ttc
COPY --from=builder /go/didianwei_non_motor_oc_model.tronmodel /workspace/conf/didianwei_non_motor_oc_model.tronmodel
COPY --from=builder /go/didianwei_non_motor_det_model.tronmodel /workspace/conf/didianwei_non_motor_det_model.tronmodel
COPY --from=builder /go/gaodianwei_non_motor_oc_model.tronmodel /workspace/conf/gaodianwei_non_motor_oc_model.tronmodel
COPY --from=builder /go/gaodianwei_non_motor_det_model.tronmodel /workspace/conf/gaodianwei_non_motor_det_model.tronmodel

COPY --from=builder /go/waimai_cls_local_model.tronmodel /workspace/conf/waimai_cls_local_model.tronmodel
COPY --from=builder /go/waimai_crnn_local_model.tronmodel /workspace/conf/waimai_crnn_local_model.tronmodel
COPY --from=builder /go/waimai_east_local_model.tronmodel /workspace/conf/waimai_east_local_model.tronmodel
COPY --from=builder /go/vehicle_plate.tronmodel /workspace/conf/vehicle_plate.tronmodel
COPY --from=builder /go/yolov3-tiny_car.tronmodel /workspace/conf/yolov3-tiny_car.tronmodel

COPY --from=builder /go/bin/analyzer /workspace/analyzer
COPY --from=builder /go/src/qiniu.com/vas-app/app/analyzer/analyzer.conf /workspace/conf/analyzer.conf

WORKDIR /workspace
RUN ldconfig
CMD /workspace/analyzer  -f /workspace/conf/analyzer.conf
