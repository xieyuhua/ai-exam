<template>
  <div class="student-page">
    <Topbar title="📋 考试中心" :userName="userName" @logout="handleLogout" />
    <div class="main-container">
      <button class="back-btn" @click="$router.push('/student')">← 返回首页</button>

      <div class="tabs">
        <button :class="['tab', { active: activeTab === 'exams' }]" @click="activeTab = 'exams'">📋 可考考试</button>
        <button :class="['tab', { active: activeTab === 'records' }]" @click="activeTab = 'records'; loadRecords()">📊 考试记录</button>
      </div>

      <div class="content">
        <!-- 考试列表 -->
        <div v-if="activeTab === 'exams'">
          <div class="loader" v-if="examsLoading"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>
          <div v-else-if="examList.length === 0" class="empty-state">
            <div class="empty-icon">📭</div>
            <h3>暂无可考考试</h3>
            <p>请联系管理员创建考试</p>
          </div>
          <div v-else>
            <div v-for="exam in examList" :key="exam.id" :class="['card', `status-${exam.status || 'ended'}`]">
              <div class="exam-left">
                <div class="exam-title-row">
                  <h3>{{ exam.title }}</h3>
                  <span :class="['badge', statusBadge(exam.status).cls]">{{ statusBadge(exam.status).icon }} {{ statusBadge(exam.status).text }}</span>
                </div>
                <div class="exam-meta">
                  <span>📝 {{ exam.questionCount }}题</span>
                  <span>⏱ {{ exam.duration }}分钟</span>
                  <span>💯 {{ exam.totalScore }}分</span>
                  <span>📅 {{ exam.startTime }} ~ {{ exam.endTime }}</span>
                </div>
              </div>
              <div class="exam-right">
                <button class="btn btn-primary" :disabled="!canStart(exam)" @click="showStartModal(exam)">{{ examBtnText(exam) }}</button>
              </div>
            </div>
          </div>
        </div>

        <!-- 考试记录 -->
        <div v-if="activeTab === 'records'">
          <div class="loader" v-if="recordsLoading"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>
          <div v-else-if="recordList.length === 0" class="empty-state">
            <div class="empty-icon">📝</div>
            <h3>暂无考试记录</h3>
            <p>参加考试后成绩将显示在这里</p>
          </div>
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead><tr><th>考试名称</th><th>考试日期</th><th>成绩</th><th>正确率</th></tr></thead>
              <tbody>
                <tr v-for="r in recordList" :key="r.id">
                  <td><strong>{{ r.examTitle }}</strong></td>
                  <td>{{ r.date || '--' }}</td>
                  <td><span :class="['badge', scoreClass(r)]">{{ Number(r.score || 0).toFixed(1) }}分</span></td>
                  <td><span :class="rateClass(r)">{{ r.correctCount || 0 }}/{{ r.totalCount || 1 }} ({{ rate(r) }}%)</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- 开始考试确认弹窗 -->
    <Modal :show="startModalShow" :title="'📝 确认开始考试'" :maxWidth="'440px'" @close="startModalShow = false">
      <div class="modal-info" v-if="selectedExam">
        <div class="row"><span>考试名称</span><strong>{{ selectedExam.title }}</strong></div>
        <div class="row"><span>题目数量</span><strong>{{ selectedExam.questionCount }} 题</strong></div>
        <div class="row"><span>考试时长</span><strong>{{ selectedExam.duration }} 分钟</strong></div>
        <div class="row"><span>总分</span><strong>{{ selectedExam.totalScore }} 分</strong></div>
      </div>
      <div class="modal-hint">开始后将进入计时，请确保在安静的环境中答题</div>
      <div class="modal-buttons">
        <button class="modal-btn cancel" @click="startModalShow = false">再想想</button>
        <button class="modal-btn confirm" @click="confirmStartExam">开始考试</button>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, checkAuth, logout, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'
import Topbar from '@/components/Topbar.vue'
import Modal from '@/components/Modal.vue'

const router = useRouter()
const userName = ref('')
const activeTab = ref('exams')
const examList = ref([])
const recordList = ref([])
const examsLoading = ref(false)
const recordsLoading = ref(false)
const startModalShow = ref(false)
const selectedExam = ref(null)

function statusBadge(status) {
  const map = {
    active: { text: '进行中', cls: 'green', icon: '🟢' },
    upcoming: { text: '未开始', cls: 'orange', icon: '🟠' },
    ended: { text: '已结束', cls: 'gray', icon: '⚫' }
  }
  return map[status] || map.ended
}
function canStart(exam) {
  if (exam.hasAttempted && !exam.allowRepeat) return false
  // 优先使用后端返回的状态，配合客户端时间做兜底判断
  if (exam.status === 'active') return true
  // 客户端时间兜底：后端状态可能未及时更新（如旧数据时区问题）
  const now = Date.now()
  const startTime = parseTime(exam.startTime)
  const endTime = parseTime(exam.endTime)
  return now >= startTime && now < endTime
}
function examBtnText(exam) {
  if (exam.hasAttempted && !exam.allowRepeat) return '✓ 已完成'
  const now = Date.now()
  const startTime = parseTime(exam.startTime)
  const endTime = parseTime(exam.endTime)
  if (exam.status === 'active' || (now >= startTime && now < endTime)) return '▶ 开始考试'
  if (exam.status === 'ended' || now >= endTime) return '已结束'
  return '未开始'
}
function parseTime(str) {
  if (!str) return NaN
  // 兼容浏览器差异：将空格替换为 T，确保 new Date() 能正确解析
  return new Date(str.replace(' ', 'T')).getTime()
}
function scoreClass(r) {
  const rt = rate(r); if (rt >= 80) return 'green'; if (rt < 60) return 'red'; return 'orange'
}
function rateClass(r) {
  const rt = rate(r); if (rt >= 80) return 'rate-high'; if (rt < 60) return 'rate-low'; return 'rate-mid'
}
function rate(r) { return Math.round((r.correctCount || 0) / (r.totalCount || 1) * 100) }

async function loadExams() {
  examsLoading.value = true
  try { const res = await api('/student/exams'); examList.value = res.data || [] } catch (e) { console.error(e) } finally { examsLoading.value = false }
}
async function loadRecords() {
  recordsLoading.value = true
  try { const res = await api('/student/records'); recordList.value = res.data || [] } catch (e) { console.error(e) } finally { recordsLoading.value = false }
}
function showStartModal(exam) { selectedExam.value = exam; startModalShow.value = true }
function confirmStartExam() {
  if (!selectedExam.value) return; startModalShow.value = false; router.push(`/exam?examId=${selectedExam.value.id}`)
}
async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const user = currentUser.value
  if (user?.isAdmin) { showToast('管理员不能参加考试，正在跳转管理端...', 'error'); setTimeout(() => router.push('/admin'), 1500); return }
  userName.value = user?.name || user?.workNo || '未知用户'
  await Promise.all([loadExams(), loadRecords()])
})
</script>

<style scoped>
.student-page { min-height: 100vh; background: linear-gradient(160deg, #f0f3ff 0%, #faf5ff 50%, #f0fdf4 100%); }
.main-container { max-width: 960px; margin: 30px auto; padding: 0 20px; }
.back-btn { display: inline-flex; align-items: center; gap: 5px; background: none; border: none; color: #667eea; font-size: 14px; font-weight: 600; cursor: pointer; margin-bottom: 16px; }
.tabs { display: flex; border-bottom: 2px solid #f0f0f0; background: white; border-radius: 12px 12px 0 0; overflow: hidden; }
.tab { flex: 1; padding: 14px 20px; text-align: center; background: transparent; border: none; cursor: pointer; font-size: 14px; font-weight: 600; color: #999; transition: all 0.3s; position: relative; }
.tab::after { content: ''; position: absolute; bottom: 0; left: 50%; transform: translateX(-50%); width: 0; height: 3px; background: linear-gradient(90deg, #667eea, #764ba2); border-radius: 3px; transition: width 0.3s; }
.tab.active { color: #667eea; }
.tab.active::after { width: 60%; }
.content { background: white; padding: 28px 30px; min-height: 420px; border-radius: 0 0 12px 12px; box-shadow: 0 2px 16px rgba(0,0,0,0.06); }
.exam-left { flex: 1; min-width: 0; }
.exam-title-row { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.exam-title-row h3 { font-size: 17px; color: #1a1a2e; font-weight: 600; }
.exam-meta { display: flex; gap: 18px; flex-wrap: wrap; font-size: 13px; color: #888; }
.exam-right { flex-shrink: 0; margin-left: 20px; }
.table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid #edf0f5; }
.badge { padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.badge.green { background: #f0fdf4; color: #16a34a; }
.badge.orange { background: #fffbeb; color: #d97706; }
.badge.red { background: #fef2f2; color: #dc2626; }
.rate-high { color: #16a34a; font-weight: 600; }
.rate-mid { color: #d97706; font-weight: 600; }
.rate-low { color: #dc2626; font-weight: 600; }
.card { display: flex; align-items: center; justify-content: space-between; border: 1px solid #edf0f5; border-radius: 12px; padding: 20px 24px; margin-bottom: 12px; transition: all 0.2s; background: white; }
.card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); border-color: #c7d2fe; }
.loader { text-align: center; padding: 40px 0; }
.loader-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #667eea; margin: 0 4px; animation: bounce 0.6s infinite alternate; }
.loader-dot:nth-child(2) { animation-delay: 0.2s; }
.loader-dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce { to { transform: translateY(-10px); opacity: 0.4; } }
.empty-state { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-state h3 { font-size: 18px; color: #333; margin-bottom: 8px; }
.empty-state p { font-size: 14px; color: #999; }
.modal-info { background: #f8f9fc; border-radius: 12px; padding: 16px 20px; margin-bottom: 18px; }
.modal-info .row { display: flex; justify-content: space-between; align-items: center; padding: 9px 0; font-size: 14px; color: #555; }
.modal-info .row:not(:last-child) { border-bottom: 1px solid #edf0f5; }
.modal-info .row strong { color: #333; font-weight: 600; }
.modal-hint { font-size: 13px; color: #999; text-align: center; margin-bottom: 22px; line-height: 1.6; }
.modal-buttons { display: flex; gap: 12px; }
.modal-btn { flex: 1; padding: 13px; border: none; border-radius: 12px; font-size: 15px; font-weight: 600; cursor: pointer; transition: all 0.3s; }
.modal-btn.cancel { background: #f0f1f5; color: #666; }
.modal-btn.cancel:hover { background: #e2e4ea; }
.modal-btn.confirm { background: linear-gradient(135deg, #667eea, #764ba2); color: white; box-shadow: 0 4px 16px rgba(102,126,234,0.35); }
.modal-btn.confirm:hover { transform: translateY(-2px); box-shadow: 0 6px 24px rgba(102,126,234,0.5); }
@media (max-width: 768px) { .main-container { margin: 16px auto; padding: 0 12px; } .content { padding: 18px 14px; } .card { flex-direction: column; align-items: stretch; gap: 14px; } .exam-right { margin-left: 0; } .modal-buttons { flex-direction: column; } }
</style>
