# Media Assets

This directory stores README and release visuals.

- `architecture.svg`: current architecture snapshot used in the top-level README.
- `dashboard-placeholder.svg`: temporary panel for upcoming dashboard screenshots or GIFs.
- `demo.gif`: 60-second real dashboard capture exported from the running `web/` app.

Recommended follow-up:

1. Replace `dashboard-placeholder.svg` with a real screenshot from `web/`.
2. Refresh the demo GIF from live UI using the commands below.

## Regenerate Demo GIF

From repository root:

```bash
cd web
npm install
npx playwright install chromium
npm run dev -- --port 3100
```

In a second terminal:

```bash
cd web
npm run capture:demo
cd ..
python3 scripts/build_demo_gif.py
```
