package algorithm

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ─── Types ────────────────────────────────────────────────────────────────────

// RuleIndex is the hot in-memory cache built once from Apriori output.
// Lookups are O(1) map reads — no DB calls, no recomputation per request.
type RuleIndex struct {
	mu           sync.RWMutex
	rules        map[string][]AssociationRule // sorted-key → rules ranked by lift
	popularItems []string                     // menu_item_ids ranked by order count
	builtAt      time.Time
	ttl          time.Duration
}

// Recommendation is the value returned to the HTTP handler.
type Recommendation struct {
	MenuItemIDs []string `json:"menu_item_ids"`

	// Source tells the frontend which fallback level fired so it can show
	// appropriate copy:
	//   "rule_4" → "Customers who ordered these 4 items also ordered…"
	//   "rule_3" / "rule_2" / "rule_1" → "People who ordered X often add…"
	//   "popular" → "Popular items you might like"
	Source string `json:"source"`
}

// ─── Build the index ──────────────────────────────────────────────────────────

// NewRuleIndex indexes every rule by its antecedent key for O(1) lookup.
// Rules from Apriori are already sorted by lift desc — we preserve that order.
func NewRuleIndex(
	rules []AssociationRule,
	popularItems []string,
	ttl time.Duration,
) *RuleIndex {
	idx := &RuleIndex{
		rules:        make(map[string][]AssociationRule, len(rules)),
		popularItems: popularItems,
		builtAt:      time.Now(),
		ttl:          ttl,
	}
	for _, r := range rules {
		key := sortedKey(r.Antecedent)
		idx.rules[key] = append(idx.rules[key], r)
	}
	return idx
}

// IsStale reports whether the cache should be rebuilt.
func (idx *RuleIndex) IsStale() bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return time.Since(idx.builtAt) > idx.ttl
}

// ─── Recommendation lookup ────────────────────────────────────────────────────

// Recommend returns suggested menu_item_ids given what the user has selected.
//
// Fallback chain:
//  1. Take up to the last 4 selected items as the context window
//  2. Lookup exact N-item antecedent key in the rule cache  (source: "rule_N")
//  3. Try all (N-1)-item subsets of the window
//  4. … continues down to 1-item subsets
//  5. If nothing matched → return popular items              (source: "popular")
//
// exclude: set of menu_item_ids already in the cart — never recommended back.
// limit:   max number of recommendations to return.
func (idx *RuleIndex) Recommend(
	_ context.Context,
	selectedIDs []string,
	exclude map[string]bool,
	limit int,
) Recommendation {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Cold start (no data yet) or empty cart → popular items
	if len(idx.rules) == 0 || len(selectedIDs) == 0 {
		fmt.Println("fall back to popular : ")
		return idx.popularFallback(exclude, limit)
	}

	// Cap to last 4 items (most recent selections are most relevant)
	window := selectedIDs
	if len(window) > 4 {
		window = window[len(window)-4:]
	}

	// Walk down: full window → 3-subsets → 2-subsets → single items
	for size := len(window); size >= 1; size-- {
		subsets := subsetsOfSize(window, size)
		items, source := idx.lookupSubsets(subsets, exclude, limit, size)
		if len(items) > 0 {
			return Recommendation{MenuItemIDs: items, Source: source}
		}
	}

	return idx.popularFallback(exclude, limit)
}

// lookupSubsets checks every subset in the batch, merges all matching
// consequents, deduplicates, and ranks by lift desc then confidence desc.
func (idx *RuleIndex) lookupSubsets(
	subsets [][]string,
	exclude map[string]bool,
	limit, subsetSize int,
) ([]string, string) {

	type scored struct{ lift, confidence float64 }
	scores := make(map[string]scored)

	for _, sub := range subsets {
		key := sortedKey(sub)
		rules, ok := idx.rules[key]
		if !ok {
			continue
		}
		for _, rule := range rules {
			for _, item := range rule.Consequent {
				if exclude[item] {
					continue
				}
				// Keep the best (highest lift) score seen across all matching rules
				if ex, exists := scores[item]; !exists || rule.Lift > ex.lift {
					scores[item] = scored{rule.Lift, rule.Confidence}
				}
			}
		}
	}

	if len(scores) == 0 {
		return nil, ""
	}

	type ranked struct {
		id         string
		lift       float64
		confidence float64
	}
	candidates := make([]ranked, 0, len(scores))
	for id, s := range scores {
		candidates = append(candidates, ranked{id, s.lift, s.confidence})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].lift != candidates[j].lift {
			return candidates[i].lift > candidates[j].lift
		}
		return candidates[i].confidence > candidates[j].confidence
	})

	result := make([]string, 0, limit)
	for _, c := range candidates {
		if len(result) >= limit {
			break
		}
		result = append(result, c.id)
	}

	return result, sourceLabel(subsetSize)
}

// popularFallback returns popular items filtered by the exclude set.
func (idx *RuleIndex) popularFallback(exclude map[string]bool, limit int) Recommendation {
	result := make([]string, 0, limit)
	for _, id := range idx.popularItems {
		if !exclude[id] {
			result = append(result, id)
		}
		if len(result) >= limit {
			break
		}
	}
	return Recommendation{MenuItemIDs: result, Source: "popular"}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// subsetsOfSize returns all combinations of exactly `size` items from `items`.
func subsetsOfSize(items []string, size int) [][]string {
	var result [][]string
	var build func(start int, cur []string)
	build = func(start int, cur []string) {
		if len(cur) == size {
			cp := make([]string, size)
			copy(cp, cur)
			result = append(result, cp)
			return
		}
		for i := start; i < len(items); i++ {
			build(i+1, append(cur, items[i]))
		}
	}
	build(0, nil)
	return result
}

// sortedKey produces the canonical "|"-joined sorted string for a slice of IDs.
// This must match the key format used inside algorithm.ItemSet.key().
func sortedKey(ids []string) string {
	cp := make([]string, len(ids))
	copy(cp, ids)
	sort.Strings(cp)
	key := ""
	for i, id := range cp {
		if i > 0 {
			key += "|"
		}
		key += id
	}
	return key
}

func sourceLabel(size int) string {
	switch size {
	case 4:
		return "rule_4"
	case 3:
		return "rule_3"
	case 2:
		return "rule_2"
	case 1:
		return "rule_1"
	default:
		return "popular"
	}
}
