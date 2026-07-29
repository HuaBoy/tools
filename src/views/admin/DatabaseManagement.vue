<script setup>
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import GlassCard from '@/components/GlassCard.vue';
import { useLogsStore } from '@/stores/logs';
import adminApi from '@/api/admin';

const logsStore = useLogsStore();

const databases = ref([]);
const isLoading = ref(false);
const searchName = ref('');
const filterType = ref('');

const showDialog = ref(false);
const editId = ref(null);
const submitting = ref(false);
const form = reactive({
  name: '',
  db_type: 'postgres',
  host: '',
  port: 5432,
  db_name: '',
  username: '',
  password: '',
  status: 1,
  remark: ''
});

const typeOptions = [
  { label: 'PostgreSQL', value: 'postgres' },
  { label: 'MySQL', value: 'mysql' },
  { label: 'SQLite', value: 'sqlite' }
];

const loadDatabases = async () => {
  isLoading.value = true;
  try {
    const res = await adminApi.getDatabases(searchName.value, filterType.value);
    databases.value = res.data || [];
  } catch (e) {
    ElMessage.error('加载服务器数据库失败: ' + e.message);
  } finally {
    isLoading.value = false;
  }
};

const handleSearch = () => loadDatabases();

const resetForm = () => {
  form.name = '';
  form.db_type = 'postgres';
  form.host = '';
  form.port = 5432;
  form.db_name = '';
  form.username = '';
  form.password = '';
  form.status = 1;
  form.remark = '';
};

const handleAdd = () => {
  editId.value = null;
  resetForm();
  showDialog.value = true;
};

const handleEdit = (db) => {
  editId.value = db.id;
  form.name = db.name;
  form.db_type = db.db_type || 'postgres';
  form.host = db.host || '';
  form.port = db.port || 0;
  form.db_name = db.db_name || '';
  form.username = db.username || '';
  form.password = '';
  form.status = db.status === 0 ? 0 : 1;
  form.remark = db.remark || '';
  showDialog.value = true;
};

const handleSave = async () => {
  if (!form.name.trim()) {
    ElMessage.warning('请输入数据库名称');
    return;
  }
  const payload = {
    name: form.name.trim(),
    db_type: form.db_type,
    host: form.host.trim(),
    port: Number(form.port) || 0,
    db_name: form.db_name.trim(),
    username: form.username.trim(),
    status: form.status ? 1 : 0,
    remark: form.remark.trim()
  };
  if (form.password) payload.password = form.password;
  submitting.value = true;
  try {
    if (editId.value) {
      await adminApi.updateDatabase(editId.value, payload);
      ElMessage.success('更新成功');
      logsStore.addLog('编辑', '服务器数据库', `更新数据库: ${form.name}`);
    } else {
      await adminApi.createDatabase(payload);
      ElMessage.success('创建成功');
      logsStore.addLog('创建', '服务器数据库', `创建数据库: ${form.name}`);
    }
    showDialog.value = false;
    loadDatabases();
  } catch (e) {
    ElMessage.error('保存失败: ' + e.message);
  } finally {
    submitting.value = false;
  }
};

const handleDelete = async (db) => {
  try {
    await ElMessageBox.confirm(`确定要删除数据库 "${db.name}" 吗？`, '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    });
    await adminApi.deleteDatabase(db.id);
    ElMessage.success('删除成功');
    logsStore.addLog('删除', '服务器数据库', `删除数据库: ${db.name}`);
    loadDatabases();
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败: ' + e.message);
  }
};

const handleToggleStatus = async (db, val) => {
  const next = val ? 1 : 0;
  try {
    await adminApi.updateDatabase(db.id, { status: next });
    ElMessage.success('状态已更新');
    logsStore.addLog('更新', '服务器数据库', `切换状态: ${db.name}`);
  } catch (e) {
    ElMessage.error('更新失败: ' + e.message);
    db.status = val ? 0 : 1;
  }
};

const connectionString = (db) => {
  if (!db.host) return '未配置连接';
  const port = db.port ? ':' + db.port : '';
  const dbn = db.db_name ? '/' + db.db_name : '';
  return `${db.host}${port}${dbn}`;
};

onMounted(loadDatabases);
</script>

<template>
  <GlassCard class="database-management">
    <div class="db-header">
      <div class="db-title">
        <h2>服务器数据库</h2>
        <p>管理服务器上的数据库实例连接信息（PostgreSQL / MySQL / SQLite）</p>
      </div>
      <button class="btn-primary" @click="handleAdd">+ 新增数据库</button>
    </div>

    <div class="content-panel">
      <div class="toolbar">
        <div class="search-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input v-model="searchName" @keyup.enter="handleSearch" placeholder="搜索数据库名称" />
        </div>
        <select v-model="filterType" class="filter-select" @change="handleSearch">
          <option value="">全部类型</option>
          <option v-for="t in typeOptions" :key="t.value" :value="t.value">{{ t.label }}</option>
        </select>
        <button class="btn-ghost" @click="handleSearch">查询</button>
        <button class="btn-ghost" @click="() => { searchName=''; filterType=''; handleSearch(); }">重置</button>
      </div>

      <div v-if="isLoading" class="loading">加载中...</div>
      <div v-else-if="databases.length" class="database-grid">
        <div v-for="db in databases" :key="db.id" class="db-card" :class="{ disabled: db.status === 0 }">
          <div class="db-card-top">
            <div class="db-icon">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg>
            </div>
            <div class="db-meta">
              <div class="db-name">{{ db.name }}</div>
              <span class="db-type-tag" :class="'type-' + (db.db_type || 'postgres')">{{ (db.db_type || 'postgres').toUpperCase() }}</span>
            </div>
          </div>
          <div class="db-conn">
            <div class="conn-row"><span class="conn-label">连接</span><span class="conn-value">{{ connectionString(db) }}</span></div>
            <div class="conn-row"><span class="conn-label">用户</span><span class="conn-value">{{ db.username || '—' }}</span></div>
            <div class="conn-row" v-if="db.remark"><span class="conn-label">备注</span><span class="conn-value">{{ db.remark }}</span></div>
          </div>
          <div class="db-card-bottom">
            <label class="switch-wrap">
              <span class="switch-label">{{ db.status === 0 ? '已禁用' : '已启用' }}</span>
              <input type="checkbox" :checked="db.status !== 0" @change="handleToggleStatus(db, $event.target.checked)" />
            </label>
            <div class="db-actions">
              <button class="action-link" @click="handleEdit(db)">编辑</button>
              <button class="action-link danger" @click="handleDelete(db)">删除</button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/></svg>
        <span>暂无数据库配置</span>
        <p>点击「新增数据库」添加服务器数据库连接</p>
      </div>
    </div>

    <el-dialog v-model="showDialog" :title="editId ? '编辑数据库' : '新增数据库'" width="520px">
      <div class="form-grid">
        <div class="form-item full">
          <label>名称 <span class="req">*</span></label>
          <input v-model="form.name" placeholder="如：生产主库" />
        </div>
        <div class="form-item">
          <label>类型</label>
          <select v-model="form.db_type">
            <option v-for="t in typeOptions" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </div>
        <div class="form-item">
          <label>状态</label>
          <label class="switch-inline">
            <input type="checkbox" v-model="form.status" :true-value="1" :false-value="0" />
            <span>{{ form.status ? '启用' : '禁用' }}</span>
          </label>
        </div>
        <div class="form-item">
          <label>主机</label>
          <input v-model="form.host" placeholder="127.0.0.1" />
        </div>
        <div class="form-item">
          <label>端口</label>
          <input v-model="form.port" type="number" placeholder="5432" />
        </div>
        <div class="form-item">
          <label>数据库名</label>
          <input v-model="form.db_name" placeholder="postgres" />
        </div>
        <div class="form-item">
          <label>用户名</label>
          <input v-model="form.username" placeholder="postgres" />
        </div>
        <div class="form-item">
          <label>密码</label>
          <input v-model="form.password" type="password" :placeholder="editId ? '留空则不修改' : '请输入密码'" />
        </div>
        <div class="form-item full">
          <label>备注</label>
          <textarea v-model="form.remark" rows="2" placeholder="可选备注"></textarea>
        </div>
      </div>
      <template #footer>
        <button class="btn-ghost" @click="showDialog = false">取消</button>
        <button class="btn-primary" :disabled="submitting" @click="handleSave">{{ submitting ? '保存中...' : '保存' }}</button>
      </template>
    </el-dialog>
  </GlassCard>
</template>

<style scoped>
.database-management { width: 100%; }

.db-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 16px; margin-bottom: 16px; flex-wrap: wrap;
}
.db-title h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; }
.db-title p { font-size: 13px; color: var(--text-tertiary); margin: 4px 0 0; }

.content-panel {
  background: #ffffff; border: 1px solid #E2E8F0; border-radius: 12px;
  overflow: hidden; padding: 16px;
}

.toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; flex-wrap: wrap; }
.search-box {
  display: flex; align-items: center; gap: 8px; flex: 1; min-width: 200px;
  background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 12px;
}
.search-box input { border: none; background: transparent; outline: none; width: 100%; color: var(--text-primary); font-size: 13px; }
.filter-select, .btn-ghost, .btn-primary {
  height: 36px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid var(--border-color);
}
.filter-select { padding: 0 10px; background: var(--bg-input); color: var(--text-primary); }
.btn-ghost { padding: 0 16px; background: var(--bg-input); color: var(--text-secondary); }
.btn-ghost:hover { border-color: #165DFF; color: #165DFF; }
.btn-primary {
  padding: 0 18px; background: #165DFF; color: #fff; border: none; font-weight: 600;
}
.btn-primary:hover { background: #0F4CD0; }
.btn-primary:disabled { opacity: .6; cursor: not-allowed; }

.loading { text-align: center; padding: 40px; color: var(--text-tertiary); }

.database-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }
.db-card {
  display: flex; flex-direction: column; gap: 12px; padding: 16px;
  background: var(--bg-input); border: 1px solid var(--border-color); border-radius: 10px; transition: all .2s;
}
.db-card:hover { border-color: rgba(22,93,255,.4); box-shadow: 0 4px 14px rgba(22,93,255,.12); }
.db-card.disabled { opacity: .55; }
.db-card-top { display: flex; align-items: center; gap: 12px; }
.db-icon {
  width: 40px; height: 40px; border-radius: 10px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  background: rgba(22,93,255,.1); color: #165DFF;
}
.db-meta { flex: 1; min-width: 0; }
.db-name { font-size: 15px; font-weight: 600; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.db-type-tag { font-size: 10px; padding: 1px 6px; border-radius: 6px; font-weight: 700; }
.type-postgres { background: rgba(22,93,255,.12); color: #165DFF; }
.type-mysql { background: rgba(76,175,80,.14); color: #2E7D32; }
.type-sqlite { background: rgba(255,153,0,.14); color: #B26A00; }

.db-conn { display: flex; flex-direction: column; gap: 6px; padding: 10px 0; border-top: 1px solid var(--border-color); border-bottom: 1px solid var(--border-color); }
.conn-row { display: flex; gap: 8px; font-size: 12px; }
.conn-label { color: var(--text-tertiary); min-width: 32px; }
.conn-value { color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.db-card-bottom { display: flex; align-items: center; justify-content: space-between; gap: 8px; flex-wrap: wrap; }
.switch-wrap { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--text-secondary); cursor: pointer; }
.switch-label { white-space: nowrap; }
.db-actions { display: flex; gap: 12px; }
.action-link { background: none; border: none; color: #165DFF; cursor: pointer; font-size: 13px; padding: 0; }
.action-link:hover { text-decoration: underline; }
.action-link.danger { color: #F53F3F; }

.empty-state { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 50px 0; color: var(--text-tertiary); }
.empty-state p { font-size: 12px; margin: 0; }

/* 对话框表单 */
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-item { display: flex; flex-direction: column; gap: 6px; }
.form-item.full { grid-column: 1 / -1; }
.form-item label { font-size: 12px; color: var(--text-secondary); }
.form-item .req { color: #F53F3F; }
.form-item input, .form-item select, .form-item textarea {
  border: 1px solid var(--border-color); border-radius: 8px; padding: 8px 10px;
  background: var(--bg-input); color: var(--text-primary); font-size: 13px; outline: none; resize: vertical;
}
.form-item input:focus, .form-item select:focus, .form-item textarea:focus { border-color: #165DFF; }
.switch-inline { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text-secondary); }

@media screen and (max-width: 560px) {
  .form-grid { grid-template-columns: 1fr; }
  .database-grid { grid-template-columns: 1fr; }
}
</style>
