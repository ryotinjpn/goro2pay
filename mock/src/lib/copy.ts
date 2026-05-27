export const COPY = {
  brand: 'ゴロゴロPay',

  landing: {
    h1Top: '考えるな。',
    h1Bottom: '押せ。',
    sub: '面倒は、こちらで引き受ける。',
    demoLabel: '体験用ダメ予算',
    demoButtonInitial: '押す。',
    demoButtonDone: '— 上出来だ。',
    ctaPrimary: '始めろ',
    ctaSecondary: '戻る',
  },

  main: {
    balanceLabel: '残りダメ予算',
    btnIdleMain: 'めんどくさい',
    btnIdleSub: '— 押せ。考えるな。',
    suggestBubble: 'そろそろだろ。',
    btnSuggestMain: '押す。',
    btnDeadSub: '— 上出来だ。使い切ったな。',
    deadVerdict: '今月は、終わりだ。',
    winVerdict: '決まりだ。',
    increaseTemplate: (yen: number) =>
      `¥${yen.toLocaleString()}。来月もこの調子だ。`,
  },

  complete: {
    headline: 'いい判断だ。',
    body: '面倒は片付いた。',
    metricsTemplate: (count: number) => `今月 ${count} 度、いい判断だった。`,
    backLink: '次を待て。',
  },
} as const;
