import { area, curveMonotoneX, line } from 'd3-shape'
type ChartSize = {
  width: number
  height: number
}

export type ChartDay = {
  date: string
  value: number
}

export type TrainingChartPoint = ChartDay & {
  x: number
  y: number
}

export type TrainingChart = {
  linePath: string
  areaPath: string
  points: TrainingChartPoint[]
  maxValue: number
  bounds: {
    top: number
    right: number
    bottom: number
    left: number
  }
}

const chartPadding = {
  top: 20,
  right: 20,
  bottom: 36,
  left: 94,
}

export function buildTrainingChart(
  days: ChartDay[],
  size: ChartSize,
): TrainingChart {
  const plotWidth = size.width - chartPadding.left - chartPadding.right
  const baselineY = size.height - chartPadding.bottom
  const plotHeight = baselineY - chartPadding.top
  const maxValue = Math.max(0, ...days.map((day) => day.value))
  const lastIndex = Math.max(days.length - 1, 1)

  const points = days.map((day, index) => {
    const x = chartPadding.left + (index / lastIndex) * plotWidth
    const y =
      maxValue === 0
        ? baselineY
        : chartPadding.top +
          (1 - day.value / maxValue) * plotHeight

    return {
      ...day,
      x,
      y,
    }
  })

  const linePath =
    line<TrainingChartPoint>()
      .x((point) => point.x)
      .y((point) => point.y)
      .curve(curveMonotoneX)(points) ?? ''

  const areaPath =
    area<TrainingChartPoint>()
      .x((point) => point.x)
      .y0(baselineY)
      .y1((point) => point.y)
      .curve(curveMonotoneX)(points) ?? ''

  return {
    linePath,
    areaPath,
    points,
    maxValue,
    bounds: {
      top: chartPadding.top,
      right: size.width - chartPadding.right,
      bottom: baselineY,
      left: chartPadding.left,
    },
  }
}
