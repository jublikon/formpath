const requiredMajor = 24
const currentVersion = process.versions.node
const currentMajor = Number.parseInt(currentVersion.split('.')[0], 10)

if (currentMajor !== requiredMajor) {
  console.error(
    [
      `Formpath web requires Node.js ${requiredMajor}.x.`,
      `Current Node.js version: ${currentVersion}.`,
      'Run `nvm use` before `npm ci` or `npm run dev`.',
    ].join('\n'),
  )
  process.exit(1)
}
