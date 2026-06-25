import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { IntegrationPanel } from './IntegrationPanel'

function renderPanel(options: {
  connected: boolean
  syncing?: boolean
  hasActivities?: boolean
}) {
  return renderToStaticMarkup(
    <IntegrationPanel
      connected={options.connected}
      syncing={options.syncing ?? false}
      hasActivities={options.hasActivities ?? false}
      onSync={() => undefined}
    />,
  )
}

describe('IntegrationPanel', () => {
  it('keeps account switching available after Strava is connected', () => {
    const html = renderPanel({ connected: true, hasActivities: true })

    expect(html).toContain('Change Account')
    expect(html).toContain('href="/auth/strava"')
    expect(html).toContain('Sync activities')
  })

  it('keeps the sync action disabled while syncing', () => {
    const html = renderPanel({
      connected: true,
      syncing: true,
      hasActivities: true,
    })

    expect(html).toContain('Change Account')
    expect(html).toContain('Syncing...')
    expect(html).toContain('disabled=""')
  })

  it('shows the initial Strava connect action before authentication', () => {
    const html = renderPanel({ connected: false })

    expect(html).toContain('Connect with Strava')
    expect(html).toContain('href="/auth/strava"')
    expect(html).not.toContain('Sync activities')
  })
})
