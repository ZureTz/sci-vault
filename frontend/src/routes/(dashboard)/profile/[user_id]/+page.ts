import type { PageLoad } from './$types';

import userApi from '$lib/api/user';

export const load: PageLoad = ({ params }) => {
	const profile = userApi.getProfile(params.user_id).catch((error: unknown) => {
		console.error(
			'error fetching user profile: ' + (error instanceof Error ? error.message : String(error))
		);
		return null;
	});

	return { profile };
};
