// 原生能力桥接 - APP 端通过 JSBridge 调用原生功能
// Android: window.AndroidBridge.xxx()
// 微信小程序: wx.xxx()
// iOS(后期): window.webkit.messageHandlers.xxx()

const isAndroid = () => /android/i.test(navigator.userAgent)
const isMiniApp = () => /miniprogram/i.test(navigator.userAgent) || (typeof wx !== 'undefined' && wx.miniProgram)

export const nativeBridge = {
  // 扫码
  async scanCode() {
    if (isMiniApp()) return new Promise((resolve) => wx.scanCode({ success: r => resolve(r.result) }))
    if (isAndroid() && window.AndroidBridge) return window.AndroidBridge.scanCode()
    throw new Error('扫码功能仅支持 APP/小程序端')
  },

  // NFC 读取
  async readNFC() {
    if (isAndroid() && window.AndroidBridge) return window.AndroidBridge.readNFC()
    throw new Error('NFC 仅支持 Android APP 端')
  },

  // 蓝牙扫描
  async scanBLE(timeout = 5000) {
    if (isAndroid() && window.AndroidBridge) return window.AndroidBridge.scanBLE(timeout)
    throw new Error('蓝牙仅支持 APP 端')
  },

  // 获取定位
  async getLocation() {
    if (isMiniApp()) return new Promise((resolve, reject) => {
      wx.getLocation({ success: resolve, fail: reject })
    })
    return new Promise((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(
        pos => resolve({ lat: pos.coords.latitude, lng: pos.coords.longitude }),
        reject
      )
    })
  },

  // 拍照
  async takePhoto() {
    if (isMiniApp()) return new Promise((resolve) => {
      wx.chooseImage({ count: 1, sourceType: ['camera'], success: r => resolve(r.tempFilePaths[0]) })
    })
    if (isAndroid() && window.AndroidBridge) return window.AndroidBridge.takePhoto()
    throw new Error('拍照仅支持 APP/小程序端')
  },

  // 安装 APK（OTA 自更新）
  async installAPK(path) {
    if (isAndroid() && window.AndroidBridge) return window.AndroidBridge.installAPK(path)
    throw new Error('安装仅支持 Android APP 端')
  },

  // 获取设备信息
  async getDeviceInfo() {
    if (isMiniApp()) {
      return new Promise((resolve) => {
        wx.getSystemInfo({ success: resolve })
      })
    }
    return { platform: navigator.platform, userAgent: navigator.userAgent }
  }
}

// 挂载到全局，供 hardwareService 调用
window.__nativeBridge = nativeBridge