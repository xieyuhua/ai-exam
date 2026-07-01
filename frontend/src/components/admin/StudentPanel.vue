<template>
  <div>
    <div class="section-title">
      👤 员工管理
      <div style="display:flex;gap:8px;">
        <button class="btn btn-primary" @click="openCreate">＋ 添加员工</button>
        <button class="btn btn-secondary" @click="importShow = true">📥 批量导入</button>
      </div>
    </div>
    <div class="search-bar">
      <input type="text" placeholder="搜索账号 / 姓名..." v-model="search" />
    </div>

    <div v-if="filtered.length === 0" class="empty-state"><div class="empty-icon">👤</div><h3>暂无员工</h3></div>

    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr><th>账号</th><th>姓名</th><th>来源</th><th>密码登录</th><th>创建时间</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="s in filtered" :key="s.id">
            <td>{{ s.workNo }}</td>
            <td><strong>{{ s.name }}</strong></td>
            <td>
              <span :class="['source-tag', s.source]">{{ sourceLabel(s.source) }}</span>
            </td>
            <td>
              <span v-if="s.hasPassword" class="pwd-enabled">✅ 已设置</span>
              <span v-else class="pwd-disabled">❌ 未设置</span>
            </td>
            <td>{{ s.createdAt ? s.createdAt.substring(0, 10) : '--' }}</td>
            <td>
              <button class="btn btn-secondary btn-sm" @click="openEdit(s)">编辑</button>
              <button class="btn btn-danger btn-sm" @click="handleDelete(s)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <button :disabled="page <= 1" @click="page--; load()">上一页</button>
      <span class="page-info">第 {{ page }}/{{ totalPages }} 页（共 {{ total }} 条）</span>
      <button :disabled="page >= totalPages" @click="page++; load()">下一页</button>
    </div>

    <!-- 添加/编辑弹窗 -->
    <Modal :show="modalShow" :title="editing ? '编辑员工' : '添加员工'" @close="closeModal">
      <div class="form-group"><label>账号 <span class="required">*</span></label><input type="text" v-model="form.workNo" placeholder="请输入账号" :disabled="!!editing" /></div>
      <div class="form-group"><label>姓名 <span class="required">*</span></label><input type="text" v-model="form.name" placeholder="请输入姓名" /></div>
      <div class="form-group">
        <label>密码 <span class="hint">{{ editing ? '(留空不修改)' : '(留空则默认 123456)' }}</span></label>
        <input type="text" v-model="form.password" placeholder="请输入密码" />
      </div>
      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="closeModal">取消</button>
        <button class="modal-btn confirm" @click="save">保存</button>
      </div>
    </Modal>

    <!-- 批量导入弹窗 -->
    <Modal :show="importShow" title="📥 批量导入员工" @close="importShow = false">
      <div class="import-area" @click="$refs.studentFile?.click()">
        <div style="font-size:44px;">📄</div>
        <p>点击选择 Excel 文件（.xlsx）</p>
        <button class="btn btn-secondary">选择文件</button>
        <input type="file" ref="studentFile" accept=".xlsx,.xls" @change="handleFile" style="display:none;" />
      </div>
      <div v-if="importFileInfo" class="import-preview">
        <div class="file-info"><strong>{{ importFileInfo.name }}</strong></div>
      </div>
      <div class="import-hint">
        💡 Excel 表头需包含：<b>账号</b>、<b>姓名</b>、<b>密码</b>（密码列可选，不填默认 123456）
      </div>
      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="importShow = false">取消</button>
        <button class="modal-btn confirm" :disabled="!importFileInfo" @click="doImport">开始导入</button>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, showToast } from './shared'
import Modal from '@/components/Modal.vue'

const students = ref([])
const search = ref('')
const page = ref(1)
const total = ref(0)
const pageSize = 10
const modalShow = ref(false)
const editing = ref(null)
const importShow = ref(false)
const importFileInfo = ref(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const filtered = computed(() => students.value.filter(s => {
  if (!search.value) return true
  const kw = search.value.toLowerCase()
  return s.workNo?.toLowerCase().includes(kw) || s.name?.toLowerCase().includes(kw)
}))

const form = ref({ workNo: '', name: '', password: '' })

function sourceLabel(source) {
  const map = { import: '📥 管理员导入', wxwork: '💬 企业微信' }
  return map[source] || source || '--'
}

async function load() {
  const params = new URLSearchParams({ page: page.value, pageSize })
  if (search.value) params.set('keyword', search.value)
  const res = await api(`/admin/students?${params.toString()}`)
  students.value = res.data?.list || res.data || []
  total.value = res.data?.total || 0
}

function openCreate() { editing.value = null; form.value = { workNo: '', name: '', password: '' }; modalShow.value = true }
function openEdit(s) { editing.value = s; form.value = { workNo: s.workNo, name: s.name, password: '' }; modalShow.value = true }
function closeModal() { modalShow.value = false }

async function save() {
  if (!form.value.workNo.trim()) return showToast('请输入账号', 'error')
  if (!form.value.name.trim()) return showToast('请输入姓名', 'error')

  const data = { workNo: form.value.workNo.trim(), name: form.value.name.trim() }
  if (form.value.password.trim()) data.password = form.value.password.trim()
  // 新建时如果没有填密码，给默认值
  if (!editing.value && !data.password) data.password = '123456'

  if (editing.value) {
    const res = await api(`/admin/students/${editing.value.id}`, { method: 'PUT', body: JSON.stringify(data) })
    if (res.code === 200) { showToast('更新成功', 'success'); closeModal(); load() }
    else showToast(res.message || '更新失败', 'error')
  } else {
    const res = await api('/admin/students', { method: 'POST', body: JSON.stringify(data) })
    if (res.code === 200) { showToast('创建成功', 'success'); closeModal(); load() }
    else showToast(res.message || '创建失败', 'error')
  }
}

async function handleDelete(s) {
  if (!confirm(`确定删除员工「${s.name}」吗？`)) return
  const res = await api(`/admin/students/${s.id}`, { method: 'DELETE' })
  if (res.code === 200) { showToast('删除成功', 'success'); load() }
  else showToast(res.message || '删除失败', 'error')
}

function handleFile(e) { importFileInfo.value = e.target.files[0] }

async function doImport() {
  if (!importFileInfo.value) return
  try {
    const formData = new FormData()
    formData.append('file', importFileInfo.value)
    const token = document.cookie.match(/(?:^|;\s*)exam_token=([^;]*)/)?.[1] || localStorage.getItem('exam_token')
    const res = await fetch('/api/admin/students/import', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    })
    const data = await res.json()
    if (data.code === 200) {
      showToast(data.message || `导入成功：新增 ${data.data?.added || 0} 人`, 'success')
      importShow.value = false; importFileInfo.value = null; load()
    } else {
      showToast(data.message || '导入失败', 'error')
    }
  } catch (e) {
    showToast('导入失败：' + e.message, 'error')
  }
}

onMounted(load)
defineExpose({ load })
</script>

<style scoped>
.section-title {
  font-size: 18px; font-weight: 700; color: var(--text);
  margin-bottom: 20px; padding-bottom: 14px;
  border-bottom: 2px solid #f1f5f9;
  display: flex; justify-content: space-between; align-items: center; gap: 10px;
}
.table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid var(--border); }

/* 来源标签 */
.source-tag {
  display: inline-block; padding: 2px 10px; border-radius: 20px;
  font-size: 12px; font-weight: 600; white-space: nowrap;
}
.source-tag.import { background: #e8f4fd; color: #0f4c81; }
.source-tag.wxwork { background: #e6f9f0; color: #07c160; }

/* 密码状态 */
.pwd-enabled { color: #16a34a; font-size: 12px; }
.pwd-disabled { color: #ccc; font-size: 12px; }

/* 表单 */
.form-group { margin-bottom: 14px; }
.form-group label { display: block; font-size: 13px; font-weight: 600; color: #444; margin-bottom: 6px; }
.form-group .required { color: #ef4444; }
.form-group .hint { font-weight: 400; color: #999; font-size: 11px; }
.form-group input {
  width: 100%; padding: 10px 14px; border: 2px solid #e0e0e0;
  border-radius: 10px; font-size: 14px; transition: all 0.3s; outline: none;
}
.form-group input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.08); }
.form-group input:disabled { background: #f5f5f5; color: #999; }

.import-area {
  border: 2px dashed var(--border); border-radius: var(--radius);
  padding: 36px; text-align: center; cursor: pointer; margin-bottom: 14px;
  transition: all var(--transition);
}
.import-area:hover { border-color: var(--primary); background: #f8faff; }
.import-preview { padding: 12px 16px; background: #f0fdf4; border-radius: var(--radius); border: 1px solid #bbf7d0; margin-bottom: 12px; }
.import-hint { color: #888; font-size: 12px; line-height: 1.6; }
.import-hint b { color: #6366f1; }
.modal-buttons { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }
.modal-btn {
  padding: 10px 24px; border: none; border-radius: var(--radius-sm);
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all var(--transition);
}
.modal-btn.cancel { background: #f1f5f9; color: var(--text-secondary); }
.modal-btn.cancel:hover { background: #e2e8f0; }
.modal-btn.confirm { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; }
.modal-btn.confirm:hover { box-shadow: 0 4px 16px rgba(99,102,241,0.35); }
</style>
