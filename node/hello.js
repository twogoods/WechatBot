const puppeteer = require('puppeteer');
// https://cloud.tencent.com/community/article/529168
// http://cnodejs.org/topic/5a238b818eab6ee92a694622
async function run(){
    const browser = await puppeteer.launch({
      executablePath: '/Users/twogoods/Desktop/chrome-mac/Chromium.app/Contents/MacOS/Chromium',
      headless:false,
      slowMo: 150
    });
    const page = await browser.newPage();
    page.setViewport({
      width:1200,
      height:1200
    })

    await page.goto('https://segmentfault.com');
    await page.waitFor(1000)
    await page.click('.SFLogin')
    await page.type("input[name=username]","1271314078@qq.com");
    await page.type("input[name=password]","*********")
    await page.click('button.btn.btn-primary.pull-right.pl20.pr20')
    await page.waitFor(1000)
    await page.click('#shouldGoTo.btn.btn-primary')
    await page.waitFor(1000)
    await page.type("input[name=title]","大波是傻逼吗？");
    await page.focus("input.sf-typeHelper-input");
    await page.click('a[data-tag*="1040000000089449"]');
    await page.click("textarea#myEditor");
    await page.type('textarea#myEditor',"大波这个傻逼说程序员不好玩...我来教育他什么叫好玩.....");
    //browser.close();
}
run();