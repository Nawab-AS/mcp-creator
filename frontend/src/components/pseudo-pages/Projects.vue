<script setup lang="ts">
import { ref } from 'vue'
import SideModal from '../sideModal.vue'

// each project contains: star, name, path, last modified, model used
const projects = ref([
    {star: false, name: 'Project 1', path: '/path/to/project1', lastModified: '3 days ago', modelUsed: 'slow'},
    {star: true, name: 'Project 2', path: '/path/to/project2', lastModified: 'Less than a minute ago', modelUsed: 'fast'},
])

const settingsModal = ref({
    open: false,
    projectName: ''
})

function openOptions(projectName: string) {
    settingsModal.value.projectName = projectName
    settingsModal.value.open = true
}

</script>

<template>
    <div id="projects">
        <div id="header">
            <h1>Projects</h1>
            <span id="buttons">
                <button>Import</button>
                <button>Create</button>
            </span>
        </div>

        <table id="project-list">
            <colgroup>
                <col style="width: 8%;">
                <col style="width: 7%;">
                <col style="width: 31%;">
                <col style="width: 26%;">
                <col style="width: 20%;">
                <col style="width: 8%;">
            </colgroup>
            <thead>
                <tr id="table-header">
                    <th class="padding"></th>
                    <th><img src="../../assets/images/star.svg" alt="starred" class="star"></th>
                    <th>Name</th>
                    <th>Modified</th>
                    <th>Model</th>
                    <th class="padding"></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="project in projects" :key="project.name">
                    <td class="padding"></td>
                    <td>
                        <img src="../../assets/images/star.svg" alt="starred" 
                            :class="{ 'unstarred': !project.star, 'star': true }"
                            @click="project.star = !project.star">
                    </td>
                    <td class="project-name">
                        {{ project.name }} <br/> <code>{{ project.path }}</code>
                    </td>
                    <td>{{ project.lastModified }}</td> <!-- TODO: show friendly dates -->
                    <td>{{ project.modelUsed }}</td>
                    <td @click="openOptions(project.name)">
                        <img src="../../assets/images/options.svg" alt="options" class="options">
                    </td>
                </tr>
            </tbody>
        </table>

        <!-- settings modal -->
        <SideModal :open="settingsModal.open" @close="settingsModal.open = false">
            <h2>Settings for {{ settingsModal.projectName }}</h2>
            <p>Here you can configure settings for the project.</p>

        </SideModal>
    </div>
</template>

<style scoped>
#header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 20px 15px;
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
    cursor: pointer;
    transition-duration: 0.5s;
}

#header button:hover {
    filter: brightness(1.1);
}

#header button:active {
    filter: brightness(0.9);
}

table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0 10px;
    table-layout: fixed;
}

th, td {
    text-align: left;
    padding: 5px 10px;
}

th {
    border: 1px solid #454545;
    background-color: #242424;
    font-size: 1rem;
}

th.padding {
    border-right: none;
    border-left: none;
}

tr {
    transition-duration: 0.3s;
}

td {
    font-size: 0.95rem;
}

tbody tr:hover {
    background-color: #2a2a2a;
}


.star {
    transition-duration: 0.2s;
    scale: 0.8;
}

td>img.star.unstarred {
    opacity: 0;
}

td>img.star:hover {
    opacity: 0.9;
}

td>img.star.unstarred:hover {
    opacity: 0.4;
}

td.project-name {
    word-break: break-all;
    font-size: 1rem;
}

td.project-name code {
    font-size: 0.8rem;
    color: #888;
}


</style>