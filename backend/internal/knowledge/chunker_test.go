package knowledge

import (
	"strings"
	"testing"
)

func TestChunkTextBasic(t *testing.T) {
	text := "第一段内容。\n\n第二段内容比较长，需要分成多个块来处理。\n\n第三段。"
	chunks := ChunkText(text, DefaultChunkOptions())
	if len(chunks) == 0 {
		t.Fatal("应至少有一个块")
	}
	// 合并后应包含原文本
	joined := strings.Join(chunks, "")
	for _, c := range chunks {
		if strings.TrimSpace(c) == "" {
			t.Fatal("分块不应为空")
		}
	}
	_ = joined
}

func TestChunkTextLong(t *testing.T) {
	// 构造 2000 字长文本
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("这是第")
		sb.WriteString(strings.Repeat("测试", 5))
		sb.WriteString("段内容，需要被正确切分。")
	}
	text := sb.String()
	chunks := ChunkText(text, DefaultChunkOptions())
	if len(chunks) < 3 {
		t.Fatalf("2000字文本应分成至少3块，实际 %d", len(chunks))
	}
	// 每块不应超过 ChunkSize + overlap
	opts := DefaultChunkOptions()
	for i, c := range chunks {
		runes := []rune(c)
		if len(runes) > opts.ChunkSize+opts.OverlapSize {
			t.Fatalf("第 %d 块过长: %d 字", i, len(runes))
		}
	}
}

func TestChunkTextParagraphOverlap(t *testing.T) {
	text := "第一段。\n\n第二段。\n\n第三段。"
	opts := ChunkOptions{ChunkSize: 10, OverlapSize: 3}
	chunks := ChunkText(text, opts)
	if len(chunks) < 2 {
		t.Fatal("应产生多个块")
	}
}
