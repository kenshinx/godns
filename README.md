GODNS
====

A simple and fast dns cache server written by go.


Similar as [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) ,but support some difference features:


* Keep hosts records in redis instead of the local file /etc/hosts  

* Atuo-Reload when hosts configuration changed. (Yes,dnsmasq need restart)

* Cache records save in memory or redis configurable


## Install & Running

1. Install  

		$ go get github.com/kenshinx/godns


2. Build  

		$ cd $GOPATH/src/github.com/kenshinx/godns 
		$ go build -o godns *.go


3. Running  

		$ sudo ./godns -c godns.conf


4. Use

		$ sudo vi /etc/resolv.conf
		nameserver 127.0.0.1



## Configuration

All the configuration on `godns.conf` a TOML formating config file.   
More about Toml :[https://github.com/mojombo/toml](https://github.com/mojombo/toml)


#### resolv.conf

Upstream server can be configuration by change file from somewhere other that "/etc/resolv.conf"

```
[resolv]
resolv-file = "/etc/resolv.conf"
```
If multi `namerserver` set at resolv.conf, the upsteam server will try in order of up to botton



#### cache

Only the local memory storage backend implemented now.  The redis backend is in todo list

```
[cache]
backend = "memory"   
expire = 600  # default expire time 10 minutes
maxcount = 100000
```



#### hosts

Force resolv domain to assigned ip, support two types hosts configuration:

* locale hosts file
* remote redis hosts

__hosts file__  

can be assigned at godns.conf,default : `/etc/hosts`

```
[hosts]
host-file = "/etc/hosts"
```


__redis hosts__ 

This is a espeical requirment in our system. Must maintain a gloab hosts configuration, 
and support update the hosts record from other remote server.
so "redis-hosts" is be supported, and will query the reids when each dns request reached.  

The hosts record is organized with redis hash map. and the key the map is configired.

```
[hosts]
redis-key = "godns:hosts"
```

_Insert hosts records into redis_

```
redis > hset godns:hosts www.sina.com.cn 1.1.1.1
```



## Benchmak


__Debug close__

```
$ go test -bench=.

testing: warning: no tests to run
PASS
BenchmarkDig-4	   10000	    202732 ns/op
ok  	_/Users/kenshin/workspace/gogo/godns	2.489s
```

The result : 4032 queries/per second

The enviroment of test:

MacBook Air 

* CPU:  
Inter Core i5 1.7G  
Double cores

* MEM:  
8G




## TODO

* The redis cache backend
* Update ttl





