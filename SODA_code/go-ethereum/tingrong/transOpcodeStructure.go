package tingrong


//add new file

var BLOCK_TIME_LIST [100000][3]string
var CLEAR_BLOCK_TIME_LIST [100000][3]string
var COUNT_ARRAY int

//add new
var TxHash 		string
var CALL_LAYER	int
var CALL_STACK 	[]string	//call contract
var ALL_STACK   []string    //all contract
var BLOCKING_FLAG bool		//是否阻断交易
var EXTERNAL_FLAG bool 		//external call/create
var PLUGIN_SNAPSHOT_FLAG bool
var PLUGIN_SNAPSHOT_ID int
var CALLVALID_MAP map[int]bool