# dots_crawler

[dots](https://eventdots.jp/)の新着イベントをSlackへ流すクローラーです。

SlackはWeb Hookを利用します。
URLとチャンネル名は、環境変数に設定してください。
環境変数は、以下の2つです。

- `SLACK_WEBHOOK_URL`: SlackのWeb Hook URL
- `SLACK_WEBHOOK_CHANNEL`: 投稿するチャンネル名(#付き)
