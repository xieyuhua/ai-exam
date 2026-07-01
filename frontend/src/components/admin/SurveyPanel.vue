<template>
  <div>
    <div class="section-title">
      <span>📋 问卷管理</span>
      <div class="title-actions">
        <button class="btn btn-primary" @click="openCreate">＋ 新增问卷</button>
      </div>
    </div>

    <!-- 搜索 & 状态 Tab -->
    <div class="toolbar">
      <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input type="text" v-model="keyword" placeholder="搜索问卷标题..." @keyup.enter="load()" />
        <button v-if="keyword" class="clear-btn" @click="keyword=''; load()">✕</button>
      </div>
      <div class="status-tabs">
        <button v-for="t in statusTabs" :key="t.key" :class="['stab', { active: statusFilter === t.key }]" @click="statusFilter = t.key; load()">
          {{ t.label }} <span v-if="t.key === '' && total > 0" class="stab-count">({{ total }})</span>
        </button>
      </div>
    </div>

    <!-- 加载中 -->
    <div class="loader" v-if="loading"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>

    <!-- 空状态 -->
    <div v-else-if="list.length === 0" class="empty-state">
      <div class="empty-icon">📋</div>
      <h3>{{ statusFilter ? '暂无该状态的问卷' : '暂无问卷' }}</h3>
      <p>{{ statusFilter ? '试试切换其他状态筛选' : '点击右上角"新增问卷"创建第一个问卷' }}</p>
    </div>

    <!-- 问卷列表 -->
    <template v-else>
      <div class="survey-grid">
        <div v-for="s in list" :key="s.id" class="survey-card">
          <div class="sc-header">
            <span :class="['status-dot', s.status]"></span>
            <span class="sc-status-text">{{ statusLabel(s.status) }}</span>
            <span class="sc-qcount">📝 {{ s.questionCount || 0 }} 题</span>
          </div>
          <h3 class="sc-title">{{ s.title }}</h3>
          <p class="sc-desc" v-if="s.description">{{ s.description }}</p>
          <div class="sc-meta">
            <span>📅 {{ fmtDate(s.startTime) }} ~ {{ fmtDate(s.endTime) }}</span>
          </div>
          <div class="sc-stats" v-if="s.totalCompleted > 0">
            <span class="sc-completed">👥 {{ s.totalCompleted }} 人已填写</span>
          </div>
          <div class="sc-footer">
            <span class="sc-allow" v-if="s.allowRepeat">🔄 允许重复</span>
            <span v-else></span>
            <div class="sc-actions">
              <button class="btn-text" @click="openEdit(s)">编辑</button>
              <button class="btn-text" @click="viewDetail(s)">统计</button>
              <button class="btn-text export" @click="exportSingle(s)">导出</button>
              <button class="btn-text danger" @click="handleDelete(s)">删除</button>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination" v-if="total > pageSize">
        <button :disabled="page <= 1" @click="page--; load()">‹ 上一页</button>
        <span class="page-info">第 {{ page }}/{{ totalPages }} 页（共 {{ total }} 条）</span>
        <button :disabled="page >= totalPages" @click="page++; load()">下一页 ›</button>
      </div>
    </template>

    <!-- ========== 创建/编辑弹窗 ========== -->
    <Modal :show="modalShow" :title="editing ? '✏️ 编辑问卷' : '📋 新增问卷'" :maxWidth="'780px'" @close="closeModal">
      <div class="form-scroll">
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="fs-title">📌 基本信息</div>
          <div class="form-group"><label>问卷标题 <span class="required">*</span></label><input type="text" v-model="form.title" placeholder="请输入问卷标题" /></div>
          <div class="form-group"><label>描述（可选）</label><textarea v-model="form.description" placeholder="问卷补充说明..." rows="2"></textarea></div>
          <div class="form-row">
            <div class="form-group"><label>开始时间 <span class="required">*</span></label><input type="datetime-local" v-model="form.startTime" /></div>
            <div class="form-group"><label>结束时间 <span class="required">*</span></label><input type="datetime-local" v-model="form.endTime" /></div>
          </div>
          <label class="switch-item">
            <span class="switch-label">允许重复提交</span>
            <label class="toggle"><input type="checkbox" v-model="form.allowRepeat" /><span class="toggle-track"></span></label>
          </label>
        </div>

        <!-- 题目管理 -->
        <div class="form-section">
          <div class="fs-title">📝 问卷题目（{{ form.questions.length }} 题）</div>

          <div class="question-list">
            <div v-for="(q, qi) in form.questions" :key="qi" class="question-card">
              <!-- 题头 -->
              <div class="qc-header">
                <span class="qc-num">{{ qi + 1 }}</span>
                <select v-model="q.type" class="qc-type-select">
                  <option value="single">☉ 单选题</option>
                  <option value="multiple">☐ 多选题</option>
                  <option value="fill">✎ 填空题</option>
                  <option value="essay">📄 简答题</option>
                </select>
                <div class="qc-header-right">
                  <label class="required-label"><input type="checkbox" v-model="q.required" /> 必填</label>
                  <button class="btn-icon danger" @click="form.questions.splice(qi, 1)" :disabled="form.questions.length <= 1" title="删除此题">🗑️</button>
                </div>
              </div>

              <!-- 题目内容 -->
              <div class="qc-body">
                <input type="text" v-model="q.title" placeholder="请输入题目内容..." class="qc-title-input" />
              </div>

              <!-- 选项（仅选择题） -->
              <div v-if="['single', 'multiple'].includes(q.type)" class="qc-options">
                <div class="qc-opt-label">选项设置</div>
                <div v-for="(opt, oi) in q.options" :key="oi" class="option-item">
                  <div class="opt-header">
                    <span class="opt-label-badge">{{ String.fromCharCode(65 + oi) }}</span>
                    <select v-model="opt.type" class="opt-type-select">
                      <option value="text">📝 文字</option><option value="image">🖼️ 图片</option><option value="video">🎬 视频</option>
                    </select>
                    <button class="btn-icon danger" @click="q.options.splice(oi, 1)" :disabled="q.options.length <= 1" title="删除选项">🗑️</button>
                  </div>
                  <div class="opt-input-area">
                    <input v-if="opt.type === 'text'" type="text" v-model="opt.content" :placeholder="`选项 ${String.fromCharCode(65 + oi)} 文字`" />
                    <div v-if="opt.type === 'image'" class="media-input">
                      <input type="text" v-model="opt.url" placeholder="图片链接 URL（如 https://...）" />
                      <div class="media-preview" v-if="opt.url"><img :src="opt.url" /></div>
                    </div>
                    <div v-if="opt.type === 'video'" class="media-input">
                      <input type="text" v-model="opt.url" placeholder="视频链接 URL（支持 mp4 等）" />
                      <div class="media-preview" v-if="opt.url"><video :src="opt.url" controls muted /></div>
                    </div>
                  </div>
                </div>
                <button class="btn-dashed" @click="addQOption(q)">＋ 添加选项</button>
              </div>
            </div>
          </div>

          <button class="btn-dashed" @click="addQuestion">＋ 添加题目</button>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="closeModal">取消</button>
        <button class="btn btn-primary" @click="save">💾 保存问卷</button>
      </div>
    </Modal>

    <!-- ========== 查看统计弹窗 ========== -->
    <Modal :show="detailShow" :title="'📊 ' + detailData.title" :maxWidth="'750px'" @close="detailShow = false">
      <div class="detail-top" v-if="detailData.surveyId">
        <!-- 摘要卡片 -->
        <div class="detail-summary">
          <span :class="['status-tag', detailData.status]">{{ statusLabel(detailData.status || detailData.status) }}</span>
        </div>
        <div class="summary-cards">
          <div class="sum-card">
            <div class="sum-icon">👥</div>
            <div class="sum-info">
              <div class="sum-value">{{ detailData.totalCompleted || 0 }}</div>
              <div class="sum-label">填写人数</div>
            </div>
          </div>
          <div class="sum-card">
            <div class="sum-icon">📝</div>
            <div class="sum-info">
              <div class="sum-value">{{ (detailData.questions || []).length }}</div>
              <div class="sum-label">题目数量</div>
            </div>
          </div>
          <div class="sum-card">
            <div class="sum-icon">💬</div>
            <div class="sum-info">
              <div class="sum-value">{{ detailData.totalResponses || 0 }}</div>
              <div class="sum-label">总答题次数</div>
            </div>
          </div>
        </div>

        <!-- 按题目统计 -->
        <div class="sq-list" v-if="(detailData.questions || []).length > 0">
          <div v-for="(q, qi) in detailData.questions" :key="q.id || qi" class="sq-item">
            <div class="sq-title">
              <span class="sq-num">{{ qi + 1 }}</span>
              <span class="sq-text">{{ q.title }}</span>
              <span :class="['sq-type-tag', q.type]">{{ q.type === 'single' ? '单选' : q.type === 'multiple' ? '多选' : q.type === 'fill' ? '填空' : '简答' }}</span>
              <span v-if="q.required" class="sq-required">必填</span>
            </div>
            <div class="sq-resp-count" v-if="detailData.totalCompleted > 0">
              <span class="sq-resp-label">作答 {{ q.responseCount || 0 }}/{{ detailData.totalCompleted }} 人</span>
            </div>

            <!-- 选择题：选项柱状图 -->
            <div v-if="['single', 'multiple'].includes(q.type) && q.options?.length" class="sq-options">
              <div v-for="(opt, oi) in q.options" :key="oi" class="sq-opt-item">
                <div class="sq-opt-label">
                  <span class="sq-opt-badge">{{ opt.label }}</span>
                  <span>{{ opt.content || (opt.type === 'image' ? '🖼️ 图片' : '🎬 视频') }}</span>
                </div>
                <div class="stat-bar-wrap">
                  <div class="stat-bar" :style="{ width: Math.round(opt.percent || 0) + '%' }">
                    <span class="stat-bar-text" v-if="(opt.percent || 0) > 15">{{ opt.count || 0 }}</span>
                  </div>
                </div>
                <div class="sq-opt-bottom">
                  <span>{{ opt.count || 0 }} 人</span>
                  <span class="sq-opt-pct">{{ Math.round((opt.percent || 0) * 10) / 10 }}%</span>
                </div>
              </div>
            </div>

            <!-- 填空/简答题：文字答案列表 -->
            <div v-if="['fill', 'essay'].includes(q.type) && q.textResponses?.length" class="sq-text-list">
              <div class="sq-text-header">📄 提交内容（最近 {{ q.textResponses.length }} 条）</div>
              <div v-for="(txt, ti) in q.textResponses" :key="ti" class="sq-text-item">{{ txt || '(空)' }}</div>
            </div>
            <div v-if="['fill', 'essay'].includes(q.type) && (!q.textResponses || q.textResponses.length === 0) && detailData.totalCompleted > 0" class="sq-text-empty">
              暂无提交内容
            </div>
          </div>
        </div>
        <div v-else class="empty-state" style="padding:30px"><div class="empty-icon">📭</div><p>暂无统计信息</p></div>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, showToast } from './shared'
import Modal from '@/components/Modal.vue'

const list = ref([])
const keyword = ref('')
const statusFilter = ref('')
const page = ref(1)
const total = ref(0)
const pageSize = 12
const loading = ref(false)
const modalShow = ref(false)
const editing = ref(null)
const detailShow = ref(false)
const detailData = ref({ questions: [] })

const statusTabs = [
  { key: '', label: '全部' },
  { key: 'upcoming', label: '未开始' },
  { key: 'active', label: '进行中' },
  { key: 'ended', label: '已结束' }
]

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const form = ref({
  title: '', description: '', startTime: '', endTime: '', allowRepeat: false,
  questions: [{ title: '', type: 'single', required: true, options: [{ type: 'text', content: '', url: '' }, { type: 'text', content: '', url: '' }] }]
})

// ── 工具函数 ──
function fmtDate(t) {
  if (!t) return '--'
  const d = new Date(t)
  return `${d.getMonth()+1}/${d.getDate()} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}
function toDatetimeLocal(t) {
  if (!t) return ''
  const d = new Date(t)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}
function statusLabel(s) {
  return { active: '🟢 进行中', upcoming: '🟠 未开始', ended: '⚫ 已结束' }[s] || s
}

// ── 数据加载 ──
async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: page.value, pageSize })
    if (keyword.value) params.set('keyword', keyword.value)
    if (statusFilter.value) params.set('status', statusFilter.value)
    const res = await api(`/admin/surveys?${params}`)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) { showToast('加载失败: ' + e.message, 'error') }
  finally { loading.value = false }
}

// ── 创建 / 编辑 ──
function openCreate() {
  editing.value = null
  form.value = {
    title: '', description: '', startTime: '', endTime: '', allowRepeat: false,
    questions: [{ title: '', type: 'single', required: true, options: [{ type: 'text', content: '', url: '' }, { type: 'text', content: '', url: '' }] }]
  }
  modalShow.value = true
}

function openEdit(s) {
  editing.value = s
  form.value = {
    title: s.title, description: s.description || '',
    startTime: toDatetimeLocal(s.startTime), endTime: toDatetimeLocal(s.endTime),
    allowRepeat: !!s.allowRepeat, questions: []
  }
  loadSurveyDetail(s.id, true)
  modalShow.value = true
}

async function loadSurveyDetail(id, forEdit) {
  try {
    const res = await api(`/admin/surveys/${id}`)
    const data = res.data
    if (forEdit && data.questions) {
      form.value.questions = data.questions.map(q => ({
        title: q.title, type: q.type, required: q.required !== false,
        options: (q.options || []).map(o => ({ type: o.type || 'text', content: o.content || '', url: o.url || '' }))
      }))
      if (form.value.questions.length === 0) {
        form.value.questions = [{ title: '', type: 'single', required: true, options: [{ type: 'text', content: '', url: '' }, { type: 'text', content: '', url: '' }] }]
      }
    }
  } catch (e) { /* ignore */ }
}

function closeModal() { modalShow.value = false }

function addQuestion() {
  form.value.questions.push({ title: '', type: 'single', required: true, options: [{ type: 'text', content: '', url: '' }, { type: 'text', content: '', url: '' }] })
}
function addQOption(q) {
  q.options.push({ type: 'text', content: '', url: '' })
}

async function save() {
  if (!form.value.title.trim()) return showToast('请输入标题', 'error')
  if (form.value.questions.some(q => !q.title.trim())) return showToast('请填写所有题目内容', 'error')
  const data = {
    title: form.value.title, description: form.value.description,
    startTime: form.value.startTime ? form.value.startTime + ':00' : '',
    endTime: form.value.endTime ? form.value.endTime + ':00' : '',
    allowRepeat: form.value.allowRepeat,
    questions: form.value.questions.map((q, qi) => ({
      title: q.title, type: q.type, required: q.required !== false, sortOrder: qi,
      options: ['single', 'multiple'].includes(q.type)
        ? q.options.map((o, oi) => ({ label: String.fromCharCode(65 + oi), type: o.type, content: o.content, url: o.url }))
        : []
    }))
  }
  try {
    const url = editing.value ? `/admin/surveys/${editing.value.id}` : '/admin/surveys'
    const method = editing.value ? 'PUT' : 'POST'
    const res = await api(url, { method, body: JSON.stringify(data) })
    if (res.code === 200) { showToast(editing.value ? '更新成功' : '创建成功', 'success'); closeModal(); load() }
    else showToast(res.message || '操作失败', 'error')
  } catch (e) { showToast('操作失败', 'error') }
}

// ── 查看统计 ──
async function viewDetail(s) {
  try {
    const res = await api(`/admin/surveys/${s.id}/statistics`)
    const stats = res.data || { questions: [] }
    // 合并问卷基本信息
    detailData.value = {
      ...stats,
      title: s.title,
      status: s.status,
      startTime: s.startTime,
      endTime: s.endTime
    }
    detailShow.value = true
  } catch (e) { showToast('加载统计失败', 'error') }
}

// ── 导出单个问卷 ──
async function exportSingle(s, format = 'xlsx') {
  try {
    const url = `/api/admin/surveys/${s.id}/export?format=${format}`
    const resp = await fetch(url, { credentials: 'include' })
    if (!resp.ok) { showToast('导出失败', 'error'); return }
    const blob = await resp.blob()
    const safeName = (s.title || '问卷').replace(/[/\\:*?"<>|]/g, '_')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `${safeName}_问卷数据.${format}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(a.href)
    showToast('导出成功', 'success')
  } catch (e) { showToast('导出失败', 'error') }
}

// ── 删除 ──
async function handleDelete(s) {
  if (!confirm(`确定删除问卷「${s.title}」？\n此操作不可撤销！`)) return
  try {
    const r = await api(`/admin/surveys/${s.id}`, { method: 'DELETE' })
    if (r.code === 200) { showToast('删除成功', 'success'); load() }
    else showToast(r.message || '删除失败', 'error')
  } catch (e) { showToast('删除失败', 'error') }
}

defineExpose({ load })
onMounted(() => load())
</script>

<style scoped>
/* ── 基础 ── */
.section-title {
  font-size: 20px; font-weight: 700; margin-bottom: 20px; padding-bottom: 14px;
  border-bottom: 2px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center;
}
.title-actions { display: flex; align-items: center; gap: 10px; }

/* ── 工具栏 ── */
.toolbar { display: flex; gap: 12px; align-items: center; margin-bottom: 20px; flex-wrap: wrap; }
.search-wrap {
  position: relative; flex: 1; min-width: 200px; max-width: 320px;
  display: flex; align-items: center;
}
.search-wrap .search-icon { position: absolute; left: 12px; font-size: 14px; opacity: 0.5; }
.search-wrap input {
  width: 100%; padding: 10px 36px 10px 36px; border: 1px solid #e2e8f0; border-radius: 10px;
  font-size: 14px; background: #f8fafc; transition: all 0.2s;
}
.search-wrap input:focus { outline: none; border-color: #a5b4fc; background: #fff; box-shadow: 0 0 0 3px rgba(99,102,241,0.08); }
.clear-btn {
  position: absolute; right: 10px; background: none; border: none; cursor: pointer;
  color: #94a3b8; font-size: 14px; padding: 4px; line-height: 1;
}
.clear-btn:hover { color: #ef4444; }

/* 状态 Tabs */
.status-tabs { display: flex; gap: 4px; background: #f1f5f9; border-radius: 10px; padding: 3px; }
.stab {
  padding: 7px 16px; border: none; border-radius: 8px; background: transparent;
  cursor: pointer; font-size: 13px; font-weight: 500; color: #64748b; transition: all 0.2s;
}
.stab:hover { color: #334155; }
.stab.active { background: #fff; color: #059669; box-shadow: 0 1px 3px rgba(0,0,0,0.08); font-weight: 600; }
.stab-count { font-size: 11px; opacity: 0.7; }

/* ── 加载 / 空状态 ── */
.loader { text-align: center; padding: 48px 0; }
.loader-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #059669; margin: 0 4px; animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.2s; }
.loader-dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce { to { transform: translateY(-10px); opacity: 0.4; } }
.empty-state { text-align: center; padding: 48px 0; color: #94a3b8; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-state h3 { color: #64748b; margin-bottom: 6px; }
.empty-state p { font-size: 13px; }

/* ── 卡片网格 ── */
.survey-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px; }
.survey-card {
  background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px;
  transition: all 0.2s; display: flex; flex-direction: column;
}
.survey-card:hover { border-color: #a7f3d0; box-shadow: 0 4px 16px rgba(5,150,105,0.06); transform: translateY(-1px); }

.sc-header { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.status-dot.active { background: #22c55e; animation: pulse 2s infinite; }
.status-dot.upcoming { background: #f59e0b; }
.status-dot.ended { background: #94a3b8; }
@keyframes pulse { 0%, 100% { box-shadow: 0 0 0 0 rgba(34,197,94,0.4); } 50% { box-shadow: 0 0 0 6px rgba(34,197,94,0); } }
.sc-status-text { font-size: 12px; font-weight: 600; }
.sc-qcount { font-size: 12px; color: #94a3b8; margin-left: auto; }

.sc-title { font-size: 16px; font-weight: 700; margin-bottom: 6px; line-height: 1.4; word-break: break-word; }
.sc-desc { font-size: 13px; color: #94a3b8; margin-bottom: 12px; line-height: 1.5; word-break: break-word; }
.sc-meta { font-size: 12px; color: #94a3b8; margin-bottom: 6px; }
.sc-stats { margin-top: auto; padding-top: 12px; }
.sc-completed { font-size: 13px; font-weight: 500; color: #059669; }
.sc-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 14px; padding-top: 12px; border-top: 1px solid #f1f5f9; }
.sc-allow { font-size: 12px; color: #94a3b8; }
.sc-actions { display: flex; gap: 4px; }

.btn-text {
  padding: 6px 12px; border: none; background: none; cursor: pointer;
  font-size: 13px; color: #059669; border-radius: 6px; transition: all 0.15s; font-weight: 500;
}
.btn-text:hover { background: #ecfdf5; }
.btn-text.danger { color: #ef4444; }
.btn-text.danger:hover { background: #fef2f2; }
.btn-text.export { color: #f59e0b; }
.btn-text.export:hover { background: #fffbeb; }

/* ── 弹窗表单 ── */
.form-scroll { max-height: calc(80vh - 140px); overflow-y: auto; padding-right: 4px; }
.form-section { margin-bottom: 24px; }
.fs-title { font-size: 14px; font-weight: 700; color: #475569; margin-bottom: 14px; padding-bottom: 8px; border-bottom: 1px dashed #e2e8f0; }
.form-group { margin-bottom: 14px; }
.form-group label { display: block; font-size: 13px; font-weight: 600; color: #475569; margin-bottom: 6px; }
.required { color: #ef4444; }
.form-group input[type="text"],
.form-group input[type="datetime-local"],
.form-group textarea,
.form-group select {
  width: 100%; padding: 10px 14px; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: 14px; transition: all 0.2s; background: #f8fafc;
}
.form-group input:focus, .form-group textarea:focus, .form-group select:focus {
  outline: none; border-color: #a5b4fc; background: #fff; box-shadow: 0 0 0 3px rgba(99,102,241,0.06);
}
.form-group textarea { resize: vertical; font-family: inherit; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }

/* Switch */
.switch-item { display: flex; align-items: center; gap: 10px; cursor: pointer; margin-top: 4px; }
.switch-label { font-size: 13px; font-weight: 500; color: #475569; }
.toggle { position: relative; display: inline-block; width: 42px; height: 24px; cursor: pointer; }
.toggle input { display: none; }
.toggle-track {
  position: absolute; inset: 0; background: #cbd5e1; border-radius: 12px; transition: all 0.3s;
}
.toggle-track::after {
  content: ''; position: absolute; top: 2px; left: 2px; width: 20px; height: 20px;
  background: #fff; border-radius: 50%; transition: all 0.3s; box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}
.toggle input:checked + .toggle-track { background: #059669; }
.toggle input:checked + .toggle-track::after { transform: translateX(18px); }

/* ── 题目卡片 ── */
.question-card { border: 1px solid #e2e8f0; border-radius: 12px; padding: 18px; margin-bottom: 14px; background: #fff; transition: border-color 0.2s; }
.question-card:hover { border-color: #a5b4fc; }
.qc-header { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
.qc-num {
  width: 30px; height: 30px; display: flex; align-items: center; justify-content: center;
  background: #eef2ff; color: #6366f1; font-weight: 700; font-size: 14px; border-radius: 8px; flex-shrink: 0;
}
.qc-type-select { padding: 7px 10px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 13px; background: #f8fafc; }
.qc-header-right { display: flex; align-items: center; gap: 10px; margin-left: auto; }
.required-label { display: flex; align-items: center; gap: 4px; font-size: 13px; cursor: pointer; color: #64748b; }
.qc-body { margin-bottom: 12px; }
.qc-title-input { width: 100%; padding: 9px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background: #f8fafc; }
.qc-title-input:focus { outline: none; border-color: #a5b4fc; background: #fff; }
.qc-opt-label { font-size: 12px; font-weight: 600; color: #94a3b8; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px; }

/* 选项共用样式 */
.option-item { border: 1px solid #e2e8f0; border-radius: 8px; padding: 12px; margin-bottom: 8px; background: #fafbfc; }
.opt-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.opt-label-badge {
  width: 26px; height: 26px; display: flex; align-items: center; justify-content: center;
  background: #eef2ff; color: #6366f1; font-weight: 700; font-size: 12px; border-radius: 6px; flex-shrink: 0;
}
.opt-type-select { padding: 5px 8px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 12px; background: #f8fafc; }
.opt-input-area input[type="text"] { width: 100%; padding: 8px 10px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; }
.opt-input-area input[type="text"]:focus { outline: none; border-color: #a5b4fc; }
.media-input input { margin-bottom: 6px; }
.media-preview img, .media-preview video { max-width: 160px; max-height: 90px; border-radius: 6px; border: 1px solid #e2e8f0; }
.btn-icon { width: 30px; height: 30px; display: flex; align-items: center; justify-content: center; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 13px; }
.btn-icon:hover:not(:disabled) { border-color: #ef4444; background: #fef2f2; }
.btn-icon:disabled { opacity: 0.3; cursor: not-allowed; }
.btn-dashed {
  width: 100%; padding: 11px; border: 2px dashed #e2e8f0; border-radius: 10px; background: #f8fafc;
  cursor: pointer; font-size: 13px; color: #64748b; transition: all 0.2s;
}
.btn-dashed:hover { border-color: #6366f1; color: #6366f1; background: #eef2ff; }

/* ── 弹窗底部 ── */
.modal-footer { display: flex; gap: 10px; justify-content: flex-end; padding-top: 16px; border-top: 1px solid #f1f5f9; margin-top: 4px; }

/* ── 统计弹窗 ── */
.detail-top { padding: 4px 0; }
.detail-summary { display: flex; gap: 10px; align-items: center; margin-bottom: 14px; flex-wrap: wrap; }
.status-tag { padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.status-tag.active { background: #f0fdf4; color: #16a34a; }
.status-tag.upcoming { background: #fffbeb; color: #d97706; }
.status-tag.ended { background: #f1f5f9; color: #64748b; }
.detail-total { font-size: 14px; font-weight: 500; color: #059669; }

/* 摘要卡片 */
.summary-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 22px; }
.sum-card {
  display: flex; align-items: center; gap: 12px; padding: 16px; border-radius: 12px;
  background: linear-gradient(135deg, #f0fdf4, #ecfdf5); border: 1px solid #a7f3d0;
}
.sum-icon { font-size: 28px; flex-shrink: 0; }
.sum-value { font-size: 18px; font-weight: 700; color: #1e293b; line-height: 1.3; }
.sum-label { font-size: 11px; color: #64748b; margin-top: 2px; }

/* 题目列表 */
.sq-list { display: grid; gap: 16px; }
.sq-item { padding: 18px; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; }
.sq-item:hover { border-color: #a7f3d0; }
.sq-title { display: flex; align-items: center; gap: 8px; font-weight: 600; margin-bottom: 8px; flex-wrap: wrap; }
.sq-num {
  width: 26px; height: 26px; display: flex; align-items: center; justify-content: center;
  background: #ecfdf5; color: #059669; font-weight: 700; font-size: 13px; border-radius: 8px; flex-shrink: 0;
}
.sq-text { flex: 1; word-break: break-word; }
.sq-type-tag { font-size: 11px; padding: 2px 8px; border-radius: 10px; font-weight: 600; }
.sq-type-tag.single { background: #eef2ff; color: #6366f1; }
.sq-type-tag.multiple { background: #f0fdf4; color: #059669; }
.sq-type-tag.fill { background: #fffbeb; color: #d97706; }
.sq-type-tag.essay { background: #fef2f2; color: #ef4444; }
.sq-required { font-size: 11px; padding: 2px 8px; border-radius: 10px; background: #fef2f2; color: #ef4444; }
.sq-resp-count { margin-bottom: 10px; }
.sq-resp-label { font-size: 12px; color: #94a3b8; }

/* 选项统计条 */
.sq-options { padding-left: 4px; }
.sq-opt-item { margin-bottom: 10px; }
.sq-opt-label { display: flex; align-items: center; gap: 6px; font-size: 13px; margin-bottom: 5px; }
.sq-opt-badge {
  width: 24px; height: 24px; display: flex; align-items: center; justify-content: center;
  background: #eef2ff; color: #6366f1; font-weight: 700; font-size: 11px; border-radius: 6px; flex-shrink: 0;
}
.stat-bar-wrap { height: 26px; background: #f1f5f9; border-radius: 13px; overflow: hidden; }
.stat-bar {
  height: 100%; border-radius: 13px; min-width: 4px;
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex; align-items: center; padding-left: 12px;
  background: linear-gradient(90deg, #059669, #34d399);
}
.stat-bar-text { font-size: 11px; color: #fff; font-weight: 700; text-shadow: 0 1px 2px rgba(0,0,0,0.15); }
.sq-opt-bottom { display: flex; justify-content: space-between; margin-top: 4px; font-size: 12px; color: #64748b; }
.sq-opt-pct { color: #059669; font-weight: 600; }

/* 文本答案列表 */
.sq-text-list { margin-top: 10px; padding: 10px 12px; background: #f8fafc; border-radius: 8px; }
.sq-text-header { font-size: 12px; color: #94a3b8; margin-bottom: 8px; font-weight: 600; }
.sq-text-item { padding: 6px 10px; background: #fff; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; margin-bottom: 4px; color: #334155; word-break: break-word; }
.sq-text-empty { margin-top: 8px; font-size: 12px; color: #94a3b8; font-style: italic; }

/* ── 分页 ── */
.pagination { display: flex; justify-content: center; align-items: center; gap: 16px; margin-top: 28px; padding-top: 16px; border-top: 1px solid #f1f5f9; }
.pagination button { padding: 8px 18px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 13px; font-weight: 500; color: #475569; transition: all 0.2s; }
.pagination button:hover:not(:disabled) { border-color: #059669; color: #059669; }
.pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
.page-info { font-size: 13px; color: #94a3b8; }

/* ── 响应式 ── */
@media (max-width: 768px) {
  .toolbar { flex-direction: column; align-items: stretch; }
  .search-wrap { max-width: none; }
  .status-tabs { overflow-x: auto; -webkit-overflow-scrolling: touch; }
  .survey-grid { grid-template-columns: 1fr; }
  .form-row { grid-template-columns: 1fr; }
}
</style>
