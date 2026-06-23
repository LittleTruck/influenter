package openai

import (
	"context"
	"fmt"
	"math"

	openai "github.com/sashabaranov/go-openai"
)

// defaultEmbeddingModel 預設使用的 embedding 模型（成本低、效果佳）
const defaultEmbeddingModel = "text-embedding-3-small"

// embeddingBatchSize 單次 API 呼叫最多嵌入的文字數，避免超過請求上限
const embeddingBatchSize = 96

// EmbeddingModel 取得目前使用的 embedding 模型名稱
func (s *Service) EmbeddingModel() string {
	if s.config.EmbeddingModel != "" {
		return s.config.EmbeddingModel
	}
	return defaultEmbeddingModel
}

// EmbeddingDimensions 取得 embedding 維度（0 表示使用模型預設維度）
func (s *Service) EmbeddingDimensions() int {
	return s.config.EmbeddingDimensions
}

// CreateEmbeddings 將多筆文字批次轉為向量，回傳順序與輸入一致。
// 內部會自動分批呼叫 API，以避免單次請求過大。
func (s *Service) CreateEmbeddings(ctx context.Context, inputs []string) ([][]float64, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	model := s.EmbeddingModel()
	dimensions := s.EmbeddingDimensions()

	out := make([][]float64, 0, len(inputs))
	for start := 0; start < len(inputs); start += embeddingBatchSize {
		end := start + embeddingBatchSize
		if end > len(inputs) {
			end = len(inputs)
		}
		batch := inputs[start:end]

		req := openai.EmbeddingRequest{
			Input: batch,
			Model: openai.EmbeddingModel(model),
		}
		if dimensions > 0 {
			req.Dimensions = dimensions
		}

		resp, err := s.client.CreateEmbeddings(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create embeddings: %w", err)
		}
		if len(resp.Data) != len(batch) {
			return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(resp.Data), len(batch))
		}

		for _, d := range resp.Data {
			vec := make([]float64, len(d.Embedding))
			for j, f := range d.Embedding {
				vec[j] = float64(f)
			}
			out = append(out, vec)
		}
	}

	return out, nil
}

// CreateEmbedding 將單筆文字轉為向量。
func (s *Service) CreateEmbedding(ctx context.Context, input string) ([]float64, error) {
	vecs, err := s.CreateEmbeddings(ctx, []string{input})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return vecs[0], nil
}

// CosineSimilarity 計算兩個向量的餘弦相似度，範圍約為 [-1, 1]。
// 長度不一致或任一向量為零向量時回傳 0。
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
