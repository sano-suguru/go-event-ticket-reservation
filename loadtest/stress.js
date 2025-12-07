import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// カスタムメトリクス
const reservationSuccess = new Counter('reservation_success');
const reservationConflict = new Counter('reservation_conflict');
const reservationError = new Counter('reservation_error');
const reservationDuration = new Trend('reservation_duration_ms');
const apiLatency = new Trend('api_latency_ms');

// 設定
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';
const MAX_VUS = parseInt(__ENV.MAX_VUS) || 200;

// ストレステストシナリオ
export const options = {
  scenarios: {
    // シナリオ1: 段階的負荷増加（100→200→500 VU）
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },    // ウォームアップ
        { duration: '1m', target: 100 },    // 100 VU
        { duration: '1m', target: 100 },    // 維持
        { duration: '1m', target: 200 },    // 200 VU
        { duration: '1m', target: 200 },    // 維持
        { duration: '1m', target: MAX_VUS }, // 最大負荷
        { duration: '1m', target: MAX_VUS }, // 維持
        { duration: '30s', target: 0 },     // クールダウン
      ],
      gracefulRampDown: '10s',
      exec: 'mixedWorkload',
    },
    // シナリオ2: 瞬間的な高負荷（スパイクテスト）
    spike_test: {
      executor: 'shared-iterations',
      vus: 100,
      iterations: 100,
      maxDuration: '30s',
      startTime: '7m',  // ストレステスト後
      exec: 'concurrentReservation',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'],  // p95<1s, p99<2s
    http_req_failed: ['rate<0.05'],                   // エラー率5%未満
    reservation_success: ['count>0'],
    api_latency_ms: ['p(95)<500'],
  },
};

// テスト前のセットアップ
export function setup() {
  console.log(`ストレステスト開始: 最大 ${MAX_VUS} VU`);
  
  // テスト用イベントを作成
  const eventRes = http.post(`${BASE_URL}/api/v1/events`, JSON.stringify({
    name: `ストレステストイベント ${Date.now()}`,
    description: 'ストレステスト用 - 大規模負荷',
    venue: 'テスト会場',
    start_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    end_at: new Date(Date.now() + 25 * 60 * 60 * 1000).toISOString(),
    total_seats: 1000,  // 大規模イベント
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(eventRes, {
    'イベント作成成功': (r) => r.status === 201,
  });

  const event = JSON.parse(eventRes.body);

  // 座席を一括作成（1000席）
  const bulkSeats = [];
  for (let i = 1; i <= 1000; i++) {
    bulkSeats.push({
      seat_number: `SEAT-${i.toString().padStart(4, '0')}`,
      row: `R${Math.ceil(i / 50)}`,
      section: String.fromCharCode(65 + Math.floor((i - 1) / 200)), // A-E
      price: 5000 + (Math.floor(i / 100) * 1000),
    });
  }

  // 100席ずつ一括作成
  const seats = [];
  for (let i = 0; i < bulkSeats.length; i += 100) {
    const batch = bulkSeats.slice(i, i + 100);
    const bulkRes = http.post(
      `${BASE_URL}/api/v1/events/${event.id}/seats/bulk`,
      JSON.stringify({ seats: batch }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    if (bulkRes.status === 201) {
      const created = JSON.parse(bulkRes.body);
      if (created.seats) {
        seats.push(...created.seats);
      }
    }
  }

  console.log(`セットアップ完了: イベント=${event.id}, 座席数=${seats.length}`);
  return { event, seats };
}

// シナリオ1: 混合ワークロード（読み取り80%、書き込み20%）
export function mixedWorkload(data) {
  const userId = `user-${__VU}-${__ITER}`;
  const rand = Math.random();

  if (rand < 0.4) {
    // 40%: イベント・座席一覧取得
    readOperations(data);
  } else if (rand < 0.8) {
    // 40%: 空席数取得（キャッシュテスト）
    availableCountCheck(data);
  } else {
    // 20%: 予約作成（書き込み）
    createReservation(data, userId);
  }

  sleep(0.1 + Math.random() * 0.2);  // 100-300ms のランダム待機
}

function readOperations(data) {
  const start = Date.now();
  
  group('読み取り操作', () => {
    // イベント一覧
    const eventsRes = http.get(`${BASE_URL}/api/v1/events`);
    check(eventsRes, { 'イベント一覧取得': (r) => r.status === 200 });
    
    // 座席一覧
    const seatsRes = http.get(`${BASE_URL}/api/v1/events/${data.event.id}/seats`);
    check(seatsRes, { '座席一覧取得': (r) => r.status === 200 });
  });
  
  apiLatency.add(Date.now() - start);
}

function availableCountCheck(data) {
  const start = Date.now();
  
  const res = http.get(`${BASE_URL}/api/v1/events/${data.event.id}/seats/available/count`);
  check(res, { '空席数取得': (r) => r.status === 200 });
  
  apiLatency.add(Date.now() - start);
}

function createReservation(data, userId) {
  // ランダムな座席を選択（競合を減らす）
  const randomIndex = Math.floor(Math.random() * data.seats.length);
  const targetSeat = data.seats[randomIndex];
  
  if (!targetSeat) return;

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/reservations`, JSON.stringify({
    event_id: data.event.id,
    seat_ids: [targetSeat.id],
    idempotency_key: `stress-${userId}-${Date.now()}-${Math.random()}`,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'X-User-ID': userId,
    },
  });
  const duration = Date.now() - start;
  reservationDuration.add(duration);

  if (res.status === 201) {
    reservationSuccess.add(1);
  } else if (res.status === 400 || res.status === 409) {
    reservationConflict.add(1);
  } else {
    reservationError.add(1);
  }

  check(res, {
    '予約リクエスト処理': (r) => r.status === 201 || r.status === 400 || r.status === 409,
  });
}

// シナリオ2: 同時予約テスト（100人が同じ座席を狙う）
export function concurrentReservation(data) {
  const userId = `spike-user-${__VU}`;
  // 全員が同じ座席を狙う（意図的な競合テスト）
  const targetSeat = data.seats[0];

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/reservations`, JSON.stringify({
    event_id: data.event.id,
    seat_ids: [targetSeat.id],
    idempotency_key: `spike-${userId}-${Date.now()}`,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'X-User-ID': userId,
    },
  });
  const duration = Date.now() - start;
  reservationDuration.add(duration);

  if (res.status === 201) {
    reservationSuccess.add(1);
    console.log(`✅ VU${__VU}: 予約成功 (${duration}ms)`);
  } else if (res.status === 400 || res.status === 409) {
    reservationConflict.add(1);
  } else {
    reservationError.add(1);
    console.log(`❌ VU${__VU}: エラー ${res.status}`);
  }

  check(res, {
    '予約リクエスト処理': (r) => r.status === 201 || r.status === 400 || r.status === 409,
  });
}

// テスト後のクリーンアップ
export function teardown(data) {
  http.del(`${BASE_URL}/api/v1/events/${data.event.id}`);
  console.log('クリーンアップ完了');
}
