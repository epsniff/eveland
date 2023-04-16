package dbmarketorders

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/epsniff/eveland/src/evesdk"
	"github.com/stretchr/testify/assert"
)

type MockEveLand struct {
	marketOrders []*evesdk.MarketOrder
}

func NewMockEveLand(marketOrders []*evesdk.MarketOrder) *MockEveLand {
	return &MockEveLand{
		marketOrders: marketOrders,
	}
}

func (m *MockEveLand) ListAllMarketOrdersForRegion(ctx context.Context, region *evesdk.Region) ([]*evesdk.MarketOrder, error) {
	return m.marketOrders, nil
}

func TestLoadMarketOrders(t *testing.T) {
	// Prepare mock data
	mockMarketOrders := []*evesdk.MarketOrder{
		{OrderID: 1, Price: 2.0, SystemID: 30000142, TypeID: 42, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: true},
		{OrderID: 2, Price: 66.0, SystemID: 30000142, TypeID: 24, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: true},
		{OrderID: 3, Price: 22.0, SystemID: 30000142, TypeID: 42, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: false},
		//
		{OrderID: 4, Price: 10.0, SystemID: 30000123, TypeID: 34, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: true},
		{OrderID: 5, Price: 10.0, SystemID: 30000123, TypeID: 34, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: false},
		{OrderID: 6, Price: 10.0, SystemID: 30000123, TypeID: 34, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: true},
		{OrderID: 7, Price: 10.0, SystemID: 30000123, TypeID: 34, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: false},
		{OrderID: 8, Price: 10.0, SystemID: 30000123, TypeID: 34, VolumeRemain: 100, VolumeTotal: 100, Issued: time.Now(), Duration: 90, IsBuyOrder: true},
	}

	mockEveLand := NewMockEveLand(mockMarketOrders)
	n, err := createRandomTempSubdir()
	if err != nil {
		t.Fatal("error creating random temp dir: ", err)
	}
	dbm, err := New(mockEveLand, n)
	if err != nil {
		t.Fatal("error creating new OrderDataDB: ", err)
	}

	// Prepare test data
	region := &evesdk.Region{
		RegionID: 10000002,
		Name:     "The Forge",
	}

	// Run the test function
	cnt, err := dbm.LoadMarketOrders(context.Background(), region)
	// Check if there are no errors
	assert.NoError(t, err)
	assert.Equal(t, 8, cnt)

	// Check if the data was loaded correctly
	bos, sos, err := dbm.GetMarketOrdersBySystemID(context.TODO(), 30000142)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bos))
	assert.Equal(t, 1, len(sos))

	// bos[42].Pop()
}

func createRandomTempSubdir() (string, error) {
	baseTempDir := os.TempDir()

	// Generate a random name for the folder
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random folder name: %v", err)
	}
	randomFolderName := hex.EncodeToString(randomBytes)

	// Create the random folder
	randomDirPath := filepath.Join(baseTempDir, randomFolderName)
	err = os.MkdirAll(randomDirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create random folder: %v", err)
	}

	return randomDirPath, nil
}
