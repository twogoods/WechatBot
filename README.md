# WeChatBot
这是一个微信机器人的探索项目。微信机器人这种东西微信官方肯定是不支持的，于是乎可能的方式就是hook,web api这样的方式。类似hook的方式可以使用Xposed框架，当然这个我没玩过，不懂；另一种就是基于微信web版。这个就是基于微信web版所做的两种不同的尝试。
## go
go目录下的是go实现的版本，它就是纯粹的使用微信web api，关于api，前人都已经为我们总结好了看[这个](https://github.com/Urinx/WeixinBot)，思路都是一样的，这里只是用golang实现了而已，其中的一个坑是使用的httpclient一定要开启cookie，逃....
## node
直接使用web api的方式会有极大的概率被微信封账号，至于封账号的逻辑就很难说了，直接调用api虽然可以实现，但终究是跟官方web页面发送的请求有所区别，又或者说短时间发送大量消息，24小时不断发送，这些都可能成为微信封号的判断维度。对于利用web版的微信，模拟一个普通用户的操作是能做的最大的努力了。于是页面自动化成了一个思路。怎么实现呢？[Puppeteer](https://github.com/GoogleChrome/puppeteer)！加上headless模式简直神器啊！api极简单极易上手，看官方helloworld就知道分分钟截个图

```
const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');
  await page.screenshot({path: 'example.png'});

  await browser.close();
})();
```
node目录下是一个使用Puppeteer实现的可以发消息和文件的demo。

