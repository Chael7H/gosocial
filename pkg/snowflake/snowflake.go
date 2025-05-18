package snowflake

import (
	"fmt"
	"github.com/sony/sonyflake"
	"time"
)

// 雪花算法索尼版
var (
	sonyFlake     *sonyflake.Sonyflake //实例
	sonyMachineID uint16               //机器ID
)

func getMachineID() (uint16, error) { //返回全局定义的机器ID
	return sonyMachineID, nil
}

// Init 需传入当前机器ID
func Init(startTime string, machineID uint16) (err error) {
	sonyMachineID = machineID
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return err
	}
	settings := sonyflake.Settings{
		StartTime: st,
		MachineID: getMachineID,
	}
	sonyFlake = sonyflake.NewSonyflake(settings) //用配置生成sonyflake节点
	return nil
}

// GetID GenID生成Id
func GenID() (id int64, err error) {
	if sonyFlake == nil {
		err = fmt.Errorf("sonyFlake not initialized")
		return
	}
	ids, err := sonyFlake.NextID()
	id = int64(ids)
	return
}
