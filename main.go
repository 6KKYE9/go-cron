// go-cron 是 cron 表达式解析器与下次触发计算器。
// 支持标准 5 字段格式（分 时 日 月 周），可解释表达式含义，
// 并从给定时间起算出最近 N 次触发时间。
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	expr := flag.String("e", "", "cron 表达式（分 时 日 月 周，5 字段）")
	from := flag.String("from", "", "起始时间，格式 2006-01-02 15:04（默认现在）")
	n := flag.Int("n", 5, "显示接下来几次触发")
	tz := flag.String("tz", "", "时区名（如 Asia/Shanghai、America/New_York，默认本地时区）")
	last := flag.Bool("last", false, "同时显示最近一次触发时间")
	help := flag.Bool("h", false, "显示帮助")
	flag.Parse()

	if *help || *expr == "" {
		fmt.Print(`go-cron cron 表达式解析 / 触发时间计算（零依赖）

用法:
  go-cron -e "0 9 * * 1-5"                 # 工作日早 9 点
  go-cron -e "*/15 * * * *" -n 3           # 每 15 分钟，看接下来 3 次
  go-cron -e "30 8 1 * *" -from "2026-08-06 00:00"
  go-cron -e "0 9 * * *" -tz "America/New_York" -last   # 指定时区并看上次触发

字段: 分(0-59) 时(0-23) 日(1-31) 月(1-12) 周(0-6, 0=周日)
`)
		return
	}

	sched, err := Parse(*expr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "解析失败:", err)
		os.Exit(1)
	}
	fmt.Println("表达式:", *expr)
	fmt.Println("含义  :", sched.Describe())

	loc := time.Local
	if *tz != "" {
		l, err := time.LoadLocation(*tz)
		if err != nil {
			fmt.Fprintln(os.Stderr, "时区解析失败:", err)
			os.Exit(1)
		}
		loc = l
	}

	base := time.Now().In(loc)
	if *from != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04", *from, loc); err == nil {
			base = t
		} else {
			fmt.Fprintln(os.Stderr, "起始时间解析失败:", err)
			os.Exit(1)
		}
	}

	if *last {
		prev := sched.Prev(base)
		if prev.IsZero() {
			fmt.Println("最近一次触发: （4 年内未找到）")
		} else {
			fmt.Println("最近一次触发:", prev.Format("2006-01-02 15:04 Mon"))
		}
	}

	times := sched.NextN(base, *n)
	fmt.Printf("接下来 %d 次触发（自 %s 起，时区 %s）:\n", len(times), base.Format("2006-01-02 15:04"), loc)
	for i, t := range times {
		fmt.Printf("  %d. %s\n", i+1, t.Format("2006-01-02 15:04 Mon"))
	}
}
