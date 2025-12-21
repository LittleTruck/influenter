<script setup lang="ts">
import { BaseCard } from '~/components/base'

withDefaults(defineProps<{
	variant?: 'card' | 'plain'
}>(), {
	variant: 'card'
})
</script>

<template>
	<div>
		<BaseCard v-if="variant === 'card'">
			<template v-if="$slots.header" #header>
				<slot name="header" />
			</template>
			<slot />
			<template v-if="$slots.footer" #footer>
				<slot name="footer" />
			</template>
		</BaseCard>
		<div v-else class="rounded-md border border-gray-200 dark:border-neutral-700 p-4">
			<slot />
		</div>
	</div>
</template>



