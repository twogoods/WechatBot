const puppeteer = require('puppeteer');
const axios = require('axios');
const fs = require('fs');
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
        height: 1000
    })
    await weChatPage.waitFor(500)
    await weChatPage.screenshot({ path: 'qrcode.png' });
    await weChatPage.waitFor(10000)

    wuPage = await browser.newPage();-
    await wuPage.goto("https://www.nihaowua.com/");
}


async function positionChat(name) {
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
}

async function sendMessage(name, msg) {
    await positionChat(name);
    await weChatPage.type("pre#editArea", msg);
    await weChatPage.waitFor(1000);
    await weChatPage.click("a.btn.btn_send");
}

async function sendFile(name, filePath) {
    await positionChat(name);
    await weChatPage.waitFor(2000);
    var elementHandle = await weChatPage.$('input.webuploader-element-invisible');
    elementHandle.uploadFile(filePath)
    await weChatPage.waitFor(2000)
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
const name = "是时候跨年"
async function run() {
    await initWeChat();
    var flag = 0;
    var path = "/Users/twogoods/code/wechatbot/node/meizi/"
    while (true) {
        var index = flag % 10
        await sendFile(name, path + index + ".jpg");
        let msg = await wuDuanzi();
        await sendMessage(name, msg);
        flag++;
        await timeout(10 * 60 * 1000);
    }
}

async function timeout(delay) {
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