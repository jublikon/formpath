type IntegrationPanelProps = {
  connected: boolean
  syncing: boolean
  hasActivities: boolean
  onSync: () => void
}

export function IntegrationPanel({
  connected,
  syncing,
  hasActivities,
  onSync,
}: IntegrationPanelProps) {
  let description = 'Connect Strava to import your activities.'

  if (connected && hasActivities) {
    description =
      'Connected. Sync new activities or change the connected account.'
  } else if (connected) {
    description =
      'Connected. Sync your activities or change the connected account.'
  } else if (hasActivities) {
    description = 'Reconnect Strava to sync new activities.'
  }

  return (
    <section
      className="integration-bar"
      aria-labelledby="integration-heading"
      aria-busy={syncing}
    >
      <div>
        <p className="integration-label">Data source</p>
        <h2 id="integration-heading">Strava</h2>
        <p className="integration-copy">{description}</p>
      </div>

      {connected ? (
        <div
          className="integration-actions"
          role="group"
          aria-label="Strava actions"
        >
          <a className="primary-action" href="/auth/strava">
            Change Account
          </a>
          <button
            className="secondary-action"
            type="button"
            disabled={syncing}
            onClick={onSync}
          >
            {syncing ? 'Syncing...' : 'Sync activities'}
          </button>
        </div>
      ) : (
        <a className="primary-action" href="/auth/strava">
          {hasActivities ? 'Reconnect Strava' : 'Connect with Strava'}
        </a>
      )}
    </section>
  )
}
