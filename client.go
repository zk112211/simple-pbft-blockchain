package main

import (
	"encoding/json"
	"log"
	"net/rpc"
	"PBFTblockchain/blockchain"
	"PBFTblockchain/RPC" // 替换为您的 RPC 包路径
)

func main() {
	// 连接到 RPC 服务器（主节点）
	client, err := rpc.Dial("tcp", "主节点地址:端口")
	if err != nil {
		log.Fatalf("连接 RPC 服务器失败: %v", err)
	}

	// 创建一个示例交易
	tx := blockchain.Transaction{
		Sender:   "Alice",
		Receiver: "Bob",
		Amount:   100,
	}

	// 将交易序列化为 JSON
	txData, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("序列化交易失败: %v", err)
	}

	// 准备 RPC 请求参数
	args := RPC.RequestArgs{Data: string(txData)}
	var response RPC.Response

	// 发起 RPC 调用
	err = client.Call("PBFTService.ReceiveRequest", args, &response)
	if err != nil {
		log.Fatalf("调用 RPC 方法失败: %v", err)
	}

	// 处理响应
	if response.Ack {
		log.Println("请求成功处理")
	} else {
		log.Println("请求处理失败")
	}
}
