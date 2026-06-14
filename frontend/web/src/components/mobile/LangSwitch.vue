<script setup lang="ts">
const props = defineProps<{
  modelValue: 'en' | 'zh'
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: 'en' | 'zh']
}>()

function setLang(lang: 'en' | 'zh') {
  if (props.disabled || props.modelValue === lang) return
  emit('update:modelValue', lang)
}
</script>

<template>
  <div class="lang-switch" role="group" aria-label="切换语言">
    <button
      type="button"
      class="lang-switch__btn"
      :class="{ 'lang-switch__btn--active': modelValue === 'zh' }"
      :disabled="disabled"
      @click="setLang('zh')"
    >
      中文
    </button>
    <button
      type="button"
      class="lang-switch__btn"
      :class="{ 'lang-switch__btn--active': modelValue === 'en' }"
      :disabled="disabled"
      @click="setLang('en')"
    >
      EN
    </button>
  </div>
</template>
