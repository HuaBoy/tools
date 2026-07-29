// 销售岗数据访问层
// 当前为前端单一数据源（src/data/salesData.js），仅做展示、不落库。
// 后续若对接外部数据源自动采集，只需将 getSalesData 改为 fetch 外部接口即可，
// 页面无需改动。
import {
  salesMeta,
  nationalData,
  regionalData,
  cr20Data,
  newsData,
  sourceList
} from '@/data/salesData'

export const salesApi = {
  /**
   * 获取销售岗行业洞察数据
   * @returns {Promise<{meta:object,national:Array,regional:object,cr20:object,news:Array,sources:Array}>}
   */
  async getSalesData() {
    return {
      meta: salesMeta,
      national: nationalData,
      regional: regionalData,
      cr20: cr20Data,
      news: newsData,
      sources: sourceList
    }
  }
}
