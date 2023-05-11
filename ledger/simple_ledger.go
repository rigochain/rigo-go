package ledger

import (
	"fmt"
	"github.com/cosmos/iavl"
	"github.com/rigochain/rigo-go/types/xerrors"
	tmdb "github.com/tendermint/tm-db"
	"sort"
	"sync"
)

type SimpleLedger[T ILedgerItem] struct {
	db          tmdb.DB
	tree        *iavl.MutableTree
	cachedItems *memItems[T]
	getNewItem  func() T

	mtx sync.RWMutex
}

func NewSimpleLedger[T ILedgerItem](name, dbDir string, cacheSize int, cb func() T) (*SimpleLedger[T], xerrors.XError) {
	if db, err := tmdb.NewDB(name, "goleveldb", dbDir); err != nil {
		return nil, xerrors.From(err)
	} else if tree, err := iavl.NewMutableTree(db, cacheSize); err != nil {
		_ = db.Close()
		return nil, xerrors.From(err)
	} else if _, err := tree.Load(); err != nil {
		_ = db.Close()
		return nil, xerrors.From(err)
	} else {
		return &SimpleLedger[T]{
			db:          db,
			tree:        tree,
			cachedItems: newMemItems[T](),
			getNewItem:  cb,
		}, nil
	}
}

func (ledger *SimpleLedger[T]) ImmutableLedgerAt(n int64, cacheSize int) (*SimpleLedger[T], xerrors.XError) {
	ledger.mtx.RLock()
	defer ledger.mtx.RUnlock()

	tree, err := iavl.NewMutableTree(ledger.db, cacheSize)
	if err != nil {
		return nil, xerrors.From(err)
	}

	_, err = tree.LazyLoadVersion(n)
	if err != nil {
		return nil, xerrors.From(err)
	}

	return &SimpleLedger[T]{
		tree:        tree,
		cachedItems: newMemItems[T](),
		getNewItem:  ledger.getNewItem,
	}, nil
}

func (ledger *SimpleLedger[T]) Set(item T) xerrors.XError {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	ledger.cachedItems.setUpdatedItem(item)
	ledger.cachedItems.setGotItem(item)
	return nil
}

func (ledger *SimpleLedger[T]) CancelSet(key LedgerKey) xerrors.XError {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	ledger.cachedItems.delUpdatedItem(key)
	ledger.cachedItems.delGotItem(key)
	return nil
}

func (ledger *SimpleLedger[T]) Get(key LedgerKey) (T, xerrors.XError) {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	return ledger.get(key)
}

func (ledger *SimpleLedger[T]) get(key LedgerKey) (T, xerrors.XError) {

	// search in cachedItems
	if item, ok := ledger.cachedItems.getGotItem(key); ok {
		return item, nil
	}

	var emptyNil T
	if item, xerr := ledger.read(key); xerr != nil {
		return emptyNil, xerr
	} else {
		ledger.cachedItems.setGotItem(item)
		return item, nil
	}
}

func (ledger *SimpleLedger[T]) Del(key LedgerKey) (T, xerrors.XError) {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	return ledger.del(key)
}

func (ledger *SimpleLedger[T]) del(key LedgerKey) (T, xerrors.XError) {
	var emptyNil T

	if item, err := ledger.get(key); err != nil {
		return emptyNil, err
	} else {
		ledger.cachedItems.delGotItem(key)       // delete(ledger.gotItems, key)
		ledger.cachedItems.delUpdatedItem(key)   // delete(ledger.updatedItems, key)
		ledger.cachedItems.appendRemovedKey(key) // ledger.removedKeys = append(ledger.removedKeys, key)
		return item, nil
	}
}

func (ledger *SimpleLedger[T]) CancelDel(key LedgerKey) xerrors.XError {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	ledger.cachedItems.delRemovedKey(key)
	return nil
}

func (ledger *SimpleLedger[T]) IterateAllItems(cb func(T) xerrors.XError) xerrors.XError {
	ledger.mtx.RLock()
	defer ledger.mtx.RUnlock()

	stopped, err := ledger.tree.Iterate(func(key []byte, value []byte) bool {
		item := ledger.getNewItem()
		if xerr := item.Decode(value); xerr != nil {
			panic(fmt.Errorf("unable to load stake - txhash:%X, error:%v", key, xerr))
			return true
		} else if item.Key() != ToLedgerKey(key) {
			panic(fmt.Errorf("wrong key - Key:%X vs. stake's txhash:%X", key, item.Key()))
			return true
		} else if xerr := cb(item); xerr != nil {
			return true
		}
		return false
	})

	if err != nil {
		return xerrors.From(err)
	} else if stopped {
		return xerrors.NewOrdinary("stop to iterate ledger tree")
	}
	return nil
}

func (ledger *SimpleLedger[T]) IterateGotItems(cb func(T) xerrors.XError) xerrors.XError {
	ledger.mtx.RLock()
	defer ledger.mtx.RUnlock()

	return iterateItems(ledger.cachedItems.gotItems, cb)
}

func (ledger *SimpleLedger[T]) IterateUpdatedItems(cb func(T) xerrors.XError) xerrors.XError {
	ledger.mtx.RLock()
	defer ledger.mtx.RUnlock()

	return iterateItems(ledger.cachedItems.updatedItems, cb)
}

func (ledger *SimpleLedger[T]) Read(key LedgerKey) (T, xerrors.XError) {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	var emptyNil T
	if item, xerr := ledger.read(key); xerr != nil {
		return emptyNil, xerr
	} else {
		// Do not call ledger.cachedItems.setGotItem(...)
		// Read() only reads a item from tree
		return item, nil
	}
}

func (ledger *SimpleLedger[T]) read(key LedgerKey) (T, xerrors.XError) {
	var emptyNil T
	item := ledger.getNewItem()

	if bz, err := ledger.tree.Get(key[:]); err != nil {
		return emptyNil, xerrors.From(err)
	} else if bz == nil {
		return emptyNil, xerrors.ErrNotFoundResult
	} else if err := item.Decode(bz); err != nil {
		return emptyNil, xerrors.From(err)
	} else if key != item.Key() {
		return emptyNil, xerrors.NewOrdinary("simple_ledger: the key is compromised - the requested key is not equal to the key encoded in value")
	} else {
		return item, nil
	}
}

func (ledger *SimpleLedger[T]) Commit() ([]byte, int64, xerrors.XError) {
	ledger.mtx.Lock()
	defer ledger.mtx.Unlock()

	var keys LedgerKeyList
	for k, _ := range ledger.cachedItems.updatedItems {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	for _, k := range keys {
		_val := ledger.cachedItems.updatedItems[k]
		_key := _val.Key()
		if bz, err := _val.Encode(); err != nil {
			return nil, -1, err
		} else if _, err := ledger.tree.Set(_key[:], bz); err != nil {
			return nil, -1, xerrors.From(err)
		}
	}

	for _, k := range ledger.cachedItems.removedKeys {
		var vk LedgerKey
		copy(vk[:], k[:])
		if _, _, err := ledger.tree.Remove(vk[:]); err != nil {
			return nil, -1, xerrors.From(err)
		}
	}

	if r1, r2, err := ledger.tree.SaveVersion(); err != nil {
		return r1, r2, xerrors.From(err)
	} else {
		ledger.cachedItems.reset()
		return r1, r2, nil
	}
}

func (ledger *SimpleLedger[T]) Clone() ILedger[T] {
	return &SimpleLedger[T]{
		tree:        ledger.tree,
		cachedItems: newMemItems[T](),
		getNewItem:  ledger.getNewItem,
	}
}

func (ledger *SimpleLedger[T]) Close() xerrors.XError {
	if ledger.db != nil {
		if err := ledger.db.Close(); err != nil {
			return xerrors.From(err)
		}
	}

	ledger.db = nil
	ledger.tree = nil
	return nil
}

var _ ILedger[ILedgerItem] = (*SimpleLedger[ILedgerItem])(nil)
