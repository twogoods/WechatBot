const puppeteer = require('puppeteer');
var browser;
var page;
async function init() {
    browser = await puppeteer.launch({
        executablePath: '/Users/twogoods/Desktop/chrome-mac/Chromium.app/Contents/MacOS/Chromium',
        headless: false,
        slowMo: 150
    });
    page = await browser.newPage();
}
async function get() {
    await page.reload();
    await page.waitFor(1000);
    let content = await page.evaluate(() => {
        return document.querySelectorAll('article.post p')[0].innerText;
    });
    console.log(content);
}
async function run(){
    await init();
    await page.goto("https://www.nihaowua.com/");
    setInterval(function () {
        await get();
    }, 100);
}
run();