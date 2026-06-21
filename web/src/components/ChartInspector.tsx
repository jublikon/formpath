import { useState, type PointerEvent } from 'react'
import {
  findNearestPointIndex,
  movePointIndex,
  pointerClientXToChartX,
} from '../lib/chartInteraction'
import type { TrainingChartPoint } from '../lib/trainingChart'

type ChartInspectorProps = {
  points: TrainingChartPoint[]
  width: number
  height: number
  color: string
  formatDate: (date: string) => string
  formatValue: (value: number) => string
  label: string
}

export function ChartInspector({
  points,
  width,
  height,
  color,
  formatDate,
  formatValue,
  label,
}: ChartInspectorProps) {
  const [activeIndex, setActiveIndex] = useState<number | null>(null)
  const activePoint =
    activeIndex === null ? undefined : points[activeIndex]

  function inspectPointer(event: PointerEvent<HTMLDivElement>) {
    const bounds = event.currentTarget.getBoundingClientRect()
    const chartX = pointerClientXToChartX(
      event.clientX,
      bounds.left,
      bounds.width,
      width,
    )
    const nearestIndex = findNearestPointIndex(points, chartX)

    if (nearestIndex >= 0) {
      setActiveIndex(nearestIndex)
    }
  }

  function moveSelection(direction: -1 | 1) {
    setActiveIndex((currentIndex) =>
      movePointIndex(currentIndex, direction, points.length),
    )
  }

  const pointXPercent = activePoint ? (activePoint.x / width) * 100 : 0
  const pointYPercent = activePoint ? (activePoint.y / height) * 100 : 0
  const horizontalPosition =
    pointXPercent < 20 ? 'start' : pointXPercent > 80 ? 'end' : 'center'
  const verticalPosition = pointYPercent < 32 ? 'below' : 'above'
  const activeDescription = activePoint
    ? `${formatDate(activePoint.date)}: ${formatValue(activePoint.value)}`
    : ''

  return (
    <div className="chart-inspector">
      <div
        className="chart-inspector-target"
        tabIndex={0}
        role="group"
        aria-label={`Inspect ${label}. Use the left and right arrow keys to inspect daily values.`}
        onFocus={() => {
          if (activeIndex === null && points.length > 0) {
            setActiveIndex(0)
          }
        }}
        onBlur={() => setActiveIndex(null)}
        onKeyDown={(event) => {
          if (event.key === 'ArrowLeft') {
            event.preventDefault()
            moveSelection(-1)
          }

          if (event.key === 'ArrowRight') {
            event.preventDefault()
            moveSelection(1)
          }

          if (event.key === 'Escape') {
            setActiveIndex(null)
          }
        }}
        onPointerDown={(event) => {
          inspectPointer(event)

          if (event.pointerType !== 'mouse') {
            event.currentTarget.setPointerCapture(event.pointerId)
          }
        }}
        onPointerMove={(event) => {
          if (
            event.pointerType === 'mouse' ||
            event.currentTarget.hasPointerCapture(event.pointerId)
          ) {
            inspectPointer(event)
          }
        }}
        onPointerUp={(event) => {
          if (event.currentTarget.hasPointerCapture(event.pointerId)) {
            event.currentTarget.releasePointerCapture(event.pointerId)
          }
        }}
        onPointerCancel={(event) => {
          if (event.currentTarget.hasPointerCapture(event.pointerId)) {
            event.currentTarget.releasePointerCapture(event.pointerId)
          }
          setActiveIndex(null)
        }}
        onPointerLeave={(event) => {
          if (event.pointerType === 'mouse') {
            setActiveIndex(null)
          }
        }}
      />

      {activePoint ? (
        <div aria-hidden="true">
          <span
            className="chart-inspector-line"
            style={{ left: `${pointXPercent}%` }}
          />
          <span
            className="chart-inspector-point"
            style={{
              backgroundColor: color,
              left: `${pointXPercent}%`,
              top: `${pointYPercent}%`,
            }}
          />
          <div
            className={`chart-tooltip chart-tooltip-${horizontalPosition} chart-tooltip-${verticalPosition}`}
            style={{
              left: `${pointXPercent}%`,
              top: `${pointYPercent}%`,
            }}
          >
            <span className="chart-tooltip-date">
              {formatDate(activePoint.date)}
            </span>
            <strong className="chart-tooltip-value">
              {formatValue(activePoint.value)}
            </strong>
          </div>
        </div>
      ) : null}

      <span className="visually-hidden" aria-live="polite">
        {activeDescription}
      </span>
    </div>
  )
}
