package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rigochain/rigo-go/types/bytes"
	rigo_crypto "github.com/rigochain/rigo-go/types/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	prvKeyHex       = "83b8749ffd3b90bb26bdfa430f8df21d881df9962eb96b4ee68b3f60c57c5ccb"
	expectedBTCAddr = "7612536BD0991DB67E60DA9ECA1E3E276889B8DC"
)

func TestEcRecover(t *testing.T) {
	// create and check signature
	prvKey, err := rigo_crypto.ImportPrvKeyHex(prvKeyHex)
	require.NoError(t, err)

	pubKey := prvKey.PublicKey

	randBytes := bytes.RandBytes(1024)
	sig, err := rigo_crypto.Sign(randBytes, prvKey)
	require.NoError(t, err)
	require.True(t, rigo_crypto.VerifySig(rigo_crypto.CompressPubkey(&pubKey), randBytes, sig))

	addr0, _, err := rigo_crypto.Sig2Addr(randBytes, sig)
	require.NoError(t, err)
	require.Equal(t, expectedBTCAddr, addr0.String())

	// test for rigo_ecrecover
	ecr_input := make([]byte, 128)
	copy(ecr_input, rigo_crypto.DefaultHash(randBytes))
	ecr_input[63] = sig[64]
	copy(ecr_input[64:], sig)

	ecr := &rigo_ecrecover{}
	addr1, err := ecr.Run(ecr_input)
	require.NoError(t, err)
	require.Equal(t, common.LeftPadBytes(addr0, 32), addr1)
	require.Equal(t, expectedBTCAddr, bytes.HexBytes(common.TrimLeftZeroes(addr1)).String())
}
