export interface StoredUser {
	id: string;
	username: string;
	email: string;
}

const STORAGE_KEY = 'user';

function loadFromStorage(): StoredUser {
	if (typeof localStorage === 'undefined') return { id: '', username: '', email: '' };
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (raw) return JSON.parse(raw) as StoredUser;
	} catch {
		// ignore malformed data
	}
	return { id: '', username: '', email: '' };
}

// ── reactive state ────────────────────────────────────────────────────────────
let _user = $state<StoredUser>(loadFromStorage());
let _avatarUrl = $state<string | undefined>(undefined);

// ── public getters (read-only views) ─────────────────────────────────────────
export function getUser() {
	return _user;
}

export function getAvatarUrl() {
	return _avatarUrl;
}

// ── mutations ─────────────────────────────────────────────────────────────────
/** Call after login / register. Persists to localStorage. */
export function setUser(u: StoredUser) {
	_user = u;
	localStorage.setItem(STORAGE_KEY, JSON.stringify(u));
}

export function setAvatarUrl(url: string | undefined) {
	_avatarUrl = url;
}

/** Call on logout or 401. Clears store and localStorage. */
export function clearUser() {
	_user = { id: '', username: '', email: '' };
	_avatarUrl = undefined;
	localStorage.removeItem(STORAGE_KEY);
}
