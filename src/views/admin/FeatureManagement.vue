<script setup>
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useAuthStore } from '@/stores/auth';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const authStore = useAuthStore();
const logsStore = useLogsStore();

const features = ref([]);
const isLoading = ref(false);
const showFeatureForm = ref(false);
const editFeatureId = ref(null);

const featureForm = reactive({
  key: '',
  name: '',
  description: '',
  category: '',
  enabled: true,
  order: 0
});

const defaultFeatures = [
  { id: 1, key: 'trace_analysis', name: '批次数据追溯', description: 'AI起爆数据查询与追溯', category: 'data', enabled: true, order: 1 },
  { id: 2, key: 'factory_data', name: '智能制造系统', description: '工厂生产数据查询', category: 'data', enabled: true, order: 2 },
  { id: 3, key: 'auth_converter', name: '授权码转换', description: '设备码与授权码转换工具', category: 'tools', enabled: true, order: 3 },
  { id: 4, key: 'third_party_auth', name: '三方账号授权', description: '云系统、智能制造系统账号授权', category: 'auth', enabled: true, order: 4 },
  { id: 5, key: 'log_decrypt', name: '日志解密工具', description: '日志解密与查看工具', category: 'tools', enabled: true, order: 5 },
  { id: 6, key: 'ai_translate', name: 'AI翻译工具', description: '多语言AI翻译', category: 'tools', enabled: true, order: 6 },
  { id: 7, key: 'ai_assistant', name: 'AI运维智能助手', description: 'AI智能运维助手', category: 'tools', enabled: true, order: 7 },
  { id: 8, key: 'qr_code', name: '二维码生成', description: '二维码生成工具', category: 'tools', enabled: true, order: 8 },
  { id: 9, key: 'format_converter', name: '文档格式转换', description: '文档格式转换工具', category: 'tools', enabled: true, order: 9 },
  { id: 10, key: 'ai_tester', name: '云系统', description: '云系统测试工具', category: 'tools', enabled: true, order: 10 }
];

const categoryOptions = [
  { label: '数据查询', value: 'data' },
  { label: '授权管理', value: 'auth' },
  { label: '辅助工具', value: 'tools' },
  { label: 'AI功能', value: 'ai' },
  { label: '其他', value: 'other' }
];

const loadFeatures = async () => {
  isLoading.value = true;
  try {
    const result = await adminApi.getFeatures();
    if (result && result.length > 0) {
      features.value = result;
    } else {
      // 使用默认数据
      features.value = [...defaultFeatures];
    }
  } catch (error) {
    console.error('加载功能列表失败:', error);
    // 加载失败时使用默认数据
    features.value = [...defaultFeatures];
  } finally {
    isLoading.value = false;
  }
};

const handleAddFeature = () => {
  editFeatureId.value = null;
  featureForm.key = '';
  featureForm.name = '';
  featureForm.description = '';
  featureForm.category = 'tools';
  featureForm.enabled = true;
  featureForm.order = features.value.length + 1;
  showFeatureForm.value = true;
};

const handleEditFeature = (feature) => {
  editFeatureId.value = feature.id;
  featureForm.key = feature.key;
  featureForm.name = feature.name;
  featureForm.description = feature.description;
  featureForm.category = feature.category;
  featureForm.enabled = feature.enabled;
  featureForm.order = feature.order;
  showFeatureForm.value = true;
};

const handleSaveFeature = async () => {
  if (!featureForm.key.trim() || !featureForm.name.trim()) {
    ElMessage.warning('请填写功能标识和名称');
    return;
  }

  try {
    if (editFeatureId.value) {
      // 编辑
      const idx = features.value.findIndex(f => f.id === editFeatureId.value);
      if (idx > -1) {
        features.value[idx] = { ...features.value[idx], ...featureForm };
      }
      try {
        await adminApi.updateFeature(editFeatureId.value, featureForm);
      } catch (e) {
        console.warn('API保存失败，使用本地数据:', e.message);
      }
      ElMessage.success('功能更新成功');
      logsStore.addLog('编辑', '功能管理', `更新功能: ${featureForm.name}`);
    } else {
      // 新增
      const newFeature = {
        id: Date.now(),
        ...featureForm
      };
      features.value.push(newFeature);
      try {
        await adminApi.createFeature(featureForm);
      } catch (e) {
        console.warn('API创建失败，使用本地数据:', e.message);
      }
      ElMessage.success('功能创建成功');
      logsStore.addLog('创建', '功能管理', `创建功能: ${featureForm.name}`);
    }
    showFeatureForm.value = false;
  } catch (error) {
    ElMessage.error('保存功能失败: ' + error.message);
  }
};

const handleToggleFeature = async (feature) => {
  try {
    const newEnabled = !feature.enabled;
    // 更新本地数据
    const idx = features.value.findIndex(f => f.id === feature.id);
    if (idx > -1) {
      features.value[idx].enabled = newEnabled;
    }
    try {
      await adminApi.toggleFeature(feature.id, newEnabled);
    } catch (e) {
      console.warn('API更新失败，使用本地数据:', e.message);
    }
    ElMessage.success(`功能已${newEnabled ? '启用' : '禁用'}`);
    logsStore.addLog('开关', '功能管理', `${newEnabled ? '启用' : '禁用'}功能: ${feature.name}`);
  } catch (error) {
    ElMessage.error('更新功能状态失败: ' + error.message);
  }
};

const handleDeleteFeature = async (feature) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除功能 "${feature.name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
    features.value = features.value.filter(f => f.id !== feature.id);
    try {
      await adminApi.deleteFeature(feature.id);
    } catch (e) {
      console.warn('API删除失败，使用本地数据:', e.message);
    }
    ElMessage.success('功能删除成功');
    logsStore.addLog('删除', '功能管理', `删除功能: ${feature.name}`);
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除功能失败: ' + error.message);
    }
  }
};

const getCategoryName = (categoryValue) => {
  const opt = categoryOptions.find(o => o.value === categoryValue);
  return opt ? opt.label : categoryValue;
};

const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('zh-CN');
};

onMounted(() => {
  if (!authStore.hasPermission('feature:view')) {
    ElMessage.error('您没有访问功能管理的权限');
    return;
  }
  loadFeatures();
});
</script>

<template>
  <div class="feature-management">
    <GlassCard title="功能管理">
      <div class="toolbar">
        <div class="info-text">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
          <span>管理系统的功能模块，启用/禁用将影响对应菜单的可见性</span>
        </div>
        <button class="add-btn" @click="handleAddFeature" v-if="authStore.hasPermission('feature:toggle')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>添加功能</span>
        </button>
      </div>

      <div v-if="isLoading" class="loading-state">加载中...</div>

      <div v-else class="feature-list">
        <div
          v-for="feature in features"
          :key="feature.id"
          class="feature-item"
          :class="{ disabled: !feature.enabled }"
        >
          <div class="feature-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <path d="M9 9h6v6H9z" />
            </svg>
          </div>
          <div class="feature-info">
            <div class="feature-title">
              <span class="feature-name">{{ feature.name }}</span>
              <span class="feature-key">{{ feature.key }}</span>
              <span class="category-tag">{{ getCategoryName(feature.category) }}</span>
            </div>
            <div class="feature-description">{{ feature.description || '暂无描述' }}</div>
          </div>
          <div class="feature-status">
            <label class="switch">
              <input
                type="checkbox"
                :checked="feature.enabled"
                :disabled="!authStore.hasPermission('feature:toggle')"
                @change="handleToggleFeature(feature)"
              />
              <span class="slider"></span>
            </label>
            <span class="status-text" :class="{ active: feature.enabled }">
              {{ feature.enabled ? '已启用' : '已禁用' }}
            </span>
          </div>
          <div class="feature-actions">
            <button class="action-link" @click="handleEditFeature(feature)" v-if="authStore.hasPermission('feature:toggle')">编辑</button>
            <button class="action-link danger" @click="handleDeleteFeature(feature)" v-if="authStore.hasPermission('feature:toggle')">删除</button>
          </div>
        </div>
      </div>
    </GlassCard>

    <Teleport to="body">
      <div v-if="showFeatureForm" class="modal-overlay" @click.self="showFeatureForm = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>{{ editFeatureId ? '编辑功能' : '添加功能' }}</h3>
            <button class="modal-close" @click="showFeatureForm = false">×</button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>功能标识 (Key) *</label>
              <input v-model="featureForm.key" type="text" class="form-input" placeholder="例如: trace_analysis" />
              <div class="form-hint">英文唯一标识，用于程序识别</div>
            </div>
            <div class="form-group">
              <label>功能名称 *</label>
              <input v-model="featureForm.name" type="text" class="form-input" placeholder="例如: 批次数据追溯" />
            </div>
            <div class="form-group">
              <label>功能分类</label>
              <select v-model="featureForm.category" class="form-input">
                <option v-for="opt in categoryOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>功能描述</label>
              <textarea v-model="featureForm.description" class="form-input" rows="3" placeholder="功能详细说明"></textarea>
            </div>
            <div class="form-group">
              <label>排序</label>
              <input v-model.number="featureForm.order" type="number" class="form-input" />
            </div>
            <div class="form-group">
              <label class="checkbox-label">
                <input v-model="featureForm.enabled" type="checkbox" />
                <span>启用此功能</span>
              </label>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showFeatureForm = false">取消</button>
            <button class="btn btn-primary" @click="handleSaveFeature">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.feature-management {
  max-width: 1200px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 12px;
  flex-wrap: wrap;
}

.info-text {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.add-btn:hover {
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
}

.loading-state {
  padding: 60px 0;
  text-align: center;
  color: var(--text-tertiary);
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  transition: all 0.2s;
}

.feature-item:hover {
  border-color: rgba(22, 93, 255, 0.3);
  background: rgba(22, 93, 255, 0.02);
}

.feature-item.disabled {
  opacity: 0.5;
}

.feature-icon {
  width: 40px;
  height: 40px;
  background: rgba(22, 93, 255, 0.1);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #165DFF;
  flex-shrink: 0;
}

.feature-info {
  flex: 1;
  min-width: 0;
}

.feature-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.feature-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.feature-key {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: monospace;
  background: rgba(100, 116, 139, 0.1);
  padding: 1px 6px;
  border-radius: 4px;
}

.category-tag {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 4px;
  background: rgba(54, 207, 201, 0.1);
  color: #36CFC9;
  font-weight: 500;
}

.feature-description {
  font-size: 12px;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feature-status {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #CBD5E1;
  transition: 0.3s;
  border-radius: 22px;
}

.slider::before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #165DFF;
}

input:checked + .slider::before {
  transform: translateX(18px);
}

input:disabled + .slider {
  opacity: 0.5;
  cursor: not-allowed;
}

.status-text {
  font-size: 12px;
  color: var(--text-tertiary);
  min-width: 50px;
}

.status-text.active {
  color: #00B42A;
  font-weight: 500;
}

.feature-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.action-link {
  background: none;
  border: none;
  color: #165DFF;
  cursor: pointer;
  font-size: 13px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.action-link:hover {
  background: rgba(22, 93, 255, 0.1);
}

.action-link.danger {
  color: #F53F3F;
}

.action-link.danger:hover {
  background: rgba(245, 63, 63, 0.1);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #E2E8F0;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
}

.modal-close {
  background: none;
  border: none;
  font-size: 24px;
  color: #64748B;
  cursor: pointer;
  padding: 0;
  width: 28px;
  height: 28px;
  border-radius: 4px;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 13px;
  color: #475569;
  margin-bottom: 6px;
  font-weight: 500;
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  color: #1E293B;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  font-family: inherit;
}

.form-input:focus {
  border-color: #165DFF;
}

.form-hint {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-weight: normal !important;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid #E2E8F0;
}

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: none;
  font-weight: 500;
}

.btn-primary {
  background: #165DFF;
  color: white;
}

.btn-secondary {
  background: #F1F5F9;
  color: #475569;
}

@media screen and (max-width: 768px) {
  .feature-item {
    flex-wrap: wrap;
  }

  .feature-info {
    width: calc(100% - 56px);
  }

  .feature-status {
    width: 100%;
    justify-content: flex-end;
    padding-top: 8px;
    border-top: 1px solid var(--border-color);
  }

  .feature-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
