// node screenshot.js 'http://grafana:3000/d/9dMXL2N7z/cmd-application?orgId=1&from=now-24h&to=now&kiosk' screenshot.png

const puppeteer = require('puppeteer');
const args = process.argv.slice(2);

const url = args[0];
const fname = args[1];

(async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.setViewport({ width: 1280, height: 1500 })
  await page.goto(url, { waitUntil: 'networkidle2' });
  await page.screenshot({ path: fname, fullPage: true });

  await page.close()
  await browser.close();
})();
