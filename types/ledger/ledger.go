package ledger

import (
	"bytes"
	"github.com/cosmos/iavl"
	"github.com/kysee/arcanus/types/xerrors"
	db "github.com/tendermint/tm-db"
	"sync"
)

type ILedgerItem interface {
	GetKey() []byte
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type ILedger[T ILedgerItem] interface {
	Add(T)
	Find([]byte) (T, error)
	Iterate(func(T) error) error
	ResetMem() error
	Commit() ([]byte, int64, error)
}

type defaultTreeLedger[T ILedgerItem] struct {
	tree         *iavl.MutableTree
	cachedItems  []T
	addedItems   []T
	updatedItems []T
	removedItems []T
	mtx          sync.RWMutex
}

func DefaultTreeLedger[T ILedgerItem](name, dbDir string, cacheSize int) (*defaultTreeLedger[T], error) {
	if db, err := db.NewDB(name, "goleveldb", dbDir); err != nil {
		return nil, err
	} else if tree, err := iavl.NewMutableTree(db, cacheSize); err != nil {
		_ = db.Close()
		return nil, err
	} else if _, err := tree.Load(); err != nil {
		_ = db.Close()
		return nil, err
	} else {
		return &defaultTreeLedger[T]{
			tree: tree,
		}, nil
	}
}

func (l *defaultTreeLedger[T]) Add(itm T) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	l.addedItems = append(l.addedItems, itm)
}

func (l *defaultTreeLedger[T]) Find(key []byte) (T, error) {
	l.mtx.RLock()
	defer l.mtx.RUnlock()

	for _, v := range l.cachedItems {
		if bytes.Compare(key, v.GetKey()) == 0 {
			return v, nil
		}
	}

	var empty T

	if d, err := l.tree.Get(key); err != nil {
		return empty, err
	} else {
		var itm T
		if err := itm.Unmarshal(d); err != nil {
			return empty, err
		}
		l.cachedItems = append(l.cachedItems, itm)
		return itm, nil
	}
}

func (l *defaultTreeLedger[T]) Iterate(f func(T) error) error {
	stopped, err := l.tree.Iterate(func(key []byte, value []byte) bool {
		var itm T
		if err := itm.Unmarshal(value); err != nil {
			return true
		}
		if err := f(itm); err != nil {
			return true
		}
		return false
	})
	if err != nil {
		return xerrors.NewFrom(err)
	} else if stopped {
		return xerrors.New("Stop to load stakers tree")
	}
	return nil
}

func (l *defaultTreeLedger[T]) ResetMem() error {
	return nil
}

func (l *defaultTreeLedger[T]) Commit() ([]byte, int64, error) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	for _, itm := range l.addedItems {
		if d, err := itm.Marshal(); err != nil {
			return nil, 0, err
		} else if _, err := l.tree.Set(itm.GetKey(), d); err != nil {
			return nil, 0, err
		}
	}
	for _, itm := range l.updatedItems {
		if d, err := itm.Marshal(); err != nil {
			return nil, 0, err
		} else if _, err := l.tree.Set(itm.GetKey(), d); err != nil {
			return nil, 0, err
		}
	}
	for _, itm := range l.removedItems {
		if _, _, err := l.tree.Remove(itm.GetKey()); err != nil {
			return nil, 0, err
		}
	}

	return l.tree.SaveVersion()
}
