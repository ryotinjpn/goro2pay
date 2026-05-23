// LC-34 UlidGenerator
//
// Frontend で idempotencyKey を発行する純関数。
// Q-N4 確定通り Frontend 側で ULID を発行 (NFRC-C05)、24h TTL は Backend 側で管理。
import { ulid } from "ulid";

/**
 * Crockford Base32 形式の 26 文字 ULID を返す。
 *
 * - 時系列ソート可能 (タイムスタンプ前置)
 * - crypto.getRandomValues 互換のエントロピーを使用 (ulid lib のデフォルト)
 *
 * 戻り値の例: `01HZQK7FCS3X4N9QYBV0WJ7M5T`
 */
export function generateUlid(): string {
  return ulid();
}
