package core

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"math/rand"
	"sync"
	"zinx-mmo-game/pb"

	"github.com/golang/protobuf/proto"
)

type Player struct {
	Pid  int32              // 用户ID
	Conn ziface.IConnection // 当前玩家的连接
	X    float32            // 平面X坐标
	Y    float32            // 高度
	Z    float32            // 平面y坐标
	V    float32            // 旋转0-360角度
}

var PidGen int32 = 1  // 用户ID生成器
var IdLock sync.Mutex // 保护PidGen的互斥锁

func NewPlayer(conn ziface.IConnection) *Player {
	// 生成一个PID
	IdLock.Lock()
	id := PidGen
	PidGen++
	IdLock.Unlock()

	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), // 随机在160坐标点 基于X轴偏移若干坐标
		Y:    0,
		Z:    float32(140 + rand.Intn(20)), // 随机在140坐标点 基于Y轴偏移若干坐标
		V:    0,
	}

	return p
}

func (p *Player) SendMsg(msgId int32, data proto.Message) {
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("marshal data err", err)
		return
	}

	if p.Conn == nil {
		fmt.Println("conn is closed")
		return
	}

	if err := p.Conn.SendMsg(uint32(msgId), msg); err != nil {
		fmt.Println("sendMsg err", err)
		return
	}

	return
}

func (p *Player) SyncPid() {
	data := &pb.SyncPid{
		Pid: p.Pid,
	}

	p.SendMsg(1, data)
}

func (p *Player) BroadCastStartPosition() {
	data := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	p.SendMsg(200, data)
}

// 玩家广播世界消息
func (p *Player) Talk(content string) {
	// 1 组件 MsgID:200 proto数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	// 2 获取所有玩家
	players := WorldManagerObj.GetAllPlayers()

	// 3 向所有玩家（包括自己）发送MsgID:200消息
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}
}

func (p *Player) SyncSurrounding() {

	// 获取周边玩家ID
	pids := WorldManagerObj.AOIManager.GetPidsByPos(p.X, p.Z)
	// 根据pid得到所有玩家对象
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldManagerObj.GetPlayerByPid(int32(pid)))
	}

	// 1 同步自己位置到周围玩家
	msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //广播
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Z: p.Z,
				Y: p.Y,
				V: p.V,
			},
		},
	}
	for _, player := range players {
		player.SendMsg(200, msg)
	}

	// 2 同步周围玩家在自己的视野中
	playersData := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		p := &pb.Player{
			Pid: player.Pid,
			P: &pb.Position{
				X: player.X,
				Z: player.Z,
				Y: player.Y,
				V: player.V,
			},
		}
		playersData = append(playersData, p)
	}

	syncPlayers := &pb.SyncPlayers{
		Ps: playersData[:],
	}
	p.SendMsg(202, syncPlayers)
}
