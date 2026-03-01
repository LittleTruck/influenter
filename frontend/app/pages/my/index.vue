<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
navigateTo('/my/collaboration-items', { replace: true })
</script>
