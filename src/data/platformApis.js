// 三方平台接口清单及说明 —— 前端单一数据源
//
// 维护说明：
//   各平台接口由对应开发/运维同学按实际可用接口维护。
//   本文件仅用于"三方授权"页面的接口说明展示，不负责真实请求。

// 通用返回结构（盛景系平台）
const BLADE_RESP = {
  code: 200,
  msg: '操作成功',
  data: { records: [], total: 0, size: 10, current: 1 }
}

// 通用请求头（需登录后的业务接口）
const AUTH_HEADER = [
  { key: 'Blade-Auth', value: 'bearer {access_token}', desc: '登录接口返回的访问令牌' },
  { key: 'Tenant-Id', value: '000000', desc: '租户ID，与登录时保持一致' }
]

export const platformApis = [
  // ==================== 盛景平台（云系统） ====================
  {
    id: 'mp',
    name: '盛景平台',
    shortName: '云系统',
    domain: 'https://mp.holyview.cn:9443',
    loginTokenKey: 'mp_token',
    description: '提供起爆设备数据、爆破记录、雷管追溯数据等查询能力，供云平台侧数据展示与追溯使用。',
    baseAuth: [
      { key: 'Authorization', value: 'Basic <base64(saber:saber_secret)>', desc: 'Basic 认证，用户名密码均为 saber' },
      { key: 'Tenant-Id', value: '000000', desc: '租户ID，固定为 000000' }
    ],
    apis: [
      {
        method: 'POST',
        path: '/api/blade-auth/oauth/token',
        name: '登录获取令牌',
        desc: '账号密码登录，获取访问令牌（access_token），作为后续业务接口的认证凭证。',
        headers: [
          { key: 'Authorization', value: 'Basic <base64(saber:saber_secret)>', desc: 'Basic 认证，用户名密码均为 saber' },
          { key: 'Tenant-Id', value: '000000', desc: '租户ID，固定为 000000' }
        ],
        params: [
          { key: 'tenantId', required: true, desc: '租户ID，固定为 000000' },
          { key: 'username', required: true, desc: '登录用户名（如 admin）' },
          { key: 'password', required: true, desc: '密码的 MD5 值' },
          { key: 'grant_type', required: true, desc: '授权类型，固定为 password' },
          { key: 'scope', required: false, desc: '授权范围，固定为 all' },
          { key: 'type', required: false, desc: '登录类型，固定为 account' }
        ],
        response: '返回 { code: 200, data: { access_token, token_type, expires_in, ... } }，登录成功后将 access_token 用于 Blade-Auth 请求头。',
        successCheck: 'code === 200 且 data.access_token 非空'
      },
      {
        method: 'GET',
        path: '/api/blade-detonate/blastDeviceFactory/page',
        name: '设备数据分页查询',
        desc: '查询起爆设备 / 雷管厂家数据，支持按手持机编号、控制器编号、版本、批次号、区域等多条件筛选。',
        headers: AUTH_HEADER,
        params: [
          { key: 'tenantId', required: false, desc: '租户ID（默认 000000）' },
          { key: 'companyName', required: false, desc: '雷管企业 / 使用单位名称' },
          { key: 'controllerCode', required: false, desc: '控制器编号' },
          { key: 'controllerVersion', required: false, desc: '控制器版本' },
          { key: 'deviceCode', required: false, desc: '手持机编号（SN，如 DZ600000016）' },
          { key: 'softwareVersion', required: false, desc: '软件版本' },
          { key: 'deviceVersion', required: false, desc: '手持机版本' },
          { key: 'deviceHardware', required: false, desc: '手持机硬件型号' },
          { key: 'deviceScene', required: false, desc: '版本场景' },
          { key: 'batchNo', required: false, desc: '批次号' },
          { key: 'provinceCode', required: false, desc: '省份编码' },
          { key: 'cityCode', required: false, desc: '城市编码' },
          { key: 'current', required: false, desc: '当前页码，默认 1' },
          { key: 'size', required: false, desc: '每页条数，默认 10' }
        ],
        response: `返回 { code: 200, data: { records, total } }，records 为设备列表。示例：${JSON.stringify(BLADE_RESP.data)}`
      },
      {
        method: 'GET',
        path: '/api/blade-detonate/blastTask/page',
        name: '爆破记录分页查询',
        desc: '查询爆破作业任务记录，支持按时间范围、设备编号、控制器、作业人员等条件筛选。',
        headers: AUTH_HEADER,
        params: [
          { key: 'startDate', required: false, desc: '开始日期（yyyy-MM-dd，如 2026-08-13）' },
          { key: 'endDate', required: false, desc: '结束日期（yyyy-MM-dd）' },
          { key: 'deviceCode', required: false, desc: '手持机编号（SN，如 DZ600000016）' },
          { key: 'controllerCode', required: false, desc: '控制器编号' },
          { key: 'controllerVersion', required: false, desc: '控制器版本' },
          { key: 'deviceVersion', required: false, desc: '手持机版本' },
          { key: 'explosionDate', required: false, desc: '爆破时间' },
          { key: 'uploadDlTime', required: false, desc: '上传时间' },
          { key: 'tenantName', required: false, desc: '雷管企业名称' },
          { key: 'deptName', required: false, desc: '使用单位名称' },
          { key: 'detonatorCount', required: false, desc: '雷管数量' },
          { key: 'taskName', required: false, desc: '工程名称' },
          { key: 'blasterUserName', required: false, desc: '作业人员姓名' },
          { key: 'blasterUserPhone', required: false, desc: '作业人员手机号' },
          { key: 'current', required: false, desc: '当前页码，默认 1' },
          { key: 'size', required: false, desc: '每页条数，默认 10' }
        ],
        response: `返回 { code: 200, data: { records, total } }，records 为爆破记录列表，含爆破时间、雷管数量、作业人员等字段。`
      },
      {
        method: 'GET',
        path: '/api/blade-detonator-factory/detonatorProducttestData/page',
        name: '追溯数据分页查询',
        desc: '查询雷管产品追溯数据，支持按批次号、上传时间、追溯码、起爆标识等条件筛选。',
        headers: AUTH_HEADER,
        params: [
          { key: 'tenantId', required: false, desc: '租户ID' },
          { key: 'companyId', required: false, desc: '企业ID' },
          { key: 'uploadStartDate', required: false, desc: '上传开始日期' },
          { key: 'uploadEndDate', required: false, desc: '上传结束日期' },
          { key: 'batchNo', required: false, desc: '批次号' },
          { key: 'shellCodeStart', required: false, desc: '壳码起' },
          { key: 'uid', required: false, desc: 'UID 编码' },
          { key: 'idlStart', required: false, desc: 'IDL 起' },
          { key: 'explosionMark', required: false, desc: '起爆标识' },
          { key: 'type', required: false, desc: '类型' },
          { key: 'moduleType', required: false, desc: '模块类型' },
          { key: 'traceableCode', required: false, desc: '追溯码' },
          { key: 'factoryId', required: false, desc: '工厂ID' },
          { key: 'startDate', required: false, desc: '开始日期' },
          { key: 'endDate', required: false, desc: '结束日期' },
          { key: 'explosionStartDate', required: false, desc: '起爆开始时间' },
          { key: 'explosionEndDate', required: false, desc: '起爆结束时间' },
          { key: 'current', required: false, desc: '当前页码，默认 1' },
          { key: 'size', required: false, desc: '每页条数，默认 10' }
        ],
        response: '返回 { code: 200, data: { records, total } }，records 为追溯数据列表。'
      },
      {
        method: 'GET',
        path: '/api/blade-detonator-factory/detonatorProducttestData/export',
        name: '追溯数据导出',
        desc: '按查询条件导出追溯数据文件（Excel）。',
        headers: AUTH_HEADER,
        params: [
          { key: 'uploadStartDate', required: false, desc: '上传开始日期' },
          { key: 'uploadEndDate', required: false, desc: '上传结束日期' },
          { key: 'batchNo', required: false, desc: '批次号' },
          { key: 'date', required: false, desc: '日期' },
          { key: 'date1', required: false, desc: '结束日期' }
        ],
        response: '成功时返回文件流，浏览器下载导出文件。'
      }
    ]
  },

  // ==================== 智能制造系统（MES） ====================
  {
    id: 'smart',
    name: '智能制造系统',
    shortName: 'MES',
    domain: 'http://218.90.146.230:20001',
    loginTokenKey: 'smart_factory_token / iot_token',
    description: '智能制造（MES）平台，提供工厂生产数据查询能力，用于工厂数据追溯与分析。',
    baseAuth: [
      { key: 'Authorization', value: 'Basic <base64(saber_identity_client:saber_identity_secret)>', desc: 'Basic 认证（identity 客户端）' },
      { key: 'Tenant-Id', value: '000000', desc: '租户ID，固定为 000000' }
    ],
    apis: [
      {
        method: 'POST',
        path: '/iot-api/api/blade-auth/oauth/token',
        name: '登录获取令牌',
        desc: '智能制造系统账号密码登录，获取访问令牌（access_token）。',
        headers: [
          { key: 'Authorization', value: 'Basic <base64(saber_identity_client:saber_identity_secret)>', desc: 'Basic 认证（identity 客户端）' },
          { key: 'Tenant-Id', value: '000000', desc: '租户ID，固定为 000000' }
        ],
        params: [
          { key: 'tenantId', required: true, desc: '租户ID，固定为 000000' },
          { key: 'username', required: true, desc: '登录用户名' },
          { key: 'password', required: true, desc: '密码的 MD5 值' },
          { key: 'grant_type', required: true, desc: '授权类型，固定为 password' },
          { key: 'scope', required: false, desc: '授权范围，固定为 all' },
          { key: 'type', required: false, desc: '登录类型，固定为 account' }
        ],
        response: '返回 { code: 200, data: { access_token, ... } }。',
        successCheck: 'code === 200 且 data.access_token 非空'
      },
      {
        method: 'GET',
        path: '/api/blade-iot/factoryDataQuery/page',
        name: '工厂数据分页查询',
        desc: '查询工厂生产数据，支持按设备类型、批次号、ID 编码、起爆租户等多条件筛选。',
        headers: AUTH_HEADER,
        params: [
          { key: 'deviceType', required: false, desc: '设备类型编码' },
          { key: '$deviceType', required: false, desc: '设备类型名称（模糊）' },
          { key: 'modeTypeDiy', required: false, desc: '自定义模式类型编码' },
          { key: '$modeTypeDiy', required: false, desc: '自定义模式类型名称（模糊）' },
          { key: '$mode', required: false, desc: '模式（模糊）' },
          { key: 'batchNo', required: false, desc: '批次号' },
          { key: 'idHex', required: false, desc: 'ID 十六进制编码' },
          { key: 'blastTenantId', required: false, desc: '起爆租户ID' },
          { key: 'current', required: false, desc: '当前页码，默认 1' },
          { key: 'size', required: false, desc: '每页条数，默认 10' }
        ],
        response: '返回 { code: 200, data: { records, total } }，records 为工厂数据列表。'
      }
    ]
  }
]
