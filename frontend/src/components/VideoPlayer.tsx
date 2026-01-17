import React, { useRef, useEffect, useCallback } from 'react';
import {
  Card,
  Space,
  Button,
  Slider,
  Typography,
  Tooltip,
  Spin,
  Empty,
  Switch,
  Select,
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  StepBackwardOutlined,
  StepForwardOutlined,
  SoundOutlined,
  MutedOutlined,
  FullscreenOutlined,
  SyncOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useVideoStore } from '../stores/videoStore';

const { Text } = Typography;

interface VideoPlayerProps {
  sessionId?: string;
  videoPath?: string;
  onTimeUpdate?: (timeMs: number) => void;
  onSyncTimeRequest?: (timeMs: number) => void;
  height?: number | string;
  showControls?: boolean;
  showSyncToggle?: boolean;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  sessionId,
  videoPath: propVideoPath,
  onTimeUpdate,
  onSyncTimeRequest,
  height = 300,
  showControls = true,
  showSyncToggle = true,
}) => {
  const { t } = useTranslation();
  const videoRef = useRef<HTMLVideoElement>(null);

  const {
    serviceInfo,
    currentVideoPath,
    currentMetadata,
    currentDataURL,
    isPlaying,
    currentTimeMs,
    volume,
    muted,
    playbackRate,
    syncEnabled,
    videoOffset,
    isFullscreen,
    isLoading,
    error,
    loadServiceInfo,
    loadSessionVideo,
    loadVideoAsDataURL,
    setPlaying,
    setCurrentTime,
    setVolume,
    setMuted,
    setPlaybackRate,
    setSyncEnabled,
    setIsFullscreen,
    timeToVideoTime,
    videoTimeToSessionTime,
    clearVideo,
  } = useVideoStore();

  // Initialize service info
  useEffect(() => {
    loadServiceInfo();
  }, []);

  // Load video when sessionId or videoPath changes
  useEffect(() => {
    if (sessionId) {
      loadSessionVideo(sessionId);
    } else if (propVideoPath) {
      loadVideoAsDataURL(propVideoPath);
    }

    return () => {
      // Don't clear on unmount if we want to preserve state
    };
  }, [sessionId, propVideoPath]);

  // Sync video element with store state
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    video.volume = volume;
    video.muted = muted;
    video.playbackRate = playbackRate;
  }, [volume, muted, playbackRate]);

  // Handle play/pause
  useEffect(() => {
    const video = videoRef.current;
    if (!video || !currentDataURL) return;

    if (isPlaying) {
      video.play().catch(() => setPlaying(false));
    } else {
      video.pause();
    }
  }, [isPlaying, currentDataURL]);

  // Handle time updates from video
  const handleTimeUpdate = useCallback(() => {
    const video = videoRef.current;
    if (!video) return;

    const timeMs = Math.round(video.currentTime * 1000);
    setCurrentTime(timeMs);

    if (onTimeUpdate) {
      onTimeUpdate(syncEnabled ? videoTimeToSessionTime(timeMs) : timeMs);
    }
  }, [syncEnabled, videoTimeToSessionTime, onTimeUpdate]);

  // Handle seeking to specific time
  const seekTo = useCallback((timeMs: number) => {
    const video = videoRef.current;
    if (!video) return;

    video.currentTime = timeMs / 1000;
    setCurrentTime(timeMs);
  }, []);

  // External seek request (from event timeline)
  const handleSyncTimeRequest = useCallback((sessionTimeMs: number) => {
    if (syncEnabled) {
      const videoTimeMs = timeToVideoTime(sessionTimeMs);
      if (videoTimeMs >= 0 && currentMetadata && videoTimeMs <= currentMetadata.durationMs) {
        seekTo(videoTimeMs);
      }
    }
  }, [syncEnabled, timeToVideoTime, currentMetadata, seekTo]);

  // Expose sync function
  useEffect(() => {
    if (onSyncTimeRequest) {
      // This is a bit hacky, but allows parent to request time sync
      (window as any).__videoPlayerSeekTo = handleSyncTimeRequest;
    }
    return () => {
      delete (window as any).__videoPlayerSeekTo;
    };
  }, [handleSyncTimeRequest, onSyncTimeRequest]);

  // Format time for display
  const formatTime = (ms: number) => {
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  // Toggle fullscreen
  const toggleFullscreen = () => {
    const container = videoRef.current?.parentElement;
    if (!container) return;

    if (!document.fullscreenElement) {
      container.requestFullscreen();
      setIsFullscreen(true);
    } else {
      document.exitFullscreen();
      setIsFullscreen(false);
    }
  };

  // Skip forward/backward
  const skip = (seconds: number) => {
    const newTime = Math.max(0, Math.min(
      currentTimeMs + seconds * 1000,
      currentMetadata?.durationMs || 0
    ));
    seekTo(newTime);
  };

  // Check if video service is available
  if (!serviceInfo?.available) {
    return (
      <Card size="small" style={{ height }}>
        <Empty
          description={
            <Space direction="vertical" size="small">
              <Text>{t('video.ffmpeg_not_found', 'FFmpeg not found')}</Text>
              <Text type="secondary">
                {t('video.ffmpeg_required', 'FFmpeg is required for video playback. Please install FFmpeg and restart.')}
              </Text>
            </Space>
          }
        />
      </Card>
    );
  }

  // Loading state
  if (isLoading) {
    return (
      <Card size="small" style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin tip={t('video.loading', 'Loading video...')} />
      </Card>
    );
  }

  // No video state
  if (!currentVideoPath && !currentDataURL) {
    return (
      <Card size="small" style={{ height }}>
        <Empty description={t('video.no_video', 'No video available')} />
      </Card>
    );
  }

  // Error state
  if (error) {
    return (
      <Card size="small" style={{ height }}>
        <Empty
          description={
            <Space direction="vertical" size="small">
              <Text type="danger">{t('video.error', 'Error loading video')}</Text>
              <Text type="secondary">{error}</Text>
              <Button
                size="small"
                icon={<ReloadOutlined />}
                onClick={() => {
                  clearVideo();
                  if (sessionId) loadSessionVideo(sessionId);
                  else if (propVideoPath) loadVideoAsDataURL(propVideoPath);
                }}
              >
                {t('common.retry', 'Retry')}
              </Button>
            </Space>
          }
        />
      </Card>
    );
  }

  const durationMs = currentMetadata?.durationMs || 0;

  return (
    <Card
      size="small"
      bodyStyle={{ padding: 0 }}
      style={{ height: 'auto' }}
    >
      {/* Video Element */}
      <div
        style={{
          position: 'relative',
          backgroundColor: '#000',
          height: typeof height === 'number' ? height - 80 : `calc(${height} - 80px)`,
        }}
      >
        {currentDataURL ? (
          <video
            ref={videoRef}
            src={currentDataURL}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'contain',
            }}
            onTimeUpdate={handleTimeUpdate}
            onPlay={() => setPlaying(true)}
            onPause={() => setPlaying(false)}
            onEnded={() => setPlaying(false)}
          />
        ) : (
          <div
            style={{
              width: '100%',
              height: '100%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#666',
            }}
          >
            <Text type="secondary">
              {t('video.large_video', 'Video too large for embedded playback')}
            </Text>
          </div>
        )}
      </div>

      {/* Controls */}
      {showControls && (
        <div style={{ padding: '8px 12px' }}>
          {/* Progress Bar */}
          <Slider
            min={0}
            max={durationMs}
            value={currentTimeMs}
            onChange={(value) => seekTo(value)}
            tooltip={{
              formatter: (value) => formatTime(value || 0),
            }}
            style={{ margin: '0 0 8px 0' }}
          />

          {/* Control Buttons */}
          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
            <Space>
              {/* Play/Pause */}
              <Tooltip title={isPlaying ? t('video.pause', 'Pause') : t('video.play', 'Play')}>
                <Button
                  type="text"
                  icon={isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                  onClick={() => setPlaying(!isPlaying)}
                />
              </Tooltip>

              {/* Skip backward */}
              <Tooltip title={t('video.skip_back', 'Skip -5s')}>
                <Button
                  type="text"
                  icon={<StepBackwardOutlined />}
                  onClick={() => skip(-5)}
                />
              </Tooltip>

              {/* Skip forward */}
              <Tooltip title={t('video.skip_forward', 'Skip +5s')}>
                <Button
                  type="text"
                  icon={<StepForwardOutlined />}
                  onClick={() => skip(5)}
                />
              </Tooltip>

              {/* Time display */}
              <Text type="secondary" style={{ minWidth: 100 }}>
                {formatTime(currentTimeMs)} / {formatTime(durationMs)}
              </Text>
            </Space>

            <Space>
              {/* Sync toggle */}
              {showSyncToggle && (
                <Tooltip title={t('video.sync_events', 'Sync with event timeline')}>
                  <Space size="small">
                    <SyncOutlined style={{ color: syncEnabled ? '#1890ff' : '#999' }} />
                    <Switch
                      size="small"
                      checked={syncEnabled}
                      onChange={setSyncEnabled}
                    />
                  </Space>
                </Tooltip>
              )}

              {/* Playback speed */}
              <Select
                size="small"
                value={playbackRate}
                onChange={setPlaybackRate}
                style={{ width: 70 }}
                options={[
                  { label: '0.5x', value: 0.5 },
                  { label: '1x', value: 1 },
                  { label: '1.5x', value: 1.5 },
                  { label: '2x', value: 2 },
                ]}
              />

              {/* Volume */}
              <Tooltip title={muted ? t('video.unmute', 'Unmute') : t('video.mute', 'Mute')}>
                <Button
                  type="text"
                  icon={muted ? <MutedOutlined /> : <SoundOutlined />}
                  onClick={() => setMuted(!muted)}
                />
              </Tooltip>

              {/* Fullscreen */}
              <Tooltip title={t('video.fullscreen', 'Fullscreen')}>
                <Button
                  type="text"
                  icon={<FullscreenOutlined />}
                  onClick={toggleFullscreen}
                />
              </Tooltip>
            </Space>
          </Space>
        </div>
      )}
    </Card>
  );
};

export default VideoPlayer;
