import { createClient } from '@supabase/supabase-js';

const supabaseUrl = 'http://106.14.191.34:8000';
const supabaseKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE';

const supabase = createClient(supabaseUrl, supabaseKey);

async function testConnection() {
  console.log('正在测试Supabase连接...');
  try {
    const { data, error } = await supabase.from('information_schema.tables').select('table_name').eq('table_schema', 'public');
    if (error) {
      console.log('查询表列表失败:', error.message);
      return null;
    }
    console.log('现有表:', data.map(t => t.table_name));
    return data;
  } catch (e) {
    console.log('连接异常:', e.message);
    return null;
  }
}

async function createUserTable() {
  console.log('尝试创建users表...');
  try {
    const { data, error } = await supabase.rpc('create_users_table', {});
    if (error) {
      console.log('RPC调用失败:', error.message);
      return false;
    }
    console.log('表创建成功');
    return true;
  } catch (e) {
    console.log('RPC异常:', e.message);
    return false;
  }
}

async function insertSampleData() {
  console.log('插入示例数据...');
  const users = [
    {
      username: 'admin',
      email: 'admin@example.com',
      password: '123456',
      role: 'admin',
      permissions: JSON.stringify(['user:view', 'user:create', 'user:edit', 'user:delete', 'role:assign', 'permission:manage', 'feature:toggle', 'data:query', 'data:export', 'log:view']),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    },
    {
      username: 'testuser',
      email: 'test@example.com',
      password: '123456',
      role: 'user',
      permissions: JSON.stringify(['data:query', 'data:export']),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
  ];
  
  try {
    const { data, error } = await supabase.from('users').insert(users);
    if (error) {
      console.log('插入数据失败:', error.message);
      return false;
    }
    console.log('数据插入成功:', data);
    return true;
  } catch (e) {
    console.log('插入异常:', e.message);
    return false;
  }
}

async function main() {
  await testConnection();
  await createUserTable();
  
  try {
    const { data, error } = await supabase.from('users').select('*');
    if (error) {
      console.log('查询users表失败:', error.message);
      return;
    }
    console.log('users表数据:', data);
  } catch (e) {
    console.log('查询异常:', e.message);
  }
}

main();
