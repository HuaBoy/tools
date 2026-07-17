// 硬件服务 - 芯片/模块/设备/固件/OTA/BLE/MQTT
import { api } from './http'

const native = () => window.__nativeBridge || {}

export const hardwareService = {
  // === 设备 ===
  getDevices(params) { return api.get('/api/v1/hardware/devices', params) },
  registerDevice(data) { return api.post('/api/v1/hardware/devices', data) },
  getDeviceDetail(id) { return api.get('/api/v1/hardware/devices/' + id) },
  updateDevice(id, data) { return api.patch('/api/v1/hardware/devices/' + id, data) },

  // === 芯片 ===
  getChips(params) { return api.get('/api/v1/hardware/chips', params) },
  registerChip(data) { return api.post('/api/v1/hardware/chips', data) },
  bindChipToDevice(chipId, deviceId) {
    return api.post('/api/v1/hardware/chips/bind', { chipId, deviceId })
  },

  // === 模块 ===
  getModules(params) { return api.get('/api/v1/hardware/modules', params) },
  registerModule(data) { return api.post('/api/v1/hardware/modules', data) },

  // === 固件 ===
  getFirmwareVersions(deviceType) { return api.get('/api/v1/hardware/firmware', { deviceType }) },
  pushOTA(deviceId, version) { return api.post('/api/v1/hardware/firmware/ota', { deviceId, version }) },
  getOTAStatus(taskId) { return api.get('/api/v1/hardware/firmware/ota/' + taskId) },

  // === 诊断 ===
  runSelfTest(deviceId) { return api.post('/api/v1/hardware/devices/' + deviceId + '/selftest') },
  getDiagnosisResult(taskId) { return api.get('/api/v1/hardware/diagnosis/' + taskId) },

  // === BLE（APP 原生桥接） ===
  scanBLE(timeout = 5000) { return native().scanBLE?.(timeout) || Promise.reject(new Error('BLE 仅支持 APP 端')) },
  connectBLE(deviceId) { return native().connectBLE?.(deviceId) || Promise.reject(new Error('BLE 仅支持 APP 端')) },

  // === MQTT（APP 原生桥接） ===
  subscribeMQTT(topic, cb) { return native().mqttSubscribe?.(topic, cb) || Promise.reject(new Error('MQTT 仅支持 APP 端')) },
  publishMQTT(topic, msg) { return native().mqttPublish?.(topic, msg) || Promise.reject(new Error('MQTT 仅支持 APP 端')) }
}