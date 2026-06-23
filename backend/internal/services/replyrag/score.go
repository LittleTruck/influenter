package replyrag

import "strings"

// 檢索與排序參數
const (
	// poolLimit 候選池上限（最新的寄出回信 + 優質範例）
	poolLimit = 120
	// defaultTopK 預設注入的 few-shot 範例數量
	defaultTopK = 3
	// maxTopK few-shot 範例數量上限
	maxTopK = 5
	// minCosine 跨品牌候選的最低餘弦相似度門檻；同品牌或優質範例不受此限
	minCosine = 0.15

	// 加權：在語意相似度之上，依使用者指定的「品牌、案件類型、語氣」維度加分
	brandBoost = 0.15 // 同品牌
	typeBoost  = 0.08 // 同合作類型
	goodBoost  = 0.30 // 已標記為優質範例

	// 文字截斷長度（建立向量文件與顯示用）
	incomingDocTruncate = 1200
	replyDocTruncate    = 2000
)

// scoreCandidate 在語意相似度（cosine）之上疊加品牌／類型／優質範例加權。
func scoreCandidate(cosine float64, brandMatch, typeMatch, isGood bool) float64 {
	score := cosine
	if brandMatch {
		score += brandBoost
	}
	if typeMatch {
		score += typeBoost
	}
	if isGood {
		score += goodBoost
	}
	return score
}

// shouldInclude 決定候選是否進入排序：同品牌或優質範例一律納入，
// 其餘須通過最低相似度門檻，避免不相關的回信被當成範例。
func shouldInclude(cosine float64, brandMatch, isGood bool) bool {
	return brandMatch || isGood || cosine >= minCosine
}

// equalFoldTrim 比較兩個字串是否相同（忽略大小寫與前後空白），且皆非空。
func equalFoldTrim(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	return strings.EqualFold(a, b)
}

// truncateRunes 以字元（rune）為單位安全截斷，避免切斷多位元組中文字。
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
