import { mkdir } from 'node:fs/promises'
import path from 'node:path'

import { chromium } from 'playwright'

const url = process.env.NEBULA_DEMO_URL || 'http://127.0.0.1:3100'
const outputDir = path.resolve(process.cwd(), '../docs/media/frames')

async function main() {
  await mkdir(outputDir, { recursive: true })

  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: 1500, height: 960 } })

  await page.goto(url, { waitUntil: 'networkidle', timeout: 60_000 })
  await page.waitForSelector('h1')

  for (let i = 0; i < 10; i += 1) {
    await page.waitForTimeout(1000)
    const framePath = path.join(outputDir, `frame-${String(i).padStart(2, '0')}.png`)
    await page.screenshot({ path: framePath, fullPage: false })
  }

  await browser.close()
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
