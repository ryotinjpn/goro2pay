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
    winVerdict: '決まりだ',
    increaseTemplate: (yen: number) =>
      `¥${yen.toLocaleString()}。来月もこの調子だ。`,
  },

  complete: {
    headline: 'いい判断だ。',
    body: '面倒は片付いた。',
    metricsTemplate: (count: number) => `今月 ${count} 度、いい判断だった。`,
    backLink: '次を待て。',
  },

  signup: {
    h1Line1: '面倒は、',
    h1Line2: 'こちらで引き受ける。',
    sub: '30 秒で済む。',
    submit: '登録する',
    submitting: '登録中…',
    emailLabel: 'メールアドレス',
    passwordLabel: 'パスワード',
    hintLength: '8 文字以上',
    hintUpper: '英大文字を含む',
    hintLower: '英小文字を含む',
    hintDigit: '数字を含む',
  },

  login: {
    h1: '戻ってきたか。',
    sub: 'また面倒になったか。',
    submit: 'ログイン',
    submitting: 'ログイン中…',
    emailLabel: 'メールアドレス',
    passwordLabel: 'パスワード',
  },

  budget: {
    h1: 'ダメ予算を設定する',
    desc: '月間ダメ予算を 1,000〜100,000 円 (1,000 円刻み) で設定してください。即時反映され、月末にリセットされます。',
    inputLabel: '金額を直接入力',
    submit: 'ダメ予算を設定する',
  },

  error: {
    h1: '申し訳ありません',
    body: '予期しないエラーが発生しました。',
    button: 'もう一度',
  },

  notFound: {
    h1: 'ページが見つかりません',
    body: 'お探しのページは見つけられませんでした。',
    link: 'トップに戻る',
  },

  logout: {
    h2: 'やめるのか？',
    sub: '戻ってこい。',
    primary: 'やめる。',
    secondary: '戻る。',
  },

  sessionExpired: {
    h2: '離れすぎたな。',
    body: 'セッションが切れました。再度ログインしてください。',
    button: '戻る。',
  },

  insufficient: {
    h2: '残りダメ予算が足りません…',
    bodyLine1: '今月のダメ予算を使い切りました。',
    bodyLine2: 'もっとダメになる準備はできていますか？',
    primary: 'もっとダメになる',
    secondary: '閉じる',
  },

  raiseBudget: {
    h2: '翌月予算を増額しますか？',
    recommendLabel: '推奨',
    recommendNote: 'あなたには必要です',
    primary: '増額する',
    secondary: '今月はがんばる',
  },
} as const;
