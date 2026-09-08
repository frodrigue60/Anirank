package public

import (
	infrastructurecache "anirank/api/internal/infrastructure/cache"
	"context"
	"testing"
	"time"
)

func TestViewDeduperSuppressesViewsUntilTTLExpires(t *testing.T) {
	deduper := newViewDeduper(24 * time.Hour)
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)

	if !deduper.MarkIfNew("view:client:song", now) {
		t.Fatal("expected first view to be reserved")
	}
	if deduper.MarkIfNew("view:client:song", now.Add(time.Hour)) {
		t.Fatal("expected repeated view inside TTL to be suppressed")
	}
	if !deduper.MarkIfNew("view:client:song", now.Add(24*time.Hour)) {
		t.Fatal("expected view at TTL expiry to be reserved again")
	}
}

func TestShouldCountViewUsesMemoryWhenDistributedCacheIsDisabled(t *testing.T) {
	usecase := &CatalogUsecase{cache: infrastructurecache.NewNoOpCache()}
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)

	if !usecase.shouldCountView(context.Background(), "view:client:42", now) {
		t.Fatal("expected first view to count")
	}
	if usecase.shouldCountView(context.Background(), "view:client:42", now.Add(time.Minute)) {
		t.Fatal("expected local fallback to suppress repeated view")
	}
	if !usecase.shouldCountView(context.Background(), "view:other-client:42", now.Add(time.Minute)) {
		t.Fatal("expected a different client key to count")
	}
}
