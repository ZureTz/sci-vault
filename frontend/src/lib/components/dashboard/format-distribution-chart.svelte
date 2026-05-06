<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { PieChart } from 'layerchart';

	import * as Chart from '$lib/components/ui/chart';
	import type { ChartConfig } from '$lib/components/ui/chart';
	import type { FormatBucket } from '$lib/api/stats';

	let { data }: { data: FormatBucket[] } = $props();

	function bucketKey(contentType: string): string {
		switch (contentType) {
			case 'application/pdf':
				return 'pdf';
			case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
				return 'docx';
			case 'application/vnd.openxmlformats-officedocument.presentationml.presentation':
				return 'pptx';
			case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
				return 'xlsx';
			case 'text/plain':
				return 'txt';
			case 'text/markdown':
				return 'md';
			default:
				return 'other';
		}
	}

	const PALETTE = [
		'var(--chart-1)',
		'var(--chart-2)',
		'var(--chart-3)',
		'var(--chart-4)',
		'var(--chart-5)'
	];

	const buckets = $derived.by(() => {
		const out: Record<string, number> = {};
		for (const row of data) {
			const k = bucketKey(row.content_type);
			out[k] = (out[k] ?? 0) + row.count;
		}
		return Object.entries(out)
			.map(([key, count]) => ({
				key,
				label: $_(`dashboard.formats.${key}`),
				count
			}))
			.sort((a, b) => b.count - a.count);
	});

	const config = $derived.by<ChartConfig>(() => {
		const cfg: ChartConfig = {};
		buckets.forEach((b, i) => {
			cfg[b.key] = { label: b.label, color: PALETTE[i % PALETTE.length] };
		});
		return cfg;
	});

	const total = $derived(buckets.reduce((acc, b) => acc + b.count, 0));
</script>

{#if total === 0}
	<div class="flex h-50 w-full items-center justify-center text-sm text-muted-foreground">
		{$_('dashboard.charts.empty')}
	</div>
{:else}
	<Chart.Container {config} class="h-55 w-full">
		<PieChart
			data={buckets}
			key="key"
			label="label"
			value="count"
			innerRadius={0.55}
			cornerRadius={2}
			padAngle={0.012}
			legend
		>
			{#snippet tooltip()}
				<Chart.Tooltip hideLabel />
			{/snippet}
		</PieChart>
	</Chart.Container>
{/if}
