# gz-oss
fuse oss;阿里云oss映射为自定义文件系统
依赖的Oss包，见fantasycool/oss项目

考虑使用spark在阿里机器上跑数据分析的需求可以考虑使用这个组件，在aliyun机器上部署hadoop运维成本很高，而且磁盘会丢数据哦！使用fuse
的方式映射文件系统可以使spark以读取本地文件的方式来使用oss进行数据分析。

组件在读取时已经做了read缓存，大大减少了请求次数，所以成本还是比较低廉的。endpoint采用指定内网的方式的时候，流量等都是不计费的。

目前基于这个组件进行spark读取数据分析的的经过项目验证了。
TODO:
1:文件信息获取
2:写文件
