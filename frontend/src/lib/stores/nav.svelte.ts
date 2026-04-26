/**
 * Tracks where the user came from when they entered a /documents/[id] detail
 * page, so the sidebar can keep the originating section highlighted instead of
 * unconditionally treating doc details as part of the "Documents" group.
 *
 * The detail page is reachable from My Documents, Search, History, the
 * Dashboard, and from "similar document" cards on another detail page; the
 * origin is captured on entry, preserved while hopping between detail pages,
 * and cleared the moment the user navigates anywhere else.
 *
 * The value is the URL pathname of the origin (e.g. "/search", "/mine/history",
 * "/documents/mine"), since pathname is what the sidebar's `isActive` checks
 * compare against.
 */
let _docDetailOrigin = $state<string | null>(null);

export function getDocDetailOrigin(): string | null {
	return _docDetailOrigin;
}

export function setDocDetailOrigin(pathname: string | null) {
	_docDetailOrigin = pathname;
}
