command
env GOOS=windows GOARCH=amd64 go build -o resolver
./icloudverify -host mx01.mail.icloud.com:25 -b 1 -i input.csv -o apple.csv -v -greeting smtpclient.apple -workers 120 -delay 20
