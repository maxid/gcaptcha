package gcaptcha

import (
	"math"
	"time"
)

// The protocol document records that the Java validator only checks the drop
// position, not the drag behaviour: `SimpleImageCaptchaValidator` compares the
// percentage with a 0.02 tolerance and ignores the track. A Go implementation that
// silently rejected human-looking-but-slightly-jittery tracks would therefore be
// *stricter* than production and break the front end.
//
// These checks are opt-in for that reason. Enabling them changes observable
// behaviour, so it must be a deliberate decision by the deployment.

// TrackPolicy turns the behavioural checks on.
type TrackPolicy struct {
	// MinPoints rejects tracks that are too short to be a real drag.
	MinPoints int
	// MinDuration rejects instant "drags"; a human takes longer.
	MinDuration time.Duration
	// MaxDuration rejects a drag that ran far past the challenge lifetime.
	MaxDuration time.Duration
	// MaxJitter bounds the vertical wobble in pixels relative to the drag height.
	MaxJitter float64
	// MaxY is the largest allowed vertical offset of the piece.
	MaxY int
}

// DefaultTrackPolicy is a permissive starting point: it rejects obviously scripted
// submissions (single point, zero elapsed time) without judging hand tremor.
func DefaultTrackPolicy() TrackPolicy {
	return TrackPolicy{
		MinPoints:   3,
		MinDuration: 50 * time.Millisecond,
		MaxDuration: 2 * time.Minute,
		MaxJitter:   60,
		MaxY:        360,
	}
}

// WithTrackPolicy enables behavioural verification with the given policy.
func WithTrackPolicy(policy TrackPolicy) Option {
	return func(s *Service) {
		s.trackPolicy = &policy
	}
}

// WithSecondaryVerification extends the challenge lifetime and prefixes the stored
// id, mirroring the Java `secondary` mode (TTL 120s, `captcha:secondary` prefix).
// It is off by default, as in production.
func WithSecondaryVerification(enabled bool) Option {
	return func(s *Service) { s.secondary = enabled }
}

// secondaryTTL is the Java value when secondary verification is on.
const secondaryTTL = 120 * time.Second

// checkBehaviour applies the optional track policy. A nil policy accepts every
// track, matching the recorded contract.
func checkBehaviour(policy *TrackPolicy, track Track) bool {
	if policy == nil {
		return true
	}

	points := track.TrackList
	if len(points) < policy.MinPoints {
		return false
	}

	elapsed := time.Duration(track.StopTime-track.StartTime) * time.Millisecond
	if elapsed < policy.MinDuration {
		return false
	}
	if policy.MaxDuration > 0 && elapsed > policy.MaxDuration {
		return false
	}

	var minY, maxY float64
	first := true
	for _, point := range points {
		if first {
			minY, maxY, first = point.Y, point.Y, false
			continue
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	if policy.MaxJitter > 0 && maxY-minY > policy.MaxJitter {
		return false
	}
	if policy.MaxY > 0 && math.Abs(float64(track.Top)) > float64(policy.MaxY) {
		return false
	}
	return true
}
