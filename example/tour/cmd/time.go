package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
	"time"
	"tour/internal/timer"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "时间格式处理",
	Long:  "时间格式处理",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
var nowTimeCmd = &cobra.Command{
	Use:   "now",
	Short: "获取当前时间",
	Long:  "获取当前时间",
	Run: func(cmd *cobra.Command, args []string) {
		nowTime := timer.GetNowTime()
		log.Printf("output result: %s, %d", nowTime.Format("2006-01-02 15:04:05"), nowTime.Unix())
	},
}

var calculateTimeCmd = &cobra.Command{
	Use:   "calc",
	Short: "计算所需时间",
	Long:  "计算所需时间",
	Run: func(cmd *cobra.Command, args []string) {
		var currentTimer time.Time
		var layout = "2006-01-02 15:04:05" //格式化string
		location,_:=time.LoadLocation("Asia/Shanghai")
		//所有解析与格式化的操作都最好指定时区信息
		if calculateTime == "" {
			currentTimer = timer.GetNowTime()
		} else {
			var err error
			if !strings.Contains(calculateTime, " ") {
				layout = "2006-01-02"
			}
			currentTimer, err = time.ParseInLocation(layout, calculateTime, location)
			if err != nil {
				t, _ := strconv.Atoi(calculateTime)
				currentTimer = time.Unix(int64(t), 0).In(location)
			}
		}
		calculateTime, err := timer.GetCalculateTime(currentTimer, duration)
		if err != nil {
			log.Fatalf("timer.etCalculateTime err: %v", err)
		}
		log.Printf("output result: %s %d", calculateTime.In(location).Format(layout), calculateTime.Unix())
	},
}

var calculateTime string
var duration string

func init() {
	timeCmd.AddCommand(nowTimeCmd)
	timeCmd.AddCommand(calculateTimeCmd)

	calculateTimeCmd.Flags().StringVarP(&calculateTime, "calculate", "c", "", `需要计算的时间，有效单位为时间戳或已格式化后的时间`)
	calculateTimeCmd.Flags().StringVarP(&duration, "duration", "d", "", `持续时间，有效时间单位为"ns", "us" (or "μs"), "ms", "s", "m","h"`)
}
