import http from 'k6/http';
import { check, sleep } from 'k6';

// スモークテスト: 基本的な動作確認
export const options = {
  vus: 1,
  duration: '10s',
  thresholds: {
    http_req_failed: ['rate<0.01'],     // エラー率1%未満
    http_req_duration: ['p(95)<1000'],  // 95%が1秒以内
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';

export default function () {
  // ヘルスチェック
  const healthRes = http.get(`${BASE_URL}/api/v1/health`);
  check(healthRes, {
    'ヘルスチェック成功': (r) => r.status === 200,
  });

  // メトリクスエンドポイント
  const metricsRes = http.get(`${BASE_URL}/metrics`);
  check(metricsRes, {
    'メトリクス取得成功': (r) => r.status === 200,
  });

  // イベント一覧（空でもOK）
  const eventsRes = http.get(`${BASE_URL}/api/v1/events`);
  check(eventsRes, {
    'イベント一覧取得成功': (r) => r.status === 200,
  });

  sleep(1);
}
