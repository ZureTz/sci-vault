/**
 * Persistent search state — survives navigation between search and document detail pages.
 * Kept in-memory only (not localStorage) so it resets on full page reload.
 */
import type { SearchResultItem, SearchHistoryItem } from '$lib/api/search';

let _query = $state('');
let _results = $state<SearchResultItem[]>([]);
let _hasSearched = $state(false);
let _history = $state<SearchHistoryItem[]>([]);

// Recent searches filtered by what the user has typed (case-insensitive substring).
// Empty query → show everything.
const _filteredHistory = $derived.by(() => {
	const q = _query.trim().toLowerCase();
	if (!q) return _history;
	return _history.filter((h) => h.query.toLowerCase().includes(q));
});

export function getSearchState() {
	return {
		get query() {
			return _query;
		},
		get results() {
			return _results;
		},
		get hasSearched() {
			return _hasSearched;
		},
		get history() {
			return _history;
		},
		get filteredHistory() {
			return _filteredHistory;
		}
	};
}

export function setSearchState(query: string, results: SearchResultItem[], hasSearched: boolean) {
	_query = query;
	_results = results;
	_hasSearched = hasSearched;
}

export function setSearchQuery(query: string) {
	_query = query;
}

export function setSearchHistory(items: SearchHistoryItem[]) {
	_history = items;
}
