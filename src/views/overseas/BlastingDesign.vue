<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import GlassCard from '@/components/GlassCard.vue'

const downloadUrl = '/downloads/blasting-design-setup.exe'
const downloadName = '爆破设计软件安装包'
const fileSize = '156 MB'
const version = 'V2.3.1'
const updateDate = '2026-06-15'

const features = [
  { icon: 'map-pin', title: '爆破方案设计', desc: '支持多孔、多排、复杂地形的爆破方案参数设计' },
  { icon: 'bar-chart-2', title: '爆破参数计算', desc: '自动计算装药量、孔间距、延时等关键参数' },
  { icon: 'eye', title: '三维可视化', desc: '3D 展示爆破区域地形、孔位分布及爆破效果模拟' },
  { icon: 'file-text', title: '报告自动生成', desc: '一键导出爆破设计方案报告，符合行业规范格式' },
  { icon: 'database', title: '数据管理', desc: '方案库管理，支持历史方案复用和参数对比' },
  { icon: 'shield', title: '安全校核', desc: '内置安全距离、振动阈值等校核规则' }
]

function handleDownload() {
  const a = document.createElement('a')
  a.href = downloadUrl
  a.download = downloadName
  a.click()
  ElMessage.success('正在下载 ' + downloadName)
}
</script>

<template>
  <div class="blasting-design">
    <GlassCard title="爆破设计软件">
      <div class="hero">
        <div class="hero-icon">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
          </svg>
        </div>
        <div class="hero-text">
          <h2>爆破设计软件</h2>
          <p class="subtitle">专业爆破方案设计工具，支持参数计算、三维可视化、方案管理</p>
        </div>
        <div class="version-badge">最新版 {{ version }}</div>
      </div>
    </GlassCard>

    <div class="content-grid">
      <!-- 下载区域 -->
      <div class="download-section">
        <GlassCard title="软件下载">
          <div class="download-card">
            <div class="download-info">
              <div class="download-icon">
                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                </svg>
              </div>
              <div class="download-meta">
                <h4>{{ downloadName }}</h4>
                <p>
                  <span class="meta-tag">版本 {{ version }}</span>
                  <span class="meta-tag">{{ fileSize }}</span>
                  <span class="meta-tag">更新于 {{ updateDate }}</span>
                </p>
                <p class="sys-req">支持系统：Windows 10/11 64位</p>
              </div>
            </div>
            <button class="download-btn" @click="handleDownload">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
              下载安装包
            </button>
          </div>
        </GlassCard>
      </div>

      <!-- 功能特性 -->
      <GlassCard title="功能特性">
        <div class="features-grid">
          <div v-for="f in features" :key="f.icon" class="feature-item">
            <div class="feature-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="16" /><line x1="8" y1="12" x2="16" y2="12" />
              </svg>
            </div>
            <div>
              <h4>{{ f.title }}</h4>
              <p>{{ f.desc }}</p>
            </div>
          </div>
        </div>
      </GlassCard>

      <!-- 说明区域 -->
      <GlassCard title="软件说明">
        <div class="description">
          <section>
            <h4>软件简介</h4>
            <p>爆破设计软件是一款专业用于爆破工程方案设计的辅助工具，集爆破参数计算、三维地形可视化、方案设计与管理于一体，适用于矿山开采、隧道掘进、拆除爆破等各类爆破工程场景。</p>
          </section>
          <section>
            <h4>核心功能</h4>
            <ul>
              <li><strong>参数设计：</strong>支持孔径、孔深、间距、排距、装药量等关键参数的自定义与自动计算</li>
              <li><strong>三维可视化：</strong>基于地形数据生成 3D 爆破模型，直观展示孔位分布与爆破效果</li>
              <li><strong>方案库管理：</strong>保存历史方案，支持参数对比和方案复用</li>
              <li><strong>报告导出：</strong>一键生成符合行业标准的爆破设计方案报告</li>
              <li><strong>安全校核：</strong>内置爆破安全规范，自动校核振动、飞石等安全指标</li>
            </ul>
          </section>
          <section>
            <h4>适用场景</h4>
            <ul>
              <li>露天矿山爆破设计</li>
              <li>隧道及地下工程爆破</li>
              <li>拆除爆破方案设计</li>
              <li>爆破参数优化与仿真</li>
            </ul>
          </section>
        </div>
      </GlassCard>
    </div>
  </div>
</template>

<style scoped>
.blasting-design { max-width: 100%; }

.hero {
  display: flex; align-items: center; gap: 20px; flex-wrap: wrap;
}
.hero-icon {
  width: 64px; height: 64px; border-radius: 16px;
  background: linear-gradient(135deg, #F77234, #F6AB16);
  display: flex; align-items: center; justify-content: center;
  color: white; flex-shrink: 0;
}
.hero-text { flex: 1; min-width: 200px; }
.hero-text h2 { margin: 0 0 4px; font-size: 20px; color: var(--text-primary); }
.subtitle { margin: 0; font-size: 13px; color: var(--text-secondary); }
.version-badge {
  padding: 4px 14px; border-radius: 20px;
  background: rgba(22, 93, 255, 0.1); border: 1px solid rgba(22, 93, 255, 0.3);
  color: #165DFF; font-size: 13px; white-space: nowrap;
}

.content-grid { margin-top: 20px; display: flex; flex-direction: column; gap: 20px; }

/* 下载区域 */
.download-card {
  display: flex; justify-content: space-between; align-items: center;
  gap: 20px; flex-wrap: wrap;
}
.download-info {
  display: flex; align-items: center; gap: 16px; flex: 1;
}
.download-icon {
  width: 60px; height: 60px; border-radius: 12px;
  background: var(--bg-input); display: flex; align-items: center;
  justify-content: center; color: var(--text-secondary); flex-shrink: 0;
}
.download-meta h4 { margin: 0 0 6px; font-size: 15px; color: var(--text-primary); }
.meta-tag {
  display: inline-block; padding: 1px 8px; border-radius: 4px;
  background: var(--bg-input); color: var(--text-tertiary); font-size: 12px;
  margin-right: 8px; margin-bottom: 4px;
}
.sys-req { margin: 6px 0 0; font-size: 12px; color: var(--text-tertiary); }

.download-btn {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 28px; border-radius: 10px;
  background: linear-gradient(135deg, #165DFF, #0E42D2);
  border: none; color: white; font-size: 15px; font-weight: 500;
  cursor: pointer; transition: all 0.2s; white-space: nowrap;
}
.download-btn:hover {
  transform: translateY(-1px); box-shadow: 0 4px 16px rgba(22, 93, 255, 0.4);
}

/* 功能特性 */
.features-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px;
}
.feature-item {
  display: flex; gap: 12px; padding: 16px;
  background: var(--bg-input); border-radius: 10px;
}
.feature-icon {
  width: 40px; height: 40px; border-radius: 10px;
  background: rgba(247, 114, 52, 0.1); color: #F77234;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.feature-item h4 { margin: 0 0 4px; font-size: 14px; color: var(--text-primary); }
.feature-item p { margin: 0; font-size: 12px; color: var(--text-secondary); line-height: 1.5; }

/* 说明区域 */
.description section { margin-bottom: 20px; }
.description section:last-child { margin-bottom: 0; }
.description h4 { margin: 0 0 8px; font-size: 15px; color: var(--text-primary); }
.description p { margin: 0; font-size: 13px; color: var(--text-secondary); line-height: 1.8; }
.description ul { margin: 0; padding-left: 18px; }
.description li { font-size: 13px; color: var(--text-secondary); line-height: 1.8; margin-bottom: 4px; }
.description li strong { color: var(--text-primary); }
</style>
