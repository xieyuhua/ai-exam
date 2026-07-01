<template>
  <div>
    <div class="section-title">
      📊 成绩管理
    </div>
    <div class="search-bar">
      <input type="text" placeholder="搜索账号 / 考试名称..." v-model="search" />
      <select v-model="categoryFilter">
        <option value="">全部分类</option>
        <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
    </div>

    <div v-if="scores.length === 0" class="empty-state"><div class="empty-icon">📊</div><h3>暂无成绩数据</h3></div>

    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr><th>账号</th><th>姓名</th><th>考试名称</th><th>分类</th><th>成绩</th><th>正确率</th><th>考试时间</th></tr>
        </thead>
        <tbody>
          <tr v-for="s in filtered" :key="s.id">
            <td>{{ s.workNo || '--' }}</td>
            <td><strong>{{ s.studentName || '--' }}</strong></td>
            <td>{{ s.examTitle }}</td>
            <td>{{ s.categoryName || '--' }}</td>
            <td><span :class="['badge', scoreClass(s.score)]">{{ Number(s.score || 0).toFixed(1) }}分</span></td>
            <td>
              <span :class="rateClass(s)">
                {{ s.correctCount || 0 }}/{{ s.totalCount || 0 }} ({{ rate(s) }}%)
              </span>
            </td>
            <td>{{ s.createdAt ? s.createdAt.substring(0, 10) : '--' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <button :disabled="page <= 1" @click="page--; load()">上一页</button>
      <span class="page-info">第 {{ page }}/{{ totalPages }} 页（共 {{ total }} 条）</span>
      <button :disabled="page >= totalPages" @click="page++; load()">下一页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, showToast } from './shared'

const props = defineProps({ categories: { type: Array, default: () => [] } })

const scores = ref([])
const search = ref('')
const categoryFilter = ref('')
const page = ref(1)
const total = ref(0)
const pageSize = 10

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const filtered = computed(() => scores.value.filter(s => {
  if (search.value) {
    const kw = search.value.toLowerCase()
    if (!s.studentName?.toLowerCase().includes(kw) &&
        !s.workNo?.toLowerCase().includes(kw) &&
        !s.examTitle?.toLowerCase().includes(kw)) return false
  }
  if (categoryFilter.value && String(s.categoryId) !== String(categoryFilter.value)) return false
  return true
}))

function scoreClass(score) {
  const s = Number(score) || 0
  if (s >= 80) return 'green'
  if (s < 60) return 'red'
  return 'orange'
}

function rate(r) {
  return Math.round((r.correctCount || 0) / (r.totalCount || 1) * 100)
}

function rateClass(r) {
  const rt = rate(r)
  if (rt >= 80) return 'rate-high'
  if (rt < 60) return 'rate-low'
  return 'rate-mid'
}

async function load() {
  const params = new URLSearchParams({ page: page.value, pageSize })
  if (categoryFilter.value) params.set('categoryId', categoryFilter.value)
  if (search.value) params.set('keyword', search.value)
  const res = await api(`/admin/scores?${params.toString()}`)
  scores.value = res.data?.list || res.data || []
  total.value = res.data?.total || 0
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
.badge.green { background: #f0fdf4; color: #16a34a; }
.badge.orange { background: #fffbeb; color: #d97706; }
.badge.red { background: #fef2f2; color: #dc2626; }
.rate-high { color: #16a34a; font-weight: 600; }
.rate-mid { color: #d97706; font-weight: 600; }
.rate-low { color: #dc2626; font-weight: 600; }
</style>
