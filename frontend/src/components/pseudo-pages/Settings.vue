<script setup lang="ts">
import { OpenStorageDir, Uninstall } from '../../../wailsjs/go/main/App'

</script>


<template>
    <div id="settings">
        <h1 id="header">Settings</h1>

        <div id="buttons">
            <button class="destructive" @click="Uninstall()">Uninstall</button>
            <button class="secondary" @click="OpenStorageDir()">Open Storage Directory</button>
        </div>
    </div>
</template>

<style scoped>
#settings {
    min-height: 100%;
    padding: 20px 16px;
}

h1 {
    margin: 0;
    margin-bottom: 24px;
    color: #f3f6fb;
    font-size: 2rem;
}

#buttons {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 30px;
}

#buttons button {
    width: fit-content;
    padding: 7px 11px;
    font: inherit;
    font-size: 0.85rem;
    font-weight: 600;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-control);
    cursor: pointer;
}

#buttons button.destructive {
    background-color: var(--danger);
    border-color: var(--danger-border);
    color: #fff;
}

#buttons button.secondary {
    background-color: #3c3c3c;
    color: var(--text-primary);
}
</style>