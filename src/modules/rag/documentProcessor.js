/**
 * 文档处理模块
 * 支持多种文档格式的解析和预处理
 * 支持格式: txt, md, json, pdf, docx, xlsx, xls, pptx, ppt
 */

import * as pdfjsLib from 'pdfjs-dist';
import PdfWorker from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
import mammoth from 'mammoth';
import * as XLSX from 'xlsx';
import JSZip from 'jszip';

// 配置 pdf.js worker
pdfjsLib.GlobalWorkerOptions.workerSrc = PdfWorker;

class DocumentProcessor {
  constructor() {
    this.supportedFormats = ['txt', 'md', 'json', 'pdf', 'docx', 'xlsx', 'xls', 'pptx', 'ppt'];
    this.maxChunkSize = 1000; // 每个分块的最大字符数
    this.chunkOverlap = 200;  // 分块重叠字符数
  }

  /**
   * 处理文档文件
   * @param {File} file - 文档文件
   * @returns {Promise<Array>} - 处理后的文档分块
   */
  async processFile(file) {
    const extension = this.getFileExtension(file.name);
    
    if (!this.supportedFormats.includes(extension)) {
      throw new Error(`不支持的文件格式: ${extension}`);
    }

    try {
      const content = await this.readFileContent(file, extension);
      const chunks = this.splitIntoChunks(content);
      
      return chunks.map((chunk, index) => ({
        id: `${file.name}_${Date.now()}_${index}`,
        text: chunk,
        metadata: {
          filename: file.name,
          extension,
          chunkIndex: index,
          totalChunks: chunks.length,
          processedAt: new Date().toISOString(),
          fileSize: file.size,
          fileType: file.type
        }
      }));
    } catch (error) {
      console.error('文档处理失败:', error);
      throw error;
    }
  }

  /**
   * 处理纯文本
   * @param {string} text - 原始文本
   * @param {Object} metadata - 元数据
   * @returns {Array} - 处理后的文档分块
   */
  processText(text, metadata = {}) {
    const chunks = this.splitIntoChunks(text);
    
    return chunks.map((chunk, index) => ({
      id: `text_${Date.now()}_${index}`,
      text: chunk,
      metadata: {
        ...metadata,
        chunkIndex: index,
        totalChunks: chunks.length,
        processedAt: new Date().toISOString(),
        source: 'text_input'
      }
    }));
  }

  /**
   * 读取文件内容并提取纯文本
   */
  async readFileContent(file, extension) {
    try {
      switch (extension) {
        case 'txt':
        case 'md':
          return await file.text();

        case 'json': {
          const raw = await file.text();
          return this.jsonToText(JSON.parse(raw));
        }

        case 'pdf': {
          const buf = await file.arrayBuffer();
          return await this.extractTextFromPDF(buf);
        }

        case 'docx': {
          const buf = await file.arrayBuffer();
          return await this.extractTextFromDOCX(buf);
        }

        case 'xlsx':
        case 'xls': {
          const buf = await file.arrayBuffer();
          return await this.extractTextFromXLSX(buf);
        }

        case 'pptx':
        case 'ppt': {
          const buf = await file.arrayBuffer();
          return await this.extractTextFromPPTX(buf);
        }

        default:
          throw new Error(`未处理的文件格式: ${extension}`);
      }
    } catch (error) {
      console.error('读取文件失败:', error);
      throw error;
    }
  }

  /**
   * 从 PDF 提取文本
   */
  async extractTextFromPDF(arrayBuffer) {
    const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
    let fullText = '';

    for (let i = 1; i <= pdf.numPages; i++) {
      const page = await pdf.getPage(i);
      const textContent = await page.getTextContent();
      const pageText = textContent.items.map((item) => item.str).join(' ');
      fullText += pageText + '\n';
    }

    return fullText;
  }

  /**
   * 从 DOCX 提取文本
   */
  async extractTextFromDOCX(arrayBuffer) {
    const result = await mammoth.extractRawText({ arrayBuffer });
    return result.value || '';
  }

  /**
   * 从 Excel 提取文本
   */
  async extractTextFromXLSX(arrayBuffer) {
    const wb = XLSX.read(arrayBuffer, { type: 'array' });
    let text = '';

    wb.SheetNames.forEach((name) => {
      const ws = wb.Sheets[name];
      text += `【工作表: ${name}】\n`;
      text += XLSX.utils.sheet_to_txt(ws) + '\n';
    });

    return text;
  }

  /**
   * 从 PPTX 提取文本（OXML 本质是 zip 包）
   */
  async extractTextFromPPTX(arrayBuffer) {
    const zip = await JSZip.loadAsync(arrayBuffer);
    const slideFiles = Object.keys(zip.files)
      .filter((f) => /^ppt\/slides\/slide\d+\.xml$/i.test(f))
      .sort((a, b) => {
        const na = parseInt(a.match(/\d+/)[0], 10);
        const nb = parseInt(b.match(/\d+/)[0], 10);
        return na - nb;
      });

    let text = '';
    for (const f of slideFiles) {
      const xml = await zip.files[f].async('string');
      const matches = xml.match(/<a:t>([\s\S]*?)<\/a:t>/g) || [];
      const slideText = matches
        .map((m) => m.replace(/<a:t>/, '').replace(/<\/a:t>/, ''))
        .join(' ');
      text += slideText + '\n';
    }

    return text.trim();
  }

  /**
   * JSON转文本
   */
  jsonToText(jsonData) {
    if (typeof jsonData === 'string') {
      return jsonData;
    }
    
    const processObject = (obj, depth = 0) => {
      let result = '';
      const indent = '  '.repeat(depth);
      
      if (Array.isArray(obj)) {
        obj.forEach((item, index) => {
          result += `${indent}${index + 1}. ${processObject(item, depth + 1)}\n`;
        });
      } else if (obj && typeof obj === 'object') {
        Object.entries(obj).forEach(([key, value]) => {
          if (typeof value === 'object' && value !== null) {
            result += `${indent}${key}:\n${processObject(value, depth + 1)}`;
          } else {
            result += `${indent}${key}: ${value}\n`;
          }
        });
      } else {
        result += `${indent}${obj}\n`;
      }
      
      return result;
    };
    
    return processObject(jsonData);
  }

  /**
   * 文本分块
   */
  splitIntoChunks(text) {
    if (!text || text.length === 0) {
      return [];
    }

    const chunks = [];
    let start = 0;
    
    while (start < text.length) {
      let end = start + this.maxChunkSize;
      
      // 如果不在段落边界，向前找段落结束
      if (end < text.length) {
        // 优先在段落边界分割
        const paragraphEnd = text.lastIndexOf('\n\n', end);
        if (paragraphEnd > start + this.maxChunkSize / 2) {
          end = paragraphEnd;
        } else {
          // 次优在句子边界分割
          const sentenceEnd = Math.max(
            text.lastIndexOf('. ', end),
            text.lastIndexOf('。', end),
            text.lastIndexOf('! ', end),
            text.lastIndexOf('? ', end)
          );
          if (sentenceEnd > start + this.maxChunkSize / 3) {
            end = sentenceEnd + 1;
          } else {
            // 最后在单词边界分割
            const wordEnd = text.lastIndexOf(' ', end);
            if (wordEnd > start + this.maxChunkSize / 4) {
              end = wordEnd;
            }
          }
        }
      } else {
        end = text.length;
      }
      
      const chunk = text.substring(start, end).trim();
      if (chunk.length > 0) {
        chunks.push(chunk);
      }
      
      // 移动起始位置，考虑重叠
      start = Math.max(start + 1, end - this.chunkOverlap);
    }
    
    return chunks;
  }

  /**
   * 文本预处理
   */
  preprocessText(text) {
    if (!text) return '';
    
    return text
      .toLowerCase()
      .replace(/\s+/g, ' ')          // 合并空白字符
      .replace(/[^\w\u4e00-\u9fff\s.,!?;:()\-]/g, ' ') // 保留常用标点
      .trim();
  }

  /**
   * 提取关键词
   */
  extractKeywords(text, maxKeywords = 10) {
    const words = text.toLowerCase()
      .split(/[^\w\u4e00-\u9fff]+/)
      .filter(word => word.length > 1);
    
    const freq = {};
    words.forEach(word => {
      freq[word] = (freq[word] || 0) + 1;
    });
    
    return Object.entries(freq)
      .sort(([, a], [, b]) => b - a)
      .slice(0, maxKeywords)
      .map(([word]) => word);
  }

  /**
   * 获取文件扩展名
   */
  getFileExtension(filename) {
    return filename.split('.').pop().toLowerCase();
  }

  /**
   * 验证文件
   */
  validateFile(file) {
    const extension = this.getFileExtension(file.name);
    
    if (!this.supportedFormats.includes(extension)) {
      return {
        valid: false,
        error: `不支持的文件格式: ${extension}。支持格式: ${this.supportedFormats.join(', ')}`
      };
    }
    
    if (file.size > 10 * 1024 * 1024) { // 10MB限制
      return {
        valid: false,
        error: '文件大小超过10MB限制'
      };
    }
    
    return { valid: true };
  }
}

// 单例模式导出
const documentProcessor = new DocumentProcessor();
export default documentProcessor;
