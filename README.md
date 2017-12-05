# Befor Use
1. client_secret.jsonを取得してフォルダに設置
  - 以下の生地を参考に
    - https://qiita.com/konojunya/items/e2611a65d98a9abf1bf7
2. config.toml.skeltonをconfig.tomlにリネームし、適切な情報(calendarID, 名前)を入力する

# Usage
1. main.exeを実行
  - exeファイルは他のコンフィグファイル(client_secret.json, config.toml)と同じフォルダに設置すること。
2. url.txtが生成されるので、そのURLをブラウザからアクセス(初回だけ)
3. client_secret.jsonを取得したGoogleアカウントUserで許可をする(初回だけ)
4. コードが表示されるのでコピー(初回だけ)
5. config.tomlのauthorizationCodeに記入(初回だけ)
6. もう一度main.exeを実行
7. Desktopにファイルが生成される。
  - 自分と同じシフト開始時間のメンバリストを生成