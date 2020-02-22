// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	//add new 
	"github.com/ethereum/collector"
	// "syscall"
	"strconv"
	"fmt"
	"github.com/ethereum/go-ethereum/tingrong"
	"github.com/ethereum/go-ethereum/cmd/pluginManage"
	"github.com/ethereum/go-ethereum/fei"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the block and state according to any hard-fork specs
	if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
		misc.ApplyDAOHardFork(statedb)
	}

	//add new
	if p.config.TransferDataPlg.GetOpcodeRegister("handle_BLOCK_INFO"){
		blockcollector := collector.NewBlockCollector()
		blockcollector.Op = "Block"+fmt.Sprintf("%v", header.Number)
		blockcollector.ParentHash = header.ParentHash.String()
		blockcollector.UncleHash = header.UncleHash.String()
		blockcollector.Coinbase = header.Coinbase.String()
		blockcollector.StateRoot = header.Root.String()
		blockcollector.TxHashRoot = header.TxHash.String()
		blockcollector.ReceiptHash = header.ReceiptHash.String()
		blockcollector.Difficulty = header.Difficulty.String()
		blockcollector.Number = header.Number.String()
		blockcollector.GasLimit = header.GasLimit
		blockcollector.GasUsed = header.GasUsed
		blockcollector.Time = header.Time
		blockcollector.Extra = header.Extra
		blockcollector.MixDigest = header.MixDigest.String()
		blockcollector.Nonce = header.Nonce.Uint64()
		p.config.TransferDataPlg.SendDataToPlugin("handle_BLOCK_INFO",blockcollector.SendBlockInfo("handle_BLOCK_INFO"))
	}
	//add new

	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}

	

	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), block.Uncles())

	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)

	//add new 
	vmenv.SetTxStart(true)
	vmenv.ChainConfig().TransferDataPlg.Start()

	//feifei add new --api
	if fei.IsReg {
		// //whole folder fresh有问题，如果全部移除，是无法删掉旧的的。需要用new去新增
		// pluginManage.StartRun(vmenv.ChainConfig().TransferDataPlg)
		// fei.IsReg = false

		//single plugin
		pluginManage.RegisterMethod(vmenv.ChainConfig().TransferDataPlg, fei.RegPath)
		fei.RegPath = fei.Clear
		fei.IsReg = false
	}

	if fei.IsUn {
		vmenv.ChainConfig().TransferDataPlg.UnRegisterPlg()
		fei.IsUn = false
		fei.UnPlg = fei.Clear
	}

	if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("TXSTART"){
		vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("TXSTART", collector.SendFlag("TXSTART"))
	}

	tingrong.CALL_LAYER = 0
	tingrong.CALL_STACK = nil
	tingrong.ALL_STACK = nil
	tingrong.EXTERNAL_FLAG = true
	tingrong.BLOCKING_FLAG = false
	tingrong.PLUGIN_SNAPSHOT_ID = 0 
	tingrong.CALLVALID_MAP = make(map[int]bool)

	tingrong.TxHash = tx.Hash().String()

	// if vmenv.BlockNumber.Int64() >= 2000001{
	// 	if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("ENDSIGNAL") {
	// 		vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("ENDSIGNAL", collector.SendFlag("ENDSIGNAL"))
	// 	}
	// 	fmt.Println("!!!!!!!!!!!!!!   ", vmenv.BlockNumber, "   EXIT")
	// 	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	// }

	// fmt.Println("tingrong.txhash",tingrong.TxHash)

	if msg.To() != nil{
		tingrong.CALL_LAYER += 1
		tingrong.CALL_STACK = append(tingrong.CALL_STACK,msg.To().String()+"#"+strconv.Itoa(tingrong.CALL_LAYER))
		tingrong.ALL_STACK = append(tingrong.ALL_STACK,msg.To().String())
		// fmt.Println("option:EXTERNALINFOSTART")
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("contract:",msg.To().String())
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}

	tcstart := collector.NewTransCollector()

	//external collector
	if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("EXTERNALINFOSTART"){
		tcstart.Op = "EXTERNALINFOSTART"
		tcstart.TxHash = tx.Hash().String()
		tcstart.BlockNumber = vmenv.BlockNumber.String()
		tcstart.BlockTime = vmenv.Time.String()
		tcstart.From = msg.From().String()
		tcstart.Value = msg.Value().String()
		tcstart.GasPrice = msg.GasPrice().String()
		tcstart.GasLimit = msg.Gas()
		tcstart.Nonce = tx.Nonce()
		tcstart.CallLayer = 1
		if msg.To() != nil {
			tcstart.CallType = "CALL"
			tcstart.To = msg.To().String()
			
			callcollector := collector.NewCallCollector()
			if vmenv.StateDB.Exist(*msg.To()) {
				callcollector.ContractCode = vmenv.StateDB.GetCode(*msg.To())
			}
			callcollector.InputData = msg.Data()			
			tcstart.CallInfo = *callcollector
		}
		vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("EXTERNALINFOSTART", tcstart.SendTransInfo("EXTERNALINFOSTART"))

	}

	
	//add new 

	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := ApplyMessage(vmenv, msg, gp)


	//add new 
	if tingrong.BLOCKING_FLAG == true{
		statedb.RevertToSnapshot(tingrong.PLUGIN_SNAPSHOT_ID)
	}
	//add new 


	tcend := collector.NewTransCollector()

	//add new 
	if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("EXTERNALINFOEND"){
		tcend.Op = "EXTERNALINFOEND"
		tcend.TxHash = tx.Hash().String()
		tcend.GasUsed = gas
		tcend.CallLayer = 1
	}
	//add new 

	if err != nil {
		//add new 
		if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("EXTERNALINFOEND"){
			tcend.IsSuccess = false
			vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("EXTERNALINFOEND", tcend.SendTransInfo("EXTERNALINFOEND"))	
		}
		// fmt.Println("externalend err tingrong.STACK_FLAG",tingrong.STACK_FLAG)
		//add new 
		return nil, 0, err
	}
	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
		//add new 
		if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("EXTERNALINFOEND"){
			tcend.CallType = "CREATE"
			tcend.To = receipt.ContractAddress.String()
			createcollector := collector.NewCreateCollector()
			createcollector.ContractAddr = receipt.ContractAddress.String()
			createcollector.ContractDeployCode = msg.Data()	
			if vmenv.StateDB.Exist(receipt.ContractAddress) {
				createcollector.ContractRuntimeCode = vmenv.StateDB.GetCode(receipt.ContractAddress)
			}		
			tcend.CreateInfo = *createcollector
		}
		//add new 
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = statedb.BlockHash()
	receipt.BlockNumber = header.Number
	receipt.TransactionIndex = uint(statedb.TxIndex())

	//add new 
	if !failed {
		if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("EXTERNALINFOEND"){
			tcend.IsSuccess = true
			vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("EXTERNALINFOEND", tcend.SendTransInfo("EXTERNALINFOEND"))
		}
	} else {
		if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("EXTERNALINFOEND"){
			tcend.IsSuccess = false
			vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("EXTERNALINFOEND", tcend.SendTransInfo("EXTERNALINFOEND"))
		}
	}

	tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
	// fmt.Println("option:EXTERNALINFOEND")
	// fmt.Println("layer:",tingrong.CALL_LAYER)
	// fmt.Println("stack:",tingrong.CALL_STACK)
	// fmt.Println("tingrong.ALL_STACK",tingrong.ALL_STACK)

	if vmenv.ChainConfig().TransferDataPlg.GetOpcodeRegister("TXEND"){
		vmenv.ChainConfig().TransferDataPlg.SendDataToPlugin("TXEND", collector.SendFlag("TXEND"))
		vmenv.ChainConfig().TransferDataPlg.Stop()
	}

	
	//add new 

	return receipt, gas, err
}
