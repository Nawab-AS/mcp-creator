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
  return charCodeSum % 360
})

</script>

<template>
  <main>
    <div class="initials">
      <p>{{ Initials?.toUpperCase() }}</p>
    </div>
  </main>
</template>

<style scoped>

.initials {
  width: 30px;
  height: 30px;
  background-color: v-bind('"hsl(" + color + ", 100%, 50%)"');
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
