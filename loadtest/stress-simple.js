import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// カスタムメトリクス
const reservationSuccess = new Counter('reservation_success');
const reservationConflict = new Counter('reservation_conflict');
const reservationError = new Counter('reservation_error');
const reservationDuration = new Trend('reservation_duration_ms');

// 設定
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// 簡略化されたストレステスト
export const options = {
  stages: [
    { duration: '10s', target: 50 },   // 50 VU まで増加
    { duration: '20s', target: 100 },  // 100 VU まで増加
    { duration: '30s', target: 100 },  // 維持
    { duration: '20s', target: 200 },  // 200 VU まで増加
    { duration: '30s', target: 200 },  // 維持
    { duration: '10s', target: 0 },    // クールダウン
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'],
    http_req_failed: ['rate<0.1'],
  },
};

// グローバル変数（セットアップデータ）
let eventId = null;
let seatIds = [];

// テスト前のセットアップ
export function setup() {
  console.log('ストレステスト開始: 最大 200 VU');
  
  // テスト用イベントを作成
  const eventRes = http.post(`${BASE_URL}/api/v1/events`, JSON.stringify({
    name: `ストレステストイベント ${Date.now()}`,
    description: 'ストレステスト用',
    venue: 'テスト会場',
    start_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    end_at: new Date(Date.now() + 25 * 60 * 60 * 1000).toISOString(),
    total_seats: 500,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (eventRes.status !== 201) {
    console.log(`イベント作成失敗: ${eventRes.status}`);
    return null;
  }

  const event = JSON.parse(eventRes.body);
  console.log(`イベント作成成功: ${event.id}`);

  // 座席を一括作成（500席）
  const bulkSeats = [];
  for (let i = 1; i <= 500; i++) {
    bulkSeats.push({
      seat_number: `S-${i.toString().padStart(4, '0')}`,
      row: `R${Math.ceil(i / 25)}`,
      section: String.fromCharCode(65 + Math.floor((i - 1) / 100)),
      price: 5000,
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
      try {
        const created = JSON.parse(bulkRes.body);
        if (created.seats) {
          seats.push(...created.seats);
        }
      } catch (e) {
        console.log(`座席パース失敗: ${e}`);
      }
    }
  }

  console.log(`セットアップ完了: 座席数=${seats.length}`);
  return { eventId: event.id, seatIds: seats.map(s => s.id) };
}

// メイン処理
export default function(data) {
  if (!data || !data.eventId) {
    console.log('セットアップデータなし');
    return;
  }

  const userId = `user-${__VU}-${__ITER}`;
  const rand = Math.random();

  if (rand < 0.5) {
    // 50%: 読み取り操作
    readOperations(data);
  } else if (rand < 0.8) {
    // 30%: 空席数取得
    availableCount(data);
  } else {
    // 20%: 予約作成
    createReservation(data, userId);
  }

  sleep(0.05 + Math.random() * 0.1);
}

function readOperations(data) {
  const eventsRes = http.get(`${BASE_URL}/api/v1/events`);
  check(eventsRes, { 'イベント一覧取得': (r) => r.status === 200 });
  
  const seatsRes = http.get(`${BASE_URL}/api/v1/events/${data.eventId}/seats`);
  check(seatsRes, { '座席一覧取得': (r) => r.status === 200 });
}

function availableCount(data) {
  const res = http.get(`${BASE_URL}/api/v1/events/${data.eventId}/seats/available/count`);
  check(res, { '空席数取得': (r) => r.status === 200 });
}

function createReservation(data, userId) {
  if (!data.seatIds || data.seatIds.length === 0) return;
  
  const randomIndex = Math.floor(Math.random() * data.seatIds.length);
  const targetSeatId = data.seatIds[randomIndex];

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/reservations`, JSON.stringify({
    event_id: data.eventId,
    seat_ids: [targetSeatId],
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

// クリーンアップ
export function teardown(data) {
  if (data && data.eventId) {
    http.del(`${BASE_URL}/api/v1/events/${data.eventId}`);
    console.log('クリーンアップ完了');
  }
}
