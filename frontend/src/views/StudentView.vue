<template>
  <div class="home-page">
    <Topbar title="海斛集团" :userName="userName" @logout="handleLogout" />
    <div class="main-container">
      <div class="welcome-section">
        <h1>欢迎回来，{{ userName }}</h1>
        <p>选择您要使用的功能模块</p>
      </div>

      <div class="module-grid">
        <!-- 考试中心 -->
        <div class="module-card exam" @click="$router.push('/student/exams')">
          <div class="module-icon">📋</div>
          <div class="module-info">
            <h3>考试中心</h3>
            <p>查看可考考试、参加考试、查看成绩记录</p>
          </div>
          <div class="module-badge" v-if="activeExamCount > 0">{{ activeExamCount }} 场进行中</div>
          <div class="module-arrow">→</div>
        </div>

        <!-- 投票 -->
        <div class="module-card vote" @click="$router.push('/student/votes')">
          <div class="module-icon">🗳️</div>
          <div class="module-info">
            <h3>投票</h3>
            <p>参与投票活动，表达您的意见</p>
          </div>
          <div class="module-badge" v-if="activeVoteCount > 0">{{ activeVoteCount }} 个进行中</div>
          <div class="module-arrow">→</div>
        </div>

        <!-- 问卷 -->
        <div class="module-card survey" @click="$router.push('/student/surveys')">
          <div class="module-icon">📋</div>
          <div class="module-info">
            <h3>问卷调查</h3>
            <p>填写问卷，帮助我们做得更好</p>
          </div>
          <div class="module-badge" v-if="activeSurveyCount > 0">{{ activeSurveyCount }} 个进行中</div>
          <div class="module-arrow">→</div>
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
const activeExamCount = ref(0)
const activeVoteCount = ref(0)
const activeSurveyCount = ref(0)

async function loadStats() {
  try {
    const [examsRes, votesRes, surveysRes] = await Promise.all([
      api('/student/exams'),
      api('/student/votes'),
      api('/student/surveys')
    ])
    const exams = examsRes.data || []
    const votes = votesRes.data || []
    const surveys = surveysRes.data || []
    activeExamCount.value = exams.filter(e => e.status === 'active' && (!e.hasAttempted || e.allowRepeat)).length
    activeVoteCount.value = votes.filter(v => !v.hasVoted && v.status !== 'ended').length
    activeSurveyCount.value = surveys.filter(s => !s.hasCompleted && s.status !== 'ended').length
  } catch (e) { console.error(e) }
}

async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const user = currentUser.value
  if (user?.isAdmin) { showToast('管理员不能访问员工端', 'error'); setTimeout(() => router.push('/admin'), 1500); return }
  userName.value = user?.name || user?.workNo || '未知用户'
  await loadStats()
})
</script>

<style scoped>
.home-page { min-height: 100vh; background: linear-gradient(160deg, #f0f3ff 0%, #faf5ff 50%, #f0fdf4 100%); }
.main-container { max-width: 900px; margin: 0 auto; padding: 0 20px; }
.welcome-section { text-align: center; padding: 50px 20px 30px; }
.welcome-section h1 { font-size: 28px; font-weight: 800; color: #1a1a2e; margin-bottom: 8px; }
.welcome-section p { font-size: 15px; color: #888; }
.module-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; padding-bottom: 40px; }
.module-card {
  display: flex; align-items: center; gap: 18px;
  background: white; border-radius: 16px; padding: 28px 24px;
  box-shadow: 0 2px 16px rgba(0,0,0,0.05); cursor: pointer;
  transition: all 0.3s; position: relative; overflow: hidden;
  border: 1px solid transparent;
}
.module-card:hover { transform: translateY(-4px); box-shadow: 0 8px 30px rgba(0,0,0,0.1); }
.module-card.exam:hover { border-color: #c7d2fe; }
.module-card.vote:hover { border-color: #ddd6fe; }
.module-card.survey:hover { border-color: #a7f3d0; }
.module-card.records:hover { border-color: #fde68a; }
.module-icon { font-size: 42px; flex-shrink: 0; line-height: 1; }
.module-info { flex: 1; min-width: 0; }
.module-info h3 { font-size: 18px; font-weight: 700; color: #1a1a2e; margin-bottom: 6px; }
.module-info p { font-size: 13px; color: #888; line-height: 1.5; }
.module-badge { position: absolute; top: 16px; right: 16px; padding: 3px 10px; border-radius: 10px; font-size: 11px; font-weight: 600; background: #eef2ff; color: #6366f1; }
.module-arrow { font-size: 20px; color: #ccc; flex-shrink: 0; transition: all 0.3s; }
.module-card:hover .module-arrow { color: #667eea; transform: translateX(4px); }

@media (max-width: 768px) {
  .module-grid { grid-template-columns: 1fr; gap: 12px; }
  .welcome-section { padding: 30px 16px 20px; }
  .welcome-section h1 { font-size: 22px; }
  .module-card { padding: 22px 18px; }
  .module-icon { font-size: 36px; }
}
</style>
