<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  open: boolean,
  close: (bool?: boolean) => boolean | void | Promise<boolean | void>,
  // close(<cancelled>) --> should close?
}>()


// forced open flash
const flash_duration = 200
const forcedOpen = ref(false)
async function close(cancelled?: boolean) {
   forcedOpen.value = await props.close?.(cancelled) || false

   // flash `warning`
   if (!forcedOpen.value) return;
   setTimeout(() => {
	   forcedOpen.value = false
   }, flash_duration/2)

}
</script>

<template>
    <Teleport to="#modals">
        <Transition name="fade">
            <div v-if="props.open" id="modal-bg" @click.self="close()"></div>
        </Transition>
        <Transition name="slide" appear>
            <div v-if="props.open" id="side-modal" @click.stop :class="{ flash: forcedOpen }">
                <slot>No slot content</slot>
				<span>
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
	transition-duration: v-bind(flash_duration/2 + 'ms');
}

#side-modal.flash {
	border-color: #ff0000;
	background-color: #ff0000;
}


#side-modal button {
	margin: 10vh 10px 0 0;
	padding: 5px 10px;
	font-size: 1rem;
	border: none;
	border-radius: 5px;
	cursor: pointer;
	transition-duration: 0.5s;
}

#side-modal button.save {
	background-color: #4CAF50;
	color: white;
}

#side-modal button.cancel {
	background-color: #f44336;
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