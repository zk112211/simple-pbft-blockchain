# PBFT Blockchain Consensus System

A blockchain system implementation based on PBFT (Practical Byzantine Fault Tolerance) consensus algorithm, developed in Go language.

## Project Introduction

This project implements a complete PBFT consensus blockchain system, including blockchain core functionality, PBFT consensus algorithm, RPC communication services, and node management. The system can handle Byzantine faults, ensuring consensus can be reached even in the presence of malicious nodes.

## Features

### 🏗️ Blockchain Core Functionality
- **Block Structure**: Supports standard blockchain fields such as timestamp, transaction list, previous block hash, etc.
- **Transaction Processing**: Complete transaction creation, signing, verification, and pool management
- **Cryptographic Support**: Digital signature and verification based on ECDSA
- **Data Persistence**: Supports JSON format storage and loading of blockchain data

### 🔐 PBFT Consensus Algorithm
- **Three-Phase Consensus**: Pre-prepare → Prepare → Commit → Reply
- **Byzantine Fault Tolerance**: Can tolerate up to f malicious nodes (total nodes ≥ 3f+1)
- **View Management**: Supports view switching and primary node election
- **Message Verification**: Complete message integrity verification and sequence number management

### 🌐 Network Communication
- **RPC Services**: Node-to-node communication based on Go standard RPC library
- **HTTP Interface**: Supports HTTP POST request client communication
- **Message Broadcasting**: Message broadcasting and forwarding mechanism between nodes
- **Asynchronous Processing**: Asynchronous message processing based on channels

### 🖥️ Node Management
- **Multi-Node Support**: Supports deployment and management of multiple consensus nodes
- **State Synchronization**: Automatic node state synchronization and recovery
- **Fault Handling**: Timeout mechanisms and error recovery
- **Logging**: Complete operation logs and stage tracking

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Client Apps   │    │   Client Apps   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      RPC Service Layer    │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │    PBFT Consensus Engine  │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │   Blockchain Core Layer   │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │     Data Storage Layer    │
                    └───────────────────────────┘
```

## Directory Structure

```
PBFTblockchain/
├── blockchain/           # Blockchain core implementation
│   └── blockchain.go     # Block, transaction, blockchain data structures and methods
├── pbft/                 # PBFT consensus algorithm implementation
│   ├── pbft_type.go      # PBFT related data type definitions
│   └── pbft.go          # PBFT consensus algorithm core logic
├── RPC/                  # RPC communication services
│   ├── node.go          # Node management and message handling
│   ├── PBFTservice.go   # PBFT RPC service interface
│   └── log.go           # Logging functionality
├── server.go             # Server startup entry point
├── client.go             # Client example code
├── go.mod               # Go module dependency management
└── README.md            # Project documentation
```

## Installation Requirements

- **Go Version**: 1.21.5 or higher
- **Operating System**: All platforms supported by Go (Windows, macOS, Linux)
- **Network**: TCP/IP communication support

## Quick Start

### 1. Clone the Project
```bash
git clone <repository-url>
cd PBFTblockchain
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Start the Server
```bash
go run server.go
```

### 4. Run the Client
```bash
go run client.go
```

## Configuration

### Node Configuration
Configure node information in `RPC/node.go`:
```go
NodeTable: map[string]string{
    "Apple":  "localhost:1111",
    "MS":     "localhost:1112", 
    "Google": "localhost:1113",
    "IBM":    "localhost:1114",
}
```

### Consensus Parameters
- **Fault Tolerance Nodes**: f = 1 (can tolerate up to 1 malicious node)
- **Minimum Nodes**: 3f + 1 = 4 nodes
- **Timeout**: 1 second

## Usage Examples

### Create Transaction
```go
tx := blockchain.Transaction{
    Sender:   "Alice",
    Receiver: "Bob", 
    Amount:   100,
}
```

### Start Consensus
```go
prePrepareMsg, err := pbftState.StartConsensus(request)
```

### Handle Consensus Messages
```go
// Pre-prepare phase
prepareMsg, err := state.PrePrepare(prePrepareMsg)

// Prepare phase  
commitMsg, err := state.Prepare(prepareMsg)

// Commit phase
replyMsg, requestMsg, err := state.Commit(commitMsg)
```

## API Interfaces

### PBFT Service Interfaces
- `StartConsensusRPC`: Start consensus process
- `PrePrepareRPC`: Handle pre-prepare messages
- `PrepareRPC`: Handle prepare messages
- `CommitRPC`: Handle commit messages

### Blockchain Interfaces
- `NewBlockChain`: Create new blockchain
- `AddBlock`: Add new block
- `AddTransactionToPool`: Add transaction to pool
- `SignTransaction`: Transaction signing
- `VerifySignature`: Signature verification

## Consensus Process

1. **Request**: Client sends request to primary node
2. **Pre-prepare**: Primary node creates pre-prepare message and broadcasts
3. **Prepare**: Replica nodes verify and send prepare messages
4. **Commit**: Nodes collect sufficient prepare messages and send commit messages
5. **Reply**: Reply to client after consensus is reached

## Performance Characteristics

- **Consensus Latency**: Typically 3 rounds of message exchange
- **Throughput**: Supports concurrent transaction processing
- **Scalability**: Supports dynamic node joining/leaving
- **Fault Tolerance**: Can tolerate up to 1/3 malicious nodes

## Development Guidelines

### Code Standards
- Follow Go standard coding conventions
- Complete error handling
- Detailed comments
- Modular design

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./blockchain
go test ./pbft
```

## Troubleshooting

### Common Issues
1. **Port Conflicts**: Check node port configuration
2. **Network Connection**: Confirm firewall settings
3. **Consensus Failure**: Check node count and configuration

### Debug Mode
Enable detailed log output:
```go
// Add more fmt.Printf statements in code
fmt.Printf("Debug: %+v\n", variable)
```

## Contributing

Welcome to submit Issues and Pull Requests to improve the project!

### Development Process
1. Fork the project
2. Create feature branch
3. Submit changes
4. Create Pull Request

## License

This project is licensed under the MIT License. See LICENSE file for details.

## Contact

For questions or suggestions, please contact through:
- Submit GitHub Issue
- Send email to project maintainers

## Changelog

### v1.0.0
- Initial version release
- Implemented basic PBFT consensus algorithm
- Support for multi-node deployment
- Complete blockchain functionality

---

**Note**: This is an educational blockchain implementation and is not recommended for production use. Before using in production environments, please conduct thorough security audits and testing.
