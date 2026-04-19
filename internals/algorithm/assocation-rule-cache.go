package algorithm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CacheManager owns the live *RuleIndex and keeps it fresh.
//
// Design:
//   - One atomic pointer stores the current index.
//   - Reads (Recommend calls) load the pointer with no locking — zero contention.
//   - Rebuilds happen in a background goroutine; on completion the pointer is
//     swapped atomically. In-flight reads continue using the old snapshot.
//   - If a rebuild fails, the old snapshot stays live and we retry next tick.
type CacheManager struct {
	orderRepository repository.OrderRepo
	db              *pgxpool.Pool
	ttl             time.Duration  // rebuild interval, e.g. 1 * time.Hour
	ptr             unsafe.Pointer // *RuleIndex, swapped atomically
	rebuildMu       sync.Mutex
}

// NewCacheManager creates a manager but does NOT start the background loop.
// Call Start() once at application startup.
func NewCacheManager(db *pgxpool.Pool, ttl time.Duration, orderRepo repository.OrderRepo) *CacheManager {
	return &CacheManager{db: db, ttl: ttl, orderRepository: orderRepo}
}

// Start performs the first synchronous build (so the app is ready immediately).
// Periodic refresh is handled by a dedicated jobs goroutine.
func (cm *CacheManager) Start(ctx context.Context) error {
	if err := cm.rebuild(ctx); err != nil {
		return fmt.Errorf("recommendation: initial cache build: %w", err)
	}
	return nil
}

// Get returns the current hot RuleIndex. Always non-nil after Start returns.
func (cm *CacheManager) Get() *RuleIndex {
	return (*RuleIndex)(atomic.LoadPointer(&cm.ptr))
}

// Refresh forces an immediate synchronous rebuild.
// It shares the same lock path as scheduled rebuilds to prevent overlap.
func (cm *CacheManager) Refresh(ctx context.Context) error {
	return cm.rebuild(ctx)
}

// rebuild fetches fresh data, runs Apriori, and hot-swaps the index pointer.
func (cm *CacheManager) rebuild(ctx context.Context) error {
	cm.rebuildMu.Lock()
	defer cm.rebuildMu.Unlock()

	log.Println("[recommendation] building cache…")

	// 1. Fetch last 6 months of completed orders → map[order_id][]menu_item_id
	orderMap, err := cm.orderRepository.GetOrderItemsForAprioriAlgorithm(ctx)
	if err != nil {
		return fmt.Errorf("fetch orders: %w", err)
	}

	// 2. Fetch top-N popular items (used as the final fallback)
	popular, err := cm.orderRepository.GetPopularItems(ctx, 10)
	if err != nil {
		return fmt.Errorf("fetch popular items: %w", err)
	}
	popular = mergePopularWithDefault(popular)

	// 3. Run Apriori (skipped gracefully if there are no orders yet)
	var rules []AssociationRule
	if len(orderMap) > 0 {
		_, rules = Apriori(
			orderMap,
			0.03, // minSupport:    present in ≥ 3 % of orders
			0.20, // minConfidence: rule correct ≥ 20 % of the time
			4,    // maxItemsetSize: matches our 4-item context window
		)
	}
	if len(rules) == 0 {
		rules = DefaultAssociationRules()
	}

	// 4. Atomically swap in the new index
	next := NewRuleIndex(rules, popular, cm.ttl)
	atomic.StorePointer(&cm.ptr, unsafe.Pointer(next))

	log.Printf("[recommendation] cache ready — %d rules, %d popular items", len(rules), len(popular))
	return nil
}

func mergePopularWithDefault(popular []string) []string {
	seen := make(map[string]bool, len(popular)+len(DefaultPopularMenuItemIDs))
	merged := make([]string, 0, len(popular)+len(DefaultPopularMenuItemIDs))

	for _, id := range popular {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		merged = append(merged, id)
	}
	for _, id := range DefaultPopularMenuItemIDs {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		merged = append(merged, id)
	}

	return merged
}
