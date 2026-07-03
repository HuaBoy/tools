/**
 * 设备类型映射工具
 *
 * 设备类型由软件版本（APP 版本）首字母决定：
 *   L → 小勇士设备
 *   I → DT40设备
 *   Q → 全面屏设备
 *   K → 煤许设备
 *   T → DT40 + 小勇士设备（同时适用）
 *
 * 用于：
 *   - VersionHistory.vue 版本履历中显示适用设备类型
 *   - FactoryDataQuery.vue 工厂数据中显示设备类型
 *   - 任何需要根据 APP 版本识别设备类型的场景
 */

/**
 * 软件版本首字母 → 设备类型映射表
 */
export const DEVICE_TYPE_BY_SOFTWARE_VERSION = {
  L: { name: '小勇士设备', code: 'L', color: '#165DFF', desc: '小勇士系列产品' },
  I: { name: 'DT40设备', code: 'I', color: '#722ED1', desc: 'DT40 系列产品' },
  Q: { name: '全面屏设备', code: 'Q', color: '#13C2C2', desc: '全面屏系列产品' },
  K: { name: '煤许设备', code: 'K', color: '#FA8C16', desc: '煤安/煤许系列产品' },
  T: { name: 'DT40和小勇士设备', code: 'T', color: '#52C41A', desc: 'DT40 + 小勇士通用版本' }
};

/**
 * 默认设备类型（未匹配时）
 */
export const DEFAULT_DEVICE_TYPE = { name: '通用设备', code: '', color: '#86909C', desc: '通用设备' };

/**
 * 从软件版本号中提取首字母
 * 支持格式：
 *   v3.2.0 → v
 *   V3.2.0 → V
 *   L1.0.0 → L
 *   I2.5 → I
 *   Q3.0.1 → Q
 *   3.2.0 → '' (无字母)
 */
export const extractSoftwareVersionLetter = (appVersion) => {
  if (!appVersion || typeof appVersion !== 'string') return '';
  const trimmed = appVersion.trim();
  if (!trimmed) return '';

  // 提取第一个英文字母
  const match = trimmed.match(/[A-Za-z]/);
  if (match) {
    return match[0].toUpperCase();
  }
  return '';
};

/**
 * 根据软件版本（APP 版本）获取设备类型
 *
 * @param {string} appVersion - 软件版本号，如 v3.2.0、I2.5、L1.0
 * @returns {{ name: string, code: string, color: string, desc: string }}
 */
export const getDeviceTypeByAppVersion = (appVersion) => {
  const letter = extractSoftwareVersionLetter(appVersion);
  if (!letter) return { ...DEFAULT_DEVICE_TYPE };

  const type = DEVICE_TYPE_BY_SOFTWARE_VERSION[letter];
  if (type) {
    return { ...type };
  }
  return { ...DEFAULT_DEVICE_TYPE };
};

/**
 * 获取设备类型名称（字符串）
 */
export const getDeviceTypeName = (appVersion) => {
  return getDeviceTypeByAppVersion(appVersion).name;
};

/**
 * 获取设备类型显示样式
 * 返回 { name, code, color, desc }
 */
export const getDeviceTypeDisplay = (appVersion) => {
  return getDeviceTypeByAppVersion(appVersion);
};

/**
 * 批量获取设备类型（用于列表展示）
 */
export const getDeviceTypesByAppVersions = (appVersions) => {
  if (!Array.isArray(appVersions)) return [];
  return appVersions.map(v => getDeviceTypeByAppVersion(v));
};

/**
 * 判断软件版本是否属于指定设备类型
 *
 * 例如：isAppVersionMatchDeviceType('T2.0.0', 'DT40') → true
 *      isAppVersionMatchDeviceType('T2.0.0', '小勇士') → true
 */
export const isAppVersionMatchDeviceType = (appVersion, deviceName) => {
  const type = getDeviceTypeByAppVersion(appVersion);
  if (!type) return false;

  // T 类型（DT40 + 小勇士）需要特殊处理
  if (type.code === 'T') {
    return /DT40|小勇士/i.test(deviceName);
  }
  return type.name.includes(deviceName) || deviceName.includes(type.name.replace('设备', ''));
};

/**
 * 获取所有支持的设备类型列表
 */
export const getAllDeviceTypes = () => {
  return Object.values(DEVICE_TYPE_BY_SOFTWARE_VERSION);
};
