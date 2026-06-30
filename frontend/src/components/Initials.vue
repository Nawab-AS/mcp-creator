<script lang="ts" setup>
import { computed } from 'vue'

const props = defineProps({
  Initials: String,
})

const color = computed(() => {
  if (!props.Initials) return '#000000'
  const charCodeSum = props.Initials
    .split('')
    .reduce((sum: number, char: string) => sum + char.charCodeAt(0)*100, 0)
  const hue = charCodeSum % 360
  return `hsl(${hue}, 70%, 50%)`
})

</script>

<template>
  <main>
    <div
      class="initials"
      :style="{ backgroundColor: color }"
    >
      <p>{{ Initials?.toUpperCase() }}</p>
    </div>
  </main>
</template>

<style scoped>

.initials {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease-in-out;
  transition-delay: 0.1s;
}

p {
  margin: 0;
  padding: 0;
  font-size: 1rem;
  font-weight: bolder;
  color: white;
  user-select: none;
}

.initials:hover {
  filter: brightness(1.1);
  scale: 1.1;
  cursor: pointer;
}
</style>
