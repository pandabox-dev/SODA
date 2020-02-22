package collector

//add new file

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

//internal info
type Collector struct {
	Op                  string   	`json:"op"`                  //op
	Pc                  uint64   	`json:"pc"`                  //pc
	CallLayer			int   		`json:"calllayer"`			 //调用层数
	CallContract        string      `json:"callcontract"`
	From                string   	`json:"from"`                //form
	To                  string  	`json:"to"`                  //to
	GasUsed             uint64  	`json:"gasused"`             //实际gas消耗
	Value               string  	`json:"value"`               //ether
	OpArgs              []string	`json:"opargs"`              //计算参数
	OpValuePre          string		`json:"opvaluepre"`          //之前的值
	OpValueNow          string 	 	`json:"opvaluenow"`          //之后的值
	OpResult            string   	`json:"opresult"`            //计算的结果值
	ByteArgs            []byte   	`json:"byteargs"`            //input字节参数
	RetArgs             []byte   	`json:"retargs"`             //执行后返回的字节
	PcNext              string   	`json:"pcnext"`              //下一字节码位
	GasTmp              string   	`json:"gastmp"`              //预先分配的gas
	PreArgs             []byte   	`json:"preargs"`             //暂时只用于内部调用mem data
	ByteCode			[]byte 	 	`json:"bytecode"`			 //用于4种call调用合约的bytecode获取
	InternalErr         string   	`json:"internalerr"`         //内部调用错误信息
	IsInternalSucceeded bool      	`json:"isinternalsucceeded"` //内部调用是否成功
	IsCallValid			bool 		`json:"iscallvalid"`
}

type CallCollector struct{
	InputData    		[]byte 		`json:"trans_inputdata"`
	ContractCode 		[]byte 		`json:"trans_contractcode"`
}

type CreateCollector struct {
	ContractAddr      	string 		`json:"contractaddr"`
	ContractDeployCode 	[]byte 		`json:"contractinputcode"`
	ContractRuntimeCode []byte 		`json:"contractretcode"`
}

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
	CallType	 		string 			`json:"trans_calltype"`	//外部调用类型 call/create
	CallLayer           int 			`json:"trans_calllayer"`
	CreateInfo   		CreateCollector `json:"trans_createcollector"`
	CallInfo            CallCollector	`json:"trans_callcollector"`
	Nonce				uint64			`json:"trans_nonce"`
	Pc					uint64			`json:"trans_pc"`
	IsSuccess			bool 			`json:"trans_issucess"`
}

type CollectorDataT struct {
	Option             	string          `json:"option"`
	InsInfo            	Collector       `json:"ins_info"`
	TransInfo			TransCollector 	`json:"trans_info"`
	BlockInfo			BlockCollector	`json:"block_info"`
}

func NewCollector() *Collector {
	e := &Collector{}
	e.IsInternalSucceeded = true
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
func NewCollectorDataT() *CollectorDataT {
	return &CollectorDataT{}
}


func (e *Collector) SendInsInfo() *CollectorDataT {
	daT := CollectorDataT{}
	daT.Option = e.Op

	daT.InsInfo = *e

	return &daT
}


//外部交易数据
func (tc *TransCollector) SendTransInfo(option string) *CollectorDataT {
	data := CollectorDataT{}
	data.Option = option
	data.TransInfo = *tc
	return &data
}

func (bc *BlockCollector) SendBlockInfo(option string) *CollectorDataT {
	data := CollectorDataT{}
	data.Option = option
	data.BlockInfo = *bc
	return &data
}


func SendFlag(op string) *CollectorDataT {
	daT := CollectorDataT{}
	daT.Option = op


	return &daT
}
