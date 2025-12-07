import http from 'k6/http';
import { check } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// カスタムメトリクス
const reservationSuccess = new Counter('reservation_success');
const reservationConflict = new Counter('reservation_conflict');
const reservationError = new Counter('reservation_error');
const reservationDuration = new Trend('reservation_duration_ms');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const CONCURRENT_USERS = parseInt(__ENV.CONCURRENT_USERS) || 100;

// 100人が同時に同じ座席を予約
export const options = {
  scenarios: {
    concurrent_reservation: {
      executor: 'shared-iterations',
      vus: CONCURRENT_USERS,
      iterations: CONCURRENT_USERS,
      maxDuration: '30s',
    },
  },
  thresholds: {
    reservation_success: ['count==1'],  // 必ず1人だけ成功
    reservation_error: ['count==0'],    // エラーは0件
  },
};

export function setup() {
  console.log(`${CONCURRENT_USERS}人同時予約テスト開始`);
  
  // イベント作成
  const eventRes = http.post(`${BASE_URL}/api/v1/events`, JSON.stringify({
    name: `競合テスト ${Date.now()}`,
    description: '100人同時予約テスト',
    venue: 'テスト会場',
    start_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    end_at: new Date(Date.now() + 25 * 60 * 60 * 1000).toISOString(),
    total_seats: 1,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const event = JSON.parse(eventRes.body);

  // 1座席だけ作成（全員がこれを狙う）
  const seatRes = http.post(`${BASE_URL}/api/v1/events/${event.id}/seats`, JSON.stringify({
    seat_number: 'VIP-001',
    row: 'VIP',
    section: 'A',
    price: 50000,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const seat = JSON.parse(seatRes.body);
  console.log(`セットアップ完了: event=${event.id}, seat=${seat.id}`);
  
  return { eventId: event.id, seatId: seat.id };
}

export default function(data) {
  const userId = `concurrent-user-${__VU}`;
  
  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/reservations`, JSON.stringify({
    event_id: data.eventId,
    seat_ids: [data.seatId],
    idempotency_key: `concurrent-${userId}-${Date.now()}`,
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
    // 競合は正常動作なのでログ出力しない
  } else {
    reservationError.add(1);
    console.log(`❌ VU${__VU}: エラー ${res.status} - ${res.body}`);
  }

  check(res, {
    '予約処理完了': (r) => r.status === 201 || r.status === 400 || r.status === 409,
  });
}

export function teardown(data) {
  http.del(`${BASE_URL}/api/v1/events/${data.eventId}`);
  console.log('クリーンアップ完了');
}
