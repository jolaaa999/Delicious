<script setup lang="ts">
import { ref } from 'vue'
import { uploadImage, resolveImageUrl } from '@/api/upload'

const props = defineProps<{
  modelValue: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const uploading = ref(false)
const error = ref('')

async function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files?.length) return

  uploading.value = true
  error.value = ''
  const newUrls = [...props.modelValue.filter(Boolean)]

  try {
    for (const file of Array.from(files)) {
      const res = await uploadImage(file)
      newUrls.push(res.url)
    }
    emit('update:modelValue', newUrls)
  } catch (err: unknown) {
    error.value = (err as { message?: string })?.message || '上传失败'
  } finally {
    uploading.value = false
    input.value = ''
  }
}

function remove(idx: number) {
  const next = [...props.modelValue]
  next.splice(idx, 1)
  emit('update:modelValue', next.length ? next : [])
}
</script>

<template>
  <div class="image-uploader">
    <div v-if="modelValue.filter(Boolean).length" class="preview-grid">
      <div v-for="(url, idx) in modelValue" :key="idx">
        <div v-if="url" class="preview-item">
          <img :src="resolveImageUrl(url)" :alt="`图片 ${idx + 1}`" />
          <button type="button" class="preview-remove" aria-label="删除" @click="remove(idx)">×</button>
          <span v-if="idx === 0 && url" class="preview-cover">封面</span>
        </div>
      </div>
    </div>

    <label class="upload-trigger" :class="{ 'upload-trigger--loading': uploading }">
      <input
        type="file"
        accept="image/jpeg,image/png,image/webp,image/gif"
        multiple
        hidden
        :disabled="uploading"
        @change="onFileChange"
      />
      <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M12 16V8M12 8L9 11M12 8L15 11" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
        <path d="M4 16V18C4 19.1 4.9 20 6 20H18C19.1 20 20 19.1 20 18V16" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
      </svg>
      <span>{{ uploading ? '上传中…' : '点击上传图片' }}</span>
    </label>

    <p v-if="error" class="upload-error">{{ error }}</p>
    <p class="upload-hint">支持 JPG / PNG / WebP，首张为封面</p>
  </div>
</template>

<style scoped>
.image-uploader {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.preview-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--color-bg-muted);
}

.preview-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-cover {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 2px;
  font-size: 10px;
  text-align: center;
  color: #fff;
  background: rgba(212, 98, 42, 0.85);
}

.upload-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: border-color var(--transition-fast), color var(--transition-fast);
}

.upload-trigger svg {
  width: 28px;
  height: 28px;
}

.upload-trigger:active:not(.upload-trigger--loading) {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.upload-trigger--loading {
  opacity: 0.6;
  cursor: wait;
}

.upload-error {
  font-size: 13px;
  color: var(--color-danger);
}

.upload-hint {
  font-size: 12px;
  color: var(--color-text-muted);
  margin: 0;
}
</style>
