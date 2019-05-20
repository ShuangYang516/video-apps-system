package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"qbox.us/cc/config"
	"qiniu.com/vas-app/biz/export/client"
)

var rootCmd = &cobra.Command{
	Use:   "export [filepath]",
	Short: "export events to file folder",
	Long:  `export events to file folder`,
	Args:  cobra.ExactArgs(1),
	Run:   rootRun,
}

var (
	endpoint string
	start    string
	end      string

	typo string

	cameraId  string
	class     int
	eventType int
	limit     int

	marking string

	hasLabel   int
	labelScore float64

	hasFace    int
	similarity float64

	fontfile string

	appName = "export"
	conf    Config
)

type Config struct {
	Devices map[string]DeviceConfig `json:"devices"`
}

type DeviceConfig struct {
	Position string `json:"position"`
	Addr     string `json:"addr"`
	SBBH     string `json:"sbbh"`
}

func init() {
	rootCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "vas-app dashboard http address (required)")
	rootCmd.Flags().StringVarP(&start, "start", "s", "", "export start time: format \"20060102150405\" (required)")
	rootCmd.Flags().StringVarP(&end, "end", "d", "", "export end time: format \"20060102150405\" (required)")

	rootCmd.Flags().StringVarP(&typo, "type", "t", "", "export type: non_motor|vehicle")

	rootCmd.Flags().StringVarP(&cameraId, "cameraId", "i", "", "export cameraId")
	rootCmd.Flags().IntVarP(&class, "class", "c", 0, "export class: 3:饿了么 4:美团 ")
	rootCmd.Flags().IntVarP(&eventType, "eventType", "", 0, `export eventType: 违法类型，
	非机动车-2201:闯红灯 2202:逆行  2203:停车越线 2204:非机动车占用机动车道 2205:非机动车占用人行道；
	机动车-2101:机动车大弯小转 2102:实线变道 2106:不按导向线行驶 2108:网格线停车 2109:不礼让行人`)
	rootCmd.Flags().IntVarP(&limit, "limit", "l", 0, "export limit")

	rootCmd.Flags().StringVarP(&marking, "marking", "m", "", "打标状态 init:原始状态 illegal:违规 discard:作废")

	rootCmd.Flags().IntVarP(&hasLabel, "hasLabel", "", 0, "是否包含标牌, 1:包含 2:不包含")
	rootCmd.Flags().Float64VarP(&labelScore, "labelScore", "", 0, "lable score: 0-1")

	rootCmd.Flags().IntVarP(&hasFace, "hasFace", "", 0, "是否包含人脸, 1:包含 2:不包含")
	rootCmd.Flags().Float64VarP(&similarity, "similarity", "", 0, "face similarity: 0-100")

	rootCmd.Flags().StringVarP(&fontfile, "fontfile", "f", "./font/PingFang-SC-Regular.ttf", "filepath of fontfile(.ttf format)")

	if err := rootCmd.MarkFlagRequired("endpoint"); err != nil {
		fmt.Println("endpoint is required")
		os.Exit(2)
	}
	if err := rootCmd.MarkFlagRequired("start"); err != nil {
		fmt.Println("start is required")
		os.Exit(2)
	}
	if err := rootCmd.MarkFlagRequired("end"); err != nil {
		fmt.Println("end is required")
		os.Exit(2)
	}
	// if err := rootCmd.MarkFlagRequired("type"); err != nil {
	// 	fmt.Println("type is required")
	// 	os.Exit(2)
	// }
}

func rootRun(_ *cobra.Command, args []string) {
	var (
		filePath string
		err      error
	)

	// load default config
	if e := config.LoadEx(&conf, "./conf/"+appName+".conf"); e != nil {
		log.Fatal("config.Load failed:", e)
	}
	buf, _ := json.MarshalIndent(conf, "", "    ")
	log.Printf("loaded conf \n%s", string(buf))

	filePath = args[0]
	fmt.Println("reading from", filePath)

	cli, err := client.New(endpoint)
	if err != nil {
		fmt.Printf("client.New(%s): %v\n", endpoint, err)
		return
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Printf("time.LoadLocation(\"Local\"): %v\n", err)
		return
	}

	startTime, err := time.ParseInLocation("20060102150405", start, loc)
	if err != nil {
		fmt.Printf("time.ParseInLocation(%s): %v\n", start, err)
		return
	}
	endTime, err := time.ParseInLocation("20060102150405", end, loc)
	if err != nil {
		fmt.Printf("time.ParseInLocation(%s): %v\n", end, err)
		return
	}

	req := &client.GetEventReq{
		Start:     int(startTime.Unix() * 1000),
		End:       int(endTime.Unix() * 1000),
		CameraId:  cameraId,
		Type:      typo,
		EventType: eventType,
		Limit:     limit,
		Class:     class,

		Marking: marking,

		HasLabel:   hasLabel,
		LabelScore: labelScore,

		HasFace:    hasFace,
		Similarity: similarity,
	}

	resp, err := cli.GetEvents(req)
	if err != nil {
		fmt.Printf("cli.GetEvents(%#v): %v\n", req, err)
		return
	}
	if resp.Code != 0 {
		fmt.Printf("cli.GetEvents(%#v): %s\n", req, resp.Msg)
		return
	}

	list := resp.Data
	ProcessEvents(list, conf.Devices, filePath)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
