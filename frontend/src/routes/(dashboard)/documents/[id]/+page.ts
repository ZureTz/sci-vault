import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => {
	return {
		id: parseInt(params.id, 10)
	};
};
