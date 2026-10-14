<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from "vue";
import { EventsOn, EventsEmit, EventsOff } from "../../wailsjs/runtime/runtime.js";

const showModal = ref(false);

const handleSoftExit = () => {
    showModal.value = true;
};

const handleStopServers = () => {
    showModal.value = false;
    EventsEmit("exit-ack");
};

const handleCancel = () => {
    showModal.value = false;
};

onMounted(() => {
    EventsOn("Soft-exit", handleSoftExit);
});

onUnmounted(() => {
    EventsOff("Soft-exit");
});
</script>

<template>
    <div class="exit-warning-host">
        <Teleport to="#modals">
            <div v-if="showModal" class="modal-overlay" @click="handleCancel">
                <div class="modal" @click.stop>
                    <div class="modal-header">
                        <h3>Exit Application</h3>
                    </div>
                    <div class="modal-body">
                        <p>Exiting will stop all running MCP servers. Are you sure you want to continue?</p>
                    </div>
                    <div class="modal-footer">
                        <button class="btn btn-cancel" @click="handleCancel">Cancel</button>
                        <button class="btn btn-danger" @click="handleStopServers">Stop Servers</button>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<style scoped>
.exit-warning-host {
    height: 0px;
    width: 0px;
}

.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    padding: 20px;
    background-color: rgba(0, 0, 0, 0.58);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease-out;
}

.modal {
    width: min(100%, 420px);
    background-color: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-surface);
    padding: 0;
    box-shadow: 0 12px 36px rgba(0, 0, 0, 0.45);
    animation: slideUp 0.2s ease-out;
}

.modal-header {
    padding: 16px 18px;
    border-bottom: 1px solid var(--border);
}

.modal-header h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
}

.modal-body {
    padding: 18px;
}

.modal-body p {
    margin: 0;
    color: var(--text-secondary);
    line-height: 1.5;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 14px 18px;
    border-top: 1px solid var(--border);
}

.btn {
    padding: 6px 10px;
    border-radius: var(--radius-control);
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    border: 1px solid transparent;
    transition: background-color 0.2s ease, border-color 0.2s ease;
}

.btn-cancel {
    background-color: #3c3c3c;
    color: var(--text-primary);
    border-color: var(--border-strong);
}

.btn-cancel:hover {
    background-color: #484848;
    border-color: #666;
}

.btn-danger {
    background-color: var(--danger);
    border-color: var(--danger-border);
    color: #fff;
}

.btn-danger:hover {
    background-color: var(--danger-hover);
    border-color: #d14343;
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

@keyframes slideUp {
    from { 
        opacity: 0;
        transform: translateY(20px);
    }
    to { 
        opacity: 1;
        transform: translateY(0);
    }
}
</style>