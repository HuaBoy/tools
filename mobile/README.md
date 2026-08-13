# AI起爆运维执行平台 - 安卓 APP

<div align="center">

**AI起爆运维移动端（Android APK + 微信小程序）**

![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D)
![UniApp](https://img.shields.io/badge/UniApp-3.0-2F54EB)
![Android](https://img.shields.io/badge/Android-9.0+-3DDC84)
![License](https://img.shields.io/badge/License-Proprietary-red)

</div>

## 📖 项目简介

**AI起爆运维执行平台**移动端，面向起爆作业现场的一线运维人员。通过自然语言与 AI 交互，覆盖设备诊断、起爆数据查询、硬件管理、知识库检索等核心场景，一套代码可同时编译为 **Android APK**、**iOS App**、**微信小程序** 与 **H5**。

| 端 | 支持 | 说明 |
|----|------|------|
| Android | ✅ | 原生 APK，minSdk 24（Android 7.0+），包名 `com.ai.blasting` |
| iOS | ✅ | 支持蓝牙/NFC/相机权限配置 |
| 微信小程序 | ✅ | 无需安装，扫码即用 |
| H5 | ✅ | 浏览器访问 |

## ✨ 核心功能

### 🏠 AI 工作台（首页）
- **AI 自然语言交互**：输入需求即可执行（如"诊断批次号 B20240701-003"）
- **常用 AI 场景卡片**：一键进入高频操作
- **快速工具**：常用功能快捷入口
- **硬件管理入口**：芯片 · 模块 · 固件 · BLE

### 🔧 硬件管理
| 模块 | 说明 |
|------|------|
| 硬件管理 | 起爆器设备列表与状态 |
| 芯片管理 | 芯片信息管理 |
| 固件升级 | 固件版本管理与 OTA 升级 |

### 其他模块
| 页面 | 说明 |
|------|------|
| 智能诊断 | 起爆器故障智能诊断 |
| AI 起爆数据 | 起爆任务数据查询 |
| AI 知识库 | 爆破知识智能检索 |
| 通用工具 | 现场工具集 |
| 系统管理 | 个人中心 / 系统设置 |
| 登录 | 账号密码登录（后端 JWT） |

### 🔌 硬件能力（原生模块）
- **蓝牙（BLE）**：连接起爆器设备通信
- **NFC**：设备识别与读写
- **OTA**：固件空中升级
- **相机**：扫描设备二维码
- **安装包自升级**：`REQUEST_INSTALL_PACKAGES` 权限支持应用内更新

## 📱 页面结构

```
src/
├── pages/
│   ├── index/index.vue          # AI 工作台（首页）
│   ├── login/login.vue          # 登录
│   ├── diagnosis/index.vue      # 智能诊断
│   ├── data/index.vue           # AI 起爆数据
│   ├── hardware/
│   │   ├── index.vue            # 硬件管理
│   │   ├── chip/index.vue       # 芯片管理
│   │   └── firmware/index.vue   # 固件升级
│   ├── knowledge/index.vue      # AI 知识库
│   ├── tools/index.vue          # 通用工具
│   └── admin/index.vue          # 系统管理
├── components/                  # 公共组件
├── services/                    # API 服务层
├── stores/                      # Pinia 状态管理
├── utils/                       # 工具函数
├── static/                      # 静态资源（tabBar 图标等）
├── App.vue                      # 应用入口
├── main.js                      # Vue 实例
├── manifest.json                # 应用配置（权限/包名/版本）
└── pages.json                   # 路由与 tabBar 配置
```

### 底部 TabBar
**AI工作台 · 诊断 · 硬件 · 工具 · 我的**（5 个 Tab，选中态主色 `#165DFF`）

## 🚀 快速开始

### 环境要求
- Node.js 16+
- HBuilderX（推荐，用于打包 APK）或 Vue CLI
- 微信开发者工具（小程序调试）

### 安装依赖

```bash
npm install
```

### 开发调试

```bash
# H5 浏览器调试
npm run dev:h5

# 微信小程序调试
npm run dev:mp-weixin
```

### 打包发布

```bash
# 编译 Android APK（生成 App 资源，配合 HBuilderX 云打包）
npm run build:app

# 编译微信小程序
npm run build:mp-weixin

# 编译 H5
npm run build:h5
```

> **APK 打包方式**：运行 `npm run build:app` 后，将 `dist/build/app` 目录用 HBuilderX 打开，通过"发行 → 原生App-云打包"生成 APK；或使用本地离线打包（需 Android SDK）。

## 📄 应用配置（manifest.json）

### Android 配置
| 项 | 值 |
|----|----|
| 包名 | `com.ai.blasting` |
| minSdk | 24（Android 7.0） |
| targetSdk | 33（Android 13） |
| ABI | armeabi-v7a / arm64-v8a |
| 版本 | 1.0.0（versionCode 100） |

### Android 权限
```xml
BLUETOOTH / BLUETOOTH_ADMIN / BLUETOOTH_SCAN / BLUETOOTH_CONNECT
ACCESS_FINE_LOCATION（蓝牙发现）
NFC / CAMERA / REQUEST_INSTALL_PACKAGES
```

### iOS 隐私描述
蓝牙使用、蓝牙外设、相机扫码、定位（蓝牙发现）

## 🛠 技术栈

- **框架**：UniApp 3.x（Vue 3 Composition API）
- **状态管理**：Pinia 2.x
- **UI 组件**：uni-ui
- **样式**：SCSS
- **构建**：Vite 5 + `@dcloudio/vite-plugin-uni`
- **API**：uni.request（对接 AI起爆后端 Go 服务）

## 📦 依赖

```json
{
  "dependencies": {
    "@dcloudio/uni-app": "^3.0.0",
    "@dcloudio/uni-app-plus": "^3.0.0",
    "@dcloudio/uni-h5": "^3.0.0",
    "@dcloudio/uni-mp-weixin": "^3.0.0",
    "@dcloudio/uni-ui": "^1.4.0",
    "vue": "^3.4.0",
    "pinia": "^2.1.0"
  }
}
```

## 📁 目录说明

```
mobile/
├── src/           # 源码
├── index.html     # H5 入口
├── package.json   # 依赖与脚本
└── vite.config.js # Vite 配置
```

## 🤝 相关项目

| 项目 | 说明 |
|------|------|
| [AI起爆一体化工具集](https://gitee.com/makersoft/app-store) | Web 管理端（Vue3 + Element Plus） |
| Go 后端 | API 服务（登录/JWT/业务接口） |

## 📝 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0.0 | - | 首个版本：AI 工作台、硬件管理、诊断、知识库、工具集 |

---

*Copyright © 2026 · AI起爆运维执行平台*
