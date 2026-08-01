<script setup lang="ts">
import { EventsOn } from "../../../wailsjs/runtime/runtime";

import { computed, onMounted, ref, watch } from 'vue'
import Sections from '../Selector.vue'

// backend imports
import { GetModels, DownloadModel, DeleteModel } from '../../../wailsjs/go/main/App'
import type { backend } from '../../../wailsjs/go/models'


// fetch models list
const models = ref<backend.Model[]>([])
const bufferedModels = ref(new Map<string, "Downloading" | "Deleting" | "Available" | "Installed">());
onMounted(async () => {
    models.value = await GetModels()
})

EventsOn("model-download-started", (modelName: string) => {
    bufferedModels.value.set(modelName, 'Downloading')
})
EventsOn("model-download-completed", (modelName: string) => {
    bufferedModels.value.set(modelName, 'Installed')
})
EventsOn("model-delete-started", (modelName: string) => {
    bufferedModels.value.set(modelName, 'Deleting')
})
EventsOn("model-delete-completed", (modelName: string) => {
    bufferedModels.value.set(modelName, 'Available')
})


const filter = ref('')
const visibleModels = computed(() => models.value.filter((item) => 
    item.installed === (filter.value === 'Installed')
))

watch(filter, (_) => { // collapse bufferedModels
    for (const [modelName, status] of bufferedModels.value.entries()) {
        const model = models.value.find((m) => m.name === modelName)
        if (!model) continue;
        if (status.endsWith('ing')) continue;
        model.installed = bufferedModels.value.get(modelName) === 'Installed'
        bufferedModels.value.delete(modelName)
    }
})


async function modifyModel(name: string) {
    if (bufferedModels.value.get(name) != undefined) {
        const status = bufferedModels.value.get(name)
        if (!status || status.endsWith('ing')) return;

        if (status === 'Installed') {
            await DeleteModel(name)
        } else if (status === 'Available') {
            await DownloadModel(name)
        }
        return
    }

    if (filter.value === 'Installed') {
        await DeleteModel(name)
    } else {
        await DownloadModel(name)
    }
}
</script>

<template>
    <div id="models">
        <div id="header">
            <h1>Models</h1>
            <!-- <span>
                <button>Import</button>
            </span> -->
        </div>
        <Sections :options="['Available', 'Installed']" v-model="filter"/>
        <div id="model-list">
            <div
                v-for="m in visibleModels"
                :key="m.name"
                class="model-card"
            >
                <h3>{{ m.name }}</h3>
                <p>{{ m.description }}</p>
                <div class="model-meta">
                    <code v-if="m.size_mb">{{ m.size_mb }} MB</code>
                    <button
                        class="model-action"
                        :class="{ destructive: bufferedModels.get(m.name) === 'Installed' || (bufferedModels.get(m.name) === undefined && filter === 'Installed') }"
                        @click="modifyModel(m.name)"
                        :disabled="(bufferedModels.get(m.name) ?? '').endsWith('ing')"
                    >
                        <span v-if="bufferedModels.get(m.name)?.endsWith('ing')">{{ bufferedModels.get(m.name) }}...</span>
                        <span v-else-if="bufferedModels.has(m.name)">{{ bufferedModels.get(m.name) === 'Installed' ? 'Delete' : 'Install' }}</span>
                        <span v-else>{{ filter === 'Installed' ? 'Delete' : 'Download' }}</span>
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
#models {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
}

#header {
    /* flex: 0 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between; */
    margin: 20px 15px;
}

#header>h1 {
    font-size: 2rem;
    margin: 0;
}

/* #header button {
    margin-left: 10px;
    padding: 5px 10px;
    font-size: 1rem;
    border: none;
    border-radius: 5px;
    background-color: #454545;
    color: white;
    cursor: pointer;
    transition-duration: 0.5s;
} */

#status {
    margin: 0 10px 8px;
    color: #b7c6e9;
    font-size: 0.95rem;
}


#model-list {
    flex: 1 1 auto;
    min-height: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 12px;
    padding: 16px;
    overflow-y: auto;
    align-content: start;
}

.model-card {
    display: flex;
    flex-direction: column;
    min-height: 186px;
    padding: 14px;
    width: 100%;
    box-sizing: border-box;
    border: 1px solid #414141;
    border-radius: 6px;
    background-color: #242424;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.025);
    transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
    cursor: pointer;
}

.model-card:hover {
    border-color: #5c5c5c;
    background-color: #292929;
    transform: translateY(-1px);
}

.model-card h3 {
    margin: 0;
    min-height: 2.6rem;
    color: #f3f6fb;
    font-size: 1rem;
    font-weight: 600;
    line-height: 1.3;
    text-align: left;
    overflow-wrap: anywhere;
}

.model-card p {
    text-align: left;
    margin: 10px 0 0;
    color: #bdbdbd;
    display: -webkit-box;
    font-size: 0.875rem;
    line-height: 1.45;
    line-clamp: 3;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
}

.model-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    min-height: 26px;
    margin-top: auto;
    padding-top: 12px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.model-meta code {
    padding: 3px 7px;
    border: 1px solid #4e4e4e;
    border-radius: 3px;
    background-color: #1b1b1b;
    color: #c8d8ee;
    font-family: inherit;
    font-size: 0.75rem;
    font-weight: 600;
}

.model-action {
    padding: 5px 9px;
    border: 1px solid #176dcf;
    border-radius: 4px;
    background-color: #004f94;
    color: #fff;
    font: inherit;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s ease, border-color 0.2s ease;
}

.model-action:hover {
    border-color: #298fff;
    background-color: #176dcf;
}

.model-action.destructive {
    border-color: #8f2424;
    background-color: #6d1d1d;
}

.model-action.destructive:hover {
    border-color: #d14343;
    background-color: #9f2929;
}

.model-action:disabled {
    border-color: #4e4e4e;
    background-color: #3c3c3c;
    color: #9b9b9b;
    cursor: not-allowed;
}

@media (max-width: 560px) {
    #header {
        margin: 16px 12px;
    }

    #header>h1 {
        font-size: 1.6rem;
    }

    #model-list {
        grid-template-columns: minmax(0, 1fr);
        padding: 12px;
    }
}

</style>
