<template>
  <div class="vote-page">
    <Topbar title="🗳️ 投票" :userName="userName" @logout="handleLogout" />
    <div class="main-container">
      <button class="back-btn" @click="$router.push('/student/votes')">← 返回投票列表</button>

      <div v-if="loading" class="loader"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>

      <template v-else-if="vote">
        <div class="vote-card">
          <h2>{{ vote.title }}</h2>
          <p class="desc" v-if="vote.description">{{ vote.description }}</p>
          <div class="meta-row">
            <span :class="['badge', vote.status === 'active' ? 'green' : 'orange']">{{ vote.status === 'active' ? '🟢 进行中' : '🟠 未开始' }}</span>
            <span>{{ vote.voteType === 'single' ? '单选' : '多选' + (vote.maxChoices > 1 ? `(最多${vote.maxChoices}项)` : '') }}</span>
            <span>已投 {{ vote.totalVotes }} 人</span>
          </div>

          <!-- 投票选项 -->
          <div class="options-list" v-if="!vote.hasVoted || vote.allowRepeat">
            <div
              v-for="opt in vote.options" :key="opt.id"
              :class="['option-item', { selected: isSelected(opt.id) }]"
              @click="toggleOption(opt.id)"
            >
              <div class="opt-check">{{ isSelected(opt.id) ? '✅' : '⬜' }}</div>
              <div class="opt-media" v-if="opt.type === 'image' && opt.url">
                <img :src="opt.url" />
              </div>
              <div class="opt-media" v-if="opt.type === 'video' && opt.url">
                <video :src="opt.url" controls muted></video>
              </div>
              <div class="opt-text" v-if="opt.content">{{ opt.content }}</div>
            </div>
          </div>

          <!-- 已投票结果 -->
          <div v-if="vote.hasVoted && !vote.allowRepeat" class="result-info">
            <div class="result-tip">✅ 您已投票</div>
          </div>

          <!-- 提交按钮 -->
          <div v-if="!vote.hasVoted || vote.allowRepeat">
            <button
              class="submit-btn"
              :disabled="selectedIds.length === 0 || submitting"
              @click="doSubmit"
            >{{ submitting ? '提交中...' : '🗳️ 提交投票' }}</button>
          </div>

          <!-- 结果展示（公开时或已投后） -->
          <div v-if="vote.isPublic || vote.hasVoted" class="results-section">
            <h3>📊 投票结果</h3>
            <div v-for="opt in vote.options" :key="opt.id" class="result-item">
              <div class="r-label">
                <span v-if="opt.content">{{ opt.content }}</span>
                <span v-if="opt.type === 'image' && opt.url">🖼️ 图片</span>
                <span v-if="opt.type === 'video' && opt.url">🎬 视频</span>
              </div>
              <div class="r-bar-wrap">
                <div class="r-bar" :style="{ width: Math.round(opt.percent || 0) + '%' }"></div>
              </div>
              <span class="r-stat">{{ opt.count }} 票 ({{ Math.round(opt.percent || 0) }}%)</span>
            </div>
          </div>
        </div>
      </template>
      <div v-else class="empty-state"><div class="empty-icon">🗳️</div><h3>投票不存在</h3></div>
    </div>
    <Toast />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, checkAuth, logout, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'
import Topbar from '@/components/Topbar.vue'
import Toast from '@/components/Toast.vue'

const route = useRoute()
const router = useRouter()
const userName = ref('')
const vote = ref(null)
const selectedIds = ref([])
const loading = ref(true)
const submitting = ref(false)

function isSelected(id) {
  if (vote.value?.voteType === 'single') return selectedIds.value[0] === id
  return selectedIds.value.includes(id)
}

function toggleOption(id) {
  if (!vote.value) return
  if (vote.value.voteType === 'single') {
    selectedIds.value = selectedIds.value[0] === id ? [] : [id]
  } else {
    const idx = selectedIds.value.indexOf(id)
    if (idx > -1) selectedIds.value.splice(idx, 1)
    else if (selectedIds.value.length < (vote.value.maxChoices || 10)) selectedIds.value.push(id)
  }
}

async function loadVote() {
  loading.value = true
  try {
    const res = await api(`/student/votes/${route.query.id}`)
    vote.value = res.data
  } catch (e) { showToast('加载投票失败', 'error') } finally { loading.value = false }
}

async function doSubmit() {
  if (selectedIds.value.length === 0) return showToast('请选择投票选项', 'error')
  submitting.value = true
  try {
    const res = await api('/student/votes/submit', {
      method: 'POST', body: JSON.stringify({ voteId: Number(route.query.id), optionIds: selectedIds.value })
    })
    if (res.code === 200) { showToast('投票成功', 'success'); selectedIds.value = []; await loadVote() }
    else showToast(res.message || '投票失败', 'error')
  } catch (e) { showToast('投票失败', 'error') } finally { submitting.value = false }
}

async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const u = currentUser.value
  if (u?.isAdmin) { showToast('管理员不能参与投票', 'error'); setTimeout(() => router.push('/admin'), 1000); return }
  userName.value = u?.name || u?.workNo || ''
  await loadVote()
})
</script>

<style scoped>
.vote-page { min-height: 100vh; background: linear-gradient(160deg, #ede9fe 0%, #f3f0ff 50%, #ede9fe 100%); }
.main-container { max-width: 680px; margin: 30px auto; padding: 0 20px; }
.back-btn { display: inline-flex; align-items: center; gap: 5px; background: none; border: none; color: #667eea; font-size: 14px; font-weight: 600; cursor: pointer; margin-bottom: 16px; }
.vote-card { background: white; border-radius: 16px; padding: 32px; box-shadow: 0 2px 20px rgba(0,0,0,0.06); }
.vote-card h2 { font-size: 22px; color: #1a1a2e; margin-bottom: 10px; }
.desc { color: #666; font-size: 14px; margin-bottom: 16px; line-height: 1.6; }
.meta-row { display: flex; gap: 14px; flex-wrap: wrap; margin-bottom: 20px; font-size: 13px; }
.badge { padding: 3px 12px; border-radius: 12px; font-weight: 600; font-size: 12px; }
.badge.green { background: #f0fdf4; color: #16a34a; }
.badge.orange { background: #fffbeb; color: #d97706; }
.options-list { display: grid; gap: 12px; margin-bottom: 20px; }
.option-item {
  display: flex; align-items: center; gap: 12px; padding: 14px 18px;
  border: 2px solid #e5e7eb; border-radius: 12px; cursor: pointer;
  transition: all 0.3s; background: #fafbfc;
}
.option-item:hover { border-color: #a78bfa; background: #f8f7ff; }
.option-item.selected { border-color: #7c3aed; background: #f5f3ff; }
.opt-check { font-size: 20px; flex-shrink: 0; }
.opt-media img, .opt-media video { max-width: 160px; max-height: 100px; border-radius: 8px; }
.opt-text { flex: 1; font-size: 15px; color: #333; font-weight: 500; }
.submit-btn {
  width: 100%; padding: 15px; background: linear-gradient(135deg, #7c3aed, #8b5cf6);
  color: white; border: none; border-radius: 12px; font-size: 16px; font-weight: 700;
  cursor: pointer; transition: all 0.3s; box-shadow: 0 4px 16px rgba(124,58,237,0.35);
}
.submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.submit-btn:hover:not(:disabled) { transform: translateY(-2px); }
.result-info { padding: 14px; background: #f0fdf4; border-radius: 10px; text-align: center; margin: 12px 0; }
.result-tip { color: #16a34a; font-weight: 600; }
.results-section { margin-top: 28px; padding-top: 20px; border-top: 2px solid #f0f0f0; }
.results-section h3 { font-size: 16px; margin-bottom: 14px; color: #333; }
.result-item { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.r-label { width: 120px; font-size: 13px; color: #555; flex-shrink: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.r-bar-wrap { flex: 1; height: 20px; background: #e5e7eb; border-radius: 10px; overflow: hidden; }
.r-bar { height: 100%; background: linear-gradient(90deg, #7c3aed, #a78bfa); border-radius: 10px; transition: width 0.6s; min-width: 2px; }
.r-stat { font-size: 12px; color: #888; width: 80px; flex-shrink: 0; text-align: right; }

@media (max-width: 768px) {
  .main-container { margin: 16px auto; padding: 0 12px; }
  .vote-card { padding: 20px 16px; border-radius: 12px; }
  .option-item { padding: 12px; }
}
</style>
