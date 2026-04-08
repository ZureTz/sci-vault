/**
 * Streaming translation using fetch + ReadableStream.
 *
 * Axios (XHR) is not used here because XHR buffers internally and does not
 * expose true streaming; fetch ReadableStream reads chunks at the network
 * level, which works correctly through proxies like Cloudflare.
 */
export async function translateSummary(
	text: string,
	targetLanguage: string,
	onChunk: (chunk: string) => void,
	onDone: () => void,
	onError: (error: string) => void
): Promise<void> {
	const token = localStorage.getItem('token');

	let response: Response;
	try {
		response = await fetch('/api/v1/translate/summary', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				...(token ? { Authorization: `Bearer ${token}` } : {})
			},
			body: JSON.stringify({ text, target_language: targetLanguage })
		});
	} catch {
		onError('network error');
		return;
	}

	if (!response.ok) {
		onError(`HTTP ${response.status}`);
		return;
	}

	const reader = response.body?.getReader();
	if (!reader) {
		onError('ReadableStream not supported');
		return;
	}

	const decoder = new TextDecoder();
	let buffer = '';

	try {
		while (true) {
			const { done, value } = await reader.read();
			if (done) break;

			buffer += decoder.decode(value, { stream: true });
			const lines = buffer.split('\n');
			buffer = lines.pop() ?? '';

			for (const line of lines) {
				if (line.startsWith('data: ')) {
					const data = line.slice(6);
					if (data === '[DONE]') {
						onDone();
						return;
					}
					onChunk(data);
				} else if (line.startsWith('event: error')) {
					onError('Translation failed');
					return;
				}
			}
		}
	} finally {
		reader.releaseLock();
	}

	onDone();
}
