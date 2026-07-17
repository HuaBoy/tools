/**
 * 本地向量存储模块
 * 使用内存和文件系统存储向量数据
 */

import embeddingInstance from './embedding.js';

class VectorStore {
  constructor() {
    this.vectors = new Map(); // id -> {vector, metadata}
    this.index = new Map();   // 用于快速搜索的索引
    this.vectorSize = 384;    // 向量维度
    this.initialized = false;
    this.storagePath = './knowledge/vector_db';
  }

  async initialize() {
    if (this.initialized) return;
    
    try {
      // 从本地文件加载已有向量
      await this.loadFromFile();
      this.initialized = true;
      console.log('向量存储初始化完成，当前向量数:', this.vectors.size);
    } catch (error) {
      console.warn('向量存储加载失败，使用空存储:', error);
      this.initialized = true;
    }
  }

  async addDocument(id, text, metadata = {}) {
    await this.initialize();
    
    // 生成向量
    const vector = await embeddingInstance.embedText(text);
    
    // 存储向量和元数据
    this.vectors.set(id, {
      vector,
      metadata: {
        ...metadata,
        text,
        timestamp: new Date().toISOString(),
        id
      }
    });

    // 更新索引
    this.updateIndex(id, vector);
    
    // 异步保存到文件
    this.saveToFile().catch(console.error);
    
    return id;
  }

  async addDocuments(documents) {
    const results = [];
    for (const doc of documents) {
      const result = await this.addDocument(doc.id, doc.text, doc.metadata);
      results.push(result);
    }
    return results;
  }

  async search(query, options = {}) {
    await this.initialize();
    const {
      topK = 5,
      minScore = 0.3,
      filter = null
    } = options;

    // 生成查询向量
    const queryVector = await embeddingInstance.embedText(query);
    
    // 计算相似度
    const results = [];
    
    for (const [id, data] of this.vectors.entries()) {
      // 应用过滤器
      if (filter && !filter(data.metadata)) {
        continue;
      }
      
      const similarity = embeddingInstance.cosineSimilarity(queryVector, data.vector);
      
      if (similarity >= minScore) {
        results.push({
          id,
          score: similarity,
          text: data.metadata.text,
          metadata: data.metadata
        });
      }
    }

    // 按相似度排序
    results.sort((a, b) => b.score - a.score);
    
    // 返回topK结果
    return results.slice(0, topK);
  }

  async semanticSearch(query, context = '', options = {}) {
    // 结合上下文进行语义搜索
    const enhancedQuery = context ? `${context} ${query}` : query;
    return this.search(enhancedQuery, options);
  }

  getDocument(id) {
    return this.vectors.get(id);
  }

  deleteDocument(id) {
    const existed = this.vectors.delete(id);
    if (existed) {
      this.saveToFile().catch(console.error);
    }
    return existed;
  }

  clear() {
    this.vectors.clear();
    this.index.clear();
    this.saveToFile().catch(console.error);
  }

  getStats() {
    return {
      totalVectors: this.vectors.size,
      vectorSize: this.vectorSize,
      lastUpdated: new Date().toISOString()
    };
  }

  updateIndex(id, vector) {
    // 简单的索引策略：存储前10个最大值的索引
    const maxIndices = [];
    for (let i = 0; i < vector.length; i++) {
      maxIndices.push({ index: i, value: Math.abs(vector[i]) });
    }
    
    maxIndices.sort((a, b) => b.value - a.value);
    const topIndices = maxIndices.slice(0, 10).map(item => item.index);
    
    for (const idx of topIndices) {
      if (!this.index.has(idx)) {
        this.index.set(idx, new Set());
      }
      this.index.get(idx).add(id);
    }
  }

  async saveToFile() {
    try {
      const data = {
        vectors: Array.from(this.vectors.entries()),
        metadata: {
          version: '1.0',
          savedAt: new Date().toISOString(),
          vectorSize: this.vectorSize
        }
      };

      // 在实际应用中，这里应该写入文件
      // 由于浏览器限制，这里使用localStorage作为示例
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('vector_store', JSON.stringify(data));
      }
      
      console.log('向量存储已保存');
    } catch (error) {
      console.error('保存向量存储失败:', error);
    }
  }

  async loadFromFile() {
    try {
      // 从localStorage加载（示例）
      if (typeof localStorage !== 'undefined') {
        const saved = localStorage.getItem('vector_store');
        if (saved) {
          const data = JSON.parse(saved);
          
          for (const [id, vectorData] of data.vectors) {
            this.vectors.set(id, vectorData);
            this.updateIndex(id, vectorData.vector);
          }
          
          console.log('向量存储已从本地加载');
        }
      }
    } catch (error) {
      console.warn('加载向量存储失败:', error);
    }
  }

  // 批量操作
  async batchAdd(documents) {
    return this.addDocuments(documents);
  }

  async batchSearch(queries, options = {}) {
    const results = [];
    for (const query of queries) {
      const result = await this.search(query, options);
      results.push(result);
    }
    return results;
  }
}

// 单例模式导出
const vectorStore = new VectorStore();
export default vectorStore;