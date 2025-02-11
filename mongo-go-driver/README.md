# Q & A 

## How do you get IP from a spawnhost for whitelisting Atlas?

If the spawn host is on a private network, there's not a local utility that can tell you what the public IP is as it appears to other Internet hosts. The best way is to use a service like ifconfig.me:
```
curl -4 ifconfig.me

```
