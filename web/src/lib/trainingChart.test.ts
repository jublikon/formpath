import { describe, expect, it } from 'vitest'
import type { TrainingDay } from './trainingOverview'
import { buildTrainingChart } from './trainingChart'

function trainingDays(values: number[]): TrainingDay[] {
  return values.map((movingSeconds, index) => ({
    date: `2026-06-${String(index + 1).padStart(2, '0')}`,
    movingSeconds,
  }))
}

describe('buildTrainingChart', () => {
  it('creates line and area paths from daily training values', () => {
    const chart = buildTrainingChart(trainingDays([0, 1800, 3600, 900]), {
      width: 800,
      height: 260,
    })

    expect(chart.linePath).toMatch(/^M/)
    expect(chart.areaPath).toMatch(/^M/)
    expect(chart.linePath).toContain('C')
    expect(chart.areaPath).toContain('Z')
  })

  it('places higher training values higher in the chart', () => {
    const chart = buildTrainingChart(trainingDays([0, 1800, 3600]), {
      width: 800,
      height: 260,
    })

    expect(chart.points[2].y).toBeLessThan(chart.points[1].y)
    expect(chart.points[1].y).toBeLessThan(chart.points[0].y)
  })

  it('keeps an all-zero series on the baseline', () => {
    const chart = buildTrainingChart(trainingDays([0, 0, 0]), {
      width: 800,
      height: 260,
    })

    expect(new Set(chart.points.map((point) => point.y))).toEqual(
      new Set([chart.bounds.bottom]),
    )
    expect(chart.maxMovingSeconds).toBe(0)
  })
})
