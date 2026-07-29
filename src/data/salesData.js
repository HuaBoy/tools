// 销售岗 · 民爆行业洞察 —— 前端单一数据源（仅展示，不落库）
//
// 维护说明：
//   运营 / 行业研究员按工信部安全生产司、中国爆破器材行业协会、各集团年报、
//   券商研究所（国金/国信/中休信/开源/头豹等）、第三方机构（智研咨询/观研天下/
//   QYResearch/普华有策）、地方省民爆协会的最新权威发布更新本文件。
//
// 可见性（visibility）：
//   'public'   —— 公开版可见（行业聚合数据）
//   'internal' —— 仅内部版可见（企业级/省级明细等对外不披露数据）
//
// 注：当前数值为示例占位，请替换为最新权威数据。

export const salesMeta = {
  // 数据版本时间：每次更新数据时修改此处
  updatedAt: '2026-07-22',
  // 维护提示
  note: '当前为示例占位数据，请运营按工信部安全生产司、中国爆破器材行业协会及各家最新权威发布及时更新。'
}

// 全国总量
export const nationalData = [
  {
    label: '工业炸药产销总量',
    value: '452.3',
    unit: '万吨',
    period: '2025年',
    trend: '同比 +4.8%',
    trendUp: true,
    source: '中国爆破器材行业协会',
    visibility: 'public'
  },
  {
    label: '工业雷管产销总量',
    value: '19.6',
    unit: '亿发',
    period: '2025年',
    trend: '电子雷管占比 96%',
    trendUp: true,
    source: '中国爆破器材行业协会',
    visibility: 'public'
  },
  {
    label: '民爆行业总产值',
    value: '432.5',
    unit: '亿元',
    period: '2025年',
    trend: '同比 +5.2%',
    trendUp: true,
    source: '工信部安全生产司',
    visibility: 'public'
  },
  {
    label: '民爆产品销售总值',
    value: '418.0',
    unit: '亿元',
    period: '2025年',
    trend: '同比 +4.6%',
    trendUp: true,
    source: '工信部安全生产司',
    visibility: 'public'
  },
  {
    label: '爆破服务总收入',
    value: '386.2',
    unit: '亿元',
    period: '2025年',
    trend: '同比 +7.1%',
    trendUp: true,
    source: '中国爆破器材行业协会',
    visibility: 'public'
  },
  {
    label: '硝酸铵均价（原料）',
    value: '2,180',
    unit: '元/吨',
    period: '2025年',
    trend: '同比 -3.1%',
    trendUp: false,
    source: '第三方机构（智研咨询）',
    visibility: 'public'
  }
]

// 区域产销分布
export const regionalData = {
  // 各大区（公开版）
  macro: [
    { region: '华东', output: '118.5', share: '26.2%' },
    { region: '华中', output: '76.3', share: '16.9%' },
    { region: '西南', output: '68.9', share: '15.2%' },
    { region: '西北', output: '64.1', share: '14.2%' },
    { region: '华北', output: '52.7', share: '11.7%' },
    { region: '华南', output: '41.2', share: '9.1%' },
    { region: '东北', output: '30.6', share: '6.7%' }
  ],
  // 重点省份明细（内部版，对外不披露）
  province: [
    { region: '新疆', output: '38.4' },
    { region: '内蒙古', output: '31.2' },
    { region: '贵州', output: '27.6' },
    { region: '四川', output: '24.3' },
    { region: '山西', output: '19.8' },
    { region: '湖南', output: '17.5' },
    { region: '广东', output: '15.1' }
  ]
}

// CR20 集中度
export const cr20Data = {
  period: '2025年',
  source: '中国爆破器材行业协会 / 工信部安全生产司',
  metrics: [
    {
      label: '前20大企业合计产量',
      value: '286.0',
      unit: '万吨',
      sub: '约占行业 63.3%',
      visibility: 'public'
    },
    {
      label: '前20大企业合计产值',
      value: '312.0',
      unit: '亿元',
      sub: '约占行业 72.1%',
      visibility: 'public'
    },
    {
      label: '行业 CR20 集中度',
      value: '63.3',
      unit: '%',
      sub: '按产量计',
      visibility: 'public'
    }
  ],
  publicNote:
    '公开版仅披露前20家企业的合计产量、合计产值与行业CR20集中度，不披露单家企业销量与营收（企业明细属行业内部统计资料，不对外放开）。',
  // 内部版专属对标信息（对外不披露）
  internal: {
    label: '本集团对标（内部）',
    text:
      '本集团工业炸药产量位列行业第一梯队，具体排名、产量与营收对标数据见内部经营分析，对外不披露。'
  }
}

// 行业动态资讯
export const newsData = [
  {
    date: '2026-07-15',
    tag: '政策',
    title: '工信部安全生产司部署民爆行业安全生产治本攻坚',
    desc: '持续推进电子雷管全面替代，强化生产、储存、运输全链条安全监管。',
    source: '工信部安全生产司',
    visibility: 'public'
  },
  {
    date: '2026-07-08',
    tag: '协会',
    title: '中国爆破器材行业协会发布上半年行业运行简报',
    desc: '上半年工业炸药产销平稳，爆破服务收入保持增长，区域集中度稳中有升。',
    source: '中国爆破器材行业协会',
    visibility: 'public'
  },
  {
    date: '2026-06-26',
    tag: '研报',
    title: '国金证券：民爆行业景气延续，关注一体化与电子雷管',
    desc: '看好民爆一体化龙头企业，原料价格下行利好盈利，电子雷管渗透持续提升。',
    source: '券商研究所（国金）',
    visibility: 'public'
  },
  {
    date: '2026-06-18',
    tag: '市场',
    title: '智研咨询：2025中国民爆行业市场规模与竞争格局分析',
    desc: '行业CR20持续提升，头部企业通过并购整合扩大份额，区域分布向中西部倾斜。',
    source: '第三方机构（智研咨询）',
    visibility: 'public'
  },
  {
    date: '2026-07-10',
    tag: '内部研判',
    title: '本集团二季度区域订单结构与竞品动态',
    desc: '内部经营分析：华东、西南重点项目需求旺盛，需关注同区域竞品报价策略。',
    source: '内部经营分析',
    visibility: 'internal'
  },
  {
    date: '2026-06-30',
    tag: '内部对标',
    title: '本集团产能利用率与许可产能对照',
    desc: '内部：当前许可产能与产能利用率对标情况，详见经营月报，对外不披露。',
    source: '内部经营分析',
    visibility: 'internal'
  }
]

// 数据来源清单（页脚展示）
export const sourceList = [
  '工信部安全生产司',
  '中国爆破器材行业协会',
  '特能集团',
  '葛洲坝',
  '宏大',
  '南岭民爆易普利',
  '江南化工',
  '保利联合',
  '凯龙股份',
  '国泰集团',
  '雪峰科技',
  '高争民爆',
  '同德化工',
  '券商研究所（国金/国信/中休信/开源/头豹等）',
  '第三方机构（智研咨询/观研天下/QYResearch/普华有策）',
  '地方省民爆协会'
]
