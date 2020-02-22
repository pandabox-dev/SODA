package main

import (
	"fmt"
	"github.com/ethereum/collector"
	"github.com/json-iterator/go"
	"math/big"
	"strings"
)

// var logger pluginlog.ErrTxLog
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RunOpCode struct {
	MethodName string   `json:"methodname"`
	OpCode     []string `json:"option"`
}

type DaoInfo struct {
	BlockNumber      string              `json:"blocknumber"`
	TxHash           string              `json:"txhash"`
	GasUsed          uint64              `json:"gasused"`
	Cycle            string              `json:"cycle"`
	Victim           string              `json:"victim"`
	TotalCycleCount  uint64              `json:"totalcyclecount"`
	TotalValueCount  string              `json:"totalvaluecount"`
	InternalLog      string              `json:"internallog"`
}

type Node struct {
	address     string  `json:"address"`
	invalue     big.Int `json:"invalue"`
	outvalue    big.Int `json:"outvalue"`
	parent      *Node   `json:"parent`
	children    []*Node `json:"children"`
	ancestors   []*Node `json:"ancestors"`
}

// 全局变量
var (
	head Node // 外部交易的From节点
	root Node // 交易调用树的根节点
	cur_p *Node // 指向当前所在的合约节点指针
	txhash string // 交易hash
	gasused uint64 // 交易的gas
	blocknumber string // 交易的块高
)

// 插件入口函数：梦开始的地方
func Run() []byte {
	var data = RunOpCode{
		MethodName: "P1",
		OpCode:     []string{"EXTERNALINFOSTART", "EXTERNALINFOEND", "CALLSTART", "CALLEND", "CALLCODESTART", "CALLCODEEND"},
	}

	// logger.InitialFileLog("./plugin_log/thedaolog/thedaolog")
	b, err := json.Marshal(&data)
	if err != nil {
		return []byte{}
	}

	return b
}

// 数据处理函数：做梦的地方
func Recv(m *collector.CollectorDataT) (byte, string) {
	switch m.Option {
	// 该OpCode 是外部调用开始	
	case "EXTERNALINFOSTART":
		txhash = m.TransInfo.TxHash
		blocknumber = m.TransInfo.BlockNumber
		var val big.Int
		val.SetString(m.TransInfo.Value, 10)
		head = Node {
			// 统一小写
			address:     strings.ToLower(m.TransInfo.From),
			invalue:     *big.NewInt(int64(0)),
			outvalue:    val,
			parent:      nil,
			children:    nil,
			ancestors:   nil,
		}
		root = Node {
			// 统一小写
			address:     strings.ToLower(m.TransInfo.To),
			invalue:     val,
			outvalue:    *big.NewInt(int64(0)),
			// parent:      &head,
			parent:      nil,
			children:    nil,
			ancestors:   nil,
		}
		// 理论上EOA 没啥关系
		// root.ancestors = append(root.ancestors, &head)
		// head.children = append(head.children, &root)
		cur_p = &root

	// 该OpCode 是内部调用开始
	case "CALLSTART", "CALLCODESTART":
		var val big.Int
		val.SetString(m.InsInfo.Value, 10)
		node := Node {
			// 统一小写
			address:     strings.ToLower(m.InsInfo.To),
			invalue:     val,
			outvalue:    *big.NewInt(int64(0)),
			parent:      cur_p,
			children:    nil,
			ancestors:   nil,
		}
		node.ancestors = append([]*Node{cur_p}, cur_p.ancestors...)
		cur_p.children = append(cur_p.children, &node)
		cur_p.outvalue.Add(&cur_p.outvalue, &val)
		cur_p = &node

	// 该OpCode 是内部调用结束
	case "CALLEND", "CALLCODEEND":
		cur_p = cur_p.parent
		// 如果当前调用失败，在树中删除相应节点
		if !m.InsInfo.IsInternalSucceeded || !m.InsInfo.IsCallValid {
			cur_p.children = cur_p.children[:len(cur_p.children)-1]
		}

	// 该OpCode 是外部调用结束
	case "EXTERNALINFOEND":
		// 使用的gas
		gasused = m.TransInfo.GasUsed
		// 如果该交易失败
		if !m.TransInfo.IsSuccess {
			return 0x00, ""
		}
		// 处理环 && 调用关系
		result := procCycleInfo()
		if result != ""{
			return 0x01, result
		}
	}
	return 0x00, ""
}

// 判断有没有环，顺便把所有的调用记录下来
func procCycleInfo() string {
	// cur_p = &head
	cur_p = &root // 换成EOA调用的第一个地址
	nodes := []*Node{cur_p}
	// 该交易下总的情况（环，总的交易额，总的环个数，总的环
	cycles := make(map[string]bool)
	totalValueCount := *big.NewInt(int64(0))
	var totalCycleCount uint64
	totalCycleCount = 0
	totalCycleStr := "["
	totalCallStr := "["
	victim := ""
	// 通过迭代的方式后序遍历树结构
	for len(nodes) > 0 {
		cur_p = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		// 遍历的时候记录一下调用关系（没有前后关系）
		if cur_p.parent != nil {
			totalCallStr += "{\"addr\":\"" + cur_p.parent.address + "--" + cur_p.address + "\",\"value\":" + fmt.Sprintf("%v", cur_p.invalue.String()) + "},"
		}
		// 寻找环啦
		if cur_p.ancestors != nil && cur_p.children != nil {
			for _, child := range cur_p.children {
				// 自己调用自己不算，新增一个，要是没向外转钱（自己和儿子都没转钱）也不算
				if child.address == cur_p.address || (child.outvalue.Int64() == 0 && !isChildrenOutValue(child)) {
					continue
				}
				// 看看之前有没有调用过当前节点调用过的节点，
				// 若有，则形成了一个环A--...--B(cur_p)--A
				for _, ancestor := range cur_p.ancestors {
					// 找到一个环了（可能有多个），需要继续在这下面找一个一样的
					if ancestor.address == child.address {
						cycleArr := []*Node{child}
						tem_p := cur_p
						// 找到调用的路线
						for tem_p != ancestor {
							cycleArr = append(cycleArr, tem_p)
							tem_p = tem_p.parent
						}
						cycleArr = append(cycleArr, ancestor)
						// 将环的调用字符串存储到map 中
						cycleStr := ""
						for _, call := range cycleArr {
							if cycleStr == "" {
								cycleStr = call.address
							} else {
								cycleStr = call.address + "--" + cycleStr
							}
						}
						// 在这下面找还有没有一样的环，直接搜索吧…… 短路逻辑先判断是否含有第二个环，有就不用继续找了
						if isPathExist(cycleArr, child) || isContinuedCall(cycleArr[len(cycleArr)-2].address, child) {
							var value *big.Int
							victim, value = procCycleVictims(cycleArr)
							totalValueCount.Add(&totalValueCount, value)
							cycles[cycleStr] = true
							break // 找到直接退出
						}
					}
				}
			}
		}
		// 孩子节点进入栈中
		for _, child := range cur_p.children {
			nodes = append(nodes, child)
		}
	}
	// 判断有没有环数大于等于2的环，事实上我每次环数大于等于2时才入字典，因此替换成是否存在cycles[string]bool
	for cycle, _ := range cycles {
		// 此时认为是一个重入攻击
		totalCycleCount += 1
		totalCycleStr += "{\"detail\":\"" + cycle + "\"},"
	}
	totalCycleStr += "]"
	totalCallStr += "]"

	// 没有环或者就妹转钱
	if totalCycleCount == 0 {
		return ""
	}

	var daoInfo = DaoInfo {
		BlockNumber:      blocknumber,
		TxHash:           txhash,
		GasUsed:          gasused,
		Cycle:            totalCycleStr,
		Victim:           victim,
		TotalCycleCount:  totalCycleCount,
		TotalValueCount:  totalValueCount.String(),
		InternalLog:      "", // 先不写，太多了 占空间
		// InternalLog:      totalCallStr,
	}
	daoJsonData, err := json.Marshal(daoInfo)
	if err != nil {
		panic(err)
	}
	// 写日志
	// Log(&logger, string(daoJsonData))
	return string(daoJsonData)
}

// 判断路径是否存在（找有没有第二个相同的环），从后往前搜索，因为我的cycleArr是倒着的
func isPathExist(cycleArr []*Node, node *Node) bool {
	nodes := []*Node{node}
	res := false
	for len(nodes) > 0 {
		tem_p := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		res = res || isPathExist_(cycleArr, tem_p, len(cycleArr)-1)
		for _, child := range tem_p.children {
			nodes = append(nodes, child)
		}
	}
	return res
}

// 递归子函数，新增一个，第二个环也要转了钱钱才算
func isPathExist_(cycleArr []*Node, node *Node, index int) bool {
	if node.address != cycleArr[index].address {
		return false
	} else if index == 0 {
		// 向外转了钱钱才算，或者它调用的合约转了钱钱
		return node.outvalue.Int64() != 0 || isChildrenOutValue(node)
	}
	res := false
	for _, child := range node.children {
		res = res || isPathExist_(cycleArr, child, index-1)
	}
	return res
}

// 孩子有没有向外转过钱钱
func isChildrenOutValue(node *Node) bool {
	for _, child := range node.children {
		if child.outvalue.Int64() != 0 {
			return true
		}
	}
	return false
}

// 有没有继续调用某个地址的合约
func isContinuedCall(address string, node *Node) bool {
	for _, child := range node.children {
		if child.address == address {
			return true
		}
	}
	return false
}

// 找出一个环中的受害者
func procCycleVictims(cycleArr []*Node) (string, *big.Int) {
	victims := make(map[string]*big.Int)
	victim := ""
	value := big.NewInt(int64(0))
	for _, node := range cycleArr {
		// 转出的钱钱比收入的多，它就是受害者了
		if node.outvalue.Cmp(&(node.invalue)) == 1 {
			if value, ok := victims[node.address]; ok {
				value.Add(value, &(node.outvalue))
			} else {
				victims[node.address] = big.NewInt(int64(0))
				victims[node.address].Add(victims[node.address], &(node.outvalue))
			}
		}
	}
	if len(victims) == 0 {
		victims[cycleArr[0].address] = big.NewInt(int64(0))
		victims[cycleArr[0].address].Add(victims[cycleArr[0].address], &(cycleArr[0].outvalue))
	}
	// 只要被偷钱最多的
	for k, v := range victims {
		if value.Cmp(v) != 1 {
			victim, value = k, v
		}
	}
	if victim == "0xd2e16a20dd7b1ae54fb0312209784478d069c7b0" {
		victim = "0xbb9bc244d798123fde783fcc1c72d3bb8c189413"
	}
	return victim, value
}

// 打印出树，调试用
func printTree(node *Node) string {
	if node.children == nil {
		var ancestorsAddr []string
		for _, ancestor := range node.ancestors {
			ancestorsAddr = append(ancestorsAddr, ancestor.address)
		}
		return fmt.Sprintf("{\"address\": \"%s\", \"ancestors\": \"%s\"}", node.address, strings.Join(ancestorsAddr, ", "))
	} else {
		var ancestorsAddr []string
		for _, ancestor := range node.ancestors {
			ancestorsAddr = append(ancestorsAddr, ancestor.address)
		}
		res := fmt.Sprintf("{\"address\": \"%s\", \"ancestors\": \"%s\", \"children\": [", node.address, strings.Join(ancestorsAddr, ", "))
		var childJsonArr []string
		for _, child := range node.children {
			childJsonArr = append(childJsonArr, printTree(child))
		}
		res += strings.Join(childJsonArr, ", ")
		res += "]}"
		return res
	}
}