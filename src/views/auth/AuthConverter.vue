<script setup>
import { ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';

const logsStore = useLogsStore();
const rawDeviceCode = ref('');
const authCode = ref('');
const historyRecords = ref([
  { id: 1, deviceCode: 'A910', authCode: '793179', expireDate: '2025-12-31', createTime: '2024-01-15 10:30' },
  { id: 2, deviceCode: 'B234', authCode: '323751', expireDate: '2025-12-31', createTime: '2024-01-15 11:20' },
  { id: 3, deviceCode: 'CDEF', authCode: '918479', expireDate: '2025-12-31', createTime: '2024-01-16 09:15' }
]);

const letterToNumMap = {
  'A': '7', 'B': '3', 'C': '9', 'D': '1', 'E': '8', 'F': '4', 'G': '6',
  'H': '2', 'I': '5', 'J': '0',
  'K': '2', 'L': '4', 'M': '6', 'N': '8', 'O': '1', 'P': '3', 'Q': '5',
  'R': '7', 'S': '9', 'T': '0',
  'U': '2', 'V': '4', 'W': '6', 'X': '8', 'Y': '0', 'Z': '1'
};

const numToLetterMap = {
  '0': 'K', '1': 'L', '2': 'M', '3': 'N', '4': 'O', '5': 'P', '6': 'Q',
  '7': 'R', '8': 'S', '9': 'T'
};

const weights = [13, 7, 19, 11];

const generateAuthCode = (deviceCode) => {
  const code = deviceCode.toUpperCase().trim();

  if (code.length !== 4) {
    return null;
  }

  for (const char of code) {
    if (!/^[A-Z0-9]$/.test(char)) {
      return null;
    }
  }

  const source = code.split('');

  const mapped = source.map(char => {
    if (/^[A-Z]$/.test(char)) {
      return letterToNumMap[char];
    } else {
      return numToLetterMap[char];
    }
  });

  const weightedSum = source.reduce((sum, char, index) => {
    return sum + char.charCodeAt(0) * weights[index];
  }, 0);

  const offset = ((source[0].charCodeAt(0) + source[1].charCodeAt(0) * 2) % 26) + 1;
  const sumMod26 = weightedSum % 26;

  const pos1 = mapped[0];
  const pos2 = mapped[1];
  const pos3 = String(weightedSum % 10);
  const pos4 = String(offset % 10);
  const pos5 = String.fromCharCode(((mapped[2].charCodeAt(0) + offset) % 26) + 65);
  const pos6 = String.fromCharCode(((mapped[3].charCodeAt(0) + sumMod26) % 26) + 65);

  const intermediate = pos1 + pos2 + pos3 + pos4 + pos5 + pos6;

  const finalCode = intermediate.split('').map(char => {
    if (/^[A-Z]$/.test(char)) {
      return letterToNumMap[char];
    }
    return char;
  }).join('');

  return finalCode;
};

const handleConvert = () => {
  if (!rawDeviceCode.value.trim()) {
    ElMessage.warning('请输入原始设备码');
    return;
  }

  const deviceCodes = rawDeviceCode.value.trim().split('\n').filter(code => code.trim());

  if (deviceCodes.length === 0) {
    ElMessage.warning('请输入至少一个设备码');
    return;
  }

  const results = [];
  const errors = [];

  deviceCodes.forEach((code, index) => {
    const trimmedCode = code.trim();
    if (trimmedCode.length !== 4) {
      errors.push(`第${index + 1}行: ${trimmedCode} (长度不为4位)`);
      return;
    }

    const result = generateAuthCode(trimmedCode);

    if (result === null) {
      errors.push(`第${index + 1}行: ${trimmedCode} (格式不正确)`);
    } else {
      results.push({ deviceCode: trimmedCode.toUpperCase(), authCode: result });
    }
  });

  if (errors.length > 0) {
    ElMessage.warning(`部分设备码转换失败:\n${errors.join('\n')}`);
  }

  if (results.length > 0) {
    authCode.value = results.map(r => `${r.deviceCode} -> ${r.authCode}`).join('\n');

    const now = new Date();
    const currentDate = now.toISOString().split('T')[0];
    const currentTime = now.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });

    results.forEach(result => {
      historyRecords.value.unshift({
        id: Date.now() + Math.random(),
        deviceCode: result.deviceCode,
        authCode: result.authCode,
        expireDate: '2025-12-31',
        createTime: `${currentDate} ${currentTime}`
      });
    });

    logsStore.addLog('转换', '授权管理', `批量转换 ${results.length} 个设备码`);
    ElMessage.success(`成功转换 ${results.length} 个授权码`);
  }
};

const handleCopy = () => {
  if (!authCode.value) {
    ElMessage.warning('没有可复制的授权码');
    return;
  }
  navigator.clipboard.writeText(authCode.value).then(() => {
    ElMessage.success('授权码已复制到剪贴板');
    logsStore.addLog('复制', '授权管理', '复制授权码');
  });
};

const handleClear = () => {
  rawDeviceCode.value = '';
  authCode.value = '';
  logsStore.addLog('清空', '授权管理', '清空输入内容');
};

const handleBatchExport = () => {
  if (historyRecords.value.length === 0) {
    ElMessage.warning('没有可导出的记录');
    return;
  }

  const headers = ['设备码', '授权码', '授权有效期', '创建时间'];
  const content = [headers.join(',')].concat(
    historyRecords.value.map(item => {
      const expireDate = item.expireDate || '2025-12-31';
      return `${item.deviceCode},${item.authCode},${expireDate},${item.createTime}`;
    })
  ).join('\n');

  const blob = new Blob(['\uFEFF' + content], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  const now = new Date();
  const dateStr = now.toISOString().split('T')[0];
  a.download = `auth_records_${dateStr}.csv`;
  a.click();
  URL.revokeObjectURL(url);
  logsStore.addLog('导出', '授权管理', '批量导出授权记录');
  ElMessage.success('批量导出成功');
};

const handleClearHistory = () => {
  if (historyRecords.value.length === 0) {
    ElMessage.warning('暂无历史记录');
    return;
  }
  ElMessageBox.confirm('确定要清空所有历史记录吗？', '清空确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    historyRecords.value = [];
    ElMessage.success('历史记录已清空');
  }).catch(() => {});
};
</script>

<template>
  <div class="auth-converter">
    <GlassCard title="授权码转换工具">
      <!-- 授权码转换工具区 -->
      <section class="section">
        <div class="section-header">
          <div class="section-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1 7.778 7.778L12 17.778l-7.778 7.778a5.5 5.5 0 0 1 7.778-7.778L12 17.778l-7.778 7.778a5.5 5.5 0 0 1 7.778-7.778" />
            </svg>
            <span>授权码转换</span>
          </div>
        </div>

        <div class="converter-form">
          <div class="form-row">
            <div class="form-item">
              <label class="form-label">原始设备码</label>
              <textarea
                v-model="rawDeviceCode"
                class="code-input"
                placeholder="请输入设备码，每行一个&#10;例如：&#10;A910&#10;B234&#10;CDEF&#10;1234"
                rows="4"
              ></textarea>
              <p class="input-hint">支持批量输入，每行一个4位设备码（字母A-Z和数字0-9）</p>
            </div>

            <div class="form-item">
              <label class="form-label">授权码输出</label>
              <textarea
                v-model="authCode"
                class="code-input output"
                placeholder="授权码将在此显示..."
                rows="4"
                readonly
              ></textarea>
              <p class="input-hint">6位纯数字授权码</p>
            </div>
          </div>

          <div class="form-actions">
            <button class="convert-btn" @click="handleConvert">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="23 7 13.5 16.5 8.5 11.5 1 18" />
                <polyline points="17 7 23 7 23 13" />
              </svg>
              <span>一键转换</span>
            </button>
            <button class="action-btn primary" @click="handleCopy">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
              <span>复制结果</span>
            </button>
            <button class="action-btn secondary" @click="handleClear">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 6h18" />
                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
              </svg>
              <span>清空内容</span>
            </button>
          </div>
        </div>
      </section>

      <div class="section-divider"></div>

      <!-- 历史转换记录区 -->
      <section class="section">
        <div class="section-header">
          <div class="section-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
              <path d="M3 3v5h5" />
              <path d="M12 7v5l4 2" />
            </svg>
            <span>历史转换记录</span>
            <span class="record-count">共 {{ historyRecords.length }} 条</span>
          </div>
          <div class="section-actions">
            <button class="action-btn primary" @click="handleBatchExport" :disabled="historyRecords.length === 0">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
              <span>全部导出</span>
            </button>
            <button class="action-btn danger" @click="handleClearHistory" :disabled="historyRecords.length === 0">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 6h18" />
                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
              </svg>
              <span>清空</span>
            </button>
          </div>
        </div>

        <div class="tasks-table">
          <table>
            <thead>
              <tr>
                <th>设备码</th>
                <th>授权码</th>
                <th>授权有效期</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in historyRecords" :key="record.id">
                <td>{{ record.deviceCode }}</td>
                <td>{{ record.authCode }}</td>
                <td>{{ record.expireDate }}</td>
                <td>{{ record.createTime }}</td>
              </tr>
              <tr v-if="historyRecords.length === 0">
                <td colspan="4" class="empty-row">暂无历史记录</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </GlassCard>
  </div>
</template>

<style scoped>
.auth-converter {
  max-width: 100%;
  width: 100%;
}

.section {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.section-title svg {
  color: #165DFF;
}

.section-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.record-count {
  display: inline-block;
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 10px;
  font-weight: 500;
  margin-left: 4px;
}

.section-divider {
  height: 1px;
  background: var(--border-color);
  margin: 28px 0;
}

.converter-form {
  padding: 16px;
  background: var(--bg-input);
  border-radius: 8px;
}

.form-row {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  width: 100%;
}

.form-item {
  flex: 1 1 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.form-label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
  font-weight: 500;
}

.input-hint {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 6px;
  line-height: 1.4;
}

.code-input {
  width: 100%;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px;
  color: var(--text-primary);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 2px;
  resize: vertical;
  outline: none;
  transition: border-color 0.2s;
  min-height: 100px;
  box-sizing: border-box;
  line-height: 1.6;
}

.code-input:focus {
  border-color: rgba(22, 93, 255, 0.6);
}

.code-input.output {
  background: rgba(0, 180, 42, 0.06);
  border-color: rgba(0, 180, 42, 0.3);
}

.code-input::placeholder {
  color: var(--text-tertiary);
  letter-spacing: normal;
  font-weight: normal;
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
  flex-wrap: wrap;
}

.convert-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 24px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 8px;
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.convert-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  min-width: 100px;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.primary {
  background: rgba(0, 180, 42, 0.1);
  border: 1px solid rgba(0, 180, 42, 0.3);
  color: #00B42A;
}

.action-btn.primary:hover:not(:disabled) {
  background: rgba(0, 180, 42, 0.2);
}

.action-btn.secondary {
  background: rgba(100, 116, 139, 0.1);
  border: 1px solid rgba(100, 116, 139, 0.3);
  color: #64748B;
}

.action-btn.secondary:hover:not(:disabled) {
  background: rgba(100, 116, 139, 0.2);
}

.action-btn.danger {
  background: rgba(245, 63, 63, 0.1);
  border: 1px solid rgba(245, 63, 63, 0.3);
  color: #F53F3F;
}

.action-btn.danger:hover:not(:disabled) {
  background: rgba(245, 63, 63, 0.2);
}

.tasks-table {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.tasks-table table {
  width: 100%;
  border-collapse: collapse;
}

.tasks-table th,
.tasks-table td {
  padding: 12px 16px;
  text-align: left;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
}

.tasks-table th {
  color: var(--text-tertiary);
  font-weight: 500;
  background: rgba(22, 93, 255, 0.04);
}

.tasks-table td {
  color: var(--text-secondary);
}

.tasks-table tbody tr:last-child td {
  border-bottom: none;
}

.tasks-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.05);
}

.empty-row {
  text-align: center;
  color: var(--text-tertiary);
  padding: 40px !important;
}

/* 平板 (768px以下) */
@media screen and (max-width: 768px) {
  .form-row {
    flex-direction: column;
    gap: 16px;
  }

  .form-item {
    min-width: 100%;
  }

  .form-actions {
    flex-direction: column;
  }

  .convert-btn,
  .action-btn {
    width: 100%;
    justify-content: center;
  }

  .section-header {
    flex-direction: column;
    align-items: stretch;
  }

  .section-actions {
    width: 100%;
  }

  .section-actions .action-btn {
    flex: 1;
  }

  .tasks-table th,
  .tasks-table td {
    padding: 10px 12px;
    font-size: 12px;
  }
}

/* 手机 (480px以下) */
@media screen and (max-width: 480px) {
  .converter-form {
    padding: 12px;
  }

  .code-input {
    min-height: 90px;
    padding: 10px;
    font-size: 14px;
    letter-spacing: 1px;
  }

  .convert-btn,
  .action-btn {
    font-size: 13px;
    padding: 10px 16px;
  }

  .form-label {
    font-size: 12px;
  }

  .input-hint {
    font-size: 10px;
  }

  .section-title {
    font-size: 14px;
  }

  .tasks-table th,
  .tasks-table td {
    padding: 8px 10px;
    font-size: 11px;
  }
}

/* 横屏手机/小平板优化 */
@media screen and (max-height: 500px) and (orientation: landscape) {
  .code-input {
    min-height: 70px;
  }
}
</style>
