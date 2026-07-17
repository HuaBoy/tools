// 设备检测 composable - 平台/屏幕/网络
import { ref, onMounted } from 'vue'

export function useDevice() {
  const platform = ref('web')
  const isMobile = ref(false)
  const isOnline = ref(navigator.onLine)
  const screenWidth = ref(window.innerWidth)

  function detect() {
    const ua = navigator.userAgent
    if (/miniprogram/i.test(ua)) platform.value = 'miniapp'
    else if (/android/i.test(ua)) platform.value = 'android'
    else if (/iphone|ipad/i.test(ua)) platform.value = 'ios'
    else platform.value = 'web'

    isMobile.value = platform.value === 'android' || platform.value === 'ios' || screenWidth.value <= 768
  }

  onMounted(() => {
    detect()
    window.addEventListener('resize', () => { screenWidth.value = window.innerWidth; detect() })
    window.addEventListener('online', () => { isOnline.value = true })
    window.addEventListener('offline', () => { isOnline.value = false })
  })

  return { platform, isMobile, isOnline, screenWidth }
}