<script setup lang="ts">
import { EventsOn } from "../../wailsjs/runtime/runtime"
import { onBeforeUnmount, ref } from "vue"

type Operation = "download" | "delete"

type Update = {
    operation: Operation
    modelName: string
    progress: number
    complete: boolean
}

const updates = ref<Update[]>([])

function getUpdate(operation: Operation, modelName: string) {
    return updates.value.find((update) => update.operation === operation && update.modelName === modelName)
}

function removeUpdate(operation: Operation, modelName: string) {
    updates.value = updates.value.filter((update) => update.operation !== operation || update.modelName !== modelName)
}

function startUpdate(operation: Operation, modelName: string) {
    const existingUpdate = getUpdate(operation, modelName)
    if (existingUpdate) {
        existingUpdate.progress = 0
        existingUpdate.complete = false
        return
    }

    updates.value.push({ operation, modelName, progress: 0, complete: false })
}

function setProgress(operation: Operation, modelName: string, progress: number) {
    const update = getUpdate(operation, modelName)
    if (update) {
        update.progress = Math.max(0, Math.min(progress, 1))
    }
}

function completeUpdate(operation: Operation, modelName: string) {
    const update = getUpdate(operation, modelName)
    if (!update) {
        return
    }

    update.progress = 1
    update.complete = true
    setTimeout(() => {
        removeUpdate(operation, modelName)
    }, 3000)
}

const unsubscribeListeners = (['download', 'delete'] as const).flatMap((operation) => [
    EventsOn(`model-${operation}-started`, (modelName: string) => startUpdate(operation, modelName)),
    EventsOn(`model-${operation}-progress`, (modelName: string, progress: number) => setProgress(operation, modelName, progress)),
    EventsOn(`model-${operation}-completed`, (modelName: string) => completeUpdate(operation, modelName)),
])

onBeforeUnmount(() => {
    unsubscribeListeners.forEach((unsubscribe) => unsubscribe())
})
</script>

<template>
    <section id="progress-updates" aria-live="polite" aria-label="Model activity">
        <TransitionGroup name="progress-update" tag="div" class="updates-list">
            <article v-for="update in updates" :key="`${update.operation}-${update.modelName}`" class="progress-update" :class="{ complete: update.complete }">
                <div class="update-header">
                    <p>
                        <span class="update-status">
                            {{ update.complete ? (update.operation === 'download' ? 'Downloaded' : 'Deleted') : (update.operation === 'download' ? 'Downloading' : 'Deleting') }}
                        </span>
                        <span class="model-name" :title="update.modelName">{{ update.modelName }}</span>
                    </p>
                    <span class="progress-value">{{ Math.round(update.progress * 100) }}%</span>
                </div>
                <div class="progress-bar" :aria-label="`${update.modelName}: ${Math.round(update.progress * 100)}%`">
                    <div class="progress" :style="{ width: (update.progress * 100) + '%' }"></div>
                </div>
            </article>
        </TransitionGroup>
    </section>
</template>

<style scoped>
#progress-updates {
    width: 100%;
    box-sizing: border-box;
    padding: 10px;
    overflow-y: auto;
    max-height: 200px;
}

Article {
    width: calc(200px - 40px);
}

.updates-list {
    display: grid;
    gap: 10px;
}

.progress-update {
    border: 1px solid #454545;
    border-radius: 5px;
    padding: 8px;
    background-color: #202020;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.025);
}

.progress-update-enter-active,
.progress-update-leave-active,
.progress-update-move {
    transition: opacity 0.22s ease, transform 0.22s ease;
}

.progress-update-enter-from,
.progress-update-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}

.update-header {
    display: flex;
    align-items: start;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 7px;
}

.update-header p {
    margin: 0;
    min-width: 0;
    font-size: 0.75rem;
    line-height: 1.35;
    text-align: left;
}

.update-status {
    display: block;
    color: #9db9d8;
    font-size: 0.7rem;
    font-weight: 600;
}

.complete .update-status {
    color: #86c996;
}

.model-name {
    display: block;
    overflow: hidden;
    color: #ececec;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.progress-value {
    flex: 0 0 auto;
    padding-top: 2px;
    color: #9db9d8;
    font-size: 0.7rem;
    font-variant-numeric: tabular-nums;
}

.progress-bar {
    width: 100%;
    height: 4px;
    background-color: #3c3c3c;
    border-radius: 2px;
    overflow: hidden;
}

.progress {
    height: 100%;
    border-radius: inherit;
    background-color: #2185ff;
    transition: width 0.2s ease;
}

.complete .progress {
    background-color: #4ca964;
}
</style>