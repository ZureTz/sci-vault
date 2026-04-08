import request from './request';

export async function translateSummary(
	text: string,
	targetLanguage: string,
	onChunk: (chunk: string) => void,
	onDone: () => void,
	onError: (error: string) => void
): Promise<void> {
	let processed = 0;
	let done = false;

	function parseChunks(raw: string) {
		const newText = raw.slice(processed);
		processed = raw.length;

		for (const line of newText.split('\n')) {
			if (line.startsWith('data: ')) {
				const data = line.slice(6);
				if (data === '[DONE]') {
					done = true;
					onDone();
					return;
				}
				onChunk(data);
			} else if (line.startsWith('event: error')) {
				done = true;
				onError('Translation failed');
				return;
			}
		}
	}

	try {
		await request.post('/translate/summary', { text, target_language: targetLanguage }, {
			responseType: 'text',
			onDownloadProgress: (event) => {
				const responseText = (event.event?.target as XMLHttpRequest)?.responseText ?? '';
				parseChunks(responseText);
			}
		});
	} catch (error: unknown) {
		if (!done) onError(String(error));
		return;
	}

	if (!done) onDone();
}
