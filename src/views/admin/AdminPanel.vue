<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useAuthStore } from '@/stores/auth';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const authStore = useAuthStore();
const logsStore = useLogsStore();

const activeTab = ref('users');

const userSearchForm = reactive({
  search: '',
  role: ''
});

const users = ref([]);
const userTotal = ref(0);
const userCurrentPage = ref(1);
const userPageSize = ref(10);
const isLoadingUsers = ref(false);

const showUserForm = ref(false);
const editUserId = ref(null);
const userForm = reactive({
  username: '',
  email: '',
  role: 'user',
  permissions: []
});

const roles = ref([]);
const allPermissions = ref([]);

const features = ref([]);
const isLoadingFeatures = ref(false);

const roleOptions = computed(() => {
  return roles.value.map(r => ({ label: r.name, value: r.id }));
});

const permissionOptions = computed(() => {
  return allPermissions.value.map(p => ({ label: p.name, value: p.id }));
});

const loadUsers = async () => {
  isLoadingUsers.value = true;
  try {
    const result = await adminApi.getUsers(userCurrentPage.value, userPageSize.value, userSearchForm.search);
    users.value = result.data;
    userTotal.value = result.total;
  } catch (error) {
    ElMessage.error('加载用户列表失败: ' + error.message);
  } finally {
    isLoadingUsers.value = false;
  }
};

const loadRoles = async () => {
  try {
    roles.value = await adminApi.getRoles();
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

const loadFeatures = async () => {
  isLoadingFeatures.value = true;
  try {
    features.value = await adminApi.getFeatures();
  } catch (error) {
    ElMessage.error('加载功能列表失败: ' + error.message);
  } finally {
    isLoadingFeatures.value = false;
  }
};

const handleSearchUsers = () => {
  userCurrentPage.value = 1;
  loadUsers();
};

const handleAddUser = () => {
  editUserId.value = null;
  userForm.username = '';
  userForm.email = '';
  userForm.role = 'user';
  userForm.permissions = [];
  showUserForm.value = true;
};

const handleEditUser = async (user) => {
  editUserId.value = user.id;
  userForm.username = user.username || '';
  userForm.email = user.email || '';
  userForm.role = user.role || 'user';
  userForm.permissions = user.permissions || [];
  showUserForm.value = true;
};

const handleSaveUser = async () => {
  if (!userForm.username || !userForm.email) {
    ElMessage.warning('请填写用户名和邮箱');
    return;
  }

  try {
    if (editUserId.value) {
      await adminApi.updateUser(editUserId.value, {
        username: userForm.username,
        role: userForm.role,
        permissions: userForm.permissions
      });
      ElMessage.success('用户信息更新成功');
      logsStore.addLog('编辑', '用户管理', `编辑用户: ${userForm.username}`);
    } else {
      const password = Math.random().toString(36).substr(2, 8);
      await adminApi.register(userForm.email, password, userForm.username);
      ElMessage.success(`用户创建成功，初始密码: ${password}`);
      logsStore.addLog('创建', '用户管理', `创建用户: ${userForm.username}`);
    }
    showUserForm.value = false;
    loadUsers();
  } catch (error) {
    ElMessage.error('保存用户失败: ' + error.message);
  }
};

const handleDeleteUser = async (user) => {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });

    await adminApi.deleteUser(user.id);
    ElMessage.success('用户删除成功');
    logsStore.addLog('删除', '用户管理', `删除用户: ${user.username}`);
    loadUsers();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除用户失败: ' + error.message);
    }
  }
};

const handleResetPassword = async (user) => {
  try {
    await ElMessageBox.confirm('确定要重置该用户密码吗？系统将发送重置邮件。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });

    await adminApi.resetPassword(user.email);
    ElMessage.success('密码重置邮件已发送');
    logsStore.addLog('重置密码', '用户管理', `重置用户密码: ${user.username}`);
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重置密码失败: ' + error.message);
    }
  }
};

const handleToggleFeature = async (feature) => {
  try {
    await adminApi.toggleFeature(feature.id, !feature.enabled);
    feature.enabled = !feature.enabled;
    ElMessage.success(feature.enabled ? '功能已开启' : '功能已关闭');
    logsStore.addLog('功能开关', '功能管理', `${feature.enabled ? '开启' : '关闭'}功能: ${feature.name}`);
  } catch (error) {
    ElMessage.error('操作失败: ' + error.message);
  }
};

const handleAddFeature = async () => {
  const name = prompt('请输入功能名称:');
  if (!name) return;

  try {
    await adminApi.createFeature({
      name,
      key: name.toLowerCase().replace(/\s+/g, '_'),
      description: '',
      enabled: true,
      order: features.value.length + 1
    });
    ElMessage.success('功能创建成功');
    logsStore.addLog('创建', '功能管理', `创建功能: ${name}`);
    loadFeatures();
  } catch (error) {
    ElMessage.error('创建功能失败: ' + error.message);
  }
};

const handleEditFeature = async (feature) => {
  const name = prompt('请输入功能名称:', feature.name);
  if (!name) return;

  try {
    await adminApi.updateFeature(feature.id, { name });
    ElMessage.success('功能更新成功');
    logsStore.addLog('编辑', '功能管理', `编辑功能: ${name}`);
    loadFeatures();
  } catch (error) {
    ElMessage.error('更新功能失败: ' + error.message);
  }
};

const handleDeleteFeature = async (feature) => {
  try {
    await ElMessageBox.confirm('确定要删除该功能吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });

    await adminApi.deleteFeature(feature.id);
    ElMessage.success('功能删除成功');
    logsStore.addLog('删除', '功能管理', `删除功能: ${feature.name}`);
    loadFeatures();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除功能失败: ' + error.message);
    }
  }
};

const getRoleName = (roleId) => {
  const role = roles.value.find(r => r.id === roleId);
  return role ? role.name : roleId;
};

const getPermissionNames = (permissionIds) => {
  return permissionIds.map(id => {
    const perm = allPermissions.value.find(p => p.id === id);
    return perm ? perm.name : id;
  }).join(', ');
};

onMounted(() => {
  loadUsers();
  loadRoles();
  loadPermissions();
  loadFeatures();
});
</script>

<template>
  <div class="admin-page">
    <GlassCard>
      <div class="page-header">
        <h2>后台管理系统</h2>
        <p>用户管理、权限管理、功能管理</p>
      </div>

      <div class="tab-container">
        <div class="tabs">
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'users' }"
            @click="activeTab = 'users'"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
              <path d="M16 3.13a4 4 0 0 1 0 7.75" />
            </svg>
            <span>用户管理</span>
          </button>
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'permissions' }"
            @click="activeTab = 'permissions'"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10" />
              <path d="m9 12 2 2 4-4" />
            </svg>
            <span>权限管理</span>
          </button>
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'features' }"
            @click="activeTab = 'features'"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="7" height="7" />
              <rect x="14" y="3" width="7" height="7" />
              <rect x="14" y="14" width="7" height="7" />
              <rect x="3" y="14" width="7" height="7" />
            </svg>
            <span>功能管理</span>
          </button>
        </div>

        <div class="tab-content">
          <div v-if="activeTab === 'users'" class="tab-pane">
            <div class="search-bar">
              <input 
                v-model="userSearchForm.search"
                type="text" 
                placeholder="搜索用户名或邮箱"
                class="search-input"
                @keyup.enter="handleSearchUsers"
              />
              <button class="search-btn" @click="handleSearchUsers">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="11" cy="11" r="8" />
                  <path d="m21 21-4.35-4.35" />
                </svg>
                <span>搜索</span>
              </button>
              <button class="add-btn" @click="handleAddUser">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 5v14" />
                  <path d="M5 12h14" />
                </svg>
                <span>添加用户</span>
              </button>
            </div>

            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>用户名</th>
                    <th>邮箱</th>
                    <th>角色</th>
                    <th>权限</th>
                    <th>创建时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="user in users" :key="user.id">
                    <td>{{ user.username }}</td>
                    <td>{{ user.email }}</td>
                    <td>
                      <span class="role-tag" :class="user.role">
                        {{ getRoleName(user.role) }}
                      </span>
                    </td>
                    <td class="permissions-cell">
                      {{ getPermissionNames(user.permissions) || '-' }}
                    </td>
                    <td>{{ user.created_at ? new Date(user.created_at).toLocaleString() : '-' }}</td>
                    <td>
                      <button class="action-btn edit" @click="handleEditUser(user)">编辑</button>
                      <button class="action-btn reset" @click="handleResetPassword(user)">重置密码</button>
                      <button class="action-btn delete" @click="handleDeleteUser(user)">删除</button>
                    </td>
                  </tr>
                  <tr v-if="users.length === 0">
                    <td colspan="6" class="empty-row">
                      <span v-if="isLoadingUsers">加载中...</span>
                      <span v-else>暂无数据</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="pagination" v-if="userTotal > userPageSize">
              <button 
                class="page-btn" 
                :disabled="userCurrentPage <= 1"
                @click="userCurrentPage--; loadUsers()"
              >上一页</button>
              <span class="page-info">{{ userCurrentPage }} / {{ Math.ceil(userTotal / userPageSize) }}</span>
              <button 
                class="page-btn" 
                :disabled="userCurrentPage >= Math.ceil(userTotal / userPageSize)"
                @click="userCurrentPage++; loadUsers()"
              >下一页</button>
            </div>
          </div>

          <div v-if="activeTab === 'permissions'" class="tab-pane">
            <div class="permissions-grid">
              <div class="permission-card">
                <h3>角色列表</h3>
                <div class="role-list">
                  <div v-for="role in roles" :key="role.id" class="role-item">
                    <span class="role-tag" :class="role.id">{{ role.name }}</span>
                    <span class="role-desc">{{ role.description }}</span>
                  </div>
                </div>
              </div>

              <div class="permission-card">
                <h3>权限列表</h3>
                <div class="permission-list">
                  <div v-for="perm in allPermissions" :key="perm.id" class="permission-item">
                    <span class="perm-id">{{ perm.id }}</span>
                    <span class="perm-name">{{ perm.name }}</span>
                    <span class="perm-desc">{{ perm.description }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="current-user-info">
              <h3>当前用户权限</h3>
              <div class="user-perm-card">
                <div class="user-perm-header">
                  <span class="user-name">{{ authStore.userProfile?.username || '-' }}</span>
                  <span class="user-role" :class="authStore.role">{{ getRoleName(authStore.role) }}</span>
                </div>
                <div class="user-perm-list">
                  <span 
                    v-for="perm in authStore.permissions" 
                    :key="perm"
                    class="perm-tag"
                  >
                    {{ getPermissionNames([perm]) }}
                  </span>
                  <span v-if="authStore.permissions.length === 0" class="no-perm">暂无权限</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'features'" class="tab-pane">
            <div class="feature-header">
              <h3>功能开关管理</h3>
              <button class="add-btn" @click="handleAddFeature">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 5v14" />
                  <path d="M5 12h14" />
                </svg>
                <span>添加功能</span>
              </button>
            </div>

            <div class="feature-list">
              <div v-for="feature in features" :key="feature.id" class="feature-item">
                <div class="feature-info">
                  <span class="feature-name">{{ feature.name }}</span>
                  <span class="feature-key">{{ feature.key }}</span>
                  <span class="feature-desc">{{ feature.description || '-' }}</span>
                </div>
                <div class="feature-actions">
                  <button 
                    class="toggle-btn"
                    :class="{ enabled: feature.enabled }"
                    @click="handleToggleFeature(feature)"
                  >
                    {{ feature.enabled ? '开启' : '关闭' }}
                  </button>
                  <button class="action-btn edit" @click="handleEditFeature(feature)">编辑</button>
                  <button class="action-btn delete" @click="handleDeleteFeature(feature)">删除</button>
                </div>
              </div>
              <div v-if="features.length === 0" class="empty-state">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="3" width="7" height="7" />
                  <rect x="14" y="3" width="7" height="7" />
                  <rect x="14" y="14" width="7" height="7" />
                  <rect x="3" y="14" width="7" height="7" />
                </svg>
                <span>暂无功能配置</span>
                <p>点击上方按钮添加功能</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="showUserForm" class="modal-overlay" @click.self="showUserForm = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>{{ editUserId ? '编辑用户' : '添加用户' }}</h3>
            <button class="close-btn" @click="showUserForm = false">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>用户名</label>
              <input v-model="userForm.username" type="text" class="form-input" placeholder="请输入用户名" />
            </div>
            <div class="form-group">
              <label>邮箱</label>
              <input v-model="userForm.email" type="email" class="form-input" placeholder="请输入邮箱" />
            </div>
            <div class="form-group">
              <label>角色</label>
              <select v-model="userForm.role" class="form-select">
                <option v-for="role in roles" :key="role.id" :value="role.id">
                  {{ role.name }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>权限</label>
              <div class="permission-select">
                <label 
                  v-for="perm in allPermissions" 
                  :key="perm.id"
                  class="perm-checkbox"
                >
                  <input 
                    type="checkbox" 
                    :checked="userForm.permissions.includes(perm.id)"
                    @change="(e) => {
                      if (e.target.checked) {
                        userForm.permissions.push(perm.id);
                      } else {
                        userForm.permissions = userForm.permissions.filter(p => p !== perm.id);
                      }
                    }"
                  />
                  <span>{{ perm.name }}</span>
                </label>
              </div>
            </div>
            <p v-if="!editUserId" class="hint-text">
              注意：新用户创建后，系统会生成初始密码并在提示中显示。
            </p>
          </div>
          <div class="modal-footer">
            <button class="btn btn-cancel" @click="showUserForm = false">取消</button>
            <button class="btn btn-confirm" @click="handleSaveUser">确定</button>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<style scoped>
.admin-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
  
  h2 {
    font-size: 24px;
    font-weight: 600;
    color: #1E293B;
    margin: 0 0 8px 0;
  }
  
  p {
    color: #64748B;
    margin: 0;
  }
}

.tab-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #E2E8F0;
}

.tabs {
  display: flex;
  border-bottom: 1px solid #E2E8F0;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 14px 20px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #64748B;
  transition: all 0.2s;
  border-bottom: 2px solid transparent;
  
  &:hover {
    color: #165DFF;
    background: rgba(22, 93, 255, 0.05);
  }
  
  &.active {
    color: #165DFF;
    border-bottom-color: #165DFF;
    background: rgba(22, 93, 255, 0.03);
  }
}

.tab-content {
  padding: 20px;
}

.search-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.search-input {
  flex: 1;
  padding: 10px 14px;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  
  &:focus {
    border-color: #165DFF;
  }
}

.search-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: #165DFF;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
  
  &:hover {
    background: #1550D6;
  }
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: #00B42A;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
  
  &:hover {
    background: #00A627;
  }
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  
  th {
    text-align: left;
    padding: 12px 16px;
    background: #F8FAFC;
    color: #475569;
    font-size: 13px;
    font-weight: 600;
    border-bottom: 2px solid #E2E8F0;
  }
  
  td {
    padding: 12px 16px;
    border-bottom: 1px solid #E2E8F0;
    font-size: 13px;
    color: #334155;
  }
  
  tr:hover {
    background: #F8FAFC;
  }
}

.empty-row {
  text-align: center;
  color: #94A3B8;
}

.role-tag {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  
  &.admin { background: rgba(245, 63, 63, 0.1); color: #F53F3F; }
  &.manager { background: rgba(22, 93, 255, 0.1); color: #165DFF; }
  &.editor { background: rgba(251, 146, 60, 0.1); color: #FB923C; }
  &.viewer { background: rgba(14, 165, 233, 0.1); color: #0EA5E9; }
  &.user { background: rgba(148, 163, 184, 0.1); color: #64748B; }
}

.permissions-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-btn {
  padding: 4px 10px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  margin-right: 6px;
  transition: all 0.2s;
  
  &.edit {
    background: rgba(22, 93, 255, 0.1);
    color: #165DFF;
    
    &:hover { background: rgba(22, 93, 255, 0.2); }
  }
  
  &.reset {
    background: rgba(251, 146, 60, 0.1);
    color: #FB923C;
    
    &:hover { background: rgba(251, 146, 60, 0.2); }
  }
  
  &.delete {
    background: rgba(245, 63, 63, 0.1);
    color: #F53F3F;
    
    &:hover { background: rgba(245, 63, 63, 0.2); }
  }
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 20px;
}

.page-btn {
  padding: 6px 16px;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 13px;
  color: #64748B;
  transition: all 0.2s;
  
  &:hover:not(:disabled) {
    border-color: #165DFF;
    color: #165DFF;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.page-info {
  font-size: 13px;
  color: #64748B;
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.permission-card {
  background: #F8FAFC;
  border-radius: 12px;
  padding: 16px;
  
  h3 {
    font-size: 14px;
    font-weight: 600;
    color: #1E293B;
    margin: 0 0 12px 0;
  }
}

.role-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.role-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: white;
  border-radius: 8px;
}

.role-desc {
  font-size: 12px;
  color: #94A3B8;
}

.permission-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 300px;
  overflow-y: auto;
}

.permission-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  background: white;
  border-radius: 6px;
  font-size: 12px;
  
  .perm-id {
    color: #94A3B8;
    min-width: 120px;
    font-family: monospace;
  }
  
  .perm-name {
    color: #1E293B;
    font-weight: 500;
    flex: 1;
  }
  
  .perm-desc {
    color: #94A3B8;
  }
}

.current-user-info {
  h3 {
    font-size: 14px;
    font-weight: 600;
    color: #1E293B;
    margin: 0 0 12px 0;
  }
}

.user-perm-card {
  background: #F8FAFC;
  border-radius: 12px;
  padding: 16px;
}

.user-perm-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  
  .user-name {
    font-size: 16px;
    font-weight: 600;
    color: #1E293B;
  }
  
  .user-role {
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;
  }
}

.user-perm-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.perm-tag {
  padding: 4px 10px;
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
  border-radius: 4px;
  font-size: 12px;
}

.no-perm {
  color: #94A3B8;
  font-size: 12px;
}

.feature-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  
  h3 {
    font-size: 16px;
    font-weight: 600;
    color: #1E293B;
    margin: 0;
  }
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.feature-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: #F8FAFC;
  border-radius: 10px;
}

.feature-info {
  display: flex;
  align-items: center;
  gap: 12px;
  
  .feature-name {
    font-size: 14px;
    font-weight: 600;
    color: #1E293B;
  }
  
  .feature-key {
    font-size: 12px;
    color: #94A3B8;
    font-family: monospace;
    background: white;
    padding: 2px 8px;
    border-radius: 4px;
  }
  
  .feature-desc {
    font-size: 12px;
    color: #94A3B8;
  }
}

.feature-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-btn {
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
  
  &:not(.enabled) {
    background: rgba(148, 163, 184, 0.1);
    color: #94A3B8;
  }
  
  &.enabled {
    background: rgba(0, 180, 42, 0.1);
    color: #00B42A;
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: #94A3B8;
  
  svg {
    margin-bottom: 12px;
  }
  
  span {
    font-size: 14px;
    margin-bottom: 4px;
  }
  
  p {
    font-size: 12px;
    margin: 0;
  }
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
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #E2E8F0;
  
  h3 {
    font-size: 16px;
    font-weight: 600;
    color: #1E293B;
    margin: 0;
  }
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #94A3B8;
  padding: 4px;
  
  &:hover {
    color: #64748B;
  }
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 16px;
  
  label {
    display: block;
    font-size: 13px;
    font-weight: 500;
    color: #475569;
    margin-bottom: 6px;
  }
}

.form-input, .form-select {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  
  &:focus {
    border-color: #165DFF;
  }
}

.permission-select {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.perm-checkbox {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
}

.hint-text {
  font-size: 12px;
  color: #F53F3F;
  margin: 0;
  padding: 8px;
  background: rgba(245, 63, 63, 0.05);
  border-radius: 6px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid #E2E8F0;
}

.btn {
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  
  &.btn-cancel {
    background: #F1F5F9;
    color: #64748B;
    
    &:hover {
      background: #E2E8F0;
    }
  }
  
  &.btn-confirm {
    background: #165DFF;
    color: white;
    
    &:hover {
      background: #1550D6;
    }
  }
}
</style>