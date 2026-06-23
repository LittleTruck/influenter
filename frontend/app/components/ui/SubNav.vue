<script setup lang="ts">
defineProps<{
  /** 區塊標題（Col 2 頂部標籤，例如「管理」） */
  label?: string
  items: Array<{
    label: string
    icon: string
    to: string
  }>
}>()
</script>

<template>
  <nav class="sub-nav">
    <p v-if="label" class="sub-nav__label">{{ label }}</p>
    <NuxtLink
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="sub-nav__item"
      active-class="sub-nav__item--active"
    >
      <span class="sub-nav__icon">
        <UIcon :name="item.icon" class="size-3.5" />
      </span>
      <span class="truncate">{{ item.label }}</span>
    </NuxtLink>
  </nav>
</template>

<style scoped>
/* Col 2｜子頁導覽 — 依 redesign.md */
.sub-nav {
  width: 152px;
  flex-shrink: 0;
  align-self: stretch;
  background: #ffffff;
  border-right: 0.5px solid #e8e6fa;
  padding: 20px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sub-nav__label {
  font-size: 10px;
  color: #b4b2a9;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  font-weight: 500;
  margin-bottom: 10px;
}

.sub-nav__item {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  color: #888780;
  padding: 8px 10px;
  border-radius: 7px;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.sub-nav__item:hover {
  background: #f6f5ff;
  color: #3c3489;
}

.sub-nav__item--active {
  background: #eeedfe;
  color: #3c3489;
  font-weight: 500;
}

/* 子導覽項目的小 icon 色塊（16×16，border-radius 4px） */
.sub-nav__icon {
  display: grid;
  place-items: center;
  width: 16px;
  height: 16px;
  border-radius: 4px;
  background: #ffffff;
  color: #534ab7;
  box-shadow: inset 0 0 0 1px #e8e6fa;
  flex-shrink: 0;
}

.sub-nav__item--active .sub-nav__icon {
  color: #3c3489;
}

/* 深色模式 */
:global(.dark) .sub-nav {
  background: #14121f;
  border-right-color: #2e2a52;
}
:global(.dark) .sub-nav__item {
  color: #9b99a3;
}
:global(.dark) .sub-nav__item:hover {
  background: rgba(83, 74, 183, 0.12);
  color: #c9c4f0;
}
:global(.dark) .sub-nav__item--active {
  background: rgba(83, 74, 183, 0.18);
  color: #c9c4f0;
}
:global(.dark) .sub-nav__icon {
  background: #1a1730;
  color: #afa9ec;
  box-shadow: inset 0 0 0 1px #2e2a52;
}
:global(.dark) .sub-nav__item--active .sub-nav__icon {
  color: #c9c4f0;
}
</style>
