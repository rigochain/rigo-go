package genesis

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

func MainnetGenesisDoc(chainId string) (*tmtypes.GenesisDoc, error) {
	genDoc, err := tmtypes.GenesisDocFromJSON(jsonBlobMainnetGenesis)
	if err != nil {
		return nil, err
	}
	return genDoc, nil
}

var jsonBlobMainnetGenesis = []byte(`{
  "genesis_time": "2023-09-20T06:45:01.333168Z",
  "chain_id": "mainnet",
  "initial_height": "0",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1",
      "time_iota_ms": "1000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": [
        "secp256k1"
      ]
    },
    "version": {
      "app_version": "1"
    }
  },
  "validators": [
    {
      "address": "EE273B53C7D654CBF0D660D6781D7FC15BAED707",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "A+wvscISumm42piANW2CJYH/qiU0bx2okyPabIkIf9aS"
      },
      "power": "10000000",
      "name": "Validator#1"
    },
    {
      "address": "1D459DCCC6A7AE76496044E3FB886BFB0FB714D4",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "A/lvsF0qTmZL7CLHrbzj5YvmFlGe8Tv9SuPGcz3F3lkS"
      },
      "power": "10000000",
      "name": "Validator#2"
    },
    {
      "address": "9E57D69287D5EEF9E8F03206B5B53FC6AA222CB9",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "AvyfnTQ8kAWtXlPU+9vKss1IjkR23pNp/bjN4BeTvRsz"
      },
      "power": "10000000",
      "name": "Validator#3"
    },
    {
      "address": "0F9894E0262002254F300BFEC2B2087096C6586E",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "Ah+5oHoVb1nODcAoAKQXiJO0mcs4gknXB3QIjHsP8dDF"
      },
      "power": "10000000",
      "name": "Validator#4"
    },
    {
      "address": "EFAC306C301C7B110DF0631F3B74F4C1AC93EEBF",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "A3CSHEWwsZgk0lfF/Qhw/1Lbu3gHlpDRzCZ/59zYwpe3"
      },
      "power": "10000000",
      "name": "Validator#5"
    }
  ],
  "app_hash": "9965D70B33A5B4EDEB50BFC295DE1DEB83AE7A210C0F5FFF560F3331A1766B07",
  "app_state": {
    "assetHolders": [
      {
        "address": "64621D5A507D566451E3947341B6AED9B500B40E",
        "balance": "923800000000000000000000000"
      },
      {
        "address": "2EFE6619BB6B784025D839C76410D089B43B4BB6",
        "balance": "6000000000000000000000000"
      },
      {
        "address": "9600F2786A663E4D679D470555CF1FB871836888",
        "balance": "200000000000000000000000"
      },
      {
        "address": "46937D479A555F5CEE42E47346AEF4C36DB15EC3",
        "balance": "20000000000000000000000000"
      }
    ],
    "govParams": {
      "version": "1",
      "maxValidatorCnt": "21",
      "minValidatorStake": "7000000000000000000000000",
      "rewardPerPower": "4756468797",
      "lazyRewardBlocks": "2592000",
      "lazyApplyingBlocks": "259200",
      "gasPrice": "250000000000",
      "minTrxGas": "4000",
      "maxTrxGas": "25000000",
      "maxBlockGas": "18446744073709551615",
      "minVotingPeriodBlocks": "259200",
      "maxVotingPeriodBlocks": "2592000",
      "minSelfStakeRatio": "50",
      "maxUpdatableStakeRatio": "30",
      "slashRatio": "50",
      "signedBlocksWindow": "10000",
      "minSignedBlocks": "500"
    }
  }
}
`)
