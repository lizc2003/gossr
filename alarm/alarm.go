package alarm

import "github.com/lizc2003/gossr/common/util"

var gAlarm *util.RobotDingDing

func InitAlarm(env string, url string, secret string) {
	if url != "" {
		gAlarm = util.NewRobotDingDing(env, "gossr", url, secret)
	}
}

func SendMessage(msg string) {
	if gAlarm != nil {
		gAlarm.SendMsg(msg)
	}
}
