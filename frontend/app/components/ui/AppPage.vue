<template>
	<div class="flex flex-col gap-4">
		<!-- Page header -->
		<div class="flex items-center justify-between">
			<div class="min-w-0">
				<!-- title slot -->
				<slot name="title">
					<h1 class="text-xl font-semibold truncate">{{ title }}</h1>
				</slot>
				<!-- toolbar slot -->
				<div v-if="$slots.toolbar" class="mt-2">
					<slot name="toolbar" />
				</div>
			</div>
			<!-- primary actions slot (R-01: 主要動作固定在右上角) -->
			<div class="shrink-0 flex items-center gap-2">
				<slot name="actions" />
			</div>
		</div>

		<!-- page content -->
		<div>
			<slot />
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps<{
	title?: string
}>()
</script>



