<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { GetProjects, CopyToClipboard } from '../../../wailsjs/go/main/App'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import type { backend } from '../../../wailsjs/go/models'

const projects = ref<backend.Project[]>([])
const loading = ref(true)
const error = ref('')

const runningProjects = computed(() => projects.value.filter(project => project.status === 0 || project.status === undefined))

async function refreshProjects() {
    try {
        projects.value = await GetProjects()
        error.value = ''
    } catch (cause) {
        error.value = cause instanceof Error ? cause.message : 'Unable to load projects'
    } finally {
        loading.value = false
    }
}

const unsubscribeStatus = EventsOn('project-status', (projectName: string, status: number) => {
    const project = projects.value.find(item => item.name === projectName)
    if (project) {
        project.status = status
    } else {
        refreshProjects()
    }
})

onMounted(refreshProjects)
onBeforeUnmount(unsubscribeStatus)

function statusLabel(status?: number) {
    if (status === 2) return 'Starting'
    if (status === 3) return 'Stopping'
    if (status === 1) return 'Offline'
    if (status === 4) return 'Unknown'
    return 'Running'
}
</script>

<template>
    <div id="home">
        <header class="page-header">
            <div>
                <h1>Server overview</h1>
            </div>
            <div class="status-summary">
                <span class="status-dot" :class="{ inactive: runningProjects.length === 0 }"></span>
                <span><strong>{{ runningProjects.length }}/{{ projects.length }}</strong> servers running</span>
            </div>
        </header>

        <section class="server-list" aria-label="Running servers">
            <p v-if="loading" class="message">Loading projects...</p>
            <p v-else-if="error" class="message error-message">{{ error }}</p>
            <p v-else-if="projects.length === 0" class="message">No projects yet.</p>
            <template v-else>
                <article v-for="project in projects" :key="project.name" class="server-card">
                    <div class="server-heading">
                        <span class="server-icon">{{ project.name.charAt(0).toUpperCase() }}</span>
                        <div>
                            <h2>{{ project.name }}</h2>
                            <p>{{ project.path }}</p>
                        </div>
                    </div>
                    <div class="server-details">
                        <span class="running-status" :class="{ offline: project.status === 1 }">
                            <span class="status-dot"></span>{{ statusLabel(project.status) }}
                        </span>
                        <code @click="CopyToClipboard(`http://localhost:${project.port}/mcp`)">Streamable HTTP: http://localhost:{{ project.port }}/mcp</code>
                    </div>
                </article>
            </template>
        </section>
    </div>
</template>

<style scoped>
#home {
    display: flex;
    flex-direction: column;
    min-height: 100%;
    box-sizing: border-box;
    padding: 20px 16px;
}

.page-header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 24px;
}

h1, h2, p {
    margin-top: 0;
}

h1 {
    margin-bottom: 0;
    color: #f3f6fb;
    font-size: 2rem;
}

.status-summary, .running-status {
    display: inline-flex;
    align-items: center;
    gap: 7px;
}

.status-summary {
    padding: 7px 10px;
    border: 1px solid #315f38;
    border-radius: 5px;
    background-color: #1f3222;
    color: #c9ebce;
    font-size: 0.875rem;
    white-space: nowrap;
}

.status-summary strong {
    color: #fff;
}

.status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: #55bd62;
    box-shadow: 0 0 0 3px rgba(85, 189, 98, 0.14);
}

.status-dot.inactive {
    background-color: #777;
    box-shadow: none;
}

.server-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 12px;
    max-width: 860px;
}

.message {
    margin: 8px 0;
    color: #bdbdbd;
}

.error-message {
    color: #e58585;
}

.server-card {
    min-height: 150px;
    padding: 16px;
    box-sizing: border-box;
    border: 1px solid #414141;
    border-radius: 6px;
    background-color: #242424;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.025);
}

.server-heading {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.server-icon {
    display: grid;
    place-items: center;
    flex: 0 0 auto;
    width: 32px;
    height: 32px;
    border: 1px solid #59677a;
    border-radius: 5px;
    background-color: #2b323c;
    color: #d7e7ff;
    font-size: 0.875rem;
    font-weight: 700;
}

h2 {
    margin-bottom: 5px;
    color: #f3f6fb;
    font-size: 1rem;
    font-weight: 600;
}

.server-heading p {
    margin-bottom: 0;
    color: #bdbdbd;
    font-size: 0.875rem;
}

.server-details {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-top: 26px;
    padding-top: 12px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.running-status {
    color: #a9dcae;
    font-size: 0.8rem;
    font-weight: 600;
}

.running-status .status-dot {
    width: 6px;
    height: 6px;
    box-shadow: none;
}

.running-status.offline {
    color: #e58585;
}

.running-status.offline .status-dot {
    background-color: #e58585;
}

code {
    padding: 4px 7px;
    border: 1px solid #4e4e4e;
    border-radius: 3px;
    background-color: #1b1b1b;
    color: #c8d8ee;
    font-family: inherit;
    font-size: 0.75rem;
    font-weight: 600;
}

@media (max-width: 560px) {
    #home {
        padding: 16px 12px;
    }

    .page-header {
        align-items: flex-start;
        flex-direction: column;
        margin-bottom: 18px;
    }

    h1 {
        font-size: 1.6rem;
    }
}
</style>