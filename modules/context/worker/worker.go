package worker

import (
	"context"
	"log"
	"time"

	"joints-be/config"
	contextService "joints-be/modules/context/service"
	transcriptEntity "joints-be/modules/transcript/entity"
	transcriptRepo "joints-be/modules/transcript/repository"
)

type ContextWorker struct {
	contextSvc     contextService.ContextService
	transcriptRepo transcriptRepo.TranscriptRepository
	interval       time.Duration
	recentChunks   int
}

func NewContextWorker(
	contextSvc contextService.ContextService,
	transcriptRepo transcriptRepo.TranscriptRepository,
) *ContextWorker {
	cfg := config.AppConfig
	interval := 1 * time.Second
	recentChunks := 5
	if cfg != nil {
		if cfg.ContextWorkerInterval > 0 {
			interval = cfg.ContextWorkerInterval
		}
		if cfg.ContextRecentChunks > 0 {
			recentChunks = cfg.ContextRecentChunks
		}
	}

	return &ContextWorker{
		contextSvc:     contextSvc,
		transcriptRepo: transcriptRepo,
		interval:       interval,
		recentChunks:   recentChunks,
	}
}

// Start menjalankan background worker loop sampai context di-cancel (graceful shutdown).
func (w *ContextWorker) Start(ctx context.Context) {
	log.Printf("[worker] context worker started (interval=%v, recentChunks=%d)", w.interval, w.recentChunks)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[worker] context worker stopping gracefully...")
			return
		case <-ticker.C:
			w.processPendingTranscripts(ctx)
		}
	}
}

func (w *ContextWorker) processPendingTranscripts(ctx context.Context) {
	// Ambil pending chunks
	chunks, err := w.transcriptRepo.FindPending(ctx, 10)
	if err != nil {
		log.Printf("[worker] failed to fetch pending transcripts: %v", err)
		return
	}

	for _, chunk := range chunks {
		// Stop jika shutdown signal diterima
		if ctx.Err() != nil {
			return
		}
		w.processSingleTranscript(ctx, chunk)
	}
}

func (w *ContextWorker) processSingleTranscript(ctx context.Context, chunk transcriptEntity.TranscriptChunk) {
	startTime := time.Now()

	// 1. Mark as processing to protect against duplicate processing
	if err := w.transcriptRepo.UpdateStatus(ctx, chunk.ID, transcriptEntity.TranscriptStatusProcessing); err != nil {
		log.Printf("[worker] failed to mark transcript as processing id=%s error=%v", chunk.ID, err)
		return
	}

	// 2. Fetch recent transcript chunks up to current sequence
	recents, err := w.transcriptRepo.FindRecentBySession(ctx, chunk.SessionID, chunk.Sequence, w.recentChunks)
	if err != nil {
		log.Printf("[worker] failed to fetch recent transcripts session_id=%s sequence=%d error=%v", chunk.SessionID, chunk.Sequence, err)
		_ = w.transcriptRepo.UpdateStatus(ctx, chunk.ID, transcriptEntity.TranscriptStatusFailed)
		return
	}

	texts := make([]string, len(recents))
	for i, r := range recents {
		texts[i] = r.Text
	}

	// 3. Process context via ContextService
	ctxResult, err := w.contextSvc.ProcessTranscript(ctx, chunk.SessionID, chunk.ID, chunk.Sequence, texts)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[worker] transcript processing failed session_id=%s transcript_id=%s sequence=%d duration_ms=%d error=%v",
			chunk.SessionID, chunk.ID, chunk.Sequence, duration.Milliseconds(), err)
		_ = w.transcriptRepo.UpdateStatus(ctx, chunk.ID, transcriptEntity.TranscriptStatusFailed)
		return
	}

	// 4. Mark as completed
	if err := w.transcriptRepo.UpdateStatus(ctx, chunk.ID, transcriptEntity.TranscriptStatusCompleted); err != nil {
		log.Printf("[worker] failed to mark transcript as completed id=%s error=%v", chunk.ID, err)
		return
	}

	contextType := "no_change"
	if ctxResult != nil {
		contextType = ctxResult.Type
	}

	log.Printf("[worker] transcript processed successfully session_id=%s transcript_id=%s sequence=%d duration_ms=%d context_type=%s",
		chunk.SessionID, chunk.ID, chunk.Sequence, duration.Milliseconds(), contextType)
}

