import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
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
    port: 5173,
    open: false,
    proxy: {
      '/api/v1': {
        target: 'http://您的阿里云服务器IP:8080',
        changeOrigin: true,
        secure: false,
        ws: false
      },
      '/api': {
        target: 'https://mp.holyview.cn:9443',
        changeOrigin: true,
        secure: true,
        ws: true,
        headers: {
          'Authorization': 'Basic ' + Buffer.from('saber:saber_secret').toString('base64')
        }
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
      }
    }
  }
})
