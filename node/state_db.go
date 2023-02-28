package node

import (
	"encoding/binary"
	"github.com/rigochain/rigo-go/ctrlers/types"
	"github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
	"sync"
)

const (
	keyBlockHeight  = "bh"
	keyBlockAppHash = "ah"
	keyBlockGasInfo = "bg"
)

type StateDB struct {
	db tmdb.DB

	mtx   sync.RWMutex
	cache map[string][]byte
}

func openStateDB(name, dir string) (*StateDB, error) {
	// The returned 'db' instance is safe in concurrent use.
	db, err := tmdb.NewDB(name, "goleveldb", dir)
	if err != nil {
		return nil, err
	}

	return &StateDB{
		db:    db,
		cache: make(map[string][]byte),
	}, nil
}

func (stdb *StateDB) Close() error {
	stdb.mtx.Lock()
	defer stdb.mtx.Unlock()

	stdb.cache = map[string][]byte{}
	return stdb.db.Close()
}

func (stdb *StateDB) LastBlockHeight() int64 {
	v := stdb.get(keyBlockHeight)
	if v == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(v))
}

func (stdb *StateDB) PutLastBlockHeight(bh int64) error {
	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, uint64(bh))
	return stdb.put(keyBlockHeight, v)
}

func (stdb *StateDB) LastBlockAppHash() []byte {
	return stdb.get(keyBlockAppHash)
}

func (stdb *StateDB) PutLastBlockAppHash(v []byte) error {
	return stdb.put(keyBlockAppHash, v)
}

func (stdb *StateDB) LastBlockGasInfo() *types.BlockContext {
	bz := stdb.get(keyBlockGasInfo)
	if bz == nil {
		return nil
	}
	ret := &types.BlockContext{}
	if err := json.Unmarshal(bz, ret); err != nil {
		return nil
	}
	return ret
}

func (stdb *StateDB) PutLastBlockContext(ctx *types.BlockContext) error {
	bz, err := json.Marshal(ctx)
	if err != nil {
		return err
	}
	return stdb.put(keyBlockGasInfo, bz)
}

func (stdb *StateDB) putCache(k string, v []byte) {
	stdb.mtx.Lock()
	defer stdb.mtx.Unlock()

	stdb.cache[k] = v
}

func (stdb *StateDB) getCache(k string) []byte {
	stdb.mtx.RLock()
	defer stdb.mtx.RUnlock()

	v := stdb.cache[k]
	return v
}

func (stdb *StateDB) get(k string) []byte {
	if v := stdb.getCache(k); v != nil {
		return v
	}

	if v, err := stdb.db.Get([]byte(k)); err == nil {
		stdb.putCache(k, v)
		return v
	}

	return nil
}

func (stdb *StateDB) put(k string, v []byte) error {
	if err := stdb.db.SetSync([]byte(k), v); err != nil {
		return err
	}
	stdb.putCache(k, v)
	return nil
}
