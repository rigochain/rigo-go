package stake

import (
	"errors"
	"github.com/rigochain/rigo-go/types/xerrors"
	"sync"
)

type BlockMarker struct {
	BlockHeights []int64 `json:"blockHeights"`
	mtx          sync.RWMutex
}

func (bm *BlockMarker) Mark(height int64) xerrors.XError {
	bm.mtx.Lock()
	defer bm.mtx.Unlock()

	lastIdx := len(bm.BlockHeights) - 1
	if lastIdx >= 0 && bm.BlockHeights[lastIdx] >= height {
		return xerrors.From(errors.New("height must bigger than last marked height"))
	}

	bm.BlockHeights = append(bm.BlockHeights, height)
	return nil
}

func (bm *BlockMarker) CountInWindow(startHeight, endHeight int64, rewin bool) int {
	if startHeight > endHeight {
		return 0
	}

	bm.mtx.RLock()
	defer bm.mtx.RUnlock()

	count := 0
	preIdx := -1
	for i, h := range bm.BlockHeights {
		if h < startHeight {
			preIdx = i
		}
		if h >= startHeight && h <= endHeight {
			count++
		}
		if h >= endHeight {
			break
		}
	}

	if rewin && preIdx > 0 {
		bm.BlockHeights = bm.BlockHeights[preIdx+1:]
	}

	return count
}
