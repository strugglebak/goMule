# goMule
一个用 go 实现的 torrent 客户端，基于 go version 1.17.6

# Features
- [x] 支持[核心协议](https://www.bittorrent.org/beps/bep_0003.html)
- [x] 支持下载进度展示
- [x] 支持 `.torrent` 文件解析
- [x] 支持解析 `tracker` 响应
- [x] 支持 `p2p` 协议下载
- [x] 支持 `peers` 之间的并发下载
- [x] 最大支持 5 个 pipelined 下载请求

# 安装

```bash
go get github.com/strugglebak/goMule
```

# 用法和调试

可以尝试用 [这个页面](https://cdimage.debian.org/debian-cd/current/amd64/bt-cd/#indexlist) 下载 torrent，目前尝试用的例子是 `debian-11.2.0-amd64-netinst.iso.torrent`

```bash
git clone git@github.com:strugglebak/goMule.git
cd goMule
go build
./goMule debian-11.2.0-amd64-netinst.iso.torrent debian.iso
```

# 截图

![](./assets/screen_shot.png)

# Roadmaps

- [ ] 支持 [Fast extension](http://bittorrent.org/beps/bep_0006.html)
- [ ] 支持 [磁力链接](http://bittorrent.org/beps/bep_0009.html)
- [ ] 支持 [多 tracker](http://bittorrent.org/beps/bep_0012.html)
- [ ] 支持 [UDP tracker](http://bittorrent.org/beps/bep_0015.html)
- [ ] 支持 [DHT](http://bittorrent.org/beps/bep_0005.html)
- [ ] 支持 [PEX](http://bittorrent.org/beps/bep_0011.html)

...
