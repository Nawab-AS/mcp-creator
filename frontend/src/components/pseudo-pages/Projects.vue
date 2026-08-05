<script setup lang="ts">
import { ref, onMounted, toRaw, watch } from 'vue'
import SideModal from '../sideModal.vue'

import { GetProjects, GetModels, ModifyProject, CreateProject, DeleteProject as DeleteProjectFunc, ReindexProject as ReindexProjectFunc, SelectDirDialog, GetAvailablePort } from '../../../wailsjs/go/main/App'
import type { backend } from '../../../wailsjs/go/models'

const projects = ref<backend.Project[]>([])
const models = ref<string[]>([])

onMounted(async () => await refreshProjects(false))

async function refreshProjects(soft=true) {
    projects.value = await GetProjects()
    models.value = await GetModels().then(models => models.filter(m => m.installed).map(m => m.name))
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
    type: 'Modify' as 'Modify' | 'Create',
    error: '',
    project: null as backend.Project | null,
})

watch(
    () => [
        settingsModal.value.project?.name,
        settingsModal.value.project?.path,
        settingsModal.value.project?.modelUsed,
        settingsModal.value.project?.port,
    ],
    () => {
        settingsModal.value.error = ''
    },
)

watch(
    () => settingsModal.value.open,
    (open) => {
        if (!open) {
            // allow the SideModal component to fully unmount first (avoid blank modal)
            setTimeout(() => {
                settingsModal.value = {
                    open: false,
                    initialName: '',
                    type: 'Modify',
                    error: '',
                    project: null,
                }
            }, 200)
        }
    },
)

watch(
    () => settingsModal.value.project?.port,
    (port) => {
        port = Number(port)
        if (Number.isNaN(port)) {
            settingsModal.value.project!.port = 1024
            return
        }
        if (port < 1024) { settingsModal.value.project!.port = 1024 }
        if (port > 49150) { settingsModal.value.project!.port = 49150 }
    },
)

async function toggleStar(projectName: string) {
    const project = projects.value.find(p => p.name === projectName)
    if (!project) return
    if (project) {
        project.star = !project.star
    }
    const result = await ModifyProject(projectName, 'star', project.star)
    if (result.statusCode !== 200) {
        console.error(`Failed to modify project: ${result.message}`)
    }
}

function openOptions(projectName: string) {
    const project = projects.value.find(p => p.name === projectName)
    if (!project) return

    settingsModal.value = {
        open: true,
        initialName: project.name,
        type: 'Modify',
        error: '',
        project: structuredClone(toRaw(project)) as backend.Project, // deep copy
    }
}

async function updateModifiedProjects(cancelled: boolean = false) {
    if (cancelled) {
        settingsModal.value.open = false
        return
    }

    if (!settingsModal.value.project) return true;

    if (settingsModal.value.type === 'Create') {
        const result = await CreateProject(
            settingsModal.value.project.name,
            settingsModal.value.project.path,
            settingsModal.value.project.modelUsed,
            settingsModal.value.project.port
        )

        if (result.statusCode !== 201) {
            settingsModal.value.error = result.message
            return true
        }

        settingsModal.value.open = false
        await refreshProjects(false)
        return

    } else if (settingsModal.value.type === 'Modify') {
        let project = projects.value.find(p => p.name === settingsModal.value.initialName)

        if (!project || !settingsModal.value.project) { // unlikely but just in case
            settingsModal.value.open = false
            return
        }

        const updatedProject = settingsModal.value.project

        // if the name changed, update first to avoid conflicts
        if (project.name !== updatedProject.name) {
            const result = await ModifyProject(project.name, 'name', updatedProject.name)
            if (result.statusCode !== 200) {
                settingsModal.value.error = result.message
                return true
            }
            project.name = updatedProject.name
        }

        const keys = Object.keys(updatedProject) as (keyof backend.Project)[]
        for (const key of keys) {
            if (project[key] !== updatedProject[key]) {
                const result = await ModifyProject(project.name, key, updatedProject[key])
                if (result.statusCode !== 200) {
                    settingsModal.value.error = result.message
                    return true
                }
            }
        }
        settingsModal.value.open = false
        await refreshProjects(false)
    }
}

async function selectProjectDirectory(title: string) {
    if (!settingsModal.value.project) return

    const selectedDirectory = await SelectDirDialog(title)
    if (selectedDirectory) {
        settingsModal.value.project.path = selectedDirectory
    }
}

async function DeleteProject(projectName: string) {
    const result = await DeleteProjectFunc(projectName)
    if (result.statusCode !== 200) {
        settingsModal.value.error = result.message
        return
    }
    settingsModal.value.open = false
    await refreshProjects(false)
}

async function ReindexProject(projectName: string) {
    settingsModal.value.open = false
    ReindexProjectFunc(projectName)
    await refreshProjects(false)
}

async function createProjectModal() {
    let project: backend.Project = {
        name: '',
        path: '',
        star: false,
        port: await GetAvailablePort(),
        lastModified: new Date().toISOString(),
        modelUsed: '',
    }

    settingsModal.value = {
        open: true,
        initialName: project.name,
        type: 'Create',
        error: '',
        project,
    }
}

</script>

<template>
    <div id="projects">
        <div id="header">
            <h1>Projects</h1>
            <span id="buttons">
                <button class="create" @click="createProjectModal()">Create</button>
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
                        {{ project.name }} <br/>
                        <p class="path"> <span> <code>/{{ project.path }}</code> </span> </p>
                    </td>
                    <td>{{ friendlyDate(project.lastModified) }}</td>
                    <td>{{ project.modelUsed }}</td>
                    <td @click="openOptions(project.name)">
                        <img src="../../assets/images/options.svg" alt="options" class="options">
                    </td>
                </tr>
            </tbody>
        </table>
        <p v-if="projects.length === 0" id="no-projects">No projects?<br />Create one!</p>

        <!-- settings modal -->
        <SideModal :open="settingsModal.open" :close="updateModifiedProjects">
            <h2>{{ settingsModal.type == 'Create' ? 'Create Project' : 'Modify Project' }}</h2>
            <br />
            <div v-if="settingsModal.project" id="project-settings">
                <label for="project-name">Project Name</label>
                <input type="text" id="project-name" v-model="settingsModal.project.name" placeholder="My New Project" maxlength="20"/>
                <p class="error" v-if="settingsModal.error.startsWith('name: ')">{{ settingsModal.error.slice('name: '.length) }}</p>
                <br /><br />

                <label for="project-path">Project Folder</label>
                <div id="project-path">
                    <input type="text" v-model="settingsModal.project.path" readonly placeholder="Select a project folder"/>
                    <button type="button" @click="selectProjectDirectory('Select your project folder')">Browse</button>
                </div>
                <p class="error" v-if="settingsModal.error.startsWith('path: ')">{{ settingsModal.error.slice('path: '.length) }}</p>
                <br />

                <div id="project-model">
                    <label for="project-model">Model</label>
                    <select v-model="settingsModal.project.modelUsed">
                        <option default value="" disabled>Select a model</option>
                        <option v-for="model in models" :key="model" :value="model">{{ model }}</option>
                    </select>
                </div>
                <p class="error" v-if="settingsModal.error.startsWith('model: ')">{{ settingsModal.error.slice('model: '.length) }}</p>
                <br />

                <details id="project-advanced">
                    <summary>Advanced Options</summary>
                    <br />
                    <div id="project-port">
                        <span>
                            <label for="project-port">Port</label>
                            <input type="number" id="project-port" v-model.number="settingsModal.project.port"/>
                        </span>
                        <p class="error" v-if="settingsModal.error.startsWith('port: ')">{{ settingsModal.error.slice('port: '.length) }}</p>
                    </div>
                </details>
                <br /><br />
                <span v-if="settingsModal.type === 'Modify'" id="modal-actions">
                    <button type="button" id="reindex" @click="ReindexProject(settingsModal.project.name)">Reindex</button>
                    <button type="button" id="delete" @click="DeleteProject(settingsModal.project.name)">Delete</button>
                </span>
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

#header button.create {
    background-color: #328435;
    color: white;
}

#project-settings {
    width: 300px;
}

#project-name {
    outline: none;
    border: 1px solid #454545;
    border-radius: 5px;
    padding: 5px 10px;
    width: calc(100% - 30px);
    margin-top: 10px;
    font-size: 0.8rem;
    color: white;
    background-color: #242424;
}

#project-path {
    display: flex;
    gap: 5px;
    margin-top: 5px;
}

#project-path input {
    flex: 1;
    outline: none;
    direction: rtl;
    text-align: left;
    padding: 5px 12px;
    border: 1px solid #454545;
    border-radius: 5px;
    background-color: #242424;
    color: white;
    font-size: 0.9rem;
}

#project-path button {
    padding: 5px 12px;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    background-color: #454545;
    color: white;
    font-size: 0.9rem;
}

#project-model {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

#project-model select {
    outline: none;
    border: 1px solid #454545;
    border-radius: 5px;
    padding: 5px 10px;
    width: calc(100% - 30px);
    font-size: 0.8rem;
    color: white;
    background-color: #242424;
}

p.error {
    color: #b71c1c;
    font-size: 0.8rem;
    margin: 5px 0 0 0;
}

#project-advanced {
    margin-top: 10px;
    border: 1px solid #454545;
    border-radius: 5px;
    padding: 10px;
}

#project-port label, #project-port p.error {
    margin-left: 20px;
}

#project-port input {
    margin-left: 10px;
    outline: none;
    border: 1px solid #454545;
    border-radius: 5px;
    padding: 5px 10px;
    width: 60px;
    text-align: center;
    font-size: 0.8rem;
    color: white;
    background-color: #242424;
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
    font-size: 1rem;
}

td.project-name>p.path {
    margin: 0;
    font-size: 0.8rem;
    color: #888;
    max-width: min-content;

    display: flex;
    flex-direction: row-reverse;
    overflow: hidden;
    white-space: nowrap;
    -webkit-mask-image: linear-gradient(to right, transparent 0%, #000 15px);
    mask-image: linear-gradient(to right, transparent 0%, #000 15px);
}


#no-projects {
    text-align: center;
    font-size: 1.2rem;
    color: #888;
    margin-top: 20vh;
}

#modal-actions {
    display: flex;
    gap: 10px;
}

#modal-actions > button {
	padding: 5px 10px;
	font-size: 1rem;
	border: none;
	border-radius: 5px;
	cursor: pointer;
	transition-duration: 0.5s;
}

button#reindex {
    background-color: #454545;
    color: white;
}

button#delete {
    background-color: #b71c1c;
    color: white;
}

</style>