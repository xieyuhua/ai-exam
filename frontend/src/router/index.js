import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Login',
    component: () => import('@/views/LoginView.vue')
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/AdminView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/student',
    name: 'StudentHome',
    component: () => import('@/views/StudentView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  },
  {
    path: '/student/exams',
    name: 'StudentExams',
    component: () => import('@/views/ExamListView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  },
  {
    path: '/student/votes',
    name: 'StudentVotes',
    component: () => import('@/views/VoteListView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  },
  {
    path: '/student/surveys',
    name: 'StudentSurveys',
    component: () => import('@/views/SurveyListView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  },
  {
    path: '/exam',
    name: 'Exam',
    component: () => import('@/views/ExamView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  },
  {
    path: '/vote',
    name: 'Vote',
    component: () => import('@/views/VoteView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  },
  {
    path: '/survey',
    name: 'Survey',
    component: () => import('@/views/SurveyView.vue'),
    meta: { requiresAuth: true, requiresStudent: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
