package service

import (
	"context"
	"testing"
	"time"

	"github.com/hammo/influScope/indexer/internal/domain"
	"github.com/hammo/influScope/pkg/models"
)

// --- MOCKS ---

type mockMessage struct {
	body     []byte
	ackCount int
}

func (m *mockMessage) Body() []byte { return m.body }
func (m *mockMessage) Ack() error   { m.ackCount++; return nil }

type mockConsumer struct {
	messages []*mockMessage
	index    int
}

func (m *mockConsumer) Next(ctx context.Context) (domain.Message, error) {
	if m.index >= len(m.messages) {
		// Block forever so the service loop doesn't spin infinitely after reading all messages
		<-ctx.Done()
		return nil, ctx.Err()
	}
	msg := m.messages[m.index]
	m.index++
	return msg, nil
}
func (m *mockConsumer) Close() error { return nil }

type mockAnalytics struct {
	rate float64
	err  error
}

func (m *mockAnalytics) GetEngagement(ctx context.Context, u string, f int, p string) (float64, error) {
	return m.rate, m.err
}
func (m *mockAnalytics) Close() error { return nil }

type mockSearch struct {
	savedCount int
	err        error
}

func (m *mockSearch) IndexProfile(ctx context.Context, profile *models.Influencer) error {
	if m.err != nil {
		return m.err
	}
	m.savedCount++
	return nil
}

type mockMetrics struct {
	indexed int
	errors  int
}

func (m *mockMetrics) IncIndexed() { m.indexed++ }
func (m *mockMetrics) IncError()   { m.errors++ }

// --- TESTS ---

func TestMessageConsumptionFlow(t *testing.T) {
	// 1. Setup Mocks
	msg1 := &mockMessage{body: []byte(`{"username": "user1", "followers": 5000}`)}
	msg2 := &mockMessage{body: []byte(`{"username": "user2", "followers": 100}`)}
	consumer := &mockConsumer{messages: []*mockMessage{msg1, msg2}}
	analytics := &mockAnalytics{rate: 5.5}
	search := &mockSearch{}
	metrics := &mockMetrics{}

	// 2. Init Service
	svc := NewIndexerService(consumer, analytics, search, metrics)

	// 3. Run Service in background briefly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go svc.Start(ctx)
	<-ctx.Done() // Wait for timeout

	// 4. Assertions
	if search.savedCount != 2 {
		t.Errorf("Expected 2 profiles saved to search, got %d", search.savedCount)
	}
	if metrics.indexed != 2 {
		t.Errorf("Expected 2 indexed metrics, got %d", metrics.indexed)
	}
	if msg1.ackCount != 1 || msg2.ackCount != 1 {
		t.Errorf("Expected both messages to be ACKed exactly once")
	}
}

func TestBadJSONMessage(t *testing.T) {
	msg := &mockMessage{body: []byte(`{bad json}`)}
	consumer := &mockConsumer{messages: []*mockMessage{msg}}
	search := &mockSearch{}
	metrics := &mockMetrics{}

	svc := NewIndexerService(consumer, &mockAnalytics{}, search, metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	go svc.Start(ctx)
	<-ctx.Done()

	// Bad JSON should be acked (discarded) but NOT sent to Elasticsearch
	if search.savedCount != 0 {
		t.Errorf("Expected 0 saves, got %d", search.savedCount)
	}
	if metrics.errors != 0 {
		t.Errorf("Expected 0 ES errors, got %d", metrics.errors)
	}
	if msg.ackCount != 1 {
		t.Errorf("Expected bad message to be ACKed (discarded)")
	}
}
