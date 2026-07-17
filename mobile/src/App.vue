<script setup>
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'

onLaunch(() => {
  console.log('AI起爆运维执行平台 - 移动端启动')
  // 检测运行环境
  // #ifdef APP-PLUS
  console.log('运行环境: Android APK')
  // #endif
  // #ifdef MP-WEIXIN
  console.log('运行环境: 微信小程序')
  // #endif
})

onShow(() => { console.log('App Show') })
onHide(() => { console.log('App Hide') })
</script>

<style lang="scss">
@import '@dcloudio/uni-ui/lib/uni.scss';

page {
  background-color: #F5F7FA;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
</style>