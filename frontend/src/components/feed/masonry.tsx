import React, { ReactNode, useEffect, useState } from 'react'

type BreakpointCols = number | { default: number;[key: number]: number }

interface MasonryProps {
    breakpointCols?: BreakpointCols
    className?: string
    columnClassName?: string
    columnAttrs?: React.HTMLAttributes<HTMLDivElement>
    children: ReactNode
}

const DEFAULT_COLUMNS = 2

export const Masonry: React.FC<MasonryProps> = ({
    breakpointCols = DEFAULT_COLUMNS,
    className = 'my-masonry-grid',
    columnClassName = 'my-masonry-grid_column',
    columnAttrs = {},
    children,
}) => {
    const [columnCount, setColumnCount] = useState<number>(
        typeof breakpointCols === 'object' && breakpointCols.default
            ? breakpointCols.default
            : typeof breakpointCols === 'number'
                ? breakpointCols
                : DEFAULT_COLUMNS
    )

    useEffect(() => {
        const calculateColumnCount = () => {
            const windowWidth = window.innerWidth || Infinity

            const breakpointColsObject =
                typeof breakpointCols === 'object'
                    ? breakpointCols
                    : { default: breakpointCols }

            let matchedBreakpoint = Infinity
            let columns = breakpointColsObject.default || DEFAULT_COLUMNS

            for (const breakpoint in breakpointColsObject) {
                const optBreakpoint = parseInt(breakpoint, 10)
                const isCurrentBreakpoint =
                    optBreakpoint > 0 && windowWidth <= optBreakpoint

                if (isCurrentBreakpoint && optBreakpoint < matchedBreakpoint) {
                    matchedBreakpoint = optBreakpoint
                    columns = breakpointColsObject[breakpoint]
                }
            }

            columns = Math.max(1, columns)

            if (columns !== columnCount) {
                setColumnCount(columns)
            }
        }

        calculateColumnCount()
        window.addEventListener('resize', calculateColumnCount)

        return () => {
            window.removeEventListener('resize', calculateColumnCount)
        }
    }, [breakpointCols, columnCount])

    const itemsInColumns = (): ReactNode[][] => {
        const items = React.Children.toArray(children)
        const columns: ReactNode[][] = Array.from({ length: columnCount }, () => [])

        items.forEach((item, index) => {
            const columnIndex = index % columnCount
            columns[columnIndex].push(item)
        })

        return columns
    }

    const renderColumns = () => {
        const childrenInColumns = itemsInColumns()
        const columnWidth = `${100 / childrenInColumns.length}%`

        return childrenInColumns.map((items, i) => (
            <div
                key={i}
                {...columnAttrs}
                className={columnClassName}
                style={{ ...columnAttrs.style, width: columnWidth }}
            >
                {items}
            </div>
        ))
    }

    return <div className={className}>{renderColumns()}</div>
}