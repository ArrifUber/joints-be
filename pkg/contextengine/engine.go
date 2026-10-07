package contextengine

import "context"

// ContextInput adalah data yang dikirim ke AI engine untuk dianalisis.
type ContextInput struct {
	Subject           string
	Topic             string
	Subtopic          string
	CurrentContext    *ContextResult // nil jika belum ada konteks sebelumnya
	RecentTranscripts []string
}

// ContextResult adalah hasil analisis dari AI engine.
type ContextResult struct {
	// Type menentukan jenis update:
	// "concept_update", "topic_changed", "formula_detected", "no_change"
	Type     string   `json:"type"`
	Concept  string   `json:"concept,omitempty"`
	Summary  string   `json:"summary,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Formula  string   `json:"formula,omitempty"`
}

// ContextEngine adalah abstraction untuk AI provider.
// Implementasi konkret dapat berupa Mock, Gemini, OpenAI, dsb.
// Business logic tidak boleh bergantung pada implementasi spesifik —
// hanya pada interface ini.
type ContextEngine interface {
	Analyze(ctx context.Context, input ContextInput) (*ContextResult, error)
}

