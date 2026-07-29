<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useAuthStore } from '@/stores/auth';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const authStore = useAuthStore();
const logsStore = useLogsStore();

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
  nickname: '',
  password: '',
  confirmPassword: '',
  role: null,        // 角色 ID
  status: 1
});

const roles = ref([]);

const roleOptions = computed(() => {
  return roles.value.map(r => ({ label: r.name, value: r.id }));
});

const loadUsers = async () => {
  isLoadingUsers.value = true;
  try {
    const result = await adminApi.getUsers(userCurrentPage.value, userPageSize.value, userSearchForm.search, userSearchForm.role);
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

const handleAddUser = () => {
  editUserId.value = null;
  userForm.username = '';
  userForm.email = '';
  userForm.nickname = '';
  userForm.password = '';
  userForm.confirmPassword = '';
  userForm.role = roles.value.length ? roles.value[0].id : null;
  userForm.status = 1;
  showUserForm.value = true;
};

const handleEditUser = (user) => {
  editUserId.value = user.id;
  userForm.username = user.username;
  userForm.email = user.email || '';
  userForm.nickname = user.nickname || '';
  userForm.password = '';
  userForm.confirmPassword = '';
  userForm.role = user.role_id || null;
  userForm.status = user.status === undefined ? 1 : user.status;
  showUserForm.value = true;
};

const handleSaveUser = async () => {
  if (!userForm.username.trim()) {
    ElMessage.warning('请输入用户名');
    return;
  }
  if (!userForm.role) {
    ElMessage.warning('请选择角色');
    return;
  }

  if (!editUserId.value) {
    if (!userForm.password.trim()) {
      ElMessage.warning('请输入密码');
      return;
    }
    if (userForm.password !== userForm.confirmPassword) {
      ElMessage.warning('两次输入的密码不一致');
      return;
    }
  }

  try {
    if (editUserId.value) {
      const updates = {
        nickname: userForm.nickname || userForm.username,
        email: userForm.email,
        role_id: userForm.role,
        status: userForm.status
      };
      if (userForm.password.trim()) {
        updates.password = userForm.password;
      }
      await adminApi.updateUser(editUserId.value, updates);
      ElMessage.success('用户更新成功');
      logsStore.addLog('编辑', '用户管理', `更新用户: ${userForm.username}`);
    } else {
      await adminApi.createUser({
        username: userForm.username,
        password: userForm.password,
        nickname: userForm.nickname || userForm.username,
        email: userForm.email,
        role_id: userForm.role,
        status: userForm.status
      });
      ElMessage.success('用户创建成功');
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
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
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
    const { value: newPassword } = await ElMessageBox.prompt(
      `请输入用户 "${user.username}" 的新密码：`,
      '重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info',
        inputType: 'password',
        inputPattern: /^.{6,}$/,
        inputErrorMessage: '密码长度至少6位'
      }
    );
    await adminApi.resetPassword(user.id, newPassword);
    ElMessage.success('密码重置成功');
    logsStore.addLog('重置密码', '用户管理', `重置密码: ${user.username}`);
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重置密码失败: ' + error.message);
    }
  }
};

const handleSearchUsers = () => {
  userCurrentPage.value = 1;
  loadUsers();
};

const handlePageChange = (page) => {
  userCurrentPage.value = page;
  loadUsers();
};

const getRoleName = (roleId) => {
  const role = roles.value.find(r => r.id === roleId);
  return role ? role.name : roleId;
};

const getRoleClass = (roleId) => {
  if (roleId === 1) return 'admin';
  if (roleId === 2) return 'manager';
  if (roleId === 3) return 'editor';
  if (roleId === 4) return 'viewer';
  return 'user';
};

const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('zh-CN');
};

onMounted(async () => {
  // 确保认证状态已加载
  if (!authStore.isLoggedIn) {
    try {
      await authStore.loadAuthState();
    } catch (e) {
      console.warn('加载认证状态失败:', e);
    }
  }

  if (!authStore.isLoggedIn) {
    ElMessage.error('请先登录');
    return;
  }

  if (!authStore.hasPermission('user:view')) {
    ElMessage.error('您没有访问用户管理的权限');
    return;
  }
  loadUsers();
  loadRoles();
});
</script>

<template>
  <div class="user-management">
    <GlassCard title="用户管理">
      <div class="content-panel">
        <div class="toolbar">
        <div class="search-area">
          <input
            v-model="userSearchForm.search"
            type="text"
            class="search-input"
            placeholder="搜索用户名或邮箱"
            @keyup.enter="handleSearchUsers"
          />
          <select v-model="userSearchForm.role" class="role-select">
            <option value="">全部角色</option>
            <option v-for="r in roleOptions" :key="r.value" :value="r.value">
              {{ r.label }}
            </option>
          </select>
          <button class="search-btn" @click="handleSearchUsers">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <span>搜索</span>
          </button>
        </div>
        <button class="add-btn" @click="handleAddUser" v-if="authStore.hasPermission('user:create')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>添加用户</span>
        </button>
      </div>

      <div v-if="isLoadingUsers" class="loading-state">加载中...</div>

      <div v-else class="user-table">
        <table>
          <thead>
            <tr>
              <th>用户名</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>{{ user.username }}</td>
              <td>{{ user.email || '-' }}</td>
              <td>
                <span class="role-tag" :class="'role-' + (getRoleClass(user.role_id))">
                  {{ user.role_name || getRoleName(user.role_id) }}
                </span>
              </td>
              <td>
                <span class="status-tag" :class="user.status === 1 ? 'enabled' : 'disabled'">
                  {{ user.status === 1 ? '启用' : '禁用' }}
                </span>
              </td>
              <td>{{ formatDate(user.created_at) }}</td>
              <td class="actions">
                <button class="action-link" @click="handleEditUser(user)" v-if="authStore.hasPermission('user:edit')">编辑</button>
                <button class="action-link warn" @click="handleResetPassword(user)">重置密码</button>
                <button class="action-link danger" @click="handleDeleteUser(user)" v-if="authStore.hasPermission('user:delete')">删除</button>
              </td>
            </tr>
            <tr v-if="users.length === 0">
              <td colspan="6" class="empty-row">暂无用户数据</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="userTotal > userPageSize" class="pagination">
        <button
          class="page-btn"
          :disabled="userCurrentPage <= 1"
          @click="handlePageChange(userCurrentPage - 1)"
        >上一页</button>
        <span class="page-info">{{ userCurrentPage }} / {{ Math.ceil(userTotal / userPageSize) }}</span>
        <button
          class="page-btn"
          :disabled="userCurrentPage >= Math.ceil(userTotal / userPageSize)"
          @click="handlePageChange(userCurrentPage + 1)"
        >下一页</button>
      </div>
      </div>
    </GlassCard>

    <Teleport to="body">
      <div v-if="showUserForm" class="modal-overlay" @click.self="showUserForm = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>{{ editUserId ? '编辑用户' : '添加用户' }}</h3>
            <button class="modal-close" @click="showUserForm = false">×</button>
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
              <label>昵称</label>
              <input v-model="userForm.nickname" type="text" class="form-input" placeholder="请输入昵称" />
            </div>
            <div class="form-group">
              <label>{{ editUserId ? '新密码（不填则保持不变）' : '密码' }}</label>
              <input v-model="userForm.password" type="password" class="form-input" placeholder="请输入密码" />
            </div>
            <div v-if="!editUserId" class="form-group">
              <label>确认密码</label>
              <input v-model="userForm.confirmPassword" type="password" class="form-input" placeholder="请再次输入密码" />
            </div>
            <div class="form-group">
              <label>角色</label>
              <select v-model="userForm.role" class="form-input">
                <option v-for="r in roleOptions" :key="r.value" :value="r.value">
                  {{ r.label }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>状态</label>
              <select v-model="userForm.status" class="form-input">
                <option :value="1">启用</option>
                <option :value="0">禁用</option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showUserForm = false">取消</button>
            <button class="btn btn-primary" @click="handleSaveUser">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.user-management {
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

.search-area {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.search-input,
.role-select {
  padding: 8px 12px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.search-input {
  min-width: 200px;
}

.search-input:focus,
.role-select:focus {
  border-color: rgba(22, 93, 255, 0.6);
}

.search-btn,
.add-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.search-btn {
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
}

.search-btn:hover {
  background: rgba(22, 93, 255, 0.2);
}

.add-btn {
  background: linear-gradient(135deg, #165DFF 0%, #0F4CD0 100%);
  color: white;
}

.add-btn:hover {
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
}

.loading-state {
  padding: 60px 0;
  text-align: center;
  color: var(--text-tertiary);
}

.user-table {
  overflow-x: auto;
}

.user-table table {
  width: 100%;
  border-collapse: collapse;
}

.user-table th,
.user-table td {
  padding: 12px 16px;
  text-align: left;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
}

.user-table th {
  color: var(--text-tertiary);
  font-weight: 500;
  background: rgba(22, 93, 255, 0.04);
}

.user-table td {
  color: var(--text-secondary);
}

.user-table tbody tr:hover {
  background: rgba(22, 93, 255, 0.04);
}

.empty-row {
  text-align: center;
  color: var(--text-tertiary);
  padding: 40px !important;
}

.role-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.role-tag.role-admin {
  background: rgba(245, 63, 63, 0.1);
  color: #F53F3F;
}

.role-tag.role-manager {
  background: rgba(255, 125, 0, 0.1);
  color: #FF7D00;
}

.role-tag.role-editor {
  background: rgba(22, 93, 255, 0.1);
  color: #165DFF;
}

.role-tag.role-viewer {
  background: rgba(100, 116, 139, 0.1);
  color: #64748B;
}

.role-tag.role-user {
  background: rgba(0, 180, 42, 0.1);
  color: #00B42A;
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

.action-link.warn {
  color: #FF7D00;
}

.action-link.danger {
  color: #F53F3F;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
}

.page-btn {
  padding: 6px 12px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 13px;
  color: var(--text-secondary);
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
  max-width: 600px;
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
  display: flex;
  align-items: center;
  justify-content: center;
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
}

.form-input:focus {
  border-color: #165DFF;
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
  padding: 8px;
  background: #F8FAFC;
  border-radius: 6px;
}

.permission-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  cursor: pointer;
  color: #475569;
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
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-area {
    flex-direction: column;
  }

  .search-input {
    min-width: 0;
    width: 100%;
  }

  .permissions-grid {
    grid-template-columns: 1fr;
  }

  .user-table th,
  .user-table td {
    padding: 8px 10px;
    font-size: 12px;
  }
}
</style>
