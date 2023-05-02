package genesis

import (
	tmtypes "github.com/tendermint/tendermint/types"
)

func DevnetGenesisDoc() (*tmtypes.GenesisDoc, error) {
	return tmtypes.GenesisDocFromJSON(jsonBlobDevnetGenesis)
}

var jsonBlobDevnetGenesis = []byte(`{
  "genesis_time": "2021-08-06T08:29:24.827484Z",
  "chain_id": "DEVNET",
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
      "address": "B9ADC66777A1900A8F5CF22F07E7641CC3C3CF48",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "A9uTzg7ST634ZtmfksDrQspEGt2n+GUsmk31X6P2yRjj"
      },
      "power": "10",
      "name": ""
    },
    {
      "address": "82F4C6D5498A2CA1E194E3C5AD9AD1EEEC9E7AF0",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "AroU/zplM+sz7oIYUUycXmZx4mruoplpXaRhpteoQpZa"
      },
      "power": "10",
      "name": ""
    },
    {
      "address": "51398BA5613C62D2566B523C6E49D94B88F55D54",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "AnhrTQzRUNs1VLRuc3WF5kOXJ6qLE0nUNGprYM1rvQF1"
      },
      "power": "10",
      "name": ""
    },
    {
      "address": "0EC62329BE52FDB338448C53DDB082A2E0AAF864",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "AmlK2xCcj6RbabvdwtfmK65Zty4lX9YPswpsoUqW2LHI"
      },
      "power": "10",
      "name": ""
    },
    {
      "address": "D8AEAC7E12BD6488036505262FE71767D3996792",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "AtjFfclYPh0/cC46/BeOJIuJWPsjhOg5js8azJyiaal2"
      },
      "power": "10",
      "name": ""
    }
  ],
  "app_hash": "0BB80C199AC01DA31D82B8DA95C941F91E796908A6546B96D0EF1BE55CED9E16",
  "app_state": {
    "assetHolders": [
      {
        "address": "B9ADC66777A1900A8F5CF22F07E7641CC3C3CF48",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "82F4C6D5498A2CA1E194E3C5AD9AD1EEEC9E7AF0",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "51398BA5613C62D2566B523C6E49D94B88F55D54",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "0EC62329BE52FDB338448C53DDB082A2E0AAF864",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "D8AEAC7E12BD6488036505262FE71767D3996792",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "22E94B8CD68867197BBEC78BD5F290E77EB0955E",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "99F954D4EA8DB0CFB7932404E004C7C5DE35977F",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "2E8F04F3F5CE9C8EB60586043D16F6D542539A47",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "CC2A6D5B73D438A282C6C811C5E6837DE40A3CB1",
        "balance": "0x52b7d2dcc80cd2e4000000"
      },
      {
        "address": "4979E92CFF082C6ADAE085CA51818B70D7754B21",
        "balance": "0x52b7d2dcc80cd2e4000000"
      }
    ], 
	"govRule": {
      "version": "0",
      "maxValidatorCnt": "21",
      "amountPerPower": "0xde0b6b3a7640000",
      "rewardPerPower": "0x3b9aca00",
      "lazyRewardBlocks": "10",
      "lazyApplyingBlocks": "10",
      "minTrxFee": "0xde0b6b3a7640000",
      "minVotingPeriodBlocks": "259200",
	  "maxVotingPeriodBlocks": "2592000"
	}
  }
}
`)
