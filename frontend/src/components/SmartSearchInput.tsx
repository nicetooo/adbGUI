/**
 * SmartSearchInput - 智能搜索输入框
 * 支持 AI 自然语言搜索（需要配置 LLM）
 *
 * 设计原则：
 * - AI 搜索只通过手动触发（快捷键或点击按钮）
 * - AI 解析成功后清空输入框，显示解析结果标签
 * - 普通输入直接用于传统文本搜索
 */
import React, { useCallback, useRef } from 'react';
import { Input, Tooltip, Spin, Tag, Space, theme, message } from 'antd';
import { SearchOutlined, RobotOutlined, ThunderboltOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAIStore } from '../stores/aiStore';
import { useSmartSearchStore } from '../stores/smartSearchStore';

// Parsed query result from AI
export interface NLParsedQuery {
  types?: string[];
  sources?: string[];
  levels?: string[];
  keywords?: string[];
  timeRange?: {
    startMs?: number;
    endMs?: number;
    last?: string;
  };
  context?: string;
}

export interface NLQueryResult {
  query: NLParsedQuery;
  explanation: string;
  confidence: number;
  suggestions?: string[];
}

export interface SmartSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  onParsedQuery?: (result: NLQueryResult | null) => void;
  placeholder?: string;
  sessionId?: string;
  disabled?: boolean;
  style?: React.CSSProperties;
  allowClear?: boolean;
  // Context hint for what kind of data is being searched
  searchContext?: 'logcat' | 'events' | 'network' | 'general';
}

const SmartSearchInput: React.FC<SmartSearchInputProps> = ({
  value,
  onChange,
  onParsedQuery,
  placeholder,
  sessionId,
  disabled,
  style,
  allowClear = true,
  searchContext = 'general',
}) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const { serviceInfo, config, parseNaturalQuery } = useAIStore();
  
  // Store state
  const {
    isParsing,
    aiFilterActive,
    lastParsedResult,
    setIsParsing,
    setAiFilterActive,
    setLastParsedResult,
    clearAiFilter,
  } = useSmartSearchStore();
  
  const inputRef = useRef<any>(null);

  // Check if AI natural search is available
  const isAIAvailable = serviceInfo?.status === 'ready' &&
    config?.enabled &&
    config?.features?.naturalSearch;

  // Handle input change - just pass through for traditional search
  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    onChange(newValue);

    // If AI filter was active and user starts typing, clear it
    if (aiFilterActive && newValue) {
      clearAiFilter();
      onParsedQuery?.(null);
    }

    // If input is cleared, also clear AI filter
    if (!newValue) {
      clearAiFilter();
      onParsedQuery?.(null);
    }
  }, [onChange, aiFilterActive, onParsedQuery, clearAiFilter]);

  // Parse with AI - manual trigger only
  const handleAIParse = useCallback(async () => {
    if (!value || !isAIAvailable || isParsing) return;

    setIsParsing(true);
    message.info(t('smart_search.parsing', 'AI parsing...'));

    try {
      const result = await parseNaturalQuery(value, sessionId || '');
      if (result && result.query) {
        // Check if AI returned meaningful results
        const hasResults = result.query.types?.length ||
          result.query.sources?.length ||
          result.query.levels?.length ||
          result.query.keywords?.length ||
          result.query.timeRange;

        if (hasResults) {
          setLastParsedResult(result);
          setAiFilterActive(true);
          onParsedQuery?.(result);
          // Clear input after successful AI parse
          onChange('');
          message.success(result.explanation || t('smart_search.applied', 'AI filter applied'));
        } else {
          message.warning(t('smart_search.no_results', 'AI could not parse the query'));
        }
      }
    } catch (err) {
      console.error('[SmartSearch] AI parse failed:', err);
      message.error(t('smart_search.parse_failed', 'AI parsing failed'));
    } finally {
      setIsParsing(false);
    }
  }, [value, isAIAvailable, isParsing, parseNaturalQuery, sessionId, onParsedQuery, onChange, t]);

  // Clear AI filter
  const handleClearAIFilter = useCallback(() => {
    clearAiFilter();
    onParsedQuery?.(null);
    inputRef.current?.focus();
  }, [onParsedQuery, clearAiFilter]);

  // Handle keyboard shortcuts
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    // Ctrl/Cmd + Enter to trigger AI search
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      if (value && isAIAvailable) {
        handleAIParse();
      }
    }
  }, [value, isAIAvailable, handleAIParse]);

  // Get context-aware placeholder
  const getPlaceholder = () => {
    if (placeholder) return placeholder;

    if (isAIAvailable) {
      const shortcutHint = navigator.platform.includes('Mac') ? '⌘+Enter' : 'Ctrl+Enter';
      switch (searchContext) {
        case 'logcat':
          return t('smart_search.placeholder_logcat_with_hint', { shortcut: shortcutHint, defaultValue: `Search logs (${shortcutHint} for AI)` });
        case 'events':
          return t('smart_search.placeholder_events_with_hint', { shortcut: shortcutHint, defaultValue: `Search events (${shortcutHint} for AI)` });
        case 'network':
          return t('smart_search.placeholder_network_with_hint', { shortcut: shortcutHint, defaultValue: `Search requests (${shortcutHint} for AI)` });
        default:
          return t('smart_search.placeholder_general_with_hint', { shortcut: shortcutHint, defaultValue: `Search (${shortcutHint} for AI)` });
      }
    }

    return t('smart_search.placeholder_basic', 'Search...');
  };

  // Render AI filter tags
  const renderAIFilterTags = () => {
    if (!aiFilterActive || !lastParsedResult?.query) return null;

    const { query } = lastParsedResult;
    const tags: React.ReactNode[] = [];

    if (query.sources?.length) {
      query.sources.forEach(source => {
        tags.push(
          <Tag key={`source-${source}`} color="green" style={{ fontSize: 10, padding: '0 4px' }}>
            {source}
          </Tag>
        );
      });
    }

    if (query.types?.length) {
      query.types.forEach(type => {
        tags.push(
          <Tag key={`type-${type}`} color="blue" style={{ fontSize: 10, padding: '0 4px' }}>
            {type}
          </Tag>
        );
      });
    }

    if (query.levels?.length) {
      query.levels.forEach(level => {
        const color = level === 'error' || level === 'fatal' ? 'red' : level === 'warn' ? 'orange' : 'default';
        tags.push(
          <Tag key={`level-${level}`} color={color} style={{ fontSize: 10, padding: '0 4px' }}>
            {level}
          </Tag>
        );
      });
    }

    if (query.timeRange?.last) {
      tags.push(
        <Tag key="time" color="purple" style={{ fontSize: 10, padding: '0 4px' }}>
          {query.timeRange.last}
        </Tag>
      );
    }

    if (query.keywords?.length) {
      query.keywords.slice(0, 3).forEach(kw => {
        tags.push(
          <Tag key={`kw-${kw}`} style={{ fontSize: 10, padding: '0 4px' }}>
            "{kw}"
          </Tag>
        );
      });
    }

    return tags;
  };

  // Get shortcut hint based on platform
  const shortcutHint = navigator.platform.includes('Mac') ? '⌘+Enter' : 'Ctrl+Enter';

  return (
    <div style={{ position: 'relative', ...style }}>
      {/* AI Filter Active State */}
      {aiFilterActive && lastParsedResult ? (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '4px 8px',
          backgroundColor: token.colorPrimaryBg,
          borderRadius: token.borderRadius,
          border: `1px solid ${token.colorPrimaryBorder}`,
        }}>
          <RobotOutlined style={{ color: token.colorPrimary, fontSize: 14 }} />
          <Space size={4} wrap style={{ flex: 1 }}>
            {renderAIFilterTags()}
          </Space>
          <Tooltip title={t('smart_search.clear_filter', 'Clear AI Filter')}>
            <CloseCircleOutlined
              style={{ color: token.colorTextSecondary, cursor: 'pointer' }}
              onClick={handleClearAIFilter}
            />
          </Tooltip>
        </div>
      ) : (
        /* Normal Search Input */
        <Input
          ref={inputRef}
          prefix={
            isParsing ? (
              <Spin size="small" />
            ) : (
              <SearchOutlined />
            )
          }
          suffix={
            isAIAvailable && value ? (
              <Tooltip
                title={
                  <span>
                    {t('smart_search.ai_search', 'AI Smart Search')}
                    <br />
                    <span style={{ opacity: 0.7, fontSize: 11 }}>{shortcutHint}</span>
                  </span>
                }
              >
                <ThunderboltOutlined
                  style={{
                    color: token.colorTextSecondary,
                    cursor: isParsing ? 'wait' : 'pointer',
                    transition: 'color 0.2s',
                  }}
                  onClick={isParsing ? undefined : handleAIParse}
                  onMouseEnter={(e) => !isParsing && (e.currentTarget.style.color = token.colorPrimary)}
                  onMouseLeave={(e) => !isParsing && (e.currentTarget.style.color = token.colorTextSecondary)}
                />
              </Tooltip>
            ) : undefined
          }
          placeholder={getPlaceholder()}
          value={value}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          disabled={disabled || isParsing}
          allowClear={allowClear}
          style={{ width: '100%' }}
        />
      )}

      {/* AI Filter explanation */}
      {aiFilterActive && lastParsedResult?.explanation && (
        <div style={{
          fontSize: 11,
          color: token.colorTextSecondary,
          marginTop: 4,
          fontStyle: 'italic',
        }}>
          {lastParsedResult.explanation}
        </div>
      )}
    </div>
  );
};

export default SmartSearchInput;
