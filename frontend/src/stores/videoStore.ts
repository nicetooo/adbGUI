import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

// Wails runtime bindings
const GetVideoServiceInfo = (window as any).go?.main?.App?.GetVideoServiceInfo;
const GetVideoMetadata = (window as any).go?.main?.App?.GetVideoMetadata;
const GetVideoFrame = (window as any).go?.main?.App?.GetVideoFrame;
const GetVideoThumbnails = (window as any).go?.main?.App?.GetVideoThumbnails;
const GetSessionVideoInfo = (window as any).go?.main?.App?.GetSessionVideoInfo;
const ListRecordings = (window as any).go?.main?.App?.ListRecordings;
const ReadVideoFileAsDataURL = (window as any).go?.main?.App?.ReadVideoFileAsDataURL;

// Types
export interface VideoServiceInfo {
  available: boolean;
  ffmpegPath?: string;
  ffprobePath?: string;
}

export interface VideoMetadata {
  path: string;
  duration: number;
  durationMs: number;
  width: number;
  height: number;
  frameRate: number;
  codec: string;
  bitRate: number;
  totalFrames: number;
  thumbnailPath?: string;
}

export interface VideoThumbnail {
  timeMs: number;
  base64: string;
  width: number;
  height: number;
}

export interface SessionVideoInfo {
  hasVideo: boolean;
  videoPath?: string;
  videoOffset?: number;
  metadata?: VideoMetadata;
  error?: string;
}

export interface Recording {
  name: string;
  path: string;
  size: number;
  modified: number;
  duration?: number;
  width?: number;
  height?: number;
}

interface VideoState {
  // Service info
  serviceInfo: VideoServiceInfo | null;

  // Current video
  currentVideoPath: string | null;
  currentMetadata: VideoMetadata | null;
  currentDataURL: string | null;

  // Thumbnails
  thumbnails: VideoThumbnail[];

  // Playback state
  isPlaying: boolean;
  currentTimeMs: number;
  volume: number;
  muted: boolean;
  playbackRate: number;

  // Sync state
  syncEnabled: boolean;
  videoOffset: number; // Offset from session start

  // UI state
  isFullscreen: boolean;

  // Loading states
  isLoading: boolean;
  isLoadingThumbnails: boolean;

  // Error
  error: string | null;

  // Recordings list
  recordings: Recording[];

  // Actions
  loadServiceInfo: () => Promise<void>;
  loadVideo: (videoPath: string) => Promise<void>;
  loadVideoAsDataURL: (videoPath: string) => Promise<void>;
  loadSessionVideo: (sessionId: string) => Promise<SessionVideoInfo | null>;
  loadThumbnails: (intervalMs?: number, width?: number) => Promise<void>;
  getFrame: (timeMs: number, width?: number) => Promise<string | null>;

  // Playback controls
  setPlaying: (playing: boolean) => void;
  setCurrentTime: (timeMs: number) => void;
  setVolume: (volume: number) => void;
  setMuted: (muted: boolean) => void;
  setPlaybackRate: (rate: number) => void;

  // Sync controls
  setSyncEnabled: (enabled: boolean) => void;
  setVideoOffset: (offset: number) => void;

  // UI controls
  setIsFullscreen: (isFullscreen: boolean) => void;

  // Utility
  timeToVideoTime: (sessionTimeMs: number) => number;
  videoTimeToSessionTime: (videoTimeMs: number) => number;

  // Recordings
  loadRecordings: () => Promise<void>;

  // Clear
  clearVideo: () => void;
}

export const useVideoStore = create<VideoState>()(
  immer((set, get) => ({
    serviceInfo: null,
    currentVideoPath: null,
    currentMetadata: null,
    currentDataURL: null,
    thumbnails: [],
    isPlaying: false,
    currentTimeMs: 0,
    volume: 1,
    muted: false,
    playbackRate: 1,
    syncEnabled: true,
    videoOffset: 0,
    isFullscreen: false,
    isLoading: false,
    isLoadingThumbnails: false,
    error: null,
    recordings: [],

    loadServiceInfo: async () => {
      if (!GetVideoServiceInfo) return;

      try {
        const info = await GetVideoServiceInfo();
        set((state) => {
          state.serviceInfo = info;
        });
      } catch (err) {
        console.error('Failed to load video service info:', err);
      }
    },

    loadVideo: async (videoPath: string) => {
      if (!GetVideoMetadata) return;

      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        const metadata = await GetVideoMetadata(videoPath);
        set((state) => {
          state.currentVideoPath = videoPath;
          state.currentMetadata = metadata;
          state.currentTimeMs = 0;
          state.isLoading = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isLoading = false;
        });
      }
    },

    loadVideoAsDataURL: async (videoPath: string) => {
      if (!ReadVideoFileAsDataURL) return;

      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        // First load metadata
        if (GetVideoMetadata) {
          const metadata = await GetVideoMetadata(videoPath);
          set((state) => {
            state.currentMetadata = metadata;
          });
        }

        // Then load as data URL (for small videos)
        const dataURL = await ReadVideoFileAsDataURL(videoPath);
        set((state) => {
          state.currentVideoPath = videoPath;
          state.currentDataURL = dataURL;
          state.currentTimeMs = 0;
          state.isLoading = false;
        });
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isLoading = false;
        });
      }
    },

    loadSessionVideo: async (sessionId: string) => {
      if (!GetSessionVideoInfo) return null;

      set((state) => {
        state.isLoading = true;
        state.error = null;
      });

      try {
        const info = await GetSessionVideoInfo(sessionId);

        if (info.hasVideo && info.videoPath) {
          set((state) => {
            state.currentVideoPath = info.videoPath;
            state.currentMetadata = info.metadata || null;
            state.videoOffset = info.videoOffset || 0;
            state.currentTimeMs = 0;
            state.isLoading = false;
          });

          // Try to load as data URL for playback
          if (info.metadata && info.metadata.durationMs < 60000) {
            // Only load as data URL if video is less than 1 minute
            await get().loadVideoAsDataURL(info.videoPath);
          }
        } else {
          set((state) => {
            state.currentVideoPath = null;
            state.currentMetadata = null;
            state.isLoading = false;
            if (info.error) {
              state.error = info.error;
            }
          });
        }

        return info;
      } catch (err) {
        set((state) => {
          state.error = String(err);
          state.isLoading = false;
        });
        return null;
      }
    },

    loadThumbnails: async (intervalMs = 5000, width = 160) => {
      if (!GetVideoThumbnails) return;

      const { currentVideoPath } = get();
      if (!currentVideoPath) return;

      set((state) => {
        state.isLoadingThumbnails = true;
      });

      try {
        const thumbnails = await GetVideoThumbnails(currentVideoPath, intervalMs, width);
        set((state) => {
          state.thumbnails = thumbnails || [];
          state.isLoadingThumbnails = false;
        });
      } catch (err) {
        console.error('Failed to load thumbnails:', err);
        set((state) => {
          state.isLoadingThumbnails = false;
        });
      }
    },

    getFrame: async (timeMs: number, width = 320) => {
      if (!GetVideoFrame) return null;

      const { currentVideoPath } = get();
      if (!currentVideoPath) return null;

      try {
        return await GetVideoFrame(currentVideoPath, timeMs, width);
      } catch (err) {
        console.error('Failed to get frame:', err);
        return null;
      }
    },

    setPlaying: (playing: boolean) => {
      set((state) => {
        state.isPlaying = playing;
      });
    },

    setCurrentTime: (timeMs: number) => {
      set((state) => {
        state.currentTimeMs = timeMs;
      });
    },

    setVolume: (volume: number) => {
      set((state) => {
        state.volume = Math.max(0, Math.min(1, volume));
      });
    },

    setMuted: (muted: boolean) => {
      set((state) => {
        state.muted = muted;
      });
    },

    setPlaybackRate: (rate: number) => {
      set((state) => {
        state.playbackRate = rate;
      });
    },

    setSyncEnabled: (enabled: boolean) => {
      set((state) => {
        state.syncEnabled = enabled;
      });
    },

    setVideoOffset: (offset: number) => {
      set((state) => {
        state.videoOffset = offset;
      });
    },

    setIsFullscreen: (isFullscreen: boolean) => {
      set((state) => {
        state.isFullscreen = isFullscreen;
      });
    },

    timeToVideoTime: (sessionTimeMs: number) => {
      const { videoOffset } = get();
      return sessionTimeMs - videoOffset;
    },

    videoTimeToSessionTime: (videoTimeMs: number) => {
      const { videoOffset } = get();
      return videoTimeMs + videoOffset;
    },

    loadRecordings: async () => {
      if (!ListRecordings) return;

      try {
        const recordings = await ListRecordings();
        set((state) => {
          state.recordings = recordings || [];
        });
      } catch (err) {
        console.error('Failed to load recordings:', err);
      }
    },

    clearVideo: () => {
      set((state) => {
        state.currentVideoPath = null;
        state.currentMetadata = null;
        state.currentDataURL = null;
        state.thumbnails = [];
        state.currentTimeMs = 0;
        state.isPlaying = false;
        state.videoOffset = 0;
        state.error = null;
      });
    },
  }))
);
