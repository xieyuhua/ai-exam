<template>
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-icon purple">📋</div>
      <div>
        <div class="stat-value">{{ stats.exams }}</div>
        <div class="stat-label">考试总数</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon blue">🎓</div>
      <div>
        <div class="stat-value">{{ stats.students }}</div>
        <div class="stat-label">员工人数</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon indigo">🗳️</div>
      <div>
        <div class="stat-value">{{ stats.votes }}</div>
        <div class="stat-label">投票总数</div>
        <div class="stat-sub">{{ stats.activeVotes || 0 }} 进行中</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon teal">📋</div>
      <div>
        <div class="stat-value">{{ stats.surveys }}</div>
        <div class="stat-label">问卷总数</div>
        <div class="stat-sub">{{ stats.activeSurveys || 0 }} 进行中</div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({ stats: { type: Object, default: () => ({ exams: 0, students: 0, votes: 0, surveys: 0, activeVotes: 0, activeSurveys: 0 }) } })
</script>

<style scoped>
.stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-bottom: 24px; }
.stat-card {
  background: var(--white); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 20px 22px;
  display: flex; align-items: center; gap: 16px;
  transition: all var(--transition);
}
.stat-card:hover { box-shadow: var(--shadow); transform: translateY(-2px); }
.stat-icon {
  width: 48px; height: 48px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; font-size: 22px; flex-shrink: 0;
}
.stat-icon.purple { background: #eef2ff; color: #6366f1; }
.stat-icon.green { background: #ecfdf5; color: #10b981; }
.stat-icon.blue { background: #eff6ff; color: #3b82f6; }
.stat-icon.orange { background: #fffbeb; color: #f59e0b; }
.stat-icon.indigo { background: #f5f3ff; color: #8b5cf6; }
.stat-icon.teal { background: #f0fdfa; color: #14b8a6; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text); line-height: 1; }
.stat-label { font-size: 13px; color: var(--text-secondary); margin-top: 4px; }
.stat-sub { font-size: 11px; color: #10b981; margin-top: 2px; font-weight: 500; }
</style>
