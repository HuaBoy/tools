import { createClient } from '@supabase/supabase-js';

const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || '';
const supabaseKey = import.meta.env.VITE_SUPABASE_ANON_KEY || '';

// 判断是否配置了真实的 Supabase（同时支持官方云与自托管实例）
const isSupabaseConfigured = !!(
  supabaseUrl &&
  /^https?:\/\//.test(supabaseUrl) &&
  !supabaseUrl.includes('your-project') &&
  !supabaseUrl.includes('your_project') &&
  supabaseKey &&
  !supabaseKey.includes('your-anon-key') &&
  !supabaseKey.includes('your_anon_key') &&
  !supabaseKey.includes('请从Supabase') &&
  supabaseKey.length > 20
);

// 调试信息（生产环境会自动去除）
if (typeof window !== 'undefined') {
  console.log('[Supabase] 配置状态:', isSupabaseConfigured ? '已配置' : '未配置（本地模式）');
  if (!isSupabaseConfigured) {
    console.log('[Supabase] 提示：使用本地账号登录，默认密码：123456');
  }
}

let supabaseClient = null;

if (isSupabaseConfigured) {
  try {
    supabaseClient = createClient(supabaseUrl, supabaseKey, {
      auth: {
        persistSession: true,
        autoRefreshToken: true
      }
    });
  } catch (e) {
    console.warn('Supabase 初始化失败:', e.message);
    supabaseClient = null;
  }
}

// 链式查询的安全代理：任何方法都返回一个可链式调用且最终返回安全 Promise 的对象
const createSafeQuery = () => {
  const handler = {
    get(_target, prop) {
      if (prop === 'then') return undefined; // 避免被误判为 thenable
      // 终结方法：返回安全 Promise
      if (prop === 'single' || prop === 'maybeSingle') {
        return () => Promise.resolve({ data: null, error: { message: 'Supabase 未配置' } });
      }
      // 链式方法：返回新的代理对象
      return new Proxy(function() {}, handler);
    },
    apply() {
      return new Proxy(function() {}, handler);
    }
  };
  return new Proxy(function() {}, handler);
};

// 提供一个安全的代理，当未配置或调用失败时返回友好的错误
const safeSupabase = supabaseClient || {
  auth: {
    signInWithPassword: async () => ({
      data: { user: null, session: null },
      error: { message: 'Supabase 未配置，请使用本地账号（密码 123456）登录' }
    }),
    signUp: async () => ({
      data: { user: null, session: null },
      error: { message: 'Supabase 未配置' }
    }),
    signOut: async () => ({ error: null }),
    signInWithOAuth: async () => ({ data: null, error: { message: 'Supabase 未配置' } }),
    resetPasswordForEmail: async () => ({ data: null, error: { message: 'Supabase 未配置' } }),
    getUser: async () => ({ data: { user: null }, error: null }),
    getSession: async () => ({ data: { session: null }, error: null }),
    onAuthStateChange: () => ({
      data: { subscription: { unsubscribe: () => {} } }
    }),
    updateUser: async () => ({ data: { user: null }, error: { message: 'Supabase 未配置' } })
  },
  from: () => createSafeQuery(),
  storage: {
    from: () => ({
      upload: async () => ({ data: null, error: { message: 'Supabase 未配置' } }),
      download: async () => ({ data: null, error: { message: 'Supabase 未配置' } }),
      getPublicUrl: () => ({ data: { publicUrl: '' } })
    })
  },
  rpc: async () => ({ data: null, error: { message: 'Supabase 未配置' } })
};

export const supabase = safeSupabase;
export const isSupabaseReady = isSupabaseConfigured;
export default safeSupabase;
