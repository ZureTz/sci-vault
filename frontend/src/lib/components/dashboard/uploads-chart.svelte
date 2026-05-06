<script lang="ts">
	import { AreaChart } from 'layerchart';

	import * as Chart from '$lib/components/ui/chart';
	import type { ChartConfig } from '$lib/components/ui/chart';
	import type { DayCount } from '$lib/api/stats';

	let { data, label }: { data: DayCount[]; label: string } = $props();

	const parsed = $derived(data.map((d) => ({ ...d, dateValue: new Date(`${d.date}T00:00:00Z`) })));

	const config = $derived.by<ChartConfig>(() => ({
		count: { label, color: 'var(--chart-1)' }
	}));
</script>

<Chart.Container {config} class="h-50 w-full">
	<AreaChart
		data={parsed}
		x="dateValue"
		y="count"
		series={[{ key: 'count', label, value: 'count', color: 'var(--color-count)' }]}
		legend={false}
		grid
		rule={false}
		props={{
			area: { 'fill-opacity': 0.25 },
			xAxis: {
				format: (d: Date) => d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
			}
		}}
	>
		{#snippet tooltip()}
			<Chart.Tooltip
				indicator="line"
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
	</AreaChart>
</Chart.Container>
