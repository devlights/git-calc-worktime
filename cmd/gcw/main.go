package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
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

const (
	dateFormat = "2006-01-02 15:04:05 -0700"
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
		localTz *time.Location
		err     error
	)
	localTz, err = time.LoadLocation(args.timeZone)
	if err != nil {
		return fmt.Errorf("ローカルタイムゾーン取得エラー: %w", err)
	}

	var (
		// git --no-pager -C /path/to/repository log --author=git-user-name --format="%H %ai"
		gitCmd    = exec.Command("git", "--no-pager", "-C", args.dir, "log", fmt.Sprintf("--author=%s", args.userName), "--format=%H %ai")
		cmdStdout io.ReadCloser
	)
	cmdStdout, err = gitCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("gitコマンド実行エラー: %w", err)
	}

	err = gitCmd.Start()
	if err != nil {
		return fmt.Errorf("gitコマンド実行エラー: %w", err)
	}

	var (
		workweek = make(map[int]int)
		weekend  = make(map[int]int)
		scanner  = bufio.NewScanner(cmdStdout)
	)
	for scanner.Scan() {
		var (
			line      = scanner.Text()
			fields    = strings.Fields(line) // e.g., e71ef8f2ea60e651c272cb51127121e1d7597928 2024-08-15 16:54:57 +0900
			timestamp time.Time
		)
		if len(fields) < 2 {
			continue
		}

		timestamp, err = time.Parse(dateFormat, fields[1]+" "+fields[2]+" "+fields[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "日付解析エラー (行: %s): %v\n", line, err)
			continue
		}

		var (
			localTime = timestamp.In(localTz)
			hour      = localTime.Hour()
		)
		switch localTime.Weekday() {
		case time.Saturday:
			fallthrough
		case time.Sunday:
			weekend[hour]++
		default:
			workweek[hour]++
		}
	}

	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("読み取りエラー: %w", err)
	}

	err = gitCmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("gitコマンドが非ゼロのステータスで終了: %d, エラー: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("gitコマンド実行エラー: %w", err)
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
