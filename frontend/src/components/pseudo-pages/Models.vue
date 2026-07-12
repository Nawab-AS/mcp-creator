<script setup lang="ts">
import { reactive } from 'vue'

const model = reactive([
    { name: 'slow', description: 'Fast on older hardware and simple datasets.' },
    { name: 'balanced', description: 'Good for modern hardware and balanced performance.' },
    { name: 'accurate', description: 'Good on powerful hardware and complex datasets.' },
    { name: 'fast', description: 'Fast on modern hardware and simple datasets.' }
])
</script>

<template>
    <div id="models">
        <div id="header">
            <h1>Models</h1>
            <span>
                <button>Import</button>
                <button>Install</button>
            </span>
        </div>
        <div id="model-list">
            <div v-for="m in model" :key="m.name" class="model-card">
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
    margin: 10px 0;
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
    transition: all 0.2s ease-in-out;
}

#header button:last-child {
    background-color: #004f94;
}

#header button:hover {
    filter: brightness(1.1);
}


#model-list {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 10px;
}

.model-card {
    border: 1px solid #444;
    border-radius: 10px;
    padding: 10px;
    height: 150px;
    width: 150px;
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
