<template>
  <div class="list-page">
    <Topbar title="📋 问卷列表" :userName="userName" @logout="handleLogout" />
    <div class="main-container">
      <button class="back-btn" @click="$router.push('/student')">← 返回首页</button>

      <div class="content">
        <div class="loader" v-if="loading"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>
        <div v-else-if="surveyList.length === 0" class="empty-state">
          <div class="empty-icon">📋</div>
          <h3>暂无问卷</h3>
          <p>管理员尚未创建问卷调查</p>
        </div>
        <div v-else>
          <div v-for="survey in surveyList" :key="survey.id" class="card">
            <div class="card-left">
              <div class="title-row">
                <h3>{{ survey.title }}</h3>
                <span :class="['badge', surveyStatusCls(survey)]">{{ surveyStatusText(survey) }}</span>
              </div>
              <div class="card-meta" v-if="survey.description">
                <span>{{ survey.description }}</span>
              </div>
              <div class="card-info">
                <span>📝 {{ survey.questionCount || 0 }} 题</span>
                <span>📅 {{ survey.startTime }} ~ {{ survey.endTime }}</span>
              </div>
            </div>
            <div class="card-right">
              <button class="btn btn-primary" @click="$router.push(`/survey?id=${survey.id}`)">填写问卷 →</button>
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
const surveyList = ref([])
const loading = ref(false)

function surveyStatusCls(survey) {
  if (survey.hasCompleted) return 'green'
  if (survey.status === 'ended') return 'gray'
  if (survey.status === 'upcoming') return 'orange'
  return 'green'
}
function surveyStatusText(survey) {
  if (survey.hasCompleted) return '✓ 已填写'
  if (survey.status === 'ended') return '已结束'
  if (survey.status === 'upcoming') return '未开始'
  return '进行中'
}

async function loadSurveys() {
  loading.value = true
  try { const res = await api('/student/surveys'); surveyList.value = res.data || [] } catch (e) { console.error(e) } finally { loading.value = false }
}
async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const u = currentUser.value
  if (u?.isAdmin) { showToast('管理员不能填写问卷', 'error'); setTimeout(() => router.push('/admin'), 1000); return }
  userName.value = u?.name || u?.workNo || ''
  await loadSurveys()
})
</script>

<style scoped>
.list-page { min-height: 100vh; background: linear-gradient(160deg, #ecfdf5 0%, #f0fdf4 50%, #ecfdf5 100%); }
.main-container { max-width: 800px; margin: 30px auto; padding: 0 20px; }
.back-btn { display: inline-flex; align-items: center; gap: 5px; background: none; border: none; color: #059669; font-size: 14px; font-weight: 600; cursor: pointer; margin-bottom: 16px; }
.content { background: white; border-radius: 16px; padding: 28px 30px; min-height: 420px; box-shadow: 0 2px 16px rgba(0,0,0,0.06); }
.card { display: flex; align-items: center; justify-content: space-between; border: 1px solid #edf0f5; border-radius: 12px; padding: 20px 24px; margin-bottom: 12px; transition: all 0.2s; }
.card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); border-color: #a7f3d0; }
.card-left { flex: 1; min-width: 0; }
.card-right { flex-shrink: 0; margin-left: 20px; }
.title-row { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; flex-wrap: wrap; }
.title-row h3 { font-size: 17px; color: #1a1a2e; font-weight: 600; }
.card-meta { font-size: 13px; color: #888; margin-bottom: 8px; }
.card-info { display: flex; gap: 14px; font-size: 13px; color: #999; flex-wrap: wrap; }
.badge { padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.badge.green { background: #f0fdf4; color: #059669; }
.badge.orange { background: #fffbeb; color: #d97706; }
.badge.gray { background: #f1f5f9; color: #64748b; }
.loader { text-align: center; padding: 40px 0; }
.loader-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #059669; margin: 0 4px; animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.2s; }
.loader-dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce { to { transform: translateY(-10px); opacity: 0.4; } }
.empty-state { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-state h3 { font-size: 18px; color: #333; margin-bottom: 8px; }
.empty-state p { font-size: 14px; color: #999; }
.btn { padding: 10px 22px; border: none; border-radius: 10px; font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.3s; }
.btn-primary { background: linear-gradient(135deg, #059669, #10b981); color: white; }
.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(5,150,105,0.3); }
@media (max-width: 768px) { .main-container { margin: 16px auto; padding: 0 12px; } .content { padding: 18px 14px; } .card { flex-direction: column; align-items: stretch; gap: 14px; } .card-right { margin-left: 0; } }
</style>
