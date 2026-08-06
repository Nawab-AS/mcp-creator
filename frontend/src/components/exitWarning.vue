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

.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease-out;
}

.modal {
    background-color: #2d2d2d;
    border: 1px solid #444;
    border-radius: 8px;
    padding: 0;
    min-width: 380px;
    max-width: 450px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
    animation: slideUp 0.2s ease-out;
}

.modal-header {
    padding: 16px 20px;
    border-bottom: 1px solid #444;
}

.modal-header h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #eee;
}

.modal-body {
    padding: 20px;
}

.modal-body p {
    margin: 0;
    color: #ccc;
    line-height: 1.5;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 20px;
    border-top: 1px solid #444;
}

.btn {
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    border: none;
    transition: all 0.15s ease;
}

.btn-cancel {
    background-color: #3a3a3a;
    color: #eee;
    border: 1px solid #555;
}

.btn-cancel:hover {
    background-color: #444;
    border-color: #666;
}

.btn-danger {
    background-color: #c0392b;
    color: white;
}

.btn-danger:hover {
    background-color: #e74c3c;
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