import { area, curveMonotoneX, line } from 'd3-shape'
import type { TrainingDay } from './trainingOverview'

type ChartSize = {
  width: number
  height: number
}

export type TrainingChartPoint = TrainingDay & {
  x: number
  y: number
}

export type TrainingChart = {
  linePath: string
  areaPath: string
  points: TrainingChartPoint[]
  maxMovingSeconds: number
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
  left: 48,
}

export function buildTrainingChart(
  days: TrainingDay[],
  size: ChartSize,
): TrainingChart {
  const plotWidth = size.width - chartPadding.left - chartPadding.right
  const baselineY = size.height - chartPadding.bottom
  const plotHeight = baselineY - chartPadding.top
  const maxMovingSeconds = Math.max(
    0,
    ...days.map((day) => day.movingSeconds),
  )
  const lastIndex = Math.max(days.length - 1, 1)

  const points = days.map((day, index) => {
    const x = chartPadding.left + (index / lastIndex) * plotWidth
    const y =
      maxMovingSeconds === 0
        ? baselineY
        : chartPadding.top +
          (1 - day.movingSeconds / maxMovingSeconds) * plotHeight

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
    maxMovingSeconds,
    bounds: {
      top: chartPadding.top,
      right: size.width - chartPadding.right,
      bottom: baselineY,
      left: chartPadding.left,
    },
  }
}
