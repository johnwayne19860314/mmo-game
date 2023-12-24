package main

import (
	"fmt"
	"time"
	"zinx-mmo-game/pb"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
)

func clientStartup(conn ziface.IConnection) {
	clietTalkData := &pb.Talk{
		Content: "hello",
	}
	data, err := proto.Marshal(clietTalkData)
	if err != nil {
		//err := fmt.Errorf("failed to marshal clientdata %v",clietTalkData)
		fmt.Println("failed to marshall data")
	}

	for {

		conn.SendMsg(2, data)
		time.Sleep(13 * time.Second)
	}
}

type clientHandle struct{
	znet.BaseRouter
}

func (c *clientHandle) Handle(req ziface.IRequest) {
	msgId := req.GetMsgID()
	fmt.Printf("client get a message , the msgid is %v \n", msgId)
	req.GetConnection().SendMsg(2, []byte("error message"))
	fmt.Println("client send an error message")
}

func main() {
	client := znet.NewClient("127.0.0.1", 8899)
	client.SetOnConnStart(clientStartup)
	client.AddRouter(200,&clientHandle{})
	client.Start()
	select {}
}
