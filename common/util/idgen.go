package util

import (
	"github.com/lizc2003/gossr/common/tlog"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const IDGEN_BASE_TIME = 1562860800

type IdGen struct {
	serverIdVal int64
	lowBits     uint
	maxOffset   int
	offset      int
	lastEpoch   int64
	mutex       sync.Mutex
}

func NewIdGen(serverId int) (*IdGen, error) {
	return NewIdGenWithBits(serverId, 5, 21)
}

func NewBigIdGen(serverId int) (*IdGen, error) {
	return NewIdGenWithBits(serverId, 16, 32)
}

func NewIdGenWithBits(serverId int, serverIdBits int, lowBits int) (*IdGen, error) {
	maxId := 1 << uint(serverIdBits)
	if serverId < 0 || serverId >= maxId || serverIdBits <= 0 || serverIdBits >= lowBits {
		return nil, errors.New("serverId invalid")
	}

	serverShiftBits := uint(lowBits - serverIdBits)
	return &IdGen{
		serverIdVal: int64(serverId) << serverShiftBits,
		lowBits:     uint(lowBits),
		maxOffset:   1 << serverShiftBits,
		offset:      0,
		lastEpoch:   0,
	}, nil
}

func (this *IdGen) GenId() int64 {
	return this.genIdByEpoch(time.Now().Unix())
}

func (this *IdGen) GenStringId() string {
	n := this.genIdByEpoch(time.Now().Unix())
	return strconv.FormatInt(n, 16)
}

func (this *IdGen) GenShortStringId() string {
	n := this.genIdByEpoch(time.Now().Unix())
	return strconv.FormatInt(n, 36)
}

func (this *IdGen) genIdByEpoch(epoch int64) int64 {
	var offset int
	this.mutex.Lock()
	if epoch < this.lastEpoch {
		tlog.Warningf("clock is back: %d from previous: %d", epoch, this.lastEpoch)
		epoch = this.lastEpoch
	} else if epoch > this.lastEpoch {
		this.lastEpoch = epoch
		this.offset = 0
	}
	this.offset++
	offset = this.offset
	this.mutex.Unlock()

	if offset >= this.maxOffset {
		tlog.Warningf("maximum id reached in 1 second in epoch: %d", epoch)
		return this.genIdByEpoch(epoch + 1)
	}

	return ((epoch - IDGEN_BASE_TIME) << this.lowBits) | this.serverIdVal | int64(offset)
}

func IdGenTest() {
	idGen, err := NewIdGenWithBits(1, 5, 21)
	if err == nil {
		fmt.Println(idGen)
		for k := 0; k < 30; k++ {
			go func() {
				idmap := make(map[int64]bool)
				for i := 0; i < 531075; i++ {
					id := idGen.GenId()
					idmap[id] = true
				}
				fmt.Println(len(idmap))
			}()
		}

		time.Sleep(5 * time.Second)
	} else {
		fmt.Println(err)
	}
}
