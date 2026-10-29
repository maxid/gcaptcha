package gcaptcha

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func decodePNGSize(t *testing.T, dataURI string) (int, int) {
	t.Helper()
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:image/png;base64") {
		t.Fatalf("invalid data URI prefix: %.40q", dataURI)
	}
	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Width, cfg.Height
}

func newTestService() (*Service, *MemoryStore) {
	store := NewMemoryStore()
	return New(store, WithRand(rand.New(rand.NewSource(7)))), store
}

func solution(t *testing.T, store *MemoryStore, id string) float64 {
	t.Helper()
	store.mu.Lock()
	defer store.mu.Unlock()
	item, ok := store.items[id]
	if !ok {
		t.Fatalf("challenge %s not stored", id)
	}
	return item.data.Percentage
}

func trackAt(width int, percentage float64) Track {
	return Track{
		BgImageWidth: width,
		TrackList:    []Point{{X: 0, Y: 0}, {X: percentage * float64(width), Y: 0}},
	}
}

func TestGenerateAndValidate(t *testing.T) {
	svc, store := newTestService()
	ch, err := svc.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID == "" || ch.Type != TypeSlider {
		t.Fatalf("challenge = %+v", ch)
	}
	if w, h := decodePNGSize(t, ch.BackgroundImage); w != 600 || h != 360 {
		t.Fatalf("background = %dx%d", w, h)
	}
	if w, h := decodePNGSize(t, ch.TemplateImage); w != 110 || h != 110 {
		t.Fatalf("template = %dx%d", w, h)
	}

	pct := solution(t, store, ch.ID)
	ok, err := svc.Validate(context.Background(), ch.ID, trackAt(ch.BackgroundImageWidth, pct))
	if err != nil || !ok {
		t.Fatalf("validate: ok=%v err=%v", ok, err)
	}
	// Single use.
	ok, err = svc.Validate(context.Background(), ch.ID, trackAt(ch.BackgroundImageWidth, pct))
	if err != nil || ok {
		t.Fatalf("second validate: ok=%v err=%v", ok, err)
	}
}

func TestValidateRejectsWrongPosition(t *testing.T) {
	svc, store := newTestService()
	ch, err := svc.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	pct := solution(t, store, ch.ID)
	if ok, err := svc.Validate(context.Background(), ch.ID, trackAt(ch.BackgroundImageWidth, pct+0.1)); err != nil || ok {
		t.Fatalf("wrong position accepted: ok=%v err=%v", ok, err)
	}
}

func TestValidateToleranceBoundary(t *testing.T) {
	svc, store := newTestService()
	ch, err := svc.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	pct := solution(t, store, ch.ID)
	if ok, _ := svc.Validate(context.Background(), ch.ID, trackAt(ch.BackgroundImageWidth, pct+0.019)); !ok {
		t.Fatal("within tolerance should pass")
	}
}

func TestValidateExpired(t *testing.T) {
	store := NewMemoryStore()
	now := time.Unix(1_700_000_000, 0)
	store.SetClock(func() time.Time { return now })
	svc := New(store, WithRand(rand.New(rand.NewSource(7))), WithTTL(20*time.Second))
	ch, err := svc.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(30 * time.Second)
	if ok, _ := svc.Validate(context.Background(), ch.ID, Track{BgImageWidth: 600, TrackList: []Point{{X: 0}, {X: 300}}}); ok {
		t.Fatal("expired challenge accepted")
	}
}

func TestValidateMissingOrMalformed(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	if ok, _ := svc.Validate(ctx, "", trackAt(600, 0.5)); ok {
		t.Fatal("empty id accepted")
	}
	if ok, _ := svc.Validate(ctx, "missing", trackAt(600, 0.5)); ok {
		t.Fatal("missing id accepted")
	}
	ch, err := svc.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := svc.Validate(ctx, ch.ID, Track{BgImageWidth: 600}); ok {
		t.Fatal("empty track accepted")
	}
	if ok, _ := svc.Validate(ctx, ch.ID, Track{BgImageWidth: 0, TrackList: []Point{{X: 0}, {X: 1}}}); ok {
		t.Fatal("zero width accepted")
	}
}

func TestGenerateProducesUniqueIDs(t *testing.T) {
	svc, _ := newTestService()
	a, err := svc.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.Generate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if a.ID == b.ID {
		t.Fatal("duplicate challenge id")
	}
}
