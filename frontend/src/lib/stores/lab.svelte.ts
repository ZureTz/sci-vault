/**
 * Shared reactive signal that triggers a lab list reload in any subscriber.
 * Pages that create or join a lab call `invalidateLabs()`;
 * the sidebar watches `labsVersion` and re-fetches when it changes.
 */
let labsVersion = $state(0);

export function invalidateLabs() {
	labsVersion++;
}

export function getLabsVersion() {
	return labsVersion;
}

/**
 * Active lab state — tracks which lab the user is currently working in.
 * Persisted to localStorage so the selection survives page reloads.
 */
interface ActiveLab {
	id: number;
	name: string;
	role: 'owner' | 'member';
}

let _activeLab = $state<ActiveLab | null>(null);

// Restore from localStorage on module load
if (typeof localStorage !== 'undefined') {
	try {
		const stored = localStorage.getItem('active_lab');
		if (stored) {
			_activeLab = JSON.parse(stored);
		}
	} catch {
		// ignore parse errors
	}
}

export function getActiveLab(): ActiveLab | null {
	return _activeLab;
}

export function setActiveLab(lab: ActiveLab | null) {
	_activeLab = lab;
	if (lab) {
		localStorage.setItem('active_lab', JSON.stringify(lab));
		localStorage.setItem('active_lab_id', String(lab.id));
	} else {
		localStorage.removeItem('active_lab');
		localStorage.removeItem('active_lab_id');
	}
}
