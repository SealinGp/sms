# SMS

树莓派与Air780E搭建的短信收发平台

```
GET:
  /random_key?range=(不提供则使用默认值)&length=(默认为8)
POST:
  /send_sms?key=(访问密钥,如果已通过网页登录则不需要)&sender=(发送者)&phone=(手机号)&message=(短信内容)
```

具体配置信息在config.ini中

## 如果想搭建, 请先看完此篇博客, 博客中有详细的说明和所需材料

[https://blog.akvicor.com/posts/project/sms](https://blog.akvicor.com/posts/project/sms/)
