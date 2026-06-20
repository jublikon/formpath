import { formatDuration, formatPeriod } from '../lib/formatters'
import { buildTrainingChart } from '../lib/trainingChart'
import type { TrainingDay } from '../lib/trainingOverview'

type TrainingVolumeChartProps = {
  days: TrainingDay[]
  startDate: string
  endDate: string
}

const chartWidth = 800
const chartHeight = 260

function formatChartDate(dateKey: string): string {
  const [year, month, day] = dateKey.split('-').map(Number)
  return new Intl.DateTimeFormat(undefined, {
    day: 'numeric',
    month: 'short',
  }).format(new Date(year, month - 1, day))
}

function formatAccessibleChartDate(dateKey: string): string {
  const [year, month, day] = dateKey.split('-').map(Number)
  return new Intl.DateTimeFormat(undefined, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(year, month - 1, day))
}

export function TrainingVolumeChart({
  days,
  startDate,
  endDate,
}: TrainingVolumeChartProps) {
  const values = days.map((day) => ({
    date: day.date,
    value: day.movingSeconds,
  }))
  const chart = buildTrainingChart(values, {
    width: chartWidth,
    height: chartHeight,
  })
  const middleIndex = Math.floor(days.length / 2)
  const labelPoints = [
    chart.points[0],
    chart.points[middleIndex],
    chart.points.at(-1),
  ]
  const description =
    chart.maxValue === 0
      ? 'No moving time was recorded during this period.'
      : `The highest daily moving time was ${formatDuration(chart.maxValue)}.`

  return (
    <section className="training-chart-card" aria-labelledby="chart-heading">
      <div className="section-heading">
        <div>
          <p className="section-eyebrow">Training volume</p>
          <h2 id="chart-heading">Daily moving time</h2>
        </div>
        <p className="period">{formatPeriod(startDate, endDate)}</p>
      </div>

      <svg
        className="training-chart"
        viewBox={`0 0 ${chartWidth} ${chartHeight}`}
        preserveAspectRatio="none"
        role="img"
        aria-labelledby="training-chart-title training-chart-description"
      >
        <title id="training-chart-title">
          Daily moving time for the last 28 days
        </title>
        <desc id="training-chart-description">{description}</desc>

        <defs>
          <linearGradient
            id="training-volume-gradient"
            x1="0"
            y1="0"
            x2="0"
            y2="1"
          >
            <stop offset="0%" stopColor="#397368" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#397368" stopOpacity="0.02" />
          </linearGradient>
        </defs>

        {[
          chart.bounds.top,
          (chart.bounds.top + chart.bounds.bottom) / 2,
          chart.bounds.bottom,
        ].map((y) => (
          <line
            className="chart-grid-line"
            key={y}
            x1={chart.bounds.left}
            x2={chart.bounds.right}
            y1={y}
            y2={y}
          />
        ))}

        <path
          className="chart-area"
          d={chart.areaPath}
          fill="url(#training-volume-gradient)"
        />
        <path className="chart-line" d={chart.linePath} />
      </svg>

      <div className="chart-labels" aria-hidden="true">
        {labelPoints.map((point, index) => (
          <span
            className={`chart-label chart-label-${index}`}
            key={point?.date ?? index}
            style={{
              left: point ? `${(point.x / chartWidth) * 100}%` : '0%',
            }}
          >
            {point ? formatChartDate(point.date) : ''}
          </span>
        ))}
      </div>

      <ul className="visually-hidden">
        {days.map((day) => (
          <li key={day.date}>
            {formatAccessibleChartDate(day.date)}:{' '}
            {formatDuration(day.movingSeconds)}
          </li>
        ))}
      </ul>
    </section>
  )
}
