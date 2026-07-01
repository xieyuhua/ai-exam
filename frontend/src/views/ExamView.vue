<template>
  <div class="exam-page">
    <div class="exam-container">
      <!-- Header -->
      <div class="exam-header" v-show="examState === 'exam'">
        <h1>📝 {{ examTitle }}</h1>
        <div class="exam-info">
          <div class="info-item"><div class="label">总题数</div><div class="value">{{ questions.length }}</div></div>
          <div class="info-item"><div class="label">已答</div><div class="value">{{ answeredCount }}</div></div>
          <div class="info-item"><div class="label">未答</div><div class="value">{{ unansweredCount }}</div></div>
        </div>
      </div>

      <!-- Progress -->
      <div class="progress-container" v-show="examState === 'exam'">
        <div class="progress-bar"><div class="progress-fill" :style="{ width: progressWidth }"></div></div>
        <div class="progress-text">{{ currentIndex + 1 }} / {{ questions.length }}</div>
      </div>

      <!-- Loading -->
      <div v-if="examState === 'loading'" class="loading-container">
        <div class="loading-spinner"></div>
        <div class="loading-text">正在加载考试数据...</div>
      </div>

      <!-- Error -->
      <div v-if="examState === 'error'" class="loading-container">
        <div style="font-size:50px;margin-bottom:20px;">😢</div>
        <div style="color:#f44336;font-size:16px;">{{ errorMsg }}</div>
        <button class="btn btn-primary" style="margin-top:20px;width:auto;" @click="fetchExamData">🔄 重新加载</button>
      </div>

      <!-- Questions -->
      <div class="question-container" v-show="examState === 'exam'">
        <div class="question-number">
          第 {{ currentIndex + 1 }} 题
          <span :class="['question-type-badge', currentQuestion?.type || 'single']">{{ typeLabel }}</span>
          <span class="question-score">{{ currentQuestion?.score || 0 }}分</span>
        </div>
        <div class="question-title">{{ currentQuestion?.title }}</div>

        <!-- 选项类题目 -->
        <div v-if="isMultipleSelect" class="options-list">
          <div
            v-for="(opt, i) in currentQuestion?.options || []"
            :key="i"
            :class="['option-item', { selected: isOptionSelected(opt.label) }]"
            @click="toggleOption(opt.label)"
          >
            <div class="option-marker">{{ opt.label }}</div>
            <div class="option-text">{{ opt.content }}</div>
            <div class="option-checkbox"></div>
          </div>
          <div v-if="currentQuestion?.type === 'multiple'" class="hint-text">💡 多选题，请选择所有符合条件的选项</div>
        </div>

        <!-- 填空题 -->
        <input
          v-if="currentQuestion?.type === 'fill'"
          type="text"
          class="text-answer-input"
          :value="getTextAnswer()"
          @input="setTextAnswer($event.target.value)"
          placeholder="请输入你的答案..."
        />

        <!-- 简答题 -->
        <textarea
          v-if="currentQuestion?.type === 'essay'"
          class="text-answer-input textarea"
          :value="getTextAnswer()"
          @input="setTextAnswer($event.target.value)"
          placeholder="请输入你的答案..."
        ></textarea>
      </div>

      <!-- Buttons -->
      <div class="button-container" v-show="examState === 'exam'">
        <button class="btn btn-prev" :disabled="currentIndex === 0" @click="prevQuestion">← 上一题</button>
        <button class="btn btn-next" v-if="!isLastQuestion" @click="nextQuestion">下一题 →</button>
        <button class="btn btn-submit" v-if="isLastQuestion" @click="submitModalShow = true">🎯 提交试卷</button>
      </div>

      <!-- Result -->
      <div class="result-container" v-show="examState === 'result'">
        <div class="result-icon">🎉</div>
        <h2>考试结束</h2>
        <div class="result-score">{{ score }}分</div>
        <div class="result-details">
          <div class="result-item"><div class="label">总题数</div><div class="value">{{ totalQuestions }}</div></div>
          <div class="result-item"><div class="label">答对</div><div class="value" style="color:#4caf50;">{{ correctCount }}</div></div>
          <div class="result-item"><div class="label">答错</div><div class="value" style="color:#f44336;">{{ wrongCount }}</div></div>
          <div class="result-item"><div class="label">正确率</div><div class="value">{{ scoreDisplay }}%</div></div>
        </div>
        <div style="display:flex;gap:15px;justify-content:center;margin-top:30px;">
          <button class="btn btn-prev" @click="$router.push('/student/exams')">← 返回考试列表</button>
          <button v-if="canViewAnswer" class="btn btn-next" @click="showReview">📖 查看答案解析</button>
        </div>
      </div>

      <!-- Review -->
      <div class="review-container" v-show="examState === 'review'">
        <div class="review-header"><h2>📋 答案详情</h2></div>
        <div class="review-content">
          <div v-for="(q, idx) in questions" :key="q.id" class="review-question">
            <div class="review-question-header">
              <span class="review-question-num">第 {{ idx + 1 }} 题</span>
              <span :class="['question-type-badge', q.type]">{{ typeLabelMap[q.type] || q.type }}</span>
              <span :class="['review-result-badge', reviewStatusClass(q)]">{{ reviewStatusText(q) }}</span>
            </div>
              <div class="review-question-title">
                {{ q.title }}
                <span class="question-score" style="font-size:13px;margin-left:8px;">{{ q.score || 0 }}分</span>
              </div>

            <div v-if="isMultipleSelectType(q.type)" class="review-options">
              <div
                v-for="opt in q.options || []"
                :key="opt.label"
                :class="['review-option', reviewOptionClass(q, opt.label)]"
              >
                <div class="review-option-marker">{{ opt.label }}</div>
                <div class="review-option-text">{{ opt.content }}</div>
                <div class="review-option-icon">{{ reviewOptionIcon(q, opt.label) }}</div>
              </div>
            </div>

            <div style="margin-top:15px;font-size:14px;color:#666;">
              <strong>你的答案：</strong>{{ getUserAnswerStr(q) }}
              | <strong>正确答案：</strong>{{ getCorrectAnswerStr(q) }}
            </div>
            <div v-if="q.explanation" class="explanation-box">
              <h4>💡 答案解析</h4>
              <p>{{ q.explanation }}</p>
            </div>
          </div>
        </div>
        <div style="display:flex;gap:15px;justify-content:center;margin-top:30px;">
          <button class="btn btn-prev" @click="$router.push('/student/exams')">← 返回考试列表</button>
          <button class="btn btn-next" @click="examState = 'result'">📊 返回成绩</button>
        </div>
      </div>
    </div>

    <!-- 答题卡遮罩 -->
    <div v-if="sheetOpen" class="answer-sheet-overlay" @click="sheetOpen = false"></div>

    <!-- 答题卡 -->
    <div class="answer-sheet-toggle">
      <button class="answer-sheet-btn" @click="sheetOpen = !sheetOpen">📋 答题卡</button>
      <div :class="['answer-sheet-panel', { show: sheetOpen }]">
        <h3 style="margin-bottom:15px;text-align:center;">答题卡</h3>
        <div class="answer-grid">
          <div
            v-for="(q, i) in questions"
            :key="q.id"
            :class="['answer-cell', { answered: !!userAnswers[q.id], current: i === currentIndex }]"
            @click="jumpToQuestion(i)"
          >{{ i + 1 }}</div>
        </div>
        <div style="margin-top:15px;font-size:12px;color:#666;text-align:center;">
          <span style="color:#667eea;">■</span> 已答 <span style="color:#e0e0e0;">■</span> 未答 <span style="color:#f5576c;">□</span> 当前
        </div>
      </div>
    </div>

    <!-- Submit Modal -->
    <div :class="['modal-overlay', { show: submitModalShow }]" @click.self="submitModalShow = false">
      <div class="modal-dialog">
        <div class="modal-icon">📋</div>
        <h2>确认提交试卷？</h2>
        <p v-if="unansweredCount > 0">您还有 <strong>{{ unansweredCount }}</strong> 道题未作答，确定提交吗？</p>
        <p v-else>您已完成所有题目，确定提交吗？</p>
        <div class="modal-buttons">
          <button class="modal-btn cancel" @click="submitModalShow = false">继续答题</button>
          <button class="modal-btn confirm" @click="submitExam">确认提交</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { api, checkAuth, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'

const router = useRouter()
const route = useRoute()

// State
const examState = ref('loading') // loading | exam | result | review | error
const examTitle = ref('在线考试')
const questions = ref([])
const currentIndex = ref(0)
const userAnswers = ref({}) // { questionId: [answerLabels] }
const examResults = ref([])
const canViewAnswer = ref(false)
const errorMsg = ref('')
const examId = ref(0)

const submitModalShow = ref(false)
const sheetOpen = ref(false)

const score = ref(0)
const totalQuestions = ref(0)
const correctCount = ref(0)
const wrongCount = ref(0)

const typeLabelMap = { single: '单选题', multiple: '多选题', judge: '判断题', fill: '填空题', essay: '简答题' }

// Computed
const currentQuestion = computed(() => questions.value[currentIndex.value])
const typeLabel = computed(() => typeLabelMap[currentQuestion.value?.type] || currentQuestion.value?.type || '')
const isMultipleSelect = computed(() => {
  const t = currentQuestion.value?.type
  return t === 'single' || t === 'multiple' || t === 'judge'
})
const isLastQuestion = computed(() => currentIndex.value === questions.value.length - 1)
const answeredCount = computed(() => Object.keys(userAnswers.value).length)
const unansweredCount = computed(() => questions.value.length - answeredCount.value)
const progressWidth = computed(() => `${((currentIndex.value + 1) / (questions.value.length || 1)) * 100}%`)
const scoreDisplay = computed(() => Number(score.value || 0).toFixed(1))

function isMultipleSelectType(type) {
  return ['single', 'multiple', 'judge'].includes(type)
}

// Options
function isOptionSelected(label) {
  return (userAnswers.value[currentQuestion.value?.id] || []).includes(label)
}

function toggleOption(label) {
  const q = currentQuestion.value
  if (!q) return
  const isMulti = q.type === 'multiple'
  const current = userAnswers.value[q.id] || []

  if (isMulti) {
    const idx = current.indexOf(label)
    if (idx > -1) current.splice(idx, 1)
    else current.push(label)
    userAnswers.value[q.id] = [...current]
  } else {
    userAnswers.value[q.id] = [label]
  }
}

function getTextAnswer() {
  return (userAnswers.value[currentQuestion.value?.id] || [])[0] || ''
}

function setTextAnswer(val) {
  const q = currentQuestion.value
  if (!q) return
  if (val.trim()) {
    userAnswers.value[q.id] = [val.trim()]
  } else {
    delete userAnswers.value[q.id]
  }
}

// Navigation
function prevQuestion() {
  if (currentIndex.value > 0) currentIndex.value--
}
function nextQuestion() {
  if (currentIndex.value < questions.value.length - 1) currentIndex.value++
}
function jumpToQuestion(i) {
  currentIndex.value = i
  sheetOpen.value = false
}

// Exam
async function fetchExamData() {
  examState.value = 'loading'
  try {
    examId.value = parseInt(route.query.examId) || 0
    if (!examId.value) throw new Error('缺少考试ID')

    const studentName = currentUser.value?.name || currentUser.value?.workNo || 'unknown'
    const data = await api(`/exam/questions?examId=${examId.value}&studentName=${encodeURIComponent(studentName)}`)

    if (data.code !== 200) throw new Error(data.message || '获取考试数据失败')

    const resData = data.data
    if (resData.isCompleted) {
      canViewAnswer.value = resData.canViewAnswer ?? true
      examResults.value = resData.results || []
      questions.value = Array.isArray(resData.list) ? resData.list : (resData.list ? [resData.list] : [])
      score.value = resData.totalScore || 0
      totalQuestions.value = resData.totalQuestions || questions.value.length
      correctCount.value = resData.correctCount || 0
      wrongCount.value = resData.wrongCount || 0
      examState.value = 'result'
      return
    }

    questions.value = Array.isArray(resData.list) ? resData.list : (resData.list ? [resData.list] : [])
    if (questions.value.length === 0) throw new Error('考试题目为空')

    examTitle.value = resData.title || '在线考试'
    examState.value = 'exam'
  } catch (err) {
    errorMsg.value = '加载考试数据失败：' + err.message
    examState.value = 'error'
  }
}

async function submitExam() {
  submitModalShow.value = false
  try {
    const submitData = {
      examId: examId.value,
      studentName: currentUser.value?.name || currentUser.value?.workNo || 'unknown',
      answers: Object.entries(userAnswers.value).map(([qId, ans]) => ({
        questionId: parseInt(qId),
        selectedOptions: ans
      }))
    }

    const data = await api('/exam/submit', { method: 'POST', body: JSON.stringify(submitData) })
    if (data.code !== 200) throw new Error(data.message || '提交失败')

    const result = data.data
    canViewAnswer.value = result.canViewAnswer ?? true
    examResults.value = result.results || []
    score.value = result.totalScore || 0
    totalQuestions.value = result.totalQuestions || questions.value.length
    correctCount.value = result.correctCount || 0
    wrongCount.value = result.wrongCount || 0
    examState.value = 'result'
  } catch (err) {
    showToast('提交失败：' + err.message, 'error')
  }
}

// Review
function showReview() {
  if (!canViewAnswer.value) { showToast('当前考试不允许查看答案解析', 'error'); return }
  examState.value = 'review'
}

function reviewResult(q) {
  return examResults.value.find(r => r.questionId == q.id) || { isUnanswered: true, isCorrect: false, userAnswer: [], correctAnswer: [] }
}

function reviewStatusClass(q) {
  const r = reviewResult(q)
  if (r.isUnanswered) return 'unanswered'
  return r.isCorrect ? 'correct' : 'wrong'
}

function reviewStatusText(q) {
  const r = reviewResult(q)
  if (r.isUnanswered) return '未作答'
  return r.isCorrect ? '✓ 回答正确' : '✗ 回答错误'
}

function reviewOptionClass(q, label) {
  const r = reviewResult(q)
  const correctAns = r.correctAnswer || []
  const userAns = r.userAnswer || []
  if (correctAns.includes(label)) return 'correct-answer'
  if (userAns.includes(label) && !correctAns.includes(label)) return 'wrong-answer'
  return ''
}

function reviewOptionIcon(q, label) {
  const r = reviewResult(q)
  const correctAns = r.correctAnswer || []
  const userAns = r.userAnswer || []
  if (correctAns.includes(label)) return ' ✓'
  if (userAns.includes(label) && !correctAns.includes(label)) return ' ✗'
  return ''
}

function getUserAnswerStr(q) {
  const r = reviewResult(q)
  return (r.userAnswer || []).join(', ') || '未作答'
}

function getCorrectAnswerStr(q) {
  const r = reviewResult(q)
  return (r.correctAnswer || []).join(', ')
}

// Keyboard
function onKeydown(e) {
  if (examState.value !== 'exam') return
  if (e.key === 'ArrowLeft') prevQuestion()
  if (e.key === 'ArrowRight') nextQuestion()
  if (e.key === 'Escape') submitModalShow.value = false
}

onMounted(async () => {
  const authed = await checkAuth()
  if (!authed) { router.push('/'); return }
  if (currentUser.value?.isAdmin) { router.push('/admin'); return }
  document.addEventListener('keydown', onKeydown)
  await fetchExamData()
})
</script>

<style scoped>
.exam-page {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh; padding: 20px;
}
.exam-container {
  max-width: 900px; margin: 0 auto; background: white;
  border-radius: 20px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); overflow: hidden;
}

/* Header */
.exam-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white; padding: 30px; text-align: center;
}
.exam-header h1 { font-size: 28px; margin-bottom: 10px; }
.exam-info { display: flex; justify-content: space-around; margin-top: 20px; flex-wrap: wrap; gap: 15px; }
.info-item { background: rgba(255,255,255,0.2); padding: 10px 20px; border-radius: 10px; min-width: 100px; }
.info-item .label { font-size: 12px; opacity: 0.8; }
.info-item .value { font-size: 20px; font-weight: bold; margin-top: 5px; }

/* Progress */
.progress-container { padding: 20px 30px; background: #f8f9fa; }
.progress-bar { height: 8px; background: #e0e0e0; border-radius: 4px; overflow: hidden; }
.progress-fill { height: 100%; background: linear-gradient(90deg, #667eea 0%, #764ba2 100%); transition: width 0.3s ease; border-radius: 4px; }
.progress-text { text-align: right; margin-top: 5px; color: #666; font-size: 14px; }

/* Question */
.question-container { padding: 40px; min-height: 400px; }
.question-number { color: #667eea; font-size: 16px; font-weight: bold; margin-bottom: 10px; display: flex; align-items: center; gap: 10px; }
.question-score {
  margin-left: auto; padding: 4px 14px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white; border-radius: 20px; font-size: 13px; font-weight: bold;
}
.question-type-badge {
  display: inline-block; padding: 4px 12px; border-radius: 15px; font-size: 12px; margin-left: 10px;
}
.question-type-badge.single { background: #e3f2fd; color: #1976d2; }
.question-type-badge.multiple { background: #fff3e0; color: #f57c00; }
.question-type-badge.judge { background: #fce4ec; color: #c62828; }
.question-type-badge.fill { background: #e8f5e9; color: #2e7d32; }
.question-type-badge.essay { background: #f3e5f5; color: #7b1fa2; }
.question-title { font-size: 20px; color: #333; margin-bottom: 30px; line-height: 1.6; }

/* Options */
.options-list { display: flex; flex-direction: column; gap: 15px; }
.option-item {
  display: flex; align-items: center; padding: 18px 20px;
  border: 2px solid #e0e0e0; border-radius: 12px;
  cursor: pointer; transition: all 0.3s ease; background: #fafafa;
}
.option-item:hover { border-color: #667eea; background: #f0f4ff; transform: translateX(5px); }
.option-item.selected { border-color: #667eea; background: linear-gradient(135deg, #e8f0fe 0%, #f0e6ff 100%); }
.option-item.selected .option-marker { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
.option-marker {
  width: 36px; height: 36px; border-radius: 50%; background: #e0e0e0;
  display: flex; align-items: center; justify-content: center;
  font-weight: bold; color: #666; margin-right: 15px; flex-shrink: 0;
}
.option-text { font-size: 16px; color: #333; line-height: 1.5; }
.option-checkbox {
  margin-left: auto; width: 24px; height: 24px; border: 2px solid #ddd; border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
}
.option-item.selected .option-checkbox { background: #667eea; border-color: #667eea; }
.option-item.selected .option-checkbox::after { content: '✓'; color: white; font-weight: bold; }
.hint-text { color: #999; font-size: 14px; margin-top: 10px; }
.text-answer-input { width: 100%; padding: 14px 18px; border: 2px solid #e0e0e0; border-radius: 12px; font-size: 16px; margin-top: 10px; font-family: inherit; }
.text-answer-input:focus { border-color: #667eea; box-shadow: 0 0 0 3px rgba(102,126,234,0.1); outline: none; }
.textarea { min-height: 120px; resize: vertical; }

/* Buttons */
.button-container { display: flex; justify-content: space-between; padding: 30px 40px; background: #f8f9fa; border-top: 1px solid #e0e0e0; }
.btn { padding: 12px 35px; border: none; border-radius: 10px; font-size: 16px; font-weight: bold; cursor: pointer; transition: all 0.3s ease; display: flex; align-items: center; gap: 8px; }
.btn-prev { background: white; color: #667eea; border: 2px solid #667eea; }
.btn-prev:hover:not(:disabled) { background: #667eea; color: white; }
.btn-next { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
.btn-next:hover { transform: translateY(-2px); box-shadow: 0 5px 20px rgba(102,126,234,0.4); }
.btn-submit { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; }
.btn-submit:hover { transform: translateY(-2px); box-shadow: 0 5px 20px rgba(245,87,108,0.4); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Loading */
.loading-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80px 40px; }
.loading-spinner { width: 50px; height: 50px; border: 4px solid #e0e0e0; border-top-color: #667eea; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.loading-text { margin-top: 20px; color: #666; font-size: 16px; }

/* Result */
.result-container { padding: 50px; text-align: center; }
.result-icon { font-size: 80px; margin-bottom: 20px; }
.result-score {
  font-size: 60px; font-weight: bold;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin: 20px 0;
}
.result-details { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 20px; margin-top: 30px; }
.result-item { padding: 20px; background: #f8f9fa; border-radius: 12px; }
.result-item .label { color: #666; font-size: 14px; }
.result-item .value { font-size: 24px; font-weight: bold; color: #333; margin-top: 5px; }

/* Review */
.review-container { padding: 30px 40px; }
.review-header { text-align: center; margin-bottom: 30px; }
.review-header h2 { color: #333; font-size: 24px; margin-bottom: 10px; }
.review-question { background: white; border-radius: 15px; padding: 30px; box-shadow: 0 5px 20px rgba(0,0,0,0.08); margin-bottom: 20px; }
.review-question-header { display: flex; align-items: center; gap: 15px; margin-bottom: 20px; flex-wrap: wrap; }
.review-question-num { font-size: 18px; font-weight: bold; color: #667eea; }
.review-result-badge { padding: 5px 15px; border-radius: 20px; font-size: 13px; font-weight: bold; }
.review-result-badge.correct { background: #e8f5e9; color: #4caf50; }
.review-result-badge.wrong { background: #ffebee; color: #f44336; }
.review-result-badge.unanswered { background: #fff3e0; color: #ff9800; }
.review-question-title { font-size: 18px; color: #333; margin-bottom: 25px; line-height: 1.6; }
.review-options { display: flex; flex-direction: column; gap: 12px; }
.review-option { display: flex; align-items: center; padding: 15px 20px; border-radius: 10px; border: 2px solid #e0e0e0; background: #fafafa; }
.review-option.correct-answer { border-color: #4caf50; background: #e8f5e9; }
.review-option.wrong-answer { border-color: #f44336; background: #ffebee; }
.review-option-marker {
  width: 32px; height: 32px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-weight: bold; margin-right: 15px; flex-shrink: 0; background: #e0e0e0; color: #666;
}
.review-option.correct-answer .review-option-marker { background: #4caf50; color: white; }
.review-option.wrong-answer .review-option-marker { background: #f44336; color: white; }
.review-option-text { flex: 1; font-size: 15px; color: #333; }
.review-option-icon { font-size: 20px; margin-left: 10px; }
.explanation-box {
  margin-top: 25px; padding: 20px; background: linear-gradient(135deg, #e8f0fe 0%, #f0e6ff 100%);
  border-radius: 12px; border-left: 4px solid #667eea;
}
.explanation-box h4 { color: #667eea; font-size: 15px; margin-bottom: 10px; }
.explanation-box p { color: #555; line-height: 1.8; font-size: 14px; }

/* Answer Sheet Overlay */
.answer-sheet-overlay { position: fixed; inset: 0; z-index: 99; background: transparent; }

/* Answer Sheet */
.answer-sheet-toggle { position: fixed; right: 20px; top: 50%; transform: translateY(-50%); z-index: 100; }
.answer-sheet-btn {
  padding: 15px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white; border: none; cursor: pointer; font-weight: bold;
  border-radius: 15px 15px 0 0;
}
.answer-sheet-panel { display: none; padding: 20px; max-height: 500px; overflow-y: auto; background: white; border-radius: 0 0 15px 15px; box-shadow: 0 5px 20px rgba(0,0,0,0.2); }
.answer-sheet-panel.show { display: block; }
.answer-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 10px; width: 220px; }
.answer-cell {
  width: 40px; height: 40px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-weight: bold; cursor: pointer; border: 2px solid #e0e0e0; color: #666;
  transition: all 0.3s ease;
}
.answer-cell.answered { background: #667eea; color: white; border-color: #667eea; }
.answer-cell.current { border-color: #f5576c; box-shadow: 0 0 0 3px rgba(245,87,108,0.3); }
.answer-cell:hover { transform: scale(1.1); }

/* Submit Modal */
.modal-overlay {
  display: none; position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5); z-index: 1000; align-items: center; justify-content: center;
}
.modal-overlay.show { display: flex; }
.modal-dialog { background: white; border-radius: 20px; padding: 40px; max-width: 500px; width: 90%; text-align: center; animation: modalIn 0.3s ease; }
@keyframes modalIn { from { transform: scale(0.8); opacity: 0; } to { transform: scale(1); opacity: 1; } }
.modal-icon { font-size: 60px; margin-bottom: 20px; }
.modal-dialog h2 { color: #333; margin-bottom: 15px; }
.modal-dialog p { color: #666; margin-bottom: 25px; line-height: 1.6; }
.modal-buttons { display: flex; gap: 15px; justify-content: center; }
.modal-btn { padding: 12px 30px; border: none; border-radius: 10px; font-size: 16px; font-weight: bold; cursor: pointer; }
.modal-btn.cancel { background: #e0e0e0; color: #666; }
.modal-btn.confirm { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; }
.modal-btn:hover { transform: translateY(-2px); }

@media (max-width: 768px) {
  .exam-page { padding: 0; }
  .exam-container { border-radius: 0; min-height: 100vh; }
  .exam-header { padding: 20px 15px; }
  .exam-header h1 { font-size: 20px; }
  .question-container { padding: 20px 15px; }
  .question-title { font-size: 17px; }
  .button-container { padding: 15px; flex-wrap: wrap; gap: 10px; }
  .btn { padding: 14px 12px; font-size: 14px; flex: 1; justify-content: center; }
  .result-container { padding: 30px 15px; }
  .result-score { font-size: 45px; }
  .result-details { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .answer-sheet-toggle { right: 0; top: auto; bottom: 100px; transform: none; border-radius: 15px 0 0 15px; }
  .answer-sheet-panel { position: fixed; bottom: 0; left: 0; right: 0; max-height: 60vh; background: white; border-radius: 18px 18px 0 0; box-shadow: 0 -5px 25px rgba(0,0,0,0.2); }
  .answer-grid { width: 100%; }
  .answer-cell { width: 100%; aspect-ratio: 1; }
}
</style>
