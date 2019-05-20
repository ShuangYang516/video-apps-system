package proto

const (
	EventTypeAssembly    int = 1  //聚集
	EventTypeBanner      int = 2  //横幅
	EventTypeHeadCount   int = 11 //人流计数
	EventTypeHeadDensity int = 12 //客流密度
)

const EventTypeNone = 0

const EventTypeNonMotorizedOffset = 2200

//非机动车
const (
	EventTypeNonMotorChuangHongDeng = iota + EventTypeNonMotorizedOffset + 1
	EventTypeNonMotorNiXing
	EventTypeNonMotorTingCheYueXian
	EventTypeNonMotorZhanJiDongCheDao
	EventTypeNonMotorZhanRenXingDao
	EventTypeNonMotorWeiFanJinLingBiaoZhi
)

//机动车
const (
	EventTypeVehicleDaWanXiaoZhuan           = 2101 // 大弯小转
	EventTypeVehiclShiXianBianDao            = 2102 // 实线变道
	EventTypeVehicleBuAnDaoXiangXianXiangShi = 2106 // 不按导向线行驶
	EventTypeVehicleWangGeXianTingChe        = 2108 // 网格线停车
	EventTypeVehicleBuLiRangXingRen          = 2109 // 不礼让行人
)

const EventTypeConstructionOffset = 2300

// 施工监管
const (
	EventTypeConstructionChaoshi = iota + EventTypeConstructionOffset + 1 // 施工超时
	EventTypeConstructionChaoqi
	EventTypeConstructionChaoquyu
)

const (
	EventTypeNonMotor     = "non_motor"
	EventTypeVehicle      = "vehicle"
	EventTypeConstruction = "construction"
)

// 注意：增加非机动车和机动车事件类型，请同时更新此map !!!
var EventTypeMap = map[string][]int{
	EventTypeNonMotor: []int{
		// EventTypeNonMotorChuangHongDeng,
		EventTypeNonMotorNiXing,
		// EventTypeNonMotorTingCheYueXian,
		EventTypeNonMotorZhanJiDongCheDao,
		EventTypeNonMotorZhanRenXingDao,
		// EventTypeNonMotorWeiFanJinLingBiaoZhi,
	},
	EventTypeVehicle: []int{
		EventTypeVehicleDaWanXiaoZhuan,
		EventTypeVehiclShiXianBianDao,
		EventTypeVehicleBuAnDaoXiangXianXiangShi,
		// EventTypeVehicleWangGeXianTingChe,
		EventTypeVehicleBuLiRangXingRen,
	},
	EventTypeConstruction: []int{
		EventTypeConstructionChaoshi,
		EventTypeConstructionChaoqi,
		EventTypeConstructionChaoquyu,
	},
}

func MapEventType(t int) string {
	switch t {
	case EventTypeNonMotorChuangHongDeng:
		return "闯红灯"
	case EventTypeNonMotorNiXing:
		return "逆行行驶"
	case EventTypeNonMotorTingCheYueXian:
		return "停车越线"
	case EventTypeNonMotorZhanJiDongCheDao:
		return "在机动车道行驶"
	case EventTypeNonMotorZhanRenXingDao:
		return "在人行道行驶"
	case EventTypeNonMotorWeiFanJinLingBiaoZhi:
		return "违反禁令标志"
	case EventTypeVehicleDaWanXiaoZhuan:
		return "大弯小转"
	case EventTypeVehiclShiXianBianDao:
		return "违反禁止标线"
	case EventTypeVehicleBuAnDaoXiangXianXiangShi:
		return "不按导向车道行驶"
	case EventTypeVehicleWangGeXianTingChe:
		return "网格线停车"
	case EventTypeVehicleBuLiRangXingRen:
		return "不礼让行人"
	case EventTypeConstructionChaoshi:
		return "施工超时"
	case EventTypeConstructionChaoqi:
		return "施工超期"
	case EventTypeConstructionChaoquyu:
		return "施工超区域"
	}
	return "其他"
}
