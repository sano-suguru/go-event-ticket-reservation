# Go製イベントチケット予約システム - プロジェクト企画書

## 1. 背景と課題

### 現在の状況
- **職務経歴**: 決済額1,000億円超のふるさと納税サイト「ふるなび」で6年間の開発経験
- **主要技術スタック**: C#/.NET、TypeScript、AWS、SQL Server
- **強み**: 決済システム開発、大規模トラフィック対応、高信頼性システム設計

### 課題
- **バックエンド.NETは日本の求人市場ではニッチ**
  - モダンなWeb開発の主流はGo、Rust、Kotlin、Pythonなど
  - 特にスタートアップやモダンな企業ではGoの採用が増加
  
- **経歴との技術スタックのギャップを埋める必要性**
  - .NETでの豊富な経験があるが、Go言語での実績がない
  - 言語は違えど、設計力・アーキテクチャ理解は転用可能であることを証明したい

### 目標
Goで実装したポートフォリオアプリを通じて、以下をアピールする：
1. **技術の転用可能性**: .NETで培った設計力がGoでも発揮できる
2. **本質的な理解**: 言語に依存しない、システム設計の本質的理解
3. **実務レベルの実装力**: Todoアプリではなく、実務で求められる複雑性を持つアプリ

## 2. なぜ「イベントチケット予約システム」なのか

### 2.1 Todoアプリが不適切な理由

| 観点             | Todoアプリ | あなたの経歴             |
| ---------------- | ---------- | ------------------------ |
| ユーザー数       | 基本単一   | 200-300万人規模          |
| 競合制御         | 不要       | 必須（在庫、決済）       |
| トランザクション | 単純       | 複雑（分散、補償）       |
| 整合性要件       | 低い       | 高い（金銭）             |
| パフォーマンス   | 重要度低   | 年末トラフィック急増対応 |

→ **経歴と比較して「芸がない」、差別化できない**

### 2.2 イベントチケット予約システムの適切性

#### あなたの経験との対応関係

| あなたの経験                                                                           | チケット予約システムでの対応                                                           |
| -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| **決済システム開発**<br>- 在庫引当<br>- トランザクション管理<br>- 二重予約防止         | **チケット在庫管理**<br>- 座席在庫引当<br>- 予約トランザクション<br>- 二重予約防止     |
| **カート機能開発**<br>- 複数商品の状態管理<br>- セッション管理<br>- 在庫確保           | **複数チケット予約**<br>- 複数座席の状態管理<br>- 予約有効期限管理<br>- 座席確保・解放 |
| **大規模トラフィック対応**<br>- 年末の集中アクセス<br>- ECS AutoScaling<br>- Redis活用 | **人気イベント販売開始時**<br>- 同時アクセス制御<br>- 在庫キャッシュ<br>- 分散ロック   |
| **非同期処理設計**<br>- Kinesis + Lambda<br>- イベント駆動<br>- データレプリケーション | **予約確定の非同期処理**<br>- キュー処理<br>- 遅延確定<br>- 通知処理                   |
| **監視・運用**<br>- New Relic<br>- CloudWatch<br>- X-Ray                               | **観測性の実装**<br>- 構造化ログ<br>- メトリクス<br>- トレーシング                     |

#### 技術的な複雑性のバランス

```
複雑度の軸：

[低] Todo → [中] チケット予約 → [高] 本格的なEC基盤
              ↑
              実装可能性と
              技術アピールの
              最適バランス
```

**理由:**
- ✅ **2-4週間で実装可能**な規模
- ✅ **競合制御、トランザクション、非同期処理**など実務的な複雑性
- ✅ **パフォーマンス測定**が意味を持つ
- ✅ **経歴との一貫性**が説明しやすい

## 3. 技術選定とその理由

### 3.1 言語: Go

**選定理由:**
1. **求人市場での需要**
   - モダンなWebバックエンドのデファクトスタンダードの一つ
   - マイクロサービス、API開発での採用例多数
   
2. **.NETとの類似性**
   - 静的型付け言語（C#からの移行が容易）
   - 標準ライブラリが充実
   - パフォーマンス重視の設計思想
   
3. **並行処理のサポート**
   - goroutineによる軽量な並行処理
   - あなたの非同期処理の経験が活かせる

### 3.2 フレームワーク: Echo

```go
// Echo推奨理由
- より高機能、ミドルウェア充実
- コミュニティ大きい
- エンタープライズグレードのプロジェクトで採用例多数
推奨度: ★★★★★
```

### 3.3 データベース: PostgreSQL

**選定理由:**
1. **ACID特性の完全サポート**
   - あなたの「高信頼性・高整合性システム」の経験を活かせる
   - トランザクション分離レベルの制御
   
2. **経歴との整合性**
   - SQL Server経験が直接活かせる
   - SQLの知識が転用可能
   
3. **Go言語との親和性**
   - pgx（高速なPostgreSQLドライバ）が成熟
   - database/sqlインターフェースが標準

### 3.4 キャッシュ/ロック: Redis

**選定理由:**
1. **経歴での使用実績**
   - ElastiCacheの経験
   - セッション管理での使用経験
   
2. **分散ロックの実装**
   - 在庫競合制御に必須
   - Redisのアトミック操作を活用
   
3. **パフォーマンス向上**
   - 在庫数のキャッシュ
   - レスポンス時間の改善

### 3.5 その他のツール選定

| 用途             | 選定技術                | 理由                     |
| ---------------- | ----------------------- | ------------------------ |
| ORマッパー       | sqlx                    | 軽量、Dapper経験に近い   |
| マイグレーション | golang-migrate          | 標準的、シンプル         |
| バリデーション   | go-playground/validator | デファクトスタンダード   |
| ロギング         | zap                     | 構造化ログ、高速         |
| テスト           | testify                 | アサーション、モック     |
| 負荷テスト       | k6                      | GoベースでJSシナリオ記述 |
| コンテナ         | Docker Compose          | ローカル開発環境         |
| CI               | GitHub Actions          | 経験と同じ               |
| API仕様          | OpenAPI 3.0             | Swagger UI自動生成       |

## 4. システム設計

### 4.1 アーキテクチャ概要

```
┌─────────────────────────────────────────────────┐
│              API Layer (Echo)                    │
│  - ルーティング                                    │
│  - ミドルウェア（認証、ログ、エラーハンドリング）        │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│           Application Layer                      │
│  - ユースケース実装                                 │
│  - トランザクション境界                              │
│  - ビジネスロジック                                 │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│            Domain Layer                          │
│  - エンティティ                                    │
│  - 値オブジェクト                                   │
│  - ドメインサービス                                 │
│  - リポジトリインターフェース                         │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│         Infrastructure Layer                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │PostgreSQL│  │  Redis   │  │  Logger  │      │
│  │Repository│  │  Cache   │  │          │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└─────────────────────────────────────────────────┘
```

**アーキテクチャパターン: Clean Architecture**

**理由:**
- 依存関係の方向が明確（内側に向かう単方向）
- テスタビリティが高い
- あなたの「拡張性・保守性の高いシステム設計」の経験を示せる

### 4.2 ドメインモデル

```go
// 主要なエンティティ

type Event struct {
    ID          string
    Name        string
    Description string
    Venue       string
    StartAt     time.Time
    EndAt       time.Time
    TotalSeats  int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Seat struct {
    ID          string
    EventID     string
    SeatNumber  string
    Status      SeatStatus // Available, Reserved, Confirmed
    Price       int
}

type Reservation struct {
    ID              string
    EventID         string
    UserID          string
    SeatIDs         []string
    Status          ReservationStatus // Pending, Confirmed, Cancelled
    ExpiresAt       time.Time
    ConfirmedAt     *time.Time
    IdempotencyKey  string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 4.3 データベース設計

```sql
-- events テーブル
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    venue VARCHAR(255),
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    total_seats INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 0 -- 楽観的ロック用
);

-- seats テーブル
CREATE TABLE seats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    seat_number VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL, -- available, reserved, confirmed
    price INTEGER NOT NULL,
    reserved_by UUID, -- reservation_id
    reserved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 0, -- 楽観的ロック用
    UNIQUE(event_id, seat_number)
);

-- reservations テーブル
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, confirmed, cancelled
    idempotency_key VARCHAR(255) UNIQUE, -- 冪等性保証
    expires_at TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP,
    total_amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- reservation_seats テーブル（中間テーブル）
CREATE TABLE reservation_seats (
    reservation_id UUID NOT NULL REFERENCES reservations(id),
    seat_id UUID NOT NULL REFERENCES seats(id),
    PRIMARY KEY (reservation_id, seat_id)
);

-- インデックス
CREATE INDEX idx_seats_event_status ON seats(event_id, status);
CREATE INDEX idx_reservations_user ON reservations(user_id);
CREATE INDEX idx_reservations_expires ON reservations(expires_at) WHERE status = 'pending';
CREATE INDEX idx_reservations_idempotency ON reservations(idempotency_key);
```

### 4.4 API設計

```
POST   /api/v1/events                    # イベント作成
GET    /api/v1/events                    # イベント一覧
GET    /api/v1/events/:id                # イベント詳細
GET    /api/v1/events/:id/seats          # 座席一覧・空席状況

POST   /api/v1/reservations              # 予約作成（座席確保）
GET    /api/v1/reservations/:id          # 予約詳細
POST   /api/v1/reservations/:id/confirm  # 予約確定
POST   /api/v1/reservations/:id/cancel   # 予約キャンセル
GET    /api/v1/users/:id/reservations    # ユーザーの予約一覧

GET    /api/v1/health                    # ヘルスチェック
GET    /api/v1/metrics                   # メトリクス（Prometheus形式）
GET    /swagger/*                        # Swagger UI
```

## 5. 核心的な技術実装

### 5.1 在庫競合制御: 楽観的ロック

```go
// あなたの決済システムでの経験を活かす

func (r *SeatRepository) ReserveSeats(ctx context.Context, tx *sqlx.Tx, 
    seatIDs []string, reservationID string) error {
    
    query := `
        UPDATE seats
        SET 
            status = 'reserved',
            reserved_by = $1,
            reserved_at = NOW(),
            version = version + 1
        WHERE 
            id = ANY($2)
            AND status = 'available'
            AND version = $3  -- 楽観的ロックチェック
        RETURNING id
    `
    
    var updatedIDs []string
    err := tx.SelectContext(ctx, &updatedIDs, query, 
        reservationID, pq.Array(seatIDs), currentVersion)
    
    if len(updatedIDs) != len(seatIDs) {
        return ErrSeatAlreadyReserved // 競合発生
    }
    
    return err
}
```

**アピールポイント:**
- あなたの決済システムでの「データ整合性を保つ設計」の経験
- 楽観的ロックによる高いパフォーマンスと整合性の両立

### 5.2 分散ロック（Redis）

```go
// 人気イベントの座席争奪戦でのロック制御

type RedisLock struct {
    client *redis.Client
}

func (l *RedisLock) AcquireLock(ctx context.Context, 
    eventID string, ttl time.Duration) (bool, error) {
    
    lockKey := fmt.Sprintf("lock:event:%s", eventID)
    
    // Redisのアトミック操作で分散ロック獲得
    success, err := l.client.SetNX(ctx, lockKey, "locked", ttl).Result()
    if err != nil {
        return false, err
    }
    
    return success, nil
}

func (l *RedisLock) ReleaseLock(ctx context.Context, eventID string) error {
    lockKey := fmt.Sprintf("lock:event:%s", eventID)
    return l.client.Del(ctx, lockKey).Err()
}
```

**アピールポイント:**
- あなたのElastiCache/Redis使用経験
- 分散システムにおける排他制御の理解

### 5.3 冪等性の保証

```go
// あなたの決済システムでの「二重請求防止」の経験を活かす

func (s *ReservationService) CreateReservation(ctx context.Context, 
    req *CreateReservationRequest) (*Reservation, error) {
    
    // 冪等性キーのチェック
    existing, err := s.repo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
    if err != nil && !errors.Is(err, ErrNotFound) {
        return nil, err
    }
    
    // 既に同じリクエストで作成済み
    if existing != nil {
        return existing, nil
    }
    
    // 新規予約作成
    return s.createNewReservation(ctx, req)
}
```

**アピールポイント:**
- あなたの「決済時のネットワーク障害による二重請求防止」の経験
- 分散システムにおける冪等性設計の理解

### 5.4 トランザクション境界の設計

```go
// あなたの「トランザクション管理」の経験を活かす

func (s *ReservationService) CreateReservation(ctx context.Context, 
    req *CreateReservationRequest) (*Reservation, error) {
    
    // トランザクション開始
    tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()
    
    // 1. 予約レコード作成
    reservation, err := s.repo.Create(ctx, tx, req)
    if err != nil {
        return nil, err
    }
    
    // 2. 座席を予約状態に更新（楽観的ロック）
    err = s.seatRepo.ReserveSeats(ctx, tx, req.SeatIDs, reservation.ID)
    if err != nil {
        return nil, err
    }
    
    // 3. キャッシュ更新
    err = s.cache.UpdateAvailableSeats(ctx, req.EventID)
    if err != nil {
        // キャッシュ更新失敗はログして継続
        log.Error("cache update failed", "error", err)
    }
    
    // コミット
    if err := tx.Commit(); err != nil {
        return nil, err
    }
    
    // 4. 非同期で通知送信（トランザクション外）
    s.notifier.SendReservationCreated(reservation)
    
    return reservation, nil
}
```

**アピールポイント:**
- トランザクション境界の適切な設計
- どこをトランザクション内に含めるか/外に出すかの判断
- あなたの「決済の高速なレスポンスとデータ整合性を両立」の経験

### 5.5 リトライ機構

```go
// あなたの「自動リトライ機能」の経験を活かす

type RetryConfig struct {
    MaxRetries   int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func WithRetry(ctx context.Context, config RetryConfig, 
    fn func() error) error {
    
    var lastErr error
    delay := config.InitialDelay
    
    for attempt := 0; attempt <= config.MaxRetries; attempt++ {
        if attempt > 0 {
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return ctx.Err()
            }
            
            // Exponential backoff
            delay = time.Duration(float64(delay) * config.Multiplier)
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        }
        
        if err := fn(); err != nil {
            lastErr = err
            
            // リトライ可能なエラーかチェック
            if !isRetryable(err) {
                return err
            }
            
            continue
        }
        
        return nil
    }
    
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryable(err error) bool {
    // タイムアウト、一時的なネットワークエラーなど
    var netErr net.Error
    if errors.As(err, &netErr) && netErr.Temporary() {
        return true
    }
    
    // デッドロック検出
    if errors.Is(err, sql.ErrTxDone) {
        return true
    }
    
    return false
}
```

**アピールポイント:**
- あなたのPollyライブラリ使用経験
- リトライロジックの適切な実装

### 5.6 キャッシュ戦略

```go
// あなたの「DB負荷軽減のためのキャッシュ活用」の経験を活かす

type SeatCache struct {
    redis *redis.Client
}

func (c *SeatCache) GetAvailableSeats(ctx context.Context, 
    eventID string) (int, error) {
    
    cacheKey := fmt.Sprintf("seats:available:%s", eventID)
    
    // キャッシュヒット
    val, err := c.redis.Get(ctx, cacheKey).Int()
    if err == nil {
        return val, nil
    }
    
    // キャッシュミス: DBから取得
    count, err := c.countAvailableSeatsFromDB(ctx, eventID)
    if err != nil {
        return 0, err
    }
    
    // キャッシュに保存（短いTTL）
    c.redis.Set(ctx, cacheKey, count, 30*time.Second)
    
    return count, nil
}

func (c *SeatCache) InvalidateEvent(ctx context.Context, eventID string) error {
    cacheKey := fmt.Sprintf("seats:available:%s", eventID)
    return c.redis.Del(ctx, cacheKey).Err()
}
```

**アピールポイント:**
- あなたのRedis/ElastiCache使用経験
- キャッシュ invalidation戦略

## 6. 実装フェーズ

### Phase 1: 基盤実装（1週間）

**目標: MVPの動作**

```
Day 1-2: プロジェクトセットアップ
□ Go modules初期化
□ ディレクトリ構造作成
□ Docker Compose（PostgreSQL, Redis）
□ マイグレーションツールセットアップ
□ 基本的なCI/CD（lint, test）

Day 3-4: コア機能実装
□ イベントCRUD
□ 座席管理
□ 基本的なリポジトリパターン

Day 5-7: 予約機能
□ 予約作成API
□ 楽観的ロック実装
□ トランザクション管理
□ 基本的な単体テスト
```

**マイルストーン:**
- ✅ 予約作成〜確定の一連のフローが動作
- ✅ 競合制御が機能（簡易版）

### Phase 2: 信頼性向上（1週間）

**目標: プロダクショングレードの品質**

```
Day 8-9: 分散ロック
□ Redis分散ロック実装
□ タイムアウト処理
□ デッドロック回避

Day 10-11: 冪等性・リトライ
□ 冪等性キー実装
□ リトライ機構
□ エラーハンドリング強化

Day 12-14: キャッシュ・最適化
□ Redis キャッシュ実装
□ クエリ最適化
□ インデックスチューニング
```

**マイルストーン:**
- ✅ 二重予約が確実に防止される
- ✅ ネットワーク障害時のリトライが動作
- ✅ キャッシュによるパフォーマンス改善

### Phase 3: 観測性・パフォーマンス（1週間）

**目標: 運用を見据えた実装**

```
Day 15-16: ロギング・メトリクス
□ 構造化ログ（zap）
□ メトリクス収集（Prometheus形式）
□ OpenAPI仕様書生成

Day 17-18: 負荷テスト
□ k6でシナリオ作成
□ ベンチマーク実行
□ ボトルネック特定

Day 19-21: 改善・ドキュメント
□ パフォーマンスチューニング
□ README作成
□ アーキテクチャ図作成
□ 技術的な意思決定の文書化
```

**マイルストーン:**
- ✅ 負荷テスト結果をREADMEに記載
- ✅ ログから問題追跡が可能
- ✅ パフォーマンスのボトルネックが特定・改善済み

## 7. テスト戦略

### 7.1 テストピラミッド

```
        ┌──────────┐
        │   E2E    │  少数・クリティカルパスのみ
        └──────────┘
       ┌────────────┐
       │ Integration│  中程度・リポジトリ層
       └────────────┘
      ┌──────────────┐
      │  Unit Tests  │  多数・ビジネスロジック
      └──────────────┘
```

### 7.2 具体的なテスト例

```go
// ユニットテスト: ビジネスロジック

func TestReservationService_CreateReservation_ConcurrentAccess(t *testing.T) {
    // 同時に10人が同じ座席を予約しようとする
    // 1人だけ成功することを確認
    
    var wg sync.WaitGroup
    results := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, err := service.CreateReservation(ctx, req)
            results <- err
        }()
    }
    
    wg.Wait()
    close(results)
    
    successCount := 0
    conflictCount := 0
    
    for err := range results {
        if err == nil {
            successCount++
        } else if errors.Is(err, ErrSeatAlreadyReserved) {
            conflictCount++
        }
    }
    
    assert.Equal(t, 1, successCount)
    assert.Equal(t, 9, conflictCount)
}

// 統合テスト: リポジトリ

func TestSeatRepository_ReserveSeats_OptimisticLock(t *testing.T) {
    // テスト用DB使用
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewSeatRepository(db)
    
    // 座席作成
    seat := createTestSeat(t, db, "A1")
    
    // トランザクション1で取得
    tx1, _ := db.Begin()
    seat1, _ := repo.GetByID(ctx, tx1, seat.ID)
    
    // トランザクション2で取得・更新
    tx2, _ := db.Begin()
    seat2, _ := repo.GetByID(ctx, tx2, seat.ID)
    repo.ReserveSeats(ctx, tx2, []string{seat2.ID}, "reservation-2")
    tx2.Commit()
    
    // トランザクション1で更新試行 → 楽観的ロック失敗
    err := repo.ReserveSeats(ctx, tx1, []string{seat1.ID}, "reservation-1")
    
    assert.Error(t, err)
    assert.ErrorIs(t, err, ErrOptimisticLockConflict)
}
```

### 7.3 負荷テスト

```javascript
// k6による負荷テストシナリオ

import http from 'k6/http';
import { check, sleep } from 'k6';

// シナリオ: 人気イベントの発売開始
export let options = {
    stages: [
        { duration: '30s', target: 100 },  // 100並行ユーザーまで増加
        { duration: '1m', target: 100 },   // 100並行ユーザーを維持
        { duration: '30s', target: 0 },    // クールダウン
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95%ile 500ms以下
        http_req_failed: ['rate<0.01'],   // エラー率1%未満
    },
};

export default function () {
    // イベント一覧取得
    let eventRes = http.get('http://localhost:8080/api/v1/events');
    check(eventRes, { 'events loaded': (r) => r.status === 200 });
    
    let events = JSON.parse(eventRes.body);
    if (events.length === 0) return;
    
    let eventID = events[0].id;
    
    // 座席一覧取得
    let seatsRes = http.get(`http://localhost:8080/api/v1/events/${eventID}/seats`);
    check(seatsRes, { 'seats loaded': (r) => r.status === 200 });
    
    let seats = JSON.parse(seatsRes.body);
    let availableSeats = seats.filter(s => s.status === 'available');
    
    if (availableSeats.length === 0) return;
    
    // 予約試行
    let payload = JSON.stringify({
        event_id: eventID,
        seat_ids: [availableSeats[0].id],
        idempotency_key: `key-${__VU}-${__ITER}`,
    });
    
    let reserveRes = http.post(
        'http://localhost:8080/api/v1/reservations',
        payload,
        { headers: { 'Content-Type': 'application/json' } }
    );
    
    // 成功 or 競合は両方OK（在庫切れは正常動作）
    check(reserveRes, { 
        'reservation created or conflict': (r) => 
            r.status === 201 || r.status === 409 
    });
    
    sleep(1);
}
```

## 8. READMEでのアピール構成

### 構成案

```markdown
# Event Ticket Reservation System

## 概要
イベントチケット予約システムのバックエンドAPI実装。
大規模ECサイトでの決済システム開発経験を活かし、
高信頼性・高整合性が求められる予約システムを実装。

## 技術的な特徴

### 1. 競合制御と整合性の確保
- 楽観的ロックによる在庫競合制御
- Redisを用いた分散ロック
- トランザクション境界の適切な設計

### 2. 高いパフォーマンス
- Redisキャッシュによる高速なレスポンス
- クエリ最適化とインデックス設計
- 負荷テスト結果: XXX req/sec達成

### 3. 信頼性の確保
- 冪等性の保証（二重予約防止）
- リトライ機構とエラーハンドリング
- 構造化ログによる追跡可能性

### 4. Clean Architecture
- ドメイン駆動設計
- テスタビリティの高い設計
- 依存関係の明確化

## アーキテクチャ
（図を挿入）

## 技術選定の理由
（各技術の選定理由を記載）

## 実装のハイライト

### 楽観的ロックによる在庫制御
（コードスニペット + 説明）

### 分散ロック実装
（コードスニペット + 説明）

### トランザクション設計
（コードスニペット + 説明）

## パフォーマンス

### 負荷テスト結果
- 並行ユーザー数: 100
- スループット: XXX req/sec
- 95パーセンタイルレスポンス: XXX ms
- エラー率: 0.X%

### ボトルネック分析と改善
1. 問題: XXX
   改善: XXX
   結果: XX%向上

## セットアップ

### 前提条件
- Docker & Docker Compose
- Go 1.21+
- make（オプション）

### 起動方法
```bash
# 依存サービス起動
docker-compose up -d

# マイグレーション実行
make migrate-up

# アプリケーション起動
make run

# Swagger UI
http://localhost:8080/swagger/index.html
```

## API仕様
OpenAPI 3.0仕様書をSwagger UIで確認できます。

## 今後の拡張案
- [ ] 決済統合
- [ ] 待機列システム
- [ ] イベント推薦機能

## 開発者
決済額1,000億円超のふるさと納税サイトにて、
決済システム開発、大規模トラフィック対応を担当。
本プロジェクトでは、その経験をGoで実装。

職務経歴書: [リンク]
```

## 9. 期待されるアピール効果

### 採用担当者への訴求ポイント

1. **技術の転用可能性**
   - 「.NETでの経験を他言語でも発揮できる」
   - 言語固有の知識ではなく、本質的な設計力を持つ

2. **実務レベルの実装力**
   - Todoアプリではない、実務的な複雑性
   - 競合制御、トランザクション、非同期処理など

3. **経歴との一貫性**
   - 決済システム ↔ 予約システムの類似性
   - 大規模トラフィック対応の経験
   - 高信頼性システムの設計

4. **運用を見据えた実装**
   - ログ、メトリクス、監視
   - パフォーマンステスト
   - ドキュメント化

### 面接での説明例

```
面接官: 「.NETの経験が豊富ですが、Goは初めてですか？」

あなた: 「はい、Goでの業務経験はまだありませんが、
ポートフォリオとしてイベントチケット予約システムを実装しました。

これは、私が担当したふるさと納税サイトの決済システムと
技術的な課題が類似しており、座席の在庫管理と決済の在庫管理、
予約の競合制御と決済の二重実行防止など、本質的には同じ問題です。

実装では、楽観的ロックによる競合制御、Redisでの分散ロック、
冪等性の保証、リトライ機構など、実務で必要な要素を
すべて盛り込んでいます。

また、負荷テストを実施し、100並行ユーザーで
XXX req/secのスループットを達成しました。

これらを通じて、言語は異なっても、高信頼性システムの
設計・実装能力は転用可能であることを示せたと考えています。」
```

## 10. フロントエンドの方針

### 推奨：**作らない（APIのみ）**

**理由:**

✅ **経歴に既に十分な実績がある**
- Vue.js 3、React、状態管理など豊富な経験
- カート機能の複雑なフロント実装済み
- これ以上フロントをアピールする必要性は低い

✅ **バックエンドエンジニアとしてアピールしたい**
- 志望職種がバックエンド寄り
- フロントに時間を使うより、バックエンドの質を高める方が効果的

✅ **時間を有効活用できる**
- 2-4週間という限られた期間
- フロントに1週間使うなら、負荷テストやパフォーマンス改善に注力

### 代替案：API仕様書の充実

```yaml
tools:
  - swag（Go用OpenAPI生成ツール）
  - Swagger UI（自動生成されるインタラクティブなドキュメント）
  - Postman Collection（APIテスト用）
  
提供するもの:
  - Swagger UIでブラウザから直接APIをテスト可能
  - リクエスト/レスポンスのサンプル
  - エラーコードの詳細説明
  - 認証フローの説明
```

**面接での説明:**
> 「フロントエンドは職務経歴に十分な実績があるため、今回はバックエンドの設計・実装の質を最大化することに集中しました。API仕様書はOpenAPIで完全に文書化し、Swagger UIで誰でも試せるようにしています」

## 11. インフラ構成

### 推奨：**Docker Compose + IaCコード**

**構成:**

```yaml
実装レベル:
  [必須] Docker Compose環境
    - PostgreSQL
    - Redis
    - アプリケーション
    - すべてローカルで完結
    - README通りにコマンド実行すれば動く
  
  [推奨] AWS CDKコード
    - 本番想定のインフラコード
    - 実際にデプロイはしない（コストの問題）
    - 「こう構築する」という設計を示す
  
  [オプション] 無料環境へのデプロイ
    - Railway（PostgreSQL、Redis、Webアプリ無料枠）
    - Render（同様）
    - 実際に動くデモを提供
```

### 11.1 Docker Compose（必須）

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: ticket_reservation
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://postgres:postgres@postgres:5432/ticket_reservation?sslmode=disable
      REDIS_URL: redis://redis:6379
      LOG_LEVEL: debug

volumes:
  postgres_data:
```

### 11.2 AWS CDKコード（推奨）

```typescript
// infrastructure/lib/ticket-reservation-stack.ts
import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as rds from 'aws-cdk-lib/aws-rds';
import * as elasticache from 'aws-cdk-lib/aws-elasticache';
import * as logs from 'aws-cdk-lib/aws-logs';

export class TicketReservationStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // VPC
    const vpc = new ec2.Vpc(this, 'VPC', {
      maxAzs: 2,
      natGateways: 1
    });

    // RDS PostgreSQL（Multi-AZ）
    const db = new rds.DatabaseInstance(this, 'Database', {
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.VER_15
      }),
      vpc,
      instanceType: ec2.InstanceType.of(
        ec2.InstanceClass.T3,
        ec2.InstanceSize.SMALL
      ),
      multiAz: true, // 高可用性
      allocatedStorage: 20,
      maxAllocatedStorage: 100, // ストレージの自動スケーリング
      backupRetention: cdk.Duration.days(7),
      deletionProtection: true
    });

    // ElastiCache Redis（レプリケーション）
    const redisSubnetGroup = new elasticache.CfnSubnetGroup(this, 'RedisSubnetGroup', {
      description: 'Subnet group for Redis',
      subnetIds: vpc.privateSubnets.map(subnet => subnet.subnetId),
    });

    const redis = new elasticache.CfnReplicationGroup(this, 'Redis', {
      replicationGroupDescription: 'Redis for caching and distributed locks',
      engine: 'redis',
      cacheNodeType: 'cache.t3.micro',
      numCacheClusters: 2, // プライマリ + レプリカ
      automaticFailoverEnabled: true,
      cacheSubnetGroupName: redisSubnetGroup.ref,
    });

    // ECS Cluster
    const cluster = new ecs.Cluster(this, 'Cluster', {
      vpc,
      containerInsights: true
    });

    // Fargate Task Definition
    const taskDef = new ecs.FargateTaskDefinition(this, 'TaskDef', {
      cpu: 512,
      memoryLimitMiB: 1024,
    });

    const container = taskDef.addContainer('app', {
      image: ecs.ContainerImage.fromRegistry('your-app:latest'),
      logging: ecs.LogDrivers.awsLogs({
        streamPrefix: 'ticket-reservation',
        logRetention: logs.RetentionDays.ONE_WEEK
      }),
      environment: {
        DATABASE_URL: `postgres://...`, // Secrets Managerから取得
        REDIS_URL: `redis://...`
      },
    });

    container.addPortMappings({ containerPort: 8080 });

    // Fargate Service（オートスケーリング）
    const service = new ecs.FargateService(this, 'Service', {
      cluster,
      taskDefinition: taskDef,
      desiredCount: 2, // 最小2台
      minHealthyPercent: 50,
      maxHealthyPercent: 200,
    });

    // オートスケーリング設定
    const scaling = service.autoScaleTaskCount({
      minCapacity: 2,
      maxCapacity: 10
    });

    scaling.scaleOnCpuUtilization('CpuScaling', {
      targetUtilizationPercent: 70
    });

    // Application Load Balancer
    const alb = new elbv2.ApplicationLoadBalancer(this, 'ALB', {
      vpc,
      internetFacing: true
    });

    const listener = alb.addListener('Listener', {
      port: 80,
      open: true
    });

    listener.addTargets('ECS', {
      port: 8080,
      targets: [service],
      healthCheck: {
        path: '/api/v1/health',
        interval: cdk.Duration.seconds(30)
      }
    });

    // CloudWatch Alarms
    // ... エラー率、レスポンスタイム、CPU使用率などのアラーム設定
  }
}
```

**READMEでの記載例:**

```markdown
## インフラ構成

### ローカル開発環境
Docker Composeで完結。以下のコマンドで起動：
```bash
docker-compose up -d
```

### 本番想定アーキテクチャ
AWSでの本番環境を想定したCDKコードを `infrastructure/` に配置。

**主要コンポーネント:**
- **ECS Fargate**: オートスケーリング対応（2-10台）
- **RDS PostgreSQL**: Multi-AZ構成、自動バックアップ
- **ElastiCache Redis**: レプリケーション構成
- **ALB**: ヘルスチェック、SSL終端
- **CloudWatch**: ログ集約、メトリクス、アラーム

実際のデプロイは行っていませんが、コードで設計意図を示しています。
推定月額コスト: 約$150-200（本番運用時）
```

### 11.3 無料環境へのデプロイ（オプション）

**Railway を使用した例:**

```yaml
# railway.toml
[build]
builder = "NIXPACKS"

[deploy]
startCommand = "./bin/api"
healthcheckPath = "/api/v1/health"
restartPolicyType = "ON_FAILURE"

[[services]]
name = "app"

[[services]]
name = "postgres"
image = "postgres:15"

[[services]]
name = "redis"
image = "redis:7-alpine"
```

**メリット:**
- 実際に動くデモを提供できる
- 面接官がブラウザで試せる
- Swagger UIも公開可能

**デメリット:**
- 無料枠の制限
- パフォーマンステストには不向き

## 12. 動かすために必要な追加要素

### 12.1 認証・認可（簡易版）

```go
// internal/api/middleware/auth.go

type AuthMiddleware struct {
    validAPIKeys map[string]string // API Key -> User ID
}

func NewAuthMiddleware() *AuthMiddleware {
    return &AuthMiddleware{
        validAPIKeys: map[string]string{
            "dev-admin-key-12345": "admin-user-id",
            "dev-user-key-67890":  "test-user-id",
        },
    }
}

func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        apiKey := c.Request().Header.Get("X-API-Key")
        
        if apiKey == "" {
            return echo.NewHTTPError(http.StatusUnauthorized, "API key required")
        }
        
        userID, ok := m.validAPIKeys[apiKey]
        if !ok {
            return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
        }
        
        // ユーザーIDをコンテキストに設定
        c.Set("user_id", userID)
        
        return next(c)
    }
}
```

**アピールポイント:**
- 「本格的な認証はスコープ外だが、APIキーで基本的な認証を実装」
- 「実務では OAuth2.0/JWTを使用する想定」

### 12.2 シードデータ

```go
// cmd/seed/main.go

package main

import (
    "log"
    "time"
    
    "github.com/yourusername/ticket-reservation/internal/infrastructure/database"
)

func main() {
    db := database.Connect()
    defer db.Close()
    
    log.Println("Creating seed data...")
    
    // イベント作成
    events := []struct {
        Name        string
        Venue       string
        StartAt     time.Time
        TotalSeats  int
    }{
        {
            Name:       "Rock Festival 2024",
            Venue:      "Tokyo Dome",
            StartAt:    time.Now().Add(30 * 24 * time.Hour),
            TotalSeats: 100,
        },
        {
            Name:       "Jazz Night",
            Venue:      "Blue Note Tokyo",
            StartAt:    time.Now().Add(45 * 24 * time.Hour),
            TotalSeats: 50,
        },
        {
            Name:       "Tech Conference 2024",
            Venue:      "Tokyo Big Sight",
            StartAt:    time.Now().Add(60 * 24 * time.Hour),
            TotalSeats: 200,
        },
    }
    
    for _, e := range events {
        eventID := createEvent(db, e)
        createSeats(db, eventID, e.TotalSeats)
    }
    
    log.Println("Seed data created successfully")
}

func createEvent(db *sql.DB, event struct{...}) string {
    // イベント作成ロジック
    // ...
    return eventID
}

func createSeats(db *sql.DB, eventID string, count int) {
    // 座席作成ロジック
    // A1, A2, ..., Z99 のような座席番号を生成
    // ...
}
```

### 12.3 環境変数管理

```bash
# .env.example

# ========================================
# Database Configuration
# ========================================
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ticket_reservation?sslmode=disable

# ========================================
# Redis Configuration
# ========================================
REDIS_URL=redis://localhost:6379

# ========================================
# Application Configuration
# ========================================
PORT=8080
LOG_LEVEL=debug
ENVIRONMENT=development

# ========================================
# API Keys (for development only)
# ========================================
ADMIN_API_KEY=dev-admin-key-12345
USER_API_KEY=dev-user-key-67890

# ========================================
# Performance Settings
# ========================================
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
REDIS_POOL_SIZE=10

# ========================================
# Business Logic
# ========================================
RESERVATION_EXPIRY_MINUTES=15
```

```go
// internal/config/config.go

type Config struct {
    DatabaseURL string
    RedisURL    string
    Port        string
    LogLevel    string
    Environment string
    
    AdminAPIKey string
    UserAPIKey  string
    
    DBMaxOpenConns int
    DBMaxIdleConns int
    RedisPoolSize  int
    
    ReservationExpiryMinutes int
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }
    
    return &Config{
        DatabaseURL:              getEnv("DATABASE_URL", ""),
        RedisURL:                 getEnv("REDIS_URL", "redis://localhost:6379"),
        Port:                     getEnv("PORT", "8080"),
        LogLevel:                 getEnv("LOG_LEVEL", "info"),
        Environment:              getEnv("ENVIRONMENT", "development"),
        AdminAPIKey:              getEnv("ADMIN_API_KEY", ""),
        UserAPIKey:               getEnv("USER_API_KEY", ""),
        DBMaxOpenConns:           getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
        DBMaxIdleConns:           getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
        RedisPoolSize:            getEnvAsInt("REDIS_POOL_SIZE", 10),
        ReservationExpiryMinutes: getEnvAsInt("RESERVATION_EXPIRY_MINUTES", 15),
    }, nil
}
```

### 12.4 Makefile（開発効率化）

```makefile
.DEFAULT_GOAL := help

# ========================================
# Variables
# ========================================
APP_NAME := ticket-reservation
DOCKER_IMAGE := $(APP_NAME):latest
DATABASE_URL := postgres://postgres:postgres@localhost:5432/ticket_reservation?sslmode=disable

# ========================================
# Help
# ========================================
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ========================================
# Setup & Installation
# ========================================
.PHONY: setup
setup: ## Initial setup (install dependencies, start services, migrate, seed)
	@echo "Setting up project..."
	go mod download
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	sleep 5
	make migrate-up
	make seed
	@echo "Setup complete! Run 'make run' to start the application"

.PHONY: install-tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ========================================
# Development
# ========================================
.PHONY: run
run: ## Run application locally
	go run cmd/api/main.go

.PHONY: build
build: ## Build application binary
	go build -o bin/api cmd/api/main.go

.PHONY: watch
watch: ## Run with hot reload (requires air)
	air

# ========================================
# Database
# ========================================
.PHONY: migrate-up
migrate-up: ## Run database migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" down

.PHONY: migrate-create
migrate-create: ## Create new migration file (usage: make migrate-create name=add_users_table)
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: seed
seed: ## Seed database with test data
	go run cmd/seed/main.go

.PHONY: reset-db
reset-db: ## Reset database (drop, migrate, seed)
	make migrate-down
	make migrate-up
	make seed

# ========================================
# Testing
# ========================================
.PHONY: test
test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	go test -v -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
	go test -v -run Integration ./...

.PHONY: test-coverage
test-coverage: test ## Show test coverage in browser
	go tool cover -html=coverage.out

.PHONY: load-test
load-test: ## Run k6 load test
	k6 run tests/load/reservation_test.js

# ========================================
# Code Quality
# ========================================
.PHONY: lint
lint: ## Run linter
	golangci-lint run --timeout=5m

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# ========================================
# Documentation
# ========================================
.PHONY: swagger
swagger: ## Generate Swagger documentation
	swag init -g cmd/api/main.go -o docs

# ========================================
# Docker
# ========================================
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all services
	docker-compose down

.PHONY: docker-logs
docker-logs: ## View Docker logs
	docker-compose logs -f

.PHONY: docker-clean
docker-clean: ## Remove all containers and volumes
	docker-compose down -v

# ========================================
# Cleanup
# ========================================
.PHONY: clean
clean: ## Clean up build artifacts and caches
	rm -f bin/api
	rm -f coverage.out
	rm -rf docs/
	go clean -cache

.PHONY: clean-all
clean-all: clean docker-clean ## Clean everything including Docker volumes
	@echo "All clean!"
```

### 12.5 GitHub Actions（CI/CD）

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        cache: true
    
    - name: Install dependencies
      run: go mod download
    
    - name: Install migrate tool
      run: |
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
        sudo mv migrate /usr/local/bin/
    
    - name: Run migrations
      run: migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable" up
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable
        REDIS_URL: redis://localhost:6379
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        cache: true
    
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        cache: true
    
    - name: Build
      run: go build -v -o bin/api cmd/api/main.go
    
    - name: Upload artifact
      uses: actions/upload-artifact@v3
      with:
        name: api-binary
        path: bin/api
```

### 12.6 詳細なREADME

```markdown
# Event Ticket Reservation System

[![CI](https://github.com/yourusername/ticket-reservation/workflows/CI/badge.svg)](https://github.com/yourusername/ticket-reservation/actions)
[![codecov](https://codecov.io/gh/yourusername/ticket-reservation/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/ticket-reservation)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/ticket-reservation)](https://goreportcard.com/report/github.com/yourusername/ticket-reservation)

イベントチケット予約システムのバックエンドAPI実装。大規模ECサイトでの決済システム開発経験を活かし、高信頼性・高整合性が求められる予約システムを実装。

## 📋 目次

- [クイックスタート](#クイックスタート)
- [技術的な特徴](#技術的な特徴)
- [アーキテクチャ](#アーキテクチャ)
- [API仕様](#api仕様)
- [開発](#開発)
- [テスト](#テスト)
- [パフォーマンス](#パフォーマンス)
- [デプロイ](#デプロイ)

## 🚀 クイックスタート

### 前提条件

- Docker & Docker Compose
- Go 1.21+
- make（オプション）

### セットアップ

```bash
# 1. リポジトリクローン
git clone https://github.com/yourusername/ticket-reservation.git
cd ticket-reservation

# 2. 環境変数設定
cp .env.example .env

# 3. セットアップ（Docker起動、マイグレーション、シードデータ投入）
make setup

# 4. アプリケーション起動
make run
```

### 動作確認

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

### APIの試し方

```bash
# イベント一覧取得
curl http://localhost:8080/api/v1/events

# 座席確認
curl http://localhost:8080/api/v1/events/{event_id}/seats

# 予約作成（認証が必要）
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-user-key-67890" \
  -d '{
    "event_id": "...",
    "seat_ids": ["..."],
    "idempotency_key": "unique-key-123"
  }'
```

## 💡 技術的な特徴

### 1. 競合制御と整合性の確保

- **楽観的ロック**による在庫競合制御
- **Redisを用いた分散ロック**
- トランザクション境界の適切な設計

### 2. 高いパフォーマンス

- Redisキャッシュによる高速なレスポンス
- クエリ最適化とインデックス設計
- 負荷テスト結果: 100並行ユーザーで XXX req/sec達成

### 3. 信頼性の確保

- 冪等性の保証（二重予約防止）
- リトライ機構とエラーハンドリング
- 構造化ログによる追跡可能性

### 4. Clean Architecture

- ドメイン駆動設計
- テスタビリティの高い設計
- 依存関係の明確化

## 🏗️ アーキテクチャ

詳細は [docs/architecture.md](docs/architecture.md) を参照。

## 📚 API仕様

OpenAPI 3.0仕様書をSwagger UIで確認できます。

http://localhost:8080/swagger/index.html

## 🛠️ 開発

### 便利なコマンド

```bash
make help          # すべてのコマンドを表示
make run           # アプリケーション起動
make test          # テスト実行
make lint          # リンター実行
make load-test     # 負荷テスト実行
make swagger       # API仕様書生成
```

### ディレクトリ構造

```
.
├── cmd/
│   ├── api/           # メインアプリケーション
│   └── seed/          # シードデータ投入ツール
├── internal/
│   ├── domain/        # ドメイン層
│   ├── application/   # アプリケーション層
│   ├── infrastructure/# インフラ層
│   └── api/           # API層
├── db/
│   └── migrations/    # DBマイグレーション
├── tests/
│   ├── unit/          # 単体テスト
│   ├── integration/   # 統合テスト
│   └── load/          # 負荷テスト
├── docs/              # ドキュメント
└── infrastructure/    # IaCコード（AWS CDK）
```

## 🧪 テスト

### テストの実行

```bash
# すべてのテスト
make test

# 単体テストのみ
make test-unit

# 統合テストのみ
make test-integration

# カバレッジ確認
make test-coverage

# 負荷テスト
make load-test
```

## ⚡ パフォーマンス

### 負荷テスト結果

- **並行ユーザー数**: 100
- **スループット**: XXX req/sec
- **95パーセンタイルレスポンス**: XXX ms
- **エラー率**: 0.X%

詳細は [docs/performance.md](docs/performance.md) を参照。

## 🚢 デプロイ

### ローカル環境（Docker Compose）

```bash
docker-compose up -d
```

### 本番想定（AWS）

AWS CDKによるインフラコードを `infrastructure/` に配置。
実際のデプロイは行っていませんが、以下の構成を想定:

- **ECS Fargate**: オートスケーリング対応（2-10台）
- **RDS PostgreSQL**: Multi-AZ構成
- **ElastiCache Redis**: レプリケーション構成
- **ALB**: ヘルスチェック、SSL終端

詳細は [infrastructure/README.md](infrastructure/README.md) を参照。

## 📖 技術ブログ

実装の詳細について、以下の記事で解説しています:

- [楽観的ロックによる在庫競合制御の実装](link)
- [Redisを使った分散ロックの実装](link)
- [トランザクション設計のベストプラクティス](link)

## 👤 開発者

決済額1,000億円超のふるさと納税サイトにて、決済システム開発、大規模トラフィック対応を担当。
本プロジェクトでは、その経験をGoで実装。

- **職務経歴書**: [リンク]
- **GitHub**: [@yourusername](https://github.com/yourusername)
- **Qiita**: [@Nossa](https://qiita.com/Nossa)

## 📝 ライセンス

MIT License
```

## 13. 実装スコープ：最小限〜推奨構成

### 🟢 Phase 1: 最小限（動かすために必須） - Week 1

```
必須要素:
□ Go実装（バックエンドAPI）
  □ Clean Architecture構造
  □ イベントCRUD
  □ 座席管理
  □ 予約作成・確定・キャンセル
  □ 楽観的ロック実装
  
□ データベース
  □ PostgreSQL（Docker Compose）
  □ マイグレーション
  □ 基本的なインデックス
  
□ キャッシュ
  □ Redis（Docker Compose）
  □ 基本的なキャッシュ実装
  
□ 開発環境
  □ Docker Compose設定
  □ .env設定
  □ シードデータ
  
□ ドキュメント
  □ 基本的なREADME（セットアップ手順）
  
□ テスト
  □ 主要な単体テスト
```

**マイルストーン:**
- ✅ `docker-compose up` で環境が立ち上がる
- ✅ READMEの手順通りに動作する
- ✅ 予約フローが一通り動作する

### 🟡 Phase 2: 推奨（アピール力を高める） - Week 2-3

```
推奨要素:
□ 信頼性向上
  □ 分散ロック（Redis）
  □ 冪等性キー
  □ リトライ機構
  □ エラーハンドリング強化
  
□ パフォーマンス
  □ クエリ最適化
  □ インデックスチューニング
  □ キャッシュ戦略の洗練
  
□ 観測性
  □ 構造化ログ（zap）
  □ メトリクス（Prometheus形式）
  □ OpenAPI仕様書（Swagger UI）
  
□ テスト
  □ 統合テスト
  □ テーブル駆動テスト
  □ 負荷テスト（k6）
  
□ CI/CD
  □ GitHub Actions
  □ 自動テスト実行
  □ Lint自動化
  
□ 開発効率化
  □ Makefile
  □ Pre-commit hooks
  
□ ドキュメント
  □ アーキテクチャ図
  □ 技術的な意思決定の文書化
  □ パフォーマンステスト結果
```

**マイルストーン:**
- ✅ 二重予約が確実に防止される
- ✅ 負荷テスト結果が測定・文書化されている
- ✅ CI/CDが動作し、プルリクエストで自動テストが走る

### 🔵 Phase 3: オプション（さらに差別化） - Week 4

```
オプション要素:
□ インフラ
  □ AWS CDKコード（実際のデプロイなし）
  □ Railway等への実デプロイ
  □ Terraformコード（代替）
  
□ 高度な機能
  □ 予約有効期限の自動キャンセル（バックグラウンドワーカー）
  □ 待機列システム（簡易版）
  □ WebSocket通知
  
□ ドキュメント
  □ 技術ブログ記事執筆
  □ Qiita記事公開
  □ アーキテクチャ決定記録（ADR）
  
□ その他
  □ Docker Hubへのイメージ公開
  □ CodecovでカバレッジバッジGitHubバッジの充実
  □ Postman Collectionの提供
```

## 14. 次のステップ

### Week 1: Phase 1実装（必須）
- **Day 1-2**: プロジェクトセットアップ、基本構造
- **Day 3-4**: イベント・座席管理実装
- **Day 5-7**: 予約機能実装、楽観的ロック

**デイリーゴール:**
- 毎日GitHubにコミット
- 動作する状態を維持
- テストを書きながら進める

### Week 2: Phase 2実装（推奨）
- **Day 8-10**: 信頼性向上（分散ロック、冪等性、リトライ）
- **Day 11-14**: 観測性、テスト、CI/CD、ドキュメント

**週末ゴール:**
- 負荷テストを実行し結果を記録
- READMEを充実させる

### Week 3: Phase 2完成 + Phase 3着手
- **Day 15-18**: パフォーマンスチューニング、ドキュメント充実
- **Day 19-21**: AWS CDKコード or 技術ブログ執筆

**週末ゴール:**
- ポートフォリオとして完成
- 第三者がREADME見て動かせる状態

### Week 4: ブラッシュアップ（時間次第）
- コードレビュー（自己レビュー）
- リファクタリング
- デモ環境構築
- 技術ブログ公開

---

## 15. 最終チェックリスト

転職活動に使う前に、以下を確認：

### 必須項目
- [ ] READMEの手順通りに、第三者が環境構築できる
- [ ] `make setup` で完全に立ち上がる
- [ ] Swagger UIが動作し、APIを試せる
- [ ] テストが通る（`make test`）
- [ ] CIが通る（GitHub Actions）
- [ ] 基本的なエラーハンドリングが実装されている

### 推奨項目
- [ ] 負荷テスト結果がREADMEに記載されている
- [ ] アーキテクチャ図がある
- [ ] 技術的な意思決定が文書化されている
- [ ] コードカバレッジが60%以上
- [ ] Lintエラーが0件

### オプション項目
- [ ] 実際にデプロイされたデモ環境がある
- [ ] 技術ブログ記事がある
- [ ] GitHubバッジが充実している
