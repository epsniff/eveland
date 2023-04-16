package dbitems

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/epsniff/eveland/src/evesdk"
)

type EveLand interface {
	ListAllTypeIDs(ctx context.Context) ([]int32, error)
	GetTypeData(ctx context.Context, typeID int32) (*evesdk.TypeData, error)
}

type ItemDataDB struct {
	eveSDK EveLand

	mu        sync.RWMutex
	typeCache map[int32]*evesdk.TypeData

	pdb *pebble.DB
}

func New(eveSDK EveLand, dbpath string) (*ItemDataDB, error) {
	pebDbPath, err := db_location(dbpath)
	if err != nil {
		return nil, fmt.Errorf("error preping db location: %v", err)
	}
	fmt.Println("Storing item data on disk in pebbledb at: ", pebDbPath)

	pdb, err := pebble.Open(pebDbPath, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("error opening: %v", err)
	}

	return &ItemDataDB{eveSDK: eveSDK, pdb: pdb, typeCache: make(map[int32]*evesdk.TypeData)}, nil
}

func (r *ItemDataDB) Close() error {
	err := r.pdb.Close()
	if err != nil {
		return fmt.Errorf("error closing: %v", err)
	}
	return nil
}

func (r *ItemDataDB) GetItem(ctx context.Context, id int32) (*evesdk.TypeData, error) {
	// Check the cache first.
	r.mu.RLock()
	if td, ok := r.typeCache[id]; ok {
		r.mu.RUnlock()
		return td, nil
	}
	r.mu.RUnlock()

	key := TypeIDKey(id)
	data, closer, err := r.pdb.Get([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("error while trying to read from db: err: %v", err)
	}
	defer closer.Close()

	// Unmarshal the data.
	var td *evesdk.TypeData
	err = json.Unmarshal(data, &td)
	if err != nil {
		return nil, fmt.Errorf("error while trying to unmarshal data: err: %v", err)
	}

	// Add the item to the cache.
	r.mu.Lock()
	r.typeCache[id] = td
	r.mu.Unlock()

	return td, nil
}

func (r *ItemDataDB) LoadItems(ctx context.Context) error {
	types, err := r.eveSDK.ListAllTypeIDs(context.Background())
	if err != nil {
		fmt.Println("Error while trying to list all type ids:", err)
	}
	fmt.Printf("Number of types: %v\n", len(types))

	// sem is a channel that will allow up to 4 concurrent operations.
	var sem = make(chan int, 2)
	// Create a mutex to protect the marketOrders slice	append.
	typeMu := &sync.RWMutex{}
	typeDataList := []*evesdk.TypeData{}
	var wg sync.WaitGroup
	wg.Add(len(types))
	for i, t := range types {
		go func(i int, typeId int32) {
			defer func() {
				wg.Done()
				<-sem
			}()
			try := 0
		retry:
			td, err := r.eveSDK.GetTypeData(context.Background(), typeId)
			if err != nil {
				err := fmt.Errorf("error while trying to get type data: %v", err)
				fmt.Printf("Error: try: %v, type: %v, err: %v\n", try, typeId, err)
				try++
				time.Sleep(time.Duration(try) * time.Second)
				if try < 5 {
					goto retry
				}
			}
			typeMu.Lock()
			typeDataList = append(typeDataList, td)
			typeMu.Unlock()
			if i%300 == 0 {
				fmt.Printf("Processed %d types.\n", i)
			}
		}(i, t)
		time.Sleep(10 * time.Millisecond)
	}
	wg.Wait() // Wait for all goroutines to finish.
	fmt.Printf("Number of type data: %v\n", len(typeDataList))

	for _, td := range typeDataList {
		// marshal the data as json
		data, err := json.Marshal(td)
		if err != nil {
			return fmt.Errorf("error while trying to marshal type data: %v", err)
		}
		err = r.pdb.Set(TypeIDKey(td.TypeId), data, pebble.Sync)
		if err != nil {
			return fmt.Errorf("error while trying to write to db: err: %v", err)
		}
	}

	fmt.Printf("Done writing to type database, number of types: %v\n", len(typeDataList))

	return nil
}

func TypeIDKey(typeID int32) []byte {
	return []byte(strconv.Itoa(int(typeID)))
}

func db_location(baseDir string) (string, error) {
	dbpath := filepath.Join(baseDir, "eveitems_peb_db")

	_, err := os.Stat(dbpath)
	if os.IsNotExist(err) {
		err := os.Mkdir(dbpath, 0700)
		if err != nil {
			return "", fmt.Errorf("could not create directory %s: %w", dbpath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("could not stat directory %s: %w", dbpath, err)
	}

	return dbpath, nil
}
