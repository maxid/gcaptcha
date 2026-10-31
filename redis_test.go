package gcaptcha

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisStoreOneTime(t *testing.T) {
	mr := miniredis.RunT(t)
	store := NewRedisStore(mr.Addr(), "", 0, "")
	defer store.Close()
	ctx := context.Background()

	data := ValidData{Percentage: 0.42, Type: TypeSlider}
	if err := store.Save(ctx, "id1", data, 20*time.Second); err != nil {
		t.Fatal(err)
	}
	// The key layout is Java's `...::captcha::<id>`: this store only appends the
	// id, so the caller passes the platform part plus DefaultRedisPrefix.
	storeWithPlatform := NewRedisStore(mr.Addr(), "", 0, "xdmp_platform"+DefaultRedisPrefix)
	defer storeWithPlatform.Close()
	if err := storeWithPlatform.Save(ctx, "id2", data, 20*time.Second); err != nil {
		t.Fatal(err)
	}
	if key := "xdmp_platform" + DefaultRedisPrefix + "id2"; !mr.Exists(key) {
		t.Fatalf("key %q missing; keys: %v", key, mr.Keys())
	}
	if !mr.Exists(DefaultRedisPrefix + "id1") {
		t.Fatalf("default prefix not applied: %v", mr.Keys())
	}
	if ttl := mr.TTL(DefaultRedisPrefix + "id1"); ttl <= 0 || ttl > 20*time.Second {
		t.Fatalf("ttl = %v", ttl)
	}

	got, ok, err := store.Take(ctx, "id1")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got.Percentage != data.Percentage || got.Type != data.Type {
		t.Fatalf("data = %+v", got)
	}
	if _, ok, _ := store.Take(ctx, "id1"); ok {
		t.Fatal("second take must fail")
	}
}

func TestRedisStoreCustomPrefixAndMissing(t *testing.T) {
	mr := miniredis.RunT(t)
	store := NewRedisStore(mr.Addr(), "", 0, "test:")
	defer store.Close()
	ctx := context.Background()

	if err := store.Save(ctx, "x", ValidData{Percentage: 0.1}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if !mr.Exists("test:x") {
		t.Fatal("custom prefix not used")
	}
	if _, ok, err := store.Take(ctx, "nope"); err != nil || ok {
		t.Fatalf("missing: ok=%v err=%v", ok, err)
	}
}
