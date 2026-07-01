<template>
  <div>
    <div class="section-title">
      <span>📝 全部考题</span>
      <div class="title-actions">
        <button class="btn btn-primary" @click="openCreate">＋ 添加考题</button>
        <button class="btn btn-secondary" @click="triggerImport">📥 批量导入</button>
        <button class="btn btn-success" @click="downloadTemplate">📄 下载模板</button>
      </div>
    </div>

    <!-- 搜索 & 筛选 -->
    <div class="toolbar">
      <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input type="text" v-model="search" placeholder="搜索题目内容..." @keyup.enter="load()" />
        <button v-if="search" class="clear-btn" @click="search=''; load()">✕</button>
      </div>
      <select v-model="categoryFilter" class="cat-select" @change="load()">
        <option value="">全部分类</option>
        <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
      <select v-model="typeFilter" class="type-select" @change="load()">
        <option value="">全部类型</option>
        <option value="single">单选题</option>
        <option value="multiple">多选题</option>
        <option value="judge">判断题</option>
        <option value="fill">填空题</option>
        <option value="essay">简答题</option>
      </select>
    </div>

    <div v-if="loading" class="loader"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>

    <div v-else-if="questions.length === 0" class="empty-state">
      <div class="empty-icon">{{ typeFilter ? '📝' : '📝' }}</div>
      <h3>{{ typeFilter ? '暂无该类型的考题' : '暂无考题' }}</h3>
      <p v-if="!typeFilter">点击「添加考题」或「批量导入」添加题目</p>
    </div>

    <div v-else class="question-list">
      <div v-for="q in questions" :key="q.id" class="question-card">
        <div class="q-header">
          <div class="q-title-row">
            <span class="q-id">#{{ q.id }}</span>
            <h4>{{ q.title }}</h4>
          </div>
          <span :class="['q-type', q.type]">{{ typeLabel(q.type) }}</span>
        </div>
        <div v-if="q.options && q.options.length > 0" class="q-options">
          <span v-for="opt in q.options" :key="opt.label" :class="['opt-tag', isCorrectAnswer(q, opt.label) ? 'opt-correct' : '']">
            {{ opt.label }}. {{ opt.text || opt.content }}
          </span>
        </div>
        <div class="q-answer">
          ✅ 答案：{{ formatAnswer(q) }}
        </div>
        <div v-if="q.explanation" class="q-explanation">💡 {{ q.explanation }}</div>
        <div class="q-actions">
          <button class="btn btn-secondary btn-sm" @click="openEdit(q)">编辑</button>
          <button class="btn btn-danger btn-sm" @click="handleDelete(q)">删除</button>
        </div>
      </div>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <button :disabled="page <= 1" @click="page--; load()">上一页</button>
      <span class="page-info">第 {{ page }}/{{ totalPages }} 页（共 {{ total }} 条）</span>
      <button :disabled="page >= totalPages" @click="page++; load()">下一页</button>
    </div>

    <!-- 考题弹窗 -->
    <Modal :show="modalShow" :title="editing ? '编辑考题' : '添加考题'" :maxWidth="'640px'" @close="closeModal">
      <div class="form-section">
        <h4 class="form-section-title">基本信息</h4>
        <div class="form-group" v-if="!editing && !categoryFilter">
          <label>所属分类</label>
          <select v-model="form.categoryId">
            <option value="">请选择分类</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="form-group"><label>题目类型</label><select v-model="form.type" @change="onTypeChange">
          <option value="single">单选题</option><option value="multiple">多选题</option>
          <option value="judge">判断题</option><option value="fill">填空题</option><option value="essay">简答题</option>
        </select></div>
        <div class="form-group"><label>题目内容</label><textarea v-model="form.title" placeholder="请输入题目内容" rows="3"></textarea></div>
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
          <input type="text" v-model="form.answerText" placeholder="请输入正确答案" />
        </div>
      </div>

      <div class="form-section">
        <h4 class="form-section-title">答案解析（可选）</h4>
        <div class="form-group">
          <textarea v-model="form.explanation" placeholder="输入答案解析，帮助员工理解" rows="2"></textarea>
        </div>
      </div>

      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="closeModal">取消</button>
        <button class="modal-btn confirm" @click="save">{{ editing ? '保存修改' : '添加考题' }}</button>
      </div>
    </Modal>

    <!-- 导入弹窗 -->
    <Modal :show="importShow" title="📥 批量导入考题" @close="importShow = false">
      <div v-if="!categoryFilter" class="form-group" style="margin-bottom:16px;">
        <label>导入到分类</label>
        <select v-model="importCategoryId">
          <option value="">请选择分类</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
      <div class="import-area" @click="importFile?.click()">
        <div class="import-icon">📄</div>
        <p>点击选择 Excel 文件 或拖拽到此处</p>
        <button class="btn btn-secondary">选择文件</button>
        <input type="file" ref="importFile" accept=".xlsx,.xls" @change="handleFile" style="display:none;" />
      </div>
      <div v-if="importFileInfo" class="import-preview">
        <div class="file-info"><strong>{{ importFileInfo.name }}</strong></div>
      </div>
      <div class="import-hint">💡 请先下载模板，按模板格式填写后导入</div>
      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="importShow = false">取消</button>
        <button class="modal-btn confirm" :disabled="!importFileInfo" @click="doImport">开始导入</button>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { api, showToast } from './shared'
import Modal from '@/components/Modal.vue'

const props = defineProps({
  categories: { type: Array, default: () => [] }
})
const emit = defineEmits(['changed'])

const questions = ref([])
const search = ref('')
const categoryFilter = ref('')
const typeFilter = ref('')
const page = ref(1)
const total = ref(0)
const pageSize = 12
const modalShow = ref(false)
const editing = ref(null)
const importShow = ref(false)
const importFile = ref(null)
const importFileInfo = ref(null)
const importFileData = ref(null)
const importCategoryId = ref('')
const loading = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
const isTextType = computed(() => ['fill', 'essay'].includes(form.value.type))

const defaultOptions = [
  { label: 'A', text: '' }, { label: 'B', text: '' }, { label: 'C', text: '' }, { label: 'D', text: '' }
]

const form = ref(getEmptyForm())

function getEmptyForm() {
  return { type: 'single', title: '', options: JSON.parse(JSON.stringify(defaultOptions)), answer: [], answerText: '', explanation: '', categoryId: '' }
}

function typeLabel(t) {
  return { single: '单选题', multiple: '多选题', judge: '判断题', fill: '填空题', essay: '简答题' }[t] || t
}

// 解析后端返回的 JSON 字符串字段（options/answer 在 DB 中存为 JSON 字符串）
function parseQuestion(q) {
  let options = []
  let answer = []
  try {
    options = typeof q.options === 'string' ? JSON.parse(q.options || '[]') : (q.options || [])
  } catch (e) { options = [] }
  try {
    answer = typeof q.answer === 'string' ? JSON.parse(q.answer || '[]') : (q.answer || [])
  } catch (e) { answer = [] }
  return { ...q, options, answer }
}

function formatAnswer(q) {
  if (!q.answer) return '--'
  if (Array.isArray(q.answer)) return q.answer.join(', ')
  return q.answer
}

function isCorrectAnswer(q, label) {
  if (!q.answer) return false
  const ans = Array.isArray(q.answer) ? q.answer : [q.answer]
  return ans.some(a => String(a).toUpperCase() === String(label).toUpperCase())
}

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
  // 移除的选项如果被选中，同时清除答案
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

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: page.value, pageSize })
    if (categoryFilter.value) params.set('categoryId', categoryFilter.value)
    if (search.value) params.set('keyword', search.value)
    if (typeFilter.value) params.set('type', typeFilter.value)
    const res = await api(`/admin/questions?${params.toString()}`)
    const raw = res.data?.list || res.data || []
    questions.value = raw.map(parseQuestion)
    total.value = res.data?.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function openCreate() { editing.value = null; form.value = getEmptyForm(); modalShow.value = true }

function openEdit(q) {
  editing.value = q
  const isText = ['fill', 'essay'].includes(q.type)
  // 将 RichOption（content 字段）转为表单格式（text 字段）
  const rawOpts = q.options || []
  const normalizedOpts = rawOpts.length > 0
    ? rawOpts.map(opt => ({ label: String(opt.label).toUpperCase(), text: opt.content || opt.text || '' }))
    : JSON.parse(JSON.stringify(defaultOptions))
  form.value = {
    type: q.type, title: q.title,
    options: normalizedOpts,
    answer: isText ? [] : (Array.isArray(q.answer) ? q.answer.map(a => String(a).toUpperCase()).filter(Boolean) : (q.answer ? [String(q.answer).toUpperCase()] : [])),
    answerText: isText ? (Array.isArray(q.answer) ? q.answer.join(',') : q.answer || '') : '',
    explanation: q.explanation || '',
    categoryId: ''
  }
  modalShow.value = true
}

function closeModal() { modalShow.value = false }

async function save() {
  if (!form.value.title.trim()) return showToast('请输入题目内容', 'error')

  const isText = ['fill', 'essay'].includes(form.value.type)
  if (!isText && form.value.answer.length === 0) return showToast('请选择正确答案', 'error')
  if (isText && !form.value.answerText.trim()) return showToast('请输入正确答案', 'error')

  // 确定分类 ID：优先用弹窗选择的分类，其次用筛选器当前分类
  const cid = form.value.categoryId || categoryFilter.value
  if (!cid) return showToast('请选择所属分类', 'error')

  const data = {
    title: form.value.title,
    type: form.value.type,
    categoryId: Number(cid),
    answer: isText ? (form.value.answerText || '').split(',').filter(Boolean) : (form.value.answer || []),
    explanation: form.value.explanation,
    options: isText ? [] : form.value.options
  }

  if (editing.value) {
    const res = await api(`/admin/questions/${editing.value.id}`, { method: 'PUT', body: JSON.stringify(data) })
    if (res.code === 200) { showToast('更新成功', 'success'); closeModal(); load() }
    else showToast(res.message || '更新失败', 'error')
  } else {
    const res = await api('/admin/questions', { method: 'POST', body: JSON.stringify(data) })
    if (res.code === 200) { showToast('创建成功', 'success'); closeModal(); load() }
    else showToast(res.message || '创建失败', 'error')
  }
}

async function handleDelete(q) {
  if (!confirm(`确定删除这道题目吗？此操作不可撤销。`)) return
  const res = await api(`/admin/questions/${q.id}`, { method: 'DELETE' })
  if (res.code === 200) { showToast('删除成功', 'success'); load() }
  else showToast(res.message || '删除失败', 'error')
}

function triggerImport() { importFileInfo.value = null; importCategoryId.value = ''; importShow.value = true }

function handleFile(e) {
  const file = e.target.files[0]
  if (!file) return
  importFileInfo.value = file
}

async function doImport() {
  if (!importFileInfo.value) return
  try {
    const cid = importCategoryId.value || categoryFilter.value
    if (!cid) return showToast('请先选择分类', 'error')
    const formData = new FormData()
    formData.append('file', importFileInfo.value)
    formData.append('categoryId', cid)
    const token = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)?.[1] || localStorage.getItem('exam_token')
    const res = await fetch(`/api/admin/questions/import`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    })
    const data = await res.json()
    if (data.code === 200) {
      showToast(data.message || '导入成功', 'success')
      importShow.value = false; load()
    } else {
      showToast(data.message || '导入失败', 'error')
    }
  } catch (e) {
    showToast('导入失败：' + e.message, 'error')
  }
}

function downloadTemplate() {
  const token = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)?.[1] || localStorage.getItem('exam_token')
  window.open(`/api/admin/template?token=${encodeURIComponent(token)}`, '_blank')
}

defineExpose({ load })
</script>

<style scoped>
.section-title {
  font-size: 20px; font-weight: 700; margin-bottom: 20px; padding-bottom: 14px;
  border-bottom: 2px solid #f1f5f9;
  display: flex; justify-content: space-between; align-items: center; gap: 10px; flex-wrap: wrap;
}
.title-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

/* 工具栏 */
.toolbar { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 20px; }
.search-wrap { position: relative; flex: 1; min-width: 200px; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); font-size: 14px; }
.search-wrap input {
  width: 100%; padding: 10px 36px 10px 36px; border: 2px solid var(--border);
  border-radius: var(--radius-sm); font-size: 14px; transition: border-color var(--transition);
}
.search-wrap input:focus { outline: none; border-color: var(--primary); }
.clear-btn {
  position: absolute; right: 8px; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer; color: #999; font-size: 16px;
}
.clear-btn:hover { color: #333; }

.cat-select {
  padding: 10px 14px; border: 2px solid var(--border);
  border-radius: var(--radius-sm); font-size: 13px; font-weight: 600;
  color: var(--text); background: #fff; cursor: pointer;
  min-width: 140px; transition: border-color var(--transition);
}
.cat-select:focus { outline: none; border-color: var(--primary); }

.type-select {
  padding: 10px 14px; border: 2px solid var(--border);
  border-radius: var(--radius-sm); font-size: 13px; font-weight: 600;
  color: var(--text); background: #fff; cursor: pointer;
  min-width: 120px; transition: border-color var(--transition);
}
.type-select:focus { outline: none; border-color: var(--primary); }

/* 题目列表 */
.question-list { display: flex; flex-direction: column; gap: 12px; }

.question-card {
  border: 1px solid var(--border); border-radius: 12px;
  padding: 20px 24px; transition: all var(--transition); background: var(--white);
}
.question-card:hover { box-shadow: var(--shadow); border-color: var(--primary-light); }
.q-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; gap: 10px; }
.q-title-row { display: flex; align-items: flex-start; gap: 8px; flex: 1; }
.q-id {
  background: #f1f5f9; color: #64748b; padding: 2px 8px; border-radius: 6px;
  font-size: 12px; font-weight: 700; flex-shrink: 0; margin-top: 2px;
}
.q-title-row h4 { font-size: 15px; color: var(--text); font-weight: 600; margin: 0; line-height: 1.5; }
.q-type { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 700; flex-shrink: 0; }
.q-type.single { background: #eff6ff; color: #3b82f6; }
.q-type.multiple { background: #fff7ed; color: #ea580c; }
.q-type.judge { background: #fef2f2; color: #dc2626; }
.q-type.fill { background: #f0fdf4; color: #16a34a; }
.q-type.essay { background: #faf5ff; color: #9333ea; }

.q-options { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 6px; margin-top: 8px; }
.opt-tag {
  padding: 6px 10px; background: #f8fafc; border-radius: var(--radius-sm);
  border: 1px solid var(--border); font-size: 13px; color: var(--text-secondary);
}
.opt-tag.opt-correct { background: #f0fdf4; border-color: #bbf7d0; color: #059669; font-weight: 600; }

.q-answer { margin-top: 10px; padding: 8px 14px; background: #f0fdf4; border-radius: var(--radius-sm); color: var(--success); font-weight: 600; font-size: 13px; }
.q-explanation { font-size: 13px; color: #94a3b8; margin-top: 5px; }
.q-actions { display: flex; gap: 8px; margin-top: 12px; }

/* 表单 */
.form-section { margin-bottom: 18px; }
.form-section-title {
  font-size: 13px; font-weight: 700; color: var(--text-secondary);
  margin-bottom: 10px; padding-bottom: 6px; border-bottom: 1px solid #f1f5f9;
}

.option-editor { margin-top: 4px; }
.option-row { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.opt-label {
  width: 34px; height: 34px; border-radius: 50%;
  background: #f1f5f9; display: flex; align-items: center;
  justify-content: center; font-weight: 700; color: var(--text-secondary); flex-shrink: 0;
}
.option-row input { flex: 1; padding: 8px 12px; border: 2px solid var(--border); border-radius: var(--radius-sm); font-size: 14px; }
.option-row input:focus { outline: none; border-color: var(--primary); }
.add-option-btn {
  margin-top: 8px; padding: 8px 16px; background: #f8fafc;
  border: 2px dashed var(--border); border-radius: var(--radius-sm);
  color: var(--text-secondary); cursor: pointer; font-size: 13px;
  transition: all var(--transition); width: 100%; font-weight: 500;
}
.add-option-btn:hover { border-color: var(--primary); color: var(--primary); }

/* 答案选择器 */
.answer-radios, .answer-checkboxes { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 4px; }
.answer-radio, .answer-checkbox {
  display: flex; align-items: center; gap: 5px;
  padding: 8px 16px; border: 2px solid var(--border);
  border-radius: var(--radius-sm); cursor: pointer;
  transition: all var(--transition); font-size: 13px; font-weight: 600;
}
.answer-radio:hover, .answer-checkbox:hover { border-color: var(--primary-light); }
.answer-radio.selected, .answer-checkbox.selected { border-color: var(--primary); background: #eef2ff; color: var(--primary); }
.answer-radio input, .answer-checkbox input { display: none; }

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

/* 弹窗 */
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

/* Loader */
.loader { display: flex; gap: 6px; justify-content: center; padding: 40px 0; }
.loader-dot { width: 10px; height: 10px; border-radius: 50%; background: var(--primary); animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.15s; }
.loader-dot:nth-child(3) { animation-delay: 0.3s; }
@keyframes bounce { to { opacity: 0.3; transform: translateY(-8px); } }

@media (max-width: 768px) {
  .toolbar { flex-direction: column; align-items: stretch; }
  .search-wrap { max-width: 100%; }
}
</style>
