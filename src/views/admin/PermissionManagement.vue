<script setup>
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useAuthStore } from '@/stores/auth';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const authStore = useAuthStore();
const logsStore = useLogsStore();

const roles = ref([]);
const menus = ref([]);            // 扁平化后的功能菜单列表
const expandedRoles = ref([]);
const roleMenuMap = ref({});      // roleId -> 选中的 menu id 列表
const isLoading = ref(false);

const showRoleForm = ref(false);
const editRoleId = ref(null);
const roleForm = reactive({ name: '', code: '', description: '' });

// 将树形菜单拍平，方便渲染勾选列表
const flattenMenus = (tree, acc = []) => {
  (tree || []).forEach(m => {
    acc.push({ id: m.id, name: m.name, type: m.type, parent_id: m.parent_id });
    if (m.children && m.children.length) flattenMenus(m.children, acc);
  });
  return acc;
};

const loadRoles = async () => {
  try {
    roles.value = await adminApi.getRoles();
  } catch (error) {
    ElMessage.error('加载角色列表失败: ' + error.message);
  }
};

const loadMenus = async () => {
  try {
    const tree = await adminApi.getMenus();
    menus.value = flattenMenus(tree || []);
  } catch (error) {
    ElMessage.error('加载功能菜单失败: ' + error.message);
  }
};

// 展开角色时加载其已分配的功能菜单
const loadRoleMenus = async (roleId) => {
  if (roleMenuMap.value[roleId]) return;
  try {
    const role = await adminApi.getRole(roleId);
    roleMenuMap.value[roleId] = role.menu_ids || [];
  } catch (e) {
    roleMenuMap.value[roleId] = [];
  }
};

const isRoleExpanded = (roleId) => expandedRoles.value.includes(roleId);

const toggleRole = async (roleId) => {
  const idx = expandedRoles.value.indexOf(roleId);
  if (idx > -1) {
    expandedRoles.value.splice(idx, 1);
  } else {
    expandedRoles.value.push(roleId);
    await loadRoleMenus(roleId);
  }
};

const isMenuChecked = (roleId, menuId) => (roleMenuMap.value[roleId] || []).includes(menuId);

const toggleMenu = (roleId, menuId) => {
  if (!roleMenuMap.value[roleId]) roleMenuMap.value[roleId] = [];
  const arr = roleMenuMap.value[roleId];
  const i = arr.indexOf(menuId);
  if (i > -1) arr.splice(i, 1);
  else arr.push(menuId);
};

const handleSaveRoleMenus = async (roleId) => {
  if (!authStore.hasPermission('role:assign')) {
    ElMessage.error('您没有分配权限');
    return;
  }
  try {
    await adminApi.setRoleMenus(roleId, roleMenuMap.value[roleId] || []);
    ElMessage.success('权限已保存');
    logsStore.addLog('权限管理', '权限管理', `更新角色功能权限: ${getRoleName(roleId)}`);
  } catch (error) {
    ElMessage.error('保存失败: ' + error.message);
  }
};

const handleAddRole = () => {
  editRoleId.value = null;
  roleForm.name = '';
  roleForm.code = '';
  roleForm.description = '';
  showRoleForm.value = true;
};

const handleEditRole = (role) => {
  editRoleId.value = role.id;
  roleForm.name = role.name;
  roleForm.code = role.code || '';
  roleForm.description = role.description || '';
  showRoleForm.value = true;
};

const handleSaveRole = async () => {
  if (!roleForm.name.trim()) {
    ElMessage.warning('请输入角色名称');
    return;
  }
  try {
    if (editRoleId.value) {
      await adminApi.updateRole(editRoleId.value, { name: roleForm.name, description: roleForm.description, status: 1 });
      ElMessage.success('角色更新成功');
      logsStore.addLog('编辑', '权限管理', `更新角色: ${roleForm.name}`);
    } else {
      if (!roleForm.code.trim()) {
        ElMessage.warning('请输入角色编码');
        return;
      }
      await adminApi.createRole({ name: roleForm.name, code: roleForm.code, description: roleForm.description, status: 1 });
      ElMessage.success('角色创建成功');
      logsStore.addLog('创建', '权限管理', `创建角色: ${roleForm.name}`);
    }
    showRoleForm.value = false;
    loadRoles();
  } catch (error) {
    ElMessage.error('保存角色失败: ' + error.message);
  }
};

const handleDeleteRole = async (role) => {
  try {
    await ElMessageBox.confirm(`确定删除角色 "${role.name}" 吗？`, '删除确认', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
    });
    await adminApi.deleteRole(role.id);
    ElMessage.success('角色已删除');
    logsStore.addLog('删除', '权限管理', `删除角色: ${role.name}`);
    loadRoles();
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败: ' + error.message);
  }
};

const getRoleName = (roleId) => {
  const r = roles.value.find(x => x.id === roleId);
  return r ? r.name : roleId;
};

onMounted(async () => {
  if (!authStore.isLoggedIn) {
    try { await authStore.loadAuthState(); } catch (e) {}
  }
  if (!authStore.hasPermission('role:view') && !authStore.hasPermission('permission:manage')) {
    ElMessage.error('您没有访问权限管理的权限');
    return;
  }
  isLoading.value = true;
  await Promise.all([loadRoles(), loadMenus()]);
  isLoading.value = false;
});
</script>

<template>
  <div class="permission-management">
    <GlassCard title="权限管理">
      <div v-if="isLoading" class="loading-state">加载中...</div>

      <div v-else class="content-panel">
        <div class="toolbar">
          <div class="info-text">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="16" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
            <span>管理角色，并为每个角色分配可见的功能菜单（权限）</span>
          </div>
          <button class="add-btn" @click="handleAddRole" v-if="authStore.hasPermission('role:assign')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            <span>添加角色</span>
          </button>
        </div>

        <div class="role-list">
          <div
            v-for="role in roles"
            :key="role.id"
            class="role-item"
            :class="{ expanded: isRoleExpanded(role.id) }"
          >
            <div class="role-header" @click="toggleRole(role.id)">
              <div class="role-info">
                <div class="role-title">
                  <span class="role-name">{{ role.name }}</span>
                  <span class="role-id" v-if="role.code">({{ role.code }})</span>
                </div>
                <div class="role-description">{{ role.description || '暂无描述' }}</div>
                <div class="role-stats">
                  <span class="stat-item">已分配功能: <strong>{{ (roleMenuMap[role.id] || []).length }}</strong></span>
                </div>
              </div>
              <div class="role-actions" @click.stop>
                <button class="edit-btn" @click="handleEditRole(role)" v-if="authStore.hasPermission('role:assign')">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                  </svg>
                  <span>编辑</span>
                </button>
                <button class="delete-btn" @click="handleDeleteRole(role)" v-if="authStore.hasPermission('role:assign')">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="3 6 5 6 21 6" />
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                  </svg>
                  <span>删除</span>
                </button>
                <button class="expand-btn" @click="toggleRole(role.id)">
                  <svg :class="{ rotated: isRoleExpanded(role.id) }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </button>
              </div>
            </div>

            <div v-if="isRoleExpanded(role.id)" class="role-content">
              <div v-if="menus.length" class="permission-list">
                <label
                  v-for="m in menus"
                  :key="m.id"
                  class="permission-item"
                >
                  <input
                    type="checkbox"
                    :checked="isMenuChecked(role.id, m.id)"
                    :disabled="!authStore.hasPermission('role:assign')"
                    @change="toggleMenu(role.id, m.id)"
                  />
                  <div class="permission-info">
                    <div class="permission-name">{{ m.name }}</div>
                    <div class="permission-desc">{{ m.type === 'button' ? '按钮' : '菜单' }}</div>
                  </div>
                </label>
              </div>
              <div v-else class="empty-row">暂无可分配的功能菜单</div>
              <div class="role-content-footer">
                <button class="save-btn" @click="handleSaveRoleMenus(role.id)" v-if="authStore.hasPermission('role:assign')">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                    <polyline points="17 21 17 13 7 13 7 21" />
                    <polyline points="7 3 7 8 15 8" />
                  </svg>
                  <span>保存权限</span>
                </button>
              </div>
            </div>
          </div>
          <div v-if="roles.length === 0" class="empty-row">暂无角色</div>
        </div>
      </div>
    </GlassCard>

    <Teleport to="body">
      <div v-if="showRoleForm" class="modal-overlay" @click.self="showRoleForm = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>{{ editRoleId ? '编辑角色' : '添加角色' }}</h3>
            <button class="modal-close" @click="showRoleForm = false">×</button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>角色名称 *</label>
              <input v-model="roleForm.name" type="text" class="form-input" placeholder="例如: 管理员" />
            </div>
            <div class="form-group" v-if="!editRoleId">
              <label>角色编码 *</label>
              <input v-model="roleForm.code" type="text" class="form-input" placeholder="例如: manager" />
            </div>
            <div class="form-group">
              <label>角色描述</label>
              <textarea v-model="roleForm.description" class="form-input" rows="2"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showRoleForm = false">取消</button>
            <button class="btn btn-primary" @click="handleSaveRole">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.permission-management {
  width: 100%;
}

.content-panel {
  background: #ffffff;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  overflow: hidden;
  padding: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
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
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: none;
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  color: white;
}

.add-btn:hover {
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
}

.loading-state {
  padding: 60px 0;
  text-align: center;
  color: var(--text-tertiary, #94A3B8);
}

.role-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.role-item {
  background: var(--bg-input, #F8FAFC);
  border: 1px solid var(--border-color, #E2E8F0);
  border-radius: 10px;
  overflow: hidden;
  transition: all 0.2s;
}

.role-item.expanded {
  border-color: rgba(22, 93, 255, 0.4);
}

.role-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  cursor: pointer;
  transition: background 0.2s;
}

.role-header:hover {
  background: rgba(22, 93, 255, 0.04);
}

.role-info {
  flex: 1;
}

.role-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
}

.role-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1E293B);
}

.role-id {
  font-size: 12px;
  color: var(--text-tertiary, #94A3B8);
  font-family: monospace;
}

.role-description {
  font-size: 12px;
  color: var(--text-tertiary, #94A3B8);
  margin-bottom: 8px;
}

.role-stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  font-size: 12px;
  color: var(--text-secondary, #475569);
}

.stat-item strong {
  color: #165DFF;
  font-weight: 600;
}

.role-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.edit-btn,
.save-btn,
.delete-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.edit-btn {
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
}

.edit-btn:hover {
  background: rgba(22, 93, 255, 0.2);
}

.delete-btn {
  background: rgba(245, 63, 63, 0.1);
  color: #F53F3F;
}

.delete-btn:hover {
  background: rgba(245, 63, 63, 0.2);
}

.save-btn {
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
}

.save-btn:hover {
  background: rgba(0, 180, 42, 0.2);
}

.expand-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-tertiary, #94A3B8);
  padding: 6px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.expand-btn:hover {
  background: rgba(22, 93, 255, 0.1);
}

.expand-btn svg {
  transition: transform 0.2s;
}

.expand-btn svg.rotated {
  transform: rotate(180deg);
}

.role-content {
  padding: 0 20px 20px 20px;
  border-top: 1px solid var(--border-color, #E2E8F0);
  background: rgba(22, 93, 255, 0.02);
}

.role-content-footer {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.permission-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 8px;
  margin-top: 16px;
}

.permission-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-input, #F8FAFC);
  border: 1px solid var(--border-color, #E2E8F0);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.permission-item:hover {
  border-color: rgba(22, 93, 255, 0.3);
  background: rgba(22, 93, 255, 0.04);
}

.permission-item input[type="checkbox"] {
  margin-top: 2px;
  cursor: pointer;
}

.permission-item input[type="checkbox"]:disabled {
  cursor: not-allowed;
}

.permission-info {
  flex: 1;
}

.permission-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary, #1E293B);
  margin-bottom: 2px;
}

.permission-desc {
  font-size: 11px;
  color: var(--text-tertiary, #94A3B8);
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
  .role-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .role-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .permission-list {
    grid-template-columns: 1fr;
  }
}
</style>
