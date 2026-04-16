/**
 * Persistent search state — survives navigation between search and document detail pages.
 * Kept in-memory only (not localStorage) so it resets on full page reload.
 */
import type { SearchResultItem } from '$lib/api/document';

let _query = $state('');
let _results = $state<SearchResultItem[]>([]);
let _hasSearched = $state(false);

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
