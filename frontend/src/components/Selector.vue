<!-- /components/Selector.vue -->
<script setup lang="ts">
import { reactive, onMounted } from 'vue'


// consts
const leftIndent = 20;
const barWidth = 100;
const speed = 5;
const colors = {
  bar: {
    selected: '#2185ff',
    unselected: '#242424',
  },
  text: {
    selected: '#fff',
    selectedBg: '#161616',
    unselected: '#fff',
  }
}


// Define props matching modelValue (default for v-model) and options
const props = defineProps<{
  modelValue: string
  options: string[],
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()



// onclick animation
const bar = reactive({
  left: leftIndent,
  current_left: 0,
})


function onClick(option: string, now: boolean = false) {
  // do other stuff...
  emit('update:modelValue', option)
  const index = props.options.indexOf(option)
  bar.left = leftIndent + index * (barWidth + 32) + 1
  if (now) {
    bar.current_left = bar.left
  }
}


onMounted(() => {
  onClick(props.options[0], true)
})

// move bar
setInterval(() => {
  if (Math.abs(bar.current_left - bar.left) <= speed) {
    bar.current_left = bar.left
  } else if (bar.current_left < bar.left) {
    bar.current_left += speed
  } else if (bar.current_left > bar.left) {
    bar.current_left -= speed
  }
}, 10)

</script>

<template>
  <div class="selector-buttons">
    <p 
      v-for="option in options" 
      :key="option"
      :class="{ active: modelValue === option }"
      @click="onClick(option)"
    >
      {{ option }}
  </p>
  </div>
</template>

<style scoped>

.selector-buttons {
  background-color: v-bind('colors.bar.unselected');
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  gap: 0px;
  padding: 0px;
  padding-left: v-bind('leftIndent + "px"');

  border-bottom: 2px solid transparent; 
  border-image: v-bind('`linear-gradient(to right, ${colors.bar.selected} ${bar.current_left}px, ${colors.bar.unselected} ${bar.current_left}px, ${colors.bar.unselected} ${bar.current_left + barWidth + 32}px, ${colors.bar.selected} ${bar.current_left + barWidth + 32}px)`') 1;
}


.active {
  color: v-bind('colors.text.selected');
  background-color: v-bind('colors.text.selectedBg');
}

p {
  padding: 8px 16px;
  margin: 0px;
  cursor: pointer;
  color: v-bind('colors.text.unselected');
  width: v-bind('barWidth + "px"');
  text-align: center;
}

</style>
