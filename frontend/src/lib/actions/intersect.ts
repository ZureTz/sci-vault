export function intersect(
	node: HTMLElement,
	options: IntersectionObserverInit = { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }
) {
	const observer = new IntersectionObserver((entries) => {
		entries.forEach((entry) => {
			if (entry.isIntersecting) {
				node.classList.add('intersected');
				observer.unobserve(node);
			}
		});
	}, options);

	observer.observe(node);

	return {
		destroy() {
			observer.disconnect();
		}
	};
}
