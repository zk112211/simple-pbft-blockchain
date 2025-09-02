package RPC

import (
	"PBFTblockchain/pbft"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type View struct {
	ID      int64
	Primary string
}

type Node struct {
	NodeID        string
	NodeTable     map[string]string
	View          *View
	CurrentState  *pbft.State
    CommittedMsgs []*pbft.RequestMsg
	MsgBuffer     *MsgBuffer
	MsgEntrance   chan interface{}
	MsgDelivery   chan interface{}
	Alarm         chan bool
	RpcServer     *rpc.Server
}

type MsgBuffer struct {
	ReqMsgs        []*pbft.RequestMsg
	PrePrepareMsgs []*pbft.PrePrepareMsg
	PrepareMsgs    []*pbft.VoteMsg
	CommitMsgs     []*pbft.VoteMsg
}

const channelcapcity = 20

const ResolvingTimeDuration = time.Millisecond * 1000 //1秒

// NewNode 创建新的节点
func NewNode(nodeID string) *Node{
	const viewID = 10000000000 //创建一个临时的视图编号

	node := &Node{
		NodeID: nodeID,
		NodeTable: map[string]string{  //随便指定节点id
			"Apple":  "localhost:1111",
			"MS":     "localhost:1112",
			"Google": "localhost:1113",
			"IBM":    "localhost:1114",
		},
		View: &View{
			ID:      viewID,
			Primary: "Apple",
		},

		//初始化msg
		CurrentState:  nil,
		CommittedMsgs: make([]*pbft.RequestMsg, 0),
		MsgBuffer: &MsgBuffer{
			ReqMsgs:        make([]*pbft.RequestMsg, 0),
			PrePrepareMsgs: make([]*pbft.PrePrepareMsg, 0),
			PrepareMsgs:    make([]*pbft.VoteMsg, 0),
			CommitMsgs:     make([]*pbft.VoteMsg, 0),
		},

		// 初始化通道，增加缓冲改成异步方式，有问题
		MsgEntrance: make(chan interface{}, channelcapcity),
		MsgDelivery: make(chan interface{}, channelcapcity),
		Alarm:       make(chan bool),
		RpcServer:   rpc.NewServer(),
	}

	pbftService := &PBFTService{

	}

	node.RpcServer.Register(pbftService)

	//开始并发执行
	go node.dispatchMsg() // 阻塞接收消息

	go node.alarmToDispatcher() // 超时时钟响应

	// 开始处理信息
	go node.resolveMsg() // 处理接收的消息

	return node
}

func (n *Node)StartRPCService(address string){
	listener, err := net.Listen("tcp", address)
	if err != nil {
		// 处理错误
		return
	}

	// 启动服务器监听
	go n.RpcServer.Accept(listener)
}

// Broadcast 向其他节点广播
func (node *Node)Broadcast(Msg interface{}, path string) map[string]error{
    errorMap := make(map[string]error)

	for nodeId, url := range node.NodeTable {
		if nodeId == node.NodeID{
			continue //不广播给自己
		}

		jsonMsg, err := json.Marshal(Msg)
		if err != nil {
			errorMap[nodeId] = err
			continue
		}
		send(url+path, jsonMsg) //发送下一条请求
	}

	if len(errorMap) == 0{
		return  nil
	} else{
		return  errorMap
	}
}

//回复消息
func (node *Node)Reply(Msg *pbft.ReplyMsg) error{
	//打印所有提交信息
	for _, value := range node.CommittedMsgs {
		fmt.Printf("Committed value: %s, %d, %s, %d", value.ClientID, value.Timestamp, value.SequenceID)
	}
	fmt.Print("\n")

	//读取json文件
	jsonMsg, err := json.Marshal(Msg)
	if err != nil {
		return err
	}

	// 将信息发送给Primary节点
	sendmsgtoclient(node.NodeTable[node.View.Primary], jsonMsg)
	//将节点的状态归位
	return nil
}

//得到请求
func (node *Node)GetReq(reqMsg *pbft.RequestMsg) error{
	LogMsg(reqMsg)

	// 为新共识过程创建新的状态
	err := node.createStateForNewConsensus()
	if err != nil {
		return err
	}

	// 开始共识过程
	prePrepareMsg, err := node.CurrentState.StartConsensus(reqMsg)
	if err != nil {
		return err
	}

	LogStage(fmt.Sprintf("Consensus Process (ViewID:%d)", node.CurrentState.ViewID), false)

	//发送预准备信息
	if prePrepareMsg != nil {
		fmt.Println("广播预准备信息")
		node.Broadcast(prePrepareMsg, "/preprepare")
		LogStage("Pre-prepare", true)
	}

	return nil
}

//预准备
func (node *Node) GetPrePrepared(prePrepareMsg *pbft.PrePrepareMsg) error{
	LogMsg(prePrepareMsg)

	// 创建一个新的共识状态
	err := node.createStateForNewConsensus()
	if err != nil {
		return err
	}

	//预准备信息
	prePareMsg, err := node.CurrentState.PrePrepare(prePrepareMsg)
	if err != nil {
		return err
	}

	if prePareMsg != nil {
		// 将节点id附在消息中
		prePareMsg.NodeID = node.NodeID //当前节点进入prepare 状态，广播prepare消息

		LogStage("Pre-prepare", true)
		fmt.Println("广播准备信息")
		node.Broadcast(prePareMsg, "/prepare")
		LogStage("Prepare", false)
	}

	return nil
}

//准备
func (node *Node)GetPrepared(prepareMsg *pbft.VoteMsg)error{
	LogMsg(prepareMsg)

	//准备消息
	commitMsg, err := node.CurrentState.Prepare(prepareMsg)
	if err != nil {
		return err
	}

	if commitMsg != nil {
		//将节点放在消息中
		commitMsg.NodeID = node.NodeID

		LogStage("Prepare", true)
		fmt.Println("收到准备消息，commit提交，广播commit")
		node.Broadcast(commitMsg, "/commit")
		LogStage("Commit", false) //广播了一个commit出去
	}

	return nil
}

//提交
func (node *Node)GetCommitted(commitMsg *pbft.VoteMsg)error{
	LogMsg(commitMsg)

	//得到回复消息和提交消息
	replyMsg, committedMsg, err := node.CurrentState.Commit(commitMsg)
	if err != nil { //错误返回
		fmt.Println("err1")
		return err
	}

	if replyMsg != nil {
		//错误返回
		if committedMsg == nil {
			fmt.Println("committed message is nil, even though the reply message is not nil")
			return errors.New("committed message is nil, even though the reply message is not nil")
		}

		//将节点id加到消息中
		replyMsg.NodeID = node.NodeID

		// 保存节点最后的提交消息
		node.CommittedMsgs = append(node.CommittedMsgs, committedMsg)

		LogStage("Commit", true)
		fmt.Println("收到commit消息，reply给客户端")
		node.Reply(replyMsg)
		LogStage("Reply", true)
		node = NewNode(node.NodeID)
	}

	return nil
}

//回复
func (node *Node)GetReply(msg *pbft.ReplyMsg){
	fmt.Printf("Result: %s by %s\n", msg.Result, msg.NodeID)
}

//为新共识创造状态
func (node *Node)createStateForNewConsensus()error{
	var lastSequenceID int64
	if len(node.CommittedMsgs) == 0 {
		lastSequenceID = -1
	} else {
		lastSequenceID = node.CommittedMsgs[len(node.CommittedMsgs)-1].SequenceID
	}

	//在主节点中为新的共识创造新的状态
	node.CurrentState = pbft.CreateState(node.View.ID, lastSequenceID)

	LogStage("Create the replica status", true)

	return nil
}

//分派数据
func (node *Node)dispatchMsg(){
	for {
		select {
		case msg := <-node.MsgEntrance: //接收来自通道的数据
			fmt.Println("收到http请求") // 需要解决通道的阻塞问题
			err := node.routeMsg(msg)
			if err != nil {
				fmt.Println(err)
				// TODO: send err to ErrorChannel
			}
		case <-node.Alarm:
			//channel超时，报错
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err)
				fmt.Println("!!!!!!!!!!case <-node.Alarm:!!!!!!!!!!!!")
				// TODO: send err to ErrorChannel
			}
		}
	}

}

//线路发送信息
func (node *Node)routeMsg(msg interface{}) []error{
	//对不同的消息类型进行处理，操作基本相同，只是数据类型不同
	switch msg.(type) {
	case *pbft.RequestMsg:
		if node.CurrentState == nil || node.CurrentState != nil {
			//拷贝缓存消息
			msgs := make([]*pbft.RequestMsg, len(node.MsgBuffer.ReqMsgs))
			copy(msgs, node.MsgBuffer.ReqMsgs)

			//将新到达的信息附加到msgs
			msgs = append(msgs, msg.(*pbft.RequestMsg))

			// 清理缓存
			node.MsgBuffer.ReqMsgs = make([]*pbft.RequestMsg, 0)

			// 发送消息
			fmt.Println("转发请求消息")
			node.MsgDelivery <- msgs
		} else {
			fmt.Println("当前状态不为空，节点状态未归位")
			node.MsgBuffer.ReqMsgs = append(node.MsgBuffer.ReqMsgs, msg.(*pbft.RequestMsg))
		}
	case *pbft.PrePrepareMsg:
		if node.CurrentState == nil || node.CurrentState != nil {
			// 拷贝缓存消息
			msgs := make([]*pbft.PrePrepareMsg, len(node.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.MsgBuffer.PrePrepareMsgs)

			// 附加新到达的消息
			msgs = append(msgs, msg.(*pbft.PrePrepareMsg))

			// 清空缓存
			node.MsgBuffer.PrePrepareMsgs = make([]*pbft.PrePrepareMsg, 0)

			// 发送消息
			fmt.Println("转发预准备消息")
			node.MsgDelivery <- msgs
		} else {
			node.MsgBuffer.PrePrepareMsgs = append(node.MsgBuffer.PrePrepareMsgs, msg.(*pbft.PrePrepareMsg))
		}
	case *pbft.VoteMsg:
		if msg.(*pbft.VoteMsg).MsgType == pbft.PrepareMsg {
			if node.CurrentState == nil || node.CurrentState.CurrentStage != pbft.PrePrepared {
				node.MsgBuffer.PrepareMsgs = append(node.MsgBuffer.PrepareMsgs, msg.(*pbft.VoteMsg))
			} else {
				//缓存消息
				msgs := make([]*pbft.VoteMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)

				// 附加新到达消息达msgs
				msgs = append(msgs, msg.(*pbft.VoteMsg))

				//清空缓存
				node.MsgBuffer.PrepareMsgs = make([]*pbft.VoteMsg, 0)

				// 发送消息
				fmt.Println("收到准备消息后开始回复，投票过程")
				node.MsgDelivery <- msgs
			}
		} else if msg.(*pbft.VoteMsg).MsgType == pbft.CommitMsg {
			// Copy buffered messages first.
			msgs := make([]*pbft.VoteMsg, len(node.MsgBuffer.CommitMsgs))
			copy(msgs, node.MsgBuffer.CommitMsgs)

			// Append a newly arrived message.
			msgs = append(msgs, msg.(*pbft.VoteMsg))

			// Empty the buffer.
			node.MsgBuffer.CommitMsgs = make([]*pbft.VoteMsg, 0)

			// Send messages.
			fmt.Println("收到commit提交消息")
			node.MsgDelivery <- msgs
		}
	}

	return nil

}

//当警告时，线路发送信息
func (node *Node)routeMsgWhenAlarmed() []error{
	if node.CurrentState == nil {
		// 检查请求消息并发送
		if len(node.MsgBuffer.ReqMsgs) != 0 {
			msgs := make([]*pbft.RequestMsg, len(node.MsgBuffer.ReqMsgs))
			copy(msgs, node.MsgBuffer.ReqMsgs)

			node.MsgDelivery <- msgs
		}

		// 检查预准备消息并发送
		if len(node.MsgBuffer.PrePrepareMsgs) != 0 {
			msgs := make([]*pbft.PrePrepareMsg, len(node.MsgBuffer.PrePrepareMsgs))
			copy(msgs, node.MsgBuffer.PrePrepareMsgs)

			node.MsgDelivery <- msgs
		}
	} else {
		switch node.CurrentState.CurrentStage {
		case pbft.PrePrepared:
			// C检查准备消息并发送
			if len(node.MsgBuffer.PrepareMsgs) != 0 {
				msgs := make([]*pbft.VoteMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)

				node.MsgDelivery <- msgs
			}
		case pbft.Prepared:
			// 检查提交消息并发送
			if len(node.MsgBuffer.CommitMsgs) != 0 {
				msgs := make([]*pbft.VoteMsg, len(node.MsgBuffer.CommitMsgs))
				copy(msgs, node.MsgBuffer.CommitMsgs)

				node.MsgDelivery <- msgs
			}
		}
	}

	return nil

}

//处理消息
func (node *Node) resolveMsg() {
	for {
		//从分发器中取得请求消息
		msgs := <-node.MsgDelivery
		fmt.Println("处理请求消息")
		switch msgs.(type) {
		case []*pbft.RequestMsg:
			errs := node.resolveRequestMsg(msgs.([]*pbft.RequestMsg)) //request - preprepare
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err)
				}
				// TODO: send err to ErrorChannel
			}
		case []*pbft.PrePrepareMsg:
			fmt.Println("处理预准备消息")
			errs := node.resolvePrePrepareMsg(msgs.([]*pbft.PrePrepareMsg)) //preprepare-prepare
			if len(errs) != 0 {
				for _, err := range errs {
					fmt.Println(err)
				}
				// TODO: send err to ErrorChannel
			}
		case []*pbft.VoteMsg:
			voteMsgs := msgs.([]*pbft.VoteMsg)
			if len(voteMsgs) == 0 {
				break
			}

			if voteMsgs[0].MsgType == pbft.PrepareMsg {
				fmt.Println("处理准备投票消息")
				errs := node.resolvePrepareMsg(voteMsgs) // 收到prepare 开始commit
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err)
					}
					// TODO: send err to ErrorChannel
				}
			} else if voteMsgs[0].MsgType == pbft.CommitMsg {
				fmt.Println("收到并处理commit消息")
				errs := node.resolveCommitMsg(voteMsgs) //收到commit 判断大于 2 ，开始reply
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err)
					}
					// TODO: send err to ErrorChannel
				}
			}
		}
	}
}

//分发器超时警告
func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

//处理请求消息
func (node *Node) resolveRequestMsg(msgs []*pbft.RequestMsg) []error {
	fmt.Println("处理客户端请求")
	errs := make([]error, 0)

	// 处理消息
	for _, reqMsg := range msgs {
		err := node.GetReq(reqMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

//处理预准备消息
func (node *Node) resolvePrePrepareMsg(msgs []*pbft.PrePrepareMsg) []error {
	errs := make([]error, 0)

	// 处理消息
	for _, prePrepareMsg := range msgs {
		err := node.GetPrePrepared(prePrepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

//处理准备消息
func (node *Node) resolvePrepareMsg(msgs []*pbft.VoteMsg) []error {
	errs := make([]error, 0)

	// 处理消息
	for _, prepareMsg := range msgs {
		err := node.GetPrepared(prepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

//处理提交消息
func (node *Node) resolveCommitMsg(msgs []*pbft.VoteMsg) []error {
	errs := make([]error, 0)

	// 处理消息
	for _, commitMsg := range msgs {
		err := node.GetCommitted(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

//向客户端发送信息
func sendmsgtoclient(clientURL string, msg []byte) error {
	// 创建请求
	req, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(msg))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("非预期的状态码: %d", resp.StatusCode)
	}

	// 请求成功
	return nil
}

// send 向指定的 URL 发送 JSON 编码的消息
func send(url string, msg []byte) error {
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(msg))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("非预期的状态码: %d", resp.StatusCode)
	}

	// 请求成功
	return nil
}
