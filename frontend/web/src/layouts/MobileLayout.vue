<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import MobileTabBar from '@/components/layout/MobileTabBar.vue'

const route = useRoute()
const showTabBar = computed(() => !route.meta.hideTabBar)
</script>

<template>
  <div class="mobile-layout" :class="{ 'mobile-layout--with-tabbar': showTabBar }">
    <main class="mobile-layout__main">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
    <MobileTabBar v-if="showTabBar" />
  </div>
</template>

<style scoped>
.mobile-layout {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--color-bg-page);
}

.mobile-layout--with-tabbar .mobile-layout__main {
  padding-bottom: calc(var(--tabbar-height) + var(--safe-bottom));
}

.mobile-layout__main {
  min-height: inherit;
}
</style>
