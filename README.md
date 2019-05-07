**gracego enables gracefully restart or upgrade golang application.**

## example

```bash
git clone https://github.com/vogo/gracego.git
cd gracego/examples/echo

go build -o echo_v1
ln echo_v1 echo

./echo

# restart through signal
ps -ef |grep -v grep |grep echo
kill -HUP <PID>

# upgrade through http request
curl http://127.0.0.1:8081/upgrade

```