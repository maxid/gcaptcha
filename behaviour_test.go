package gcaptcha

import (
	"context"
	"testing"
	"time"
)

func newService(t *testing.T, opts ...Option) *Service {
	t.Helper()
	return New(NewMemoryStore(), opts...)
}

// peekPercentage reads the stored solution without consuming the challenge: Take is
// one-shot, so a test that peeks must put the data back.
func peekPercentage(t *testing.T, svc *Service, id string) float64 {
	t.Helper()
	store := svc.store.(*MemoryStore)
	data, ok, err := store.Take(context.Background(), id)
	if err != nil || !ok {
		t.Fatalf("peek: ok=%v err=%v", ok, err)
	}
	if err := store.Save(context.Background(), id, data, time.Minute); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return data.Percentage
}

// A track that reaches the slot passes with the contract defaults: the Java
// validator only compares the drop position.
func TestValidatePositionOnlyByDefault(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	stored := peekPercentage(t, svc, challenge.ID)

	track := Track{
		BgImageWidth: 600,
		StartTime:    1_700_000_000_000,
		StopTime:     1_700_000_000_010, // 10ms: too fast for the optional policy
		TrackList: []Point{
			{X: 0, Y: 0, T: 0},
			{X: stored * 600, Y: 0, T: 10},
		},
	}
	if ok, err := svc.Validate(ctx, challenge.ID, track); err != nil || !ok {
		t.Fatalf("position-only validation = %v, %v", ok, err)
	}
}

// Enabling the policy rejects a scripted submission the position check would accept.
func TestTrackPolicyRejectsTooFewPoints(t *testing.T) {
	svc := newService(t, WithTrackPolicy(DefaultTrackPolicy()))
	ctx := context.Background()

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	percentage := peekPercentage(t, svc, challenge.ID)

	track := Track{
		BgImageWidth: 600,
		StartTime:    1_700_000_000_000,
		StopTime:     1_700_000_000_500,
		TrackList:    []Point{{X: percentage * 600, Y: 0}}, // one point
	}
	if ok, _ := svc.Validate(ctx, challenge.ID, track); ok {
		t.Fatal("a single-point track must be rejected when the policy is on")
	}
}

func TestTrackPolicyRejectsInstantDrag(t *testing.T) {
	svc := newService(t, WithTrackPolicy(DefaultTrackPolicy()))
	ctx := context.Background()

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	percentage := peekPercentage(t, svc, challenge.ID)

	track := Track{
		BgImageWidth: 600,
		StartTime:    1_700_000_000_000,
		StopTime:     1_700_000_000_000, // zero elapsed
		TrackList: []Point{
			{X: 0, Y: 0},
			{X: percentage * 600 / 2, Y: 0},
			{X: percentage * 600, Y: 0},
		},
	}
	if ok, _ := svc.Validate(ctx, challenge.ID, track); ok {
		t.Fatal("a zero-duration drag must be rejected when the policy is on")
	}
}

func TestTrackPolicyAcceptsAPlausibleDrag(t *testing.T) {
	svc := newService(t, WithTrackPolicy(DefaultTrackPolicy()))
	ctx := context.Background()

	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	target := peekPercentage(t, svc, challenge.ID) * 600

	track := Track{
		BgImageWidth: 600,
		StartTime:    1_700_000_000_000,
		StopTime:     1_700_000_000_600,
		Top:          100,
		TrackList: []Point{
			{X: 0, Y: 100},
			{X: target / 2, Y: 101},
			{X: target, Y: 100},
		},
	}
	if ok, err := svc.Validate(ctx, challenge.ID, track); err != nil || !ok {
		t.Fatalf("plausible drag rejected: %v, %v", ok, err)
	}
}

// Secondary verification keeps the challenge longer, like the Java mode.
func TestSecondaryVerificationExtendsTTL(t *testing.T) {
	store := NewMemoryStore()
	now := time.Unix(1_700_000_000, 0)
	store.SetClock(func() time.Time { return now })
	svc := New(store, WithSecondaryVerification(true))

	ctx := context.Background()
	challenge, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Past the normal 20s lifetime, still inside the 120s secondary window.
	now = now.Add(60 * time.Second)
	data, ok, _ := store.Take(ctx, "secondary:"+challenge.ID)
	if !ok {
		t.Fatal("secondary challenge not stored under the secondary key")
	}
	if data.Type != TypeSlider {
		t.Fatalf("type = %q", data.Type)
	}
}

func TestChallengeIsSingleUse(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	challenge, _ := svc.Generate(ctx)
	track := Track{BgImageWidth: 600, TrackList: []Point{{X: 0}, {X: 300}}}

	_, _ = svc.Validate(ctx, challenge.ID, track)
	if ok, _ := svc.Validate(ctx, challenge.ID, track); ok {
		t.Fatal("a challenge must not validate twice")
	}
}
