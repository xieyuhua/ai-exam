<template>
  <div class="survey-page">
    <Topbar title="📋 问卷" :userName="userName" @logout="handleLogout" />
    <div class="main-container">
      <button class="back-btn" @click="$router.push('/student/surveys')">← 返回问卷列表</button>

      <div v-if="loading" class="loader"><span class="loader-dot"></span><span class="loader-dot"></span><span class="loader-dot"></span></div>

      <template v-else-if="survey">
        <div class="survey-card">
          <h2>{{ survey.title }}</h2>
          <p class="desc" v-if="survey.description">{{ survey.description }}</p>
          <div class="meta-row">
            <span class="badge green" v-if="survey.status === 'active'">🟢 进行中</span>
            <span class="badge orange" v-else>🟠 未开始</span>
          </div>

          <div v-if="hasCompleted" class="completed-banner">✅ 您已完成此问卷</div>

          <!-- 问题列表 -->
          <div class="question-list" v-for="(q, qi) in survey.questions" :key="q.id">
            <div class="q-title">
              {{ qi + 1 }}. {{ q.title }}
              <span v-if="q.required !== false" class="required-mark">*</span>
            </div>

            <!-- 单选 -->
            <div v-if="q.type === 'single'" class="options-list">
              <div
                v-for="opt in q.options" :key="opt.label"
                :class="['option-item', { selected: answers[qi]?.value === opt.label }]"
                @click="answers[qi] = { ...answers[qi], value: opt.label }"
              >
                <span class="opt-radio">{{ answers[qi]?.value === opt.label ? '🔘' : '⚪' }}</span>
                <img v-if="opt.type === 'image' && opt.url" :src="opt.url" class="opt-img" />
                <video v-if="opt.type === 'video' && opt.url" :src="opt.url" controls muted class="opt-video"></video>
                <span v-if="opt.content" class="opt-text">{{ opt.content }}</span>
              </div>
            </div>

            <!-- 多选 -->
            <div v-if="q.type === 'multiple'" class="options-list">
              <div
                v-for="opt in q.options" :key="opt.label"
                :class="['option-item', { selected: (answers[qi]?.values || []).includes(opt.label) }]"
                @click="toggleMulti(qi, opt.label)"
              >
                <span class="opt-radio">{{ (answers[qi]?.values || []).includes(opt.label) ? '✅' : '⬜' }}</span>
                <img v-if="opt.type === 'image' && opt.url" :src="opt.url" class="opt-img" />
                <video v-if="opt.type === 'video' && opt.url" :src="opt.url" controls muted class="opt-video"></video>
                <span v-if="opt.content" class="opt-text">{{ opt.content }}</span>
              </div>
            </div>

            <!-- 填空 -->
            <div v-if="q.type === 'fill'">
              <input type="text" v-model="answers[qi].value" placeholder="请输入答案" class="text-input" />
            </div>

            <!-- 简答 -->
            <div v-if="q.type === 'essay'">
              <textarea v-model="answers[qi].value" placeholder="请输入您的回答" rows="3" class="text-input"></textarea>
            </div>
          </div>

          <button
            class="submit-btn"
            :disabled="submitting || hasCompleted"
            @click="doSubmit"
          >{{ submitting ? '提交中...' : '📮 提交问卷' }}</button>
        </div>
      </template>
      <div v-else class="empty-state"><div class="empty-icon">📋</div><h3>问卷不存在</h3></div>
    </div>
    <Toast />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, checkAuth, logout, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'
import Topbar from '@/components/Topbar.vue'
import Toast from '@/components/Toast.vue'

const route = useRoute()
const router = useRouter()
const userName = ref('')
const survey = ref(null)
const answers = ref([])
const loading = ref(true)
const submitting = ref(false)
const hasCompleted = ref(false)

function toggleMulti(qi, label) {
  if (!answers.value[qi]) answers.value[qi] = { values: [] }
  if (!answers.value[qi].values) answers.value[qi].values = []
  const idx = answers.value[qi].values.indexOf(label)
  if (idx > -1) answers.value[qi].values.splice(idx, 1)
  else answers.value[qi].values.push(label)
}

function initAnswers() {
  if (!survey.value?.questions) return
  answers.value = survey.value.questions.map(q => {
    if (q.type === 'multiple') return { values: [] }
    return { value: '' }
  })
}

async function loadSurvey() {
  loading.value = true
  try {
    const res = await api(`/student/surveys/${route.query.id}`)
    survey.value = res.data
    initAnswers()
  } catch (e) { showToast('加载问卷失败', 'error') } finally { loading.value = false }
}

async function doSubmit() {
  // 校验必填
  if (!survey.value) return
  for (let i = 0; i < survey.value.questions.length; i++) {
    const q = survey.value.questions[i]
    const a = answers.value[i]
    if (q.required !== false) {
      if (q.type === 'multiple' && (!a?.values || a.values.length === 0)) return showToast(`第${i + 1}题是必答题`, 'error')
      if (q.type !== 'multiple' && !a?.value) return showToast(`第${i + 1}题是必答题`, 'error')
    }
  }
  submitting.value = true
  try {
    const answerList = survey.value.questions.map((q, i) => {
      const a = answers.value[i]
      let ans = []
      if (q.type === 'multiple') ans = a?.values || []
      else if (a?.value) ans = [a.value]
      return { surveyQuestionId: q.id, answer: ans }
    })
    const res = await api('/student/surveys/submit', {
      method: 'POST', body: JSON.stringify({ surveyId: Number(route.query.id), answers: answerList })
    })
    if (res.code === 200) { showToast('提交成功', 'success'); hasCompleted.value = true }
    else showToast(res.message || '提交失败', 'error')
  } catch (e) { showToast('提交失败', 'error') } finally { submitting.value = false }
}

async function handleLogout() { await logout(); router.push('/') }

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  const u = currentUser.value
  if (u?.isAdmin) { showToast('管理员不能填写问卷', 'error'); setTimeout(() => router.push('/admin'), 1000); return }
  userName.value = u?.name || u?.workNo || ''
  await loadSurvey()
})
</script>

<style scoped>
.survey-page { min-height: 100vh; background: linear-gradient(160deg, #ecfdf5 0%, #f0fdf4 50%, #ecfdf5 100%); }
.main-container { max-width: 700px; margin: 30px auto; padding: 0 20px; }
.back-btn { display: inline-flex; align-items: center; gap: 5px; background: none; border: none; color: #059669; font-size: 14px; font-weight: 600; cursor: pointer; margin-bottom: 16px; }
.survey-card { background: white; border-radius: 16px; padding: 32px; box-shadow: 0 2px 20px rgba(0,0,0,0.06); }
.survey-card h2 { font-size: 22px; color: #1a1a2e; margin-bottom: 10px; }
.desc { color: #666; font-size: 14px; margin-bottom: 16px; line-height: 1.6; }
.meta-row { margin-bottom: 20px; }
.badge { padding: 3px 12px; border-radius: 12px; font-weight: 600; font-size: 12px; }
.badge.green { background: #f0fdf4; color: #059669; }
.badge.orange { background: #fffbeb; color: #d97706; }
.completed-banner { padding: 14px; background: #f0fdf4; border-radius: 10px; text-align: center; color: #059669; font-weight: 600; margin-bottom: 20px; }
.question-list { padding: 16px 0; border-bottom: 1px solid #f0f0f0; }
.q-title { font-size: 15px; font-weight: 600; color: #333; margin-bottom: 12px; }
.required-mark { color: #dc2626; margin-left: 4px; }
.options-list { display: grid; gap: 8px; }
.option-item { display: flex; align-items: center; gap: 10px; padding: 10px 14px; border: 2px solid #e5e7eb; border-radius: 10px; cursor: pointer; transition: all 0.2s; }
.option-item:hover { border-color: #34d399; background: #f0fdf4; }
.option-item.selected { border-color: #059669; background: #ecfdf5; }
.opt-radio { font-size: 18px; flex-shrink: 0; }
.opt-img, .opt-video { max-width: 120px; max-height: 80px; border-radius: 6px; }
.opt-text { flex: 1; font-size: 14px; color: #444; }
.text-input { width: 100%; padding: 10px 14px; border: 2px solid #e5e7eb; border-radius: 10px; font-size: 14px; margin-top: 4px; outline: none; }
.text-input:focus { border-color: #059669; }
textarea.text-input { resize: vertical; }
.submit-btn {
  width: 100%; padding: 15px; margin-top: 24px; background: linear-gradient(135deg, #059669, #10b981);
  color: white; border: none; border-radius: 12px; font-size: 16px; font-weight: 700;
  cursor: pointer; transition: all 0.3s; box-shadow: 0 4px 16px rgba(5,150,105,0.35);
}
.submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.submit-btn:hover:not(:disabled) { transform: translateY(-2px); }

@media (max-width: 768px) {
  .main-container { margin: 16px auto; padding: 0 12px; }
  .survey-card { padding: 20px 16px; }
}
</style>
