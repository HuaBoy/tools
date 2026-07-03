<script setup>import { ref, reactive, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import CryptoJS from 'crypto-js';
import { saveCredentials as saveCred, clearCredentials as clearCred, updateLoginStatus } from '@/utils/loginStatus.js';

const logsStore = useLogsStore();

const loginForm = reactive({
  tenantId: '000000',
  username: '',
  password: '',
  grant_type: 'password',
  type: 'account',
  loginIp: '114.224.178.71',
  'captcha-code': '',
  'captcha-key': '1397b4412c418709f10bb4ceacc8816b'
});

const queryForm = reactive({
  deviceCode: '',
  controllerCode: '',
  provinceCode: '',
  cityCode: '',
  current: 1,
  size: 10
});

const accessToken = ref('');
const accountInfo = ref(null);
const loginResult = ref(null);
const queryResult = ref(null);
const isQuerying = ref(false);
const isLogging = ref(false);
const selectedDevice = ref(null);

const showDeviceDetail = (device) => {
  selectedDevice.value = device;
};

const showSavedCredentials = ref(false);

const LOGIN_URL = 'https://mp.holyview.cn:9443/api/blade-auth/oauth/token';
const QUERY_URL = '/api/blade-detonate/blastDeviceFactory/page';

const TARGET_HOST = 'mp.holyview.cn:9443';
const TARGET_ORIGIN = 'https://mp.holyview.cn:9443';

const md5 = (str) => {
  return CryptoJS.MD5(str).toString();
};

const getAuthHeader = () => {
  const credentials = 'saber:saber_secret';
  return 'Basic ' + btoa(credentials);
};

const loadSavedCredentials = () => {
  try {
    const saved = localStorage.getItem('tester_credentials');
    if (saved) {
      const data = JSON.parse(saved);
      const now = Date.now();
      if (data.expireTime > now) {
        loginForm.username = data.username;
        loginForm.password = data.password;
        accessToken.value = data.accessToken;
        accountInfo.value = data.accountInfo;
        showSavedCredentials.value = true;

        // 同步登录状态到顶部菜单栏
        updateLoginStatus('mp', true);
        // 同时保存到统一凭据管理
        saveCred('mp', {
          tenantId: loginForm.tenantId,
          username: data.username,
          password: md5(data.password),
          accessToken: data.accessToken,
          tokenExpire: data.expireTime
        });

        ElMessage.success('已加载保存的登录信息');
        logsStore.addLog('加载', 'API测试', '自动加载保存的凭证');
      }
    }
  } catch (e) {
    console.error('加载凭证失败', e);
  }
};

const saveCredentials = async (username, password, token, account) => {
  try {
    const data = {
      username,
      password,
      accessToken: token,
      accountInfo: account,
      expireTime: Date.now() + 2 * 24 * 60 * 60 * 1000
    };
    localStorage.setItem('tester_credentials', JSON.stringify(data));
    logsStore.addLog('保存', 'API测试', '保存登录凭证');
  } catch (e) {
    console.error('保存凭证失败', e);
  }
};

const clearCredentials = () => {
  localStorage.removeItem('tester_credentials');
  accessToken.value = '';
  accountInfo.value = null;
  loginForm.password = '';
  showSavedCredentials.value = false;
  // 同步清除登录状态到顶部菜单栏
  updateLoginStatus('mp', false);
  // 清除统一凭据
  clearCred('mp');
  ElMessage.success('已清除保存的凭证');
  logsStore.addLog('清除', 'API测试', '清除登录凭证');
};

const handleLogin = async () => {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning('请输入用户名和密码');
    return;
  }

  isLogging.value = true;
  loginResult.value = null;

  try {
    const md5Password = md5(loginForm.password);
    
    const params = new URLSearchParams();
    params.append('tenantId', loginForm.tenantId);
    params.append('username', loginForm.username);
    params.append('password', md5Password);
    params.append('grant_type', loginForm.grant_type);
    params.append('type', loginForm.type);
    params.append('loginIp', loginForm.loginIp);
    params.append('scope', 'all');

    const fullUrl = `${LOGIN_URL}?${params.toString()}`;

    const response = await fetch(fullUrl, {
      method: 'POST',
      headers: {
        'Authorization': getAuthHeader(),
        'accept': 'application/json, text/plain, */*',
        'origin': 'https://mp.holyview.cn:9443',
        'referer': 'https://mp.holyview.cn:9443/',
        'tenant-id': loginForm.tenantId
      }
    });

    const data = await response.json();

    if (data.code === 200 && data.data && data.data.access_token) {
      accessToken.value = data.data.access_token;
      accountInfo.value = data.data;
      loginResult.value = { success: true, data };

      saveCredentials(loginForm.username, loginForm.password, data.data.access_token, data.data);
      showSavedCredentials.value = true;

      // 同步登录状态到顶部菜单栏
      updateLoginStatus('mp', true);
      // 保存到统一凭据管理（供401自动重新登录使用）
      saveCred('mp', {
        tenantId: loginForm.tenantId,
        username: loginForm.username,
        password: md5Password,
        accessToken: data.data.access_token,
        tokenExpire: Date.now() + 2 * 24 * 60 * 60 * 1000
      });

      ElMessage.success('登录成功');
      logsStore.addLog('登录', 'API测试', `用户: ${loginForm.username}`);
    } else {
      loginResult.value = { success: false, status: response.status, error: data.message || data.msg || '登录失败', detail: data };
      ElMessage.error(`登录失败 [${response.status}]: ${data.message || data.msg || '未知错误'}`);
      logsStore.addLog('登录', 'API测试', `失败 [${response.status}]: ${data.message || data.msg}`);
    }
  } catch (error) {
    loginResult.value = { success: false, error: error.message };
    ElMessage.error('网络连接失败: ' + error.message);
    logsStore.addLog('登录', 'API测试', `网络错误: ${error.message}`);
  } finally {
    isLogging.value = false;
  }
};

const handleQuery = async () => {
  if (!accessToken.value) {
    ElMessage.warning('请先登录获取Token');
    return;
  }

  isQuerying.value = true;
  queryResult.value = null;

  try {
    const params = new URLSearchParams();
    if (queryForm.deviceCode) params.append('deviceCode', queryForm.deviceCode);
    if (queryForm.controllerCode) params.append('controllerCode', queryForm.controllerCode);
    if (queryForm.provinceCode) params.append('provinceCode', queryForm.provinceCode);
    if (queryForm.cityCode) params.append('cityCode', queryForm.cityCode);
    params.append('current', queryForm.current);
    params.append('size', queryForm.size);

    const response = await fetch(`${QUERY_URL}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'authorization': 'Basic c2FiZXI6c2FiZXJfc2VjcmV0',
        'blade-auth': `bearer ${accessToken.value}`,
        'tenant-id': '000000'
      }
    });

    const data = await response.json();

    if (response.ok) {
      queryResult.value = { success: true, data };
      ElMessage.success(`查询成功，共 ${data.total || 0} 条记录`);
      logsStore.addLog('查询', 'API测试', `设备查询: ${queryForm.deviceCode || queryForm.controllerCode}`);
    } else {
      queryResult.value = { success: false, error: data.message || '查询失败' };
      if (data.code === 401) {
        ElMessage.error('Token已过期，请重新登录');
        clearCredentials();
      } else {
        ElMessage.error(data.message || '查询失败');
      }
      logsStore.addLog('查询', 'API测试', `失败: ${data.message}`);
    }
  } catch (error) {
    queryResult.value = { success: false, error: error.message };
    ElMessage.error('网络连接失败');
    logsStore.addLog('查询', 'API测试', `网络错误: ${error.message}`);
  } finally {
    isQuerying.value = false;
  }
};

const handleQueryBoth = async () => {
  if (!accessToken.value) {
    ElMessage.warning('请先登录获取Token');
    return;
  }

  isQuerying.value = true;
  queryResult.value = null;

  const searchValue = queryForm.deviceCode || queryForm.controllerCode;
  if (!searchValue) {
    ElMessage.warning('请输入SN编号或控制器编号');
    isQuerying.value = false;
    return;
  }

  try {
    const params = new URLSearchParams();
    params.append('deviceCode', searchValue);
    params.append('controllerCode', searchValue);
    params.append('current', queryForm.current || 1);
    params.append('size', queryForm.size || 10);
    if (queryForm.provinceCode) params.append('provinceCode', queryForm.provinceCode);
    if (queryForm.cityCode) params.append('cityCode', queryForm.cityCode);

    const response = await fetch(`${QUERY_URL}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json, text/plain, */*',
        'authorization': 'Basic c2FiZXI6c2FiZXJfc2VjcmV0',
        'blade-auth': `bearer ${accessToken.value}`,
        'tenant-id': '000000'
      }
    });
    
    const result = await response.json();
    
    if (result.code === 200 && result.success && result.data && result.data.records) {
      queryResult.value = { 
        success: true, 
        data: result.data,
        records: result.data.records,
        total: result.data.total,
        current: result.data.current,
        size: result.data.size,
        pages: result.data.pages
      };
      ElMessage.success(`查询成功，共 ${result.data.total} 条记录`);
      logsStore.addLog('查询', 'API测试', `查询成功: ${searchValue}, 共${result.data.total}条`);
    } else {
      queryResult.value = { success: false, error: result.msg || '未查询到数据' };
      ElMessage.warning(result.msg || '未查询到数据');
    }
  } catch (error) {
    queryResult.value = { success: false, error: error.message };
    ElMessage.error('网络连接失败');
    logsStore.addLog('查询', 'API测试', `网络错误: ${error.message}`);
  } finally {
    isQuerying.value = false;
  }
};

onMounted(() => {
  loadSavedCredentials();
});
</script>

<template>
  <div class="ai-tester">
    <GlassCard title="云系统">
      <div class="section-tabs">
        <button 
          class="tab-btn active"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
          <span>登录测试</span>
        </button>
        <button 
          class="tab-btn"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 20h9" />
            <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7.5 19.5a2.121 2.121 0 0 1-3-3z" />
          </svg>
          <span>设备查询</span>
        </button>
      </div>
    </GlassCard>

    <GlassCard title="设备查询" style="margin-top: 20px;">
      <div class="form-grid">
        <div class="form-group">
          <label>SN编号</label>
          <input 
            v-model="queryForm.deviceCode"
            type="text"
            class="form-input"
            placeholder="设备SN编号"
          />
        </div>
        <div class="form-group">
          <label>控制器编号</label>
          <input 
            v-model="queryForm.controllerCode"
            type="text"
            class="form-input"
            placeholder="控制器编号"
          />
        </div>
        <div class="form-group">
          <label>省份代码</label>
          <input 
            v-model="queryForm.provinceCode"
            type="text"
            class="form-input"
            placeholder="provinceCode"
          />
        </div>
        <div class="form-group">
          <label>城市代码</label>
          <input 
            v-model="queryForm.cityCode"
            type="text"
            class="form-input"
            placeholder="cityCode"
          />
        </div>
        <div class="form-group">
          <label>页码</label>
          <input 
            v-model.number="queryForm.current"
            type="number"
            class="form-input"
            placeholder="current"
          />
        </div>
        <div class="form-group">
          <label>每页数量</label>
          <input 
            v-model.number="queryForm.size"
            type="number"
            class="form-input"
            placeholder="size"
          />
        </div>
      </div>

      <div class="form-actions">
        <button 
          class="action-btn primary"
          :disabled="isQuerying || !accessToken"
          @click="handleQuery"
        >
          <svg v-if="!isQuerying" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="spin">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          <span>{{ isQuerying ? '查询中...' : '查询设备' }}</span>
        </button>
        <button 
          class="action-btn secondary"
          :disabled="isQuerying || !accessToken"
          @click="handleQueryBoth"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>双条件查询</span>
        </button>
      </div>

      <div v-if="queryResult" class="result-panel" :class="queryResult.success ? 'success' : 'error'">
        <div class="result-header">
          <svg v-if="queryResult.success" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
          <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
          <span>{{ queryResult.success ? '查询成功' : '查询失败' }}</span>
        </div>
        <pre v-if="!queryResult.success" class="result-content">{{ JSON.stringify(queryResult.data, null, 2) }}</pre>
      </div>

      <div v-if="queryResult?.success && queryResult.records" class="records-panel">
        <div class="records-header">
          <span>设备列表</span>
          <span class="pagination-info">共 {{ queryResult.total }} 条 | 第 {{ queryResult.current }} 页 | 每页 {{ queryResult.size }} 条</span>
        </div>
        
        <div class="table-container">
          <table class="device-table">
            <thead>
              <tr>
                <th>序号</th>
                <th>SN编号</th>
                <th>控制器编号</th>
                <th>创建时间</th>
                <th>更新时间</th>
                <th>更新账号</th>
                <th>设备状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(record, index) in queryResult.records" :key="record.id">
                <td>{{ (queryResult.current - 1) * queryResult.size + index + 1 }}</td>
                <td class="device-code">{{ record.deviceCode || '-' }}</td>
                <td class="controller-code">{{ record.controllerCode || '-' }}</td>
                <td>{{ record.createTime || '-' }}</td>
                <td>{{ record.updateTime || '-' }}</td>
                <td>{{ record.updateUserName || record.updateUser || '-' }}</td>
                <td>
                  <span class="status-tag" :class="record.status === 1 ? 'status-active' : 'status-inactive'">
                    {{ record.status === 1 ? '启用' : '禁用' }}
                  </span>
                </td>
                <td>
                  <button class="detail-btn" @click="showDeviceDetail(record)">详情</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <div v-if="selectedDevice" class="device-detail-panel">
          <div class="detail-header">
            <span>设备详情 - {{ selectedDevice.deviceCode }}</span>
            <button class="close-btn" @click="selectedDevice = null">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
          <div class="detail-content">
            <div class="detail-row">
              <span class="detail-label">ID:</span>
              <span class="detail-value">{{ selectedDevice.id }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">SN编号:</span>
              <span class="detail-value">{{ selectedDevice.deviceCode }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">设备版本:</span>
              <span class="detail-value">{{ selectedDevice.deviceVersion }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">控制器编号:</span>
              <span class="detail-value">{{ selectedDevice.controllerCode }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">控制器版本:</span>
              <span class="detail-value">{{ selectedDevice.controllerVersion }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">控制器类型:</span>
              <span class="detail-value">{{ selectedDevice.controllerTypeName }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">厂商:</span>
              <span class="detail-value">{{ selectedDevice.factoryName }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">创建时间:</span>
              <span class="detail-value">{{ selectedDevice.createTime }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">更新时间:</span>
              <span class="detail-value">{{ selectedDevice.updateTime }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">更新账号:</span>
              <span class="detail-value">{{ selectedDevice.updateUserName || selectedDevice.updateUser }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">状态:</span>
              <span class="detail-value">
                <span class="status-tag" :class="selectedDevice.status === 1 ? 'status-active' : 'status-inactive'">
                  {{ selectedDevice.status === 1 ? '启用' : '禁用' }}
                </span>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">电子围栏:</span>
              <span class="detail-value">{{ selectedDevice.electricFance ? '已设置' : '未设置' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">芯片厂商:</span>
              <span class="detail-value">{{ selectedDevice.chipManufacturer }}</span>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.ai-tester {
  max-width: 100%;
}

.section-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  
  &.active {
    background: rgba(22, 93, 255, 0.1);
    border-color: #165DFF;
    color: #165DFF;
  }
  
  &:hover:not(.active) {
    background: rgba(22, 93, 255, 0.05);
  }
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.form-group {
  label {
    display: block;
    font-size: 12px;
    color: var(--text-tertiary);
    margin-bottom: 8px;
  }
}

.form-input {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px 12px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
  
  &:focus {
    border-color: #165DFF;
  }
  
  &::placeholder {
    color: var(--text-tertiary);
  }
}

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  
  &.primary {
    background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
    color: #FFFFFF;
    
    &:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
    }
  }
  
  &.secondary {
    background: rgba(100, 116, 139, 0.2);
    border: 1px solid rgba(100, 116, 139, 0.3);
    color: #94A3B8;
    
    &:hover:not(:disabled) {
      background: rgba(100, 116, 139, 0.3);
    }
  }
  
  &.danger {
    background: rgba(245, 63, 63, 0.1);
    border: 1px solid rgba(245, 63, 63, 0.3);
    color: #F53F3F;
    
    &:hover:not(:disabled) {
      background: rgba(245, 63, 63, 0.2);
    }
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.auth-info {
  margin-top: 20px;
  padding: 16px;
  background: rgba(0, 180, 42, 0.1);
  border-radius: 8px;
}

.info-title {
  font-size: 13px;
  font-weight: 600;
  color: #00B42A;
  margin-bottom: 12px;
}

.info-item {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  
  &:last-child {
    margin-bottom: 0;
  }
}

.info-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.info-value {
  font-size: 12px;
  color: var(--text-primary);
  
  &.token-status.valid {
    color: #00B42A;
  }
}

.result-panel {
  margin-top: 20px;
  padding: 16px;
  border-radius: 8px;
  
  &.success {
    background: rgba(0, 180, 42, 0.1);
    border: 1px solid rgba(0, 180, 42, 0.3);
  }
  
  &.error {
    background: rgba(245, 63, 63, 0.1);
    border: 1px solid rgba(245, 63, 63, 0.3);
  }
}

.result-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  
  svg {
    flex-shrink: 0;
  }
  
  span {
    font-size: 14px;
    font-weight: 600;
    
    .success & {
      color: #00B42A;
    }
    
    .error & {
      color: #F53F3F;
    }
  }
}

.result-content {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
}

.records-panel {
  margin-top: 20px;
}

.records-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.pagination-info {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: normal;
}

.table-container {
  overflow-x: auto;
  background: var(--bg-input);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.device-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.device-table thead {
  background: rgba(22, 93, 255, 0.1);
}

.device-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
}

.device-table td {
  padding: 10px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
}

.device-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.05);
}

.device-code {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 500;
  color: #165DFF;
}

.controller-code {
  font-family: 'Consolas', 'Monaco', monospace;
}

.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-active {
  background: rgba(0, 180, 42, 0.2);
  color: #00B42A;
}

.status-inactive {
  background: rgba(245, 63, 63, 0.2);
  color: #F53F3F;
}

.detail-btn {
  padding: 4px 12px;
  background: rgba(22, 93, 255, 0.1);
  border: 1px solid rgba(22, 93, 255, 0.3);
  border-radius: 4px;
  color: #165DFF;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.detail-btn:hover {
  background: rgba(22, 93, 255, 0.2);
}

.device-detail-panel {
  margin-top: 20px;
  background: var(--bg-input);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(22, 93, 255, 0.1);
  border-bottom: 1px solid var(--border-color);
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 4px;
}

.close-btn:hover {
  color: var(--text-primary);
}

.detail-content {
  padding: 16px;
}

.detail-row {
  display: flex;
  gap: 12px;
  margin-bottom: 10px;
}

.detail-label {
  min-width: 100px;
  font-size: 13px;
  color: var(--text-tertiary);
  font-weight: 500;
}

.detail-value {
  font-size: 13px;
  color: var(--text-primary);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media screen and (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .action-btn {
    width: 100%;
    justify-content: center;
  }
  
  .device-table {
    font-size: 12px;
  }
  
  .device-table th,
  .device-table td {
    padding: 8px 10px;
  }
  
  .detail-row {
    flex-direction: column;
    gap: 4px;
  }
  
  .detail-label {
    min-width: auto;
  }
}

@media screen and (max-width: 480px) {
  .tab-btn {
    padding: 8px 12px;
    font-size: 12px;
  }
  
  .form-input {
    font-size: 12px;
    padding: 8px 10px;
  }
}
</style>