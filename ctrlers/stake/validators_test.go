package stake

import (
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/rigochain/rigo-go/libs"
	"github.com/rigochain/rigo-go/types/bytes"
	"github.com/rigochain/rigo-go/types/crypto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/rand"
	"sort"
	"testing"
)

func TestStakeCtrler_validatorUpdates(t *testing.T) {
	for c := 1; c < 100; c++ {
		testValidatorUpdates(t, c, 21)
	}
}

func testValidatorUpdates(t *testing.T, delegateeCnt, maxValCnt int) {
	var allDelegatees DelegateeArray
	maxCapPower := math.MaxInt64 / int64(delegateeCnt)
	for i := 0; i < delegateeCnt; i++ {
		prvKey := secp256k1.GenPrivKey()
		pubBytes := prvKey.PubKey().Bytes()
		addr, _ := crypto.PubBytes2Addr(pubBytes)
		d := NewDelegatee(addr, pubBytes)
		d.TotalPower = rand.Int63n(maxCapPower)
		allDelegatees = append(allDelegatees, d)
	}
	sort.Sort(PowerOrderDelegatees(allDelegatees))
	prePower := int64(0)
	for i, d := range allDelegatees {
		if i == 0 {
			prePower = d.TotalPower
		} else {
			require.True(t, prePower >= d.TotalPower)
		}
	}

	n := libs.MIN(maxValCnt, len(allDelegatees))
	newVals1 := selectValidators(allDelegatees, n)
	require.NotNil(t, newVals1)

	var updatedVals DelegateeArray
	sort.Sort(AddressOrderDelegatees(newVals1))
	valUps := validatorUpdates(nil, newVals1)

	for _, v := range valUps {
		pubBytes := v.PubKey.GetSecp256K1()
		addr, _ := crypto.PubBytes2Addr(pubBytes)
		d := NewDelegatee(addr, pubBytes)
		d.TotalPower = v.Power
		updatedVals = append(updatedVals, d)
	}
	require.Equal(t, newVals1, updatedVals)
	for i, _ := range newVals1 {
		require.Equal(t, newVals1[i].Addr, updatedVals[i].Addr)
		require.Equal(t, newVals1[i].TotalPower, updatedVals[i].TotalPower)
	}

	// update delegatees
	var newDelegatees DelegateeArray
	for _, d := range allDelegatees {
		choice := rand.Int() % 4
		if delegateeCnt <= 10 && choice == 1 {
			choice = 0 // un-staking partially
		}
		switch choice {
		case 0: // partially un-staking
			d.TotalPower -= rand.Int63n(d.TotalPower - 1)
		case 1: // un-staking all
			d.TotalPower = 0
		case 2: // staking
			d.TotalPower += rand.Int63n(math.MaxInt64 - d.TotalPower)
		case 3: // new staking
			newone := NewDelegatee(d.Addr, d.PubKey)
			d.TotalPower = rand.Int63n(math.MaxInt64 - 1000000)
			newDelegatees = append(newDelegatees, newone)
		}
	}
	allDelegatees = append(allDelegatees, newDelegatees...)
	sort.Sort(PowerOrderDelegatees(allDelegatees))
	prePower = int64(0)
	for i, d := range allDelegatees {
		if i == 0 {
			prePower = d.TotalPower
		} else {
			require.True(t, prePower >= d.TotalPower)
		}
	}

	n = libs.MIN(maxValCnt, len(allDelegatees))
	newVals2 := selectValidators(allDelegatees, n)
	require.NotNil(t, newVals2, "validator count", n, "allDelegatees", allDelegatees)

	sort.Sort(AddressOrderDelegatees(newVals2))
	valUps = validatorUpdates(updatedVals, newVals2)

	for _, v := range valUps {
		if v.Power == 0 {
			// remove
			found := false
			for i := 0; i < len(updatedVals); i++ {
				if bytes.Compare(v.PubKey.GetSecp256K1(), updatedVals[i].PubKey) == 0 {
					updatedVals = append(updatedVals[:i], updatedVals[i+1:]...)
					found = true
					break
				}
			}
			require.True(t, found)
		} else {
			// newer or updated
			found := false
			for i := 0; i < len(updatedVals); i++ {
				if bytes.Compare(v.PubKey.GetSecp256K1(), updatedVals[i].PubKey) == 0 {
					// updated
					updatedVals[i].TotalPower = v.Power
					found = true
					break
				}
			}
			if !found {
				// newer
				pubBytes := v.PubKey.GetSecp256K1()
				addr, _ := crypto.PubBytes2Addr(pubBytes)
				d := NewDelegatee(addr, pubBytes)
				d.TotalPower = v.Power
				updatedVals = append(updatedVals, d)
			}
		}
	}
	sort.Sort(AddressOrderDelegatees(updatedVals))

	require.Equal(t, newVals2, updatedVals)
	for i, _ := range newVals2 {
		require.Equal(t, newVals2[i].Addr, updatedVals[i].Addr)
		require.Equal(t, newVals2[i].TotalPower, updatedVals[i].TotalPower)
	}

	sort.Sort(PowerOrderDelegatees(allDelegatees))
	sort.Sort(PowerOrderDelegatees(updatedVals))
	for i, _ := range updatedVals {
		require.Equal(t, allDelegatees[i].Addr, updatedVals[i].Addr)
		require.Equal(t, allDelegatees[i].TotalPower, updatedVals[i].TotalPower)
	}
}
