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
    background-color: rgba(0, 0, 0, 0.5);
    z-index: 1000;
}

#slot {
  height: 420px;
  width: 300px;

  overflow-y: auto;
  overflow-x: hidden;
}

#side-modal {
    position: fixed;
    top: -1px;
    right: 0;
    width: 300px;
    height: calc(100vh - 40px - 2px);
    background-color: #242424;
    color: white;
    padding: 20px;
    border: 1px solid transparent;
    border-radius: 20px 0 0 20px;
    z-index: 1001;
}

#side-modal.flash {
  animation: warning-flash v-bind(flashDuration + 'ms') ease-in-out;
}

@keyframes warning-flash {
  0%, 100% {
    background-color: #242424;
    border-color: transparent;
  }
  45%, 65% {
    background-color: #6b2424;
    border-color: #e15a5a;
  }
}


#modal-actions {
  position: absolute;
  bottom: 20px;
  left: 20px;
  display: flex;
  gap: 10px;
}

#modal-actions button {
	padding: 5px 10px;
	font-size: 1rem;
	border: none;
	border-radius: 5px;
	cursor: pointer;
	transition-duration: 0.5s;
}

#side-modal button.save {
	background-color: #328435;
	color: white;
}

#side-modal button.cancel {
	background-color: #b71c1c;
	color: white;
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