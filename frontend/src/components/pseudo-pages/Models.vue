<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import Sections from '../Selector.vue'

// backend imports
import { GetModels, DownloadModel } from '../../../wailsjs/go/main/App'
import type { backend } from '../../../wailsjs/go/models'


// fetch models list
const models = ref<backend.Model[]>([])
onMounted(async () => {
    models.value = await GetModels()
})


const filter = ref('')
const selectedModel = ref('')
const visibleModels = computed(() => models.value.filter((item) => 
    item.installed === (filter.value === 'Installed')
))
watch(filter, (_) => { selectedModel.value = '' })


function selectModel(name: string) {
    if (selectedModel.value === name) {
        selectedModel.value = ''
        return
    }
    selectedModel.value = name
}
async function modifyModel() {
    await DownloadModel(selectedModel.value)
}
</script>

<template>
    <div id="models" @click="selectedModel=''">
        <div id="header">
            <h1>Models</h1>
            <span>
                <button>Import</button>
                
                <button @click="modifyModel" :disabled="!selectedModel" id="modify">
                    <span v-if="filter == 'Installed'">Delete</span>
                    <span v-else>Install</span>
                </button>
            </span>
        </div>
        <Sections :options="['Installed', 'Available']" v-model="filter"/>
        <div id="model-list">
            <div
                v-for="m in visibleModels"
                :key="m.name"
                class="model-card"
                :class="{ selected: selectedModel === m.name }"
                @click.stop="selectModel(m.name)"
            >
                <h3>{{ m.name }}</h3>
                <p>{{ m.description }}</p>
            </div>
        </div>
    </div>
</template>

<style scoped>
#header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 10px 10px;
}

#header>h1 {
    font-size: 2rem;
    margin: 0;
}

#header button {
    margin-left: 10px;
    padding: 5px 10px;
    font-size: 1rem;
    border: none;
    border-radius: 5px;
    background-color: #454545;
    color: white;
    cursor: pointer;
    transition-duration: 0.5s;
}

#header button#modify {
    background-color: v-bind('filter === "Installed" ? "#b71c1c" : "#004f94"');
}

#header button:hover {
    filter: brightness(1.1);
}

#status {
    margin: 0 10px 8px;
    color: #b7c6e9;
    font-size: 0.95rem;
}


#model-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 10px;
    padding: 10px;
}

.model-card {
    border: 1px solid #444;
    border-radius: 10px;
    padding: 10px;
    height: 150px;
    width: 100%;
    box-sizing: border-box;
    transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease;
    cursor: pointer;
}

.model-card.selected {
    border-color: #004f94;
    background-color: rgba(0, 79, 148, 0.15);
}

.model-card h3 {
    margin: 0;
    font-size: 1.2rem;
    text-align: left;
}

.model-card p {
    text-align: left;
    margin: 5px 0 0 0;

    display: -webkit-box;
    font-size: 0.9rem;
    line-height: 20px;
    line-clamp: 4;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
    overflow: hidden;
}

</style>
