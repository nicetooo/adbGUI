import { memo, useMemo, useRef, useState, useEffect, useCallback } from 'react';
import JsonView from 'react18-json-view';
import 'react18-json-view/src/style.css';
import { useTheme } from '../ThemeContext';
import { Input, Space, Typography } from 'antd';
import { SearchOutlined, CloseOutlined, UpOutlined, DownOutlined } from '@ant-design/icons';

const { Text } = Typography;

// ============================================================
// JSON Search Hook - DOM-based text highlight + navigation
// ============================================================

function useJsonSearch(containerRef: React.RefObject<HTMLDivElement | null>, searchTerm: string) {
  const [matchCount, setMatchCount] = useState(0);
  const [currentIndex, setCurrentIndex] = useState(-1);
  const matchesRef = useRef<HTMLElement[]>([]);

  // Clear all <mark> elements from container
  const clearMarks = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;
    const marks = container.querySelectorAll('mark.jv-search-hit');
    marks.forEach(mark => {
      const parent = mark.parentNode;
      if (parent) {
        parent.replaceChild(document.createTextNode(mark.textContent || ''), mark);
        parent.normalize();
      }
    });
    matchesRef.current = [];
  }, [containerRef]);

  // Apply marks for current search term
  useEffect(() => {
    clearMarks();

    const container = containerRef.current;
    if (!container || !searchTerm.trim()) {
      setMatchCount(0);
      setCurrentIndex(-1);
      return;
    }

    const term = searchTerm.toLowerCase();
    const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT);
    const hits: { node: Text; positions: number[] }[] = [];

    let textNode: Text | null;
    while ((textNode = walker.nextNode() as Text | null)) {
      const text = textNode.textContent || '';
      const lower = text.toLowerCase();
      const positions: number[] = [];
      let idx = 0;
      while ((idx = lower.indexOf(term, idx)) !== -1) {
        positions.push(idx);
        idx += term.length;
      }
      if (positions.length > 0) {
        hits.push({ node: textNode, positions });
      }
    }

    const allMarks: HTMLElement[] = [];

    // Process in reverse order so earlier text node references stay valid
    for (let h = hits.length - 1; h >= 0; h--) {
      const { node, positions } = hits[h];
      const text = node.textContent || '';
      const fragment = document.createDocumentFragment();
      let lastIdx = 0;
      const localMarks: HTMLElement[] = [];

      for (const pos of positions) {
        if (pos > lastIdx) {
          fragment.appendChild(document.createTextNode(text.slice(lastIdx, pos)));
        }
        const mark = document.createElement('mark');
        mark.className = 'jv-search-hit';
        mark.textContent = text.slice(pos, pos + term.length);
        localMarks.push(mark);
        fragment.appendChild(mark);
        lastIdx = pos + term.length;
      }
      if (lastIdx < text.length) {
        fragment.appendChild(document.createTextNode(text.slice(lastIdx)));
      }
      node.parentNode?.replaceChild(fragment, node);
      // Prepend in reverse so final order is document order
      allMarks.unshift(...localMarks);
    }

    matchesRef.current = allMarks;
    setMatchCount(allMarks.length);
    setCurrentIndex(allMarks.length > 0 ? 0 : -1);
  }, [searchTerm, containerRef, clearMarks]);

  // Update active highlight when currentIndex changes
  useEffect(() => {
    const marks = matchesRef.current;
    marks.forEach(m => m.classList.remove('jv-search-active'));
    if (currentIndex >= 0 && currentIndex < marks.length) {
      marks[currentIndex].classList.add('jv-search-active');
      marks[currentIndex].scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    }
  }, [currentIndex]);

  const goNext = useCallback(() => {
    setCurrentIndex(prev => (matchesRef.current.length > 0 ? (prev + 1) % matchesRef.current.length : -1));
  }, []);

  const goPrev = useCallback(() => {
    setCurrentIndex(prev => (matchesRef.current.length > 0 ? (prev - 1 + matchesRef.current.length) % matchesRef.current.length : -1));
  }, []);

  const reset = useCallback(() => {
    clearMarks();
    setMatchCount(0);
    setCurrentIndex(-1);
  }, [clearMarks]);

  return { matchCount, currentIndex, goNext, goPrev, reset };
}

// ============================================================
// Search Bar
// ============================================================

interface SearchBarProps {
  matchCount: number;
  currentIndex: number;
  onSearch: (term: string) => void;
  onNext: () => void;
  onPrev: () => void;
  onClose: () => void;
  isDark: boolean;
}

const SearchBar = memo(({ matchCount, currentIndex, onSearch, onNext, onPrev, onClose, isDark }: SearchBarProps) => {
  const inputRef = useRef<ReturnType<typeof Input.Search> | null>(null);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    } else if (e.key === 'Enter') {
      if (e.shiftKey) onPrev();
      else onNext();
    }
  };

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      padding: '4px 8px',
      borderBottom: `1px solid ${isDark ? '#333' : '#e8e8e8'}`,
      background: isDark ? '#252526' : '#f5f5f5',
      borderRadius: '6px 6px 0 0',
    }}>
      <SearchOutlined style={{ color: isDark ? '#888' : '#999', fontSize: 12, flexShrink: 0 }} />
      <Input
        ref={inputRef as never}
        size="small"
        placeholder="Search..."
        allowClear
        autoFocus
        onChange={e => onSearch(e.target.value)}
        onKeyDown={handleKeyDown}
        style={{
          flex: 1,
          fontSize: 12,
          background: isDark ? '#1e1e1e' : '#fff',
          borderColor: isDark ? '#444' : '#d9d9d9',
        }}
      />
      {matchCount > 0 && (
        <Text style={{ fontSize: 11, whiteSpace: 'nowrap', color: isDark ? '#aaa' : '#666', flexShrink: 0 }}>
          {currentIndex + 1}/{matchCount}
        </Text>
      )}
      {matchCount === 0 && currentIndex === -1 && (
        <Text style={{ fontSize: 11, whiteSpace: 'nowrap', color: isDark ? '#666' : '#999', flexShrink: 0 }}>
          0/0
        </Text>
      )}
      <Space.Compact size="small" style={{ flexShrink: 0 }}>
        <span
          onClick={onPrev}
          style={{ cursor: 'pointer', padding: '2px 4px', color: isDark ? '#aaa' : '#666', display: 'inline-flex', alignItems: 'center' }}
        >
          <UpOutlined style={{ fontSize: 11 }} />
        </span>
        <span
          onClick={onNext}
          style={{ cursor: 'pointer', padding: '2px 4px', color: isDark ? '#aaa' : '#666', display: 'inline-flex', alignItems: 'center' }}
        >
          <DownOutlined style={{ fontSize: 11 }} />
        </span>
      </Space.Compact>
      <span
        onClick={onClose}
        style={{ cursor: 'pointer', padding: '2px 4px', color: isDark ? '#aaa' : '#666', display: 'inline-flex', alignItems: 'center', flexShrink: 0 }}
      >
        <CloseOutlined style={{ fontSize: 11 }} />
      </span>
    </div>
  );
});
SearchBar.displayName = 'SearchBar';

// ============================================================
// Highlight Styles (injected once)
// ============================================================

const STYLE_ID = 'json-viewer-search-styles';
function ensureStyles() {
  if (document.getElementById(STYLE_ID)) return;
  const style = document.createElement('style');
  style.id = STYLE_ID;
  style.textContent = `
    mark.jv-search-hit {
      background: #e8a90e;
      color: #000;
      border-radius: 2px;
      padding: 0 1px;
      scroll-margin: 40px;
    }
    mark.jv-search-hit.jv-search-active {
      background: #f57c00;
      color: #fff;
      outline: 2px solid #f57c00;
      outline-offset: 0px;
    }
  `;
  document.head.appendChild(style);
}

// ============================================================
// JsonViewer Component
// ============================================================

interface JsonViewerProps {
  /** JSON data - can be object, array, string (will attempt parse), or any value */
  data: unknown;
  /** Max height in px or CSS string (default: 'none' = fill container) */
  maxHeight?: number | string;
  /** Collapse depth - levels to expand by default (default: Infinity = all expanded) */
  collapseDepth?: number;
  /** Whether to show copy button on hover (default: true) */
  enableCopy?: boolean;
  /** Custom style overrides */
  style?: React.CSSProperties;
  /** Font size in px (default: 12) */
  fontSize?: number;
  /** Whether to show search bar (default: true for JSON objects, false for plain strings) */
  searchable?: boolean;
  /** External search term from parent component (e.g., global search) */
  externalSearchTerm?: string;
}

/**
 * Unified JSON viewer component wrapping react18-json-view.
 * Auto-adapts to dark/light theme.
 * Built-in search with Ctrl/Cmd+F, highlight, and prev/next navigation.
 */
const JsonViewer = memo(({
  data,
  maxHeight = 'none',
  collapseDepth = Infinity,
  enableCopy = true,
  style,
  fontSize = 12,
  searchable,
  externalSearchTerm,
}: JsonViewerProps) => {
  const { isDark } = useTheme();
  const containerRef = useRef<HTMLDivElement>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // Priority: internal search > external search
  // If internal search is active, use searchTerm; otherwise use externalSearchTerm
  const activeSearchTerm = searchOpen ? searchTerm : (externalSearchTerm || '');
  const { matchCount, currentIndex, goNext, goPrev, reset } = useJsonSearch(containerRef, activeSearchTerm);

  // Inject highlight styles once
  useEffect(() => ensureStyles(), []);

  // Keyboard shortcut: Ctrl/Cmd+F to open search
  useEffect(() => {
    const wrapper = wrapperRef.current;
    if (!wrapper) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'f') {
        e.preventDefault();
        e.stopPropagation();
        setSearchOpen(true);
      }
    };
    wrapper.addEventListener('keydown', handleKeyDown);
    return () => wrapper.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleCloseSearch = useCallback(() => {
    setSearchOpen(false);
    setSearchTerm('');
    reset();
  }, [reset]);

  // Parse string data to object if possible
  const parsedData = useMemo(() => {
    if (data === null || data === undefined) return null;
    if (typeof data === 'string') {
      try {
        return JSON.parse(data);
      } catch {
        return data;
      }
    }
    return data;
  }, [data]);

  const hasHeightConstraint = maxHeight !== 'none';
  const isJsonObject = parsedData !== null && typeof parsedData === 'object';
  const showSearch = searchable ?? isJsonObject;

  // Plain string fallback
  if (typeof parsedData === 'string') {
    return (
      <pre style={{
        padding: 12,
        borderRadius: 6,
        ...(hasHeightConstraint ? { overflow: 'auto', maxHeight: typeof maxHeight === 'number' ? maxHeight : maxHeight } : {}),
        fontSize,
        fontFamily: '"JetBrains Mono", "Fira Code", "SF Mono", Menlo, monospace',
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-all',
        margin: 0,
        background: isDark ? '#1e1e1e' : '#fafafa',
        color: isDark ? '#d4d4d4' : '#333',
        border: `1px solid ${isDark ? '#333' : '#e8e8e8'}`,
        ...style,
      }}>
        {parsedData}
      </pre>
    );
  }

  if (parsedData === null) {
    return (
      <span style={{ color: isDark ? '#666' : '#999', fontSize, fontStyle: 'italic' }}>
        null
      </span>
    );
  }

  const heightValue = typeof maxHeight === 'number' ? `${maxHeight}px` : maxHeight;

  return (
    <div
      ref={wrapperRef}
      tabIndex={-1}
      style={{
        borderRadius: 6,
        border: `1px solid ${isDark ? '#333' : '#e8e8e8'}`,
        display: 'flex',
        flexDirection: 'column',
        ...(hasHeightConstraint ? { maxHeight: heightValue } : {}),
        outline: 'none',
        ...style,
      }}
      className={`json-viewer-wrapper ${isDark ? 'json-viewer-dark' : 'json-viewer-light'}`}
    >
      {/* Search toggle button */}
      {showSearch && !searchOpen && (
        <div
          style={{
            position: 'sticky',
            top: 0,
            zIndex: 5,
            display: 'flex',
            justifyContent: 'flex-end',
            padding: '4px 8px 0',
            pointerEvents: 'none',
          }}
        >
          <span
            onClick={() => setSearchOpen(true)}
            title="Search (Ctrl+F)"
            style={{
              cursor: 'pointer',
              pointerEvents: 'auto',
              padding: '2px 6px',
              borderRadius: 4,
              color: isDark ? '#666' : '#bbb',
              fontSize: 12,
              display: 'inline-flex',
              alignItems: 'center',
              transition: 'color 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.color = isDark ? '#aaa' : '#666')}
            onMouseLeave={e => (e.currentTarget.style.color = isDark ? '#666' : '#bbb')}
          >
            <SearchOutlined />
          </span>
        </div>
      )}

      {/* Search bar */}
      {searchOpen && (
        <SearchBar
          matchCount={matchCount}
          currentIndex={currentIndex}
          onSearch={setSearchTerm}
          onNext={goNext}
          onPrev={goPrev}
          onClose={handleCloseSearch}
          isDark={isDark}
        />
      )}

      {/* JSON content */}
      <div
        ref={containerRef}
        style={{
          flex: 1,
          overflow: 'auto',
          minHeight: 0,
        }}
      >
        <JsonView
          src={parsedData}
          theme={isDark ? 'a11y' : 'default'}
          collapseStringsAfterLength={Number.MAX_SAFE_INTEGER}
          collapsed={collapseDepth === 0 ? true : (({ depth }: { depth: number }) => depth >= collapseDepth) as never}
          enableClipboard={enableCopy}
          style={{
            fontSize,
            fontFamily: '"JetBrains Mono", "Fira Code", "SF Mono", Menlo, monospace',
            padding: 12,
            background: 'transparent',
            lineHeight: 1.6,
          }}
        />
      </div>
    </div>
  );
});

JsonViewer.displayName = 'JsonViewer';

export default JsonViewer;
