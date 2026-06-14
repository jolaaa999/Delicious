<script setup lang="ts">
const props = defineProps<{ modelValue: number }>()
const emit = defineEmits<{ 'update:modelValue': [value: number] }>()

function setRating(n: number) {
  emit('update:modelValue', n)
}
</script>

<template>
  <div class="star-rating" role="group" aria-label="评分">
    <button
      v-for="n in 5"
      :key="n"
      type="button"
      class="star-rating__star"
      :class="{ 'star-rating__star--active': n <= props.modelValue }"
      :aria-label="`${n} 星`"
      @click="setRating(n)"
    >
      ★
    </button>
  </div>
</template>
