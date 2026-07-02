<template>
  <div>
    <div class="exam-qm-header">
      <div class="exam-qm-info">
        <strong>📋 {{ examTitle }}</strong>
        <span>考题管理</span>
      </div>
      <div class="exam-qm-stats">
        <span class="stat-badge qty">📝 {{ questions.length }}题</span>
        <span class="stat-badge score">💯 {{ totalScore }}分</span>
      </div>
      <div style="display:flex;gap:8px;flex-wrap:wrap;">
        <button class="btn btn-primary" @click="openCreate">➕ 新增考题</button>
        <button class="btn btn-secondary" @click="triggerTableImport">📄 表格导入</button>
        <button class="btn btn-outline-export" @click="exportQuestions">📤 导出考题</button>
        <button class="btn btn-outline-danger" @click="clearQuestions">🗑️ 清空考题</button>
        <button class="btn btn-secondary" @click="refresh">🔄 刷新</button>
      </div>
    </div>

    <div v-if="questions.length === 0" class="empty-state">
      <div class="empty-icon">📝</div>
      <h3>暂无考题</h3>
      <p>请新增考题或表格导入</p>
    </div>

    <div v-else class="table-wrap">
      <table class="exam-qm-table">
        <thead>
          <tr>
            <th>序号</th>
            <th>题目内容</th>
            <th>类型</th>
            <th style="width:100px;">分值</th>
            <th style="width:120px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(eq, i) in questions" :key="eq.questionId || i">
            <td>{{ i + 1 }}</td>
            <td class="q-title-cell">{{ eq.title || '--' }}</td>
            <td>
              <span :class="['q-type', eq.type || 'single']">{{ typeLabel(eq.type) }}</span>
            </td>
            <td>
              <input
                type="number"
                class="score-input"
                :value="eq.score"
                min="0"
                step="1"
                @change="updateScore(eq, $event)"
              />
            </td>
            <td>
              <button class="btn-sm edit" @click="openEdit(eq)">编辑</button>
              <button class="btn-sm remove" @click="removeQuestion(eq)">移除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑考题弹窗 -->
    <Modal :show="modalShow" :title="editing ? '编辑考题' : '新增考题'" :maxWidth="'640px'" @close="closeModal">
      <div class="form-section">
        <h4 class="form-section-title">基本信息</h4>
        <div class="form-group">
          <label>所属分类</label>
          <select v-model="form.categoryId">
            <option value="">请选择分类</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>题目类型</label>
          <select v-model="form.type" @change="onTypeChange">
            <option value="single">单选题</option>
            <option value="multiple">多选题</option>
            <option value="judge">判断题</option>
            <option value="fill">填空题</option>
            <option value="essay">简答题</option>
          </select>
        </div>
        <div class="form-group">
          <label>题目内容</label>
          <textarea v-model="form.title" placeholder="请输入题目内容" rows="3"></textarea>
        </div>
      </div>

      <!-- 选项编辑 -->
      <div class="form-section" v-if="!isTextType">
        <h4 class="form-section-title">选项设置</h4>
        <div class="option-editor">
          <div v-for="(opt, i) in form.options" :key="i" class="option-row">
            <div class="opt-label">{{ opt.label }}</div>
            <input type="text" v-model="opt.text" :placeholder="`选项${opt.label}内容`" />
            <button class="btn btn-sm btn-outline" @click="removeOption(i)" :disabled="form.options.length <= 2">删除</button>
          </div>
        </div>
        <button class="add-option-btn" @click="addOption">＋ 添加选项</button>
      </div>

      <!-- 答案选择 -->
      <div class="form-section" v-if="!isTextType">
        <h4 class="form-section-title">正确答案</h4>
        <div :class="form.type === 'multiple' ? 'answer-checkboxes' : 'answer-radios'">
          <label
            v-for="opt in form.options" :key="opt.label"
            :class="[form.type === 'multiple' ? 'answer-checkbox' : 'answer-radio', { selected: isAnswerSelected(opt.label) }]"
            @click="toggleAnswer(opt.label)"
          >
            <input :type="form.type === 'multiple' ? 'checkbox' : 'radio'" :value="opt.label" />
            {{ opt.label }}
          </label>
        </div>
      </div>

      <!-- 文本类答案 -->
      <div class="form-section" v-if="isTextType">
        <h4 class="form-section-title">正确答案</h4>
        <div class="form-group">
          <input type="text" v-model="form.answerText" placeholder="请输入正确答案（多个答案用逗号分隔）" />
        </div>
      </div>

      <div class="form-section">
        <h4 class="form-section-title">答案解析（可选）</h4>
        <div class="form-group">
          <textarea v-model="form.explanation" placeholder="输入答案解析" rows="2"></textarea>
        </div>
      </div>

      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="closeModal">取消</button>
        <button class="modal-btn confirm" @click="save">{{ editing ? '保存修改' : '添加考题' }}</button>
      </div>
    </Modal>

    <!-- 表格导入弹窗 -->
    <Modal :show="importShow" title="📄 表格导入考题" @close="importShow = false">
      <div class="import-area" @click="importFile?.click()">
        <div class="import-icon">📄</div>
        <p>点击选择 Excel/CSV/JSON 文件</p>
        <button class="btn btn-secondary">选择文件</button>
        <input type="file" ref="importFile" accept=".xlsx,.xls,.csv,.json" @change="handleFile" style="display:none;" />
      </div>
      <div v-if="importFileInfo" class="import-preview">
        <div class="file-info"><strong>{{ importFileInfo.name }}</strong></div>
      </div>
      <div class="import-hint">
        💡 支持 .xlsx / .csv / .json 格式，表头：题型、题目内容、A、B、C、D、正确答案、分值、答案解析
        <a class="template-link" @click="downloadTemplate">📥 下载导入模板</a>
      </div>
      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="importShow = false">取消</button>
        <button class="modal-btn confirm" :disabled="!importFileInfo || importing" @click="doTableImport">{{ importing ? '导入中...' : '开始导入' }}</button>
      </div>
    </Modal>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, showToast } from './shared'
import Modal from '@/components/Modal.vue'

const props = defineProps({
  examId: { type: Number, required: true },
  examTitle: { type: String, default: '' },
  categories: { type: Array, default: () => [] }
})

const questions = ref([])

// 实时计算总分（所有考题分值之和）
const totalScore = computed(() => {
  return questions.value.reduce((sum, q) => sum + (Number(q.score) || 0), 0)
})

const importShow = ref(false)
const importFile = ref(null)
const importFileInfo = ref(null)
const importing = ref(false)

// ── 新增/编辑考题 ──
const modalShow = ref(false)
const editing = ref(null)
const saving = ref(false)

const defaultOptions = [
  { label: 'A', text: '' }, { label: 'B', text: '' }, { label: 'C', text: '' }, { label: 'D', text: '' }
]

function getEmptyForm() {
  return { type: 'single', title: '', options: JSON.parse(JSON.stringify(defaultOptions)), answer: [], answerText: '', explanation: '', categoryId: '' }
}

const form = ref(getEmptyForm())
const isTextType = computed(() => ['fill', 'essay'].includes(form.value.type))

function typeLabel(t) {
  return { single: '单选题', multiple: '多选题', judge: '判断题', fill: '填空题', essay: '简答题' }[t] || t
}

// ── 考题表单操作 ──
function onTypeChange() {
  if (form.value.type === 'judge') {
    form.value.options = [{ label: 'A', text: '正确' }, { label: 'B', text: '错误' }]
  } else if (['single', 'multiple'].includes(form.value.type)) {
    form.value.options = JSON.parse(JSON.stringify(defaultOptions))
  }
  form.value.answer = []
  form.value.answerText = ''
}

function addOption() {
  const labels = form.value.options.map(o => o.label)
  let next = 'A'
  while (labels.includes(next)) {
    next = String.fromCharCode(next.charCodeAt(0) + 1)
  }
  form.value.options.push({ label: next, text: '' })
}

function removeOption(i) {
  if (form.value.options.length <= 2) return
  const removedLabel = form.value.options[i].label
  form.value.options.splice(i, 1)
  form.value.answer = form.value.answer.filter(a => a !== removedLabel)
}

function isAnswerSelected(label) { return (form.value.answer || []).includes(label) }

function toggleAnswer(label) {
  if (form.value.type === 'multiple') {
    const idx = form.value.answer.indexOf(label)
    if (idx > -1) form.value.answer.splice(idx, 1)
    else form.value.answer.push(label)
  } else {
    form.value.answer = [label]
  }
}

function openCreate() {
  editing.value = null
  form.value = getEmptyForm()
  modalShow.value = true
}

function openEdit(eq) {
  editing.value = eq
  const isText = ['fill', 'essay'].includes(eq.type)

  // 后端返回的 options 是 []OptionPair（{label, type, content, url}），已由 JSON 反序列化为对象
  let rawOpts = []
  if (typeof eq.options === 'string') {
    try { rawOpts = JSON.parse(eq.options || '[]') } catch (e) { rawOpts = [] }
  } else {
    rawOpts = eq.options || []
  }

  // answer 同理，后端返回 []string，已反序列化
  let answer = []
  if (typeof eq.answer === 'string') {
    try { answer = JSON.parse(eq.answer || '[]') } catch (e) { answer = [] }
  } else {
    answer = eq.answer || []
  }

  // 标准化为表单使用的 {label, text}，从 OptionPair.content 取值
  const normalizedOpts = rawOpts.length > 0
    ? rawOpts.map(opt => ({
        label: String(opt.label || '').toUpperCase(),
        text: opt.content || opt.text || ''
      }))
    : JSON.parse(JSON.stringify(defaultOptions))

  // 统一大写答案标签，与表单选项 label 匹配
  const answerArr = Array.isArray(answer)
    ? answer.map(a => String(a).toUpperCase()).filter(Boolean)
    : (answer ? [String(answer).toUpperCase()] : [])

  form.value = {
    type: eq.type || 'single',
    title: eq.title || '',
    options: normalizedOpts,
    answer: isText ? [] : answerArr,
    answerText: isText ? answerArr.join(',') : '',
    explanation: eq.explanation || '',
    categoryId: eq.categoryId || ''
  }
  modalShow.value = true
}

function closeModal() { modalShow.value = false }

async function save() {
  if (!form.value.title.trim()) return showToast('请输入题目内容', 'error')

  const isText = ['fill', 'essay'].includes(form.value.type)
  if (!isText && form.value.answer.length === 0) return showToast('请选择正确答案', 'error')
  if (isText && !form.value.answerText.trim()) return showToast('请输入正确答案', 'error')

  if (!editing.value && !form.value.categoryId) return showToast('请选择所属分类', 'error')

  // 将表单的 {label, text} 转换为后端需要的 {label, type: "text", content}
  const richOptions = isText ? [] : form.value.options.map(opt => ({
    label: opt.label,
    type: 'text',
    content: opt.text || ''
  }))

  const data = {
    title: form.value.title,
    type: form.value.type,
    categoryId: Number(form.value.categoryId || editing.value?.categoryId || 0),
    answer: isText ? (form.value.answerText || '').split(',').filter(Boolean) : (form.value.answer || []),
    explanation: form.value.explanation,
    options: richOptions
  }

  saving.value = true
  try {
    if (editing.value) {
      const res = await api(`/admin/questions/${editing.value.questionId}`, { method: 'PUT', body: JSON.stringify(data) })
      if (res.code === 200) {
        showToast('更新成功', 'success')
        closeModal()
        refresh()
      } else {
        showToast(res.message || '更新失败', 'error')
      }
    } else {
      // 先创建考题
      const createRes = await api('/admin/questions', { method: 'POST', body: JSON.stringify(data) })
      if (createRes.code === 200) {
        // 再关联到考试
        const newId = createRes.data?.id
        const linkRes = await api(`/admin/exams/${props.examId}/questions`, {
          method: 'POST',
          body: JSON.stringify({ questionIds: [newId] })
        })
        if (linkRes.code === 200) {
          showToast('添加成功', 'success')
          closeModal()
          refresh()
        } else {
          showToast(linkRes.message || '关联考试失败', 'error')
        }
      } else {
        showToast(createRes.message || '创建失败', 'error')
      }
    }
  } catch (e) {
    showToast('操作失败：' + e.message, 'error')
  } finally {
    saving.value = false
  }
}

// ── 考试内考题列表 ──
async function refresh() {
  const res = await api(`/admin/exams/${props.examId}/questions`)
  questions.value = res.data?.questions || []
}

async function updateScore(eq, event) {
  const score = Number(event.target.value) || 0
  const res = await api(`/admin/exams/${props.examId}/questions/${eq.questionId}`, {
    method: 'PUT',
    body: JSON.stringify({ score })
  })
  if (res.code === 200) {
    showToast('分值更新成功', 'success')
    refresh()
  } else {
    showToast(res.message || '更新失败', 'error')
    event.target.value = eq.score
  }
}

async function removeQuestion(eq) {
  if (!confirm('确定移除此考题？')) return
  const res = await api(`/admin/exams/${props.examId}/questions/${eq.questionId}`, { method: 'DELETE' })
  if (res.code === 200) {
    showToast('移除成功', 'success')
    refresh()
  } else {
    showToast(res.message || '移除失败', 'error')
  }
}

// ── 一键清空 ──
async function clearQuestions() {
  if (questions.value.length === 0) return showToast('当前考试暂无考题', 'info')
  if (!confirm(`确定要清空当前考试下全部 ${questions.value.length} 道考题吗？此操作不可撤销。`)) return
  const res = await api(`/admin/exams/${props.examId}/questions/clear`, { method: 'DELETE' })
  if (res.code === 200) {
    showToast('已清空所有考题', 'success')
    refresh()
  } else {
    showToast(res.message || '清空失败', 'error')
  }
}

// ── 导出考题 ──
async function exportQuestions() {
  try {
    const token = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)?.[1] || localStorage.getItem('exam_token')
    const res = await fetch(`/api/admin/exams/${props.examId}/questions/export`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!res.ok) throw new Error('导出失败')
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `考题导出_${props.examTitle || props.examId}.xlsx`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    showToast('导出成功', 'success')
  } catch (e) {
    showToast('导出失败：' + e.message, 'error')
  }
}

// ── 下载导入模板 ──
async function downloadTemplate() {
  try {
    const token = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)?.[1] || localStorage.getItem('exam_token')
    const res = await fetch('/api/admin/template', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!res.ok) throw new Error('下载失败')
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = '考题导入模板.xlsx'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (e) {
    showToast('模板下载失败：' + e.message, 'error')
  }
}

// ── 表格导入 ──
function triggerTableImport() { importFileInfo.value = null; importShow.value = true }

function handleFile(e) {
  const file = e.target.files[0]
  if (!file) return
  importFileInfo.value = file
}

async function doTableImport() {
  if (!importFileInfo.value || importing.value) return
  importing.value = true
  try {
    const formData = new FormData()
    formData.append('file', importFileInfo.value)
    const token = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)?.[1] || localStorage.getItem('exam_token')
    const res = await fetch(`/api/admin/exams/${props.examId}/questions/import`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    })
    const data = await res.json()
    if (data.code === 200) {
      showToast(data.message || '导入成功', 'success')
      importShow.value = false
      refresh()
    } else {
      showToast(data.message || '导入失败', 'error')
    }
  } catch (e) {
    showToast('导入失败：' + e.message, 'error')
  } finally {
    importing.value = false
  }
}

onMounted(refresh)
</script>

<style scoped>
.exam-qm-header {
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 12px; margin-bottom: 18px;
  padding: 16px 20px; background: #f8fafc; border-radius: var(--radius);
  border: 1px solid var(--border);
}
.exam-qm-info { display: flex; gap: 20px; flex-wrap: wrap; font-size: 13px; color: var(--text-secondary); align-items: center; }
.exam-qm-info strong { color: var(--text); font-size: 15px; }
.exam-qm-stats { display: flex; gap: 10px; align-items: center; }
.stat-badge {
  padding: 5px 14px; border-radius: 20px; font-size: 13px; font-weight: 700; white-space: nowrap;
}
.stat-badge.qty { background: #eef2ff; color: #6366f1; }
.stat-badge.score { background: linear-gradient(135deg, #fef2f2, #fff7ed); color: #ea580c; }

.table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid var(--border); }
.exam-qm-table { width: 100%; border-collapse: collapse; }
.exam-qm-table th { background: #f8fafc; padding: 10px 12px; text-align: left; font-size: 13px; font-weight: 600; color: var(--text-secondary); border-bottom: 2px solid var(--border); }
.exam-qm-table td { padding: 10px 12px; border-bottom: 1px solid #f1f5f9; font-size: 13px; }
.exam-qm-table tbody tr:hover { background: #f8fafc; }
.exam-qm-table .score-input {
  width: 72px; padding: 6px 8px; border: 2px solid var(--border);
  border-radius: var(--radius-sm); text-align: center; font-size: 13px;
  transition: border-color var(--transition);
}
.exam-qm-table .score-input:focus { outline: none; border-color: var(--primary); }
.q-title-cell { max-width: 260px; overflow: hidden; text-overflow: ellipsis; }

.q-type {
  padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 700;
}
.q-type.single { background: #eff6ff; color: #3b82f6; }
.q-type.multiple { background: #fff7ed; color: #ea580c; }
.q-type.judge { background: #fef2f2; color: #dc2626; }
.q-type.fill { background: #f0fdf4; color: #16a34a; }
.q-type.essay { background: #faf5ff; color: #9333ea; }

.btn-sm { padding: 5px 12px; font-size: 12px; border: none; border-radius: var(--radius-sm); cursor: pointer; font-weight: 600; margin-right: 4px; }
.btn-sm.edit { background: var(--primary-light, #eef2ff); color: var(--primary, #6366f1); }
.btn-sm.edit:hover { background: #dde4ff; }
.btn-sm.remove { background: var(--danger); color: white; }
.btn-sm.remove:hover { background: #dc2626; }

/* 导入 */
.import-area {
  border: 2px dashed var(--border); border-radius: var(--radius);
  padding: 36px; text-align: center; cursor: pointer; margin-bottom: 18px;
  transition: all var(--transition);
}
.import-area:hover { border-color: var(--primary); background: #f8faff; }
.import-icon { font-size: 44px; margin-bottom: 10px; }
.import-preview { margin-top: 15px; padding: 14px 18px; background: #f0fdf4; border-radius: var(--radius); border: 1px solid #bbf7d0; }
.import-hint { margin-top: 10px; color: #94a3b8; font-size: 12px; }
.import-hint .template-link {
  display: inline-block; margin-left: 12px; color: var(--primary);
  cursor: pointer; font-weight: 600; text-decoration: underline;
}
.import-hint .template-link:hover { color: var(--primary-dark); }

/* 表单 */
.form-section { margin-bottom: 20px; }
.form-section-title { font-size: 14px; font-weight: 600; color: var(--text); margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid #f1f5f9; }
.form-group { margin-bottom: 14px; }
.form-group label { display: block; font-size: 13px; font-weight: 600; color: var(--text-secondary); margin-bottom: 6px; }
.form-group input,
.form-group textarea,
.form-group select {
  width: 100%; padding: 10px 12px; border: 2px solid var(--border);
  border-radius: var(--radius-sm); font-size: 14px; font-family: inherit;
  transition: border-color var(--transition); box-sizing: border-box;
}
.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus { outline: none; border-color: var(--primary); }
.form-group textarea { resize: vertical; }

.option-editor { display: flex; flex-direction: column; gap: 10px; }
.option-row { display: flex; align-items: center; gap: 10px; }
.opt-label {
  width: 32px; height: 32px; border-radius: 8px;
  background: var(--primary); color: white;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 13px; flex-shrink: 0;
}
.option-row input { flex: 1; padding: 8px 12px; border: 2px solid var(--border); border-radius: var(--radius-sm); font-size: 14px; }
.option-row input:focus { outline: none; border-color: var(--primary); }

.add-option-btn {
  margin-top: 10px; background: none; border: 1px dashed #cbd5e1;
  color: var(--primary); font-size: 13px; font-weight: 600; padding: 8px 16px;
  border-radius: var(--radius-sm); cursor: pointer; width: 100%;
  transition: all var(--transition);
}
.add-option-btn:hover { border-color: var(--primary); background: #f8faff; }

.answer-radios, .answer-checkboxes { display: flex; flex-wrap: wrap; gap: 10px; }
.answer-radio, .answer-checkbox {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border: 2px solid var(--border); border-radius: var(--radius-sm);
  cursor: pointer; font-weight: 600; font-size: 13px; transition: all var(--transition);
}
.answer-radio.selected, .answer-checkbox.selected { border-color: var(--primary); background: #eef2ff; }
.answer-radio input, .answer-checkbox input { display: none; }

.modal-buttons { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }
.modal-btn {
  padding: 10px 24px; border: none; border-radius: var(--radius-sm);
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all var(--transition);
}
.modal-btn.cancel { background: #f1f5f9; color: var(--text-secondary); }
.modal-btn.cancel:hover { background: #e2e8f0; }
.modal-btn.confirm { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; }
.modal-btn.confirm:hover { box-shadow: 0 4px 16px rgba(99,102,241,0.35); }
.modal-btn:disabled { opacity: 0.5; cursor: not-allowed; }



.empty-state { text-align: center; padding: 60px 20px; color: var(--text-secondary); }
.empty-icon { font-size: 48px; margin-bottom: 14px; }
.empty-state h3 { color: var(--text); margin-bottom: 6px; }
.empty-state p { font-size: 13px; }

.btn {
  padding: 8px 16px; border: none; border-radius: var(--radius-sm);
  font-size: 13px; font-weight: 600; cursor: pointer; transition: all var(--transition);
}
.btn-primary { background: var(--primary); color: white; }
.btn-primary:hover { background: var(--primary-dark); }
.btn-secondary { background: #e2e8f0; color: var(--text); }
.btn-secondary:hover { background: #cbd5e1; }
.btn-outline { background: none; border: 1px solid var(--border); color: var(--danger); }
.btn-outline:hover { background: #fef2f2; }
.btn-outline:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-outline-template { background: none; border: 1px solid var(--primary); color: var(--primary); }
.btn-outline-template:hover { background: #eef2ff; }
.btn-outline-export { background: none; border: 1px solid #16a34a; color: #16a34a; }
.btn-outline-export:hover { background: #f0fdf4; }
.btn-outline-danger { background: none; border: 1px solid #dc2626; color: #dc2626; }
.btn-outline-danger:hover { background: #fef2f2; }
</style>
