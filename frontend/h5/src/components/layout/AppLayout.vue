<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppTabBar from './AppTabBar.vue'

const route = useRoute()
const showTabBar = computed(() => !route.meta.hideTabBar)
</script>

<template>
  <div class="app-layout" :class="{ 'app-layout--with-tabbar': showTabBar }">
    <main class="app-layout__main">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
    <AppTabBar v-if="showTabBar" />
  </div>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
  min-height: 100dvh;
}

.app-layout--with-tabbar .app-layout__main {
  padding-bottom: calc(var(--tabbar-height) + var(--safe-bottom));
}

.app-layout__main {
  min-height: inherit;
}
</style>
