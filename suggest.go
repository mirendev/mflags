package mflags

import (
	"sort"
	"strings"
)

// maxSuggestions caps how many alternatives a "did you mean" block offers.
// Beyond a handful the list stops being a suggestion and starts being help
// output, which the closing "--help" line already points at.
const maxSuggestions = 3

// editDistance returns the number of edits between a and b, counting an
// insertion, a deletion, a substitution, or a transposition of two adjacent
// characters as one each. Transpositions matter: swapping two letters is among
// the most common ways to mistype a word, and charging two edits for it would
// put "recieve" as far from "receive" as a word with two genuinely wrong
// letters.
//
// It compares runes rather than bytes so that non-ASCII names cost what they
// look like they cost.
func editDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)

	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	// Three rolling rows of the distance matrix: curr is the row being filled
	// in, prev the row above it, and prevPrev the one above that, which only
	// the transposition case reaches back for.
	prevPrev := make([]int, len(br)+1)
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}

			best := min(prev[j]+1, min(curr[j-1]+1, prev[j-1]+cost))

			// The last two characters of each are the same pair, swapped.
			if i > 1 && j > 1 && ar[i-1] == br[j-2] && ar[i-2] == br[j-1] {
				best = min(best, prevPrev[j-2]+1)
			}

			curr[j] = best
		}
		prevPrev, prev, curr = prev, curr, prevPrev
	}

	return prev[len(br)]
}

// suggestThreshold returns the largest edit distance still worth reporting for
// a word of the given length. Short words get a tight budget because at two
// edits every three-letter word is a neighbour of every other one.
func suggestThreshold(n int) int {
	switch {
	case n <= 2:
		return 0
	case n <= 4:
		return 1
	case n <= 8:
		return 2
	default:
		return 3
	}
}

// suggestNames returns the candidates closest to unknown, nearest first, or nil
// when nothing is close enough to be worth guessing at. Prefix matches rank
// ahead of pure edit distance: someone who typed "depl" almost certainly meant
// "deploy" even though four insertions is a long way in edit-distance terms.
func suggestNames(unknown string, candidates []string) []string {
	if unknown == "" {
		return nil
	}

	threshold := suggestThreshold(len([]rune(unknown)))

	type scored struct {
		name   string
		prefix bool
		dist   int
	}

	var matches []scored
	for _, c := range candidates {
		if c == "" {
			continue
		}

		// A prefix match is strong evidence on its own — an abbreviation or a
		// half-typed word — so it bypasses the distance threshold.
		if len(c) > len(unknown) && strings.HasPrefix(c, unknown) {
			matches = append(matches, scored{name: c, prefix: true, dist: len(c) - len(unknown)})
			continue
		}

		if d := editDistance(unknown, c); d <= threshold {
			matches = append(matches, scored{name: c, dist: d})
		}
	}

	if len(matches) == 0 {
		return nil
	}

	// Alphabetical ties keep the output stable; the dispatcher's command map
	// iterates in random order.
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].prefix != matches[j].prefix {
			return matches[i].prefix
		}
		if matches[i].dist != matches[j].dist {
			return matches[i].dist < matches[j].dist
		}
		return matches[i].name < matches[j].name
	})

	if len(matches) > maxSuggestions {
		matches = matches[:maxSuggestions]
	}

	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, m.name)
	}
	return names
}
