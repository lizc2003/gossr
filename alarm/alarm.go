package alarm

import (
	"github.com/lizc2003/gossr/common/util"
	"strings"
)

var gAlarmDingding *util.RobotDingDing
var gAlarmFeishu *util.RobotFeishu

func InitAlarm(env string, url string, secret string) {
	if url != "" {
		app := "gossr"
		if strings.Index(url, "feishu") > 0 {
			gAlarmFeishu = util.NewRobotFeishu(env, app, url, secret)
		} else {
			gAlarmDingding = util.NewRobotDingDing(env, app, url, secret)
		}
	}
}

func SendMessage(msg string) {
	if gAlarmFeishu != nil {
		gAlarmFeishu.SendMsg(msg)
	} else if gAlarmDingding != nil {
		gAlarmDingding.SendMsg(msg)
	}
}
