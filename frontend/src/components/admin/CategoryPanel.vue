<template>
  <div>
    <div class="cat-header">
      <div class="cat-header-left">
        <div class="cat-icon">📂</div>
        <div>
          <h2 class="cat-title">考试分类</h2>
          <p class="cat-subtitle">管理考试分类，便于题目和考试的归类组织</p>
        </div>
      </div>
      <button class="btn btn-primary btn-add" @click="openCreate">
        <span class="btn-plus">＋</span> 添加分类
      </button>
    </div>

    <!-- 统计卡片 -->
    <div class="cat-stats">
      <div class="stat-card">
        <div class="stat-num">{{ categories.length }}</div>
        <div class="stat-label">分类总数</div>
      </div>
      <div class="stat-card">
        <div class="stat-num">{{ filtered.length }}</div>
        <div class="stat-label">当前显示</div>
      </div>
    </div>

    <!-- 搜索 -->
    <div class="cat-search">
      <span class="search-icon">🔍</span>
      <input type="text" placeholder="搜索分类名称..." v-model="search" />
      <button v-if="search" class="clear-btn" @click="search=''">✕</button>
    </div>

    <!-- 空状态 -->
    <div v-if="filtered.length === 0 && !search" class="empty-state-enhanced">
      <div class="empty-visual">
        <div class="empty-icon-big">📂</div>
        <div class="empty-circle c1"></div>
        <div class="empty-circle c2"></div>
        <div class="empty-circle c3"></div>
      </div>
      <h3>还没有分类</h3>
      <p>点击右上角「添加分类」创建第一个考试分类</p>
      <button class="btn btn-primary" @click="openCreate">＋ 创建分类</button>
    </div>

    <div v-else-if="filtered.length === 0 && search" class="empty-state">
      <div class="empty-icon">🔍</div>
      <h3>未找到匹配的分类</h3>
      <p>尝试其他关键词搜索</p>
    </div>

    <!-- 分类卡片 -->
    <div v-else class="cat-grid">
      <div v-for="c in filtered" :key="c.id" class="cat-card">
        <div class="cat-card-top">
          <div class="cat-avatar" :style="{ background: avatarColor(c.name) }">
            {{ c.name.charAt(0).toUpperCase() }}
          </div>
          <div class="cat-card-body">
            <h3 class="cat-card-title">{{ c.name }}</h3>
            <p class="cat-card-desc" v-if="c.description">{{ c.description }}</p>
            <p class="cat-card-desc empty-desc" v-else>暂无描述</p>
          </div>
        </div>
        <div class="cat-card-actions">
          <button class="cat-btn edit-btn" @click="openEdit(c)">
            <span>✎</span> 编辑
          </button>
          <button class="cat-btn delete-btn" @click="handleDelete(c)">
            <span>🗑</span> 删除
          </button>
        </div>
      </div>
    </div>

    <!-- 弹窗 -->
    <Modal :show="modalShow" :title="editing ? '编辑分类' : '添加分类'" :maxWidth="'500px'" @close="closeModal">
      <div class="modal-body">
        <div class="form-group-enhanced">
          <label class="form-label-enhanced">
            <span class="label-icon">🏷️</span> 分类名称
          </label>
          <input
            type="text"
            v-model="form.name"
            placeholder="例如：前端开发、Java基础"
            class="form-input-enhanced"
            ref="nameInputRef"
          />
          <p class="form-hint">建议使用简洁明了的名称，方便后续分类筛选</p>
        </div>
        <div class="form-group-enhanced">
          <label class="form-label-enhanced">
            <span class="label-icon">📝</span> 分类描述
          </label>
          <textarea
            v-model="form.description"
            placeholder="简要描述该分类的考试内容（选填）"
            class="form-textarea-enhanced"
            rows="3"
          ></textarea>
          <p class="form-hint">{{ form.description.length || 0 }}/200 字</p>
        </div>
      </div>
      <div class="modal-footer-enhanced">
        <button class="modal-btn-cancel" @click="closeModal">取消</button>
        <button class="modal-btn-save" @click="save">
          {{ editing ? '保存修改' : '立即创建' }}
        </button>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { api, showToast } from './shared'

const emit = defineEmits(['changed'])
const categories = ref([])
const search = ref('')
const modalShow = ref(false)
const editing = ref(null)
const form = ref({ name: '', description: '' })
const nameInputRef = ref(null)

const filtered = computed(() => {
  if (!search.value) return categories.value
  const kw = search.value.toLowerCase()
  return categories.value.filter(c => c.name.toLowerCase().includes(kw))
})

const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#ef4444', '#14b8a6', '#f97316', '#06b6d4']
function avatarColor(name) {
  let hash = 0
  for (let i = 0; i < (name || '').length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

async function load() {
  const res = await api('/admin/categories')
  categories.value = res.data || []
}

async function openCreate() {
  editing.value = null
  form.value = { name: '', description: '' }
  modalShow.value = true
  await nextTick()
  nameInputRef.value?.focus()
}
function openEdit(c) {
  editing.value = c
  form.value = { name: c.name, description: c.description || '' }
  modalShow.value = true
}
function closeModal() { modalShow.value = false }

async function save() {
  if (!form.value.name.trim()) return showToast('请输入分类名称', 'error')
  if (editing.value) {
    const res = await api(`/admin/categories/${editing.value.id}`, { method: 'PUT', body: JSON.stringify(form.value) })
    if (res.code === 200) { showToast('更新成功', 'success'); closeModal(); load(); emit('changed') }
    else showToast(res.message || '更新失败', 'error')
  } else {
    const res = await api('/admin/categories', { method: 'POST', body: JSON.stringify(form.value) })
    if (res.code === 200) { showToast('创建成功', 'success'); closeModal(); load(); emit('changed') }
    else showToast(res.message || '创建失败', 'error')
  }
}

async function handleDelete(c) {
  if (!confirm(`确定删除分类「${c.name}」吗？\n\n删除后该分类下的考试将不受影响，但建议提前处理。`)) return
  const res = await api(`/admin/categories/${c.id}`, { method: 'DELETE' })
  if (res.code === 200) { showToast('删除成功', 'success'); load(); emit('changed') }
  else showToast(res.message || '删除失败', 'error')
}

onMounted(load)
defineExpose({ load })
</script>

<style scoped>
/* ===== 头部 ===== */
.cat-header {
  display: flex; justify-content: space-between; align-items: center;
  gap: 16px; margin-bottom: 24px; flex-wrap: wrap;
}
.cat-header-left { display: flex; align-items: center; gap: 14px; }
.cat-icon {
  width: 48px; height: 48px; border-radius: 14px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  display: flex; align-items: center; justify-content: center;
  font-size: 24px; box-shadow: 0 6px 20px rgba(99,102,241,0.25);
}
.cat-title { font-size: 20px; font-weight: 700; color: var(--text); margin: 0; }
.cat-subtitle { font-size: 13px; color: var(--text-secondary); margin: 2px 0 0 0; }
.btn-add { padding: 10px 22px; font-size: 14px; border-radius: 10px; }
.btn-plus { font-size: 16px; font-weight: 700; }

/* ===== 统计卡片 ===== */
.cat-stats { display: flex; gap: 14px; margin-bottom: 20px; }
.stat-card {
  flex: 1; min-width: 100px; background: var(--white);
  border: 1px solid var(--border); border-radius: 12px;
  padding: 16px 20px; text-align: center; transition: all var(--transition);
}
.stat-card:hover { box-shadow: var(--shadow); transform: translateY(-1px); }
.stat-num { font-size: 28px; font-weight: 800; color: var(--primary); line-height: 1.2; }
.stat-label { font-size: 12px; color: var(--text-secondary); margin-top: 2px; font-weight: 500; }

/* ===== 搜索 ===== */
.cat-search { position: relative; max-width: 400px; margin-bottom: 20px; }
.cat-search .search-icon { position: absolute; left: 14px; top: 50%; transform: translateY(-50%); font-size: 15px; }
.cat-search input {
  width: 100%; padding: 11px 40px 11px 42px; border: 2px solid var(--border);
  border-radius: 10px; font-size: 14px; transition: all var(--transition); background: var(--white);
}
.cat-search input:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 0 3px rgba(99,102,241,0.1); }
.cat-search .clear-btn {
  position: absolute; right: 10px; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer; color: #999; font-size: 16px; padding: 4px 6px; border-radius: 4px;
}
.cat-search .clear-btn:hover { color: #333; background: #f1f5f9; }

/* ===== 增强空状态 ===== */
.empty-state-enhanced { text-align: center; padding: 50px 20px; }
.empty-visual { position: relative; display: inline-block; margin-bottom: 20px; }
.empty-icon-big {
  font-size: 56px; position: relative; z-index: 2;
  animation: float 3s ease-in-out infinite;
}
.empty-circle {
  position: absolute; border-radius: 50%; z-index: 1;
  background: rgba(99,102,241,0.08);
}
.empty-circle.c1 { width: 70px; height: 70px; top: -8px; left: -8px; }
.empty-circle.c2 { width: 50px; height: 50px; bottom: 2px; right: -12px; background: rgba(139,92,246,0.08); }
.empty-circle.c3 { width: 36px; height: 36px; top: 10px; right: -20px; background: rgba(236,72,153,0.06); }
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}
.empty-state-enhanced h3 { font-size: 18px; color: var(--text); margin: 0 0 8px 0; }
.empty-state-enhanced p { color: var(--text-secondary); font-size: 14px; margin: 0 0 20px 0; }

/* ===== 分类卡片网格 ===== */
.cat-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 14px;
}
.cat-card {
  background: var(--white); border: 1px solid var(--border);
  border-radius: 14px; padding: 18px 20px;
  transition: all var(--transition);
  display: flex; flex-direction: column; gap: 14px;
}
.cat-card:hover {
  box-shadow: var(--shadow); transform: translateY(-2px);
  border-color: var(--primary-light);
}
.cat-card-top { display: flex; align-items: flex-start; gap: 14px; }
.cat-avatar {
  width: 44px; height: 44px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 18px; font-weight: 800; flex-shrink: 0;
}
.cat-card-body { flex: 1; min-width: 0; }
.cat-card-title {
  font-size: 16px; font-weight: 700; color: var(--text);
  margin: 0 0 4px 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.cat-card-desc {
  font-size: 13px; color: var(--text-secondary); margin: 0;
  line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
  overflow: hidden;
}
.cat-card-desc.empty-desc { color: #cbd5e1; font-style: italic; }

.cat-card-actions {
  display: flex; gap: 8px; padding-top: 12px;
  border-top: 1px solid #f1f5f9;
}
.cat-btn {
  flex: 1; padding: 8px 0; border: none; border-radius: 8px;
  font-size: 13px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 5px;
  transition: all var(--transition);
}
.cat-btn span { font-size: 14px; }
.edit-btn { background: #eef2ff; color: var(--primary); }
.edit-btn:hover { background: #e0e7ff; }
.delete-btn { background: #fef2f2; color: #ef4444; }
.delete-btn:hover { background: #fee2e2; }

/* ===== 弹窗增强 ===== */
.modal-body { padding: 4px 0; }
.form-group-enhanced { margin-bottom: 22px; }
.form-label-enhanced {
  display: flex; align-items: center; gap: 8px;
  font-size: 14px; font-weight: 700; color: var(--text); margin-bottom: 8px;
}
.label-icon { font-size: 16px; }
.form-input-enhanced {
  width: 100%; padding: 12px 16px; border: 2px solid var(--border);
  border-radius: 10px; font-size: 15px; transition: all var(--transition);
  background: var(--white); font-family: inherit;
}
.form-input-enhanced:focus {
  outline: none; border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}
.form-textarea-enhanced {
  width: 100%; padding: 12px 16px; border: 2px solid var(--border);
  border-radius: 10px; font-size: 15px; transition: all var(--transition);
  background: var(--white); font-family: inherit; resize: vertical;
}
.form-textarea-enhanced:focus {
  outline: none; border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}
.form-hint {
  font-size: 12px; color: #94a3b8; margin: 6px 0 0 0;
  text-align: right;
}

.modal-footer-enhanced {
  display: flex; gap: 10px; justify-content: flex-end; margin-top: 28px;
  padding-top: 16px; border-top: 1px solid #f1f5f9;
}
.modal-btn-cancel {
  padding: 10px 24px; border: 2px solid var(--border); border-radius: 10px;
  background: var(--white); color: var(--text-secondary);
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all var(--transition);
}
.modal-btn-cancel:hover { background: #f8fafc; border-color: #cbd5e1; }
.modal-btn-save {
  padding: 10px 28px; border: none; border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white;
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all var(--transition);
  box-shadow: 0 4px 14px rgba(99,102,241,0.3);
}
.modal-btn-save:hover { box-shadow: 0 6px 20px rgba(99,102,241,0.45); transform: translateY(-1px); }

/* 响应式 */
@media (max-width: 768px) {
  .cat-grid { grid-template-columns: 1fr; }
  .cat-stats { flex-direction: column; }
  .cat-header { flex-direction: column; align-items: flex-start; }
}
</style>
