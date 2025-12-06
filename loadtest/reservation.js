import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// カスタムメトリクス
const reservationSuccess = new Counter('reservation_success');
const reservationConflict = new Counter('reservation_conflict');
const reservationError = new Counter('reservation_error');
const reservationDuration = new Trend('reservation_duration_ms');

// 設定
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';

// テストシナリオ
export const options = {
  scenarios: {
    // シナリオ1: 通常フロー（段階的負荷）
    normal_flow: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },  // 30秒で10VUまで増加
        { duration: '1m', target: 10 },   // 1分間10VU維持
        { duration: '30s', target: 0 },   // 30秒で0VUまで減少
      ],
      gracefulRampDown: '10s',
      exec: 'normalFlow',
    },
    // シナリオ2: 同時予約テスト（瞬間的な高負荷）
    concurrent_reservation: {
      executor: 'shared-iterations',
      vus: 50,                            // 50人が同時アクセス
      iterations: 50,                     // 50リクエスト
      maxDuration: '30s',
      startTime: '2m30s',                 // 通常フロー後に開始
      exec: 'concurrentReservation',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'],     // 95%のリクエストが500ms以内
    http_req_failed: ['rate<0.1'],        // エラー率10%未満
    reservation_success: ['count>0'],     // 最低1件は成功
  },
};

// テスト前のセットアップ
export function setup() {
  // テスト用イベントを作成
  const eventRes = http.post(`${BASE_URL}/api/v1/events`, JSON.stringify({
    name: `負荷テストイベント ${Date.now()}`,
    description: '負荷テスト用',
    venue: 'テスト会場',
    start_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    end_at: new Date(Date.now() + 25 * 60 * 60 * 1000).toISOString(),
    total_seats: 100,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(eventRes, {
    'イベント作成成功': (r) => r.status === 201,
  });

  const event = JSON.parse(eventRes.body);

  // 座席を作成（100席）
  const seats = [];
  for (let i = 1; i <= 100; i++) {
    const seatRes = http.post(`${BASE_URL}/api/v1/events/${event.id}/seats`, JSON.stringify({
      seat_number: `SEAT-${i.toString().padStart(3, '0')}`,
      row: `R${Math.ceil(i / 10)}`,
      section: 'A',
      price: 5000,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    if (seatRes.status === 201) {
      seats.push(JSON.parse(seatRes.body));
    }
  }

  console.log(`セットアップ完了: イベント=${event.id}, 座席数=${seats.length}`);
  return { event, seats };
}

// シナリオ1: 通常フロー
export function normalFlow(data) {
  const userId = `user-${__VU}-${__ITER}`;
  
  group('イベント一覧取得', () => {
    const res = http.get(`${BASE_URL}/api/v1/events`);
    check(res, {
      'イベント一覧取得成功': (r) => r.status === 200,
    });
  });

  group('イベント詳細取得', () => {
    const res = http.get(`${BASE_URL}/api/v1/events/${data.event.id}`);
    check(res, {
      'イベント詳細取得成功': (r) => r.status === 200,
    });
  });

  group('座席一覧取得', () => {
    const res = http.get(`${BASE_URL}/api/v1/events/${data.event.id}/seats`);
    check(res, {
      '座席一覧取得成功': (r) => r.status === 200,
    });
  });

  group('空席数取得', () => {
    const res = http.get(`${BASE_URL}/api/v1/events/${data.event.id}/seats/available/count`);
    check(res, {
      '空席数取得成功': (r) => r.status === 200,
    });
  });

  sleep(1);
}

// シナリオ2: 同時予約テスト（同じ座席を複数人が狙う）
export function concurrentReservation(data) {
  const userId = `user-${__VU}`;
  // 全員が同じ座席を狙う（意図的な競合テスト）
  const targetSeat = data.seats[0];

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/reservations`, JSON.stringify({
    event_id: data.event.id,
    seat_ids: [targetSeat.id],
    idempotency_key: `load-test-${userId}-${Date.now()}`,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'X-User-ID': userId,
    },
  });
  const duration = Date.now() - startTime;
  reservationDuration.add(duration);

  if (res.status === 201) {
    reservationSuccess.add(1);
    console.log(`✅ VU${__VU}: 予約成功 (${duration}ms)`);
  } else if (res.status === 400 || res.status === 409) {
    // 座席が既に予約済み or ロック取得失敗
    reservationConflict.add(1);
    console.log(`⚠️ VU${__VU}: 競合で失敗 (${duration}ms)`);
  } else {
    reservationError.add(1);
    console.log(`❌ VU${__VU}: エラー ${res.status} (${duration}ms)`);
  }

  check(res, {
    '予約リクエスト処理': (r) => r.status === 201 || r.status === 400 || r.status === 409,
  });
}

// テスト後のクリーンアップ
export function teardown(data) {
  // イベントを削除
  http.del(`${BASE_URL}/api/v1/events/${data.event.id}`);
  console.log('クリーンアップ完了');
}
