import type { CSSProperties, ReactNode } from "react";

export const metadata = {
  title: "ゴロゴロPay - デザインギャラリー",
};

// ─────── Light Aurora 共通スタイル ───────

const FRAME_STYLE: CSSProperties = {
  position: "relative",
  width: "100%",
  aspectRatio: "9 / 17",
  borderRadius: 22,
  background: "var(--color-bg-base)",
  overflow: "hidden",
  boxShadow: "0 18px 48px -24px rgba(10,10,10,.32), 0 0 0 1px var(--color-line)",
  color: "var(--color-ink)",
};

// フレーム上部に敷く暖色オーロラ
const AURORA_OVERLAY: CSSProperties = {
  position: "absolute",
  top: 0,
  left: 0,
  right: 0,
  height: "46%",
  background: "var(--grad-aurora-soft)",
  pointerEvents: "none",
};

const SCREEN_INNER: CSSProperties = {
  position: "absolute",
  inset: 0,
  padding: "18px 18px 18px",
  display: "flex",
  flexDirection: "column",
};

const HEADER_STYLE: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  paddingBottom: 18,
};

function BrandHeader({ time }: { time: string }) {
  return (
    <div style={HEADER_STYLE}>
      <span style={{ display: "inline-flex", alignItems: "baseline" }}>
        <span
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            color: "var(--color-ink)",
            fontSize: 15,
            letterSpacing: "-0.01em",
          }}
        >
          ゴロゴロ
        </span>
        <span
          style={{
            fontFamily: "var(--font-serif-ital)",
            fontStyle: "italic",
            fontWeight: 600,
            color: "var(--color-ink)",
            fontSize: 17,
            marginLeft: 3,
          }}
        >
          Pay
        </span>
      </span>
      <span
        style={{
          fontFamily: "var(--font-body)",
          fontWeight: 500,
          fontSize: 11,
          color: "var(--color-ink-soft)",
          letterSpacing: "0.22em",
        }}
      >
        {time}
      </span>
    </div>
  );
}

// ピル型ボタン (primary=黒ピル / cta=暖色グラデ / secondary=白+枠)
function PillButton({
  children,
  variant = "cta",
}: {
  children: ReactNode;
  variant?: "primary" | "cta" | "secondary";
}) {
  const base: CSSProperties = {
    display: "flex",
    width: "100%",
    padding: "14px 0",
    borderRadius: "var(--radius-pill)",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    textAlign: "center",
    fontFamily: "var(--font-body)",
    fontSize: 13,
    fontWeight: 600,
    letterSpacing: "0.1em",
  };
  if (variant === "secondary") {
    return (
      <div
        style={{
          ...base,
          background: "var(--color-bg-base)",
          border: "1px solid var(--color-line)",
          color: "var(--color-ink)",
        }}
      >
        {children}
      </div>
    );
  }
  if (variant === "primary") {
    return (
      <div
        style={{
          ...base,
          background: "var(--color-line-strong)",
          color: "#fff",
        }}
      >
        {children}
      </div>
    );
  }
  return (
    <div
      style={{
        ...base,
        background: "var(--grad-cta)",
        color: "#fff",
        boxShadow: "0 16px 36px -14px rgba(255,94,143,.55)",
      }}
    >
      {children}
    </div>
  );
}

function GoroButton({
  state,
}: {
  state: "demo" | "idle" | "suggest" | "dead";
}) {
  const isDead = state === "dead";
  const size = state === "demo" ? 132 : 200;
  const bg = isDead
    ? "linear-gradient(180deg, #d2d2d8, #9a9aa2)"
    : "var(--grad-cta)";
  const shadow = isDead
    ? "none"
    : "0 26px 60px -18px rgba(255,94,143,.6), inset 0 2px 12px rgba(255,255,255,.45)";
  const mainLabel = state === "demo" ? "押す。" : state === "suggest" ? "押す。" : "めんどくさい";
  const mainColor = isDead ? "#5a5a62" : "#fff";
  const subColor = isDead ? "#6a6a72" : "rgba(255,255,255,.9)";
  const subLabel =
    state === "idle"
      ? "— 押せ。考えるな。"
      : state === "suggest"
        ? "— CoCo壱 ¥1,200 だ。"
        : state === "dead"
          ? "— 上出来だ。使い切ったな。"
          : "";
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: "50%",
        background: bg,
        boxShadow: shadow,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        margin: "0 auto",
        gap: 3,
        textAlign: "center",
      }}
    >
      <div
        style={{
          fontFamily: "var(--font-display)",
          fontWeight: 900,
          fontSize: state === "demo" ? 21 : 26,
          letterSpacing: "0.01em",
          color: mainColor,
        }}
      >
        {mainLabel}
      </div>
      {subLabel ? (
        <div
          style={{
            fontFamily: "var(--font-serif-ital)",
            fontStyle: "italic",
            fontWeight: 500,
            fontSize: state === "demo" ? 11 : 13,
            color: subColor,
            letterSpacing: "0.01em",
          }}
        >
          {subLabel}
        </div>
      ) : null}
    </div>
  );
}

function BalanceHero({
  amount,
  count,
  rate,
  dead = false,
}: {
  amount: string;
  count: number;
  rate: number;
  dead?: boolean;
}) {
  const isWarn = rate >= 60 && rate < 80;
  const isDanger = rate >= 80;
  const amountColor = dead
    ? "#b4b4ba"
    : isDanger
      ? "var(--color-state-danger)"
      : isWarn
        ? "var(--color-state-warn)"
        : "var(--color-ink)";
  const meterFill = dead
    ? "#b4b4ba"
    : isDanger
      ? "var(--color-state-danger)"
      : isWarn
        ? "var(--color-state-warn)"
        : "linear-gradient(90deg, var(--color-accent-pink), var(--color-accent-orange))";
  return (
    <div style={{ textAlign: "center" }}>
      <div
        style={{
          fontFamily: "var(--font-body)",
          fontWeight: 600,
          fontSize: 10,
          letterSpacing: "0.24em",
          textTransform: "uppercase",
          color: "var(--color-ink-soft)",
        }}
      >
        残りダメ予算
      </div>
      <div
        style={{
          fontFamily: "var(--font-display)",
          fontWeight: 900,
          fontSize: 48,
          letterSpacing: "-0.03em",
          color: amountColor,
          lineHeight: 1,
          margin: "6px 0 14px",
          fontVariantNumeric: "tabular-nums",
        }}
      >
        {amount}
      </div>
      <div
        style={{
          height: 3,
          background: "var(--color-line)",
          borderRadius: 3,
          margin: "0 12% 14px",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            height: "100%",
            width: `${rate}%`,
            background: meterFill,
          }}
        />
      </div>
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          gap: 18,
          fontFamily: "var(--font-body)",
          fontWeight: 500,
          fontSize: 10,
          color: "var(--color-ink-soft)",
          letterSpacing: "0.18em",
          textTransform: "uppercase",
        }}
      >
        <span>今月 {count} 度</span>
        <span>消化 {rate}%</span>
      </div>
    </div>
  );
}

function FormField({ label, value, type = "text" }: { label: string; value: string; type?: "text" | "password" }) {
  return (
    <div style={{ marginTop: 16 }}>
      <div
        style={{
          fontFamily: "var(--font-body)",
          fontWeight: 600,
          fontSize: 10,
          color: "var(--color-ink-soft)",
          letterSpacing: "0.2em",
          textTransform: "uppercase",
          marginBottom: 6,
        }}
      >
        {label}
      </div>
      <div
        style={{
          background: "var(--color-bg-soft)",
          border: "1px solid var(--color-line)",
          borderRadius: 12,
          padding: "13px 14px",
          color: type === "password" && !value ? "var(--color-ink-soft)" : "var(--color-ink)",
          fontFamily: "var(--font-body)",
          fontSize: 15,
          letterSpacing: type === "password" ? "0.2em" : 0,
        }}
      >
        {type === "password" ? "••••••••" : value}
      </div>
    </div>
  );
}

// ─────── 各画面のモック ───────

function LandingMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="21:07" />
      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center" }}>
        <div style={{ textAlign: "center" }}>
          <div
            style={{
              fontFamily: "var(--font-serif-ital)",
              fontStyle: "italic",
              fontWeight: 500,
              fontSize: 18,
              lineHeight: 1.3,
              color: "var(--color-ink-mid)",
              marginBottom: 16,
            }}
          >
            面倒は、こちらで引き受ける。
          </div>
          <div
            style={{
              fontFamily: "var(--font-display)",
              fontWeight: 900,
              fontSize: 44,
              lineHeight: 0.96,
              letterSpacing: "-0.02em",
              color: "var(--color-ink)",
            }}
          >
            考えるな。
            <br />
            押せ。
          </div>
        </div>
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontWeight: 600,
            fontSize: 10,
            letterSpacing: "0.2em",
            textTransform: "uppercase",
            color: "var(--color-ink-soft)",
            textAlign: "center",
            margin: "32px 0 6px",
          }}
        >
          体験用ダメ予算
        </div>
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 36,
            letterSpacing: "-0.03em",
            color: "var(--color-ink)",
            textAlign: "center",
            marginBottom: 20,
          }}
        >
          ¥1,000
        </div>
        <div style={{ marginBottom: 18 }}>
          <GoroButton state="demo" />
        </div>
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
        <PillButton variant="primary">始めろ</PillButton>
        <PillButton variant="secondary">ログイン</PillButton>
      </div>
    </div>
  );
}

function SignupMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="21:07" />
      <div style={{ display: "flex", flexDirection: "column", flex: 1 }}>
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 30,
            lineHeight: 1.1,
            letterSpacing: "-0.02em",
            color: "var(--color-ink)",
          }}
        >
          面倒は、
          <br />
          こちらで引き受ける。
        </div>
        <div
          style={{
            fontFamily: "var(--font-serif-ital)",
            fontStyle: "italic",
            fontWeight: 500,
            fontSize: 16,
            color: "var(--color-ink-mid)",
            marginTop: 8,
          }}
        >
          30 秒で済む。
        </div>
        <div style={{ marginTop: 22 }}>
          <FormField label="メールアドレス" value="taro@example.com" />
          <FormField label="パスワード" value="" type="password" />
        </div>
        <ul
          style={{
            listStyle: "none",
            padding: 0,
            margin: "8px 0 0 0",
            display: "flex",
            flexDirection: "column",
            gap: 2,
          }}
        >
          {[
            { ok: true, label: "8 文字以上" },
            { ok: true, label: "英大文字を含む" },
            { ok: false, label: "英小文字を含む" },
            { ok: false, label: "数字を含む" },
          ].map((h) => (
            <li
              key={h.label}
              style={{
                fontFamily: "var(--font-body)",
                fontSize: 11,
                letterSpacing: "0.02em",
                color: h.ok ? "var(--color-accent-coral)" : "var(--color-ink-soft)",
              }}
            >
              {h.ok ? "✓" : "・"} {h.label}
            </li>
          ))}
        </ul>
        <div style={{ marginTop: "auto" }}>
          <PillButton variant="cta">登録する</PillButton>
        </div>
      </div>
    </div>
  );
}

function LoginMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="21:07" />
      <div style={{ display: "flex", flexDirection: "column", flex: 1 }}>
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 32,
            letterSpacing: "-0.02em",
            lineHeight: 1.1,
            color: "var(--color-ink)",
          }}
        >
          戻ってきたか。
        </div>
        <div
          style={{
            fontFamily: "var(--font-serif-ital)",
            fontStyle: "italic",
            fontWeight: 500,
            fontSize: 16,
            color: "var(--color-ink-mid)",
            marginTop: 8,
          }}
        >
          また面倒になったか。
        </div>
        <div style={{ marginTop: 22 }}>
          <FormField label="メールアドレス" value="taro@example.com" />
          <FormField label="パスワード" value="" type="password" />
        </div>
        <div style={{ marginTop: "auto" }}>
          <PillButton variant="cta">ログイン</PillButton>
        </div>
      </div>
    </div>
  );
}

function BudgetSetupMock() {
  const QUICK = [10_000, 30_000, 50_000, 80_000, 100_000];
  const selected = 30_000;
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="21:09" />
      <div style={{ display: "flex", flexDirection: "column", flex: 1 }}>
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 28,
            lineHeight: 1.15,
            letterSpacing: "-0.02em",
            color: "var(--color-ink)",
          }}
        >
          まず、線を引け。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontWeight: 400,
            fontSize: 13,
            color: "var(--color-ink-soft)",
            marginTop: 10,
            lineHeight: 1.7,
          }}
        >
          引かなきゃ、始まらん。
        </div>
        <div style={{ marginTop: 22 }}>
          <div
            style={{
              fontFamily: "var(--font-body)",
              fontWeight: 600,
              fontSize: 10,
              color: "var(--color-ink-soft)",
              letterSpacing: "0.2em",
              textTransform: "uppercase",
              marginBottom: 8,
            }}
          >
            月間ダメ予算
          </div>
          <div
            style={{
              display: "flex",
              flexWrap: "wrap",
              gap: 8,
              marginBottom: 14,
            }}
          >
            {QUICK.map((v) => {
              const isSel = v === selected;
              return (
                <div
                  key={v}
                  style={{
                    padding: "10px 16px",
                    borderRadius: "var(--radius-pill)",
                    border: isSel ? "1px solid var(--color-line-strong)" : "1px solid var(--color-line)",
                    background: isSel ? "var(--color-line-strong)" : "var(--color-bg-base)",
                    color: isSel ? "#fff" : "var(--color-ink)",
                    fontFamily: "var(--font-body)",
                    fontWeight: 600,
                    fontSize: 13,
                    letterSpacing: "0.02em",
                    whiteSpace: "nowrap",
                  }}
                >
                  {v === 30_000 ? (
                    <span style={{ color: isSel ? "#fff" : "var(--color-accent-orange)", marginRight: 2 }}>★</span>
                  ) : null}
                  ¥{v.toLocaleString()}
                </div>
              );
            })}
          </div>
          <div
            style={{
              width: "100%",
              background: "var(--color-bg-soft)",
              border: "1px solid var(--color-line)",
              borderRadius: 12,
              padding: "13px 14px",
              color: "var(--color-ink)",
              fontFamily: "var(--font-body)",
              fontSize: 15,
              fontVariantNumeric: "tabular-nums",
              textAlign: "right",
            }}
          >
            ¥30,000
          </div>
          <div
            style={{
              fontFamily: "var(--font-body)",
              fontSize: 11,
              color: "var(--color-ink-soft)",
              letterSpacing: "0.02em",
              marginTop: 6,
            }}
          >
            1,000 〜 100,000 円 / 1,000 円刻み
          </div>
        </div>
        <div style={{ marginTop: "auto" }}>
          <PillButton variant="cta">線を引け</PillButton>
        </div>
      </div>
    </div>
  );
}

function InsufficientModalMock() {
  return (
    <ModalMock
      title="足りないか。"
      sub="もっと欲しけりゃ、線を引き直せ。"
      subStyle="body"
      secondary="閉じろ。"
      primary="引き直す。"
      primaryVariant="cta"
      showBackground="suggest"
    />
  );
}

function MainIdleMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="19:43" />
      <div style={{ marginTop: 14 }}>
        <BalanceHero amount="¥28,800" count={5} rate={4} />
      </div>
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>
        <GoroButton state="idle" />
      </div>
    </div>
  );
}

function MainSuggestMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="19:43" />
      <div style={{ marginTop: 14 }}>
        <BalanceHero amount="¥17,200" count={12} rate={43} />
      </div>
      <div
        style={{
          flex: 1,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          position: "relative",
        }}
      >
        <div
          style={{
            position: "absolute",
            top: "12%",
            background: "var(--color-line-strong)",
            borderRadius: 14,
            padding: "9px 15px",
            fontFamily: "var(--font-body)",
            fontWeight: 600,
            fontSize: 13,
            color: "#fff",
            letterSpacing: "0.04em",
            boxShadow: "0 14px 30px -12px rgba(0,0,0,.35)",
          }}
        >
          そろそろだろ。
        </div>
        <GoroButton state="suggest" />
      </div>
    </div>
  );
}

function CompleteMock() {
  return (
    <div style={{ ...SCREEN_INNER, justifyContent: "flex-start" }}>
      <BrandHeader time="19:46" />
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          textAlign: "center",
          position: "relative",
          background:
            "radial-gradient(ellipse at center, rgba(255,123,77,.18), transparent 60%)",
          margin: "0 -18px",
          padding: "0 18px",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 44,
            letterSpacing: "-0.03em",
            color: "var(--color-ink)",
            marginBottom: 16,
          }}
        >
          いい判断だ。
        </div>
        <div
          style={{
            fontFamily: "var(--font-serif-ital)",
            fontStyle: "italic",
            fontWeight: 500,
            fontSize: 17,
            color: "var(--color-ink-mid)",
            marginBottom: 28,
          }}
        >
          面倒は片付いた。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontWeight: 600,
            fontSize: 15,
            color: "var(--color-ink)",
            letterSpacing: "0.02em",
          }}
        >
          CoCo壱番屋 新宿店
        </div>
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 30,
            letterSpacing: "-0.02em",
            color: "var(--color-ink)",
            marginTop: 4,
            fontVariantNumeric: "tabular-nums",
          }}
        >
          ¥1,200
        </div>
        <div
          style={{
            width: "60%",
            height: 1,
            background: "var(--color-line)",
            margin: "20px 0",
          }}
        />
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontWeight: 600,
            fontSize: 11,
            color: "var(--color-ink-soft)",
            letterSpacing: "0.18em",
            textTransform: "uppercase",
            marginBottom: 8,
          }}
        >
          今月 6 度、いい判断だった。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body)",
            fontWeight: 500,
            fontSize: 11,
            color: "var(--color-ink-soft)",
            letterSpacing: "0.18em",
            textTransform: "uppercase",
          }}
        >
          残りダメ予算 ¥27,600
        </div>
      </div>
      <div
        style={{
          textAlign: "center",
          paddingTop: 8,
        }}
      >
        <span
          style={{
            display: "inline-block",
            fontFamily: "var(--font-body)",
            fontWeight: 600,
            fontSize: 12,
            color: "var(--color-ink)",
            letterSpacing: "0.12em",
            padding: "12px 24px",
            border: "1px solid var(--color-line)",
            borderRadius: "var(--radius-pill)",
          }}
        >
          次を待て。
        </span>
      </div>
    </div>
  );
}

function MainDeadMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="23:51" />
      <div style={{ marginTop: 14 }}>
        <BalanceHero amount="¥0" count={47} rate={100} dead />
      </div>
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          position: "relative",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 28,
            letterSpacing: "-0.02em",
            color: "var(--color-ink)",
            textAlign: "center",
            marginBottom: 20,
          }}
        >
          今月は、終わりだ。
        </div>
        <GoroButton state="dead" />
      </div>
      <div
        style={{
          background: "var(--grad-cta)",
          color: "#fff",
          textAlign: "center",
          padding: "16px 0",
          borderRadius: "var(--radius-pill)",
          fontFamily: "var(--font-body)",
          fontWeight: 700,
          fontSize: 14,
          letterSpacing: "0.06em",
          boxShadow: "0 18px 40px -14px rgba(255,94,143,.6)",
        }}
      >
        ¥50,000。来月もこの調子だ。
      </div>
    </div>
  );
}

function ModalMock({
  title,
  sub,
  subStyle = "serif",
  primary,
  primaryVariant = "primary",
  secondary,
  showBackground = "idle",
}: {
  title: string;
  sub: string;
  subStyle?: "serif" | "body";
  primary?: string;
  primaryVariant?: "primary" | "cta";
  secondary?: string;
  showBackground?: "idle" | "suggest";
}) {
  const Bg = showBackground === "suggest" ? MainSuggestMock : MainIdleMock;
  const subStyleProps: CSSProperties =
    subStyle === "serif"
      ? {
          fontFamily: "var(--font-serif-ital)",
          fontStyle: "italic",
          fontWeight: 500,
          fontSize: 15,
          color: "var(--color-ink-mid)",
        }
      : {
          fontFamily: "var(--font-body)",
          fontWeight: 400,
          fontSize: 13,
          color: "var(--color-ink-soft)",
          lineHeight: 1.7,
        };
  return (
    <div style={{ position: "absolute", inset: 0 }}>
      <div style={{ filter: "blur(4px)", height: "100%" }}>
        <Bg />
      </div>
      <div
        style={{
          position: "absolute",
          inset: 0,
          background: "rgba(20,15,10,.32)",
          display: "grid",
          placeItems: "center",
          padding: 20,
        }}
      >
        <div
          style={{
            width: "100%",
            background: "var(--color-bg-base)",
            border: "1px solid var(--color-line)",
            borderRadius: "var(--radius-card)",
            boxShadow: "0 40px 80px -24px rgba(0,0,0,.3)",
            padding: "28px 22px",
            textAlign: "center",
          }}
        >
          <div
            style={{
              fontFamily: "var(--font-display)",
              fontWeight: 900,
              fontSize: 24,
              letterSpacing: "-0.02em",
              color: "var(--color-ink)",
            }}
          >
            {title}
          </div>
          <div
            style={{
              ...subStyleProps,
              marginTop: 12,
              marginBottom: 22,
            }}
          >
            {sub}
          </div>
          <div style={{ display: "flex", gap: 10 }}>
            {secondary ? (
              <div style={{ flex: 1 }}>
                <PillButton variant="secondary">{secondary}</PillButton>
              </div>
            ) : null}
            {primary ? (
              <div style={{ flex: 1 }}>
                <PillButton variant={primaryVariant}>{primary}</PillButton>
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </div>
  );
}

function LogoutModalMock() {
  return <ModalMock title="やめるのか？" sub="戻ってこい。" secondary="戻る。" primary="やめる。" primaryVariant="primary" />;
}

function SessionExpiredModalMock() {
  return (
    <ModalMock
      title="離れすぎたな。"
      sub="セッションが切れました。再ログインしてください。"
      subStyle="body"
      primary="戻る。"
      primaryVariant="primary"
      showBackground="idle"
    />
  );
}

// ─────── ギャラリー ───────

const SCREENS: Array<{
  no: string;
  title: string;
  desc: string;
  Mock: () => ReactNode;
}> = [
  { no: "01", title: "LANDING", desc: "未ログイン初見。ヒーロー＋デモボタン1回＋CTA", Mock: LandingMock },
  { no: "02", title: "SIGNUP", desc: "登録フォーム。見出しのみリヴァイ調、フォームは中立", Mock: SignupMock },
  { no: "03", title: "LOGIN", desc: "既存ユーザー復帰。見出しのみリヴァイ調", Mock: LoginMock },
  { no: "04", title: "BUDGET SETUP", desc: "予算未設定時の強制セットアップ。線を引かなきゃ始まらない", Mock: BudgetSetupMock },
  { no: "05", title: "MAIN · IDLE", desc: "ログイン後の通常状態", Mock: MainIdleMock },
  { no: "06", title: "MAIN · SUGGEST", desc: "先回り提案。吹き出しが上から目線で命令", Mock: MainSuggestMock },
  { no: "07", title: "COMPLETE", desc: "注文完了。上から目線の褒め", Mock: CompleteMock },
  { no: "08", title: "MAIN · DEAD", desc: "残高 0。脱色＋赤い増額ボタンだけ生きている", Mock: MainDeadMock },
  { no: "09", title: "INSUFFICIENT", desc: "注文時に予算不足。引き直せの命令", Mock: InsufficientModalMock },
  { no: "10", title: "LOGOUT MODAL", desc: "ログアウト確認", Mock: LogoutModalMock },
  { no: "11", title: "SESSION EXPIRED", desc: "セッション切れ通知", Mock: SessionExpiredModalMock },
];

export default function DesignGalleryPage() {
  return (
    <main
      style={{
        position: "relative",
        minHeight: "100vh",
        background: "var(--color-bg-base)",
        padding: "48px 32px 80px",
        overflow: "hidden",
      }}
    >
      {/* Light Aurora: ページ上部に暖色オーロラを敷く */}
      <div
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          height: "40vh",
          background: "var(--grad-aurora-soft)",
          pointerEvents: "none",
        }}
      />
      <header style={{ position: "relative", maxWidth: 1400, margin: "0 auto 36px" }}>
        <h1
          style={{
            margin: 0,
            fontFamily: "var(--font-display)",
            fontWeight: 900,
            fontSize: 22,
            color: "var(--color-ink)",
            letterSpacing: "-0.01em",
          }}
        >
          <span style={{ fontFamily: "var(--font-display)" }}>ゴロゴロ</span>
          <span style={{ fontFamily: "var(--font-serif-ital)", fontStyle: "italic" }}>Pay</span>{" "}
          - 全画面モック (リヴァイ調コピー統一)
        </h1>
        <p
          style={{
            margin: "8px 0 0",
            color: "var(--color-ink-soft)",
            fontFamily: "var(--font-body)",
            fontSize: 12,
            letterSpacing: "0.04em",
          }}
        >
          ランディング / サインアップ / ログイン / 予算セットアップ / メイン (idle, suggest, dead) / 完了 / 予算不足モーダル / ログアウトモーダル / セッション切れモーダル
        </p>
      </header>

      <div
        style={{
          position: "relative",
          maxWidth: 1400,
          margin: "0 auto",
          display: "grid",
          gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))",
          gap: "32px 28px",
        }}
      >
        {SCREENS.map(({ no, title, desc, Mock }) => (
          <section key={no}>
            <div
              style={{
                fontFamily: "var(--font-body)",
                fontWeight: 700,
                fontSize: 11,
                letterSpacing: "0.18em",
                color: "var(--color-accent-orange)",
                marginBottom: 4,
              }}
            >
              {no} / {title}
            </div>
            <div
              style={{
                fontSize: 11,
                color: "var(--color-ink-soft)",
                marginBottom: 14,
                fontFamily: "var(--font-body)",
              }}
            >
              {desc}
            </div>
            <div style={FRAME_STYLE}>
              <div style={AURORA_OVERLAY} />
              <Mock />
            </div>
          </section>
        ))}
      </div>
    </main>
  );
}
