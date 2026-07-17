// 硬件连接状态
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useHardwareStore = defineStore('hardware', () => {
  const bleConnected = ref(false)
  const bleDevice = ref(null)
  const mqttConnected = ref(false)
  const otaProgress = ref(0)
  const otaStatus = ref('')

  function setBLEConnected(device) { bleConnected.value = true; bleDevice.value = device }
  function setBLEDisconnected() { bleConnected.value = false; bleDevice.value = null }
  function setMQTTConnected(v) { mqttConnected.value = v }
  function setOTAProgress(p, s) { otaProgress.value = p; otaStatus.value = s }

  return { bleConnected, bleDevice, mqttConnected, otaProgress, otaStatus,
    setBLEConnected, setBLEDisconnected, setMQTTConnected, setOTAProgress }
})