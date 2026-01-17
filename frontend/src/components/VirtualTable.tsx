import React, { useRef, useMemo, useId } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { theme, Spin, Empty } from "antd";
import { useVirtualTableStore, SortConfig } from "../stores/virtualTableStore";

interface ColumnType {
    title: React.ReactNode;
    dataIndex?: string;
    key?: string;
    width?: number | string;
    render?: (text: any, record: any, index: number) => React.ReactNode;
    align?: "left" | "center" | "right";
    sorter?: (a: any, b: any) => number;
}

interface VirtualTableProps {
    columns: ColumnType[];
    dataSource: any[];
    rowKey: string | ((record: any) => string);
    scroll?: { y?: number | string };
    loading?: boolean;
    onRow?: (record: any, index: number) => React.HTMLAttributes<HTMLElement>;
    rowHeight?: number;
    className?: string;
    style?: React.CSSProperties;
}

const VirtualTable: React.FC<VirtualTableProps> = ({
    columns,
    dataSource,
    rowKey,
    scroll,
    loading = false,
    onRow,
    rowHeight = 54, // Default estimate
    className,
    style,
}) => {
    const { token } = theme.useToken();
    const parentRef = useRef<HTMLDivElement>(null);
    const tableId = useId();

    const { sortConfigs, setSortConfig: setStoreSortConfig } = useVirtualTableStore();
    const sortConfig = sortConfigs[tableId] || null;
    const setSortConfig = (config: SortConfig | null) => setStoreSortConfig(tableId, config);

    const sortedData = useMemo(() => {
        if (!sortConfig) return dataSource;
        const sorted = [...dataSource].sort(sortConfig.sorter);
        return sortConfig.order === "descend" ? sorted.reverse() : sorted;
    }, [dataSource, sortConfig]);

    const rowVirtualizer = useVirtualizer({
        count: sortedData.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => rowHeight,
        overscan: 25,
    });

    const getRowKey = (record: any, index: number) => {
        if (typeof rowKey === "function") return rowKey(record);
        return record[rowKey as string] || index;
    };

    const handleHeaderClick = (col: ColumnType) => {
        if (!col.sorter) return;
        const key = col.key || col.dataIndex;
        if (!key) return;

        let newConfig: SortConfig | null;
        if (sortConfig && sortConfig.key === key) {
            if (sortConfig.order === "ascend") {
                newConfig = { key, order: "descend", sorter: col.sorter! };
            } else {
                newConfig = null; // Cancel sort
            }
        } else {
            newConfig = { key, order: "ascend", sorter: col.sorter! };
        }
        setSortConfig(newConfig);
    };

    // Sort logic could he handled here if necessary, but assuming dataSource is already sorted or controlled by parent
    // However, Antd Table handles sorting internally if `sorter` is present and state is managed.
    // For a pure replacement of "data dump", we assume parent handles sorting or we ignore it for now to solve rendering.
    // If headers need to be clickable for sort, that's more complex. We'll start with basic display.

    const renderCell = (col: ColumnType, record: any, index: number) => {
        let content = null;
        if (col.render) {
            // Antd render: (text, record, index)
            const text = col.dataIndex ? record[col.dataIndex] : undefined;
            content = col.render(text, record, index);
        } else if (col.dataIndex) {
            content = record[col.dataIndex];
        }

        return (
            <div
                key={col.key || col.dataIndex}
                style={{
                    width: col.width,
                    flex: col.width ? "none" : 1,
                    padding: "8px 16px",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: col.align === "right" ? "flex-end" : col.align === "center" ? "center" : "flex-start",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    // whiteSpace: "nowrap", // Optional
                }}
            >
                {content}
            </div>
        );
    };

    const header = (
        <div
            style={{
                display: "flex",
                backgroundColor: token.colorBgContainer, // or token.colorFillAlter
                borderBottom: `1px solid ${token.colorBorderSecondary}`,
                fontWeight: "bold",
                position: "sticky",
                top: 0,
                zIndex: 10,
                color: token.colorTextSecondary,
            }}
        >
            {columns.map((col, idx) => {
                const key = col.key || col.dataIndex;
                const isSorted = sortConfig && sortConfig.key === key;
                return (
                    <div
                        key={key || idx}
                        onClick={() => handleHeaderClick(col)}
                        style={{
                            width: col.width,
                            flex: col.width ? "none" : 1,
                            padding: "12px 16px",
                            textAlign: col.align || "left",
                            cursor: col.sorter ? "pointer" : "default",
                            userSelect: "none",
                            display: "flex",
                            alignItems: "center",
                            gap: 8,
                            backgroundColor: isSorted ? token.controlItemBgActive : "transparent",
                        }}
                    >
                        {col.title}
                        {isSorted && (
                            <span>{sortConfig.order === "ascend" ? "▲" : "▼"}</span>
                        )}
                    </div>
                );
            })}
        </div>
    );

    const scrollY = scroll?.y;
    const heightStyle = typeof scrollY === "number" ? `${scrollY}px` : scrollY || "100%";

    return (
        <div
            className={className}
            style={{
                position: "relative",
                height: heightStyle,
                display: "flex",
                flexDirection: "column",
                backgroundColor: token.colorBgContainer,
                ...style,
            }}
        >
            {/* Header outside or inside? Antd has header separate from scroll body usually. 
          But for simple virtual list, we put header above the scroll container. */}
            {header}

            <div
                ref={parentRef}
                style={{
                    flex: 1,
                    overflowY: "auto",
                    contain: "strict",
                }}
            >
                {loading ? (
                    <div style={{ padding: 40, textAlign: "center" }}>
                        <Spin />
                    </div>
                ) : sortedData.length === 0 ? (
                    <div style={{ padding: 40 }}>
                        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
                    </div>
                ) : (
                    <div
                        style={{
                            height: `${rowVirtualizer.getTotalSize()}px`,
                            width: "100%",
                            position: "relative",
                        }}
                    >
                        {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                            const sortedRecord = sortedData[virtualRow.index];
                            const key = getRowKey(sortedRecord, virtualRow.index);
                            const rowProps = onRow ? onRow(sortedRecord, virtualRow.index) : {};

                            return (
                                <div
                                    key={key}
                                    {...rowProps}
                                    style={{
                                        position: "absolute",
                                        top: 0,
                                        left: 0,
                                        width: "100%",
                                        height: `${virtualRow.size}px`,
                                        transform: `translateY(${virtualRow.start}px)`,
                                        display: "flex",
                                        borderBottom: `1px solid ${token.colorSplit}`,
                                        transition: "background-color 0.2s",
                                        ...rowProps.style,
                                    }}
                                    className="virtual-table-row"
                                    onMouseEnter={(e) => {
                                        e.currentTarget.style.backgroundColor = token.controlItemBgHover;
                                        rowProps.onMouseEnter?.(e);
                                    }}
                                    onMouseLeave={(e) => {
                                        e.currentTarget.style.backgroundColor = "transparent";
                                        rowProps.onMouseLeave?.(e);
                                    }}
                                >
                                    {columns.map((col, cIdx) => renderCell(col, sortedRecord, virtualRow.index))}
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
};

export default VirtualTable;
