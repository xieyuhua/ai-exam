<template>
  <div class="admin-page">
    <Topbar
      title="⚙️ 管理后台"
      :userName="userName"
      :showChangePassword="true"
      @logout="handleLogout"
      @changePassword="pwdModalShow = true"
    />

    <div class="tab-container">
      <div class="tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['tab', { active: activeTab === tab.key }]"
          @click="switchTab(tab.key)"
        >{{ tab.icon }} {{ tab.label }}</button>
      </div>

      <div class="content">
        <DashboardPanel v-if="activeTab === 'dashboard'" :stats="dashboardStats" />
        <CategoryPanel v-if="activeTab === 'categories'" ref="categoryPanelRef" @changed="refreshStats" />
        <ExamPanel v-if="activeTab === 'exams'" ref="examPanelRef" :categories="categoryList" @changed="refreshStats" @manageQuestions="openExamQuestions" />

        <StudentPanel v-if="activeTab === 'students'" ref="studentPanelRef" />

        <!-- 投票管理 -->
        <VotePanel v-if="activeTab === 'votes'" ref="votePanelRef" />

        <!-- 问卷管理 -->
        <SurveyPanel v-if="activeTab === 'surveys'" ref="surveyPanelRef" />

        <div v-if="activeTab === 'examQuestions'">
          <button class="back-btn" @click="closeExamQuestions">← 返回考试列表</button>
          <ExamQuestionManager :examId="currentExam.id" :examTitle="currentExam.title" :categories="categoryList" />
        </div>
      </div>
    </div>

    <PasswordModal :show="pwdModalShow" @close="pwdModalShow = false" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, checkAuth, logout, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'
import Topbar from '@/components/Topbar.vue'
import PasswordModal from '@/components/admin/PasswordModal.vue'
import DashboardPanel from '@/components/admin/DashboardPanel.vue'
import CategoryPanel from '@/components/admin/CategoryPanel.vue'
import ExamPanel from '@/components/admin/ExamPanel.vue'
import StudentPanel from '@/components/admin/StudentPanel.vue'
import VotePanel from '@/components/admin/VotePanel.vue'
import SurveyPanel from '@/components/admin/SurveyPanel.vue'
import ExamQuestionManager from '@/components/admin/ExamQuestionManager.vue'

const router = useRouter()
const userName = ref('')
const activeTab = ref('dashboard')
const pwdModalShow = ref(false)

const tabs = [
  { key: 'dashboard', icon: '📊', label: '总览' },
  { key: 'categories', icon: '📂', label: '分类' },
  { key: 'exams', icon: '📋', label: '考试' },
  { key: 'students', icon: '👤', label: '员工' },
  { key: 'votes', icon: '🗳️', label: '投票' },
  { key: 'surveys', icon: '📋', label: '问卷' }
]

const dashboardStats = reactive({ exams: 0, students: 0, votes: 0, surveys: 0, activeExams: 0, activeVotes: 0, activeSurveys: 0 })
const categoryList = ref([])
const selectedCategoryId = ref(0)
const selectedCategoryName = ref('')
const currentExam = ref({})

const categoryPanelRef = ref(null)
const examPanelRef = ref(null)
const studentPanelRef = ref(null)
const votePanelRef = ref(null)
const surveyPanelRef = ref(null)

async function loadCategories() {
  const res = await api('/admin/categories')
  categoryList.value = res.data || []
}

async function refreshStats() {
  try {
    const [stuRes, voteRes, surveyRes, voteActiveRes, surveyActiveRes] = await Promise.all([
      api('/admin/students?pageSize=1'),
      api('/admin/votes?pageSize=1').catch(() => ({ data: { total: 0 } })),
      api('/admin/surveys?pageSize=1').catch(() => ({ data: { total: 0 } })),
      api('/admin/votes?pageSize=1&status=active').catch(() => ({ data: { total: 0 } })),
      api('/admin/surveys?pageSize=1&status=active').catch(() => ({ data: { total: 0 } }))
    ])
    dashboardStats.students = stuRes.data?.total || 0
    dashboardStats.votes = voteRes.data?.total || 0
    dashboardStats.surveys = surveyRes.data?.total || 0
    dashboardStats.activeVotes = voteActiveRes.data?.total || 0
    dashboardStats.activeSurveys = surveyActiveRes.data?.total || 0
    try {
      const eRes = await api('/admin/exams?pageSize=1')
      dashboardStats.exams = eRes.data?.total || 0
    } catch (e) {}
    const catRes = await api('/admin/categories')
    if (catRes.data) categoryList.value = catRes.data
  } catch (e) {}
}

function switchTab(key) {
  activeTab.value = key
  if (selectedCategoryId.value) selectedCategoryId.value = 0
  if (key === 'exams') examPanelRef.value?.load()
  if (key === 'students') studentPanelRef.value?.load()
  if (key === 'votes') votePanelRef.value?.load()
  if (key === 'surveys') surveyPanelRef.value?.load()
}

function openExamQuestions(exam) { currentExam.value = exam; activeTab.value = 'examQuestions' }
function closeExamQuestions() { activeTab.value = 'exams'; examPanelRef.value?.load() }
async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const user = currentUser.value
  if (!user?.isAdmin) { showToast('非管理员用户，跳转员工端...', 'error'); setTimeout(() => router.push('/student'), 1500); return }
  userName.value = user?.name || user?.workNo || '管理员'
  await Promise.all([loadCategories(), refreshStats()])
})
</script>

<style scoped>
.admin-page { min-height: 100vh; background: var(--bg); }
.tab-container { max-width: 1300px; margin: 24px auto; padding: 0 20px; }
.tabs { display: flex; background: var(--white); border-radius: var(--radius) var(--radius) 0 0; overflow: hidden; box-shadow: var(--shadow-sm); }
.tab { flex: 1; padding: 15px 10px; text-align: center; background: #f8fafc; border: none; cursor: pointer; font-size: 14px; font-weight: 600; color: var(--text-secondary); transition: all var(--transition); border-bottom: 3px solid transparent; white-space: nowrap; }
.tab.active { background: var(--white); color: var(--primary); border-bottom-color: var(--primary); }
.tab:hover:not(.active) { background: #eef2ff; color: var(--primary-dark); }
.content { background: var(--white); border-radius: 0 0 var(--radius) var(--radius); padding: 28px; box-shadow: var(--shadow); min-height: 450px; animation: fadeIn 0.25s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }
.section-title { font-size: 18px; font-weight: 700; color: var(--text); margin-bottom: 20px; padding-bottom: 14px; border-bottom: 2px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center; gap: 10px; }
.back-btn { display: inline-flex; align-items: center; gap: 5px; background: none; border: none; color: var(--primary); font-size: 14px; font-weight: 600; cursor: pointer; padding: 6px 0; margin-bottom: 15px; }
.back-btn:hover { color: var(--primary-dark); }
.meta { color: var(--text-secondary); font-size: 13px; }
@media (max-width: 768px) {
  .tab-container { padding: 0 8px; margin: 12px auto; }
  .tabs { overflow-x: auto; -webkit-overflow-scrolling: touch; }
  .tab { padding: 12px 10px; font-size: 12px; flex: none; min-width: 60px; }
  .content { padding: 16px 12px; border-radius: 0 0 8px 8px; }
}
</style>
