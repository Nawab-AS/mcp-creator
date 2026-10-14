<script setup lang="ts">
import { nextTick, ref } from 'vue'

const props = defineProps<{
  open: boolean,
  close: (bool?: boolean) => boolean | void | Promise<boolean | void>,
  // close(<cancelled>) --> should close?
}>()


const flashDuration = 250
const forcedOpen = ref(false)
let flashTimeout: ReturnType<typeof setTimeout> | undefined

async function close(cancelled?: boolean) {
  const keepOpen = await props.close?.(cancelled) || false
  if (!keepOpen) return

  clearTimeout(flashTimeout)
  forcedOpen.value = false
  await nextTick()
  forcedOpen.value = true
  flashTimeout = setTimeout(() => {
    forcedOpen.value = false
  }, flashDuration)

}
</script>

<template>
    <Teleport to="#modals">
        <Transition name="fade">
            <div v-if="props.open" id="modal-bg" @click.self="close()"></div>
        </Transition>
        <Transition name="slide" appear>
            <div v-if="props.open" id="side-modal" @click.stop :class="{ flash: forcedOpen }">
              <div id="slot">
                <slot>No slot content</slot>
              </div>
              <span id="modal-actions">
                <button @click.stop="close()" class="save">Save</button>
                <button @click.stop="close(true)" class="cancel">Cancel</button>
              </span>
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
    background-color: rgba(0, 0, 0, 0.58);
    z-index: 1000;
}

#slot {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;

  overflow-y: auto;
  overflow-x: hidden;
}

#side-modal {
    position: fixed;
    top: 0;
    right: 0;
    width: 300px;
    height: 100vh;
    background-color: var(--surface);
    color: var(--text-primary);
    padding: 20px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--border);
    border-radius: var(--radius-surface) 0 0 var(--radius-surface);
    z-index: 1001;
}

#side-modal.flash {
  animation: warning-flash v-bind(flashDuration + 'ms') ease-in-out;
}

@keyframes warning-flash {
  0%, 100% {
    background-color: var(--surface);
    border-color: transparent;
  }
  45%, 65% {
    background-color: var(--danger);
    border-color: #d14343;
  }
}


#modal-actions {
  flex: 0 0 auto;
  margin-top: 16px;
  display: flex;
  gap: 8px;
}

#modal-actions button {
  padding: 6px 10px;
  font-size: 0.8rem;
  font-weight: 600;
  border: 1px solid transparent;
  border-radius: var(--radius-control);
	cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

#side-modal button.save {
  background-color: var(--accent-deep);
  border-color: var(--accent);
	color: white;
}

#side-modal button.cancel {
  background-color: #3c3c3c;
  border-color: var(--border-strong);
  color: var(--text-primary);
}


.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}


.slide-enter-active {
  transition: all 0.2s ease-out;
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