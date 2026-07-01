<template>
  <div>
    <div class="section-title">
      <span>📋 考试列表</span>
      <button class="btn btn-primary" @click="openCreate">＋ 创建考试</button>
    </div>

    <!-- 状态 Tab + 搜索 -->
    <div class="toolbar">
      <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input type="text" v-model="search" placeholder="搜索考试名称..." @keyup.enter="load()" />
        <button v-if="search" class="clear-btn" @click="search=''; load()">✕</button>
      </div>
      <div class="status-tabs">
        <button v-for="t in statusTabs" :key="t.key" :class="['stab', { active: statusFilter === t.key }]" @click="statusFilter = t.key; load()">
          {{ t.label }}
        </button>
      </div>
      <div class="cat-filter">
        <select v-model="categoryFilter" @change="load()">
          <option value="">全部分类</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="loader"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>

    <div v-else-if="exams.length === 0" class="empty-state">
      <div class="empty-icon">{{ statusFilter ? '📋' : '📋' }}</div>
      <h3>{{ statusFilter ? '暂无该状态的考试' : '暂无考试' }}</h3>
      <p v-if="!statusFilter">点击「创建考试」添加新的考试</p>
    </div>

    <div v-else class="exam-grid">
      <div v-for="e in exams" :key="e.id" class="exam-card" :class="'status-' + e.status">
        <div class="exam-card-header">
          <div class="exam-title-row">
            <h3>{{ e.title }}</h3>
            <span :class="['badge', statusClass(e.status)]">{{ statusText(e.status) }}</span>
          </div>
          <div class="exam-cat" v-if="e.categoryName">{{ e.categoryName }}</div>
        </div>
        <div class="exam-meta">
          <div class="meta-item"><span class="meta-icon">📝</span>{{ e.questionCount || 0 }}题</div>
          <div class="meta-item"><span class="meta-icon">⏱</span>{{ e.duration }}分钟</div>
          <div class="meta-item"><span class="meta-icon">💯</span>{{ (e.actualScore > 0 ? e.actualScore : e.totalScore) || 0 }}分</div>
        </div>
        <div class="exam-time">
          <span>📅 {{ formatTime(e.startTime) }} ~ {{ formatTime(e.endTime) }}</span>
        </div>
        <div class="exam-actions">
          <button class="btn btn-outline btn-sm" @click="$emit('manageQuestions', e)">管理考题</button>
          <button class="btn btn-export btn-sm" @click="exportScores(e)">导出成绩</button>
          <button class="btn btn-secondary btn-sm" @click="openEdit(e)">编辑</button>
          <button class="btn btn-danger btn-sm" @click="handleDelete(e)">删除</button>
        </div>
      </div>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <button :disabled="page <= 1" @click="page--; load()">上一页</button>
      <span class="page-info">第 {{ page }}/{{ totalPages }} 页（共 {{ total }} 条）</span>
      <button :disabled="page >= totalPages" @click="page++; load()">下一页</button>
    </div>

    <!-- 创建/编辑弹窗 -->
    <Modal :show="modalShow" :title="editing ? '编辑考试' : '创建考试'" :maxWidth="'600px'" @close="closeModal">
      <div class="form-section">
        <h4 class="form-section-title">基本信息</h4>
        <div class="form-group"><label>所属分类</label><select v-model="form.categoryId"><option value="">请选择分类</option><option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option></select></div>
        <div class="form-group"><label>考试名称</label><input type="text" v-model="form.title" placeholder="请输入考试名称" /></div>
        <div class="form-row">
          <div class="form-group"><label>考试时长（分钟）</label><input type="number" v-model="form.duration" placeholder="30" min="1" /></div>
          <div class="form-group"><label>总分</label><input type="number" v-model="form.totalScore" placeholder="100" min="1" /></div>
        </div>
      </div>
      <div class="form-section">
        <h4 class="form-section-title">时间设置</h4>
        <div class="form-row">
          <div class="form-group"><label>开始时间</label><input type="datetime-local" v-model="form.startTime" /></div>
          <div class="form-group"><label>结束时间</label><input type="datetime-local" v-model="form.endTime" /></div>
        </div>
      </div>
      <div class="form-section">
        <h4 class="form-section-title">考试规则</h4>
        <div class="form-row">
          <div class="form-group">
            <label>查看答案解析</label>
            <div class="toggle-row">
              <label class="toggle-switch">
                <input type="checkbox" v-model="form.canViewAnswer" />
                <span class="toggle-slider"></span>
              </label>
              <span class="toggle-label">{{ form.canViewAnswer ? '允许' : '不允许' }}</span>
            </div>
          </div>
          <div class="form-group">
            <label>允许重复考试</label>
            <div class="toggle-row">
              <label class="toggle-switch">
                <input type="checkbox" v-model="form.allowRepeat" />
                <span class="toggle-slider"></span>
              </label>
              <span class="toggle-label">{{ form.allowRepeat ? '允许（可多次考试）' : '不允许（每人仅一次）' }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="form-hint">💡 创建考试后，可在考试列表中点击「管理考题」按钮添加考题</div>
      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="closeModal">取消</button>
        <button class="modal-btn confirm" @click="save">{{ editing ? '保存修改' : '创建考试' }}</button>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, showToast } from './shared'
import Modal from '@/components/Modal.vue'

const props = defineProps({ categories: { type: Array, default: () => [] } })
const emit = defineEmits(['changed', 'manageQuestions'])

const exams = ref([])
const search = ref('')
const categoryFilter = ref('')
const statusFilter = ref('')
const page = ref(1)
const total = ref(0)
const pageSize = 12
const modalShow = ref(false)
const editing = ref(null)
const loading = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const statusTabs = [
  { key: '', label: '全部' },
  { key: 'upcoming', label: '未开始' },
  { key: 'active', label: '进行中' },
  { key: 'ended', label: '已结束' }
]

const form = ref(getEmptyForm())

function getEmptyForm() {
  return { title: '', categoryId: '', duration: 30, totalScore: 100, startTime: '', endTime: '', canViewAnswer: true, allowRepeat: false }
}

function formatTime(t) {
  if (!t) return '--'
  return t.replace('T', ' ').substring(0, 16)
}

function statusClass(s) {
  return { active: 'green', upcoming: 'orange', ended: 'gray' }[s] || 'gray'
}
function statusText(s) {
  return { active: '进行中', upcoming: '未开始', ended: '已结束' }[s] || '未知'
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: page.value, pageSize })
    if (categoryFilter.value) params.set('categoryId', categoryFilter.value)
    if (search.value) params.set('keyword', search.value)
    if (statusFilter.value) params.set('status', statusFilter.value)
    const res = await api(`/admin/exams?${params.toString()}`)
    exams.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function openCreate() { editing.value = null; form.value = getEmptyForm(); modalShow.value = true }
function openEdit(e) {
  editing.value = e
  form.value = {
    title: e.title, categoryId: e.categoryId || '', duration: e.duration, totalScore: e.totalScore,
    startTime: e.startTime ? e.startTime.substring(0, 16) : '',
    endTime: e.endTime ? e.endTime.substring(0, 16) : '',
    canViewAnswer: e.canViewAnswer !== false, allowRepeat: !!e.allowRepeat
  }
  modalShow.value = true
}
function closeModal() { modalShow.value = false }

async function save() {
  if (!form.value.title.trim()) return showToast('请输入考试名称', 'error')
  if (!form.value.categoryId) return showToast('请选择分类', 'error')
  if (!form.value.startTime || !form.value.endTime) return showToast('请设置开始和结束时间', 'error')
  if (form.value.startTime >= form.value.endTime) return showToast('结束时间必须晚于开始时间', 'error')

  const data = {
    title: form.value.title, categoryId: Number(form.value.categoryId) || 0,
    duration: Number(form.value.duration), totalScore: Number(form.value.totalScore),
    startTime: form.value.startTime, endTime: form.value.endTime,
    canViewAnswer: form.value.canViewAnswer, allowRepeat: form.value.allowRepeat
  }
  if (editing.value) {
    const res = await api(`/admin/exams/${editing.value.id}`, { method: 'PUT', body: JSON.stringify(data) })
    if (res.code === 200) { showToast('更新成功', 'success'); closeModal(); load(); emit('changed') }
    else showToast(res.message || '更新失败', 'error')
  } else {
    const res = await api('/admin/exams', { method: 'POST', body: JSON.stringify(data) })
    if (res.code === 200) { showToast('创建成功', 'success'); closeModal(); load(); emit('changed') }
    else showToast(res.message || '创建失败', 'error')
  }
}

async function exportScores(e, format = 'xlsx') {
  try {
    const url = `/api/admin/exams/${e.id}/scores/export?format=${format}`
    const resp = await fetch(url, { credentials: 'include' })
    if (!resp.ok) { showToast('导出失败', 'error'); return }
    const blob = await resp.blob()
    const safeName = (e.title || '考试').replace(/[/\\\\:*?"<>|]/g, '_')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `${safeName}_成绩.${format}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(a.href)
    showToast('导出成功', 'success')
  } catch (err) { showToast('导出失败', 'error') }
}

async function handleDelete(e) {
  if (!confirm(`确定删除考试「${e.title}」吗？此操作不可撤销，将同时删除关联考题和成绩数据。`)) return
  const res = await api(`/admin/exams/${e.id}`, { method: 'DELETE' })
  if (res.code === 200) { showToast('删除成功', 'success'); load(); emit('changed') }
  else showToast(res.message || '删除失败', 'error')
}

defineExpose({ load })
onMounted(load)
</script>

<style scoped>
.section-title {
  font-size: 20px; font-weight: 700; margin-bottom: 20px; padding-bottom: 14px;
  border-bottom: 2px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center;
}
/* 工具栏 */
.toolbar {
  display: flex; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 20px;
}
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

.status-tabs { display: flex; gap: 4px; }
.stab {
  padding: 8px 16px; border: 2px solid var(--border); border-radius: 8px;
  background: #fff; cursor: pointer; font-size: 13px; font-weight: 600;
  color: var(--text-secondary); transition: all var(--transition); white-space: nowrap;
}
.stab:hover { border-color: var(--primary-light); color: var(--primary); }
.stab.active { background: var(--primary); border-color: var(--primary); color: #fff; }

.cat-filter select {
  padding: 8px 12px; border: 2px solid var(--border); border-radius: var(--radius-sm);
  font-size: 13px; color: var(--text); background: #fff; cursor: pointer; min-width: 120px;
}
.cat-filter select:focus { outline: none; border-color: var(--primary); }

/* 卡片网格 */
.exam-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px; }
.exam-card {
  background: var(--white); border: 1px solid var(--border); border-radius: 12px;
  padding: 20px 22px; transition: all var(--transition); position: relative; overflow: hidden;
}
.exam-card:hover { box-shadow: var(--shadow); transform: translateY(-2px); }
.exam-card.status-active { border-left: 4px solid #10b981; }
.exam-card.status-upcoming { border-left: 4px solid #f59e0b; }
.exam-card.status-ended { border-left: 4px solid #94a3b8; }

.exam-card-header { margin-bottom: 12px; }
.exam-title-row { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
.exam-title-row h3 { font-size: 16px; color: var(--text); font-weight: 700; margin: 0; }
.exam-cat { font-size: 12px; color: var(--primary); background: #eef2ff; display: inline-block; padding: 2px 10px; border-radius: 10px; margin-top: 4px; }

.badge { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 700; }
.badge.green { background: #ecfdf5; color: #059669; }
.badge.orange { background: #fffbeb; color: #d97706; }
.badge.gray { background: #f1f5f9; color: #64748b; }

.exam-meta { display: flex; gap: 16px; flex-wrap: wrap; margin-bottom: 8px; }
.meta-item { font-size: 13px; color: var(--text-secondary); display: flex; align-items: center; gap: 4px; }
.meta-icon { font-size: 14px; }

.exam-time { font-size: 12px; color: #94a3b8; margin-bottom: 14px; }

.exam-actions { display: flex; gap: 8px; }
.btn-export {
  background: linear-gradient(135deg, #f59e0b, #f97316); color: #fff;
}
.btn-export:hover { box-shadow: 0 4px 14px rgba(245,158,11,0.35); }

/* 表单优化 */
.form-section { margin-bottom: 20px; }
.form-section-title {
  font-size: 14px; font-weight: 700; color: var(--text);
  margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid #f1f5f9;
}
.form-hint { color: #94a3b8; font-size: 12px; margin-bottom: 10px; }

/* Toggle 开关 */
.toggle-row { display: flex; align-items: center; gap: 10px; margin-top: 8px; }
.toggle-switch { position: relative; display: inline-block; width: 44px; height: 24px; }
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider {
  position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0;
  background: #e2e8f0; border-radius: 24px; transition: all 0.3s;
}
.toggle-slider::before {
  content: ""; position: absolute; height: 18px; width: 18px;
  left: 3px; bottom: 3px; background: white; border-radius: 50%; transition: all 0.3s;
}
.toggle-switch input:checked + .toggle-slider { background: #6366f1; }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(20px); }
.toggle-label { font-size: 13px; color: var(--text-secondary); }

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

/* Loader */
.loader { display: flex; gap: 6px; justify-content: center; padding: 40px 0; }
.loader-dot { width: 10px; height: 10px; border-radius: 50%; background: var(--primary); animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.15s; }
.loader-dot:nth-child(3) { animation-delay: 0.3s; }
@keyframes bounce { to { opacity: 0.3; transform: translateY(-8px); } }

@media (max-width: 768px) {
  .exam-grid { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .search-wrap { max-width: 100%; }
}
</style>
