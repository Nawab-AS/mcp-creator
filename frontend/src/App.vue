<script lang="ts" setup>
import { ref } from "vue";
import Initials from "./components/Initials.vue";
// TODO: implement better UI

const pages = ["home", "projects", "settings"]

const name = ref("Hamid Syed")
const page = ref("home")
</script>

<template>
  <div id="app">
    <div id="sidebar">
      <div id="greeting">
        <Initials :Initials="name.split(' ').map(s => s[0]).join('')" />
        <img src="./assets/images/settings.svg" alt="settings button">
      </div>
      <hr>
      <ul id="menu">
        <li v-for="item in pages" :key="item" @click="page = item" :class="{ active: page === item }">
          <p>{{ item.charAt(0).toUpperCase() + item.slice(1) }}</p>
        </li>
      </ul>
    </div>
    <div id="content">
      <h1>{{ page.charAt(0).toUpperCase() + page.slice(1) }}</h1>
    </div>
  </div>
</template>

<style scoped>
#app {
  display: flex;
  min-height: 100vh;
}

#sidebar {
  height: 100vh;
  width: 200px;
  background-color: #242424;
  color: white;
}

#greeting {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px 10px;
}

#greeting>img {
  width: 30px;
  height: 30px;
  transition: all 0.2s ease-in-out;
  transition-delay: 0.1s;
  scale: 1.25;
}

#greeting>img:hover {
  cursor: pointer;
  filter: brightness(1.1);
  scale: 1.35;
}

hr {
  border: 0;
  height: 2px;
  background-color: #444;
  margin: 0;
}

#menu {
  list-style: none;
  padding: 0;
  height: calc(100vh - 100px);
  gap: 10px;
}

#menu>li {
  margin: 10px 10px;
  transition: all 0.3s ease-in-out;
  text-align: center;
  border-radius: 10px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

#menu>li:hover {
  cursor: pointer;
  background-color: #333;
}

#menu>li.active {
  background-color: #3e3e3e;
}

#content {
  flex: 1;
  padding: 10px;
}

</style>
