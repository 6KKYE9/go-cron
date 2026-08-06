// go-cron 是一个零依赖的 cron 表达式解析器与下次触发计算器。
// 支持标准 5 字段格式（分 时 日 月 周），可解释表达式含义，
// 并从给定时间起算出最近 N 次触发时间，方便验证定时任务配置。
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
	help := flag.Bool("h", false, "显示帮助")
	flag.Parse()

	if *help || *expr == "" {
		fmt.Print(`go-cron cron 表达式解析 / 下次触发计算（零依赖）

用法:
  go-cron -e "0 9 * * 1-5"          # 工作日早 9 点
  go-cron -e "*/15 * * * *" -n 3    # 每 15 分钟，看接下来 3 次
  go-cron -e "30 8 1 * *" -from "2026-08-06 00:00"

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

	base := time.Now()
	if *from != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04", *from, time.Local); err == nil {
			base = t
		} else {
			fmt.Fprintln(os.Stderr, "起始时间解析失败:", err)
			os.Exit(1)
		}
	}

	times := sched.NextN(base, *n)
	fmt.Printf("接下来 %d 次触发（自 %s 起）:\n", len(times), base.Format("2006-01-02 15:04"))
	for i, t := range times {
		fmt.Printf("  %d. %s\n", i+1, t.Format("2006-01-02 15:04 Mon"))
	}
}
