import type { CSSProperties, ReactNode } from "react";

export const metadata = {
  title: "ゴロゴロPay - デザインギャラリー",
};

const SCAN_LINES =
  "repeating-linear-gradient(0deg, rgba(255,255,255,.025) 0 1px, transparent 1px 4px)";

const FRAME_STYLE: CSSProperties = {
  position: "relative",
  width: "100%",
  aspectRatio: "9 / 17",
  borderRadius: 18,
  background:
    "radial-gradient(ellipse at 50% 0%, var(--color-bg-warm) 0%, var(--color-bg-mid) 55%, var(--color-bg-deep) 100%)",
  overflow: "hidden",
  boxShadow: "0 6px 24px rgba(0,0,0,.4), 0 0 0 1px rgba(201,169,107,.18)",
  color: "var(--color-cream)",
};

const SCANLINE_OVERLAY: CSSProperties = {
  position: "absolute",
  inset: 0,
  background: SCAN_LINES,
  pointerEvents: "none",
};

const SCREEN_INNER: CSSProperties = {
  position: "absolute",
  inset: 0,
  padding: "20px 18px 18px",
  display: "flex",
  flexDirection: "column",
};

const HEADER_STYLE: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "baseline",
  fontSize: 11,
  letterSpacing: "0.06em",
  color: "var(--color-mute)",
  fontFamily: "var(--font-mono-pixel)",
};

function BrandHeader({ time }: { time: string }) {
  return (
    <div style={HEADER_STYLE}>
      <span>
        <span style={{ fontFamily: "var(--font-body-mincho)", fontWeight: 900, color: "var(--color-cream)", fontSize: 12, letterSpacing: "0.06em" }}>
          ゴロゴロ
        </span>
        <span style={{ fontFamily: "var(--font-display-serif)", fontStyle: "italic", color: "var(--color-cream)", fontSize: 13 }}>
          Pay
        </span>
      </span>
      <span>{time}</span>
    </div>
  );
}

function GoldButton({
  children,
  variant = "primary",
}: {
  children: ReactNode;
  variant?: "primary" | "secondary";
}) {
  const base: CSSProperties = {
    display: "block",
    width: "100%",
    padding: "10px 0",
    borderRadius: 4,
    textAlign: "center",
    fontFamily: "var(--font-mono-pixel)",
    fontSize: 11,
    letterSpacing: "0.18em",
  };
  if (variant === "primary") {
    return (
      <div
        style={{
          ...base,
          background: "linear-gradient(180deg, var(--color-gold-100), var(--color-gold-700))",
          color: "var(--color-bg-warm)",
          fontWeight: 700,
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
        background: "transparent",
        border: "1px solid var(--color-gold-500)",
        color: "var(--color-gold-100)",
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
  const size = state === "demo" ? 130 : 200;
  const bg = isDead
    ? "radial-gradient(circle at 35% 30%, #888 0%, #555 50%, #2c2c2c 100%)"
    : "radial-gradient(circle at 35% 28%, var(--color-gold-100) 0%, var(--color-gold-500) 45%, var(--color-gold-700) 80%, var(--color-gold-900) 100%)";
  const borderGlow = isDead
    ? "0 0 0 1px rgba(106,106,106,.5)"
    : "0 0 0 1px rgba(244,217,144,.4), 0 6px 22px rgba(201,169,107,.3)";
  const mainLabel = state === "demo" ? "押す。" : state === "suggest" ? "押す。" : "めんどくさい";
  const mainColor = isDead ? "#bcbcbc" : "var(--color-bg-warm)";
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
        boxShadow: borderGlow,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        margin: "0 auto",
        color: mainColor,
        fontFamily: "var(--font-body-mincho)",
        fontSize: state === "demo" ? 14 : 18,
        fontWeight: 700,
        letterSpacing: "0.04em",
        textAlign: "center",
      }}
    >
      <div>{mainLabel}</div>
      {subLabel ? (
        <div
          style={{
            marginTop: 4,
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontSize: 9,
            color: isDead ? "#9a9a9a" : "rgba(26,18,8,.78)",
            fontWeight: 400,
            letterSpacing: 0,
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
  dimmed = false,
}: {
  amount: string;
  count: number;
  rate: number;
  dimmed?: boolean;
}) {
  const amountColor = dimmed ? "var(--color-dead-gray)" : "var(--color-cream)";
  const meterColor =
    rate >= 80 ? "var(--color-accent-danger)" : rate >= 60 ? "var(--color-accent-warn)" : "var(--color-gold-500)";
  return (
    <div style={{ textAlign: "center" }}>
      <div
        style={{
          fontFamily: "var(--font-mono-pixel)",
          fontSize: 9,
          letterSpacing: "0.2em",
          color: "var(--color-mute)",
        }}
      >
        残りダメ予算
      </div>
      <div
        style={{
          fontFamily: "var(--font-display-serif)",
          fontStyle: "italic",
          fontWeight: 700,
          fontSize: 38,
          color: amountColor,
          marginTop: 4,
          textShadow: dimmed ? "none" : "0 0 18px rgba(244,236,216,.18)",
        }}
      >
        {amount}
      </div>
      <div
        style={{
          height: 1,
          width: "70%",
          margin: "10px auto 6px",
          background: dimmed ? "var(--color-dead-gray)" : meterColor,
          opacity: dimmed ? 0.5 : 1,
        }}
      />
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          width: "70%",
          margin: "0 auto",
          fontFamily: "var(--font-mono-pixel)",
          fontSize: 9,
          color: dimmed ? "var(--color-dead-gray)" : "var(--color-mute)",
          letterSpacing: "0.16em",
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
    <div style={{ marginBottom: 12 }}>
      <div
        style={{
          fontFamily: "var(--font-mono-pixel)",
          fontSize: 9,
          color: "var(--color-mute)",
          letterSpacing: "0.2em",
          marginBottom: 4,
        }}
      >
        {label}
      </div>
      <div
        style={{
          background: "rgba(255,255,255,.04)",
          border: "1px solid #4a3a18",
          borderRadius: 4,
          padding: "8px 10px",
          color: "var(--color-cream)",
          fontFamily: type === "password" ? "monospace" : "var(--font-body-sans)",
          fontSize: 12,
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
        <div
          style={{
            textAlign: "center",
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontWeight: 700,
            fontSize: 36,
            lineHeight: 1.1,
            color: "var(--color-cream)",
          }}
        >
          考えるな。
          <br />
          押せ。
        </div>
        <div
          style={{
            textAlign: "center",
            fontFamily: "var(--font-body-mincho)",
            fontSize: 11,
            color: "var(--color-gold-500)",
            margin: "16px 0 18px",
            letterSpacing: "0.04em",
          }}
        >
          面倒は、こちらで引き受ける。
        </div>
        <div
          style={{
            textAlign: "center",
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 9,
            color: "var(--color-mute)",
            letterSpacing: "0.2em",
            marginBottom: 12,
          }}
        >
          体験用ダメ予算{" "}
          <span style={{ fontFamily: "var(--font-display-serif)", fontStyle: "italic", color: "var(--color-cream)", fontSize: 12 }}>
            ¥1,000
          </span>
        </div>
        <div style={{ marginBottom: 18 }}>
          <GoroButton state="demo" />
        </div>
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        <GoldButton variant="primary">始めろ</GoldButton>
        <GoldButton variant="secondary">ログイン</GoldButton>
      </div>
    </div>
  );
}

function SignupMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="21:07" />
      <div style={{ marginTop: 18 }}>
        <div
          style={{
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontWeight: 700,
            fontSize: 22,
            lineHeight: 1.2,
            color: "var(--color-cream)",
          }}
        >
          面倒は、
          <br />
          こちらで引き受ける。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body-mincho)",
            fontSize: 10,
            color: "var(--color-gold-500)",
            marginTop: 10,
            marginBottom: 18,
            letterSpacing: "0.04em",
          }}
        >
          30 秒で済む。
        </div>
        <FormField label="メールアドレス" value="taro@example.com" />
        <FormField label="パスワード" value="" type="password" />
        <ul
          style={{
            listStyle: "none",
            padding: 0,
            margin: "0 0 16px 0",
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 9,
            letterSpacing: "0.1em",
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
              style={{ color: h.ok ? "var(--color-gold-500)" : "var(--color-mute)", marginBottom: 2 }}
            >
              {h.ok ? "✓" : "・"} {h.label}
            </li>
          ))}
        </ul>
      </div>
      <div style={{ marginTop: "auto" }}>
        <GoldButton>登録する</GoldButton>
      </div>
    </div>
  );
}

function LoginMock() {
  return (
    <div style={SCREEN_INNER}>
      <BrandHeader time="21:07" />
      <div style={{ marginTop: 18 }}>
        <div
          style={{
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontWeight: 700,
            fontSize: 26,
            color: "var(--color-cream)",
          }}
        >
          戻ってきたか。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body-mincho)",
            fontSize: 10,
            color: "var(--color-gold-500)",
            marginTop: 8,
            marginBottom: 22,
            letterSpacing: "0.04em",
          }}
        >
          また面倒になったか。
        </div>
        <FormField label="メールアドレス" value="taro@example.com" />
        <FormField label="パスワード" value="" type="password" />
      </div>
      <div style={{ marginTop: "auto" }}>
        <GoldButton>ログイン</GoldButton>
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
      <div style={{ marginTop: 18 }}>
        <div
          style={{
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontWeight: 700,
            fontSize: 24,
            lineHeight: 1.2,
            color: "var(--color-cream)",
          }}
        >
          まず、線を引け。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body-mincho)",
            fontSize: 10,
            color: "var(--color-gold-500)",
            marginTop: 8,
            marginBottom: 18,
            letterSpacing: "0.04em",
          }}
        >
          引かなきゃ、始まらん。
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 9,
            color: "var(--color-mute)",
            letterSpacing: "0.2em",
            marginBottom: 8,
          }}
        >
          月間ダメ予算
        </div>
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: 6,
            marginBottom: 14,
          }}
        >
          {QUICK.map((v) => {
            const isSel = v === selected;
            return (
              <div
                key={v}
                style={{
                  padding: "6px 10px",
                  borderRadius: 4,
                  border: isSel
                    ? "1px solid var(--color-gold-100)"
                    : "1px solid #4a3a18",
                  background: isSel
                    ? "linear-gradient(180deg, var(--color-gold-100), var(--color-gold-700))"
                    : "transparent",
                  color: isSel ? "var(--color-bg-warm)" : "var(--color-cream)",
                  fontFamily: "var(--font-mono-pixel)",
                  fontSize: 10,
                  letterSpacing: "0.06em",
                  fontWeight: isSel ? 700 : 400,
                }}
              >
                {v === 30_000 ? "★ " : ""}¥{v.toLocaleString()}
              </div>
            );
          })}
        </div>
        <div
          style={{
            background: "rgba(255,255,255,.04)",
            border: "1px solid #4a3a18",
            borderRadius: 4,
            padding: "10px 12px",
            color: "var(--color-cream)",
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontSize: 22,
            letterSpacing: "0.02em",
            textAlign: "right",
          }}
        >
          ¥30,000
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 8,
            color: "var(--color-mute)",
            letterSpacing: "0.16em",
            marginTop: 6,
          }}
        >
          1,000 〜 100,000 円 / 1,000 円刻み
        </div>
      </div>
      <div style={{ marginTop: "auto" }}>
        <GoldButton>線を引け</GoldButton>
      </div>
    </div>
  );
}

function InsufficientModalMock() {
  return (
    <ModalMock
      title="足りないか。"
      sub="もっと欲しけりゃ、線を引き直せ。"
      secondary="閉じろ。"
      primary="引き直す。"
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
            top: "8%",
            background: "rgba(26,18,8,.92)",
            border: "1px solid var(--color-gold-500)",
            borderRadius: 14,
            padding: "8px 14px",
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontSize: 14,
            color: "var(--color-cream)",
            boxShadow: "0 0 14px rgba(201,169,107,.25)",
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
          background:
            "radial-gradient(ellipse at center, rgba(201,169,107,.32), transparent 60%)",
          margin: "0 -18px",
          padding: "0 18px",
        }}
      >
        <div
          style={{
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontWeight: 700,
            fontSize: 38,
            color: "var(--color-cream)",
            marginBottom: 14,
          }}
        >
          いい判断だ。
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 10,
            color: "var(--color-mute)",
            letterSpacing: "0.18em",
            marginBottom: 16,
          }}
        >
          面倒は片付いた。
        </div>
        <div
          style={{
            fontFamily: "var(--font-body-mincho)",
            fontSize: 14,
            color: "var(--color-cream)",
          }}
        >
          CoCo壱番屋 新宿店
        </div>
        <div
          style={{
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontSize: 22,
            color: "var(--color-gold-500)",
            marginTop: 4,
          }}
        >
          ¥1,200
        </div>
        <div
          style={{
            width: "60%",
            height: 1,
            background: "var(--color-gold-500)",
            opacity: 0.5,
            margin: "16px 0",
          }}
        />
        <div
          style={{
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 9,
            color: "var(--color-gold-500)",
            letterSpacing: "0.16em",
          }}
        >
          今月 6 度、いい判断だった。
        </div>
        <div
          style={{
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 9,
            color: "var(--color-mute)",
            letterSpacing: "0.16em",
            marginTop: 4,
          }}
        >
          残りダメ予算 ¥27,600
        </div>
      </div>
      <div
        style={{
          textAlign: "center",
          fontFamily: "var(--font-mono-pixel)",
          fontSize: 10,
          color: "var(--color-mute)",
          letterSpacing: "0.18em",
          paddingTop: 8,
        }}
      >
        次を待て。
      </div>
    </div>
  );
}

function MainDeadMock() {
  return (
    <div
      style={{
        ...SCREEN_INNER,
        filter: "saturate(.2) brightness(.7)",
      }}
    >
      <BrandHeader time="23:51" />
      <div style={{ marginTop: 14 }}>
        <BalanceHero amount="¥0" count={47} rate={100} dimmed />
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
            top: "4%",
            left: 0,
            right: 0,
            textAlign: "center",
            fontFamily: "var(--font-display-serif)",
            fontStyle: "italic",
            fontSize: 26,
            color: "var(--color-cream)",
          }}
        >
          今月は、終わりだ。
        </div>
        <GoroButton state="dead" />
      </div>
      <div style={{ filter: "saturate(2) brightness(1.4)", isolation: "isolate" }}>
        <div
          style={{
            background: "linear-gradient(180deg, #ff6464, var(--color-accent-danger))",
            color: "#fff",
            textAlign: "center",
            padding: "11px 0",
            borderRadius: 4,
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 11,
            letterSpacing: "0.1em",
            boxShadow: "0 0 18px rgba(255,79,79,.55)",
          }}
        >
          ¥50,000。来月もこの調子だ。
        </div>
      </div>
    </div>
  );
}

function ModalMock({
  title,
  sub,
  primary,
  secondary,
  showBackground = "idle",
}: {
  title: string;
  sub: string;
  primary?: string;
  secondary?: string;
  showBackground?: "idle" | "suggest";
}) {
  const Bg = showBackground === "suggest" ? MainSuggestMock : MainIdleMock;
  return (
    <div style={{ position: "absolute", inset: 0 }}>
      <div style={{ filter: "blur(4px) brightness(.55)", height: "100%" }}>
        <Bg />
      </div>
      <div
        style={{
          position: "absolute",
          inset: 0,
          background: "rgba(0,0,0,.55)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: 20,
        }}
      >
        <div
          style={{
            width: "100%",
            background:
              "linear-gradient(180deg, var(--color-bg-warm), var(--color-bg-mid))",
            border: "1px solid var(--color-gold-500)",
            borderRadius: 14,
            boxShadow: "0 0 30px rgba(201,169,107,.3)",
            padding: "20px 16px",
            position: "relative",
            overflow: "hidden",
          }}
        >
          <div
            style={{
              position: "absolute",
              inset: 0,
              background: SCAN_LINES,
              pointerEvents: "none",
            }}
          />
          <div
            style={{
              fontFamily: "var(--font-display-serif)",
              fontStyle: "italic",
              fontWeight: 700,
              fontSize: 24,
              color: "var(--color-cream)",
              textAlign: "center",
              position: "relative",
            }}
          >
            {title}
          </div>
          <div
            style={{
              fontFamily: "var(--font-body-mincho)",
              fontSize: 11,
              color: "var(--color-gold-500)",
              textAlign: "center",
              marginTop: 6,
              marginBottom: 18,
              letterSpacing: "0.04em",
              position: "relative",
            }}
          >
            {sub}
          </div>
          <div style={{ display: "flex", gap: 8, position: "relative" }}>
            {secondary ? (
              <div style={{ flex: 1 }}>
                <GoldButton variant="secondary">{secondary}</GoldButton>
              </div>
            ) : null}
            {primary ? (
              <div style={{ flex: 1 }}>
                <GoldButton>{primary}</GoldButton>
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </div>
  );
}

function LogoutModalMock() {
  return <ModalMock title="やめるのか？" sub="戻ってこい。" secondary="戻る。" primary="やめる。" />;
}

function SessionExpiredModalMock() {
  return (
    <ModalMock
      title="離れすぎたな。"
      sub="セッションが切れました。再ログインしてください。"
      primary="戻る。"
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
        minHeight: "100vh",
        background: "var(--color-bg-deep)",
        padding: "48px 32px 80px",
      }}
    >
      <header style={{ maxWidth: 1400, margin: "0 auto 36px" }}>
        <h1
          style={{
            margin: 0,
            fontFamily: "var(--font-body-mincho)",
            fontWeight: 900,
            fontSize: 22,
            color: "var(--color-cream)",
            letterSpacing: "0.04em",
          }}
        >
          <span style={{ fontFamily: "var(--font-body-mincho)" }}>ゴロゴロ</span>
          <span style={{ fontFamily: "var(--font-display-serif)", fontStyle: "italic" }}>Pay</span>{" "}
          - 全画面モック (リヴァイ調コピー統一)
        </h1>
        <p
          style={{
            margin: "8px 0 0",
            color: "var(--color-mute)",
            fontFamily: "var(--font-mono-pixel)",
            fontSize: 12,
            letterSpacing: "0.1em",
          }}
        >
          ランディング / サインアップ / ログイン / 予算セットアップ / メイン (idle, suggest, dead) / 完了 / 予算不足モーダル / ログアウトモーダル / セッション切れモーダル
        </p>
      </header>

      <div
        style={{
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
                fontFamily: "var(--font-mono-pixel)",
                fontSize: 11,
                letterSpacing: "0.18em",
                color: "var(--color-gold-500)",
                marginBottom: 4,
              }}
            >
              {no} / {title}
            </div>
            <div
              style={{
                fontSize: 11,
                color: "var(--color-mute)",
                marginBottom: 14,
                fontFamily: "var(--font-body-sans)",
              }}
            >
              {desc}
            </div>
            <div style={FRAME_STYLE}>
              <Mock />
              <div style={SCANLINE_OVERLAY} />
            </div>
          </section>
        ))}
      </div>
    </main>
  );
}
