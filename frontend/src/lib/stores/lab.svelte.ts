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
