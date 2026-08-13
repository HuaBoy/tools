import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // 默认指向生产服务器；本地测试时: $env:VITE_PROXY_TARGET="http://localhost:8080"; npm run dev
  const apiTarget = env.VITE_PROXY_TARGET || 'http://111.231.16.110'

  return {
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@import "@/assets/styles/variables.scss";'
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    open: false,
    proxy: {
      '/socket.io': {
        target: 'http://localhost:9528',
        changeOrigin: true,
        ws: true
      },
      // 主系统后端（go-server），默认生产，本地测试用 VITE_PROXY_TARGET 覆盖
      '/api/v1': {
        target: apiTarget,
        changeOrigin: true,
        secure: false,
        ws: false,
        proxyTimeout: 600000,
        // 覆盖 Origin 为白名单内地址，避免后端 CORS 对 localhost:3000 返回 403（本地开发用）
        headers: { Origin: 'http://111.231.16.110' }
      },
      // 盛景平台登录（第三方 holyview），需保留，且必须在 /api 之前匹配
      '/api/blade-auth': {
        target: 'https://mp.holyview.cn:9443',
        changeOrigin: true,
        secure: true,
        ws: true,
        headers: {
          'Authorization': 'Basic ' + Buffer.from('saber:saber_secret').toString('base64')
        }
      },
      // 主系统后端（go-server），经服务器 Caddy:80 转发到 8080
      '/api': {
        target: apiTarget,
        changeOrigin: true,
        secure: false,
        ws: false,
        headers: { Origin: 'http://111.231.16.110' }
      },
      '/smart-api': {
        target: 'http://218.90.146.230:20001',
        changeOrigin: true,
        secure: false,
        ws: false,
        rewrite: (path) => path.replace(/^\/smart-api/, ''),
        headers: {
          'Authorization': 'Basic ' + Buffer.from('saber_identity_client:saber_identity_secret').toString('base64')
        }
      },
      '/iot-api': {
        target: 'http://218.90.146.230:20001',
        changeOrigin: true,
        secure: false,
        ws: false,
        rewrite: (path) => path.replace(/^\/iot-api/, ''),
        headers: {
          'Authorization': 'Basic ' + Buffer.from('saber:saber_secret').toString('base64')
        }
      },
      '/big-screen': {
        target: 'https://mp.holyview.cn:9443',
        changeOrigin: true,
        secure: false,
        ws: true,
        rewrite: (path) => path.replace(/^\/big-screen/, '')
      },
      // 代理 bigScreen2 子资源（JS/CSS/图片等），使 iframe 同源加载
      '/bigScreen2': {
        target: 'https://mp.holyview.cn:9443',
        changeOrigin: true,
        secure: false,
        ws: true
      }
    }
  }
  }
})
