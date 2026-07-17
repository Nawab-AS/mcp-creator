<script setup lang="ts">
import { ref } from 'vue'

// each project contains: star, name, path, last modified, model used
const projects = ref([
    {star: false, name: 'Project 1', path: '/path/to/project1', lastModified: '2024-06-01 13:34', modelUsed: 'slow'},
    {star: true, name: 'Project 2', path: '/path/to/project2', lastModified: '2024-06-02 14:35', modelUsed: 'fast'},
])

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
            <thead>
                <tr>
                    <th class="padding"></th>
                    <th><img src="../../assets/images/star.svg" alt="starred"></th>
                    <th><h3>Name</h3></th>
                    <th><h3>Modified</h3></th>
                    <th><h3>Model</h3></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="project in projects" :key="project.name">
                    <td class="padding"></td>
                    <td>
                        <img src="../../assets/images/star.svg" alt="starred">
                    </td>
                    <td>
                        <p>{{ project.name }}</p>
                        <p>{{ project.path }}</p>
                    </td>
                    <td>{{ project.lastModified }}</td> <!-- TODO: show friendly dates -->
                    <td>{{ project.modelUsed }}</td>
                </tr>
            </tbody>
        </table>
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
    cursor: pointer;
    transition-duration: 0.5s;
}

#header button:hover {
    filter: brightness(1.1);
}

#header button:active {
    filter: brightness(0.9);
}

h3, p {
    margin: 0;
}

th, td {
    padding: 10px;
    text-align: left;
    border-left: 2px solid #454545;
}

th.padding, td.padding {
    border-left: none;
}

</style>