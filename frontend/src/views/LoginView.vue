<template>
  <div class="login-page">
    <div class="container">
      <!-- Header -->
      <div class="header">
        <div class="logo">🏢</div>
        <h1>海斛集团</h1>
        <p class="subtitle">员工服务平台</p>
        <div :class="['mode-tag', currentRole === 'student' ? 'student' : 'admin']">
          {{ currentRole === 'student' ? '👤 员工登录' : '⚙️ 管理员登录' }}
        </div>
      </div>

      <!-- ========== 员工端 + 企微环境 ========== -->
      <template v-if="currentRole === 'student' && isWxWork && !showPwdForm">
        <div class="wxwork-badge show">
          <span class="dot"></span> 已检测到企业微信
        </div>
        <button class="btn btn-wxwork" :disabled="wxworkLoading" @click="wxWorkLogin">
          <span class="spinner" v-if="wxworkLoading"></span>
          <span>💬 企业微信授权登录</span>
        </button>
        <div class="wxwork-divider show"><span>其他登录方式</span></div>
        <div class="minor-entry show">
          <span @click="showPwdForm = true">🔑 使用工号密码登录</span>
        </div>
        <div class="admin-entry">
          <span @click="switchToAdmin">⚙ 管理员登录</span>
        </div>
      </template>

      <!-- ========== 员工端 + 企微环境切到工号 ========== -->
      <template v-else-if="currentRole === 'student' && isWxWork && showPwdForm">
        <form @submit.prevent="passwordLogin">
          <div class="form-group">
            <label for="workNo">账号</label>
            <div class="input-wrapper">
              <span class="input-icon">👤</span>
              <input type="text" id="workNo" v-model="workNo" placeholder="请输入账号" autocomplete="username" />
            </div>
          </div>
          <div class="form-group">
            <label for="password">密码</label>
            <div class="input-wrapper">
              <span class="input-icon">🔒</span>
              <input :type="showPwd ? 'text' : 'password'" id="password" v-model="password" placeholder="请输入密码" autocomplete="current-password" />
              <button type="button" class="pw-toggle" @click="showPwd = !showPwd">{{ showPwd ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <button class="btn btn-primary" :disabled="pwdLoading">
            <span class="spinner" v-if="pwdLoading"></span>
            <span>登 录</span>
          </button>
        </form>
        <div :class="['error-msg', { show: errorMsg }]">{{ errorMsg }}</div>
        <div class="back-entry" @click="showPwdForm = false">← 返回企业微信登录</div>
      </template>

      <!-- ========== 员工端 + 非企微环境 / 管理员端 ========== -->
      <template v-else>
        <form @submit.prevent="passwordLogin">
          <div class="form-group">
            <label for="workNo">账号</label>
            <div class="input-wrapper">
              <span class="input-icon">👤</span>
              <input type="text" id="workNo" v-model="workNo" placeholder="请输入账号" autocomplete="username" />
            </div>
          </div>
          <div class="form-group">
            <label for="password">密码</label>
            <div class="input-wrapper">
              <span class="input-icon">🔒</span>
              <input :type="showPwd ? 'text' : 'password'" id="password" v-model="password" placeholder="请输入密码" autocomplete="current-password" />
              <button type="button" class="pw-toggle" @click="showPwd = !showPwd">{{ showPwd ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <button class="btn btn-primary" :disabled="pwdLoading">
            <span class="spinner" v-if="pwdLoading"></span>
            <span>登 录</span>
          </button>
        </form>
        <div :class="['error-msg', { show: errorMsg }]">{{ errorMsg }}</div>

        <div class="admin-entry" v-if="currentRole === 'student'">
          <span @click="switchToAdmin">⚙ 管理员登录</span>
        </div>
        <div :class="['back-student', { show: currentRole === 'admin' }]" @click="switchToStudent">
          ← 返回员工登录
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api, setToken, checkAuth, currentUser } from '@/stores/auth'
import { showToast } from '@/components/toastState.js'

const router = useRouter()

const currentRole = ref('student')
const showPwd = ref(false)
const showPwdForm = ref(false)

const workNo = ref('')
const password = ref('')
const errorMsg = ref('')
const pwdLoading = ref(false)
const wxworkLoading = ref(false)

const isWxWork = computed(() => /wxwork/i.test(navigator.userAgent))

function switchToAdmin() {
  currentRole.value = 'admin'
  errorMsg.value = ''
  showPwdForm.value = false
}
function switchToStudent() {
  currentRole.value = 'student'
  errorMsg.value = ''
  showPwdForm.value = false
}

async function wxWorkLogin() {
  wxworkLoading.value = true
  try {
    const res = await api('/auth/wxwork/url?role=' + currentRole.value)
    if (res.code === 200 && res.data && res.data.authUrl) {
      window.location.href = res.data.authUrl
    } else {
      showToast(res.message || '获取授权链接失败', 'error')
      wxworkLoading.value = false
    }
  } catch (e) {
    showToast('网络错误，请稍后重试', 'error')
    wxworkLoading.value = false
  }
}

async function passwordLogin() {
  const wno = workNo.value.trim()
  const pwd = password.value.trim()

  if (!wno) { errorMsg.value = '请输入账号'; return }
  if (!pwd) { errorMsg.value = '请输入密码'; return }

  errorMsg.value = ''
  pwdLoading.value = true

  try {
    const res = await api('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ workNo: wno, password: pwd, role: currentRole.value })
    })

    if (res.code === 200 && res.data) {
      setToken(res.data.token)
      const isAdmin = res.data.user?.isAdmin
      if (currentRole.value === 'admin' && !isAdmin) {
        showToast('您的账号没有管理员权限，已切换为员工身份', 'error')
        setTimeout(() => router.push('/student'), 1500)
        return
      }
      showToast('登录成功', 'success')
      setTimeout(() => router.push(isAdmin ? '/admin' : '/student'), 500)
    } else {
      errorMsg.value = res.message || '账号或密码错误'
    }
  } catch (e) {
    errorMsg.value = '网络错误，请稍后重试'
  } finally {
    pwdLoading.value = false
  }
}

function checkUrlToken() {
  const params = new URLSearchParams(window.location.search)
  const token = params.get('token')
  const role = params.get('role') || 'student'
  if (token) {
    setToken(token)
    window.history.replaceState({}, document.title, window.location.pathname)
    router.push(role === 'admin' ? '/admin' : '/student')
  }
}

async function autoLogin() {
  const authed = await checkAuth()
  if (authed && currentUser.value) {
    router.push(currentUser.value.isAdmin ? '/admin' : '/student')
  }
}

onMounted(() => {
  checkUrlToken()
  autoLogin()
})
</script>

<style scoped>
.login-page {
  background: linear-gradient(135deg, #0f4c81 0%, #1a6fb5 50%, #2196f3 100%);
  min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 20px;
  position: relative; overflow: hidden;
}
.login-page::before, .login-page::after {
  content: ''; position: fixed; border-radius: 50%; opacity: 0.06; pointer-events: none;
}
.login-page::before { width: 500px; height: 500px; background: white; top: -120px; right: -120px; }
.login-page::after { width: 350px; height: 350px; background: white; bottom: -100px; left: -100px; }

.container {
  width: 420px; max-width: 100%; background: white;
  border-radius: 24px; padding: 44px 40px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
  position: relative; z-index: 1;
  animation: fadeInUp 0.5s ease;
}
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.header { text-align: center; margin-bottom: 28px; }
.header .logo {
  width: 64px; height: 64px; background: linear-gradient(135deg, #0f4c81, #2196f3);
  border-radius: 18px; display: inline-flex; align-items: center; justify-content: center;
  font-size: 32px; margin-bottom: 16px; box-shadow: 0 8px 24px rgba(15,76,129,0.3);
}
.header h1 { font-size: 24px; font-weight: 700; color: #0f4c81; }
.header .subtitle { font-size: 13px; color: #888; margin-top: 6px; }
.mode-tag {
  display: inline-block; margin-top: 12px; padding: 4px 16px;
  border-radius: 20px; font-size: 12px; font-weight: 600;
}
.mode-tag.student { background: #e3f2fd; color: #0f4c81; }
.mode-tag.admin { background: #fff3e0; color: #e65100; }

/* 企微徽章 */
.wxwork-badge { display: none; align-items: center; justify-content: center; gap: 6px; margin-bottom: 16px; font-size: 12px; color: #07c160; font-weight: 600; }
.wxwork-badge.show { display: flex; }
.wxwork-badge .dot { width: 8px; height: 8px; background: #07c160; border-radius: 50%; }

/* 表单 */
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; color: #555; margin-bottom: 6px; font-weight: 600; }
.input-wrapper { position: relative; }
.input-wrapper .input-icon { position: absolute; left: 14px; top: 50%; transform: translateY(-50%); font-size: 16px; color: #bbb; pointer-events: none; }
.input-wrapper input {
  width: 100%; padding: 12px 14px 12px 40px;
  border: 2px solid #e0e0e0; border-radius: 12px;
  font-size: 15px; transition: all 0.3s; outline: none;
}
.input-wrapper input:focus { border-color: #2196f3; box-shadow: 0 0 0 3px rgba(33,150,243,0.1); }
.input-wrapper .pw-toggle {
  position: absolute; right: 12px; top: 50%; transform: translateY(-50%);
  font-size: 14px; cursor: pointer; color: #bbb; background: none; border: none; padding: 4px;
}

.btn { width: 100%; padding: 14px; border: none; border-radius: 12px; font-size: 16px; font-weight: 700; cursor: pointer; transition: all 0.3s; display: flex; align-items: center; justify-content: center; gap: 8px; margin-top: 6px; }
.btn-primary { background: linear-gradient(135deg, #0f4c81 0%, #2196f3 100%); color: white; }
.btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(15,76,129,0.4); }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-wxwork { background: #07c160; color: white; }
.btn-wxwork:hover:not(:disabled) { background: #06ad56; transform: translateY(-2px); box-shadow: 0 8px 24px rgba(7,193,96,0.4); }

.spinner {
  display: inline-block; width: 18px; height: 18px;
  border: 2px solid rgba(255,255,255,0.3); border-top-color: white;
  border-radius: 50%; animation: spin 0.6s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.error-msg { color: #f44336; font-size: 13px; text-align: center; margin-top: 10px; opacity: 0; max-height: 0; overflow: hidden; transition: all 0.3s; }
.error-msg.show { opacity: 1; max-height: 40px; }

/* 企微分隔 + 小字入口 */
.wxwork-divider { display: none; align-items: center; margin: 20px 0; color: #bbb; font-size: 12px; }
.wxwork-divider::before, .wxwork-divider::after { content: ''; flex: 1; height: 1px; background: #e8e8e8; }
.wxwork-divider span { padding: 0 16px; }
.wxwork-divider.show { display: flex; }

.minor-entry { display: none; text-align: center; }
.minor-entry.show { display: block; }
.minor-entry span { font-size: 13px; color: #bbb; cursor: pointer; }
.minor-entry span:hover { color: #888; }

.back-entry { text-align: center; margin-top: 14px; font-size: 12px; color: #2196f3; cursor: pointer; }
.back-entry:hover { color: #0f4c81; }

.admin-entry { text-align: center; margin-top: 24px; padding-top: 20px; border-top: 1px solid #f0f0f0; }
.admin-entry span { font-size: 12px; color: #ccc; cursor: pointer; }
.admin-entry span:hover { color: #888; }
.back-student { display: none; text-align: center; margin-top: 14px; font-size: 12px; color: #bbb; cursor: pointer; }
.back-student.show { display: block; }
.back-student:hover { color: #2196f3; }

@media (max-width: 480px) {
  .container { padding: 32px 24px; border-radius: 20px; }
  .header .logo { width: 52px; height: 52px; font-size: 26px; border-radius: 14px; }
  .header h1 { font-size: 20px; }
}
</style>
