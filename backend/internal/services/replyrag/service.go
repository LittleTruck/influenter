// Package replyrag 實作擬回信的 few-shot 檢索（RAG）：
// 從使用者過去寄出、且關聯案件的回信中，依「品牌、案件類型、語氣（語意相似度）」
// 找出最相似的數封，作為產生草稿時的範例。向量相似度以應用層 cosine 計算，
// 不依賴 pgvector；embedding 會延遲產生並快取於 emails 資料表。
package replyrag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/services/openai"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Service 擬回信 few-shot 檢索服務
type Service struct {
	db     *gorm.DB
	ai     *openai.Service
	logger *zerolog.Logger
}

// New 建立檢索服務
func New(db *gorm.DB, ai *openai.Service, logger *zerolog.Logger) *Service {
	return &Service{db: db, ai: ai, logger: logger}
}

// Query 檢索條件（本次要回覆的來信 / 案件情境）
type Query struct {
	UserID            uuid.UUID
	BrandName         string
	CollaborationType string
	IncomingSubject   string
	IncomingBody      string
	TopK              int // 0 表示使用預設值
}

// candidateRow 候選回信（寄出且關聯案件）
type candidateRow struct {
	EmailID              uuid.UUID       `gorm:"column:email_id"`
	ThreadID             *string         `gorm:"column:thread_id"`
	Subject              *string         `gorm:"column:subject"`
	BodyText             *string         `gorm:"column:body_text"`
	ReceivedAt           time.Time       `gorm:"column:received_at"`
	IsGoodExample        bool            `gorm:"column:is_good_example"`
	ContextEmbedding     pq.Float64Array `gorm:"column:context_embedding"`
	ContextEmbeddingHash *string         `gorm:"column:context_embedding_hash"`
	BrandName            string          `gorm:"column:brand_name"`
	CollaborationType    *string         `gorm:"column:collaboration_type"`
}

// incomingRow 候選回信所屬郵件串中的來信
type incomingRow struct {
	ThreadID   string    `gorm:"column:thread_id"`
	Subject    *string   `gorm:"column:subject"`
	BodyText   *string   `gorm:"column:body_text"`
	Snippet    *string   `gorm:"column:snippet"`
	ReceivedAt time.Time `gorm:"column:received_at"`
}

// RetrieveExamples 回傳與本次來信最相似的過往回信，作為 few-shot 範例。
// 任何錯誤都會回傳給呼叫端，由呼叫端決定是否 fail-open（不附範例照常擬信）。
func (s *Service) RetrieveExamples(ctx context.Context, q Query) ([]openai.DraftReplyExample, error) {
	if s.ai == nil || s.db == nil {
		return nil, nil
	}

	topK := q.TopK
	if topK <= 0 {
		topK = defaultTopK
	}
	if topK > maxTopK {
		topK = maxTopK
	}

	// 1. 撈出候選回信（優質範例優先、其次依時間）
	var candidates []candidateRow
	err := s.db.WithContext(ctx).
		Table("emails AS e").
		Select(`e.id AS email_id, e.thread_id, e.subject, e.body_text, e.received_at,
			e.is_good_example, e.context_embedding, e.context_embedding_hash,
			c.brand_name, c.collaboration_type`).
		Joins("JOIN oauth_accounts oa ON oa.id = e.oauth_account_id").
		Joins("JOIN cases c ON c.id = e.case_id AND c.deleted_at IS NULL").
		Where(`oa.user_id = ? AND e.direction = ? AND e.deleted_at IS NULL
			AND e.body_text IS NOT NULL AND e.body_text <> ''`,
			q.UserID, models.EmailDirectionOutgoing).
		Order("e.is_good_example DESC, e.received_at DESC").
		Limit(poolLimit).
		Scan(&candidates).Error
	if err != nil {
		return nil, fmt.Errorf("load candidates: %w", err)
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// 2. 配對每封回信所回應的「來信」（同一 thread 中、時間早於回信的最後一封來信）
	triggers := s.loadTriggerEmails(ctx, q.UserID, candidates)

	// 3. 確保每個候選都有 context embedding（缺少或過期則重新產生並快取）
	model := s.ai.EmbeddingModel()
	dims := s.ai.EmbeddingDimensions()

	type scored struct {
		cand      candidateRow
		trigger   *incomingRow
		embedding []float64
		doc       string
		hash      string
	}
	items := make([]*scored, 0, len(candidates))
	var toEmbedIdx []int
	var toEmbedDocs []string

	for i := range candidates {
		c := candidates[i]
		var trig *incomingRow
		if t, ok := triggers[c.EmailID]; ok {
			trig = t
		}
		doc := buildCandidateDoc(c, trig)
		hash := docHash(model, dims, doc)

		it := &scored{cand: c, trigger: trig, doc: doc, hash: hash}

		// 重用快取向量的條件：hash 相符（已含 model+維度）、非空，且若有指定維度則長度需相符。
		// docHash 已涵蓋維度變更，這裡再以長度作為防禦性檢查，避免極端情況下使用到舊維度向量。
		cached := []float64(c.ContextEmbedding)
		if c.ContextEmbeddingHash != nil && *c.ContextEmbeddingHash == hash && len(cached) > 0 && (dims <= 0 || len(cached) == dims) {
			it.embedding = cached
		} else {
			toEmbedIdx = append(toEmbedIdx, len(items))
			toEmbedDocs = append(toEmbedDocs, doc)
		}
		items = append(items, it)
	}

	if len(toEmbedDocs) > 0 {
		vecs, err := s.ai.CreateEmbeddings(ctx, toEmbedDocs)
		if err != nil {
			return nil, fmt.Errorf("embed candidates: %w", err)
		}
		if len(vecs) != len(toEmbedIdx) {
			return nil, fmt.Errorf("candidate embedding count mismatch: got %d, want %d", len(vecs), len(toEmbedIdx))
		}
		for k, idx := range toEmbedIdx {
			items[idx].embedding = vecs[k]
			s.persistEmbedding(ctx, items[idx].cand.EmailID, vecs[k], model, items[idx].hash)
		}
	}

	// 4. 計算本次來信情境的向量
	queryDoc := buildQueryDoc(q)
	queryEmb, err := s.ai.CreateEmbedding(ctx, queryDoc)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	// 5. 評分並排序
	type ranked struct {
		it    *scored
		score float64
	}
	var rankedItems []ranked
	for _, it := range items {
		if len(it.embedding) == 0 {
			continue
		}
		cos := openai.CosineSimilarity(queryEmb, it.embedding)
		brandMatch := equalFoldTrim(it.cand.BrandName, q.BrandName)
		typeMatch := equalFoldTrim(derefStr(it.cand.CollaborationType), q.CollaborationType)
		if !shouldInclude(cos, brandMatch, it.cand.IsGoodExample) {
			continue
		}
		rankedItems = append(rankedItems, ranked{
			it:    it,
			score: scoreCandidate(cos, brandMatch, typeMatch, it.cand.IsGoodExample),
		})
	}

	sort.SliceStable(rankedItems, func(i, j int) bool {
		return rankedItems[i].score > rankedItems[j].score
	})

	if len(rankedItems) > topK {
		rankedItems = rankedItems[:topK]
	}

	examples := make([]openai.DraftReplyExample, 0, len(rankedItems))
	for _, r := range rankedItems {
		ex := openai.DraftReplyExample{
			BrandName:         r.it.cand.BrandName,
			CollaborationType: derefStr(r.it.cand.CollaborationType),
			ReplyBody:         truncateRunes(derefStr(r.it.cand.BodyText), replyDocTruncate),
			Similarity:        r.score,
		}
		if r.it.trigger != nil {
			ex.IncomingSubject = derefStr(r.it.trigger.Subject)
			ex.IncomingBody = truncateRunes(incomingText(r.it.trigger), incomingDocTruncate)
		}
		examples = append(examples, ex)
	}

	s.logger.Info().
		Int("candidates", len(candidates)).
		Int("embedded", len(toEmbedDocs)).
		Int("returned", len(examples)).
		Msg("Reply RAG retrieval completed")

	return examples, nil
}

// loadTriggerEmails 為每封候選回信找出其所回應的來信。
func (s *Service) loadTriggerEmails(ctx context.Context, userID uuid.UUID, candidates []candidateRow) map[uuid.UUID]*incomingRow {
	result := make(map[uuid.UUID]*incomingRow)

	threadSet := make(map[string]struct{})
	for _, c := range candidates {
		if c.ThreadID != nil && *c.ThreadID != "" {
			threadSet[*c.ThreadID] = struct{}{}
		}
	}
	if len(threadSet) == 0 {
		return result
	}
	threadIDs := make([]string, 0, len(threadSet))
	for t := range threadSet {
		threadIDs = append(threadIDs, t)
	}

	var incomings []incomingRow
	err := s.db.WithContext(ctx).
		Table("emails AS e").
		Select("e.thread_id, e.subject, e.body_text, e.snippet, e.received_at").
		Joins("JOIN oauth_accounts oa ON oa.id = e.oauth_account_id").
		Where(`oa.user_id = ? AND e.direction = ? AND e.deleted_at IS NULL AND e.thread_id IN ?`,
			userID, models.EmailDirectionIncoming, threadIDs).
		Order("e.received_at ASC").
		Scan(&incomings).Error
	if err != nil {
		s.logger.Warn().Err(err).Msg("Reply RAG: failed to load trigger emails; continuing without pairing")
		return result
	}

	byThread := make(map[string][]incomingRow)
	for _, in := range incomings {
		byThread[in.ThreadID] = append(byThread[in.ThreadID], in)
	}

	for _, c := range candidates {
		if c.ThreadID == nil || *c.ThreadID == "" {
			continue
		}
		list := byThread[*c.ThreadID]
		var best *incomingRow
		for i := range list {
			// 取時間早於（或等於）回信、且最接近回信的那封來信
			if !list[i].ReceivedAt.After(c.ReceivedAt) {
				best = &list[i]
			}
		}
		if best != nil {
			result[c.EmailID] = best
		}
	}

	return result
}

// persistEmbedding 將新算出的向量快取回 emails 資料表（失敗僅記錄，不影響本次檢索）。
func (s *Service) persistEmbedding(ctx context.Context, emailID uuid.UUID, vec []float64, model, hash string) {
	err := s.db.WithContext(ctx).
		Model(&models.Email{}).
		Where("id = ?", emailID).
		Updates(map[string]interface{}{
			"context_embedding":       pq.Float64Array(vec),
			"context_embedding_model": model,
			"context_embedding_hash":  hash,
		}).Error
	if err != nil {
		s.logger.Warn().Err(err).Str("email_id", emailID.String()).Msg("Reply RAG: failed to cache embedding")
	}
}

// buildCandidateDoc 組出候選回信「所回應之情境」的向量文件。
// 優先使用觸發來信內容；若無法配對，退而以回信本身的主旨／內容作為情境代理。
func buildCandidateDoc(c candidateRow, trig *incomingRow) string {
	if trig != nil {
		return composeDoc(c.BrandName, derefStr(c.CollaborationType), derefStr(trig.Subject), incomingText(trig))
	}
	return composeDoc(c.BrandName, derefStr(c.CollaborationType), derefStr(c.Subject), derefStr(c.BodyText))
}

// buildQueryDoc 組出本次來信情境的向量文件（須與 candidate doc 結構一致）。
func buildQueryDoc(q Query) string {
	return composeDoc(q.BrandName, q.CollaborationType, q.IncomingSubject, q.IncomingBody)
}

func composeDoc(brand, collabType, subject, body string) string {
	parts := make([]string, 0, 4)
	if s := strings.TrimSpace(brand); s != "" {
		parts = append(parts, "品牌："+s)
	}
	if s := strings.TrimSpace(collabType); s != "" {
		parts = append(parts, "合作類型："+s)
	}
	if s := strings.TrimSpace(subject); s != "" {
		parts = append(parts, "主旨："+s)
	}
	if s := strings.TrimSpace(body); s != "" {
		parts = append(parts, "內容："+truncateRunes(s, incomingDocTruncate))
	}
	return strings.Join(parts, "\n")
}

// incomingText 取來信內文（優先 body_text，退而 snippet）。
func incomingText(in *incomingRow) string {
	if in.BodyText != nil && strings.TrimSpace(*in.BodyText) != "" {
		return *in.BodyText
	}
	return derefStr(in.Snippet)
}

// docHash 以 model + 維度 + 文件內容產生雜湊，用於判斷快取是否過期。
func docHash(model string, dims int, doc string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%s", model, dims, doc)))
	return hex.EncodeToString(sum[:])
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
