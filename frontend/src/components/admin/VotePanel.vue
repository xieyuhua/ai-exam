<template>
  <div>
    <div class="section-title">
      <span>🗳️ 投票管理</span>
      <div class="title-actions">
        <button class="btn btn-primary" @click="openCreate">＋ 新增投票</button>
      </div>
    </div>

    <!-- 搜索 & 状态 Tab -->
    <div class="toolbar">
      <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input type="text" v-model="keyword" placeholder="搜索投票标题..." @keyup.enter="load()" />
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
      <div class="empty-icon">🗳️</div>
      <h3>{{ statusFilter ? '暂无该状态的投票' : '暂无投票' }}</h3>
      <p>{{ statusFilter ? '试试切换其他状态筛选' : '点击右上角"新增投票"创建第一个投票' }}</p>
    </div>

    <!-- 投票列表 -->
    <template v-else>
      <div class="vote-grid">
        <div v-for="v in list" :key="v.id" class="vote-card">
          <div class="vc-header">
            <span :class="['status-dot', v.status]"></span>
            <span class="vc-status-text">{{ statusLabel(v.status) }}</span>
            <span class="vc-type">{{ v.voteType === 'single' ? '单选' : '多选' + (v.maxChoices > 1 ? `·最多${v.maxChoices}项` : '') }}</span>
            <span v-if="v.isPublic" class="vc-public">公开</span>
          </div>
          <h3 class="vc-title">{{ v.title }}</h3>
          <p class="vc-desc" v-if="v.description">{{ v.description }}</p>
          <div class="vc-meta">
            <span>📅 {{ fmtDate(v.startTime) }} ~ {{ fmtDate(v.endTime) }}</span>
          </div>
          <div class="vc-stats" v-if="v.totalVotes > 0">
            <span class="vc-votes">👥 {{ v.totalVotes }} 人已投票</span>
          </div>
          <div class="vc-footer">
            <span class="vc-options-count">{{ v.options?.length || 0 }} 个选项</span>
            <div class="vc-actions">
              <button class="btn-text" @click="openEdit(v)">编辑</button>
              <button class="btn-text" @click="viewDetail(v)">统计</button>
              <button class="btn-text export" @click="exportSingle(v)">导出</button>
              <button class="btn-text danger" @click="handleDelete(v)">删除</button>
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
    <Modal :show="modalShow" :title="editing ? '✏️ 编辑投票' : '🗳️ 新增投票'" :maxWidth="'720px'" @close="closeModal">
      <div class="form-scroll">
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="fs-title">📌 基本信息</div>
          <div class="form-group"><label>投票标题 <span class="required">*</span></label><input type="text" v-model="form.title" placeholder="请输入投票标题" /></div>
          <div class="form-group"><label>描述（可选）</label><textarea v-model="form.description" placeholder="投票补充说明..." rows="2"></textarea></div>
          <div class="form-row">
            <div class="form-group"><label>开始时间 <span class="required">*</span></label><input type="datetime-local" v-model="form.startTime" /></div>
            <div class="form-group"><label>结束时间 <span class="required">*</span></label><input type="datetime-local" v-model="form.endTime" /></div>
          </div>
        </div>

        <!-- 投票规则 -->
        <div class="form-section">
          <div class="fs-title">⚙️ 投票规则</div>
          <div class="form-row form-row-3">
            <div class="form-group">
              <label>投票方式</label>
              <div class="radio-group">
                <label :class="['radio-label', { active: form.voteType === 'single' }]"><input type="radio" v-model="form.voteType" value="single" /> 单选</label>
                <label :class="['radio-label', { active: form.voteType === 'multiple' }]"><input type="radio" v-model="form.voteType" value="multiple" /> 多选</label>
              </div>
            </div>
            <div class="form-group" v-if="form.voteType === 'multiple'">
              <label>最多可选</label>
              <input type="number" v-model.number="form.maxChoices" min="2" max="10" class="input-sm" />
            </div>
          </div>
          <div class="switch-row">
            <label class="switch-item">
              <span class="switch-label">允许多次投票</span>
              <label class="toggle"><input type="checkbox" v-model="form.allowRepeat" /><span class="toggle-track"></span></label>
            </label>
            <label class="switch-item">
              <span class="switch-label">结果公开可见</span>
              <label class="toggle"><input type="checkbox" v-model="form.isPublic" /><span class="toggle-track"></span></label>
            </label>
          </div>
        </div>

        <!-- 选项设置 -->
        <div class="form-section">
          <div class="fs-title">📋 投票选项（{{ form.options.length }} 项）</div>
          <div class="option-list">
            <div v-for="(opt, i) in form.options" :key="i" class="option-item">
              <div class="opt-header">
                <span class="opt-label-badge">{{ String.fromCharCode(65 + i) }}</span>
                <select v-model="opt.type" class="opt-type-select">
                  <option value="text">📝 文字</option><option value="image">🖼️ 图片</option><option value="video">🎬 视频</option>
                </select>
                <button class="btn-icon danger" @click="form.options.splice(i, 1)" :disabled="form.options.length <= 2" title="删除此选项">🗑️</button>
              </div>
              <div class="opt-input-area">
                <input v-if="opt.type === 'text'" type="text" v-model="opt.content" :placeholder="`选项 ${String.fromCharCode(65 + i)} 文字`" />
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
          </div>
          <button class="btn-dashed" @click="addOption">＋ 添加选项</button>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="closeModal">取消</button>
        <button class="btn btn-primary" @click="save">💾 保存投票</button>
      </div>
    </Modal>

    <!-- ========== 查看统计弹窗 ========== -->
    <Modal :show="detailShow" :title="'📊 ' + detailVote.title" :maxWidth="'720px'" @close="detailShow = false">
      <div class="detail-top" v-if="detailVote.id">
        <!-- 摘要卡片 -->
        <div class="detail-summary">
          <span :class="['status-tag', detailVote.status]">{{ statusLabel(detailVote.status) }}</span>
          <span class="detail-vtype">{{ detailVote.voteType === 'single' ? '单选' : '多选' + (detailVote.maxChoices > 1 ? '·最多' + detailVote.maxChoices + '项' : '') }}</span>
        </div>
        <div class="summary-cards">
          <div class="sum-card">
            <div class="sum-icon">👥</div>
            <div class="sum-info">
              <div class="sum-value">{{ detailVote.totalVotes || 0 }}</div>
              <div class="sum-label">投票人数</div>
            </div>
          </div>
          <div class="sum-card" v-if="rankedOptions.length > 0">
            <div class="sum-icon">🏆</div>
            <div class="sum-info">
              <div class="sum-value">{{ rankedOptions[0]?.content || '--' }}</div>
              <div class="sum-label">领先选项</div>
            </div>
          </div>
          <div class="sum-card">
            <div class="sum-icon">📋</div>
            <div class="sum-info">
              <div class="sum-value">{{ (detailVote.options || []).length }}</div>
              <div class="sum-label">选项总数</div>
            </div>
          </div>
        </div>

        <!-- 选项排行 -->
        <div class="stat-grid" v-if="rankedOptions.length > 0">
          <div v-for="(opt, i) in rankedOptions" :key="opt.id || i" :class="['stat-item', 'rank-' + (i + 1)]">
            <div class="stat-label">
              <span :class="['rank-badge', { gold: i === 0, silver: i === 1, bronze: i === 2 }]">
                {{ i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : '#' + (i + 1) }}
              </span>
              <span class="stat-label-text">{{ opt.content || (opt.type === 'image' ? '🖼️ 图片' : '🎬 视频') }}</span>
            </div>
            <div class="stat-media" v-if="opt.type === 'image' && opt.url"><img :src="opt.url" /></div>
            <div class="stat-media" v-if="opt.type === 'video' && opt.url"><video :src="opt.url" controls muted /></div>
            <div class="stat-bar-wrap">
              <div :class="['stat-bar', rankBarClass(i)]" :style="{ width: calcPercent(opt.count, detailVote.totalVotes) + '%' }">
                <span class="stat-bar-text" v-if="calcPercent(opt.count, detailVote.totalVotes) > 14">{{ opt.count || 0 }} 票</span>
              </div>
            </div>
            <div class="stat-bottom">
              <span class="stat-count">{{ opt.count || 0 }} 票</span>
              <span class="stat-percent">{{ calcPercent(opt.count, detailVote.totalVotes) }}%</span>
            </div>
          </div>
        </div>
        <div v-else class="empty-state" style="padding:30px"><div class="empty-icon">📭</div><p>暂无投票数据</p></div>
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
const detailVote = ref({ options: [] })

const statusTabs = [
  { key: '', label: '全部' },
  { key: 'upcoming', label: '未开始' },
  { key: 'active', label: '进行中' },
  { key: 'ended', label: '已结束' }
]

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const defaultOpts = [
  { type: 'text', content: '', url: '' },
  { type: 'text', content: '', url: '' }
]
const form = ref({ title: '', description: '', startTime: '', endTime: '', voteType: 'single', maxChoices: 1, allowRepeat: false, isPublic: true, options: JSON.parse(JSON.stringify(defaultOpts)) })

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
function calcPercent(count, total) {
  if (!total || total === 0) return 0
  return Math.round((count / total) * 100)
}
// 按票数降序排列的选项
const rankedOptions = computed(() => {
  const opts = (detailVote.value.options || []).slice()
  opts.sort((a, b) => (b.count || 0) - (a.count || 0))
  return opts
})
// 排行榜进度条颜色
function rankBarClass(rank) {
  if (rank === 0) return 'bar-gold'
  if (rank === 1) return 'bar-silver'
  if (rank === 2) return 'bar-bronze'
  return ''
}

// ── 数据加载 ──
async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: page.value, pageSize })
    if (keyword.value) params.set('keyword', keyword.value)
    if (statusFilter.value) params.set('status', statusFilter.value)
    const res = await api(`/admin/votes?${params}`)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) { showToast('加载失败: ' + e.message, 'error') }
  finally { loading.value = false }
}

// ── 创建 / 编辑 ──
function openCreate() {
  editing.value = null
  form.value = { title: '', description: '', startTime: '', endTime: '', voteType: 'single', maxChoices: 1, allowRepeat: false, isPublic: true, options: JSON.parse(JSON.stringify(defaultOpts)) }
  modalShow.value = true
}
function openEdit(v) {
  editing.value = v
  form.value = {
    title: v.title, description: v.description || '',
    startTime: toDatetimeLocal(v.startTime), endTime: toDatetimeLocal(v.endTime),
    voteType: v.voteType || 'single', maxChoices: v.maxChoices || 1,
    allowRepeat: !!v.allowRepeat, isPublic: v.isPublic !== false,
    options: v.options ? JSON.parse(JSON.stringify(v.options)) : JSON.parse(JSON.stringify(defaultOpts))
  }
  modalShow.value = true
}
function closeModal() { modalShow.value = false }
function addOption() {
  form.value.options.push({ type: 'text', content: '', url: '' })
}
async function save() {
  if (!form.value.title.trim()) return showToast('请输入标题', 'error')
  if (form.value.options.length < 2) return showToast('至少需要2个选项', 'error')
  const data = {
    title: form.value.title, description: form.value.description,
    startTime: form.value.startTime ? form.value.startTime + ':00' : '',
    endTime: form.value.endTime ? form.value.endTime + ':00' : '',
    voteType: form.value.voteType, maxChoices: form.value.maxChoices,
    allowRepeat: form.value.allowRepeat, isPublic: form.value.isPublic,
    options: form.value.options.map((o, i) => ({ label: String.fromCharCode(65 + i), type: o.type, content: o.content, url: o.url }))
  }
  try {
    const url = editing.value ? `/admin/votes/${editing.value.id}` : '/admin/votes'
    const method = editing.value ? 'PUT' : 'POST'
    const res = await api(url, { method, body: JSON.stringify(data) })
    if (res.code === 200) { showToast(editing.value ? '更新成功' : '创建成功', 'success'); closeModal(); load() }
    else showToast(res.message || '操作失败', 'error')
  } catch (e) { showToast('操作失败', 'error') }
}

// ── 查看统计 ──
async function viewDetail(v) {
  try {
    const res = await api(`/admin/votes/${v.id}`)
    detailVote.value = res.data || { options: [] }
    detailShow.value = true
  } catch (e) { showToast('加载详情失败', 'error') }
}

// ── 导出单个投票 ──
async function exportSingle(v, format = 'xlsx') {
  try {
    const url = `/api/admin/votes/${v.id}/export?format=${format}`
    const resp = await fetch(url, { credentials: 'include' })
    if (!resp.ok) { showToast('导出失败', 'error'); return }
    const blob = await resp.blob()
    const safeName = (v.title || '投票').replace(/[/\\:*?"<>|]/g, '_')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `${safeName}_投票数据.${format}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(a.href)
    showToast('导出成功', 'success')
  } catch (e) { showToast('导出失败', 'error') }
}

// ── 删除 ──
async function handleDelete(v) {
  if (!confirm(`确定删除投票「${v.title}」？\n此操作不可撤销！`)) return
  try {
    const r = await api(`/admin/votes/${v.id}`, { method: 'DELETE' })
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
.stab.active { background: #fff; color: #6366f1; box-shadow: 0 1px 3px rgba(0,0,0,0.08); font-weight: 600; }
.stab-count { font-size: 11px; opacity: 0.7; }

/* ── 加载 / 空状态 ── */
.loader { text-align: center; padding: 48px 0; }
.loader-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #6366f1; margin: 0 4px; animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.2s; }
.loader-dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce { to { transform: translateY(-10px); opacity: 0.4; } }
.empty-state { text-align: center; padding: 48px 0; color: #94a3b8; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-state h3 { color: #64748b; margin-bottom: 6px; }
.empty-state p { font-size: 13px; }

/* ── 卡片网格 ── */
.vote-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px; }
.vote-card {
  background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px;
  transition: all 0.2s; display: flex; flex-direction: column;
}
.vote-card:hover { border-color: #c7d2fe; box-shadow: 0 4px 16px rgba(99,102,241,0.06); transform: translateY(-1px); }

.vc-header { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.status-dot.active { background: #22c55e; animation: pulse 2s infinite; }
.status-dot.upcoming { background: #f59e0b; }
.status-dot.ended { background: #94a3b8; }
@keyframes pulse { 0%, 100% { box-shadow: 0 0 0 0 rgba(34,197,94,0.4); } 50% { box-shadow: 0 0 0 6px rgba(34,197,94,0); } }
.vc-status-text { font-size: 12px; font-weight: 600; }
.vc-type { font-size: 12px; color: #94a3b8; margin-left: auto; }
.vc-public { font-size: 11px; padding: 2px 8px; background: #eef2ff; color: #6366f1; border-radius: 10px; }

.vc-title { font-size: 16px; font-weight: 700; margin-bottom: 6px; line-height: 1.4; word-break: break-word; }
.vc-desc { font-size: 13px; color: #94a3b8; margin-bottom: 12px; line-height: 1.5; word-break: break-word; }
.vc-meta { font-size: 12px; color: #94a3b8; margin-bottom: 6px; }
.vc-stats { margin-top: auto; padding-top: 12px; }
.vc-votes { font-size: 13px; font-weight: 500; color: #6366f1; }
.vc-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 14px; padding-top: 12px; border-top: 1px solid #f1f5f9; }
.vc-options-count { font-size: 12px; color: #94a3b8; }
.vc-actions { display: flex; gap: 4px; }

.btn-text {
  padding: 6px 12px; border: none; background: none; cursor: pointer;
  font-size: 13px; color: #6366f1; border-radius: 6px; transition: all 0.15s; font-weight: 500;
}
.btn-text:hover { background: #eef2ff; }
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
.input-sm { max-width: 100px; padding: 9px 12px !important; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-row-3 { grid-template-columns: 1.4fr 0.6fr; }

/* 单选组 */
.radio-group { display: flex; gap: 6px; }
.radio-label {
  flex: 1; padding: 10px 12px; text-align: center; border: 2px solid #e2e8f0; border-radius: 8px;
  cursor: pointer; font-size: 13px; font-weight: 500; transition: all 0.2s;
}
.radio-label input { display: none; }
.radio-label.active { border-color: #6366f1; background: #eef2ff; color: #6366f1; }

/* Switch 开关 */
.switch-row { display: flex; gap: 24px; flex-wrap: wrap; margin-top: 8px; }
.switch-item { display: flex; align-items: center; gap: 10px; cursor: pointer; }
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
.toggle input:checked + .toggle-track { background: #6366f1; }
.toggle input:checked + .toggle-track::after { transform: translateX(18px); }

/* 选项列表 */
.option-list { margin-bottom: 8px; }
.option-item { border: 1px solid #e2e8f0; border-radius: 10px; padding: 14px; margin-bottom: 10px; background: #fff; transition: border-color 0.2s; }
.option-item:hover { border-color: #c7d2fe; }
.opt-header { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.opt-label-badge {
  width: 28px; height: 28px; display: flex; align-items: center; justify-content: center;
  background: #eef2ff; color: #6366f1; font-weight: 700; font-size: 13px; border-radius: 6px; flex-shrink: 0;
}
.opt-type-select { padding: 6px 10px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; background: #f8fafc; }
.opt-input-area input[type="text"] { width: 100%; padding: 8px 12px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; }
.opt-input-area input[type="text"]:focus { outline: none; border-color: #a5b4fc; }
.media-input input { margin-bottom: 8px; }
.media-preview img, .media-preview video { max-width: 200px; max-height: 100px; border-radius: 8px; border: 1px solid #e2e8f0; }
.btn-icon { width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 14px; }
.btn-icon:hover { border-color: #ef4444; background: #fef2f2; }
.btn-icon:disabled { opacity: 0.3; cursor: not-allowed; }
.btn-dashed {
  width: 100%; padding: 12px; border: 2px dashed #e2e8f0; border-radius: 10px; background: #f8fafc;
  cursor: pointer; font-size: 13px; color: #64748b; transition: all 0.2s;
}
.btn-dashed:hover { border-color: #6366f1; color: #6366f1; background: #eef2ff; }

/* ── 弹窗底部 ── */
.modal-footer { display: flex; gap: 10px; justify-content: flex-end; padding-top: 16px; border-top: 1px solid #f1f5f9; margin-top: 4px; }

/* ── 详情弹窗 ── */
.detail-top { padding: 4px 0; }
.detail-summary { display: flex; gap: 10px; align-items: center; margin-bottom: 14px; flex-wrap: wrap; }
.status-tag { padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.status-tag.active { background: #f0fdf4; color: #16a34a; }
.status-tag.upcoming { background: #fffbeb; color: #d97706; }
.status-tag.ended { background: #f1f5f9; color: #64748b; }
.detail-total { font-size: 14px; font-weight: 500; color: #6366f1; }
.detail-vtype { font-size: 13px; color: #64748b; }

/* 摘要卡片 */
.summary-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 22px; }
.sum-card {
  display: flex; align-items: center; gap: 12px; padding: 16px; border-radius: 12px;
  background: linear-gradient(135deg, #f8fafc, #f1f5f9); border: 1px solid #e2e8f0;
}
.sum-icon { font-size: 28px; flex-shrink: 0; }
.sum-value { font-size: 18px; font-weight: 700; color: #1e293b; line-height: 1.3; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sum-label { font-size: 11px; color: #94a3b8; margin-top: 2px; }

.stat-grid { display: grid; gap: 12px; }
.stat-item { padding: 16px; background: #f8fafc; border-radius: 12px; border: 1px solid transparent; transition: all 0.25s; }
.stat-item.rank-1 { border-color: #fcd34d; background: linear-gradient(135deg, #fffbeb, #fef3c7); }
.stat-item.rank-2 { border-color: #cbd5e1; background: linear-gradient(135deg, #f8fafc, #f1f5f9); }
.stat-item.rank-3 { border-color: #fed7aa; background: linear-gradient(135deg, #fff7ed, #ffedd5); }
.stat-label { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; font-size: 14px; font-weight: 500; }
.stat-label-text { word-break: break-word; }

/* 排名徽章 */
.rank-badge {
  width: 32px; height: 32px; display: flex; align-items: center; justify-content: center;
  font-size: 14px; border-radius: 8px; flex-shrink: 0; font-weight: 700;
  background: #eef2ff; color: #6366f1;
}
.rank-badge.gold { background: #fef3c7; color: #b45309; font-size: 18px; }
.rank-badge.silver { background: #f1f5f9; color: #64748b; font-size: 18px; }
.rank-badge.bronze { background: #ffedd5; color: #c2410c; font-size: 18px; }

.stat-media img, .stat-media video { max-width: 200px; max-height: 100px; border-radius: 8px; margin-bottom: 10px; border: 1px solid #e2e8f0; }
.stat-bar-wrap {
  height: 30px; background: #e2e8f0; border-radius: 15px; overflow: hidden;
}
.stat-bar {
  height: 100%; border-radius: 15px;
  transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1); min-width: 4px;
  display: flex; align-items: center; padding-left: 14px;
  background: linear-gradient(90deg, #6366f1, #818cf8);
}
.stat-bar.bar-gold { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.stat-bar.bar-silver { background: linear-gradient(90deg, #64748b, #94a3b8); }
.stat-bar.bar-bronze { background: linear-gradient(90deg, #ea580c, #f97316); }
.stat-bar-text { font-size: 12px; color: #fff; font-weight: 700; text-shadow: 0 1px 2px rgba(0,0,0,0.2); }
.stat-bottom { display: flex; justify-content: space-between; margin-top: 10px; }
.stat-count { font-size: 14px; color: #1e293b; font-weight: 600; }
.stat-percent { font-size: 14px; color: #6366f1; font-weight: 700; }
.stat-item.rank-1 .stat-percent { color: #b45309; }
.stat-item.rank-2 .stat-percent { color: #64748b; }
.stat-item.rank-3 .stat-percent { color: #c2410c; }

/* ── 分页 ── */
.pagination { display: flex; justify-content: center; align-items: center; gap: 16px; margin-top: 28px; padding-top: 16px; border-top: 1px solid #f1f5f9; }
.pagination button { padding: 8px 18px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 13px; font-weight: 500; color: #475569; transition: all 0.2s; }
.pagination button:hover:not(:disabled) { border-color: #6366f1; color: #6366f1; }
.pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
.page-info { font-size: 13px; color: #94a3b8; }

/* ── 响应式 ── */
@media (max-width: 768px) {
  .toolbar { flex-direction: column; align-items: stretch; }
  .search-wrap { max-width: none; }
  .status-tabs { overflow-x: auto; -webkit-overflow-scrolling: touch; }
  .vote-grid { grid-template-columns: 1fr; }
  .form-row { grid-template-columns: 1fr; }
  .form-row-3 { grid-template-columns: 1fr; }
  .switch-row { flex-direction: column; gap: 12px; }
}
</style>
