import React, { useRef, useEffect, useCallback, useImperativeHandle, forwardRef } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { theme, Spin, Empty } from "antd";

export interface VirtualListProps<T> {
  /** Data source array */
  dataSource: T[];
  /** Unique key for each item */
  rowKey: string | ((item: T, index: number) => string);
  /** Render function for each item */
  renderItem: (item: T, index: number, isSelected: boolean) => React.ReactNode;

  /** Estimated row height in pixels (default: 36) */
  rowHeight?: number;
  /** Container height (default: '100%') */
  height?: number | string;
  /** Number of items to render outside visible area (default: 10) */
  overscan?: number;

  /** Show loading spinner */
  loading?: boolean;
  /** Custom empty state text */
  emptyText?: string;
  /** Custom empty state icon */
  emptyIcon?: React.ReactNode;

  /** Callback when item is clicked */
  onItemClick?: (item: T, index: number) => void;
  /** Key of the selected item */
  selectedKey?: string | null;

  /** Enable auto-scroll to bottom when new items are added */
  autoScroll?: boolean;
  /** Callback when auto-scroll state changes */
  onAutoScrollChange?: (enabled: boolean) => void;

  /** Enable keyboard navigation with arrow keys (default: true when onItemClick is provided) */
  enableKeyboardNavigation?: boolean;

  /** Container className */
  className?: string;
  /** Container style */
  style?: React.CSSProperties;
  /** Show border between rows (default: true) */
  showBorder?: boolean;
  /** Custom row style */
  rowStyle?: React.CSSProperties;
  /** Custom row className */
  rowClassName?: string | ((item: T, index: number) => string);
}

export interface VirtualListHandle {
  /** Scroll to a specific index */
  scrollToIndex: (index: number, options?: { align?: 'start' | 'center' | 'end'; behavior?: 'auto' | 'smooth' }) => void;
  /** Scroll to the top of the list */
  scrollToTop: () => void;
  /** Scroll to the bottom of the list */
  scrollToBottom: () => void;
  /** Get the scroll container element */
  getScrollElement: () => HTMLDivElement | null;
  /** Focus the list container for keyboard navigation */
  focus: () => void;
}

function VirtualListInner<T>(
  props: VirtualListProps<T>,
  ref: React.ForwardedRef<VirtualListHandle>
) {
  const {
    dataSource,
    rowKey,
    renderItem,
    rowHeight = 36,
    height = '100%',
    overscan = 10,
    loading = false,
    emptyText,
    emptyIcon,
    onItemClick,
    selectedKey,
    autoScroll = false,
    onAutoScrollChange,
    enableKeyboardNavigation,
    className,
    style,
    showBorder = true,
    rowStyle,
    rowClassName,
  } = props;

  // Enable keyboard navigation by default when onItemClick is provided
  const keyboardNavEnabled = enableKeyboardNavigation ?? !!onItemClick;

  const { token } = theme.useToken();
  const parentRef = useRef<HTMLDivElement>(null);
  const scrollingRef = useRef(false);

  const virtualizer = useVirtualizer({
    count: dataSource.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => rowHeight,
    overscan,
  });

  const getRowKey = useCallback((item: T, index: number): string => {
    if (typeof rowKey === 'function') {
      return rowKey(item, index);
    }
    return String((item as Record<string, unknown>)[rowKey] ?? index);
  }, [rowKey]);

  const getRowClassName = useCallback((item: T, index: number): string => {
    if (typeof rowClassName === 'function') {
      return rowClassName(item, index);
    }
    return rowClassName || '';
  }, [rowClassName]);

  // Find the currently selected index based on selectedKey
  const getSelectedIndex = useCallback((): number => {
    if (selectedKey === undefined || selectedKey === null) return -1;
    return dataSource.findIndex((item, index) => getRowKey(item, index) === selectedKey);
  }, [dataSource, selectedKey, getRowKey]);

  // Handle keyboard navigation
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLDivElement>) => {
    if (!keyboardNavEnabled || !onItemClick || dataSource.length === 0) return;

    const currentIndex = getSelectedIndex();

    let nextIndex: number | null = null;

    switch (e.key) {
      case 'ArrowUp':
        e.preventDefault();
        if (currentIndex === -1) {
          // No selection, select last item
          nextIndex = dataSource.length - 1;
        } else if (currentIndex > 0) {
          nextIndex = currentIndex - 1;
        }
        break;
      case 'ArrowDown':
        e.preventDefault();
        if (currentIndex === -1) {
          // No selection, select first item
          nextIndex = 0;
        } else if (currentIndex < dataSource.length - 1) {
          nextIndex = currentIndex + 1;
        }
        break;
      case 'Home':
        e.preventDefault();
        nextIndex = 0;
        break;
      case 'End':
        e.preventDefault();
        nextIndex = dataSource.length - 1;
        break;
      case 'Enter':
      case ' ':
        // Confirm selection - already selected, just prevent default for space
        if (e.key === ' ') e.preventDefault();
        return;
      default:
        return;
    }

    if (nextIndex !== null && nextIndex !== currentIndex) {
      const nextItem = dataSource[nextIndex];
      onItemClick(nextItem, nextIndex);
      // Scroll to the new selected item
      scrollingRef.current = true;
      virtualizer.scrollToIndex(nextIndex, { align: 'auto' });
      setTimeout(() => {
        scrollingRef.current = false;
      }, 100);
    }
  }, [keyboardNavEnabled, onItemClick, dataSource, getSelectedIndex, virtualizer]);

  // Auto-scroll to bottom when new items are added
  useEffect(() => {
    if (autoScroll && dataSource.length > 0) {
      scrollingRef.current = true;
      virtualizer.scrollToIndex(dataSource.length - 1, { align: 'end' });
      const timer = setTimeout(() => {
        scrollingRef.current = false;
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [dataSource.length, autoScroll, virtualizer]);

  // Handle scroll events to detect user scroll
  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    if (scrollingRef.current) return;

    const target = e.currentTarget;
    const { scrollTop, scrollHeight, clientHeight } = target;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;

    if (onAutoScrollChange) {
      if (!isAtBottom && autoScroll) {
        onAutoScrollChange(false);
      } else if (isAtBottom && !autoScroll) {
        onAutoScrollChange(true);
      }
    }
  }, [autoScroll, onAutoScrollChange]);

  // Expose methods via ref
  useImperativeHandle(ref, () => ({
    scrollToIndex: (index: number, options?: { align?: 'start' | 'center' | 'end'; behavior?: 'auto' | 'smooth' }) => {
      scrollingRef.current = true;
      virtualizer.scrollToIndex(index, {
        align: options?.align || 'center',
        behavior: options?.behavior || 'auto',
      });
      setTimeout(() => {
        scrollingRef.current = false;
      }, options?.behavior === 'smooth' ? 500 : 100);
    },
    scrollToTop: () => {
      scrollingRef.current = true;
      parentRef.current?.scrollTo({ top: 0, behavior: 'smooth' });
      setTimeout(() => {
        scrollingRef.current = false;
      }, 500);
    },
    scrollToBottom: () => {
      scrollingRef.current = true;
      virtualizer.scrollToIndex(dataSource.length - 1, { align: 'end', behavior: 'smooth' });
      setTimeout(() => {
        scrollingRef.current = false;
      }, 500);
    },
    getScrollElement: () => parentRef.current,
    focus: () => {
      parentRef.current?.focus();
    },
  }), [virtualizer, dataSource.length]);

  const heightStyle = typeof height === 'number' ? `${height}px` : height;

  // Loading state
  if (loading && dataSource.length === 0) {
    return (
      <div
        className={className}
        style={{
          height: heightStyle,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: token.colorBgContainer,
          ...style,
        }}
      >
        <Spin />
      </div>
    );
  }

  // Empty state
  if (dataSource.length === 0) {
    return (
      <div
        className={className}
        style={{
          height: heightStyle,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: token.colorBgContainer,
          ...style,
        }}
      >
        <Empty
          image={emptyIcon || Empty.PRESENTED_IMAGE_SIMPLE}
          description={emptyText}
        />
      </div>
    );
  }

  return (
    <div
      ref={parentRef}
      className={className}
      tabIndex={keyboardNavEnabled ? 0 : undefined}
      onScroll={handleScroll}
      onKeyDown={handleKeyDown}
      style={{
        height: heightStyle,
        overflow: 'auto',
        contain: 'strict',
        backgroundColor: token.colorBgContainer,
        outline: 'none',
        ...style,
      }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualRow) => {
          const item = dataSource[virtualRow.index];
          const key = getRowKey(item, virtualRow.index);
          const isSelected = selectedKey !== undefined && selectedKey !== null && key === selectedKey;
          const extraClassName = getRowClassName(item, virtualRow.index);

          return (
            <div
              key={key}
              data-index={virtualRow.index}
              className={`virtual-list-row ${extraClassName}`.trim()}
              onClick={() => onItemClick?.(item, virtualRow.index)}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: `${virtualRow.size}px`,
                transform: `translateY(${virtualRow.start}px)`,
                borderBottom: showBorder ? `1px solid ${token.colorBorderSecondary}` : undefined,
                cursor: onItemClick ? 'pointer' : undefined,
                backgroundColor: isSelected ? token.colorPrimaryBg : undefined,
                transition: 'background-color 0.15s',
                boxSizing: 'border-box',
                ...rowStyle,
              }}
            >
              {renderItem(item, virtualRow.index, isSelected)}
            </div>
          );
        })}
      </div>

      {/* Loading overlay for incremental loading */}
      {loading && dataSource.length > 0 && (
        <div
          style={{
            position: 'absolute',
            bottom: 8,
            left: '50%',
            transform: 'translateX(-50%)',
            padding: '4px 12px',
            backgroundColor: token.colorBgElevated,
            borderRadius: token.borderRadius,
            boxShadow: token.boxShadow,
          }}
        >
          <Spin size="small" />
        </div>
      )}
    </div>
  );
}

// Export with generic type support
const VirtualList = forwardRef(VirtualListInner) as <T>(
  props: VirtualListProps<T> & { ref?: React.ForwardedRef<VirtualListHandle> }
) => React.ReactElement;

export default VirtualList;
