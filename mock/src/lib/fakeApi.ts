export type OrderResult = {
  storeName: string;
  shortName: string;
  amount: number;
};

const CANDIDATES: OrderResult[] = [
  { storeName: 'ダメ食堂 西新宿店', shortName: 'ダメ食堂', amount: 1180 },
  { storeName: 'ゴロネ亭 渋谷店', shortName: 'ゴロネ亭', amount: 940 },
  { storeName: 'よこなり弁当 本店', shortName: 'よこなり弁当', amount: 820 },
  { storeName: 'もちかえり食堂 池袋店', shortName: 'もちかえり食堂', amount: 1080 },
  { storeName: 'ぐうたらカレー 神田店', shortName: 'ぐうたらカレー', amount: 1280 },
];

export const REEL_STRIP = CANDIDATES.map((c) => c.shortName);

export function placeOrder(): Promise<OrderResult> {
  return new Promise((resolve) => {
    const result = CANDIDATES[Math.floor(Math.random() * CANDIDATES.length)];
    setTimeout(() => resolve(result), 2800);
  });
}

export function suggestNext(): OrderResult {
  return CANDIDATES[0];
}
