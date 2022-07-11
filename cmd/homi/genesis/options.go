// Copyright 2018 The klaytn Authors
// Copyright 2017 AMIS Technologies
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

package genesis

import (
	"math/big"

	"github.com/klaytn/klaytn/cmd/homi/extra"
	"github.com/klaytn/klaytn/consensus/clique"
	"github.com/klaytn/klaytn/contracts/reward/contract"
	"github.com/klaytn/klaytn/log"
	"github.com/klaytn/klaytn/params"

	"github.com/klaytn/klaytn/blockchain"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/common/hexutil"
)

type Option func(*blockchain.Genesis)

var logger = log.NewModuleLogger(log.CMDIstanbul)

func Validators(addrs ...common.Address) Option {
	return func(genesis *blockchain.Genesis) {
		extraData, err := extra.Encode("0x00", addrs)
		if err != nil {
			logger.Error("Failed to encode extra data", "err", err)
			return
		}
		genesis.ExtraData = hexutil.MustDecode(extraData)
	}
}

func ValidatorsOfClique(signers ...common.Address) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.ExtraData = make([]byte, clique.ExtraVanity+len(signers)*common.AddressLength+clique.ExtraSeal)
		for i, signer := range signers {
			copy(genesis.ExtraData[32+i*common.AddressLength:], signer[:])
		}
	}
}

func makeGenesisAccount(addrs []common.Address, balance *big.Int) map[common.Address]blockchain.GenesisAccount {
	alloc := make(map[common.Address]blockchain.GenesisAccount)
	for _, addr := range addrs {
		alloc[addr] = blockchain.GenesisAccount{Balance: balance}
	}
	return alloc
}

func Alloc(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		alloc := makeGenesisAccount(addrs, balance)
		genesis.Alloc = alloc
	}
}

func AllocWithPrecypressContract(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		alloc := makeGenesisAccount(addrs, balance)
		alloc[common.HexToAddress(contract.CypressCreditContractAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(CypressCreditBin),
			Balance: big.NewInt(0),
		}
		alloc[common.HexToAddress(contract.AddressBookContractAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(PreCypressAddressBookBin),
			Balance: big.NewInt(0),
		}
		genesis.Alloc = alloc
	}
}

func AllocWithCypressContract(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		alloc := makeGenesisAccount(addrs, balance)
		alloc[common.HexToAddress(contract.CypressCreditContractAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(CypressCreditBin),
			Balance: big.NewInt(0),
		}
		alloc[common.HexToAddress(contract.AddressBookContractAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(CypressAddressBookBin),
			Balance: big.NewInt(0),
		}
		genesis.Alloc = alloc
	}
}

func AllocWithPrebaobabContract(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		alloc := makeGenesisAccount(addrs, balance)
		alloc[common.HexToAddress(contract.AddressBookContractAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(PrebaobabAddressBookBin),
			Balance: big.NewInt(0),
		}
		genesis.Alloc = alloc
	}
}

func AllocWithBaobabContract(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		alloc := makeGenesisAccount(addrs, balance)
		alloc[common.HexToAddress(contract.AddressBookContractAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(BaobabAddressBookBin),
			Balance: big.NewInt(0),
		}
		genesis.Alloc = alloc
	}
}

func ChainID(chainID *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.ChainID = chainID
	}
}

func UnitPrice(price uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.UnitPrice = price
	}
}

func Istanbul(config *params.IstanbulConfig) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Istanbul = config
	}
}

func DeriveShaImpl(impl int) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.DeriveShaImpl = impl
	}
}

func Governance(config *params.GovernanceConfig) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Governance = config
	}
}

func Clique(config *params.CliqueConfig) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Clique = config
	}
}

func StakingInterval(interval uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Governance.Reward.StakingUpdateInterval = interval
	}
}

func ProposerInterval(interval uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Governance.Reward.ProposerUpdateInterval = interval
	}
}
