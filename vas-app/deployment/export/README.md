## 执行命令
例子：./export [filepath] -s=20190316200000 -d=20190316201200 -e=localhost:10011

参数说明：

- filepath string 文件导出路径 必填
- -s string start表示开始时间 时间格式为20060102150405 必填
- -d string end 表示结束时间 时间格式为20060102150405 必填
- -e string endpoint dashboard 服务的访问地址 必填

- -t string type non_motor-非机动车，vehicle-机动车
- -i string cameraId 摄像头 id
- -c int class 类别 3:饿了么 4:美团
- -t int eventType 事件类型 非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道；机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人
- -l int limit 限制记录条目数，默认不限制

- --marking string 打标状态 init:原始状态 illegal:违规 discard:作废

- --hasLabel int 是否包含标牌, 1:包含 2:不包含
- --labelScore float 标牌置信度, 取值范围0-1

- --hasFace int 是否包含人脸, 1:包含 2:不包含
- --similarity float 人脸相似度，默认为0, 取值范围0-100

- -f fontfile string 字体文件的路径 默认为"./font/PingFang-SC-Regular.ttf"


## 举例
### 导出预警列表
非机动车
./bin/export ./ -s=20190316200000 -d=20190316201200 -e=localhost:10011 --marking=init --type non_motor
机动车
./bin/export ./ -s=20190316200000 -d=20190316201200 -e=localhost:10011 --marking=init --type vehicle

### 导出违章列表
非机动车
./bin/export ./ -s=20190316200000 -d=20190316201200 -e=localhost:10011 --marking=illegal --type non_motor
机动车
./bin/export ./ -s=20190316200000 -d=20190316201200 -e=localhost:10011 --marking=illegal --type vehicle

## 导出废弃列表
非机动车
./bin/export ./ -s=20190316200000 -d=20190316201200 -e=localhost:10011 --marking=discard --type non_motor
机动车
./bin/export ./ -s=20190316200000 -d=20190316201200 -e=localhost:10011 --marking=discard --type vehicle

其他条件根据需要添加
