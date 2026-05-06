<script lang="ts">
	import { LineChart } from 'layerchart';

	import * as Chart from '$lib/components/ui/chart';
	import type { ChartConfig } from '$lib/components/ui/chart';
	import type { DayCount } from '$lib/api/stats';

	let {
		views,
		likes,
		viewsLabel,
		likesLabel
	}: {
		views: DayCount[];
		likes: DayCount[];
		viewsLabel: string;
		likesLabel: string;
	} = $props();

	const merged = $derived.by(() => {
		const byDate: Record<string, { dateValue: Date; views: number; likes: number }> = {};
		for (const v of views) {
			byDate[v.date] = {
				dateValue: new Date(`${v.date}T00:00:00Z`),
				views: v.count,
				likes: 0
			};
		}
		for (const l of likes) {
			const existing = byDate[l.date];
			if (existing) {
				existing.likes = l.count;
			} else {
				byDate[l.date] = {
					dateValue: new Date(`${l.date}T00:00:00Z`),
					views: 0,
					likes: l.count
				};
			}
		}
		return Object.values(byDate).sort((a, b) => a.dateValue.getTime() - b.dateValue.getTime());
	});

	const config = $derived.by<ChartConfig>(() => ({
		views: { label: viewsLabel, color: 'var(--chart-1)' },
		likes: { label: likesLabel, color: 'var(--chart-2)' }
	}));

	const series = $derived([
		{ key: 'views', label: viewsLabel, value: 'views', color: 'var(--color-views)' },
		{ key: 'likes', label: likesLabel, value: 'likes', color: 'var(--color-likes)' }
	]);
</script>

<Chart.Container {config} class="h-50 w-full">
	<LineChart
		data={merged}
		x="dateValue"
		{series}
		legend
		grid
		rule={false}
		props={{
			xAxis: {
				format: (d: Date) => d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
			}
		}}
	>
		{#snippet tooltip()}
			<Chart.Tooltip
				indicator="dot"
				labelFormatter={(value) =>
					(value instanceof Date ? value : new Date(value as string)).toLocaleDateString(
						undefined,
						{
							year: 'numeric',
							month: 'short',
							day: 'numeric'
						}
					)}
			/>
		{/snippet}
	</LineChart>
</Chart.Container>
