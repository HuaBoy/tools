<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useAuthStore } from '@/stores/auth';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const authStore = useAuthStore();
const logsStore = useLogsStore();

const menus = ref([]);
const isLoading = ref(false);
const showForm = ref(false);
const editId = ref(null);

const form = reactive({
  name: '',
  type: 'menu',
  path: '',
  icon: '',
  parent_id: 0,
  sort: 0,
  status: 1
});

// 树拍平，用于列表展示与父级选择
const flatten = (tree, acc = [], depth = 0) => {
  (tree || []).forEach(m => {
    acc.push({ ...m, depth });
    if (m.children && m.children.length) flatten(m.children, acc, depth + 1);
  });
  return acc;
};

const flatMenus = computed(() => flatten(menus.value));

const loadMenus = async () => {
  isLoading.value = true;
  try {
    menus.value = await adminApi.getMenus();
  } catch (e) {
    ElMessage.error('加载功能列表失败: ' + e.message);
  } finally {
    isLoading.value = false;
  }
};

// 父级下拉：编辑时不能选自己（简化：允许选任意其它菜单）
const parentOptions = computed(() => flatMenus.value.filter(m => m.id !== editId.value));

const handleAdd = () => {
  editId.value = null;
  form.name = '';
  form.type = 'menu';
  form.path = '';
  form.icon = '';
  form.parent_id = 0;
  form.sort = flatMenus.value.length + 1;
  form.status = 1;
  showForm.value = true;
};

const handleEdit = (m) => {
  editId.value = m.id;
  form.name = m.name;
  form.type = m.type || 'menu';
  form.path = m.path || '';
  form.icon = m.icon || '';
  form.parent_id = m.parent_id || 0;
  form.sort = m.sort || 0;
  form.status = m.status === undefined ? 1 : m.status;
  showForm.value = true;
};

const handleSave = async () => {
  if (!form.name.trim()) {
    ElMessage.warning('请输入功能名称');
    return;
  }
  try {
    const payload = {
      name: form.name,
      type: form.type,
      path: form.path,
      icon: form.icon,
      parent_id: form.parent_id,
      sort: form.sort,
      status: form.status
    };
    if (editId.value) {
      await adminApi.updateMenu(editId.value, payload);
      ElMessage.success('功能更新成功');
      logsStore.addLog('编辑', '功能管理', `更新功能: ${form.name}`);
    } else {
      await adminApi.createMenu(payload);
      ElMessage.success('功能创建成功');
      logsStore.addLog('创建', '功能管理', `创建功能: ${form.name}`);
    }
    showForm.value = false;
    loadMenus();
  } catch (e) {
    ElMessage.error('保存失败: ' + e.message);
  }
};

const handleDelete = async (m) => {
  try {
    await ElMessageBox.confirm(`确定删除功能 "${m.name}" 吗？其子功能也会一并删除。`, '删除确认', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
    });
    await adminApi.deleteMenu(m.id);
    ElMessage.success('功能已删除');
    logsStore.addLog('删除', '功能管理', `删除功能: ${m.name}`);
    loadMenus();
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败: ' + e.message);
  }
};

const typeText = (t) => (t === 'button' ? '按钮' : '菜单');
const statusText = (s) => (s === 1 ? '启用' : '禁用');
const parentName = (pid) => {
  if (!pid) return '顶级菜单';
  const p = flatMenus.value.find(m => m.id === pid);
  return p ? p.name : '顶级菜单';
};

onMounted(async () => {
  if (!authStore.isLoggedIn) {
    try { await authStore.loadAuthState(); } catch (e) {}
  }
  if (!authStore.hasPermission('feature:view')) {
    ElMessage.error('您没有访问功能管理的权限');
    return;
  }
  loadMenus();
});
</script>

<template>
  <div class="feature-management">
    <GlassCard title="功能管理">
      <div class="content-panel">
        <div class="toolbar">
          <div class="info-text">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="16" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
            <span>管理系统功能菜单（菜单 / 按钮），用于角色权限分配</span>
          </div>
          <button class="add-btn" @click="handleAdd" v-if="authStore.hasPermission('feature:toggle')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            <span>添加功能</span>
          </button>
        </div>

        <div v-if="isLoading" class="loading-state">加载中...</div>

        <div v-else class="feature-table">
          <table>
            <thead>
              <tr>
                <th>功能名称</th>
                <th>类型</th>
                <th>路由/标识</th>
                <th>父级</th>
                <th>排序</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in flatMenus" :key="m.id">
                <td>
                  <span class="indent" :style="{ paddingLeft: (m.depth * 20) + 'px' }"></span>
                  {{ m.name }}
                </td>
                <td>
                  <span class="type-tag" :class="m.type === 'button' ? 'btn' : 'menu'">{{ typeText(m.type) }}</span>
                </td>
                <td class="path-cell">{{ m.path || '-' }}</td>
                <td>{{ parentName(m.parent_id) }}</td>
                <td>{{ m.sort }}</td>
                <td>
                  <span class="status-tag" :class="m.status === 1 ? 'enabled' : 'disabled'">{{ statusText(m.status) }}</span>
                </td>
                <td class="actions">
                  <button class="action-link" @click="handleEdit(m)" v-if="authStore.hasPermission('feature:toggle')">编辑</button>
                  <button class="action-link danger" @click="handleDelete(m)" v-if="authStore.hasPermission('feature:toggle')">删除</button>
                </td>
              </tr>
              <tr v-if="flatMenus.length === 0">
                <td colspan="7" class="empty-row">暂无功能配置</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </GlassCard>

    <Teleport to="body">
      <div v-if="showForm" class="modal-overlay" @click.self="showForm = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>{{ editId ? '编辑功能' : '添加功能' }}</h3>
            <button class="modal-close" @click="showForm = false">×</button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>功能名称 *</label>
              <input v-model="form.name" type="text" class="form-input" placeholder="例如: 批次数据追溯" />
            </div>
            <div class="form-group">
              <label>类型</label>
              <select v-model="form.type" class="form-input">
                <option value="menu">菜单</option>
                <option value="button">按钮</option>
              </select>
            </div>
            <div class="form-group">
              <label>路由 / 标识 (Path)</label>
              <input v-model="form.path" type="text" class="form-input" placeholder="例如: /data/query" />
            </div>
            <div class="form-group">
              <label>图标</label>
              <input v-model="form.icon" type="text" class="form-input" placeholder="图标名称" />
            </div>
            <div class="form-group">
              <label>父级功能</label>
              <select v-model.number="form.parent_id" class="form-input">
                <option :value="0">顶级菜单</option>
                <option v-for="p in parentOptions" :key="p.id" :value="p.id">{{ ' '.repeat(p.depth) }}{{ p.name }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>排序</label>
              <input v-model.number="form.sort" type="number" class="form-input" />
            </div>
            <div class="form-group">
              <label>状态</label>
              <select v-model.number="form.status" class="form-input">
                <option :value="1">启用</option>
                <option :value="0">禁用</option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showForm = false">取消</button>
            <button class="btn btn-primary" @click="handleSave">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.feature-management {
  width: 100%;
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
  color: var(--text-tertiary, #94A3B8);
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
  color: var(--text-tertiary, #94A3B8);
}

.content-panel {
  background: #ffffff;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  overflow: hidden;
  padding: 16px;
}

.feature-table {
  overflow-x: auto;
}

.feature-table table {
  width: 100%;
  border-collapse: collapse;
}

.feature-table th,
.feature-table td {
  padding: 12px 16px;
  text-align: left;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color, #E2E8F0);
}

.feature-table th {
  color: var(--text-tertiary, #94A3B8);
  font-weight: 500;
  background: rgba(22, 93, 255, 0.04);
}

.feature-table td {
  color: var(--text-secondary, #475569);
}

.feature-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.04);
}

.indent {
  display: inline-block;
}

.path-cell {
  font-family: monospace;
  color: #64748B;
}

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.type-tag.menu {
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
}

.type-tag.btn {
  background: rgba(255, 125, 0, 0.1);
  color: #FF7D00;
}

.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.status-tag.enabled {
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
}

.status-tag.disabled {
  background: rgba(148, 163, 184, 0.1);
  color: #64748B;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-link {
  background: none;
  border: none;
  color: #165DFF;
  cursor: pointer;
  font-size: 13px;
  padding: 0;
  transition: opacity 0.2s;
}

.action-link:hover {
  opacity: 0.7;
  text-decoration: underline;
}

.action-link.danger {
  color: #F53F3F;
}

.empty-row {
  text-align: center;
  color: var(--text-tertiary, #94A3B8);
  padding: 40px !important;
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
  overflow: hidden;
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

.modal-close:hover {
  background: #F1F5F9;
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
  .feature-table th,
  .feature-table td {
    padding: 8px 10px;
    font-size: 12px;
  }
}
</style>
