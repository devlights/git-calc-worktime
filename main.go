package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	Args struct {
		dir      string
		userName string
		timeZone string
	}
)

var (
	args Args
)

func init() {
	flag.StringVar(&args.dir, "dir", ".", "Path of git repository")
	flag.StringVar(&args.userName, "user", "", "Git username")
	flag.StringVar(&args.timeZone, "tz", "Asia/Tokyo", "Local Timezone")
}

func main() {
	flag.Parse()

	if args.userName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if args.dir == "" {
		args.dir = "."
	}

	if args.timeZone == "" {
		args.timeZone = "Asia/Tokyo"
	}

	var (
		absPath string
		err     error
	)
	absPath, err = filepath.Abs(args.dir)
	if err != nil {
		log.Fatalf("無効なディレクトリ: %s (%v)", args.dir, err)
	}

	args.dir = absPath

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		cmd    = exec.Command("git", "-C", args.dir, "log", fmt.Sprintf("--author=%s", args.userName), "--format=%H %ai")
		output []byte
		err    error
	)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("gitコマンド実行エラー: %w", err)
	}

	var (
		workweek = make(map[int]int)
		weekend  = make(map[int]int)
		localTz  *time.Location
	)
	localTz, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("ローカルタイムゾーン取得エラー: %w", err)
	}

	var (
		reader  = bytes.NewReader(output)
		scanner = bufio.NewScanner(reader)
	)
	for scanner.Scan() {
		var (
			line      = scanner.Text()
			fields    = strings.Fields(line) // e.g., e71ef8f2ea60e651c272cb51127121e1d7597928 2024-08-15 16:54:57 +0900
			timestamp time.Time
			localTime time.Time
		)
		if len(fields) < 2 {
			continue
		}

		timestamp, err = time.Parse("2006-01-02 15:04:05 -0700", fields[1]+" "+fields[2]+" "+fields[3])
		if err != nil {
			fmt.Printf("日付解析エラー: %v\n", err)
			continue
		}

		localTime = timestamp.In(localTz)
		switch localTime.Weekday() {
		case time.Saturday:
			fallthrough
		case time.Sunday:
			weekend[localTime.Hour()]++
		default:
			workweek[localTime.Hour()]++
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("読み取りエラー: %w", err)
	}

	printGraph(workweek, weekend)

	return nil
}

func printGraph(workweek, weekend map[int]int) {
	fmt.Printf("%-6s %6s %-30s %6s %-30s\n", "hour", "", "Monday to Friday", "", "Saturday and Sunday")

	var (
		max  = 0
		hour = 0
	)
	for hour = 0; hour < 24; hour++ {
		if max < workweek[hour] {
			max = workweek[hour]
		}

		if max < weekend[hour] {
			max = weekend[hour]
		}
	}

	for hour = 0; hour < 24; hour++ {
		var (
			workweekCount = workweek[hour]
			weekendCount  = weekend[hour]
			workweekStars = strings.Repeat("*", int(float64(workweekCount)/float64(max)*25))
			weekendStars  = strings.Repeat("*", int(float64(weekendCount)/float64(max)*25))
		)
		fmt.Printf("%02d %6d %-30s %6d %-30s\n", hour, workweekCount, workweekStars, weekendCount, weekendStars)
	}

	var (
		totalWorkweek = sum(workweek)
		totalWeekend  = sum(weekend)
		total         = totalWorkweek + totalWeekend
	)
	fmt.Printf("\nTotal: %6d (%.1f%%) %6d (%.1f%%)\n",
		totalWorkweek, float64(totalWorkweek)*100/float64(total),
		totalWeekend, float64(totalWeekend)*100/float64(total))
}

func sum(m map[int]int) int {
	var (
		total = 0
	)
	for _, v := range m {
		total += v
	}

	return total
}
