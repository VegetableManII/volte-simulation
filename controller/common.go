package controller

import (
	"context"
	"volte/common"
)

const (
	HSS_RESP_TOKEN string = "AUTN"
	HSS_RESP_XRES  string = "XRES"
	HSS_RESP_RAND  string = "RAND"
	HSS_RESP_KASME string = "Kasme"
)

// 定义基础路由转发方法
type BaseSignallingT func(context.Context, *common.Msg, chan *common.Msg) error

// 路由转发器
type Mux struct {
	router map[[2]byte]BaseSignallingT
	// mu     sync.Mutex
}

// 路由注册
func (m *Mux) Regist(r [2]byte, f BaseSignallingT) {
	if m.router == nil {
		m.router = make(map[[2]byte]BaseSignallingT)
	}
	m.router[r] = f
}

// VoLTE网络中各个功能实体的逻辑处理器实体抽象基类对象
type Base interface {
	CoreProcessor(context.Context, chan *common.Msg, chan *common.Msg)
}
