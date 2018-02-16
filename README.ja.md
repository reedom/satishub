satishub
========

[satis][]実行サーバ。

これはなに？
-----------

- [GitLab][]の更新に合わせて[satis][]を実行して[PHP Composer][]用リポジトリ情報を生成します
- [PHP Composer][]等の求めに応じてリポジトリ情報を返却します

[GitLab]: https://gitlab.com
[satis]: https://getcomposer.org/doc/articles/handling-private-packages-with-satis.md
[PHP composer]: https://getcomposer.org/

ビルド
------

・`make`を実行することでメニューを表示します。

    $ make

    satis-hub manager 0.0.1

             setup: install required tools
              deps: install dependant packages
               fmt: re-format go source code files
              lint: run golint
              test: run tests
             build: Build satishub for development on the current environment
              dist: Build satishub binaries for distribution
      docker-build: create a new docker image

・ローカルビルド

[Go lang][]が必要です。

    $ make setup && make deps && make build

・dockerビルド

[docker][]が必要です。

    $ make docker-build

[Go lang]: https://golang.org/
[docker]: https://docker.io/

つかいかた
----------

    # ローカル
    $ bin/satishub serve -h
    # docker
    $ docker run --rm reedom/satishub serve -h

    Start satis service server

    Usage:
      satishub serve [flags]

    Flags:
          --config string          satis config file path (default "satis.json")
      -h, --help                   help for serve
          --repo string            satis output directory path (default "repo")
          --satis string           satis executable path (default "satis")
          --sns-topic-arn string   AWS Simple Notification Service ARN
          --timeout int            satis build process timeout in seconds (default 1200)
          --tlscert string         TLS certificate file path (default "satis.crt")
          --tlskey string          TLS secret key file path (default "satis.key")

    Global Flags:
          --addr string      HTTP service server listen address (default ":80")
          --debug            output verbose messages for debugging
          --no-http          do not setup HTTP server
          --no-tls           do not setup HTTPS server
          --tlsaddr string   TLS(HTTPS) service server listen address (default ":443")

| フラグ        | 環境変数                  | デフォルト | 内容                                  |
|---------------|---------------------------|------------|---------------------------------------|
| no-http       | SATIS_NO_HTTP             | false      | `true`ならHTTPサーバを起動しない      |
| addr          | SATIS_HTTP_ADDR           | :80        | HTTPサーバ起動アドレス                |
| no-tls        | SATIS_NO_TLS              | false      | `true`ならHTTPSサーバを起動しない     |
| tlsaddr       | SATIS_TLS_ADDR            | :443       | HTTPSサーバ起動アドレス               |
| tlscert       | SATIS_TLS_CERT_PATH       | satis.crt  | HTTPS用証明書ファイルへのパス         |
| tlskey        | SATIS_TLS_SECRET_KEY_PATH | satis.key  | HTTPS用証明書の秘密鍵ファイルへのパス |
| satis         | SATIS_EXEC_PATH           | satis      | [satis][]コマンドへのパス             |
| config        | SATIS_CONFIG_PATH         | satis.json | satis用configへのパス                 |
| repo          | SATIS_REPO_PATH           | repo       | satis出力ディレクトリパス             |
| timeout       | SATIS_TIMEOUT             | 1200       | satisビルド最大実行時間（秒）         |
| sns-topic-arn | SATIS_SNS_TOPIC_ARN       | -          | 実行通知用[AWS SNS Topic][]のARN      |

※[AWS SNS Topic]はsatisの実行開始・終了などについて通知を得たい場合に利用します。

[AWS SNS]: https://aws.amazon.com/sns/

・実行例

    $ cat <<EOL > satis.json
    {
      "name": "My Satis Repositoy",
      "homepage": "http://satis.example.com",
      "repositories": [],
      "require-all": true
    }
    EOL

    $ docker run --rm \
        -p 8080:80 \
        -v $PWD/satis.json:/var/satishub/satis.json \
        -v $PWD/repo:/var/satishub/repo \
        --name satishub \
        reedom/satishub serve --no-tls --debug

WEB server API
--------------

| path             | method | 内容                                   |
|------------------|--------|----------------------------------------|
| `/webhook/gitlab | POST   | [GitLab][]リポジトリ用WebHook          |
| その他`/`など    | GET    | [PHP Composer][]向けリポジトリ情報返却 |
| `/config`        | GET    | satis用configの内容を返却              |
