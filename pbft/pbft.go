package pbft
//pbft相关函数操作的实现

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// State pbft状态数据结构
type State struct {
	ViewID         int64  //视图id
	MsgLogs        *MsgLogs //储存信息
	LastSequenceID int64 //上一序列号
	CurrentStage   Stage //当前阶段
}

// MsgLogs 储存所传输的所有信息
type MsgLogs struct {
	ReqMsg      *RequestMsg //客户端请求信息
	PrepareMsgs map[string]*VoteMsg //准备阶段信息
	CommitMsgs  map[string]*VoteMsg //提交阶段信息
}

const f = 1

type Stage int //阶段数据类型定义

//几种阶段的定义
const(
	Idle       Stage = iota //空闲阶段，刚创建状态，还未接收到请求
	PrePrepared  //预准备完成
	Prepared  //准备完成
	Committed  //提交完成
)

func CreateState(ViewId int64, lastSequenceId int64) *State {
	//创建并返回State
	return &State{
		ViewID: ViewId,
		MsgLogs: &MsgLogs{
			ReqMsg: nil, //空请求信息
			PrepareMsgs: make(map[string]*VoteMsg), //创建空的VoteMsg的map
			CommitMsgs: make(map[string]*VoteMsg),
		},
		LastSequenceID: lastSequenceId,
		CurrentStage: Idle, //空闲阶段
	}
}

// StartConsensus 开始共识过程
func (state *State) StartConsensus(request *RequestMsg) (*PrePrepareMsg, error){
    sequenceID := time.Now().UnixNano() //用UnixNano生成纳秒级的系列号，保证序列号的唯一性
    if sequenceID != -1{
    	for state.LastSequenceID >= sequenceID {
    		sequenceID += 1
		}
	}
    request.SequenceID = sequenceID

    state.MsgLogs.ReqMsg = request //将请求传给共识state中MsgLogs中的ReqMsg
    digest, err := digest(request)
    if err != nil{
    	return nil, err
	}

	//成功开始，返回预准备信息,进入预准备状态
	state.CurrentStage = PrePrepared
	return &PrePrepareMsg{
		ViewID:     state.ViewID,
		SequenceID: sequenceID,
		Digest:     digest,
		RequestMsg: request,
	}, err
}

// PrePrepare 预准备过程
func (state *State)PrePrepare(prePrepareMsg *PrePrepareMsg) (*VoteMsg, error){
	state.MsgLogs.ReqMsg = prePrepareMsg.RequestMsg //将reqMsg装入prePrepareMsg中

	//确认信息，若错误则返回nil并报错
	if !state.verifyMsg(prePrepareMsg.ViewID, prePrepareMsg.SequenceID, prePrepareMsg.Digest){
		return nil, errors.New("pre-prepare massage is corrupted")
	}
	state.CurrentStage = PrePrepared //阶段被设置为预准备完成

	//返回一个投票信息
	return &VoteMsg{
		ViewID:     prePrepareMsg.ViewID,
		SequenceID: prePrepareMsg.SequenceID,
		Digest:     prePrepareMsg.Digest,
		MsgType:    PrepareMsg,
	}, nil
}

// Prepare 准备过程
func (state *State)Prepare(prepareMsg *VoteMsg) (*VoteMsg, error){
	//确认信息，若错误返回nil并报错
	if !state.verifyMsg(prepareMsg.ViewID, prepareMsg.SequenceID, prepareMsg.Digest){
		return nil, errors.New("prepare massage is corrupted")
	}

	//将prepareMsg加入到MsgLogs相应模块中
    state.MsgLogs.PrepareMsgs[prepareMsg.NodeID] = prepareMsg

    //打印当前的投票信息
    fmt.Printf("[Prepare-vote]: %d\n",len(state.MsgLogs.PrepareMsgs))

    //将当前阶段设置为准备完成
    state.CurrentStage = Prepared

    return &VoteMsg{
		ViewID:     prepareMsg.ViewID,
		SequenceID: prepareMsg.SequenceID,
		Digest:     prepareMsg.Digest,
		MsgType:    CommitMsg,
	}, nil
}

//提交过程
func (state *State)Commit(commitMsg *VoteMsg) (*ReplyMsg, *RequestMsg, error) {
    //验证消息准确性
	if !state.verifyMsg(commitMsg.ViewID, commitMsg.SequenceID, commitMsg.Digest){
		return nil, nil, errors.New("commit massage is corrupted")
	}

	//将commitMsg加入到MsgLog相应模块中
	state.MsgLogs.CommitMsgs[commitMsg.NodeID] = commitMsg

	//正确投票数
	votenum := len(state.MsgLogs.CommitMsgs)
	fmt.Printf("[Commit-num]: %d\n", votenum)

	//若正确投票达到提交标准，执行相关操作
	if state.committed(votenum){
        result := "Executed"

        //将当前阶段设置为已提交
        state.CurrentStage = Committed
        fmt.Printf("Commited, reply")

        //返回ReplyMsg
        return &ReplyMsg{
        	ViewID:  state.ViewID,
        	Timestamp: state.MsgLogs.ReqMsg.Timestamp,
        	ClientID: state.MsgLogs.ReqMsg.ClientID,
        	Result: result,
		}, state.MsgLogs.ReqMsg, nil
	}

	//未达到提交标准，返回空结果
	return nil, nil, nil

}

//生成内容摘要，并生成哈希值，用于验证信息正确性
func digest(object interface{}) (string, error) {
	//将内容转化为json格式的字节序列
    msg, err := json.Marshal(object)

    if err != nil{
    	return "", err
	}

	//对内容摘要进行哈希，用于验证信息正确性
    return hash(msg), err
}

//计算哈希值，并以字符串的形式返回
func hash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed)
}

//确认信息
func (state *State)verifyMsg(viewID int64, sequenceID int64, digestGot string) bool{
	//检查视图编号是否正确
	if state.ViewID != viewID{
		return false
	}
		return false

	//检查上一序列号是否正确
	if state.LastSequenceID != -1 {
		if state.LastSequenceID >= sequenceID {
			return false
		}
	}

	digest, err := digest(state.MsgLogs.ReqMsg)
	if err != nil{
		fmt.Println(err)
		return false
	}

	//检查得到的内容摘要哈希值是否正确，若不正确，返回false
	if digestGot != digest{
		return false
	}
	return true
}

//检验是否打到提交的标准，也就是确认消息数达到2f
func (state *State)committed(count int) bool{
	if count < 2*f{
		return false
	}

	return true
}
