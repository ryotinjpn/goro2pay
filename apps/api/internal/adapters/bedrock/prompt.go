package bedrock

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HistoryItem は Bedrock プロンプトに渡す履歴 1 件の縮退表現。
//
// PII を含めない (NFRC-C24 整合): userId / email 等は含めない。
type HistoryItem struct {
	StoreName string `json:"storeName"`
	MenuName  string `json:"menuName"`
	Category  string `json:"category"`
	Amount    int    `json:"amount"`
	DayOfWeek string `json:"dayOfWeek"`
}

// BuildPrompt は Bedrock Converse API に渡すユーザメッセージ本文を生成する。
//
// 出力フォーマットは JSON 1 オブジェクトに固定し、Bedrock 側のテキストパース
// 不整合リスクを下げる (FD §2.2)。
func BuildPrompt(history []HistoryItem, dayOfWeek string, category string) (string, error) {
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return "", fmt.Errorf("marshal history: %w", err)
	}

	var b strings.Builder
	b.WriteString("あなたは「人をダメにする」食事代行アプリのアシスタントです。\n")
	b.WriteString("以下のユーザの直近の注文履歴と現在の曜日を踏まえ、最も「ダメ化を促す」食事プランを 1 つ JSON で提案してください。\n\n")
	b.WriteString("## 履歴\n")
	b.Write(historyJSON)
	b.WriteString("\n\n## 現在の曜日\n")
	b.WriteString(dayOfWeek)
	b.WriteString("\n\n## カテゴリ\n")
	b.WriteString(category)
	b.WriteString("\n\n## 出力フォーマット (JSON のみ、それ以外の文字列は一切含めないこと)\n")
	b.WriteString(`{
  "storeName": "店舗名 (短く)",
  "menuName": "メニュー名",
  "amount": 整数 (円、500〜3000 の範囲),
  "category": "food"
}`)
	return b.String(), nil
}

// BedrockPlanResponse は Bedrock の出力 JSON をパースした構造体。
type BedrockPlanResponse struct {
	StoreName string `json:"storeName"`
	MenuName  string `json:"menuName"`
	Amount    int    `json:"amount"`
	Category  string `json:"category"`
}

// ParsePlanResponse は Bedrock の text 出力を JSON としてパースする。
//
// Bedrock 応答に余計な前後文字列が含まれる場合に備え、最初の '{' から
// 最後の '}' までを抽出してから unmarshal する。
func ParsePlanResponse(text string) (*BedrockPlanResponse, error) {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("no JSON object found in bedrock response")
	}
	jsonPart := text[start : end+1]

	var plan BedrockPlanResponse
	if err := json.Unmarshal([]byte(jsonPart), &plan); err != nil {
		return nil, fmt.Errorf("unmarshal bedrock plan: %w", err)
	}
	if plan.StoreName == "" || plan.MenuName == "" || plan.Amount <= 0 || plan.Category == "" {
		return nil, fmt.Errorf("invalid plan: missing required fields")
	}
	return &plan, nil
}
