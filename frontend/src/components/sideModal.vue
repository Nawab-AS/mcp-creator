<script setup lang="ts">

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
    (e: 'close'): void
}>()

</script>

<template>
    <Teleport to="#modals">
        <Transition name="fade">
            <div v-if="props.open" id="modal-bg" @click.self="emit('close')"></div>
        </Transition>
        <Transition name="slide" appear>
            <div v-if="props.open" id="side-modal" @click.stop>
                <slot>No slot content</slot>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
#modal-bg {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: 1000;
}


#side-modal {
    position: fixed;
    top: 0;
    right: 0;
    width: 300px;
    height: calc(100vh - 40px);
    background-color: #242424;
    color: white;
    padding: 20px;
    border-radius: 20px 0 0 20px;
    z-index: 1001;
}


.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}


.slide-enter-active {
  transition: all 0.3s ease-out;
  transition-delay: 0.1s;
}

.slide-enter-from {
  transform: translateX(100%);
}


.slide-leave-active {
  transition: all 0.2s ease-in;
}

.slide-leave-to {
  transform: translateX(100%);
}
</style>