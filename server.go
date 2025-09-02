package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"PBFTblockchain/blockchain"
	"PBFTblockchain/pbft"
	"PBFTblockchain/RPC" // 替换为您的 RPC 包路径
)

func main() {
	// 初始化区块链和 PBFT 状态
	var viewID, ls int64
	fmt.Scanln(viewID, ls)
	bc, err := blockchain.LoadBlockchainFromFile("data.json")
	if err != nil { //区块链加载失败则创建新的区块链
		bc = blockchain.NewBlockChain()
	}
	pbftState := pbft.CreateState(viewID, ls) // 根据实际的函数名和参数调整
	node := RPC.NewNode("1") //节点id

	// 创建 RPC 服务器
	server := rpc.NewServer()

	// 注册 PBFTService
	err = RPC.RegisterPBFTService(server, pbftState, bc, node)
	if err != nil {
		log.Fatalf("注册 PBFTService 失败: %v", err)
	}

	// 监听特定端口
	listener, err := net.Listen("tcp", ":端口")
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}
	defer listener.Close()

	log.Println("RPC 服务启动，监听端口")
	server.Accept(listener)
}
