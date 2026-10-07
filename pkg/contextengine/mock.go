package contextengine

import (
	"context"
	"fmt"
	"strings"
)

// MockContextEngine adalah implementasi deterministic untuk development & testing.
// Tidak memerlukan API key. Seluruh pipeline dapat diuji tanpa koneksi ke AI provider.
type MockContextEngine struct{}

func NewMockContextEngine() ContextEngine {
	return &MockContextEngine{}
}

func (m *MockContextEngine) Analyze(ctx context.Context, input ContextInput) (*ContextResult, error) {
	// Simulasi "no_change" jika tidak ada transcript
	if len(input.RecentTranscripts) == 0 {
		return &ContextResult{Type: "no_change"}, nil
	}

	combined := strings.Join(input.RecentTranscripts, " ")

	// Deteksi sederhana berdasarkan kata kunci untuk membuat mock lebih realistis
	lowerCombined := strings.ToLower(combined)

	switch {
	case containsAny(lowerCombined, "rumus", "formula", "persamaan", "f =", "e ="):
		return &ContextResult{
			Type:     "formula_detected",
			Concept:  fmt.Sprintf("Rumus dalam %s", input.Subtopic),
			Summary:  fmt.Sprintf("Mock: ditemukan formula pada topik %s.", input.Topic),
			Keywords: extractKeywords(combined),
			Formula:  "F = m × a (mock)",
		}, nil

	case containsAny(lowerCombined, "hukum", "teori", "prinsip", "konsep"):
		return &ContextResult{
			Type:     "concept_update",
			Concept:  fmt.Sprintf("Konsep dalam %s", input.Subtopic),
			Summary:  fmt.Sprintf("Mock: konsep baru terdeteksi pada konteks %s - %s.", input.Topic, input.Subtopic),
			Keywords: extractKeywords(combined),
		}, nil

	default:
		return &ContextResult{
			Type:     "concept_update",
			Concept:  input.Subtopic,
			Summary:  fmt.Sprintf("Mock context response untuk: %.100s", combined),
			Keywords: extractKeywords(combined),
		}, nil
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

// extractKeywords mengambil kata-kata unik yang panjangnya > 3 huruf sebagai keywords mock.
func extractKeywords(text string) []string {
	words := strings.Fields(text)
	seen := make(map[string]bool)
	var keywords []string
	for _, w := range words {
		w = strings.ToLower(strings.Trim(w, ".,!?;:\"'"))
		if len(w) > 3 && !seen[w] {
			seen[w] = true
			keywords = append(keywords, w)
			if len(keywords) >= 5 {
				break
			}
		}
	}
	return keywords
}

