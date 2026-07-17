/**
 * 本地文本嵌入模块
 * 使用词频向量和文本相似度算法
 */

class LocalEmbedding {
  constructor() {
    this.initialized = true;
    this.vectorSize = 128; // 简化向量维度
  }

  async initialize() {
    // 无需初始化
    return Promise.resolve();
  }

  async embedText(text) {
    if (!text || text.trim().length === 0) {
      return new Array(this.vectorSize).fill(0);
    }

    // 使用词频向量化
    return this.createWordVector(text);
  }

  createWordVector(text) {
    // 文本预处理
    const words = text.toLowerCase()
      .replace(/[^\w\u4e00-\u9fff\s]/g, ' ') // 保留中文字符
      .split(/\s+/)
      .filter(w => w.length > 1);
    
    // 创建词频向量
    const vector = new Array(this.vectorSize).fill(0);
    const wordFreq = {};
    
    // 计算词频
    words.forEach(word => {
      wordFreq[word] = (wordFreq[word] || 0) + 1;
    });
    
    // 将词频映射到向量空间
    Object.entries(wordFreq).forEach(([word, freq]) => {
      const hash = this.stringHash(word) % this.vectorSize;
      vector[hash] += freq;
    });
    
    // 归一化
    return this.normalizeVector(vector);
  }

  stringHash(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // 转换为32位整数
    }
    return Math.abs(hash);
  }

  normalizeVector(vector) {
    const norm = Math.sqrt(vector.reduce((sum, val) => sum + val * val, 0));
    if (norm > 0) {
      return vector.map(val => val / norm);
    }
    return vector;
  }

  async embedBatch(texts) {
    const embeddings = [];
    for (const text of texts) {
      const embedding = await this.embedText(text);
      embeddings.push(embedding);
    }
    return embeddings;
  }

  cosineSimilarity(vec1, vec2) {
    if (vec1.length !== vec2.length) return 0;
    
    let dotProduct = 0;
    let norm1 = 0;
    let norm2 = 0;
    
    for (let i = 0; i < vec1.length; i++) {
      dotProduct += vec1[i] * vec2[i];
      norm1 += vec1[i] * vec1[i];
      norm2 += vec2[i] * vec2[i];
    }
    
    if (norm1 === 0 || norm2 === 0) return 0;
    return dotProduct / (Math.sqrt(norm1) * Math.sqrt(norm2));
  }

  // 添加文本相似度计算（基于字符串相似度）
  textSimilarity(text1, text2) {
    if (!text1 || !text2) return 0;
    
    // 转换为小写并分词
    const words1 = text1.toLowerCase().split(/\s+/).filter(w => w.length > 1);
    const words2 = text2.toLowerCase().split(/\s+/).filter(w => w.length > 1);
    
    if (words1.length === 0 || words2.length === 0) return 0;
    
    // 计算Jaccard相似度
    const set1 = new Set(words1);
    const set2 = new Set(words2);
    const intersection = new Set([...set1].filter(x => set2.has(x)));
    const union = new Set([...set1, ...set2]);
    
    return intersection.size / union.size;
  }
}

// 单例模式导出
const embeddingInstance = new LocalEmbedding();
export default embeddingInstance;