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

package vm

import (
	"errors"
	"fmt"
	"github.com/ethereum/collector"
	"math/big"
	"strings"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/crypto/sha3"
	"github.com/ethereum/go-ethereum/tingrong"
)

var (
	bigZero                  = new(big.Int)
	tt255                    = math.BigPow(2, 255)
	errWriteProtection       = errors.New("evm: write protection")
	errReturnDataOutOfBounds = errors.New("evm: return data out of bounds")
	errExecutionReverted     = errors.New("evm: execution reverted")
	errMaxCodeSizeExceeded   = errors.New("evm: max code size exceeded")
	errInvalidJump           = errors.New("evm: invalid jump destination")
)

func opAdd(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	math.U256(y.Add(x, y))

	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opSub(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	math.U256(y.Sub(x, y))

	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opMul(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}

	stack.push(math.U256(x.Mul(x, y)))

	interpreter.intPool.put(y)
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

func opDiv(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	if y.Sign() != 0 {
		math.U256(y.Div(x, y))
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := interpreter.intPool.getZero()

	if y.Sign() == 0 || x.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() != y.Sign() {
			res.Div(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Div(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	interpreter.intPool.put(x, y)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opMod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	if y.Sign() == 0 {
		stack.push(x.SetUint64(0))
	} else {
		stack.push(math.U256(x.Mod(x, y)))
	}
	interpreter.intPool.put(y)
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

func opSmod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := interpreter.intPool.getZero()

	if y.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() < 0 {
			res.Mod(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Mod(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	interpreter.intPool.put(x, y)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opExp(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	base, exponent := stack.pop(), stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, base.String(), exponent.String())
	}
	// some shortcuts
	cmpToOne := exponent.Cmp(big1)
	var res *big.Int
	if cmpToOne < 0 { // Exponent is zero
		// x ^ 0 == 1
		res = base.SetUint64(1)
	} else if base.Sign() == 0 {
		// 0 ^ y, if y != 0, == 0
		res = base.SetUint64(0)
	} else if cmpToOne == 0 { // Exponent is one
		// x ^ 1 == x
		res = base
	} else {
		res = math.Exp(base, exponent)
		interpreter.intPool.put(base)
	}
	stack.push(res)
	interpreter.intPool.put(exponent)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	back := stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, back.String())
	}

	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		if stack.flag {
			stack.collector.OpArgs = append(stack.collector.OpArgs, num.String())
		}
		mask := back.Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if num.Bit(int(bit)) > 0 {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}
		value := math.U256(num)
		stack.push(value)
		if stack.flag {
			stack.collector.OpResult = value.String()
		}
	}
	interpreter.intPool.put(back)
	return nil, nil
}

func opNot(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String())
	}

	math.U256(x.Not(x))
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

func opLt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opGt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opSlt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(1)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(0)

	default:
		if x.Cmp(y) < 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opSgt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(0)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(1)

	default:
		if x.Cmp(y) > 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	interpreter.intPool.put(x)

	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opEq(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	if x.Cmp(y) == 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opIszero(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String())
	}
	if x.Sign() > 0 {
		x.SetUint64(0)
	} else {
		x.SetUint64(1)
	}
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

func opAnd(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	stack.push(x.And(x, y))

	interpreter.intPool.put(y)
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

func opOr(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	y.Or(x, y)

	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opXor(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String())
	}
	y.Xor(x, y)

	interpreter.intPool.put(x)
	if stack.flag {
		stack.collector.OpResult = y.String()
	}
	return nil, nil
}

func opByte(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	th, val := stack.pop(), stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, th.String(), val.String())
	}
	if th.Cmp(common.Big32) < 0 {
		b := math.Byte(val, 32, int(th.Int64()))
		val.SetUint64(uint64(b))
	} else {
		val.SetUint64(0)
	}
	interpreter.intPool.put(th)
	if stack.flag {
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String(), z.String())
	}
	if z.Cmp(bigZero) > 0 {
		x.Add(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}
	interpreter.intPool.put(y, z)
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, x.String(), y.String(), z.String())
	}
	if z.Cmp(bigZero) > 0 {
		x.Mul(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}
	interpreter.intPool.put(y, z)
	if stack.flag {
		stack.collector.OpResult = x.String()
	}
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, shift.String(), value.String())
	}
	defer interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Lsh(value, n))
	if stack.flag {
		stack.collector.OpResult = value.String()
	}
	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, shift.String(), value.String())
	}
	defer interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Rsh(value, n))
	if stack.flag {
		stack.collector.OpResult = value.String()
	}
	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, S256 returns (potentially) a new bigint, so we're popping, not peeking this one
	shift, value := math.U256(stack.pop()), math.S256(stack.pop())
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, shift.String(), value.String())
	}
	defer interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		if value.Sign() >= 0 {
			value.SetUint64(0)
		} else {
			value.SetInt64(-1)
		}
		stack.push(math.U256(value))
		if stack.flag {
			stack.collector.OpResult = value.String()
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.Rsh(value, n)
	stack.push(math.U256(value))
	if stack.flag {
		stack.collector.OpResult = value.String()
	}
	return nil, nil
}

func opSha3(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	data := memory.Get(offset.Int64(), size.Int64())

	if interpreter.hasher == nil {
		interpreter.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		interpreter.hasher.Reset()
	}
	interpreter.hasher.Write(data)
	interpreter.hasher.Read(interpreter.hasherBuf[:])

	evm := interpreter.evm
	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(interpreter.hasherBuf, data)
	}
	res := interpreter.intPool.get().SetBytes(interpreter.hasherBuf[:])
	stack.push(res)

	interpreter.intPool.put(offset, size)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, offset.String(), size.String())
		stack.collector.ByteArgs = data
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opAddress(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := interpreter.intPool.get().SetBytes(contract.Address().Bytes())
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opBalance(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, slot.String())
	}
	slot.Set(interpreter.evm.StateDB.GetBalance(common.BigToAddress(slot)))
	if stack.flag {
		stack.collector.OpResult = slot.String()
	}
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	val := interpreter.intPool.get().SetBytes(interpreter.evm.Origin.Bytes())
	stack.push(val)
	if stack.flag {
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opCaller(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	val := interpreter.intPool.get().SetBytes(contract.Caller().Bytes())
	stack.push(val)
	if stack.flag {
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opCallValue(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	val := interpreter.intPool.get().Set(contract.value)
	stack.push(val)
	if stack.flag {
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	p := stack.pop()
	by := getDataBig(contract.Input, p, big32)
	res := interpreter.intPool.get().SetBytes(by)
	stack.push(res)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, p.String())
		stack.collector.ByteArgs = by
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := interpreter.intPool.get().SetInt64(int64(len(contract.Input)))
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	pre := memory.Get(memOffset.Int64(), length.Int64())
	res := getDataBig(contract.Input, dataOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), res)

	interpreter.intPool.put(memOffset, dataOffset, length)
	if stack.flag {
		stack.collector.PreArgs = pre
		stack.collector.RetArgs = res
		stack.collector.OpArgs = append(stack.collector.OpArgs, memOffset.String(), dataOffset.String(), length.String())
	}
	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := interpreter.intPool.get().SetUint64(uint64(len(interpreter.returnData)))
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()

		end = interpreter.intPool.get().Add(dataOffset, length)
	)
	defer interpreter.intPool.put(memOffset, dataOffset, length, end)

	if !end.IsUint64() || uint64(len(interpreter.returnData)) < end.Uint64() {
		return nil, errReturnDataOutOfBounds
	}
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, memOffset.String(), dataOffset.String(), length.String())
		stack.collector.PreArgs = memory.Get(memOffset.Int64(), length.Int64())
	}
	res := interpreter.returnData[dataOffset.Uint64():end.Uint64()]
	memory.Set(memOffset.Uint64(), length.Uint64(), res)
	if stack.flag {
		stack.collector.RetArgs = res
	}
	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, slot.String())
	}
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(common.BigToAddress(slot))))
	if stack.flag {
		stack.collector.OpResult = slot.String()
	}
	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	l := interpreter.intPool.get().SetInt64(int64(len(contract.Code)))
	stack.push(l)
	if stack.flag {
		stack.collector.OpResult = l.String()
	}
	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	if stack.flag {
		stack.collector.PreArgs = memory.Get(memOffset.Int64(), length.Int64())
	}
	codeCopy := getDataBig(contract.Code, codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	interpreter.intPool.put(memOffset, codeOffset, length)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, memOffset.String(), codeOffset.String(), length.String())
		stack.collector.RetArgs = codeCopy
	}
	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		addr       = common.BigToAddress(stack.pop())
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	if stack.flag {
		stack.collector.PreArgs = memory.Get(memOffset.Int64(), length.Int64())
	}
	codeCopy := getDataBig(interpreter.evm.StateDB.GetCode(addr), codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	interpreter.intPool.put(memOffset, codeOffset, length)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, addr.String(), memOffset.String(), codeOffset.String(), length.String())
		stack.collector.RetArgs = codeCopy
	}
	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal contract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without contract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 wei, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	address := common.BigToAddress(slot)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, slot.String())
	}
	if interpreter.evm.StateDB.Empty(address) {
		slot.SetUint64(0)
	} else {
		slot.SetBytes(interpreter.evm.StateDB.GetCodeHash(address).Bytes())
	}
	if stack.flag {
		stack.collector.OpResult = slot.String()
	}
	return nil, nil
}

func opGasprice(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := interpreter.intPool.get().Set(interpreter.evm.GasPrice)
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	num := stack.pop()

	n := interpreter.intPool.get().Sub(interpreter.evm.BlockNumber, common.Big257)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, num.String())
	}
	var p *big.Int
	if num.Cmp(n) > 0 && num.Cmp(interpreter.evm.BlockNumber) < 0 {
		p = interpreter.evm.GetHash(num.Uint64()).Big()
	} else {
		p = interpreter.intPool.getZero()
	}
	stack.push(p)
	interpreter.intPool.put(num, n)
	if stack.flag {
		stack.collector.OpResult = p.String()
	}
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := interpreter.intPool.get().SetBytes(interpreter.evm.Coinbase.Bytes())
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := math.U256(interpreter.intPool.get().Set(interpreter.evm.Time))
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opNumber(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := math.U256(interpreter.intPool.get().Set(interpreter.evm.BlockNumber))
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := math.U256(interpreter.intPool.get().Set(interpreter.evm.Difficulty))
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := math.U256(interpreter.intPool.get().SetUint64(interpreter.evm.GasLimit))
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opPop(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := stack.pop()
	interpreter.intPool.put(res)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, res.String())
	}
	return nil, nil
}

func opMload(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset := stack.pop()
	val := interpreter.intPool.get().SetBytes(memory.Get(offset.Int64(), 32))
	stack.push(val)

	interpreter.intPool.put(offset)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, offset.String())
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opMstore(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()

	if stack.flag {
		stack.collector.PreArgs = memory.Get(mStart.Int64(), 32)
		stack.collector.OpArgs = append(stack.collector.OpArgs, mStart.String(), val.String())
	}
	memory.Set32(mStart.Uint64(), val)

	interpreter.intPool.put(mStart, val)
	if stack.flag {
		stack.collector.RetArgs = memory.Get(mStart.Int64(), 32)
	}
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	if stack.flag {
		stack.collector.PreArgs = []byte{memory.store[off]}
	}
	res := byte(val & 0xff)
	memory.store[off] = res
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, (big.NewInt(off)).String(), (big.NewInt(val)).String())
		stack.collector.RetArgs = []byte{res}
	}
	return nil, nil
}

func opSload(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := stack.peek()
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, loc.String())
	}
	val := interpreter.evm.StateDB.GetState(contract.Address(), common.BigToHash(loc))
	loc.SetBytes(val.Bytes())
	if stack.flag {
		stack.collector.OpResult = loc.String()
	}

	return nil, nil
}

func opSstore(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()
	if stack.flag {
		stack.collector.OpValuePre = interpreter.evm.StateDB.GetState(contract.Address(), loc).Big().String()
	}
	interpreter.evm.StateDB.SetState(contract.Address(), loc, common.BigToHash(val))
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, loc.String(), val.String())
		stack.collector.OpValueNow = val.String()
	}
	interpreter.intPool.put(val)
	return nil, nil
}

func opJump(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos := stack.pop()
	if !contract.validJumpdest(pos) {
		return nil, errInvalidJump
	}
	*pc = pos.Uint64()

	interpreter.intPool.put(pos)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, pos.String())
	}
	return nil, nil
}

func opJumpi(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos, cond := stack.pop(), stack.pop()
	if cond.Sign() != 0 {
		if !contract.validJumpdest(pos) {
			return nil, errInvalidJump
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	interpreter.intPool.put(pos, cond)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, pos.String(), cond.String())
	}
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opPc(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	res := interpreter.intPool.get().SetUint64(*pc)
	stack.push(res)
	if stack.flag {
		stack.collector.OpResult = res.String()
	}
	return nil, nil
}

func opMsize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	val := interpreter.intPool.get().SetInt64(int64(memory.Len()))
	stack.push(val)
	if stack.flag {
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opGas(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	val := interpreter.intPool.get().SetUint64(contract.Gas)
	stack.push(val)
	if stack.flag {
		stack.collector.OpResult = val.String()
	}
	return nil, nil
}

func opCreate(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		value        = stack.pop()
		offset, size = stack.pop(), stack.pop()
		input        = memory.Get(offset.Int64(), size.Int64())
		gas          = contract.Gas
	)
	if interpreter.evm.ChainConfig().IsEIP150(interpreter.evm.BlockNumber) {
		gas -= gas / 64
	}
	contract.UseGas(gas)

	//add new 
	if stack.flag {
		stack.collector.Op = "CREATESTART"
		stack.collector.CallLayer = tingrong.CALL_LAYER + 1
		stack.collector.CallContract = ""
		stack.collector.OpArgs = append(stack.collector.OpArgs, value.String(), offset.String(), size.String())
		stack.collector.ByteArgs = input
		stack.collector.Value = value.String()
		stack.collector.GasTmp = fmt.Sprintf("%v", stack.collector.GasUsed)
		stack.collector.From = contract.Address().String()
		data := stack.collector.SendInsInfo()
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(stack.collector.Op, data)
	}
	//add new 

	res, addr, returnGas, suberr := interpreter.evm.Create(contract, input, gas, value)

	//add new 
	if stack.flag{
		temp_gas := stack.collector.GasUsed
		stack.collector = collector.NewCollector()
		stack.collector.GasUsed = temp_gas - returnGas
	}
	//add new 

	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	var p *big.Int
	if interpreter.evm.ChainConfig().IsHomestead(interpreter.evm.BlockNumber) && suberr == ErrCodeStoreOutOfGas {
		p = interpreter.intPool.getZero()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		p = interpreter.intPool.getZero()
	} else {
		p = interpreter.intPool.get().SetBytes(addr.Bytes())
	}
	stack.push(p)
	contract.Gas += returnGas
	interpreter.intPool.put(value, offset, size)
	
	//add new 
	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_CREATE") {
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_CREATE"
		invokeinfo.Pc = *pc
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = addr.String()
		invokeinfo.Value = value.String()
		invokeinfo.CallType = "CREATE"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		invokeinfo.CallLayer,_ = strconv.Atoi(temp_arr[1])
		
		createcollector := collector.NewCreateCollector()
		createcollector.ContractAddr = addr.String()
		createcollector.ContractDeployCode = input
		createcollector.ContractRuntimeCode = res
		invokeinfo.CreateInfo = *createcollector
		
		invokeinfo.IsSuccess = (suberr != nil)
		interpreter.evm.ChainConfig().TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}
	if stack.flag {
		stack.collector.Op = "CREATEEND"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		stack.collector.CallLayer,_ = strconv.Atoi(temp_arr[1])
		stack.collector.CallContract = addr.String()
		stack.collector.OpResult = p.String()
		stack.collector.RetArgs = res
		stack.collector.To = addr.String()
	}
	
	if interpreter.evm.isTxStart{
		tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
		// fmt.Println("option:createend")
		// fmt.Println("contract:",addr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	//add new 

	if suberr == errExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCreate2(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		endowment    = stack.pop()
		offset, size = stack.pop(), stack.pop()
		salt         = stack.pop()
		input        = memory.Get(offset.Int64(), size.Int64())
		gas          = contract.Gas
	)

	// Apply EIP150
	gas -= gas / 64
	contract.UseGas(gas)

	//add new 
	if stack.flag {
		stack.collector.Op = "CREATE2START"
		stack.collector.CallLayer = tingrong.CALL_LAYER + 1
		stack.collector.CallContract = ""
		stack.collector.OpArgs = append(stack.collector.OpArgs, endowment.String(), offset.String(), size.String(), salt.String())
		stack.collector.ByteArgs = input
		stack.collector.GasTmp = fmt.Sprintf("%v", stack.collector.GasUsed)
		stack.collector.Value = endowment.String()
		stack.collector.From = contract.Address().String()
		data := stack.collector.SendInsInfo()
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(stack.collector.Op, data)
	}
	//add new 

	res, addr, returnGas, suberr := interpreter.evm.Create2(contract, input, gas, endowment, salt)

	//add new 
	if stack.flag{
		temp_gas := stack.collector.GasUsed
		stack.collector = collector.NewCollector()
		stack.collector.GasUsed = temp_gas - returnGas
	}
	//add new 

	// Push item on the stack based on the returned error.
	var p *big.Int
	if suberr != nil {
		p = interpreter.intPool.getZero()
	} else {
		p = interpreter.intPool.get().SetBytes(addr.Bytes())
	}
	stack.push(p)
	contract.Gas += returnGas
	interpreter.intPool.put(endowment, offset, size, salt)

	//add new 
	//add new 
	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_CREATE2") {
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_CREATE2"
		invokeinfo.Pc = *pc
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = addr.String()
		invokeinfo.Value = endowment.String()
		invokeinfo.CallType = "CREATE"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		invokeinfo.CallLayer,_ = strconv.Atoi(temp_arr[1])		
		createcollector := collector.NewCreateCollector()
		createcollector.ContractAddr = addr.String()
		createcollector.ContractDeployCode = input
		createcollector.ContractRuntimeCode = res
		invokeinfo.CreateInfo = *createcollector	
		invokeinfo.IsSuccess = (suberr != nil)
		interpreter.evm.ChainConfig().TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}
	if stack.flag {
		stack.collector.Op = "CREATE2END"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		stack.collector.CallLayer,_ = strconv.Atoi(temp_arr[1])
		stack.collector.CallContract = addr.String()
		stack.collector.OpResult = p.String()
		stack.collector.RetArgs = res
		stack.collector.To = addr.String()
	}
	
	if interpreter.evm.isTxStart{
		tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
		// fmt.Println("option:create2end")
		// fmt.Println("contract:",addr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
		// tingrong.STACK_FLAG = false
	}
	//add new 

	if suberr == errExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCall(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in interpreter.evm.callGasTemp.
	pop := stack.pop()
	interpreter.intPool.put(pop)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BigToAddress(addr)
	value = math.U256(value)
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())
	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	//add new 
	
	if interpreter.evm.isTxStart{
		tingrong.CALL_LAYER += 1
		tingrong.CALL_STACK = append(tingrong.CALL_STACK,toAddr.String()+"#"+strconv.Itoa(tingrong.CALL_LAYER))
		tingrong.ALL_STACK = append(tingrong.ALL_STACK,toAddr.String())
		// fmt.Println("option:CALLSTART")
		// fmt.Println("contract:",toAddr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}

	
	if stack.flag {
		stack.collector.Op = "CALLSTART"
		stack.collector.CallLayer = tingrong.CALL_LAYER
		stack.collector.CallContract = toAddr.String()
		stack.collector.From = contract.Address().String()
		stack.collector.To = toAddr.String()
		stack.collector.Value = value.String()
		stack.collector.OpArgs = append(stack.collector.OpArgs, pop.String(), inOffset.String(), inSize.String(), retOffset.String(), retSize.String())
		stack.collector.ByteArgs = args
		stack.collector.GasTmp = fmt.Sprintf("%v", stack.collector.GasUsed)
		stack.collector.ByteCode = interpreter.evm.StateDB.GetCode(toAddr)
		data := stack.collector.SendInsInfo()
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(stack.collector.Op, data)
	}
	
	//add new 

	ret, returnGas, err := interpreter.evm.Call(contract, toAddr, args, gas, value)

	//add new 
	if stack.flag{
		temp_gas := stack.collector.GasUsed
		stack.collector = collector.NewCollector()
		stack.collector.GasUsed = temp_gas - returnGas
	}
	//add new 

	var p *big.Int
	if err != nil {
		p = interpreter.intPool.getZero()
		//add new 
		if stack.flag {
			stack.collector.InternalErr = err.Error()
			stack.collector.IsInternalSucceeded = false
		}
		//add new 
	} else {
		p = interpreter.intPool.get().SetUint64(1)
		//add new 
		if stack.flag {
			stack.collector.IsInternalSucceeded = true
		}
		//add new 
	}
	stack.push(p)
	if err == nil || err == errExecutionReverted {
		//add new 
		if stack.flag {
			stack.collector.PreArgs = memory.Get(retOffset.Int64(), retSize.Int64())
		}
		//add new 
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		//add new 
		if stack.flag {
			stack.collector.RetArgs = ret
		}
		//add new 
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, value, inOffset, inSize, retOffset, retSize)
	
	//add new 

	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_CALL") {
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_CALL"
		invokeinfo.Pc = *pc
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = toAddr.String()
		invokeinfo.Value = value.String()
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		invokeinfo.CallLayer,_ = strconv.Atoi(temp_arr[1])

		invokeinfo.CallType = "CALL"
		callcollector := collector.NewCallCollector()
		callcollector.ContractCode = interpreter.evm.StateDB.GetCode(toAddr)	
		callcollector.InputData = args		
		invokeinfo.CallInfo = *callcollector
		invokeinfo.IsSuccess = (err==nil) 
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}

	if stack.flag {
		stack.collector.Op = "CALLEND"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		stack.collector.CallLayer,_ = strconv.Atoi(temp_arr[1])
		stack.collector.CallContract = toAddr.String()
		stack.collector.OpResult = p.String()
		//stack.collector.IsInternalSucceeded = tingrong.CALLVALID_MAP[tingrong.CALL_LAYER] && stack.collector.IsInternalSucceeded
		stack.collector.IsCallValid = tingrong.CALLVALID_MAP[tingrong.CALL_LAYER]
	}
	
	if interpreter.evm.isTxStart {
		// fmt.Println("option:CALLEND")
		// fmt.Println("contract:",toAddr.String())
		tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}	
	//add new 

	return ret, nil
}

func opCallCode(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	pop := stack.pop()
	interpreter.intPool.put(pop)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BigToAddress(addr)
	value = math.U256(value)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	//add new 
	if interpreter.evm.isTxStart{
		tingrong.CALL_LAYER += 1
		tingrong.CALL_STACK = append(tingrong.CALL_STACK,toAddr.String()+"#"+strconv.Itoa(tingrong.CALL_LAYER))
		tingrong.ALL_STACK = append(tingrong.ALL_STACK,toAddr.String())
		// fmt.Println("option:CALLCODESTART")
		// fmt.Println("contract:",toAddr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	
	if stack.flag {
		stack.collector.Op = "CALLCODESTART"
		stack.collector.CallLayer = tingrong.CALL_LAYER
		stack.collector.CallContract = toAddr.String()
		stack.collector.From = contract.Address().String()
		stack.collector.To = toAddr.String()
		stack.collector.Value = value.String()
		stack.collector.OpArgs = append(stack.collector.OpArgs, pop.String(), inOffset.String(), inSize.String(), retOffset.String(), retSize.String())
		stack.collector.ByteArgs = args
		stack.collector.GasTmp = fmt.Sprintf("%v", stack.collector.GasUsed)
		stack.collector.ByteCode = interpreter.evm.StateDB.GetCode(toAddr)
		data := stack.collector.SendInsInfo()
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(stack.collector.Op, data)
	}
	//add new 

	ret, returnGas, err := interpreter.evm.CallCode(contract, toAddr, args, gas, value)

	//add new 
	if stack.flag{
		temp_gas := stack.collector.GasUsed
		stack.collector = collector.NewCollector()
		stack.collector.GasUsed = temp_gas - returnGas
	}
	//add new 

	var p *big.Int
	if err != nil {
		p = interpreter.intPool.getZero()
		//add new 
		if stack.flag {
			stack.collector.InternalErr = err.Error()
			stack.collector.IsInternalSucceeded = false
		}
		//add new 
	} else {
		p = interpreter.intPool.get().SetUint64(1)
		//add new 
		if stack.flag {
			stack.collector.IsInternalSucceeded = true
		}
		//add new 
	}
	stack.push(p)
	if err == nil || err == errExecutionReverted {
		//add new 
		if stack.flag {
			stack.collector.PreArgs = memory.Get(retOffset.Int64(), retSize.Int64())
		}
		//add new 
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		//add new 
		if stack.flag {
			stack.collector.RetArgs = ret
		}
		//add new 
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, value, inOffset, inSize, retOffset, retSize)
	//add new 

	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_CALLCODE") {
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_CALLCODE"
		invokeinfo.Pc = *pc
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = toAddr.String()
		invokeinfo.Value = value.String()
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		invokeinfo.CallLayer,_ = strconv.Atoi(temp_arr[1])

		invokeinfo.CallType = "CALL"
		callcollector := collector.NewCallCollector()
		callcollector.ContractCode = interpreter.evm.StateDB.GetCode(toAddr)	
		callcollector.InputData = args		
		invokeinfo.CallInfo = *callcollector
		invokeinfo.IsSuccess = (err==nil)
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}

	if stack.flag {
		stack.collector.Op = "CALLCODEEND"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		stack.collector.CallLayer,_ = strconv.Atoi(temp_arr[1])
		stack.collector.CallContract = toAddr.String()
		stack.collector.OpResult = p.String()	
	}

	if interpreter.evm.isTxStart{
		// fmt.Println("option:CALLCODEEND")
		// fmt.Println("contract:",toAddr.String())
		tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	//add new 
	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	pop := stack.pop()
	interpreter.intPool.put(pop)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BigToAddress(addr)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	//add new 
	if interpreter.evm.isTxStart{
		tingrong.CALL_LAYER += 1
		tingrong.CALL_STACK = append(tingrong.CALL_STACK,toAddr.String()+"#"+strconv.Itoa(tingrong.CALL_LAYER))
		tingrong.ALL_STACK = append(tingrong.ALL_STACK,toAddr.String())
		// fmt.Println("option:DELEGATECALLSTART")
		// fmt.Println("contract:",toAddr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	
	if stack.flag {
		stack.collector.Op = "DELEGATECALLSTART"
		stack.collector.CallLayer = tingrong.CALL_LAYER
		stack.collector.CallContract = toAddr.String()
		stack.collector.From = contract.Address().String()
		stack.collector.To = toAddr.String()
		stack.collector.OpArgs = append(stack.collector.OpArgs, pop.String(), toAddr.String(), inOffset.String(), inSize.String(), retOffset.String(), retSize.String())
		stack.collector.ByteArgs = args
		stack.collector.GasTmp = fmt.Sprintf("%v", stack.collector.GasUsed)
		stack.collector.ByteCode = interpreter.evm.StateDB.GetCode(toAddr)
		data := stack.collector.SendInsInfo()
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(stack.collector.Op, data)
	}
	//add new 

	ret, returnGas, err := interpreter.evm.DelegateCall(contract, toAddr, args, gas)
	
	//add new 
	if stack.flag{
		temp_gas := stack.collector.GasUsed
		stack.collector = collector.NewCollector()
		stack.collector.GasUsed = temp_gas - returnGas
	}
	//add new 
	
	var p *big.Int
	if err != nil {
		p = interpreter.intPool.getZero()
		//add new 
		if stack.flag {
			stack.collector.InternalErr = err.Error()
			stack.collector.IsInternalSucceeded = false
		}
		//add new 
	} else {
		p = interpreter.intPool.get().SetUint64(1)
		//add new 
		if stack.flag {
			stack.collector.IsInternalSucceeded = true
		}
		//add new 
	}
	stack.push(p)
	if err == nil || err == errExecutionReverted {
		//add new 
		if stack.flag {
			stack.collector.PreArgs = memory.Get(retOffset.Int64(), retSize.Int64())
		}
		//add new 
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		//add new 
		if stack.flag {
			stack.collector.RetArgs = ret
		}
		//add new 
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	//add new 
	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_DELEGATECALL") {
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_DELEGATECALL"
		invokeinfo.Pc = *pc
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = toAddr.String()
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		invokeinfo.CallLayer,_ = strconv.Atoi(temp_arr[1])

		invokeinfo.CallType = "CALL"
		callcollector := collector.NewCallCollector()
		callcollector.ContractCode = interpreter.evm.StateDB.GetCode(toAddr)	
		callcollector.InputData = args		
		invokeinfo.CallInfo = *callcollector
		invokeinfo.IsSuccess = (err==nil)
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}
	if stack.flag {
		stack.collector.Op = "DELEGATECALLEND"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		stack.collector.CallLayer,_ = strconv.Atoi(temp_arr[1])
		stack.collector.CallContract = toAddr.String()
		stack.collector.OpResult = p.String()	
	}

	if interpreter.evm.isTxStart {
		tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
		// fmt.Println("option:DELEGATECALLEND")
		// fmt.Println("contract:",toAddr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	//add new 
	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	pop := stack.pop()
	interpreter.intPool.put(pop)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BigToAddress(addr)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	//add new 
	if interpreter.evm.isTxStart{
		tingrong.CALL_LAYER += 1
		tingrong.CALL_STACK = append(tingrong.CALL_STACK,toAddr.String()+"#"+strconv.Itoa(tingrong.CALL_LAYER))
		tingrong.ALL_STACK = append(tingrong.ALL_STACK,toAddr.String())
		// fmt.Println("option:STATICCALLSTART")
		// fmt.Println("contract:",toAddr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	
	if stack.flag {
		stack.collector.Op = "STATICCALLSTART"
		stack.collector.CallLayer = tingrong.CALL_LAYER
		stack.collector.CallContract = toAddr.String()
		stack.collector.From = contract.Address().String()
		stack.collector.To = toAddr.String()
		stack.collector.OpArgs = append(stack.collector.OpArgs, pop.String(), toAddr.String(), inOffset.String(), inSize.String(), retOffset.String(), retSize.String())
		stack.collector.ByteArgs = args
		stack.collector.GasTmp = fmt.Sprintf("%v", stack.collector.GasUsed)
		stack.collector.ByteCode = interpreter.evm.StateDB.GetCode(toAddr)
		data := stack.collector.SendInsInfo()
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(stack.collector.Op, data)
	}
	//add new 

	ret, returnGas, err := interpreter.evm.StaticCall(contract, toAddr, args, gas)
	//add new 
	if stack.flag{
		temp_gas := stack.collector.GasUsed
		stack.collector = collector.NewCollector()
		stack.collector.GasUsed = temp_gas - returnGas
	}
	//add new 
	var p *big.Int
	if err != nil {
		p = interpreter.intPool.getZero()
		//add new 
		if stack.flag {
			stack.collector.InternalErr = err.Error()
			stack.collector.IsInternalSucceeded = false
		}
		//add new 
	} else {
		p = interpreter.intPool.get().SetUint64(1)
		//add new 
		if stack.flag {
			stack.collector.IsInternalSucceeded = true
		}
		//add new 
	}
	stack.push(p)
	if err == nil || err == errExecutionReverted {
		//add new 
		if stack.flag {
			stack.collector.PreArgs = memory.Get(retOffset.Int64(), retSize.Int64())
		}
		//add new 
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		//add new 
		if stack.flag {
			stack.collector.RetArgs = ret
		}
		//add new 
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	//add new 
	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_STATICCALL") {
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_STATICCALL"
		invokeinfo.Pc = *pc
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = toAddr.String()
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		invokeinfo.CallLayer,_ = strconv.Atoi(temp_arr[1])

		invokeinfo.CallType = "CALL"
		callcollector := collector.NewCallCollector()
		callcollector.ContractCode = interpreter.evm.StateDB.GetCode(toAddr)	
		callcollector.InputData = args		
		invokeinfo.CallInfo = *callcollector
		invokeinfo.IsSuccess = (err==nil)
		interpreter.evm.chainConfig.TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}
	if stack.flag {
		stack.collector.Op = "STATICCALLEND"
		temp_str := tingrong.CALL_STACK[len(tingrong.CALL_STACK)-1]
		temp_arr := strings.Split(temp_str,"#")
		stack.collector.CallLayer,_ = strconv.Atoi(temp_arr[1])
		stack.collector.CallContract = toAddr.String()
		stack.collector.OpResult = p.String()	
	}

	if interpreter.evm.isTxStart{
		tingrong.CALL_STACK = tingrong.CALL_STACK[:len(tingrong.CALL_STACK)-1]
		// fmt.Println("option:STATICCALLEND")
		// fmt.Println("contract:",toAddr.String())
		// fmt.Println("layer:",tingrong.CALL_LAYER)
		// fmt.Println("stack:",tingrong.CALL_STACK)
	}
	
	//add new 
	return ret, nil
}

func opReturn(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	interpreter.intPool.put(offset, size)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, offset.String(), size.String())
		stack.collector.RetArgs = ret
	}
	return ret, nil
}

func opRevert(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	interpreter.intPool.put(offset, size)
	if stack.flag {
		stack.collector.OpArgs = append(stack.collector.OpArgs, offset.String(), size.String())
		stack.collector.RetArgs = ret
	}
	return ret, nil
}

func opStop(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	balance := interpreter.evm.StateDB.GetBalance(contract.Address())
	toAddr := common.BigToAddress(stack.pop())
	interpreter.evm.StateDB.AddBalance(toAddr, balance)

	interpreter.evm.StateDB.Suicide(contract.Address())
	if stack.flag {
		stack.collector.Value = balance.String()
		stack.collector.From = contract.Address().String()
		stack.collector.To = toAddr.String()
	}
	if interpreter.evm.isTxStart && interpreter.evm.ChainConfig().TransferDataPlg.GetOpcodeRegister("TRANS_SUICIDE"){
		invokeinfo := collector.NewTransCollector()
		invokeinfo.Op = "TRANS_SUICIDE"
		invokeinfo.From = contract.Address().String()
		invokeinfo.To = toAddr.String()
		invokeinfo.Value = balance.String()
		interpreter.evm.ChainConfig().TransferDataPlg.SendDataToPlugin(invokeinfo.Op, invokeinfo.SendTransInfo(invokeinfo.Op))
	}
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i] = common.BigToHash(stack.pop())
		}

		d := memory.Get(mStart.Int64(), mSize.Int64())
		interpreter.evm.StateDB.AddLog(&types.Log{
			Address: contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: interpreter.evm.BlockNumber.Uint64(),
		})

		interpreter.intPool.put(mStart, mSize)
		if stack.flag {
			stack.collector.OpArgs = append(stack.collector.OpArgs, mStart.String(), mSize.String())
			for i := 0; i < size; i++ {
				stack.collector.OpArgs = append(stack.collector.OpArgs, topics[i].String())
			}

			stack.collector.RetArgs = d
		}
		return nil, nil
	}
}

// opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		codeLen = uint64(len(contract.Code))
		integer = interpreter.intPool.get()
	)
	*pc += 1
	var res *big.Int
	if *pc < codeLen {
		res = integer.SetUint64(uint64(contract.Code[*pc]))
	} else {
		res = integer.SetUint64(0)
	}
	stack.push(res)
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		codeLen := len(contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := interpreter.intPool.get()
		value := contract.Code[startMin:endMin]
		p := integer.SetBytes(common.RightPadBytes(value, pushByteSize))
		stack.push(p)

		*pc += size
		if stack.flag {
			stack.collector.OpResult = p.String()
			stack.collector.RetArgs = value
			stack.collector.PcNext = fmt.Sprintf("%v", *pc)
		}
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.dup(interpreter.intPool, int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.swap(int(size))
		return nil, nil
	}
}
