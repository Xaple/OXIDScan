# OXID

通过windows的DCOM接口进行网卡进行信息枚举，定位多网卡主机。

```
Usage of ./OXIDScan.exe:
  -i string
    	single ip address
  -n string
    	CIDR notation of a network
  -f string
    	file containing IP addresses or network CIDRs
  -t int
    	thread num (default 2000)
  -time duration
    	timeout on connection, in seconds (default 2ns)


./OXIDScan.exe -i 192.168.1.1
./OXIDScan.exe -n 192.168.1.1/24
./OXIDScan.exe -f ip.txt
```


