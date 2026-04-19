package algorithm

import (
	"sort"
	"strings"
)

// ItemSet represents a set of menu item IDs
type ItemSet []string

// key returns a canonical string key for the itemset (sorted, joined)
func (is ItemSet) key() string {
	sorted := make([]string, len(is))
	copy(sorted, is)
	sort.Strings(sorted)
	return strings.Join(sorted, "|")
}

// FrequentItemSet holds an itemset and its support count
type FrequentItemSet struct {
	Items   ItemSet
	Support float64 // fraction of transactions containing this itemset
	Count   int
}

// AssociationRule represents an if-then rule: Antecedent => Consequent
type AssociationRule struct {
	Antecedent ItemSet
	Consequent ItemSet
	Support    float64
	Confidence float64
	Lift       float64
}

// Apriori runs the Apriori algorithm on the given order map.
//
//	orderMap        – map of order_id → []menu_item_id (each order's items)
//	minSupport      – minimum fraction (0–1), e.g. 0.05 = 5%
//	minConfidence   – minimum confidence (0–1), e.g. 0.3 = 30%
//	maxItemsetSize  – cap itemset size to keep runtime manageable (e.g. 4)
//
// Each order is treated as one transaction; only menu_item_id values are used.
func Apriori(
	orderMap map[string][]string,
	minSupport float64,
	minConfidence float64,
	maxItemsetSize int,
) ([]FrequentItemSet, []AssociationRule) {

	// Convert map → deduplicated transactions (one []string per order)
	transactions := make([][]string, 0, len(orderMap))
	for _, items := range orderMap {
		// Deduplicate menu_item_ids within the same order
		seen := make(map[string]bool, len(items))
		deduped := make([]string, 0, len(items))
		for _, id := range items {
			if !seen[id] {
				seen[id] = true
				deduped = append(deduped, id)
			}
		}
		if len(deduped) > 0 {
			transactions = append(transactions, deduped)
		}
	}

	n := len(transactions)
	if n == 0 {
		return nil, nil
	}

	minCount := int(minSupport * float64(n))
	if minCount < 1 {
		minCount = 1
	}

	// ── Step 1: build candidate 1-itemsets ──────────────────────────────────
	freq1 := generateFrequent1(transactions, minCount, n)

	allFrequent := make([]FrequentItemSet, 0, len(freq1))
	allFrequent = append(allFrequent, freq1...)

	// itemset key → support (needed for lift calculation later)
	supportMap := make(map[string]float64, len(freq1))
	for _, f := range freq1 {
		supportMap[f.Items.key()] = f.Support
	}

	prev := freq1

	// ── Step 2: iteratively generate k-itemsets ──────────────────────────────
	for k := 2; k <= maxItemsetSize && len(prev) > 0; k++ {
		candidates := aprioriGen(prev)
		if len(candidates) == 0 {
			break
		}

		counted := countCandidates(transactions, candidates, minCount, n)
		if len(counted) == 0 {
			break
		}

		for _, f := range counted {
			supportMap[f.Items.key()] = f.Support
		}
		allFrequent = append(allFrequent, counted...)
		prev = counted
	}

	// ── Step 3: generate association rules ──────────────────────────────────
	rules := generateRules(allFrequent, supportMap, minConfidence)

	return allFrequent, rules
}

// ─── helpers ────────────────────────────────────────────────────────────────

// generateFrequent1 counts every single item across all transactions.
func generateFrequent1(transactions [][]string, minCount, totalTx int) []FrequentItemSet {
	counts := make(map[string]int)
	for _, tx := range transactions {
		seen := make(map[string]bool, len(tx))
		for _, item := range tx {
			if !seen[item] {
				counts[item]++
				seen[item] = true
			}
		}
	}

	result := make([]FrequentItemSet, 0)
	for item, count := range counts {
		if count >= minCount {
			result = append(result, FrequentItemSet{
				Items:   ItemSet{item},
				Support: float64(count) / float64(totalTx),
				Count:   count,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Items[0] < result[j].Items[0]
	})
	return result
}

// aprioriGen produces candidate (k+1)-itemsets from the frequent k-itemsets
// using the F_(k-1) × F_(k-1) merge step.
func aprioriGen(prev []FrequentItemSet) []ItemSet {
	candidates := make([]ItemSet, 0)
	seen := make(map[string]bool)

	for i := 0; i < len(prev); i++ {
		for j := i + 1; j < len(prev); j++ {
			a := prev[i].Items
			b := prev[j].Items

			if !sharePrefix(a, b) {
				continue
			}

			merged := make(ItemSet, len(a)+1)
			copy(merged, a)
			merged[len(a)] = b[len(b)-1]

			k := merged.key()
			if !seen[k] {
				seen[k] = true
				candidates = append(candidates, merged)
			}
		}
	}
	return candidates
}

// sharePrefix returns true when two sorted itemsets of equal length share
// all elements except the last.
func sharePrefix(a, b ItemSet) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a)-1; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return a[len(a)-1] < b[len(b)-1]
}

// countCandidates scans all transactions and counts each candidate itemset.
func countCandidates(
	transactions [][]string,
	candidates []ItemSet,
	minCount int,
	totalTx int,
) []FrequentItemSet {

	type candidateEntry struct {
		set   map[string]bool
		count int
	}
	entries := make([]candidateEntry, len(candidates))
	for i, c := range candidates {
		s := make(map[string]bool, len(c))
		for _, item := range c {
			s[item] = true
		}
		entries[i] = candidateEntry{set: s}
	}

	for _, tx := range transactions {
		txSet := make(map[string]bool, len(tx))
		for _, item := range tx {
			txSet[item] = true
		}
		for i := range entries {
			if subsetOf(entries[i].set, txSet) {
				entries[i].count++
			}
		}
	}

	result := make([]FrequentItemSet, 0)
	for i, e := range entries {
		if e.count >= minCount {
			result = append(result, FrequentItemSet{
				Items:   candidates[i],
				Support: float64(e.count) / float64(totalTx),
				Count:   e.count,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Items.key() < result[j].Items.key()
	})
	return result
}

// subsetOf returns true when every key in sub exists in super.
func subsetOf(sub, super map[string]bool) bool {
	for k := range sub {
		if !super[k] {
			return false
		}
	}
	return true
}

// generateRules extracts association rules from all frequent itemsets.
func generateRules(
	frequent []FrequentItemSet,
	supportMap map[string]float64,
	minConfidence float64,
) []AssociationRule {

	rules := make([]AssociationRule, 0)

	for _, fis := range frequent {
		if len(fis.Items) < 2 {
			continue
		}
		subsets := properSubsets(fis.Items)
		for _, ant := range subsets {
			con := difference(fis.Items, ant)
			if len(con) == 0 {
				continue
			}

			antSupport, ok := supportMap[ItemSet(ant).key()]
			if !ok || antSupport == 0 {
				continue
			}

			confidence := fis.Support / antSupport

			if confidence < minConfidence {
				continue
			}

			conSupport, ok := supportMap[ItemSet(con).key()]
			var lift float64
			if ok && conSupport > 0 {
				lift = confidence / conSupport
			}

			rules = append(rules, AssociationRule{
				Antecedent: ant,
				Consequent: con,
				Support:    fis.Support,
				Confidence: confidence,
				Lift:       lift,
			})
		}
	}

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Lift != rules[j].Lift {
			return rules[i].Lift > rules[j].Lift
		}
		return rules[i].Confidence > rules[j].Confidence
	})
	return rules
}

// properSubsets returns all non-empty proper subsets of items.
func properSubsets(items ItemSet) []ItemSet {
	n := len(items)
	total := (1 << n) - 1
	result := make([]ItemSet, 0, total-1)
	for mask := 1; mask < total; mask++ {
		sub := make(ItemSet, 0, n)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				sub = append(sub, items[i])
			}
		}
		result = append(result, sub)
	}
	return result
}

// difference returns items in a that are not in b.
func difference(a, b ItemSet) ItemSet {
	bSet := make(map[string]bool, len(b))
	for _, v := range b {
		bSet[v] = true
	}
	result := make(ItemSet, 0, len(a))
	for _, v := range a {
		if !bSet[v] {
			result = append(result, v)
		}
	}
	return result
}
