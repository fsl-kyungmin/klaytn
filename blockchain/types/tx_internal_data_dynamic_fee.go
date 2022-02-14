// Copyright 2022 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/klaytn/klaytn/blockchain/types/accountkey"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/common/hexutil"
	"github.com/klaytn/klaytn/crypto"
	"github.com/klaytn/klaytn/fork"
	"github.com/klaytn/klaytn/kerrors"
	"github.com/klaytn/klaytn/params"
	"github.com/klaytn/klaytn/rlp"
)

type TxInternalDataDynamicFee struct {
	ChainID      *big.Int
	AccountNonce uint64
	GasTipCap    *big.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap    *big.Int // a.k.a. maxFeePerGas
	GasLimit     uint64
	Recipient    *common.Address `rlp:"nil"` // nil means contract creation
	Amount       *big.Int
	Payload      []byte
	AccessList   AccessList

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type TxInternalDataDynamicFeeJSON struct {
	Type                 TxType           `json:"typeInt"`
	TypeStr              string           `json:"type"`
	AccountNonce         hexutil.Uint64   `json:"nonce"`
	MaxPriorityFeePerGas *hexutil.Big     `json:"maxPriorityFeePerGas"`
	MaxFeePerGas         *hexutil.Big     `json:"maxFeePerGas"`
	GasLimit             hexutil.Uint64   `json:"gas"`
	Recipient            *common.Address  `json:"to"`
	Amount               *hexutil.Big     `json:"value"`
	Payload              hexutil.Bytes    `json:"input"`
	TxSignatures         TxSignaturesJSON `json:"signatures"`

	AccessList AccessList   `json:"accessList"`
	ChainID    *hexutil.Big `json:"chainId"`

	Hash *common.Hash `json:"hash"`
}

func newEmptyTxInternalDataDynamicFee() *TxInternalDataDynamicFee {
	return &TxInternalDataDynamicFee{}
}

func newTxInternalDataDynamicFee() *TxInternalDataDynamicFee {
	return &TxInternalDataDynamicFee{
		ChainID:      new(big.Int),
		AccountNonce: 0,
		GasTipCap:    new(big.Int),
		GasFeeCap:    new(big.Int),
		GasLimit:     0,
		Recipient:    nil,
		Amount:       new(big.Int),
		Payload:      []byte{},
		AccessList:   AccessList{},

		// Signature values
		V: new(big.Int),
		R: new(big.Int),
		S: new(big.Int),
	}
}

func newTxInternalDataDynamicFeeWithValues(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasTipCap *big.Int, gasFeeCap *big.Int, data []byte, accessList AccessList, chainID *big.Int) *TxInternalDataDynamicFee {
	d := newTxInternalDataDynamicFee()

	d.AccountNonce = nonce
	d.Recipient = to
	d.GasLimit = gasLimit

	if len(data) > 0 {
		d.Payload = common.CopyBytes(data)
	}

	if amount != nil {
		d.Amount.Set(amount)
	}

	if gasTipCap != nil {
		d.GasTipCap.Set(gasTipCap)
	}

	if gasFeeCap != nil {
		d.GasFeeCap.Set(gasFeeCap)
	}

	if accessList != nil {
		copy(d.AccessList, accessList)
	}

	if chainID != nil {
		d.ChainID.Set(chainID)
	}

	return d
}

func newTxInternalDataDynamicFeeWithMap(values map[TxValueKeyType]interface{}) (*TxInternalDataDynamicFee, error) {
	d := newTxInternalDataDynamicFee()

	if v, ok := values[TxValueKeyNonce].(uint64); ok {
		d.AccountNonce = v
		delete(values, TxValueKeyNonce)
	} else {
		return nil, errValueKeyNonceMustUint64
	}

	if v, ok := values[TxValueKeyTo].(common.Address); ok {
		d.Recipient = &v
		delete(values, TxValueKeyTo)
	} else {
		return nil, errValueKeyToMustAddress
	}

	if v, ok := values[TxValueKeyAmount].(*big.Int); ok {
		d.Amount.Set(v)
		delete(values, TxValueKeyAmount)
	} else {
		return nil, errValueKeyAmountMustBigInt
	}

	if v, ok := values[TxValueKeyData].([]byte); ok {
		d.Payload = common.CopyBytes(v)
		delete(values, TxValueKeyData)
	} else {
		return nil, errValueKeyDataMustByteSlice
	}

	if v, ok := values[TxValueKeyGasLimit].(uint64); ok {
		d.GasLimit = v
		delete(values, TxValueKeyGasLimit)
	} else {
		return nil, errValueKeyGasLimitMustUint64
	}

	if v, ok := values[TxValueKeyGasFeeCap].(*big.Int); ok {
		d.GasFeeCap.Set(v)
		delete(values, TxValueKeyGasFeeCap)
	} else {
		return nil, errValueKeyGasFeeCapMustBigInt
	}
	if v, ok := values[TxValueKeyGasTipCap].(*big.Int); ok {
		d.GasTipCap.Set(v)
		delete(values, TxValueKeyGasTipCap)
	} else {
		return nil, errValueKeyGasTipCapMustBigInt
	}
	if v, ok := values[TxValueKeyAccessList].(AccessList); ok {
		d.AccessList = make(AccessList, len(v))
		copy(d.AccessList, v)
		delete(values, TxValueKeyAccessList)
	} else {
		return nil, errValueKeyAccessListInvalid
	}

	if v, ok := values[TxValueKeyChainID].(*big.Int); ok {
		d.ChainID.Set(v)
		delete(values, TxValueKeyChainID)
	} else {
		return nil, errValueKeyChainIDInvalid
	}

	if len(values) != 0 {
		for k := range values {
			logger.Warn("unnecessary key", k.String())
		}
		return nil, errUndefinedKeyRemains
	}

	return d, nil
}

func (t *TxInternalDataDynamicFee) Type() TxType {
	return TxTypeDynamicFee
}

func (t *TxInternalDataDynamicFee) GetRoleTypeForValidation() accountkey.RoleType {
	return accountkey.RoleTransaction
}

func (t *TxInternalDataDynamicFee) GetAccountNonce() uint64 {
	return t.AccountNonce
}

func (t *TxInternalDataDynamicFee) GetPrice() *big.Int {
	return t.GasFeeCap
}

func (t *TxInternalDataDynamicFee) GetGasLimit() uint64 {
	return t.GasLimit
}

func (t *TxInternalDataDynamicFee) GetRecipient() *common.Address {
	return t.Recipient
}

func (t *TxInternalDataDynamicFee) GetAmount() *big.Int {
	return new(big.Int).Set(t.Amount)
}

func (t *TxInternalDataDynamicFee) GetHash() *common.Hash {
	return t.Hash
}

func (t *TxInternalDataDynamicFee) GetPayload() []byte {
	return t.Payload
}

func (t *TxInternalDataDynamicFee) GetAccessList() AccessList {
	return t.AccessList
}

func (t *TxInternalDataDynamicFee) GetGasTipCap() *big.Int {
	return t.GasTipCap
}

func (t *TxInternalDataDynamicFee) GetGasFeeCap() *big.Int {
	return t.GasFeeCap
}

func (t *TxInternalDataDynamicFee) SetHash(hash *common.Hash) {
	t.Hash = hash
}

func (t *TxInternalDataDynamicFee) SetSignature(signatures TxSignatures) {
	if len(signatures) != 1 {
		logger.Crit("TxTypeDynamicFee can receive only single signature!")
	}

	t.V = signatures[0].V
	t.R = signatures[0].R
	t.S = signatures[0].S
}

func (t *TxInternalDataDynamicFee) RawSignatureValues() TxSignatures {
	return TxSignatures{&TxSignature{t.V, t.R, t.S}}
}

func (t *TxInternalDataDynamicFee) ValidateSignature() bool {
	v := byte(t.V.Uint64())
	return crypto.ValidateSignatureValues(v, t.R, t.S, false)
}

func (t *TxInternalDataDynamicFee) RecoverAddress(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) (common.Address, error) {
	V := vfunc(t.V)
	return recoverPlain(txhash, t.R, t.S, V, homestead)
}

func (t *TxInternalDataDynamicFee) RecoverPubkey(txhash common.Hash, homestead bool, vfunc func(*big.Int) *big.Int) ([]*ecdsa.PublicKey, error) {
	V := vfunc(t.V)

	pk, err := recoverPlainPubkey(txhash, t.R, t.S, V, homestead)
	if err != nil {
		return nil, err
	}

	return []*ecdsa.PublicKey{pk}, nil
}

func (t *TxInternalDataDynamicFee) IntrinsicGas(currentBlockNumber uint64) (uint64, error) {
	return IntrinsicGas(t.Payload, t.AccessList, t.Recipient == nil, *fork.Rules(big.NewInt(int64(currentBlockNumber))))
}

func (t *TxInternalDataDynamicFee) ChainId() *big.Int {
	if t.ChainID != nil {
		return t.ChainID
	}
	return deriveChainId(t.V)
}

func (t *TxInternalDataDynamicFee) Equal(a TxInternalData) bool {
	ta, ok := a.(*TxInternalDataDynamicFee)
	if !ok {
		return false
	}

	return t.ChainID.Cmp(ta.ChainID) == 0 &&
		t.AccountNonce == ta.AccountNonce &&
		t.GasFeeCap.Cmp(ta.GasFeeCap) == 0 &&
		t.GasTipCap.Cmp(ta.GasTipCap) == 0 &&
		t.GasLimit == ta.GasLimit &&
		equalRecipient(t.Recipient, ta.Recipient) &&
		t.Amount.Cmp(ta.Amount) == 0 &&
		reflect.DeepEqual(t.AccessList, ta.AccessList) &&
		t.V.Cmp(ta.V) == 0 &&
		t.R.Cmp(ta.R) == 0 &&
		t.S.Cmp(ta.S) == 0

}

func (t *TxInternalDataDynamicFee) String() string {
	var from, to string
	tx := &Transaction{data: t}

	v, r, s := t.V, t.R, t.S

	signer := LatestSignerForChainID(t.ChainId())
	if f, err := Sender(signer, tx); err != nil { // derive but don't cache
		from = "[invalid sender: invalid sig]"
	} else {
		from = fmt.Sprintf("%x", f[:])
	}

	if t.GetRecipient() == nil {
		to = "[contract creation]"
	} else {
		to = fmt.Sprintf("%x", t.GetRecipient().Bytes())
	}
	enc, _ := rlp.EncodeToBytes(t)
	return fmt.Sprintf(`
		TX(%x)
		Contract: %v
		Chaind:   %#x
		From:     %s
		To:       %s
		Nonce:    %v
		GasTipCap: %#x
		GasFeeCap: %#x
		GasLimit  %#x
		Value:    %#x
		Data:     0x%x
		AccessList: %x
		V:        %#x
		R:        %#x
		S:        %#x
		Hex:      %x
	`,
		tx.Hash(),
		t.GetRecipient() == nil,
		t.ChainId(),
		from,
		to,
		t.GetAccountNonce(),
		t.GetGasTipCap(),
		t.GetGasFeeCap(),
		t.GetGasLimit(),
		t.GetAmount(),
		t.GetPayload(),
		t.AccessList,
		v,
		r,
		s,
		enc,
	)
}

func (t *TxInternalDataDynamicFee) SerializeForSign() []interface{} {
	return []interface{}{
		t.AccountNonce,
		t.GasTipCap,
		t.GasFeeCap,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.Payload,
		t.AccessList,
	}
}

func (t *TxInternalDataDynamicFee) SenderTxHash() common.Hash {
	return prefixedRlpHash(byte(t.Type()), []interface{}{
		t.ChainID,
		t.AccountNonce,
		t.GasTipCap,
		t.GasFeeCap,
		t.GasLimit,
		t.Recipient,
		t.Amount,
		t.Payload,
		t.AccessList,
		t.V,
		t.R,
		t.S,
	})
}

func (t *TxInternalDataDynamicFee) Validate(stateDB StateDB, currentBlockNumber uint64) error {
	if t.Recipient != nil {
		if common.IsPrecompiledContractAddress(*t.Recipient) {
			return kerrors.ErrPrecompiledContractAddress
		}
	}
	return t.ValidateMutableValue(stateDB, currentBlockNumber)
}

func (t *TxInternalDataDynamicFee) ValidateMutableValue(stateDB StateDB, currentBlockNumber uint64) error {
	return nil
}

func (t *TxInternalDataDynamicFee) IsLegacyTransaction() bool {
	return false
}

func (t *TxInternalDataDynamicFee) FillContractAddress(from common.Address, r *Receipt) {
	if t.Recipient == nil {
		r.ContractAddress = crypto.CreateAddress(from, t.AccountNonce)
	}
}

func (t *TxInternalDataDynamicFee) Execute(sender ContractRef, vm VM, stateDB StateDB, currentBlockNumber uint64, gas uint64, value *big.Int) (ret []byte, usedGas uint64, err error) {
	///////////////////////////////////////////////////////
	// OpcodeComputationCostLimit: The below code is commented and will be usd for debugging purposes.
	//start := time.Now()
	//defer func() {
	//	elapsed := time.Since(start)
	//	logger.Debug("[TxInternalDataLegacy] EVM execution done", "elapsed", elapsed)
	//}()
	///////////////////////////////////////////////////////
	if t.Recipient == nil {
		// Sender's nonce will be increased in '`vm.Create()`
		ret, _, usedGas, err = vm.Create(sender, t.Payload, gas, value, params.CodeFormatEVM)
	} else {
		stateDB.IncNonce(sender.Address())
		ret, usedGas, err = vm.Call(sender, *t.Recipient, t.Payload, gas, value)
	}
	return ret, usedGas, err
}

func (t *TxInternalDataDynamicFee) MakeRPCOutput() map[string]interface{} {
	return map[string]interface{}{
		"typeInt":              t.Type(),
		"chainId":              (*hexutil.Big)(t.ChainId()),
		"type":                 t.Type().String(),
		"gas":                  hexutil.Uint64(t.GasLimit),
		"maxPriorityFeePerGas": (*hexutil.Big)(t.GasTipCap),
		"maxFeePerGas":         (*hexutil.Big)(t.GasFeeCap),
		"input":                hexutil.Bytes(t.Payload),
		"nonce":                hexutil.Uint64(t.AccountNonce),
		"to":                   t.Recipient,
		"value":                (*hexutil.Big)(t.Amount),
		"accessList":           t.AccessList,
		"signatures":           TxSignaturesJSON{&TxSignatureJSON{(*hexutil.Big)(t.V), (*hexutil.Big)(t.R), (*hexutil.Big)(t.S)}},
	}
}

func (t *TxInternalDataDynamicFee) MarshalJSON() ([]byte, error) {
	return json.Marshal(TxInternalDataDynamicFeeJSON{
		t.Type(),
		t.Type().String(),
		(hexutil.Uint64)(t.AccountNonce),
		(*hexutil.Big)(t.GasTipCap),
		(*hexutil.Big)(t.GasFeeCap),
		(hexutil.Uint64)(t.GasLimit),
		t.Recipient,
		(*hexutil.Big)(t.Amount),
		t.Payload,
		TxSignaturesJSON{&TxSignatureJSON{(*hexutil.Big)(t.V), (*hexutil.Big)(t.R), (*hexutil.Big)(t.S)}},
		t.AccessList,
		(*hexutil.Big)(t.ChainID),
		t.Hash,
	})
}

func (t *TxInternalDataDynamicFee) UnmarshalJSON(bytes []byte) error {
	js := &TxInternalDataDynamicFeeJSON{}
	if err := json.Unmarshal(bytes, js); err != nil {
		return err
	}

	t.AccountNonce = uint64(js.AccountNonce)
	t.GasTipCap = (*big.Int)(js.MaxPriorityFeePerGas)
	t.GasFeeCap = (*big.Int)(js.MaxFeePerGas)
	t.GasLimit = uint64(js.GasLimit)
	t.Recipient = js.Recipient
	t.Amount = (*big.Int)(js.Amount)
	t.Payload = js.Payload
	t.AccessList = js.AccessList
	t.V = (*big.Int)(js.TxSignatures[0].V)
	t.R = (*big.Int)(js.TxSignatures[0].R)
	t.S = (*big.Int)(js.TxSignatures[0].S)
	t.Hash = js.Hash
	t.ChainID = (*big.Int)(js.ChainID)

	return nil
}

func (t *TxInternalDataDynamicFee) setSignatureValues(chainID, v, r, s *big.Int) {
	t.ChainID, t.V, t.R, t.S = chainID, v, r, s
}
