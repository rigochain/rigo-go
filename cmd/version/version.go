package version

import (
	"fmt"
	"github.com/tendermint/tendermint/version"
	"strconv"
)

const (
	FMT_VERSTR = "%v.%v.%v-%x@%s"
)

var (
	majorVer  uint64 = 0
	minorVer  uint64 = 0
	patchVer  uint64 = 1
	commitVer uint64 = 0

	// it is changed using ldflags.
	//  ex) -ldflags "... -X 'github.com/rigochain/rigo-go/tynode/version.GitCommit=$(LVER)'"
	GitCommit string
)

func init() {
	if GitCommit != "" {
		commitVer, _ = strconv.ParseUint(GitCommit, 16, 64)
	}
}

func String() string {
	return fmt.Sprintf(FMT_VERSTR, majorVer, minorVer, patchVer, commitVer, version.TMCoreSemVer)
}

func Uint64() uint64 {
	return (majorVer << 56) + (minorVer << 48) + (patchVer << 32) + commitVer
}

func Parse(c uint64) string {
	return fmt.Sprintf(FMT_VERSTR,
		((c >> 56) & 0xFF),
		((c >> 48) & 0xFF),
		((c >> 32) & 0xFFFF),
		(c & 0xFFFFFFFF),
		version.TMCoreSemVer)
}

func Major() uint64 {
	return majorVer
}

func Minor() uint64 {
	return minorVer
}

func Patch() uint64 {
	return patchVer
}

func CommitHash() uint64 {
	return commitVer
}
