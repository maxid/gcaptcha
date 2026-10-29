package gcaptcha

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// TypeSlider is the only challenge type implemented so far.
const TypeSlider = "SLIDER"

// Tolerant is the allowed percentage deviation (Java DEFAULT_TOLERANT).
const Tolerant = 0.02

// Point is one sample of the drag track.
type Point struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	T    float64 `json:"t"`
	Type string  `json:"type,omitempty"`
}

// Track is the client-side drag data (Java ImageCaptchaTrack).
type Track struct {
	BgImageWidth        int     `json:"bgImageWidth"`
	BgImageHeight       int     `json:"bgImageHeight"`
	TemplateImageWidth  int     `json:"templateImageWidth"`
	TemplateImageHeight int     `json:"templateImageHeight"`
	StartTime           int64   `json:"startTime"`
	StopTime            int64   `json:"stopTime"`
	Left                int     `json:"left"`
	Top                 int     `json:"top"`
	TrackList           []Point `json:"trackList"`
}

// Challenge is the generation response (Java ImageCaptchaVO field names).
type Challenge struct {
	ID                    string `json:"id"`
	Type                  string `json:"type"`
	BackgroundImage       string `json:"backgroundImage"`
	TemplateImage         string `json:"templateImage"`
	BackgroundImageTag    string `json:"backgroundImageTag"`
	TemplateImageTag      string `json:"templateImageTag"`
	BackgroundImageWidth  int    `json:"backgroundImageWidth"`
	BackgroundImageHeight int    `json:"backgroundImageHeight"`
	TemplateImageWidth    int    `json:"templateImageWidth"`
	TemplateImageHeight   int    `json:"templateImageHeight"`
	Data                  any    `json:"data"`
}

// ValidData is what the server remembers for one challenge.
type ValidData struct {
	Percentage float64 `json:"percentage"`
	Type       string  `json:"type"`
}

// Store persists challenges until they are consumed.
type Store interface {
	Save(ctx context.Context, id string, data ValidData, ttl time.Duration) error
	// Take returns the data and removes it (one-time use).
	Take(ctx context.Context, id string) (ValidData, bool, error)
}

// Service generates and validates slider challenges.
type Service struct {
	store  Store
	ttl    time.Duration
	width  int
	height int
	piece  int

	mu  sync.Mutex
	rng *rand.Rand
}

// Option configures a Service.
type Option func(*Service)

// WithTTL sets the challenge lifetime (default 20s, matching the Java default).
func WithTTL(ttl time.Duration) Option {
	return func(s *Service) {
		if ttl > 0 {
			s.ttl = ttl
		}
	}
}

// WithSize sets the background size (default 600x360).
func WithSize(width, height int) Option {
	return func(s *Service) {
		if width > 0 && height > 0 {
			s.width, s.height = width, height
		}
	}
}

// WithPieceSize sets the slider piece size (default 110).
func WithPieceSize(size int) Option {
	return func(s *Service) {
		if size > 0 {
			s.piece = size
		}
	}
}

// WithRand sets the random source (tests).
func WithRand(rng *rand.Rand) Option {
	return func(s *Service) {
		if rng != nil {
			s.rng = rng
		}
	}
}

// New creates a service backed by store.
func New(store Store, opts ...Option) *Service {
	s := &Service{
		store:  store,
		ttl:    20 * time.Second,
		width:  600,
		height: 360,
		piece:  110,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Generate creates a new challenge and remembers its solution.
func (s *Service) Generate(ctx context.Context) (Challenge, error) {
	s.mu.Lock()
	x, y := s.randomPositionLocked()
	background, template, err := renderImages(s.width, s.height, s.piece, x, y, s.rng)
	s.mu.Unlock()
	if err != nil {
		return Challenge{}, err
	}
	id, err := newUUID()
	if err != nil {
		return Challenge{}, err
	}
	if err := s.store.Save(ctx, id, ValidData{
		Percentage: float64(x) / float64(s.width),
		Type:       TypeSlider,
	}, s.ttl); err != nil {
		return Challenge{}, err
	}
	return Challenge{
		ID:                    id,
		Type:                  TypeSlider,
		BackgroundImage:       dataURI(background),
		TemplateImage:         dataURI(template),
		BackgroundImageWidth:  s.width,
		BackgroundImageHeight: s.height,
		TemplateImageWidth:    s.piece,
		TemplateImageHeight:   s.piece,
	}, nil
}

// Validate checks a drag track against the stored challenge. Challenges are
// single-use: a second validation with the same id always fails.
func (s *Service) Validate(ctx context.Context, id string, track Track) (bool, error) {
	if id == "" {
		return false, nil
	}
	data, ok, err := s.store.Take(ctx, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return checkTrack(track, data.Percentage), nil
}

func checkTrack(track Track, want float64) bool {
	if track.BgImageWidth <= 0 || len(track.TrackList) < 2 {
		return false
	}
	first := track.TrackList[0]
	last := track.TrackList[len(track.TrackList)-1]
	got := (last.X - first.X) / float64(track.BgImageWidth)
	return math.Abs(got-want) <= Tolerant
}

func (s *Service) randomPositionLocked() (int, int) {
	maxX := s.width - s.piece - 10
	if maxX < 10 {
		maxX = 10
	}
	maxY := s.height - s.piece - 10
	if maxY < 10 {
		maxY = 10
	}
	x := 10 + s.rng.Intn(maxX-10+1)
	y := 10 + s.rng.Intn(maxY-10+1)
	return x, y
}

// ErrNoStore is returned when the service is created without a store.
var ErrNoStore = errors.New("gcaptcha: store is required")
