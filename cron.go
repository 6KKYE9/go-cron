package main

// cron.go：cron 表达式解析与触发计算。
// 标准 5 字段：分 时 日 月 周（周 0=周日）。支持 * 、, 、- 、*/step。

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule 是一条 cron 计划。
type Schedule struct {
	Minute map[int]bool
	Hour   map[int]bool
	Day    map[int]bool
	Month  map[int]bool
	Week   map[int]bool
}

func newSet(min, max int) map[int]bool {
	m := make(map[int]bool)
	for i := min; i <= max; i++ {
		m[i] = true
	}
	return m
}

func parseField(field string, min, max int) (map[int]bool, error) {
	set := make(map[int]bool)
	if field == "*" {
		return newSet(min, max), nil
	}
	for _, part := range strings.Split(field, ",") {
		// 支持 */step
		step := 1
		rangePart := part
		if idx := strings.Index(part, "/"); idx >= 0 {
			rangePart = part[:idx]
			s, err := strconv.Atoi(part[idx+1:])
			if err != nil || s <= 0 {
				return nil, fmt.Errorf("步长错误: %q", part)
			}
			step = s
		}
		lo, hi := min, max
		if rangePart != "*" {
			if idx := strings.Index(rangePart, "-"); idx >= 0 {
				a, err1 := strconv.Atoi(rangePart[:idx])
				b, err2 := strconv.Atoi(rangePart[idx+1:])
				if err1 != nil || err2 != nil || a < min || b > max || a > b {
					return nil, fmt.Errorf("范围错误: %q", part)
				}
				lo, hi = a, b
			} else {
				v, err := strconv.Atoi(rangePart)
				if err != nil || v < min || v > max {
					return nil, fmt.Errorf("数值越界: %q", part)
				}
				lo, hi = v, v
			}
		}
		for v := lo; v <= hi; v += step {
			set[v] = true
		}
	}
	return set, nil
}

// Parse 解析 5 字段 cron 表达式。
func Parse(expr string) (*Schedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("需要 5 个字段，得到 %d 个", len(fields))
	}
	min, err1 := parseField(fields[0], 0, 59)
	hour, err2 := parseField(fields[1], 0, 23)
	day, err3 := parseField(fields[2], 1, 31)
	month, err4 := parseField(fields[3], 1, 12)
	week, err5 := parseField(fields[4], 0, 6)
	if err1 != nil {
		return nil, fmt.Errorf("分钟字段: %w", err1)
	}
	if err2 != nil {
		return nil, fmt.Errorf("小时字段: %w", err2)
	}
	if err3 != nil {
		return nil, fmt.Errorf("日字段: %w", err3)
	}
	if err4 != nil {
		return nil, fmt.Errorf("月字段: %w", err4)
	}
	if err5 != nil {
		return nil, fmt.Errorf("周字段: %w", err5)
	}
	return &Schedule{Minute: min, Hour: hour, Day: day, Month: month, Week: week}, nil
}

// match 判断某时间是否命中计划（日与周为「或」关系，符合标准 cron 语义）。
func (s *Schedule) match(t time.Time) bool {
	if !s.Month[int(t.Month())] {
		return false
	}
	if !s.Minute[t.Minute()] || !s.Hour[t.Hour()] {
		return false
	}
	// 日与周：任一个匹配即可（标准 cron：两字段同时非 * 时为 OR）
	dayStar := len(s.Day) == 31
	weekStar := len(s.Week) == 7
	if dayStar && weekStar {
		return true
	}
	dayOK := s.Day[t.Day()]
	weekOK := s.Week[int(t.Weekday())]
	if dayStar {
		return weekOK
	}
	if weekStar {
		return dayOK
	}
	return dayOK || weekOK
}

// Next 从 t（含）起向后找下一个命中时间，最多找 4 年。
func (s *Schedule) Next(t time.Time) time.Time {
	t = t.Truncate(time.Minute)
	if s.match(t) {
		// 若当前分钟已命中，仍返回（含起点）
	}
	for i := 0; i < 4*365*24*60; i++ {
		if s.match(t) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

// NextN 返回从 t 起接下来的 n 个触发时间。
func (s *Schedule) NextN(t time.Time, n int) []time.Time {
	out := make([]time.Time, 0, n)
	cur := t.Truncate(time.Minute)
	// 若起点本身命中，从下一分钟开始避免重复计入（更符合“接下来”语义）
	cur = cur.Add(time.Minute)
	for len(out) < n {
		nt := s.Next(cur)
		if nt.IsZero() {
			break
		}
		out = append(out, nt)
		cur = nt.Add(time.Minute)
	}
	return out
}

// Describe 用中文描述表达式含义。
func (s *Schedule) Describe() string {
	min := setSummary(s.Minute, 0, 59)
	hour := setSummary(s.Hour, 0, 23)
	day := setSummary(s.Day, 1, 31)
	month := setSummary(s.Month, 1, 12)
	week := weekSummary(s.Week)
	return fmt.Sprintf("分[%s] 时[%s] 日[%s] 月[%s] 周[%s]", min, hour, day, month, week)
}

func setSummary(m map[int]bool, min, max int) string {
	vals := make([]int, 0, len(m))
	for v := range m {
		vals = append(vals, v)
	}
	// 简单排序
	for i := 0; i < len(vals); i++ {
		for j := i + 1; j < len(vals); j++ {
			if vals[j] < vals[i] {
				vals[i], vals[j] = vals[j], vals[i]
			}
		}
	}
	if len(vals) == max-min+1 {
		return "*"
	}
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ",")
}

func weekSummary(m map[int]bool) string {
	if len(m) == 7 {
		return "*"
	}
	names := []string{"日", "一", "二", "三", "四", "五", "六"}
	parts := make([]string, 0, len(m))
	for _, v := range []int{0, 1, 2, 3, 4, 5, 6} {
		if m[v] {
			parts = append(parts, names[v])
		}
	}
	return strings.Join(parts, "/")
}
