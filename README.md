# golang-smtp-gomail-ssl-send-email-demo
golang 发送邮件demo smtp包 和使用gomail包的方式 ssl

目的为实现抓取京东页面的商品信息， 监测商品预约数量超过阈值时发送邮件。

使用smtp包简单实现 本地测试ok， 扔到阿里云服务器上不行

原因 阿里云服务器 25端口被禁了

smtp连接邮箱只能走非25端口  ssl  465 （163邮箱）

ssl方式遇到一些问题。 需要去了解一些smtp的内容
