package blockchain
//区块链基本功能的实现

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// Transaction 交易数据类型
type Transaction struct {
	Sender    string //交易发起者
	Receiver  string //交易接受者
	Amount    int64 //交易金额
	Signature string //数字签名
}

// Block 区块数据类型
type Block struct {
	Timestamp       int64 //区块创建的时间戳
	Transactions    []Transaction //区块所包含的交易
	PrevBlockHash   []byte //前一区块的哈希值
	Hash            []byte //当前区块的哈希值
}

// BlockChain 区块链数据类型
type BlockChain struct {
	Blocks          []Block //区块链中包含的区块
	TransactionPool []Transaction //区块链未处理的交易池
}

// NewBlockChain 新建区块链
func NewBlockChain() *BlockChain{
	//创世区块genesisBlock
	genesisBlock := Block{
		Timestamp: time.Now().Unix(),
		Transactions: []Transaction{}, //创世区块一般没有交易
		PrevBlockHash: []byte{}, //创世区块没有上一个区块
	}
	genesisBlock.Hash = calculateHash(genesisBlock) //计算创世区块的哈希值，并赋值给创世区块的Hash

	return &BlockChain{Blocks: []Block{genesisBlock}, TransactionPool: []Transaction{}}
}

// AddBlock 向区块链中添加一个新的区块
func (bc *BlockChain) AddBlock(){
	prevBlock := bc.Blocks[len(bc.Blocks)-1] //上一区块
	//创建新的区块
	newBlock := Block{
		Timestamp: time.Now().Unix(),
		Transactions: bc.TransactionPool, //处理未完成的交易池中的区块，并添加交易到新的区块中
		PrevBlockHash: prevBlock.Hash, //上一区块哈希值
	}
	newBlock.Hash = calculateHash(newBlock) //计算新的区块哈希值
	bc.TransactionPool = []Transaction{} //新区块创建完成，交易池中的交易处理完成，重置交易池
	bc.Blocks = append(bc.Blocks, newBlock) //将新的区块添加到区块链中
}

//计算sha256哈希值
func hash(data []byte) []byte {
	h := sha256.New() //创建哈希加密器
	h.Write(data) //将要加密的数据输入到加密器中
	hashed := h.Sum(nil) //对数据进行哈希操作
	return hashed
}

//计算区块的哈希值
func calculateHash(block Block) []byte{
	//将区块内容连接成字符串
	data := strconv.FormatInt(block.Timestamp, 10) + transactionHash(block) + hex.EncodeToString(block.PrevBlockHash)

	//对区块内容进行两次哈希计算并返回
    return hash(hash([]byte(data)))
}

//对交易进行哈希
func transactionHash(block Block) string{
    var txStr []string //将进行操作的所有交易连接的字符串切片

    //将所有交易池中的交易连接起来
    for _, tx := range block.Transactions{
    	txStr = append(txStr, tx.Sender, tx.Receiver, strconv.FormatInt(tx.Amount, 10), tx.Signature) //将交易内容加入到txStr中
	}

	//将交易连接成一个字符串
	str := strings.Join(txStr, "")

	//返回哈希后的字符串
	return hex.EncodeToString(hash([]byte(str)))
}

// AddTransactionToPool 添加交易到交易池中
func (bc *BlockChain) AddTransactionToPool(tx *Transaction) {
	bc.TransactionPool = append(bc.TransactionPool, *tx)
}

// NewKeyPair 生成新的密钥对
func NewKeyPair() (ecdsa.PrivateKey, ecdsa.PublicKey){
	//使用ecdsa创建私钥
	privkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	//错误返回
	if err != nil{
		log.Panic(err)
	}

	//返回生成的私钥和公钥
	return *privkey, privkey.PublicKey
}

// SignTransaction 对交易进行签名
func (tx *Transaction)SignTransaction(privKey ecdsa.PrivateKey) {
	//对交易内容进行哈希
	txHash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%f",tx.Sender, tx.Receiver, tx.Amount)))

	//进行ecdsa签名
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, txHash[:])

	//错误返回
	if err != nil{
		log.Panic(err)
	}

	//将签名内容连接并赋给交易中的签名参数Signature
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = string(signature)
}

// VerifySignature 验证交易签名
func VerifySignature(tx *Transaction, pubKey ecdsa.PublicKey) bool{
	//对交易进行哈希
	txHash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%f", tx.Sender, tx.Receiver, tx.Amount)))

	//将签名结果的大整数分开
	r := big.Int{}
	s := big.Int{}
	sigLen := len(tx.Signature)
	r.SetBytes([]byte(tx.Signature[:(sigLen / 2)]))
	s.SetBytes([]byte(tx.Signature[(sigLen / 2):]))

	//验证签名并返回结果
	return ecdsa.Verify(&pubKey, txHash[:], &r, &s)
}

//区块链的文件储存
func (bc BlockChain)storeBlockchain(filePath string) error{
	//使用os.create函数打开文件，若文件不存在，将其创建并打开
    file, err := os.Create(filePath)

    //创建文件时错误
    if err != nil{
    	return err
	}

	defer file.Close()

    //json编码器，将其输出定向到file
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(bc); err != nil{
    	return err
	} //json输出错误，返回错误

	return nil
}

//区块链的储存文件读取
func loadBlcokChain(filePath string) (*BlockChain, error){
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil{
		return nil, err
	}

	var bc *BlockChain
	err = json.Unmarshal(bytes, &bc)
	if err != nil {
		return nil, err
	}

    return bc, nil
}

// LoadBlockchainFromFile 从文件中加载区块链数据
func LoadBlockchainFromFile(filePath string) (*BlockChain, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，返回空的区块链实例或错误
		return NewBlockChain(), nil // 或返回错误
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 反序列化内容到 BlockChain 实例
	var blockchain BlockChain
	err = json.Unmarshal(content, &blockchain)
	if err != nil {
		return nil, err
	}

	return &blockchain, nil
}