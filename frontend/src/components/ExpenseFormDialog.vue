<!-- ExpenseFormDialog：消费记录创建/编辑弹窗（ExpenseList 与 ExpenseDetail 共用） -->
<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑消费记录' : '添加消费记录'" width="640px" append-to-body destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" placeholder="例如：周末火锅" maxlength="128" />
      </el-form-item>
      <el-form-item label="金额" prop="amount">
        <el-input-number v-model="form.amount" :min="0.01" :precision="2" :step="10" style="width: 200px" />
      </el-form-item>
      <el-form-item label="类别" prop="category">
        <el-select v-model="form.category" style="width: 200px">
          <el-option v-for="opt in CategoryOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="付款人" prop="payer_id">
        <el-select v-model="form.payer_id" style="width: 200px">
          <el-option v-for="m in members" :key="m.user_id" :label="m.nickname + ' (' + m.username + ')'" :value="m.user_id" />
        </el-select>
      </el-form-item>
      <el-form-item label="消费时间" prop="paid_at">
        <el-date-picker v-model="form.paid_at" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" placeholder="选择时间" style="width: 220px" />
      </el-form-item>
      <el-form-item label="分摊方式" prop="split_type">
        <el-radio-group v-model="form.split_type">
          <el-radio v-for="opt in SplitTypeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="参与人" prop="shares">
        <div class="share-list">
          <div v-for="(share, index) in form.shares" :key="share.user_id" class="share-row">
            <span class="share-row__name">{{ share.nickname }}</span>
            <el-input-number
              v-if="form.split_type === 'ratio'"
              v-model="share.ratio"
              :min="0"
              :precision="2"
              :step="1"
              size="small"
              style="width: 140px"
            />
            <el-input-number
              v-else-if="form.split_type === 'amount'"
              v-model="share.amount"
              :min="0"
              :precision="2"
              :step="10"
              size="small"
              style="width: 140px"
            />
            <el-button v-if="share.user_id !== form.payer_id" size="small" text type="danger" @click="removeShare(index)">
              移除
            </el-button>
          </div>
        </div>
      </el-form-item>
      <el-form-item label="小票图片">
        <el-input v-model="form.receipt_url" placeholder="小票图片 URL（可选）" maxlength="512" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { CategoryOptions, SplitTypeOptions, SplitType } from '@/constants'
import type { ExpenseInfo, ExpensePayload } from '@/api/expense'
import type { MemberInfo } from '@/api/group'

interface ShareForm {
  user_id: number
  nickname: string
  ratio: number
  amount: number
}

const props = defineProps<{ members: MemberInfo[]; expense?: ExpenseInfo | null }>()
const emit = defineEmits<{ (e: 'saved', payload: ExpensePayload): void }>()

const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const isEdit = ref(false)

const form = reactive<{
  title: string
  amount: number
  category: string
  payer_id: number
  split_type: string
  paid_at: string
  receipt_url: string
  shares: ShareForm[]
}>({
  title: '',
  amount: 100,
  category: 'dining',
  payer_id: 0,
  split_type: SplitType.EQUAL,
  paid_at: '',
  receipt_url: '',
  shares: [],
})

const rules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
  category: [{ required: true, message: '请选择类别', trigger: 'change' }],
  payer_id: [{ required: true, message: '请选择付款人', trigger: 'change' }],
  paid_at: [{ required: true, message: '请选择消费时间', trigger: 'change' }],
  split_type: [{ required: true, message: '请选择分摊方式', trigger: 'change' }],
  shares: [{ required: true, type: 'array', min: 1, message: '至少选择一位参与人', trigger: 'change' }],
}

function buildShares() {
  form.shares = props.members
    .filter((m) => m.user_id !== form.payer_id)
    .map((m) => ({ user_id: m.user_id, nickname: m.nickname || m.username, ratio: 1, amount: 0 }))
  // 付款人默认也参与
  const payer = props.members.find((m) => m.user_id === form.payer_id)
  if (payer && !form.shares.some((s) => s.user_id === payer.user_id)) {
    form.shares.unshift({ user_id: payer.user_id, nickname: payer.nickname || payer.username, ratio: 1, amount: 0 })
  }
}

function open(expense?: ExpenseInfo | null) {
  isEdit.value = !!expense
  const firstMember = props.members[0]
  form.title = expense?.title || ''
  form.amount = expense?.amount || 100
  form.category = expense?.category || 'dining'
  form.payer_id = expense?.payer_id || firstMember?.user_id || 0
  form.split_type = expense?.split_type || SplitType.EQUAL
  form.paid_at = expense?.paid_at || ''
  form.receipt_url = expense?.receipt_url || ''
  form.shares = []
  if (expense && expense.shares.length > 0) {
    form.shares = expense.shares.map((s) => ({
      user_id: s.user_id,
      nickname: s.nickname || s.username,
      ratio: s.ratio || 1,
      amount: s.share_amount || 0,
    }))
  } else {
    buildShares()
  }
  visible.value = true
}

function removeShare(index: number) {
  form.shares.splice(index, 1)
}

async function handleSubmit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  if (form.shares.length === 0) {
    ElMessage.warning('请至少选择一位参与人')
    return
  }
  const payload: ExpensePayload = {
    title: form.title,
    amount: form.amount,
    category: form.category,
    payer_id: form.payer_id,
    split_type: form.split_type,
    paid_at: form.paid_at,
    receipt_url: form.receipt_url,
    shares: form.shares.map((s) => ({
      user_id: s.user_id,
      ratio: form.split_type === SplitType.RATIO ? s.ratio : undefined,
      amount: form.split_type === SplitType.AMOUNT ? s.amount : undefined,
    })),
  }
  submitting.value = true
  try {
    emit('saved', payload)
    visible.value = false
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.share-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.share-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.share-row__name {
  width: 120px;
}
</style>
