<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useAuthStore } from '@/stores/auth';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const authStore = useAuthStore();
const logsStore = useLogsStore();

const roles = ref([]);
const allPermissions = ref([]);
const users = ref([]);

// 角色与权限的映射（模拟数据）
const rolePermissionsMap = ref({});

const isLoading = ref(false);
const showRoleForm = ref(false);
const editRoleId = ref(null);
const roleForm = reactive({
  id: '',
  name: '',
  description: '',
  permissions: []
});

const expandedRoles = ref([]);

const loadRoles = async () => {
  try {
    roles.value = await adminApi.getRoles();
    // 初始化角色权限映射
    roles.value.forEach(role => {
      if (!rolePermissionsMap.value[role.id]) {
        // 超级管理员拥有所有权限
        if (role.id === 'admin') {
          rolePermissionsMap.value[role.id] = allPermissions.value.map(p => p.id);
        } else {
          rolePermissionsMap.value[role.id] = role.defaultPermissions || [];
        }
      }
    });
  } catch (error) {
    ElMessage.error('加载角色列表失败: ' + error.message);
  }
};

const loadPermissions = async () => {
  try {
    allPermissions.value = await adminApi.getPermissions();
  } catch (error) {
    ElMessage.error('加载权限列表失败: ' + error.message);
  }
};

const loadUsers = async () => {
  try {
    const result = await adminApi.getUsers(1, 100, '');
    users.value = result.data;
  } catch (error) {
    console.error('加载用户失败:', error);
  }
};

const toggleRole = (roleId) => {
  const idx = expandedRoles.value.indexOf(roleId);
  if (idx > -1) {
    expandedRoles.value.splice(idx, 1);
  } else {
    expandedRoles.value.push(roleId);
  }
};

const isRoleExpanded = (roleId) => expandedRoles.value.includes(roleId);

const toggleRolePermission = (roleId, permissionId) => {
  if (!rolePermissionsMap.value[roleId]) {
    rolePermissionsMap.value[roleId] = [];
  }
  const idx = rolePermissionsMap.value[roleId].indexOf(permissionId);
  if (idx > -1) {
    rolePermissionsMap.value[roleId].splice(idx, 1);
  } else {
    rolePermissionsMap.value[roleId].push(permissionId);
  }
};

const isPermissionChecked = (roleId, permissionId) => {
  return (rolePermissionsMap.value[roleId] || []).includes(permissionId);
};

const getRoleUserCount = (roleId) => {
  return users.value.filter(u => u.role === roleId).length;
};

const getRolePermissions = (roleId) => {
  return rolePermissionsMap.value[roleId] || [];
};

const handleSaveRolePermissions = async (roleId) => {
  if (!authStore.hasPermission('role:assign')) {
    ElMessage.error('您没有分配角色的权限');
    return;
  }

  try {
    // 这里应该调用API保存角色权限
    // await adminApi.updateRolePermissions(roleId, rolePermissionsMap.value[roleId]);
    ElMessage.success('角色权限已保存');
    logsStore.addLog('权限管理', '权限管理', `更新角色权限: ${getRoleName(roleId)}`);
  } catch (error) {
    ElMessage.error('保存角色权限失败: ' + error.message);
  }
};

const handleEditRole = (role) => {
  editRoleId.value = role.id;
  roleForm.id = role.id;
  roleForm.name = role.name;
  roleForm.description = role.description;
  roleForm.permissions = [...(rolePermissionsMap.value[role.id] || [])];
  showRoleForm.value = true;
};

const handleSaveRole = async () => {
  if (!roleForm.name.trim()) {
    ElMessage.warning('请输入角色名称');
    return;
  }

  try {
    if (editRoleId.value) {
      rolePermissionsMap.value[editRoleId.value] = [...roleForm.permissions];
      ElMessage.success('角色更新成功');
      logsStore.addLog('编辑', '权限管理', `更新角色: ${roleForm.name}`);
    }
    showRoleForm.value = false;
  } catch (error) {
    ElMessage.error('保存角色失败: ' + error.message);
  }
};

const getRoleName = (roleId) => {
  const role = roles.value.find(r => r.id === roleId);
  return role ? role.name : roleId;
};

const getPermissionName = (permissionId) => {
  const perm = allPermissions.value.find(p => p.id === permissionId);
  return perm ? perm.name : permissionId;
};

const getPermissionDescription = (permissionId) => {
  const perm = allPermissions.value.find(p => p.id === permissionId);
  return perm ? perm.description : '';
};

const permissionGroups = computed(() => {
  const groups = {};
  allPermissions.value.forEach(p => {
    const groupName = p.id.split(':')[0] || 'other';
    if (!groups[groupName]) {
      groups[groupName] = [];
    }
    groups[groupName].push(p);
  });
  return groups;
});

onMounted(async () => {
  if (!authStore.hasPermission('role:view') && !authStore.hasPermission('permission:manage')) {
    ElMessage.error('您没有访问权限管理的权限');
    return;
  }
  isLoading.value = true;
  await Promise.all([loadPermissions(), loadRoles(), loadUsers()]);
  isLoading.value = false;
});
</script>

<template>
  <div class="permission-management">
    <GlassCard title="权限管理">
      <div v-if="isLoading" class="loading-state">加载中...</div>

      <div v-else class="role-list">
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
                <span class="role-id">({{ role.id }})</span>
              </div>
              <div class="role-description">{{ role.description }}</div>
              <div class="role-stats">
                <span class="stat-item">用户数: <strong>{{ getRoleUserCount(role.id) }}</strong></span>
                <span class="stat-item">权限数: <strong>{{ getRolePermissions(role.id).length }}</strong></span>
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
              <button
                class="save-btn"
                @click="handleSaveRolePermissions(role.id)"
                v-if="authStore.hasPermission('role:assign')"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                  <polyline points="17 21 17 13 7 13 7 21" />
                  <polyline points="7 3 7 8 15 8" />
                </svg>
                <span>保存权限</span>
              </button>
              <button class="expand-btn" @click="toggleRole(role.id)">
                <svg
                  :class="{ rotated: isRoleExpanded(role.id) }"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="6 9 12 15 18 9" />
                </svg>
              </button>
            </div>
          </div>

          <div v-if="isRoleExpanded(role.id)" class="role-content">
            <div
              v-for="(perms, groupName) in permissionGroups"
              :key="groupName"
              class="permission-group"
            >
              <div class="permission-group-title">
                {{ groupName.toUpperCase() }} 模块
              </div>
              <div class="permission-list">
                <label
                  v-for="p in perms"
                  :key="p.id"
                  class="permission-item"
                >
                  <input
                    type="checkbox"
                    :checked="isPermissionChecked(role.id, p.id)"
                    :disabled="!authStore.hasPermission('role:assign')"
                    @change="toggleRolePermission(role.id, p.id)"
                  />
                  <div class="permission-info">
                    <div class="permission-name">{{ p.name }}</div>
                    <div class="permission-desc">{{ p.description }}</div>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>

    <Teleport to="body">
      <div v-if="showRoleForm" class="modal-overlay" @click.self="showRoleForm = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>编辑角色</h3>
            <button class="modal-close" @click="showRoleForm = false">×</button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>角色ID</label>
              <input v-model="roleForm.id" type="text" class="form-input" disabled />
            </div>
            <div class="form-group">
              <label>角色名称</label>
              <input v-model="roleForm.name" type="text" class="form-input" />
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
  max-width: 1400px;
}

.loading-state {
  padding: 60px 0;
  text-align: center;
  color: var(--text-tertiary);
}

.role-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.role-item {
  background: var(--bg-input);
  border: 1px solid var(--border-color);
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
  color: var(--text-primary);
}

.role-id {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: monospace;
}

.role-description {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}

.role-stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  font-size: 12px;
  color: var(--text-secondary);
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
.save-btn {
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
  color: var(--text-tertiary);
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
  border-top: 1px solid var(--border-color);
  background: rgba(22, 93, 255, 0.02);
}

.permission-group {
  margin-top: 16px;
}

.permission-group-title {
  font-size: 12px;
  font-weight: 600;
  color: #165DFF;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.permission-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 8px;
}

.permission-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
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
  color: var(--text-primary);
  margin-bottom: 2px;
}

.permission-desc {
  font-size: 11px;
  color: var(--text-tertiary);
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
