package main

import (
	"testing"
	"time"
)

func TestParseBasic(t *testing.T) {
	s, err := Parse("0 9 * * 1-5")
	if err != nil {
		t.Fatalf("Parse 报错: %v", err)
	}
	if !s.Minute[0] || !s.Hour[9] {
		t.Fatal("分0 时9 应命中")
	}
	if len(s.Week) != 5 {
		t.Fatalf("工作日应为 5 天, got %d", len(s.Week))
	}
}

func TestParseStep(t *testing.T) {
	s, err := Parse("*/15 * * * *")
	if err != nil {
		t.Fatalf("Parse 报错: %v", err)
	}
	for _, m := range []int{0, 15, 30, 45} {
		if !s.Minute[m] {
			t.Fatalf("分钟 %d 应命中", m)
		}
	}
	if len(s.Minute) != 4 {
		t.Fatalf("应为 4 个分钟档, got %d", len(s.Minute))
	}
}

func TestParseInvalidFieldCount(t *testing.T) {
	if _, err := Parse("0 9 * *"); err == nil {
		t.Fatal("4 字段应报错")
	}
}

func TestParseRangeError(t *testing.T) {
	if _, err := Parse("0 9 40 * *"); err == nil {
		t.Fatal("日=40 越界应报错")
	}
	if _, err := Parse("99 9 * * *"); err == nil {
		t.Fatal("分=99 越界应报错")
	}
}

func TestNextNWeekdays(t *testing.T) {
	s, _ := Parse("0 9 * * 1-5")
	base := time.Date(2026, 8, 8, 0, 0, 0, 0, time.Local) // 周六
	next := s.NextN(base, 3)
	// 周六 00:00 -> 下一个工作日是周一 09:00
	if len(next) != 3 {
		t.Fatalf("应有 3 次, got %d", len(next))
	}
	if next[0].Weekday() != time.Monday || next[0].Hour() != 9 || next[0].Minute() != 0 {
		t.Fatalf("第一次应为周一 09:00, got %s", next[0].Format("Mon 15:04"))
	}
}

func TestNextEvery15Min(t *testing.T) {
	s, _ := Parse("*/15 * * * *")
	base := time.Date(2026, 8, 6, 10, 7, 0, 0, time.Local)
	next := s.NextN(base, 3)
	want := []string{"10:15", "10:30", "10:45"}
	for i, w := range want {
		got := next[i].Format("15:04")
		if got != w {
			t.Fatalf("第%d次应为 %s, got %s", i+1, w, got)
		}
	}
}

func TestMatchOrSemantics(t *testing.T) {
	// 日=1 或 周=1（周一）——只要满足其一即命中
	s, _ := Parse("0 0 1 * 1")
	t1 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local) // 8/1 是周六，但日=1 命中
	t2 := time.Date(2026, 8, 3, 0, 0, 0, 0, time.Local) // 8/3 是周一，周=1 命中
	t3 := time.Date(2026, 8, 5, 0, 0, 0, 0, time.Local) // 周三，都不命中
	if !s.match(t1) || !s.match(t2) {
		t.Fatal("日=1 或 周=1 应命中")
	}
	if s.match(t3) {
		t.Fatal("周三非1号不应命中")
	}
}

func TestPrevNWeekdays(t *testing.T) {
	s, _ := Parse("0 9 * * 1-5") // 工作日早 9 点
	// 周三 10:00 往前，最近一次应为当天 09:00
	base := time.Date(2026, 8, 5, 10, 0, 0, 0, time.Local)
	prev := s.PrevN(base, 3)
	if len(prev) != 3 {
		t.Fatalf("应有 3 次, got %d", len(prev))
	}
	if prev[0].Weekday() != time.Wednesday || prev[0].Hour() != 9 || prev[0].Minute() != 0 {
		t.Fatalf("最近一次应为周三 09:00, got %s", prev[0].Format("Mon 15:04"))
	}
	if prev[1].Weekday() != time.Tuesday || prev[2].Weekday() != time.Monday {
		t.Fatalf("再往前应为周二、周一, got %s / %s", prev[1].Format("Mon"), prev[2].Format("Mon"))
	}
}

func TestPrevEvery15Min(t *testing.T) {
	s, _ := Parse("*/15 * * * *")
	base := time.Date(2026, 8, 6, 10, 37, 0, 0, time.Local)
	prev := s.PrevN(base, 3)
	want := []string{"10:30", "10:15", "10:00"}
	for i, w := range want {
		if prev[i].Format("15:04") != w {
			t.Fatalf("第%d次应为 %s, got %s", i+1, w, prev[i].Format("15:04"))
		}
	}
}

func TestPrevAtExactHit(t *testing.T) {
	s, _ := Parse("0 9 * * *") // 每天 9:00
	// 起点恰好命中，Prev 应返回该时刻之前的上一次（前一天 9:00）
	base := time.Date(2026, 8, 6, 9, 0, 0, 0, time.Local)
	prev := s.Prev(base)
	if prev.Hour() != 9 || prev.Minute() != 0 {
		t.Fatalf("应返回 9:00, got %s", prev.Format("15:04"))
	}
	if prev.Day() != 5 {
		t.Fatalf("应为前一天 8/5, got %s", prev.Format("2006-01-02"))
	}
}
