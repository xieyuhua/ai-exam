<template>
  <div class="list-page">
    <Topbar title="🗳️ 投票列表" :userName="userName" @logout="handleLogout" />
    <div class="main-container">
      <button class="back-btn" @click="$router.push('/student')">← 返回首页</button>

      <div class="content">
        <div class="loader" v-if="loading"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>
        <div v-else-if="voteList.length === 0" class="empty-state">
          <div class="empty-icon">🗳️</div>
          <h3>暂无投票</h3>
          <p>管理员尚未创建投票活动</p>
        </div>
        <div v-else>
          <div v-for="vote in voteList" :key="vote.id" class="card">
            <div class="card-left">
              <div class="title-row">
                <h3>{{ vote.title }}</h3>
                <span :class="['badge', voteStatusCls(vote)]">{{ voteStatusText(vote) }}</span>
                <span class="public-badge" v-if="vote.isPublic" title="结果公开可见">👁 公开</span>
                <span class="public-badge private" v-else title="仅管理员可见结果">🔒 未公开</span>
              </div>
              <div class="card-meta" v-if="vote.description">
                <span>{{ vote.description }}</span>
              </div>
              <div class="card-info">
                <span>{{ vote.voteType === 'single' ? '单选' : '多选' }}</span>
                <span v-if="vote.totalVotes > 0">👥 {{ vote.totalVotes }} 人已投票</span>
                <span>📅 {{ vote.startTime }}</span>
              </div>
            </div>
            <div class="card-right">
              <button class="btn btn-primary" @click="$router.push(`/vote?id=${vote.id}`)">参与投票 →</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, checkAuth, logout, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'
import Topbar from '@/components/Topbar.vue'

const router = useRouter()
const userName = ref('')
const voteList = ref([])
const loading = ref(false)

function voteStatusCls(vote) {
  if (vote.hasVoted) return 'green'
  if (vote.status === 'ended') return 'gray'
  if (vote.status === 'upcoming') return 'orange'
  return 'green'
}
function voteStatusText(vote) {
  if (vote.hasVoted) return '✓ 已投票'
  if (vote.status === 'ended') return '已结束'
  if (vote.status === 'upcoming') return '未开始'
  return '进行中'
}

async function loadVotes() {
  loading.value = true
  try { const res = await api('/student/votes'); voteList.value = res.data || [] } catch (e) { console.error(e) } finally { loading.value = false }
}
async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const u = currentUser.value
  if (u?.isAdmin) { showToast('管理员不能参与投票', 'error'); setTimeout(() => router.push('/admin'), 1000); return }
  userName.value = u?.name || u?.workNo || ''
  await loadVotes()
})
</script>

<style scoped>
.list-page { min-height: 100vh; background: linear-gradient(160deg, #ede9fe 0%, #f3f0ff 50%, #ede9fe 100%); }
.main-container { max-width: 800px; margin: 30px auto; padding: 0 20px; }
.back-btn { display: inline-flex; align-items: center; gap: 5px; background: none; border: none; color: #7c3aed; font-size: 14px; font-weight: 600; cursor: pointer; margin-bottom: 16px; }
.content { background: white; border-radius: 16px; padding: 28px 30px; min-height: 420px; box-shadow: 0 2px 16px rgba(0,0,0,0.06); }
.card { display: flex; align-items: center; justify-content: space-between; border: 1px solid #edf0f5; border-radius: 12px; padding: 20px 24px; margin-bottom: 12px; transition: all 0.2s; }
.card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); border-color: #c4b5fd; }
.card-left { flex: 1; min-width: 0; }
.card-right { flex-shrink: 0; margin-left: 20px; }
.title-row { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; flex-wrap: wrap; }
.title-row h3 { font-size: 17px; color: #1a1a2e; font-weight: 600; }
.card-meta { font-size: 13px; color: #888; margin-bottom: 8px; }
.card-info { display: flex; gap: 14px; font-size: 13px; color: #999; flex-wrap: wrap; }
.badge { padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.badge.green { background: #f0fdf4; color: #16a34a; }
.badge.orange { background: #fffbeb; color: #d97706; }
.badge.gray { background: #f1f5f9; color: #64748b; }
.public-badge { padding: 2px 8px; border-radius: 10px; font-size: 11px; background: #eef2ff; color: #6366f1; }
.public-badge.private { background: #fef3c7; color: #d97706; }
.loader { text-align: center; padding: 40px 0; }
.loader-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #7c3aed; margin: 0 4px; animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.2s; }
.loader-dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce { to { transform: translateY(-10px); opacity: 0.4; } }
.empty-state { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-state h3 { font-size: 18px; color: #333; margin-bottom: 8px; }
.empty-state p { font-size: 14px; color: #999; }
.btn { padding: 10px 22px; border: none; border-radius: 10px; font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.3s; }
.btn-primary { background: linear-gradient(135deg, #7c3aed, #8b5cf6); color: white; }
.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(124,58,237,0.3); }
@media (max-width: 768px) { .main-container { margin: 16px auto; padding: 0 12px; } .content { padding: 18px 14px; } .card { flex-direction: column; align-items: stretch; gap: 14px; } .card-right { margin-left: 0; } }
</style>
