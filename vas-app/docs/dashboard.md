## 一、非机动车违章api

### 1.1 获取当前违章列表
GET: v1/events

|字段名|类型|说明|
|----|----|----|
|page|int|页码|
|per_page|int|每页大小，默认10|
|isClassFaceFilter|boolean|是否根据类型及人脸进行筛选, 用于非机动车，机动车不要用！！！|
|start|int|起始时间戳，毫秒级13位|
|end|int|结束时间戳，毫秒级13位|
|cameraIds|[]string|摄像头id|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventTypes|[]int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|
|classes|[]int|类别，3:饿了么 4:美团|
|marking|string|打标状态 init:原始状态 illegal:违规 discard:作废|
|hasLabel|int|是否包含标牌, 1:包含 2:不包含|
|label|string|标牌，模糊查询|
|labelScore|float|标牌置信度 0-1|
|eventId|string|事件ID，模糊查询|
|hasFace|int|是否包含人脸, 1:包含 2:不包含|
|name|string|姓名，模糊查询|
|idCard|string|身份证号码，模糊查询|
|similarity|float|人脸相似度 0-100|

注意：type为事件类型的大类，分非机动车和机动车类型，如果不设置eventType，以type大类为准，否则以eventType为准。

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "content": [
            {
                "id": "5c511735272e1c59ff0f8b50",
                "eventId": "2",
                "eventType": 32,
                "address": "rtsp://localhost/vod/video",
                "cameraId": "test",
                "snapshot": [
                    {
                        "featureUri": "http://plk5h8slj.bkt.clouddn.com/object.jpg",
                        "snapshotUri": "http://plk5h8slj.bkt.clouddn.com/snapshort.jpg",
                        "pts": [],
                        "score": 0,
                        "class": 3,
                        "label": "123456",
                        "labelScore": 0.9
                    }
                ],
                "zone": {},
                "deeperInfo": {},
                "startTime": "0001-01-01T00:00:00Z",
                "endTime": "0001-01-01T00:00:00Z",
                "createdAt": "2019-01-30T11:17:09.294+08:00",
                "updatedAt": "2019-01-30T11:17:09.294+08:00",
                "status": "",
                "mark": {
                    "marking": "illegal",
                    "isClassEdit": true,
                    "isLabelEdit": true,
                    "discardReason": ""
                },
                "eventTypeStr": "逆行",
                "classStr": [
                    "饿了么"
                ],
                "faces": [
                    {
                        "faceImageUri": "",
                        "name": "gege",
                        "nation": "",
                        "sex": "0",
                        "idCard": "123456",
                        "similarity": 90
                    }
                ]
            }
        ],
        "page": 1,
        "per_page": 1,
        "total_page": 2,
        "total_count": 2
    }
}

```

|字段名|类型|说明|
|----|----|----|
|id|string|数据库id，唯一|
|eventId|string|事件ID ，全局唯一|
|eventType|int|事件类型| 
|address|string|流地址| 
|cameraId|string|摄像头ID| 
|snapshot.featureUri|string|特写图片|
|snapshot.snapshotUri|string|截帧图片| 
|snapshot.pts|array|二维数组，检测目标坐标|  
|snapshot.score|float|置信度| 
|snapshot.class|int|类别|
|snapshot.label|string|标牌，非机动车-车牌号|
|snapshot.labelScore|float|标牌置信度|
|zone|interface|划线或监测区域配置| 
|deeperInfo|interface|深度解析信息| 
|startTime|string|事件起始时间| 
|endTime|string|事件结束时间| 
|createdAt|string|数据产生时间| 
|updatedAt|string|数据更新时间| 
|status|string|"init" "finished"| 
|mark|struct|打标结构体|
|mark.marking|string|"init" "illegal" "discard"|
|mark.isClassEdit|boolean|类别是否被编辑过|
|mark.isLabelEdit|boolean|标牌是否被编辑过|
|mark.discardReason|string|作废原因|
|eventTypeStr|string|事件类型中文|
|classStr|array|类别中文，对应于snapshot.class|
|faces|array|匹配人脸信息，与snapshot一一对应，当status为"finished"，才会返回该结果|
|faces.name|string|姓名|
|faces.nation|string|民族|
|faces.sex|string|性别 1: Male, 0: Female|
|faces.idCard|string|身份证号|
|faces.similarity|float|相似度|
|faces.faceImageUrl|string|人脸图片地址|
|page|int|页码|
|per_page|int|每页大小|
|total_page|int|总页数|
|total_count|int|总数|

### 1.2 获取违章分析，按小时,默认今天
GET: v1/events/analysis/hourly  

|字段名|类型|说明|
|----|----|----|
|cameraId|string|摄像头id|
|date|string|日期，格式(2019-01-01)|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventType|int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人转|

```json
{
    "code": 0,
    "msg": "ok",
    "data": [ //0-24小时每个时间段违章个数
        22,
        33,
        44,
        22,
        ...
    ]
}

```

### 1.3 获取违章分析，按天，默认最近七天
GET: v1/events/analysis/daily   


|字段名|类型|说明|
|----|----|----|
|cameraId|string|摄像头id|
|from|string|日期，格式(2019-01-01)|
|to|string|日期，格式(2019-01-01)|
|class|int|类别，3:饿了么 4:美团|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventType|int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|

```json
{
    "code": 0,
    "msg": "ok",
    "data":{
        "2019-01-01" : 33,
        "2019-01-02" : 11,
        "2019-01-03" : 2,
        "2019-01-04" : 333
    }
}

```


### 1.4 获取违章类别占比，日，周
GET: v1/events/analysis/class  

|字段名|类型|说明|
|----|----|----|
|cameraId|string|摄像头id|
|from|string|日期，格式(2019-01-01)|
|to|string|日期，格式(2019-01-01)|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventType|int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|

```json
{
    "code": 0,
    "msg": "ok",
    "data":[
        {"class":"美团","eventsNum":111},
        {"class":"饿了么","eventsNum":222},
        {"class":"其他","eventsNum":333}
    ]
}

```

### 1.5 获取违章类型占比，日，周
GET: v1/events/analysis/type  

|字段名|类型|说明|
|----|----|----|
|cameraId|string|摄像头id|
|from|string|日期，格式(2019-01-01)|
|to|string|日期，格式(2019-01-01)|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventType|int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|

```json
{
    "code": 0,
    "msg": "ok",
    "data":[
        {"eventType":"闯红灯","eventsNum":111},
        {"eventType":"逆行","eventsNum":222},
        {"eventType":"停车越线","eventsNum":333}
    ]
}

```
### 1.6 获取路口占比，日，周
GET: v1/events/analysis/camera

|字段名|类型|说明|
|----|----|----|
|from|string|日期，格式(2019-01-01)|
|to|string|日期，格式(2019-01-01)|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventType|int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|

```json
{
    "code": 0,
    "msg": "ok",
    "data":[
        {"cameraName":"XX路口","eventsNum":111},
    ]
}

```

### 1.7 导出违章记录
GET: v1/events/export

|字段名|类型|说明|
|----|----|----|
|start|int|起始时间戳，毫秒级13位|
|end|int|结束时间戳，毫秒级13位|
|cameraIds|[]string|摄像头id|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventTypes|[]int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|
|classes|[]int|类别，3:饿了么 4:美团|
|marking|string|打标状态 init:原始状态 illegal:违规 discard:作废|
|hasLabel|int|是否包含标牌, 1:包含 2:不包含|
|label|string|标牌，模糊查询|
|labelScore|float|标牌置信度 0-1|
|eventId|string|事件ID，模糊查询|
|hasFace|int|是否包含人脸, 1:包含 2:不包含|
|name|string|姓名，模糊查询|
|idCard|string|身份证号码，模糊查询|
|similarity|float|人脸相似度 0-100|
|limit|int|限制条数，不填默认为100000|
|ids|[]string|导出事件id数组，该数组存在，则以上查询条件都失效|

输出 csv 文件

### 1.8 导出取证记录
GET: v1/events/export/evidence

|字段名|类型|说明|
|----|----|----|
|start|int|起始时间戳，毫秒级13位|
|end|int|结束时间戳，毫秒级13位|
|cameraIds|[]string|摄像头id|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|
|eventTypes|[]int|违法类型，非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道； 机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人|
|classes|[]int|类别，3:饿了么 4:美团|
|marking|string|打标状态 init:原始状态 illegal:违规 discard:作废|
|hasLabel|int|是否包含标牌, 1:包含 2:不包含|
|label|string|标牌，模糊查询|
|labelScore|float|标牌置信度 0-1|
|eventId|string|事件ID，模糊查询|
|hasFace|int|是否包含人脸, 1:包含 2:不包含|
|name|string|姓名，模糊查询|
|idCard|string|身份证号码，模糊查询|
|similarity|float|人脸相似度 0-100|
|limit|int|限制条数，不填默认为100000|
|ids|[]string|导出事件id数组，该数组存在，则以上查询条件都失效|

输出 zip 文件

注意：由于这个操作比较耗时，所以做了限流操作，20秒内只能进行两次操作，如果操作频繁，会返回”Reach request limiting!“

### 1.9 更新事件
PUT: v1/events/:id

|字段名|类型|说明|
|----|----|----|
|id|string|日志唯一ID|

请求参数

```json
{
    "class": 3,
    "label": "123456"
}

```
|字段名|类型|说明|
|----|----|----|
|class|int|类别，3:饿了么 4:美团|
|label|string|标牌，非机动车-车牌号|

返回 event 的结构体

### 1.10 删除事件
DELETE: v1/events

请求
```json
{
    "ids": ["5c511735272e1c59ff0f8b50"]
}

```
|字段名|类型|说明|
|----|----|----|
|ids|[]string|删除的id数组|

返回
```json
{
   "code": 0,
    "msg": "ok",
    "data": {
        "success": ["5c511735272e1c59ff0f8b50"],
        "fail": []
    }
}

```

|字段名|类型|说明|
|----|----|----|
|success|[]string|删除成功的id数组|
|fail|[]string|删除失败的id数组|

### 1.11 打标违章
POST: v1/mark/illegal

请求
```json
{
    "ids": ["5c511735272e1c59ff0f8b50"]
}

```

|字段名|类型|说明|
|----|----|----|
|ids|[]string|打标的id数组|

返回
```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "success": ["5c511735272e1c59ff0f8b50"],
        "fail": []
    }
}

```
|字段名|类型|说明|
|----|----|----|
|success|[]string|打标成功的id数组|
|fail|[]string|打标失败的id数组|

### 1.12 打标作废
POST: v1/mark/discard

请求
```json
{
    "ids": ["5c511735272e1c59ff0f8b50"]
}

```

|字段名|类型|说明|
|----|----|----|
|ids|[]string|打标的id数组|

返回
```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "success": ["5c511735272e1c59ff0f8b50"],
        "fail": []
    }
}

```
|字段名|类型|说明|
|----|----|----|
|success|[]string|打标成功的id数组|
|fail|[]string|打标失败的id数组|

### 1.13 获取事件类型枚举
GET: v1/event/type/enums

请求

|字段名|类型|说明|
|----|----|----|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|

返回
```json
{
    "code": 0,
    "msg": "ok",
    "data": [
        {
            "name": "机动车大弯小转",
            "value": 2101
        }
    ]
}

```
|字段名|类型|说明|
|----|----|----|
|name|string|名称|
|value|int|值|

### 1.14 获取类别枚举
GET: v1/event/class/enums

请求

|字段名|类型|说明|
|----|----|----|
|type|string|"non_motor":非机动车  "vehicle":机动车， 不填默认为全部|

返回
```json
{
    "code": 0,
    "msg": "ok",
    "data": [
        {
            "name": "饿了么",
            "value": 3
        },
        {
            "name": "美团",
            "value": 4
        }
    ]
}

```
|字段名|类型|说明|
|----|----|----|
|name|string|名称|
|value|int|值|

## 二、流任务管理api

### 2.1 获取所有流任务
GET: v1/tasks

|字段名|类型|说明|
|----|----|----|
|page|int|页码|
|per_page|int|每页大小，默认10|

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "content": [
            {
                "id": "1",
                "region": "xuhui11",
                "cameraId": "test22",
                "streamAddr": "ss",
                "location": "徐汇1",
                "status": "ON",
                "createdAt": "2019-02-13T16:41:43.051+08:00",
                "updatedAt": "2019-02-13T16:42:08.554+08:00",
                "worker": {},
                "lastErrorMsg": "",
                "lastErrorTime": "0001-01-01T00:00:00Z",
                "restartTimes": 0,
                "config": {}
            }
        ],
        "page": 1,
        "per_page": 1,
        "total_page": 2,
        "total_count": 2
    }
}

```

|字段名|类型|说明|
|----|----|----|
|id|string|日志唯一ID|
|region|string|区域，xuhui/pudong|
|cameraId|string|摄像头ID，为空则以StreamAddr为准|
|streamAddr|string|流地址|
|location|string|位置|
|status|string|状态，"ON"/"OFF"|
|createdAt|string|数据产生时间|
|updatedAt|string|数据更新时间|
|worker|interface|表示哪个IP和PID的实例在消费这个流，用于方便排查哪个实例|
|lastErrorMsg|string|最后一次错误信息|
|lastErrorTime|string|最后一次错误时间|
|restartTimes|string|重启次数|
|config|interface|配置|
|page|int|页码|
|per_page|int|每页大小|
|total_page|int|总页数|
|total_count|int|总数|

### 2.2 新增流任务
POST: v1/tasks

请求参数

```json
{
    "id": "1",
    "region": "xuhui11",
    "cameraId": "test22",
    "streamAddr": "ss",
    "location": "徐汇1",
    "status": "ON",
    "worker": {},
    "config": {}
}

```

返回

```json
{
    "code": 0,
    "msg": "ok",
    "data": null
}

```

### 2.3 更新流任务
PUT: v1/tasks/:id

|字段名|类型|说明|
|----|----|----|
|id|string|日志唯一ID|

请求参数

```json
{
    "region": "xuhui11",
    "cameraId": "test22",
    "streamAddr": "ss",
    "location": "徐汇1",
    "status": "ON",
    "worker": {},
    "config": {}
}

```

返回

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "id": "1",
        "region": "xuhui11",
        "cameraId": "test22",
        "streamAddr": "ss",
        "location": "徐汇1",
        "status": "ON",
        "createdAt": "2019-02-13T16:41:43.051+08:00",
        "updatedAt": "2019-02-13T16:42:08.554+08:00",
        "worker": {},
        "lastErrorMsg": "",
        "lastErrorTime": "0001-01-01T00:00:00Z",
        "restartTimes": 0,
        "config": {}
    }
}

```

### 2.4 删除流任务
DELETE: v1/tasks/:id

|字段名|类型|说明|
|----|----|----|
|id|string|日志唯一ID|

返回

```json
{
    "code": 0,
    "msg": "ok",
    "data": null
}

```

### 2.5 启动流任务
POST: v1/start/tasks/:id

|字段名|类型|说明|
|----|----|----|
|id|string|日志唯一ID|

返回

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "id": "1",
        "region": "xuhui11",
        "cameraId": "test22",
        "streamAddr": "ss",
        "location": "徐汇1",
        "status": "ON",
        "createdAt": "2019-02-13T16:41:43.051+08:00",
        "updatedAt": "2019-02-13T16:42:08.554+08:00",
        "worker": {},
        "lastErrorMsg": "",
        "lastErrorTime": "0001-01-01T00:00:00Z",
        "restartTimes": 0,
        "config": {}
    }
}

```

### 2.6 停止流任务
POST: v1/stop/tasks/:id

|字段名|类型|说明|
|----|----|----|
|id|string|日志唯一ID|

返回

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "id": "1",
        "region": "xuhui11",
        "cameraId": "test22",
        "streamAddr": "ss",
        "location": "徐汇1",
        "status": "OFF",
        "createdAt": "2019-02-13T16:41:43.051+08:00",
        "updatedAt": "2019-02-13T16:42:08.554+08:00",
        "worker": {},
        "lastErrorMsg": "",
        "lastErrorTime": "0001-01-01T00:00:00Z",
        "restartTimes": 0,
        "config": {}
    }
}

```

### 2.7 获取所有流
GET: v1/streams

请求参数：
|字段名|类型|说明|
|----|----|----|
|cameraName|string|摄像头名称|

返回：

```json
{
    "code": 0,
    "msg": "ok",
    "data": [
        {
            "task_id": "xh_0001",
            "stream": "http://100.100.62.123:8088/live?port=1935&app=vchannel&stream=xh_0001",
            "camera_id": "QN0011586a282eb0f5d9a_1",
            "camera_name": "onvif_ip45"
        }
    ]
}

```

|字段名|类型|说明|
|----|----|----|
|task_id|string|流任务id|
|stream|string|http flv 流地址|
|camera_name|string|摄像头名称|

## 三、设备管理api

### 3.1 获取所有子设备
GET: v1/vms/device/sub

|字段名|类型|说明|
|----|----|----|
|page|int|页码|
|per_page|int|每页大小，默认10|

```json
{
    "code": 0,
    "data": {
        "content": [
            {
                "id": "5c74a9836042bb00016b5913",
                "type": {
                    "name": "枪机",
                    "value": 1
                },
                "status": {
                    "name": "在线",
                    "value": 2
                },
                "channel": 2,
                "created_at": "2019-02-26 10:50:43",
                "organization": {
                    "id": "000000000000000000000000",
                    "name": "默认分组"
                },
                "device_id": "5c73ff0c6042bb00016b5911",
                "user_id": "5bac7e9eea395739c32bdc14",
                "attribute": {
                    "name": "徐汇测试视频1",
                    "ip": "",
                    "discovery_protocol": {
                        "name": "流地址",
                        "value": 2
                    },
                    "description": "徐汇测试视频1",
                    "vendor": {
                        "name": "海康",
                        "value": 1
                    },
                    "account": "",
                    "upstream_url": "rtmp://100.100.62.123/vchannel/xu_test1",
                    "channel_type": {
                        "name": "",
                        "value": 0
                    },
                    "internal_channel": 0
                }
            }
        ],
        "page": 1,
        "per_page": 1,
        "total_page": 2,
        "total_count": 2
    }
}

```

|字段名|类型|说明|
|----|----|----|
|id|string|子设备id|
|type|struct|类型|
|type.name|string|类型名称|
|type.value|int|类型值|
|status|struct|状态|
|status.name|string|状态名称|
|status.value|int|状态值|
|channel|int|通道号|
|created_at|string|创建时间|
|attribute|struct|属性|
|attribute.name|string|名称|
|attribute.ip|string|IP|
|attribute.discovery_protocol|struct|对接方式|
|attribute.discovery_protocol.name|string|对接方式名称|
|attribute.discovery_protocol.value|int|对接方式值|
|description|string|描述|
|vendor|struct|厂商|
|vendor.name|string|厂商名称|
|vendor.value|int|厂商值|
|upstream_url|string|流地址|
|page|int|页码|
|per_page|int|每页大小|
|total_page|int|总页数|
|total_count|int|总数|

### 3.2 获取所有摄像头
GET: v1/cameras

请求参数：
|字段名|类型|说明|
|----|----|----|
|name|string|摄像头名称|

返回：

```json
{
    "code": 0,
    "msg": "ok",
    "data": [
         {
            "camera_id": "QN0011586a282eb0f5d9a_2",
            "name": "徐汇测试视频1"
        },
        {
            "camera_id": "QN0011586a282eb0f5d9a_1",
            "name": "onvif_ip45"
        }
    ]
}
```

|字段名|类型|说明|
|----|----|----|
|camera_id|string|摄像头id|
|name|string|摄像头名称|

### 3.3 获取设备id
GET: v1/vms/deviceid

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "deviceId": "5c73ff0c6042bb00016b5911"
    }
}
```

|字段名|类型|说明|
|----|----|----|
|deviceId|string|设备id|

### 3.4 新建子设备
POST: v1/vms/device/sub

请求参数

```json
{
	"type":1,
	"channel":1,
	"organization_id":"",
	"device_id":"5bbc7d33ea395772be628009",
	"attribute":{
		"account": "admin",
        "channel_type": 0,
        "description": "",
        "discovery_protocol": 1,
        "internal_channel": 0,
        "ip": "10.11.12.13",
        "name": "test1",
        "password": "admin",
        "upstream_url": "",
        "vendor": 1
	}
}

```

注意：device_id 可以通过 3.3 GET v1/vms/deviceid 来获取。

返回

```json
{
    "code":0,
    "data": {
        "id": "5c7e53707a71bc0001055bfe",
        "status": {
            "name": "离线",
            "value": 1
        },
        "channel": 1,
        "type": {
            "name": "枪机",
            "value": 1
        },
        "organization": {
            "id": "000000000000000000000000",
            "name": "默认分组"
        },
        "device_id": "5c78a2e57a71bc0001055bfb",
        "user_id": "5bac7e9eea395739c32bdc14",
        "attribute": {
            "name": "test1",
            "ip": "10.11.12.13",
            "discovery_protocol": {
                "name": "ONVIF",
                "value": 1
            },
            "description": "",
            "vendor": {
                "name": "海康",
                "value": 1
            },
            "account": "admin",
            "upstream_url": "",
            "channel_type": {
                "name": "",
                "value": 0
            },
            "internal_channel": 0
        }
    }
}

```

### 3.5 更新子设备
PUT: v1/vms/device/sub/:id

|字段名|类型|说明|
|----|----|----|
|id|string|子设备的id|

请求参数

```json
{
	"type":1,
	"channel":1,
	"organization_id":"",
	"device_id":"5bbc7d33ea395772be628009",
	"attribute":{
		"account": "admin",
        "channel_type": 0,
        "description": "",
        "discovery_protocol": 1,
        "internal_channel": 0,
        "ip": "10.11.12.13",
        "name": "test1",
        "password": "admin",
        "upstream_url": "",
        "vendor": 1
	}
}

```

返回

```json
{
    "code":0,
    "data": {
        "id": "5c7e53707a71bc0001055bfe",
        "status": {
            "name": "离线",
            "value": 1
        },
        "channel": 1,
        "type": {
            "name": "枪机",
            "value": 1
        },
        "organization": {
            "id": "000000000000000000000000",
            "name": "默认分组"
        },
        "device_id": "5c78a2e57a71bc0001055bfb",
        "user_id": "5bac7e9eea395739c32bdc14",
        "attribute": {
            "name": "test1",
            "ip": "10.11.12.13",
            "discovery_protocol": {
                "name": "ONVIF",
                "value": 1
            },
            "description": "",
            "vendor": {
                "name": "海康",
                "value": 1
            },
            "account": "admin",
            "upstream_url": "",
            "channel_type": {
                "name": "",
                "value": 0
            },
            "internal_channel": 0
        }
    }
}


```

### 3.6 删除子设备
DELETE: v1/vms/device/sub/:id

|字段名|类型|说明|
|----|----|----|
|id|string|子设备的id|


返回

```json
{
    "code":0,
    "data": {
        "id":"5bbcc8fdea39579e54c5a162"
    }
}

```

### 3.7 批量删除子设备
DELETE: v1/vms/device/subs

请求参数

```json
{
	"ids":["5c7e4d887a71bc0001055bfc"]
}

```
|字段名|类型|说明|
|----|----|----|
|ids|array|子设备的id数组|


返回

```json
{
    "code":0,
    "data": ["5c7e4d887a71bc0001055bfc"]
}

```

### 3.8 获取流地址
GET: v1/vms/sub/:id/play/url?stream_id=0&type=hflv

请求参数

|字段名|类型|说明|
|----|----|----|
|id|string|子设备的id|
|stream_id|int|码率类型, 0:主码流, 1:子码流|
|type|string|播放类型, hflv, hls, rtmp, close|

```json
{
    "code": 0,
    "data": "http://100.100.62.123:8088/live?port=1935&app=live&stream=QN0011586a282eb0f5d9a-2-0"
}
```


### 3.9 获取抓拍
GET: v1/vms/sub/:id/snap

|字段名|类型|说明|
|----|----|----|
|id|string|子设备的id|

```json
{
    "code": 0,
    "msg": "ok",
    "data": {
        "width": 2560,
        "height": 1440,
        "base64": "data:image/jpeg;base64,..."
    }
}
```

|字段名|类型|说明|
|----|----|----|
|width|int|宽度|
|height|int|高度|
|base64|string|base64图片|


### 3.10 获取通道数
GET: v1/vms/device/channels

返回

```json
{
    "code": 0,
    "data": [
        3,
        4,
        5,
        ...,
        128
    ]
}

```

### 3.11 获取类型枚举
GET: v1/vms/sub/type/enums

返回

```json
{
    "code": 0,
    "data": [
        {
            "name": "枪机",
            "value": 1
        },
        {
            "name": "球机",
            "value": 2
        }
    ]
}

```

### 3.12 获取厂商枚举
GET: v1/vms/sub/vendor/enums

返回

```json
{
    "code": 0,
    "data": [
         {
            "name": "海康",
            "value": 1
        },
        {
            "name": "大华",
            "value": 2
        },
        {
            "name": "宇视",
            "value": 3
        }
    ]
}

```

### 3.13 获取对接方式
GET: v1/vms/sub/discovery/protocol/:type

|字段名|类型|说明|
|----|----|----|
|type|string|类型值，1：枪机，2：球机|

返回

```json
{
    "code": 0,
    "data": [
        {
            "name": "ONVIF",
            "value": 1
        },
        {
            "name": "流地址",
            "value": 2
        },
        {
            "name": "SDK",
            "value": 3
        }
    ]
}

```

### 3.14 获取通道类型枚举
GET: v1/vms/sub/channeltype/enums

返回

```json
{
    "code": 0,
    "data": [
        {
            "name": "模拟",
            "value": 1
        },
        {
            "name": "数字",
            "value": 2
        }
    ]
}

```