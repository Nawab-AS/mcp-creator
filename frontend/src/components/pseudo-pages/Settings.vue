<script setup lang="ts">
import { OpenStorageDir, Uninstall } from '../../../wailsjs/go/main/App'

</script>


<template>
    <div id="settings">
        <h1 id="header">Settings</h1>

        <div id="buttons">
            <button @click="Uninstall()">Uninstall</button>
            <button @click="OpenStorageDir()">Open Storage Directory</button>
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
    padding: 10px 20px;
    font-size: 1rem;
    border: none;
    border-radius: 5px;
    background-color: #6d1717;
    color: #ffffff;
    cursor: pointer;
}
</style>