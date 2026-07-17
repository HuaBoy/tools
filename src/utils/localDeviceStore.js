/**
 * 设备版本履历 - 本地存储
 *
 * 数据结构（单台设备）：
 * {
 *   id: string
 *   name: string                          // 设备名称（Tab 标题）
 *   intro: {                             // 设备简介
 *     background: string,                // 背景
 *     functions: string,                 // 功能
 *     metrics: string,                   // 指标
 *     photos: string[]                  // 照片（dataURL / 路径）
 *   },
 *   manualPdf: string,                // 使用说明书（内置 pdf 路径 / URL）
 *   hwVersions: [ { version, date, notes: string[] } ],  // 硬件版本 - 版本更新说明清单
 *   swVersions: [ { version, date, notes: string[] } ],  // 软件版本 - 版本更新说明清单
 *   hwPdf: string,                    // 硬件资料（内置 pdf）
 *   swPdf: string,                    // 软件资料（内置 pdf）
 *   structPdf: string,                // 结构资料（内置 pdf）
 *   testPdf: string,                  // 测试规范（内置 pdf 测试文件）
 *   testFlow: string,                 // 测试流程图（内置图片）
 *   testDocs: string                  // 测试规范 - 资料输入框
 * }
 *
 * 沿用本地存储模式，无需后端。订阅式实时刷新。
 */

const STORAGE_KEY = 'version_history_devices'

// 示例设备（首次加载时种入，可被用户修改/删除）
const SEED_DEVICES = [
  {
    id: 'dev-seed-l',
    name: '小勇士设备（L 系列）',
    intro: {
      background: '小勇士系列为便携式单兵起爆终端，面向野外/井下等复杂环境，强调轻量化与高可靠性。',
      functions: '支持蓝牙近场控制、延时精确起爆、状态自检与异常告警；可配合 DT40 主站协同组网。',
      metrics: '工作电压 12V DC｜最大输出电流 5A｜起爆通道 4 通道｜延时精度 ±1ms｜工作温度 -20°C ~ +50°C',
      photos: []
    },
    manualPdf: '',
    hwVersions: [
      { version: 'HW-V1.2', date: '2025-11-10', notes: ['优化外壳防护等级至 IP67', '更换低功耗电源管理芯片'] },
      { version: 'HW-V1.0', date: '2025-06-01', notes: ['首版硬件定型', '完成 EMC 摸底测试'] }
    ],
    swVersions: [
      { version: 'L.1.2.11T2', date: '2026-06-20', notes: ['新增 AI 日志智能识别', '修复曲线顶部遮挡问题'] },
      { version: 'L.1.1.0', date: '2026-03-15', notes: ['新增蓝牙快速配对', '优化低电量策略'] }
    ],
    hwPdf: '',
    swPdf: '',
    structPdf: '',
    testPdf: '',
    testFlow: '',
    testDocs: ''
  },
  {
    id: 'dev-seed-i',
    name: 'DT40 设备（I 系列）',
    intro: {
      background: 'DT40 系列为多通道集中式起爆主站，适用于矿山、隧道等大规模爆破场景。',
      functions: '支持 8 通道并行控制、远程无线通信、AES 加密传输与实时状态监控。',
      metrics: '工作电压 24V DC｜最大输出电流 10A｜起爆通道 8 通道｜延时精度 ±0.5ms｜通信距离 5km',
      photos: []
    },
    manualPdf: '',
    hwVersions: [
      { version: 'HW-V2.1', date: '2026-01-20', notes: ['增加冗余电源模块', '优化散热结构'] }
    ],
    swVersions: [
      { version: 'I.2.5.0', date: '2026-05-15', notes: ['新增多语言翻译', '优化通信重连机制'] },
      { version: 'I.2.0.0', date: '2026-01-10', notes: ['系统重构升级', '新增数据查询模块'] }
    ],
    hwPdf: '',
    swPdf: '',
    structPdf: '',
    testPdf: '',
    testFlow: '',
    testDocs: ''
  }
]

const listeners = new Set()

function read() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : null
  } catch (e) {
    console.warn('读取设备数据失败:', e)
    return null
  }
}

function write(list) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(list))
  listeners.forEach((fn) => fn(list))
}

function ensureSeed() {
  const existing = read()
  if (existing === null) {
    write(SEED_DEVICES)
    return [...SEED_DEVICES]
  }
  return existing
}

export const localDeviceStore = {
  getAll() {
    return ensureSeed()
  },
  getById(id) {
    return ensureSeed().find((d) => d.id === id) || null
  },
  save(device) {
    const list = ensureSeed()
    const idx = list.findIndex((d) => d.id === device.id)
    if (idx >= 0) list[idx] = device
    else list.push(device)
    write(list)
    return device
  },
  remove(id) {
    const list = ensureSeed().filter((d) => d.id !== id)
    write(list)
  },
  create(name) {
    const device = {
      id: 'dev-' + Date.now(),
      name: name || '新建设备',
      intro: { background: '', functions: '', metrics: '', photos: [] },
      manualPdf: '',
      hwVersions: [],
      swVersions: [],
      hwPdf: '',
      swPdf: '',
      structPdf: '',
      testPdf: '',
      testFlow: '',
      testDocs: ''
    }
    write([...ensureSeed(), device])
    return device
  },
  subscribe(fn) {
    listeners.add(fn)
    return () => listeners.delete(fn)
  }
}
