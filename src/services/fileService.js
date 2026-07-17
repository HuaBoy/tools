// 文件服务 - 上传/下载/格式转换
import { api } from './http'

export const fileService = {
  upload(file, onProgress) {
    const fd = new FormData()
    fd.append('file', file)
    return api.upload('/api/v1/files/upload', fd, onProgress)
  },
  download(fileId) {
    return api.get('/api/v1/files/download/' + fileId, {}, { responseType: 'blob' })
  },
  convert(file, targetFormat) {
    const fd = new FormData()
    fd.append('file', file)
    fd.append('target', targetFormat)
    return api.upload('/api/v1/files/convert', fd)
  },
  deleteFile(fileId) {
    return api.del('/api/v1/files/' + fileId)
  }
}