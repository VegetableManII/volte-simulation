package controller

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"net"
	"sync"

	"github.com/VegetableManII/volte/modules"
	"github.com/wonderivan/logger"
)

var RandomAccess uint32 = 0x0F0F0F0F

type Ue struct {
	Vip  string
	UeID uint32
}

type EnodebEntity struct {
	TAI    string // AP接入点标识
	user   map[uint32]struct{}
	userMu sync.Mutex
}

func (e *EnodebEntity) Init() {
	e.user = make(map[uint32]struct{})
}

func (e *EnodebEntity) UeRandomAccess(data []byte, raddr *net.UDPAddr) (bool, []byte) {
	rand := parseRandAccess(data[0:4])
	if rand == RandomAccess {
		logger.Info("ue 随机接入 %x %x", rand, RandomAccess)
		h := fnv.New32()
		_, _ = h.Write([]byte(raddr.String()))
		sum := h.Sum(nil)
		ueid := parseRandAccess(sum)

		logger.Info("ueid(hex):%x ueid:%v", sum, ueid)

		e.userMu.Lock()
		e.user[uint32(ueid)] = struct{}{}
		e.userMu.Unlock()
		return true, sum
	}
	return false, nil
}

func (e *EnodebEntity) GenerateUpLinkData(data []byte, n int, mme, pgw string) ([]byte, string, error) {
	ueid := parseRandAccess(data[0:4])
	e.userMu.Lock()
	if _, ok := e.user[ueid]; !ok {
		return nil, "", errors.New("ErrNeedAccessInfo")
	}
	e.userMu.Unlock()

	dst := ""
	if data[4] == modules.EPCPROTOCAL { // EPC 消息
		binary.BigEndian.PutUint32(data[0:4], ueid)
		return data[0:n], dst, nil
	} else { // IMS 消息
		binary.BigEndian.PutUint32(data[0:4], ueid)
		dst = pgw
		return data[0:n], dst, nil
	}

}

func parseRandAccess(data []byte) uint32 {
	return binary.BigEndian.Uint32(data[0:4])
}

func (e *EnodebEntity) ParseDownLinkData(data []byte) {
	ueid := parseRandAccess(data[4:8])
	binary.BigEndian.PutUint32(data[0:4], ueid)
}
