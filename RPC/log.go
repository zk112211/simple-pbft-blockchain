package RPC

import (
	"PBFTblockchain/pbft"
	"fmt"
)

// LogMsg 输出消息日志
func LogMsg(msg interface{}) {
	switch msg.(type) {
	case *pbft.RequestMsg:
		reqMsg := msg.(*pbft.RequestMsg)
		fmt.Printf("[request] ClientID: %s, Timestamp: %d, Operation: %s\n", reqMsg.ClientID, reqMsg.Timestamp)
	case *pbft.PrePrepareMsg:
		prePrepareMsg := msg.(*pbft.PrePrepareMsg)
		fmt.Printf("[prePrepare] ClientID: %s, Operation: %s, SequenceID: %d\n", prePrepareMsg.RequestMsg.ClientID, prePrepareMsg.SequenceID)
	case *pbft.VoteMsg:
		voteMsg := msg.(*pbft.VoteMsg)
		if voteMsg.MsgType == pbft.PrepareMsg {
			fmt.Printf("[Prepare] NodeID: %s\n", voteMsg.NodeID)
		} else if voteMsg.MsgType == pbft.CommitMsg {
			fmt.Printf("[commit] NodeID: %s\n", voteMsg.NodeID)
		}
	}
}

// LogStage 输出阶段日志
func LogStage(stage string, isDone bool) {
	if isDone {
		fmt.Printf("[STAGE-DONE] %s\n", stage)
	} else {
		fmt.Printf("[STAGE-BEGIN] %s\n", stage)
	}
}
