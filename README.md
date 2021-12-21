GODNS
====

A simple and fast dns cache server written by go.


Similar to [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html), but supports some difference features:


* Keep hosts records in redis and the local file /etc/hosts  

* Auto-Reloads when hosts configuration is changed. (Yes, dnsmasq needs to be reloaded)


## Installation & Running

1. Install  

		$ go get github.com/kenshinx/godns


2. Build  

		$ cd $GOPATH/src/github.com/kenshinx/godns 
		$ go build -o godns 


3. Running  

		$ sudo ./godns -c ./etc/godns.conf

4. Test
        
        $ dig www.github.com @127.0.0.1



## Use godns 

		$ sudo vi /etc/resolv.conf
		nameserver #the ip of godns running

## Configuration

All the configuration in `godns.conf` is a TOML format config file.   
More about Toml :[https://github.com/mojombo/toml](https://github.com/mojombo/toml)


#### resolv.conf

Upstream server can be configured by changing file from somewhere other than "/etc/resolv.conf"

```
[resolv]
resolv-file = "/etc/resolv.conf"
```
If multiple `namerservers` are set in resolv.conf, the upsteam server will try in a top to bottom order


#### server-list-file
Domain-specific nameservers configuration, formatting keep compatible with Dnsmasq.
>server=/google.com/8.8.8.8

More cases please refererence [dnsmasq-china-list](https://github.com/felixonmars/dnsmasq-china-list)


#### audit

Only redis storage backend is currently implemented.

Audit logs are in format:

```
{ "remoteaddr": "127.0.0.1", "domain": "domain.com", "qtype": "A", "timestamp": "2019-04-15T12:16:21.875492605Z" }
```

Backend uses lists to store logs in redis.

Logs grouped by hour.

Redis keys have format:

```
audit-YYYY-MM-DDTHH:00
```

Example request to get audit logs from redis:

```
LRANGE audit-2019-04-15T00:00 0 -1
```


#### cache

Only the local memory storage backend is currently implemented.  The redis backend is in the todo list

```
[cache]
backend = "memory"   
expire = 600  # default expire time 10 minutes
maxcount = 100000
```



#### hosts

Force resolve domain to assigned ip, support two types hosts configuration:

* locale hosts file
* remote redis hosts

__hosts file__  

can be assigned at godns.conf,default : `/etc/hosts`

```
[hosts]
host-file = "/etc/hosts"
```
Hosts file format is described in [linux man pages](http://man7.org/linux/man-pages/man5/hosts.5.html). 
More than that , `*.` wildcard is supported additional.


__redis hosts__ 

This is a special requirment in our system. Must maintain a global hosts configuration, 
and support update the host records from other remote server.
Therefore, while "redis-hosts" be enabled, will query the redis db when each dns request is reached.  

The hosts record is organized with redis hash map. and the key of the map is configured.

```
[hosts]
redis-key = "godns:hosts"
```

_Insert hosts records into redis_

```
redis > hset godns:hosts www.test.com 1.1.1.1
```

Compared with file-backend records, redis-backend hosts support multiple A entries.

```
redis > hset godns:hosts www.test.com 1.1.1.1,2.2.2.2
```


## Benchmark


__Debug close__

```
$ go test -bench=.

testing: warning: no tests to run
PASS
BenchmarkDig-8     50000             57945 ns/op
ok      _/usr/home/keqiang/godns        3.259s
```

The result : 15342 queries/per second

The test environment:

CentOS release 6.4 

* CPU:  
Intel Xeon 2.40GHZ 
4 cores

* MEM:  
46G


## Web console

Joke: A web console for godns

[https://github.com/kenshinx/joke](https://github.com/kenshinx/joke) 

screenshot

![joke](https://raw.github.com/kenshinx/joke/master/screenshot/joke.png)



## Deployment

Deployment in productive supervisord highly recommended.

```

[program:godns]
command=/usr/local/bin/godns -c /etc/godns.conf
autostart=true
autorestart=true
user=root
stdout_logfile_maxbytes = 50MB
stdoiut_logfile_backups = 20
stdout_logfile = /var/log/godns.log

```


## TODO

* The redis cache backend
* Update ttl

## LICENSE
godns is under the MIT license. See the LICENSE file for details.



