const puppeteer = require('puppeteer');
var browser;
var weChatPage;
var wuPage;
async function initWeChat() {
    browser = await puppeteer.launch({
        executablePath: '/Users/twogoods/Desktop/chrome-mac/Chromium.app/Contents/MacOS/Chromium',
        headless: false,
        slowMo: 150
    });
    weChatPage = await browser.newPage();
    await weChatPage.goto('https://wx.qq.com/');
    weChatPage.setViewport({
        width: 1200,
        height: 800
    })
    await weChatPage.waitFor(500)
    await weChatPage.screenshot({ path: 'qrcode.png' });
    await weChatPage.waitFor(10000)

   

    wuPage = await browser.newPage();
    await wuPage.goto("https://www.nihaowua.com/");
}

async function sendMessage(name,msg) {
    // await weChatPage.evaluate(() => {
    //     document.querySelectorAll('input.frm_search')[0].innerText="";
    // });
    await weChatPage.click("input.frm_search");
    await weChatPage.waitFor(500)
    await weChatPage.type("input.frm_search", name);
    await weChatPage.waitFor(1000)
    await weChatPage.click("div.contact_item.on div.info h4");
    await weChatPage.waitFor(1000)
    await weChatPage.focus("pre#editArea");
    await weChatPage.waitFor(100)
    await weChatPage.click("pre#editArea");
    await weChatPage.waitFor(100)
    await weChatPage.type("pre#editArea", msg);
    await weChatPage.waitFor(1000);
    await weChatPage.click("a.btn.btn_send");
}

async function wuDuanzi() {
    await wuPage.reload();
    await wuPage.waitFor(2000);
    let content = await wuPage.evaluate(() => {
        return document.querySelectorAll('article.post p')[0].innerText;
    });
    console.log(content);
    return content;
}

async function run() {
    await initWeChat();
    while (true) {
        timeout(10000);
        let msg = await wuDuanzi();
        await sendMessage("小目标要有",msg);
    }
}

function timeout(delay) {
    return new Promise((resolve, reject) => {
        setTimeout(() => {
            try {
                resolve(1)
            } catch (e) {
                reject(0)
            }
        }, delay);
    })
}
run();