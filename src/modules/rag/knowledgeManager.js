/**
 * 知识库管理器
 * 整合向量搜索和文档处理功能
 */

import vectorStore from './vectorStore.js';
import documentProcessor from './documentProcessor.js';

class KnowledgeManager {
  constructor() {
    this.documents = new Map();
    this.categories = new Set();
    this.tags = new Map();
    this.initialized = false;
  }

  async initialize() {
    if (this.initialized) return;
    
    try {
      await vectorStore.initialize();
      await this.loadFromStorage();
      this.initialized = true;
      console.log('知识库管理器初始化完成');
    } catch (error) {
      console.error('知识库管理器初始化失败:', error);
    }
  }

  /**
   * 添加知识条目
   */
  async addKnowledgeItem(item) {
    await this.initialize();
    
    const id = item.id || `knowledge_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const knowledgeItem = {
      id,
      question: item.question || '',
      answer: item.answer || '',
      tags: Array.isArray(item.tags) ? item.tags : [item.tags || 'general'],
      category: item.category || 'general',
      source: item.source || 'manual',
      metadata: {
        ...item.metadata,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
    };

    // 更新分类和标签
    this.categories.add(knowledgeItem.category);
    knowledgeItem.tags.forEach(tag => {
      if (!this.tags.has(tag)) {
        this.tags.set(tag, new Set());
      }
      this.tags.get(tag).add(id);
    });

    // 存储到内存
    this.documents.set(id, knowledgeItem);

    // 添加到向量存储
    const text = `${knowledgeItem.question} ${knowledgeItem.answer}`;
    await vectorStore.addDocument(id, text, {
      type: 'knowledge',
      category: knowledgeItem.category,
      tags: knowledgeItem.tags,
      ...knowledgeItem.metadata
    });

    // 保存到本地存储
    await this.saveToStorage();

    return knowledgeItem;
  }

  /**
   * 批量添加知识条目
   */
  async batchAddKnowledgeItems(items) {
    const results = [];
    for (const item of items) {
      const result = await this.addKnowledgeItem(item);
      results.push(result);
    }
    return results;
  }

  /**
   * 搜索知识库
   */
  async searchKnowledge(query, options = {}) {
    await this.initialize();
    
    const {
      topK = 5,
      minScore = 0.3,
      category = null,
      tags = [],
      includeAnswer = true
    } = options;

    // 构建过滤器
    const filter = (metadata) => {
      if (category && metadata.category !== category) return false;
      if (tags.length > 0) {
        const hasMatchingTag = tags.some(tag => metadata.tags?.includes(tag));
        if (!hasMatchingTag) return false;
      }
      return true;
    };

    // 执行向量搜索
    const vectorResults = await vectorStore.search(query, {
      topK: topK * 2, // 获取更多结果用于筛选
      minScore,
      filter
    });

    // 获取完整知识条目
    const knowledgeResults = vectorResults
      .map(result => {
        const knowledgeItem = this.documents.get(result.id);
        if (!knowledgeItem) return null;
        
        return {
          ...knowledgeItem,
          similarity: result.score,
          score: result.score
        };
      })
      .filter(item => item !== null)
      .sort((a, b) => b.score - a.score)
      .slice(0, topK);

    // 如果向量搜索结果不足，尝试关键词搜索
    if (knowledgeResults.length < topK) {
      const keywordResults = await this.keywordSearch(query, options);
      const existingIds = new Set(knowledgeResults.map(item => item.id));
      
      for (const item of keywordResults) {
        if (!existingIds.has(item.id) && knowledgeResults.length < topK) {
          knowledgeResults.push({
            ...item,
            similarity: 0.5, // 默认相似度
            score: 0.5
          });
        }
      }
    }

    return knowledgeResults;
  }

  /**
   * 关键词搜索（回退方案）
   */
  async keywordSearch(query, options = {}) {
    const { topK = 5, category = null, tags = [] } = options;
    
    const queryWords = query.toLowerCase().split(/\s+/).filter(w => w.length > 1);
    const results = [];

    for (const [id, item] of this.documents.entries()) {
      // 应用过滤器
      if (category && item.category !== category) continue;
      if (tags.length > 0 && !tags.some(tag => item.tags.includes(tag))) continue;

      // 计算关键词匹配度
      const text = `${item.question} ${item.answer}`.toLowerCase();
      let score = 0;
      
      for (const word of queryWords) {
        if (text.includes(word)) {
          score += 1;
        }
      }

      if (score > 0) {
        results.push({
          ...item,
          score: score / queryWords.length
        });
      }
    }

    return results
      .sort((a, b) => b.score - a.score)
      .slice(0, topK);
  }

  /**
   * 获取知识条目
   */
  getKnowledgeItem(id) {
    return this.documents.get(id);
  }

  /**
   * 更新知识条目
   */
  async updateKnowledgeItem(id, updates) {
    const existing = this.documents.get(id);
    if (!existing) {
      throw new Error(`未找到知识条目: ${id}`);
    }

    const updatedItem = {
      ...existing,
      ...updates,
      metadata: {
        ...existing.metadata,
        ...updates.metadata,
        updatedAt: new Date().toISOString()
      }
    };

    // 更新向量存储
    const text = `${updatedItem.question} ${updatedItem.answer}`;
    await vectorStore.addDocument(id, text, {
      type: 'knowledge',
      category: updatedItem.category,
      tags: updatedItem.tags,
      ...updatedItem.metadata
    });

    // 更新内存存储
    this.documents.set(id, updatedItem);

    // 保存到本地存储
    await this.saveToStorage();

    return updatedItem;
  }

  /**
   * 删除知识条目
   */
  async deleteKnowledgeItem(id) {
    const existed = this.documents.delete(id);
    if (existed) {
      await vectorStore.deleteDocument(id);
      await this.saveToStorage();
    }
    return existed;
  }

  /**
   * 导入文档
   */
  async importDocument(file, options = {}) {
    const validation = documentProcessor.validateFile(file);
    if (!validation.valid) {
      throw new Error(validation.error);
    }

    const chunks = await documentProcessor.processFile(file);
    const results = [];

    for (const chunk of chunks) {
      const knowledgeItem = await this.addKnowledgeItem({
        question: `文档片段: ${chunk.metadata.filename} - 第${chunk.metadata.chunkIndex + 1}部分`,
        answer: chunk.text,
        tags: ['document', 'imported', ...(options.tags || [])],
        category: options.category || 'document',
        source: 'file_import',
        metadata: {
          ...chunk.metadata,
          importOptions: options
        }
      });
      results.push(knowledgeItem);
    }

    return results;
  }

  /**
   * 获取统计信息
   */
  getStats() {
    return {
      totalItems: this.documents.size,
      categories: Array.from(this.categories),
      tags: Array.from(this.tags.keys()),
      vectorStats: vectorStore.getStats(),
      lastUpdated: new Date().toISOString()
    };
  }

  /**
   * 导出知识库
   */
  exportKnowledge(format = 'json') {
    const data = {
      version: '1.0',
      exportedAt: new Date().toISOString(),
      items: Array.from(this.documents.values()),
      stats: this.getStats()
    };

    switch (format.toLowerCase()) {
      case 'json':
        return JSON.stringify(data, null, 2);
      case 'csv':
        return this.convertToCSV(data.items);
      default:
        throw new Error(`不支持的导出格式: ${format}`);
    }
  }

  /**
   * 转换为CSV格式
   */
  convertToCSV(items) {
    if (items.length === 0) return '';
    
    const headers = ['id', 'question', 'answer', 'category', 'tags', 'createdAt'];
    const rows = items.map(item => [
      item.id,
      `"${item.question.replace(/"/g, '""')}"`,
      `"${item.answer.replace(/"/g, '""')}"`,
      item.category,
      `"${item.tags.join(',')}"`,
      item.metadata.createdAt
    ]);

    return [headers.join(','), ...rows.map(row => row.join(','))].join('\n');
  }

  /**
   * 保存到本地存储
   */
  async saveToStorage() {
    try {
      const data = {
        version: '1.0',
        savedAt: new Date().toISOString(),
        documents: Array.from(this.documents.entries()),
        categories: Array.from(this.categories),
        tags: Array.from(this.tags.entries()).map(([tag, ids]) => ({
          tag,
          ids: Array.from(ids)
        }))
      };

      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('knowledge_base', JSON.stringify(data));
      }
      
      console.log('知识库已保存到本地存储');
    } catch (error) {
      console.error('保存知识库失败:', error);
    }
  }

  /**
   * 从本地存储加载
   */
  async loadFromStorage() {
    try {
      if (typeof localStorage !== 'undefined') {
        const saved = localStorage.getItem('knowledge_base');
        if (saved) {
          const data = JSON.parse(saved);
          
          // 恢复文档
          for (const [id, item] of data.documents) {
            this.documents.set(id, item);
          }
          
          // 恢复分类
          this.categories = new Set(data.categories);
          
          // 恢复标签
          this.tags = new Map();
          for (const tagData of data.tags || []) {
            this.tags.set(tagData.tag, new Set(tagData.ids));
          }
          
          console.log('知识库已从本地存储加载');
        }
      }
    } catch (error) {
      console.warn('加载知识库失败:', error);
    }
  }

  /**
   * 清空知识库
   */
  async clear() {
    this.documents.clear();
    this.categories.clear();
    this.tags.clear();
    await vectorStore.clear();
    
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem('knowledge_base');
    }
    
    console.log('知识库已清空');
  }
}

// 单例模式导出
const knowledgeManager = new KnowledgeManager();
export default knowledgeManager;