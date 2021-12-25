package common

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/wonderivan/logger"
)

/*
并发安全集合用来保存客户端连接
新连接接入之后保存对方的IP到集合中
当连接断开时没有从集合中删除???如何判断已经断开
*/
var clientMap *sync.Map

func ExchangeWithClient(ctx context.Context, conn *net.UDPConn, pre, post chan *Msg) {
	clientMap = new(sync.Map)
	data := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			// 释放资源
			// close(pre) // 关闭生产者通道
			logger.Warn("[%v] 信令交互协程退出", ctx.Value("Entity"))
			return
		default:
			n, remote, err := conn.ReadFromUDP(data)
			if err != nil {
				logger.Error("[%v] Server读取数据错误 %v", ctx.Value("Entity"), err)
			}
			if remote != nil || n != 0 {
				logger.Info("[%v] Read[%v] Data: %v", ctx.Value("Entity"), n, data[:n])
				distribute(data[:n], pre)
				// 检查该客户端是否已经开启线程服务
				if _, ok := clientMap.Load(remote); ok {
					continue
				} else {
					clientMap.Store(remote, ctx.Value("Entity"))
					go writeToClient(ctx, conn, remote, post)
				}
			} else {
				logger.Info("[%v] Remote[%v] Len[%v]", ctx.Value("Entity"), remote, n)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func writeToClient(ctx context.Context, conn *net.UDPConn, remote *net.UDPAddr, postConsumerC chan *Msg) {
	// 创建write buffer
	var buffer bytes.Buffer
	var n int
	for {
		select {
		case <-ctx.Done():
			// 释放资源
			logger.Warn("[%v] 发送信令至客户端协程退出", ctx.Value("Entity"))
			return
		case msg := <-postConsumerC:
			if msg.Type == 0x01 {
				err := binary.Write(&buffer, binary.BigEndian, msg.Data1)
				if err != nil {
					logger.Error("[%v] EpcMsg转化[]byte失败 %v", ctx.Value("Entity"), err)
					continue
				}
				n, err = conn.WriteToUDP(buffer.Bytes(), remote)
				if err != nil {
					logger.Error("[%v] EpcMsg广播消息发送失败 %v %v", ctx.Value("Entity"), err, buffer.Bytes())
				}
			} else {
				err := binary.Write(&buffer, binary.BigEndian, msg.Data2)
				if err != nil {
					logger.Error("[%v] SipMsg转化[]byte失败 %v", ctx.Value("Entity"), err)
					continue
				}
				n, err = conn.WriteToUDP(buffer.Bytes(), remote)
				if err != nil {
					logger.Error("[%v] SipMsg广播消息发送失败 %v %v", ctx.Value("Entity"), err, buffer.Bytes())
				}
			}
			logger.Info("[%v] Write to Client[%v] Data[%v]:%v", ctx.Value("Entity"), remote, n, buffer.Bytes())
			buffer.Reset()
		}
	}
}

// 采用分发订阅模式分发epc网络信令和sip信令
func distribute(data []byte, c chan *Msg) {
	if data[0] == 0x01 { // epc电路域协议
		msg := new(EpcMsg)
		msg.Init(data)
		c <- &Msg{
			Type:  0x01,
			Data1: msg,
		}
	} else { // ims网络域sip协议
		// todo
	}
}

// 主要用于基站和PGW实现消息的代理转发
func EnodebProxyMessage(ctx context.Context, src, dest *net.UDPConn) {
	for {
		select {
		case <-ctx.Done():
			logger.Warn("[%v] 基站转发协程退出...", ctx.Value("Entity"))
			return
		default:
			// 循环代理转发用户侧到网络侧消息
			n, err := io.Copy(dest, src) // 阻塞式,copy有deadline,如果src传输数据过快copy会一直进行复制
			if err != nil {
				logger.Error("[%v] 基站转发消息失败 %v %v", ctx.Value("Entity"), n, err)
			}
		}
	}
}
