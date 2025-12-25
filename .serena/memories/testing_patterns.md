# テストパターン

## テスト構成
```
internal/
  domain/seat/entity_test.go      # ユニットテスト（純粋Go）
  application/*_test.go           # シナリオテスト（モックリポジトリ）
  api/handler/*_test.go           # ハンドラテスト（モックサービス）
e2e/reservation_flow_test.go      # E2Eテスト（実DB/Redis）
```

## テストパターン

### 1. テーブル駆動テスト
```go
func TestSeat_Reserve(t *testing.T) {
    tests := []struct {
        name    string
        status  seat.Status
        wantErr error
    }{
        {"available seat", seat.StatusAvailable, nil},
        {"reserved seat", seat.StatusReserved, seat.ErrSeatNotAvailable},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := &seat.Seat{Status: tt.status}
            err := s.Reserve("res-123")
            assert.Equal(t, tt.wantErr, err)
        })
    }
}
```

### 2. モック（testify/mock）
```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*entity, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*entity), args.Error(1)
}

// テスト内
mockRepo.On("GetByID", ctx, "123").Return(entity, nil)
```

### 3. 並行性テスト
```go
func TestConcurrentReservation(t *testing.T) {
    var wg sync.WaitGroup
    successCount := int32(0)
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := service.Reserve(ctx, input)
            if err == nil {
                atomic.AddInt32(&successCount, 1)
            }
        }()
    }
    wg.Wait()
    assert.Equal(t, int32(1), successCount) // 1つだけ成功
}
```

## テスト実行コマンド
```bash
make test                                    # 全テスト
go test -v -run TestName ./path/to/package   # 単一テスト
make test-integration                        # 統合テスト
```

## アサーション
```go
import "github.com/stretchr/testify/assert"

assert.NoError(t, err)
assert.Equal(t, expected, actual)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.True(t, condition)
errors.Is(err, expectedErr)  // エラー比較
```
