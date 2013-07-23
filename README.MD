GODNS
====

A tiny dns cache server written by go.


Similar as [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) ,but support some difference features:


* Keep hosts configuration in redis instead of local file /etc/hosts  
  So can be updated from remote server    

* Atuo-Reload when hosts configuration changed. (Yes,dnsmasq need restart)

* Cache records save in memory or redis configurable

