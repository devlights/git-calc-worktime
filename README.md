# Overview

git log --author="$GIT_USER_NAME" --format="%H %ai" の結果を集計するプログラムです。

```gcw``` は、```Git Calc Worktime``` の略です。

# Usage

```sh
$ gcw --help
Usage of ./gcw:
-dir string
        Path of git repository (default ".")
-tz string
        Local Timezone (default "Asia/Tokyo")
-user string
        Git username
```
```sh
$ gcw -user Gitユーザ名 -dir リポジトリのパス -tz ローカルタイムゾーン(デフォルトはAsia/Tokyo)
hour        Monday to Friday                      Saturday and Sunday
00      0                                     0
01      0                                     0
02      0                                     0
03      0                                     0
04      0                                     0
05      0                                     0
06      0                                     0
07      0                                     0
08      0                                     0
09      4                                     0
10    102 **************                      0
11    115 ***************                     0
12     27 ***                                 0
13    132 ******************                  2
14     92 ************                        4
15    159 *********************              20 **
16    182 *************************          25 ***
17    148 ********************                3
18    167 **********************              8 *
19    167 **********************              4
20     68 *********                           2
21     48 ******                              0
22     17 **                                  0
23      3                                     0

Total:   1431 (95.5%)     68 (4.5%)
```

# Memo

このプログラムは、[以下のブログ記事](https://ivan.bessarabov.com/blog/famous-programmers-work-time-part-2-workweek-vs-weekend)で利用されていた[Perlスクリプト](https://gist.github.com/bessarabov/30aee15c5a7c438fe5f9f3f623222b39)をGoに移植したものです。
元のスクリプトは

	$ git log --author="$GIT_USER_NAME" --format="%H %ai" | perl script.pl

とパイプ経由で入力を受け取り処理するようになっていましたが、Windows環境でも実行しやすいように
処理内で git コマンドも実行するように変更しています。

素晴らしいアイデアを教えてくれた元記事に感謝します。

# Build

[task](https://taskfile.dev/#/)を使っています。

```sh
$ task build
```

# REFERENCES

- [At what time of day do famous programmers work?](https://ivan.bessarabov.com/blog/famous-programmers-work-time)
- [At what time of day do famous programmers work? Part 2. Workweek vs Weekend.](https://ivan.bessarabov.com/blog/famous-programmers-work-time-part-2-workweek-vs-weekend)
- [Script to generate data shown in post 'At what time of day does famous programmers work? Part 2. Workweek vs Weekend.](https://gist.github.com/bessarabov/30aee15c5a7c438fe5f9f3f623222b39) 