# echo examples

examples for gracefully restart, upgrade through http request.

```bash
git clone https://github.com/vogo/gracego.git
cd gracego/examples/echo
make zip

cd build

# hard link
ln echo myecho

# start echo server
./myecho

# start a new server to replace the old
./myecho

# restart through signal
ps -ef |grep -v grep |grep echo
kill -HUP <PID>

# sleep 5s request
curl http://127.0.0.1:8081/sleep5s

# calculate 5s request
curl http://127.0.0.1:8081/calcuate5s

# upgrade through http request
curl http://127.0.0.1:8081/upgrade
```