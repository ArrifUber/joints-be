package tests_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	contextEntity "joints-be/modules/context/entity"
	contextService "joints-be/modules/context/service"
	contextWorker "joints-be/modules/context/worker"
	sessionEntity "joints-be/modules/session/entity"
	transcriptDto "joints-be/modules/transcript/dto"
	transcriptEntity "joints-be/modules/transcript/entity"
	transcriptRepo "joints-be/modules/transcript/repository"
	transcriptService "joints-be/modules/transcript/service"
	"joints-be/pkg/contextengine"
	"joints-be/pkg/event"
)

// In-memory repositories untuk full pipeline testing
type memorySessionRepo struct {
	sessions map[string]*sessionEntity.Session
}

func (m *memorySessionRepo) Create(ctx context.Context, s *sessionEntity.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *memorySessionRepo) GetByID(ctx context.Context, id string) (*sessionEntity.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return s, nil
}

type memoryTranscriptRepo struct {
	mu     sync.Mutex
	chunks []*transcriptEntity.TranscriptChunk
}

func (m *memoryTranscriptRepo) Create(ctx context.Context, t *transcriptEntity.TranscriptChunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.chunks {
		if c.SessionID == t.SessionID && c.Sequence == t.Sequence {
			return transcriptRepo.ErrDuplicateSequence
		}
	}
	t.ID = fmt.Sprintf("chunk-%d", len(m.chunks)+1)
	m.chunks = append(m.chunks, t)
	return nil
}

func (m *memoryTranscriptRepo) FindPending(ctx context.Context, limit int) ([]transcriptEntity.TranscriptChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []transcriptEntity.TranscriptChunk
	for _, c := range m.chunks {
		if c.Status == transcriptEntity.TranscriptStatusPending {
			res = append(res, *c)
			if len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

func (m *memoryTranscriptRepo) FindBySessionID(ctx context.Context, sessionID string) ([]transcriptEntity.TranscriptChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []transcriptEntity.TranscriptChunk
	for _, c := range m.chunks {
		if c.SessionID == sessionID {
			res = append(res, *c)
		}
	}
	return res, nil
}

func (m *memoryTranscriptRepo) FindRecentBySession(ctx context.Context, sessionID string, upToSequence int64, limit int) ([]transcriptEntity.TranscriptChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []transcriptEntity.TranscriptChunk
	for _, c := range m.chunks {
		if c.SessionID == sessionID && c.Sequence <= upToSequence {
			res = append(res, *c)
		}
	}
	if len(res) > limit {
		res = res[len(res)-limit:]
	}
	return res, nil
}

func (m *memoryTranscriptRepo) UpdateStatus(ctx context.Context, id string, status transcriptEntity.TranscriptStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.chunks {
		if c.ID == id {
			c.Status = status
			return nil
		}
	}
	return nil
}

type memoryContextRepo struct {
	mu     sync.Mutex
	chunks []*contextEntity.ContextChunk
}

func (m *memoryContextRepo) Create(ctx context.Context, c *contextEntity.ContextChunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.ID = fmt.Sprintf("ctx-%d", len(m.chunks)+1)
	c.CreatedAt = time.Now()
	m.chunks = append(m.chunks, c)
	return nil
}

func (m *memoryContextRepo) FindBySessionID(ctx context.Context, sessionID string) ([]contextEntity.ContextChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []contextEntity.ContextChunk
	for _, c := range m.chunks {
		if c.SessionID == sessionID {
			res = append(res, *c)
		}
	}
	return res, nil
}

func (m *memoryContextRepo) FindLatestBySessionID(ctx context.Context, sessionID string) (*contextEntity.ContextChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.chunks) - 1; i >= 0; i-- {
		if m.chunks[i].SessionID == sessionID {
			return m.chunks[i], nil
		}
	}
	return nil, fmt.Errorf("no context found")
}

type channelEventPublisher struct {
	events chan event.Event
}

func (p *channelEventPublisher) Publish(ctx context.Context, sessionID string, ev event.Event) error {
	p.events <- ev
	return nil
}

// TestPipeline_Scenario1_2_3 menguji end-to-end pipeline (§38 Skenario 1, 2, 3)
func TestPipeline_Scenario1_2_3(t *testing.T) {
	sessionRepo := &memorySessionRepo{sessions: make(map[string]*sessionEntity.Session)}
	transRepo := &memoryTranscriptRepo{chunks: make([]*transcriptEntity.TranscriptChunk, 0)}
	ctxRepo := &memoryContextRepo{chunks: make([]*contextEntity.ContextChunk, 0)}
	eventPub := &channelEventPublisher{events: make(chan event.Event, 100)}
	engine := contextengine.NewMockContextEngine()

	// Services & Worker
	tSvc := transcriptService.NewTranscriptService(transRepo, sessionRepo)
	cSvc := contextService.NewContextService(ctxRepo, sessionRepo, transRepo, engine, eventPub)
	worker := contextWorker.NewContextWorker(cSvc, transRepo)

	// Step 1: Create Session
	sessionID := "session-fisika-101"
	_ = sessionRepo.Create(context.Background(), &sessionEntity.Session{
		ID:       sessionID,
		Subject:  "Fisika",
		Topic:    "Dinamika",
		Subtopic: "Hukum Newton",
		Status:   sessionEntity.SessionStatusActive,
	})

	// Step 2: Ingest 5 chunks
	texts := []string{
		"Halo semuanya, hari ini kita mulai belajar Dinamika.",
		"Kita akan masuk ke topik Hukum Newton.",
		"Hukum Pertama Newton membahas kelembaman benda.",
		"Hukum Kedua Newton menyatakan rumus F = m x a.",
		"Hukum Ketiga Newton membahas aksi dan reaksi.",
	}

	for i, txt := range texts {
		resp, err := tSvc.AcceptTranscript(context.Background(), sessionID, transcriptDto.CreateTranscriptRequest{
			Sequence: int64(i + 1),
			Text:     txt,
		})
		if err != nil {
			t.Fatalf("failed to accept transcript %d: %v", i+1, err)
		}
		if !resp.Accepted {
			t.Fatalf("expected accepted=true for sequence %d", i+1)
		}
	}

	// Step 3: Run worker to process all pending chunks
	workerCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start worker loop
	go worker.Start(workerCtx)

	// Step 4: Verify events received via publisher (Scenario 3)
	receivedEvents := 0
	for receivedEvents < 5 {
		select {
		case ev := <-eventPub.events:
			if ev.SessionID != sessionID {
				t.Errorf("expected session ID %s, got %s", sessionID, ev.SessionID)
			}
			receivedEvents++
		case <-time.After(1500 * time.Millisecond):
			t.Fatalf("timed out waiting for events, received %d of 5", receivedEvents)
		}
	}

	// Step 5: Verify all chunks completed
	chunks, _ := transRepo.FindBySessionID(context.Background(), sessionID)
	for _, c := range chunks {
		if c.Status != transcriptEntity.TranscriptStatusCompleted {
			t.Errorf("chunk seq %d status = %s, expected completed", c.Sequence, c.Status)
		}
	}

	// Step 6: Verify context chunks saved (Scenario 1 & 2)
	ctxChunks, _ := ctxRepo.FindBySessionID(context.Background(), sessionID)
	if len(ctxChunks) == 0 {
		t.Fatalf("expected context chunks to be created")
	}
}

