package handlers

import (
	"testing"
)

// isValidTransition 与 handler 内联校验逻辑保持一致
func isValidTransition(cur, target string) bool {
	allowed, ok := requirementStatusTransitions[cur]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

func TestStatusTransitions_Legal(t *testing.T) {
	cases := []struct{ cur, next string }{
		{"draft", "backlog"},
		{"backlog", "prd"},
		{"backlog", "rejected"},
		{"prd", "dev"},
		{"prd", "rejected"},
		{"dev", "test"},
		{"dev", "rejected"},
		{"test", "verify"},
		{"test", "rejected"},
		{"verify", "review"},
		{"verify", "rejected"},
		{"review", "done"},
		{"rework", "backlog"},
		{"rework", "prd"},
		{"rework", "dev"},
		{"rework", "test"},
		{"rework", "verify"},
		{"rework", "review"},
	}
	for _, c := range cases {
		if !isValidTransition(c.cur, c.next) {
			t.Errorf("期望合法流转 %s -> %s，但被判定为非法", c.cur, c.next)
		}
	}
}

func TestStatusTransitions_Illegal(t *testing.T) {
	cases := []struct{ cur, next string }{
		{"draft", "prd"},
		{"draft", "done"},
		{"backlog", "dev"},
		{"dev", "review"},
		{"test", "done"},
		{"done", "backlog"},
		{"done", "prd"},
		{"rejected", "prd"},
		{"rejected", "dev"},
		{"prd", "prd"},
		{"backlog", "backlog"},
		{"unknown", "draft"},
		{"draft", "unknown"},
	}
	for _, c := range cases {
		if isValidTransition(c.cur, c.next) {
			t.Errorf("期望非法流转 %s -> %s，但被判定为合法", c.cur, c.next)
		}
	}
}

func TestStatusTransitions_AllStatesCovered(t *testing.T) {
	all := []string{"draft", "backlog", "prd", "dev", "test", "verify", "review", "rework", "rejected", "done"}
	for _, s := range all {
		if _, ok := requirementStatusTransitions[s]; !ok {
			t.Errorf("状态 %s 未定义在状态机中", s)
		}
	}
}

func TestStagePlan_EachStatus(t *testing.T) {
	cases := []struct {
		status string
		want   map[string]string
	}{
		{"draft", map[string]string{"A": "doing", "B": "todo", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}},
		{"backlog", map[string]string{"A": "done", "B": "todo", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}},
		{"prd", map[string]string{"A": "done", "B": "doing", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}},
		{"dev", map[string]string{"A": "done", "B": "done", "C": "doing", "D": "doing", "E": "todo", "F": "todo", "G": "todo"}},
		{"test", map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "doing", "G": "todo"}},
		{"verify", map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "done", "G": "doing"}},
		{"review", map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "done", "G": "done"}},
		{"done", map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "done", "G": "done"}},
		{"rejected", map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "done", "G": "done"}},
	}
	for _, c := range cases {
		got := pipelineStagePlan(c.status)
		for _, st := range pipelineStageDefs {
			if got[st] != c.want[st] {
				t.Errorf("状态 %s 环节 %s：期望 %s 实际 %s", c.status, st, c.want[st], got[st])
			}
		}
	}
}

func TestStagePlan_Values(t *testing.T) {
	valid := map[string]bool{"todo": true, "doing": true, "done": true, "na": true, "rework": true}
	for _, s := range []string{"draft", "backlog", "prd", "dev", "test", "verify", "review", "done", "rejected", "unknown"} {
		plan := pipelineStagePlan(s)
		for _, st := range pipelineStageDefs {
			if !valid[plan[st]] {
				t.Errorf("状态 %s 环节 %s 出现非法值 %q", s, st, plan[st])
			}
		}
	}
}

func TestContainsString(t *testing.T) {
	list := []string{"A", "B", "C"}
	if !containsString(list, "A") {
		t.Error("期望包含 A")
	}
	if containsString(list, "D") {
		t.Error("不应包含 D")
	}
	if containsString(nil, "A") {
		t.Error("空列表不应包含任何值")
	}
}

// computeReject 模拟 CreateReview 打回时目标环节后全部置 todo 的逻辑
func computeReject(target string) map[string]string {
	m := map[string]string{}
	idx := -1
	for i, s := range pipelineStageDefs {
		if s == target {
			idx = i
		}
	}
	for i, stage := range pipelineStageDefs {
		switch {
		case i == idx:
			m[stage] = "doing"
		case i > idx:
			m[stage] = "todo"
		default:
			m[stage] = "done"
		}
	}
	return m
}

func TestRejectStageReset(t *testing.T) {
	got := computeReject("C")
	want := map[string]string{"A": "done", "B": "done", "C": "doing", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("打回 C：环节 %s 期望 %s 实际 %s", k, v, got[k])
		}
	}
	got = computeReject("A")
	if got["A"] != "doing" || got["B"] != "todo" || got["G"] != "todo" {
		t.Errorf("打回 A 结果错误: %v", got)
	}
}
