package RPC

import (
	"PBFTblockchain/blockchain"
	"PBFTblockchain/pbft"
	"net/rpc"
)

type PBFTService struct {
	state pbft.PBFT
	bc    *blockchain.BlockChain
	node  *Node
}

// RequestArgs 定义了接收请求的参数结构
type RequestArgs struct {
	Data string // 假设交易数据以序列化的字符串形式传递
}

// Response 定义了 RPC 响应的结构
type Response struct {
	Ack bool // 确认响应
}

// NewPBFTService 创建并返回一个初始化的 PBFTService 实例
func NewPBFTService(bc *blockchain.BlockChain, viewID int64, lastSequenceID int64) *PBFTService {
	// 使用提供的参数创建 PBFT 状态
	state := pbft.CreateState(viewID, lastSequenceID)

	// 返回 PBFTService 实例
	return &PBFTService{
		bc:    bc,
		state: state,
	}
}

// StartConsensusRPC 是 StartConsensus 方法的 RPC 版本
func (p *PBFTService) StartConsensusRPC(args *pbft.RequestMsg, reply *pbft.PrePrepareMsg) error {
	var err error
	reply, err = p.state.StartConsensus(args)
	return err
}

// PrePrepareRPC 是 PrePrepare 方法的 RPC 版本
func (p *PBFTService) PrePrepareRPC(args *pbft.PrePrepareMsg, reply *pbft.VoteMsg) error {
	var err error
	reply, err = p.state.PrePrepare(args)
	return err
}

// PrepareRPC 是 Prepare 方法的 RPC 版本
func (p *PBFTService) PrepareRPC(args *pbft.VoteMsg, reply *pbft.VoteMsg) error {
	var err error
	reply, err = p.state.Prepare(args)
	return err
}

// CommitRPC 是 Commit 方法的 RPC 版本
func (p *PBFTService) CommitRPC(args *pbft.VoteMsg, reply *pbft.ReplyMsg) (*pbft.RequestMsg, error) {
	var requestMsg *pbft.RequestMsg
	var err error
	reply, requestMsg, err = p.state.Commit(args)
	// 处理 requestMsg，如果需要
	return requestMsg, err
}

// RegisterPBFTService 注册RPC服务
func RegisterPBFTService(srv *rpc.Server, pbftState pbft.PBFT, bc *blockchain.BlockChain, node *Node) error {
	pbftService := &PBFTService{
		state: pbftState,
		bc: bc,
		node: node,
	}
	return srv.RegisterName("PBFTService", pbftService)
}
