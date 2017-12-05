# Befor Usage
1. client_secret.jsonを取得してフォルダに設置
2. config.toml.skeltonをconfig.tomlにリネームし、適切な情報(calendarID, 名前)を入力する

# Usage
1. `go run main.go`
2. URLが表示されるのでブラウザからアクセス(初回だけ)
3. client_secret.jsonを取得したGoogleアカウントUserで許可をする(初回だけ)
4. コードが表示されるのでコピー(初回だけ)
5. コンソールに貼り付けてエンター(初回だけ)
6. ~/.credentials/calendar-go-quickstart.jsonが生成される
  - このファイルを削除すると2～6の手順をもう一度行わなければ利用できない