package pbft
//有关pbft相关数据类型的定义

// RequestMsg 请求消息数据类型
type RequestMsg struct {
	Timestamp  int64  `json:"timestamp"` //时间戳
	ClientID   string `json:"clientID"` //客户端id
	SequenceID int64  `json:"sequenceID"` //序列号
}

// ReplyMsg 结果返回消息数据类型
type ReplyMsg struct {
	ViewID    int64  `json:"viewID"` //视图id
	Timestamp int64  `json:"timestamp"` //时间戳
	ClientID  string `json:"clientID"` //客户端id
	NodeID    string `json:"nodeID"` //节点id
	Result    string `json:"result"` //结果
}

// PrePrepareMsg 预准备消息数据类型
type PrePrepareMsg struct {
	ViewID     int64       `json:"viewID"` //视图id
	SequenceID int64       `json:"sequenceID"` //序列号
	Digest     string      `json:"digest"` //消息概要（经过哈希处理）
	RequestMsg *RequestMsg `json:"requestMsg"` //请求消息
}

// VoteMsg 投票消息数据类型
type VoteMsg struct {
	ViewID     int64  `json:"viewID"` //视图id
	SequenceID int64  `json:"sequenceID"` //序列号
	Digest     string `json:"digest"` //消息概要（经哈希处理）
	NodeID     string `json:"nodeID"` //节点id
	MsgType    `json:"msgType"` //给出投票消息的消息类型
}

// MsgType 消息类型
type MsgType int
const (
	PrepareMsg MsgType = iota
	CommitMsg
)

// PBFT PBFT相关操作接口定义
type PBFT interface {
	StartConsensus(request *RequestMsg) (*PrePrepareMsg, error) //开始共识过程
	PrePrepare(prePrepareMsg *PrePrepareMsg) (*VoteMsg, error) //预准备过程
	Prepare(prepareMsg *VoteMsg) (*VoteMsg, error) //准备过程
	Commit(commitMsg *VoteMsg) (*ReplyMsg, *RequestMsg, error) //提交过程
}