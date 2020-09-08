package collector

//add new file

type AllCollector struct {
	Option             	string          `json:"option"`
	InsInfo            	InsCollector    `json:"ins_info"`
	TransInfo			TransCollector 	`json:"trans_info"`
	BlockInfo			BlockCollector	`json:"block_info"`
}

// EVM instructions
type InsCollector struct {
	OpName              string   	      `json:"opname"`              //opcode name
	Pc                  uint64   	      `json:"pc"`                  //pc
	PcNext              string   	      `json:"pcnext"`              //next PC
	CallLayer			int   		      `json:"calllayer"`		   //call layer
	AccountValue        AccountValueInfo  `json:"calllayer"`           //from-to-value
	OpInOut             OpInOutInfo       `json:"opinout"`             //input and output of opcode
	StoreValue          SstoreValueInfo   `json:"storevalue"`          //SSTORE PreValue/CurrentValue
	Gas                 GasInfo   	      `json:"gas"`                 //pre-allocated gas and read used gas
	CheckErr            CheckInfo         `json:"checkerr"`            //check error/succeeded/valid in executing process
}

// SSTORE PreValue/CurrentValue
type SstoreValueInfo struct{
	PreValue    		string 		    `json:"prevalue"`            //pre value
	CurrentValue 		string 		    `json:"currentvalue"`        //current value
}

// check error/succeeded/valid
type CheckInfo struct{
	InternalErr         string   	    `json:"internalerr"`         //内部调用错误信息
	IsInternalSucceeded bool      	    `json:"isinternalsucceeded"` //内部调用是否成功
	IsCallValid			bool 		    `json:"iscallvalid"`
}

// gas info
type GasInfo struct{
	AllocatedGas        string   	    `json:"allocatedgas"`        //pre-allocated gas
	RealGasUsed         uint64  	    `json:"realgasused"`         //real gas used
}

// input and output of opcode
type OpInOutInfo struct{
	OpArgs              []string	    `json:"opargs"`              //opcode arguments
	RetArgs             []byte   	    `json:"retargs"`             //return byte after executed
	OpResult            string   	    `json:"opresult"`            //opcode result
	InputData           []byte   	    `json:"inputdata"`           //input data
	ByteCode			[]byte 	 	    `json:"bytecode"`			 //bytecode acquisition for 4 types of call contract
	MemoryData          []byte   	    `json:"memorydata"`          //memory data of internal call
}

// from-to-value
type AccountValueInfo struct{
	FromAddr            string   	    `json:"fromaddr"`            //form address
	ToAddr              string  	    `json:"toaddr"`              //to address
	CallContract        string          `json:"callcontract"`        //callee
	Value               string  	    `json:"value"`               //ether
}

// transactions information
type TransCollector struct {
	Op 					string 			`json:"trans_txhash"`
	TxHash       		string 			`json:"trans_txhash"`
	BlockNumber  		string 			`json:"trans_blocknumber"`
	BlockTime			string 			`json:"trans_blocktime"`
	From         		string 			`json:"trans_from"`
	To           		string 			`json:"trans_to"`
	Value       		string			`json:"trans_value"`
	GasUsed      		uint64 			`json:"trans_gasused"`
	GasPrice     		string 			`json:"trans_gasprice"`
	GasLimit     		uint64 			`json:"trans_gaslimit"`
	CallType	 		string 			`json:"trans_calltype"`	     //外部调用类型 call/create
	CallLayer           int 			`json:"trans_calllayer"`
	CreateInfo   		CreateCollector `json:"trans_createcollector"`
	CallInfo            CallCollector	`json:"trans_callcollector"`
	Nonce				uint64			`json:"trans_nonce"`
	Pc					uint64			`json:"trans_pc"`
	IsSuccess			bool 			`json:"trans_issucess"`
}

// block information
type BlockCollector struct{
	Op					string		`json:"block_Op"`
	ParentHash  		string    	`json:"block_parentHash"`
	UncleHash   		string    	`json:"block_sha3Uncles"`
	Coinbase    		string 		`json:"block_miner"`
	StateRoot        	string    	`json:"block_stateRoot"`
	TxHashRoot      	string    	`json:"block_transactionsRoot"`
	ReceiptHash 		string    	`json:"block_receiptsRoot"`
	Bloom       		[]byte      `json:"block_logsBloom"`
	Difficulty  		string      `json:"block_difficulty"`
	Number      		string      `json:"block_number"`
	GasLimit    		uint64      `json:"block_gasLimit"`
	GasUsed     		uint64      `json:"block_gasUsed"`
	Time        		uint64      `json:"block_timestamp"`
	Extra       		[]byte      `json:"block_extraData"`
	MixDigest   		string    	`json:"block_mixHash"`
	Nonce       		uint64     	`json:"block_nonce"`
}

type CreateCollector struct {
	ContractAddr      	string 		`json:"contractaddr"`
	ContractDeployCode 	[]byte 		`json:"contractinputcode"`
	ContractRuntimeCode []byte 		`json:"contractretcode"`
}

type CallCollector struct{
	InputData    		[]byte 		`json:"trans_inputdata"`
	ContractCode 		[]byte 		`json:"trans_contractcode"`
}


func NewCollector() *InsCollector {
	e := &InsCollector{}
	e.CheckErr.IsInternalSucceeded = true
	return e
}
func NewBlockCollector() *BlockCollector {
	return &BlockCollector{}
}
func NewTransCollector() *TransCollector {
	return &TransCollector{}
}

func NewCreateCollector() *CreateCollector {
	return &CreateCollector{}
}
func NewCallCollector() *CallCollector {
	return &CallCollector{}
}
func NewCollectorDataT() *AllCollector {
	return &AllCollector{}
}


func (e *InsCollector) SendInsInfo() *AllCollector {
	daT := AllCollector{}
	daT.Option = e.OpName

	daT.InsInfo = *e

	return &daT
}

//external transaction info
func (tc *TransCollector) SendTransInfo(option string) *AllCollector {
	data := AllCollector{}
	data.Option = option
	data.TransInfo = *tc
	return &data
}

func (bc *BlockCollector) SendBlockInfo(option string) *AllCollector {
	data := AllCollector{}
	data.Option = option
	data.BlockInfo = *bc
	return &data
}


func SendFlag(op string) *AllCollector {
	daT := AllCollector{}
	daT.Option = op


	return &daT
}
