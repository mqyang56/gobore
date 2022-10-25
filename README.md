# gobore
A modern, simple TCP tunnels in Go that exposes local ports to a remote server, bypassing standard NAT connection firewalls. 

## Quick Start
1. Running the follow command on the cloud server[IP: 127.0.0.1].
```shell
docker run --rm --network host --name gobore-server mqyang56/gobore:1.1.5 ./gobore server --secret gobore

# Output
2022-10-25T03:44:38.136Z	debug	gobore/shared.go:75	msg	{"msg": "{\"type\":\"Authenticate\",\"authenticate\":\"dce9c649e6f6240c27c324829e092805673e2400a78db66d6a3d4d8458cdda46\"}"}
2022-10-25T03:44:38.340Z	debug	gobore/shared.go:75	msg	{"msg": "{\"type\":\"Hello\"}"}
2022-10-25T03:44:38.340Z	debug	gobore/server.go:85	Receive clientMessage: gobore.clientMessage{Type:"Hello", Authenticate:"", Hello:0x0, Accept:""}
2022-10-25T03:44:38.340Z	info	gobore/server.go:114	new client	{"port": 16002}
```
2. Running the follow command on the local server[IP: 127.0.0.1].
```shell
docker run --rm --network host --name gobore-client mqyang56/gobore:1.1.5 ./gobore client --secret gobore --local-host 127.0.0.1 --local-port 22 --to 127.0.0.1

# Output
2022-10-25T03:44:38.135Z	debug	gobore/shared.go:75	msg	{"msg": "{\"type\":\"Challenge\",\"challenge\":\"44088020-6bec-4574-8acb-0142b9d996b0\"}"}
2022-10-25T03:44:38.340Z	debug	gobore/shared.go:75	msg	{"msg": "{\"type\":\"Hello\",\"hello\":16002}"}
2022-10-25T03:44:38.340Z	info	gobore/client.go:50	connected to server	{"port": 16002}
2022-10-25T03:44:38.341Z	info	gobore/client.go:51	listening at 	{"host": "127.0.0.1", "port": 22}
```
3. Running the follow command on any hosts to access the local server.
```shell
ssh root@127.0.0.1 -p 16002
```


