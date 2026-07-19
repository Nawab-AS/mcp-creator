<script setup lang="ts">
import { ref, onMounted, toRaw } from 'vue'
import SideModal from '../sideModal.vue'

import { GetProjects, ModifyProject } from '../../../wailsjs/go/main/App'
import type { backend } from '../../../wailsjs/go/models'

const projects = ref<backend.Project[]>([])

onMounted(async () => await refreshProjects(false))

async function refreshProjects(soft=true) {
    projects.value = await GetProjects()
    if (soft) return
    projects.value.sort((a, b) => {
        if (a.star && !b.star) return -1
        if (!a.star && b.star) return 1
        return a.name.localeCompare(b.name)
    })
}

function friendlyDate(dateString: string) {
    const date = new Date(dateString)
    const now = new Date()
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000)

    if (diffInSeconds < 60) {
        return `Less than a minute ago`
    } else if (diffInSeconds < 60*60) {
        return `${Math.floor(diffInSeconds / 60)} minutes ago`
    } else if (diffInSeconds < 60*60*24) {
        return `${Math.floor(diffInSeconds / 3600)} hours ago`
    } else if (diffInSeconds < 60*60*24*7) {
        return `${Math.floor(diffInSeconds / 86400)} days ago`
    } else if (diffInSeconds < 60*60*24*7*4) {
        return `${Math.floor(diffInSeconds / 604800)} weeks ago`
    } else if (diffInSeconds < 60*60*24*7*4*12) {
        return `${Math.floor(diffInSeconds / 2592000)} months ago`
    } else {
        return date.toLocaleDateString()
    }
}

const settingsModal = ref({
    open: false,
    initialName: '',
    project: null as backend.Project | null,
})

async function toggleStar(projectName: string) {
    const project = projects.value.find(p => p.name === projectName)
    if (!project) return
    if (project) {
        project.star = !project.star
    }
    await ModifyProject(projectName, 'star', project.star)
}

function openOptions(projectName: string) {
    const project = projects.value.find(p => p.name === projectName)
    if (!project) return

    settingsModal.value = {
        open: true,
        initialName: project.name,
        project: structuredClone(toRaw(project)) as backend.Project, // deep copy
    }
}

async function updateModifiedProjects() {
    let project = projects.value.find(p => p.name === settingsModal.value.initialName)

    if (!project || !settingsModal.value.project) { // unlikely but just in case
        settingsModal.value = { open: false, initialName: '', project: null }
        return
    }


    const updatedProject = settingsModal.value.project
    updatedProject.name = updatedProject.name.trim()

    // if the name changed, update first to avoid conflicts
    if (project.name !== updatedProject.name) {
        await ModifyProject(project.name, 'name', updatedProject.name)
        project.name = updatedProject.name
    }

    const keys = Object.keys(updatedProject) as (keyof backend.Project)[]
    for (const key of keys) {
        if (project[key] !== updatedProject[key]) {
            await ModifyProject(project.name, key, updatedProject[key])
        }
    }
    settingsModal.value = { open: false, initialName: '', project: null }
    await refreshProjects(false)
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
                    <td @click="toggleStar(project.name)">
                        <img src="../../assets/images/star.svg" alt="starred" 
                            :class="{ 'unstarred': !project.star, 'star': true }">
                    </td>
                    <td class="project-name">
                        {{ project.name }} <br/> <code>{{ project.path }}</code>
                    </td>
                    <td>{{ friendlyDate(project.lastModified) }}</td>
                    <td>{{ project.modelUsed }}</td>
                    <td @click="openOptions(project.name)">
                        <img src="../../assets/images/options.svg" alt="options" class="options">
                    </td>
                </tr>
            </tbody>
        </table>

        <!-- settings modal -->
        <SideModal :open="settingsModal.open" @close="updateModifiedProjects()">
            <h2>Settings</h2>
            <br />
            <div v-if="settingsModal.project">
                <label for="project-name">Project Name:</label>
                <input type="text" id="project-name" v-model="settingsModal.project.name" />
                <br /><br />
                <label for="project-path">Project Path:</label>
                <input type="text" id="project-path" v-model="settingsModal.project.path" />
                <br /><br />
                <label for="project-model">Model Used:</label>
                <input type="text" id="project-model" v-model="settingsModal.project.modelUsed" />
            </div>
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

#header>h1, h2 {
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


.star, .options {
    transition-duration: 0.2s;
}

.star {
    scale: 0.8;
}

.star.unstarred, .options {
    opacity: 0;
}

td:hover>img.star {
    opacity: 0.9;
}

td:hover>img.star.unstarred, tr:hover>td>img.options {
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