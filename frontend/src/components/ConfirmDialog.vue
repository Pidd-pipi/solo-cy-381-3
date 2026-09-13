<!-- ConfirmDialog：通用二次确认弹窗 -->
<template>
  <el-dialog v-model="visible" :title="title" width="420px" append-to-body>
    <div class="confirm-dialog__body">
      <el-icon class="confirm-dialog__icon" color="#e6a23c" :size="28"><WarningFilled /></el-icon>
      <span>{{ message }}</span>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="danger" :loading="loading" @click="handleConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    title?: string
    message?: string
    loading?: boolean
  }>(),
  { title: '确认操作', message: '确定执行该操作吗？', loading: false },
)

const emit = defineEmits<{ (e: 'confirm'): void }>()
const visible = ref(false)

function open() {
  visible.value = true
}

function handleConfirm() {
  emit('confirm')
  visible.value = false
}

defineExpose({ open })
</script>

<style scoped>
.confirm-dialog__body {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
