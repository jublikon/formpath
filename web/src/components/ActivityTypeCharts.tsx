import {
  formatDuration,
  formatKilometers,
  formatPeriod,
} from '../lib/formatters'
import {
  buildTrainingChart,
  type ChartDay,
} from '../lib/trainingChart'
import type { TrainingDay } from '../lib/trainingOverview'

type ActivityTypeChartsProps = {
  days: TrainingDay[]
  startDate: string
  endDate: string
}

type ActivityChartConfig = {
  id: string
  eyebrow: string
  title: string
  color: string
  values: ChartDay[]
  formatValue: (value: number) => string
  emptyDescription: string
  maximumDescription: (value: number) => string
}

const chartWidth = 480
const chartHeight = 220

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

function ActivityChart({ config }: { config: ActivityChartConfig }) {
  const chart = buildTrainingChart(config.values, {
    width: chartWidth,
    height: chartHeight,
  })
  const middleIndex = Math.floor(config.values.length / 2)
  const dateLabelPoints = [
    chart.points[0],
    chart.points[middleIndex],
    chart.points.at(-1),
  ]
  const yAxisValues = [chart.maxValue, chart.maxValue / 2, 0]
  const yAxisPositions = [
    chart.bounds.top,
    (chart.bounds.top + chart.bounds.bottom) / 2,
    chart.bounds.bottom,
  ]
  const description =
    chart.maxValue === 0
      ? config.emptyDescription
      : config.maximumDescription(chart.maxValue)
  const titleId = `${config.id}-title`
  const descriptionId = `${config.id}-description`

  return (
    <article className="activity-chart-card">
      <div className="activity-chart-heading">
        <p className="section-eyebrow">{config.eyebrow}</p>
        <h3>{config.title}</h3>
      </div>

      <div className="chart-canvas">
        <svg
          className="training-chart"
          viewBox={`0 0 ${chartWidth} ${chartHeight}`}
          preserveAspectRatio="none"
          role="img"
          aria-labelledby={`${titleId} ${descriptionId}`}
        >
          <title id={titleId}>{config.title} for the last 28 days</title>
          <desc id={descriptionId}>{description}</desc>

          <defs>
            <linearGradient
              id={`${config.id}-gradient`}
              x1="0"
              y1="0"
              x2="0"
              y2="1"
            >
              <stop offset="0%" stopColor={config.color} stopOpacity="0.28" />
              <stop offset="100%" stopColor={config.color} stopOpacity="0.02" />
            </linearGradient>
          </defs>

          {yAxisPositions.map((y, index) => (
            <g key={y}>
              <line
                className="chart-grid-line"
                x1={chart.bounds.left}
                x2={chart.bounds.right}
                y1={y}
                y2={y}
              />
              <text
                className="chart-y-label"
                x={chart.bounds.left - 12}
                y={y}
              >
                {config.formatValue(yAxisValues[index])}
              </text>
            </g>
          ))}

          <path
            className="chart-area"
            d={chart.areaPath}
            fill={`url(#${config.id}-gradient)`}
          />
          <path
            className="chart-line"
            d={chart.linePath}
            style={{ stroke: config.color }}
          />
        </svg>

        <div className="chart-labels" aria-hidden="true">
          {dateLabelPoints.map((point, index) => (
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
      </div>

      <p className="chart-summary">{description}</p>

      <ul className="visually-hidden">
        {config.values.map((day) => (
          <li key={day.date}>
            {formatAccessibleChartDate(day.date)}: {config.formatValue(day.value)}
          </li>
        ))}
      </ul>
    </article>
  )
}

export function ActivityTypeCharts({
  days,
  startDate,
  endDate,
}: ActivityTypeChartsProps) {
  const charts: ActivityChartConfig[] = [
    {
      id: 'running-distance',
      eyebrow: 'Running',
      title: 'Daily running distance',
      color: '#397368',
      values: days.map((day) => ({
        date: day.date,
        value: day.runDistanceMeters,
      })),
      formatValue: formatKilometers,
      emptyDescription: 'No running distance was recorded during this period.',
      maximumDescription: (value) =>
        `The highest daily running distance was ${formatKilometers(value)}.`,
    },
    {
      id: 'cycling-distance',
      eyebrow: 'Cycling',
      title: 'Daily cycling distance',
      color: '#b4663c',
      values: days.map((day) => ({
        date: day.date,
        value: day.rideDistanceMeters,
      })),
      formatValue: formatKilometers,
      emptyDescription: 'No cycling distance was recorded during this period.',
      maximumDescription: (value) =>
        `The highest daily cycling distance was ${formatKilometers(value)}.`,
    },
    {
      id: 'workout-time',
      eyebrow: 'Workout',
      title: 'Daily workout moving time',
      color: '#665aa7',
      values: days.map((day) => ({
        date: day.date,
        value: day.workoutMovingSeconds,
      })),
      formatValue: formatDuration,
      emptyDescription: 'No workout time was recorded during this period.',
      maximumDescription: (value) =>
        `The highest daily workout time was ${formatDuration(value)}.`,
    },
  ]

  return (
    <section
      className="activity-charts"
      aria-labelledby="activity-charts-heading"
    >
      <div className="section-heading">
        <div>
          <p className="section-eyebrow">Training by activity</p>
          <h2 id="activity-charts-heading">Your four-week rhythm</h2>
        </div>
        <p className="period">{formatPeriod(startDate, endDate)}</p>
      </div>

      <div className="activity-chart-grid">
        {charts.map((chart) => (
          <ActivityChart config={chart} key={chart.id} />
        ))}
      </div>
    </section>
  )
}
