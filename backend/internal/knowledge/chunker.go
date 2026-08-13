package knowledge

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ChunkOptions 分块配置
type ChunkOptions struct {
	ChunkSize   int // 每块最大字符数（按 rune 计）
	OverlapSize int // 相邻块重叠字符数
}

// DefaultChunkOptions 默认分块配置（中文约 500 字）
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{
		ChunkSize:   500,
		OverlapSize: 80,
	}
}

// ChunkText 将文本按段落智能分块
// 策略：
//  1. 先按段落（换行）切分
//  2. 贪心合并段落到接近 ChunkSize
//  3. 超长段落按句子边界二次切分
//  4. 相邻块之间保留 OverlapSize 重叠
func ChunkText(text string, opts ChunkOptions) []string {
	if opts.ChunkSize <= 0 {
		opts = DefaultChunkOptions()
	}
	if opts.OverlapSize < 0 {
		opts.OverlapSize = 0
	}

	// 规范化空白
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// 1. 按段落切分
	paragraphs := strings.Split(text, "\n")
	var chunks []string
	var current strings.Builder
	currentLen := 0

	flush := func() {
		if currentLen == 0 {
			return
		}
		chunks = append(chunks, current.String())
		// 保留重叠
		overlap := ""
		if opts.OverlapSize > 0 && currentLen > opts.OverlapSize {
			runes := []rune(current.String())
			overlap = string(runes[len(runes)-opts.OverlapSize:])
		}
		current.Reset()
		currentLen = 0
		if overlap != "" {
			current.WriteString(overlap)
			currentLen = utf8.RuneCountInString(overlap)
		}
	}

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		paraLen := utf8.RuneCountInString(para)

		// 当前块 + 换行 + 段落 超出限制 → 先刷出当前块
		if currentLen > 0 && currentLen+1+paraLen > opts.ChunkSize {
			flush()
		}

		// 段落本身超过 ChunkSize → 按句子/窗口二次切分
		if paraLen > opts.ChunkSize {
			// 先刷出当前累积
			flush()
			for _, sub := range splitLongText(para, opts) {
				sub = strings.TrimSpace(sub)
				if sub != "" {
					chunks = append(chunks, sub)
				}
			}
			continue
		}

		// 正常累积段落
		if currentLen > 0 {
			current.WriteString("\n")
			currentLen++
		}
		current.WriteString(para)
		currentLen += paraLen
	}
	flush()

	return chunks
}

// splitLongText 对超长文本按句子边界和固定窗口切分
func splitLongText(text string, opts ChunkOptions) []string {
	runes := []rune(text)
	size := opts.ChunkSize
	overlap := opts.OverlapSize

	var result []string
	pos := 0
	total := len(runes)

	for pos < total {
		end := pos + size
		if end > total {
			end = total
		}

		// 尝试在句子边界（。！？；\n 等）处断开
		if end < total {
			for i := end - 1; i > pos+size/2; i-- {
				if isSentenceBoundary(runes[i]) {
					end = i + 1
					break
				}
			}
		}

		chunk := strings.TrimSpace(string(runes[pos:end]))
		if chunk != "" {
			result = append(result, chunk)
		}

		// 下一块起点：end - overlap（如果还有后续内容）
		if end >= total {
			break
		}
		nextPos := end - overlap
		if nextPos <= pos {
			nextPos = end
		}
		pos = nextPos
	}

	return result
}

// isSentenceBoundary 判断是否为句子边界字符
func isSentenceBoundary(r rune) bool {
	switch r {
	case '。', '！', '？', '；', ';', '\n', '\r', '.', '!', '?':
		return true
	default:
		return unicode.IsSpace(r)
	}
}
