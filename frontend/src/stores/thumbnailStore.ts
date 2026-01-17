import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface ThumbnailState {
  thumbnails: Record<string, string>; // path -> thumbnail data URL
  loadingPaths: Set<string>;
  
  setThumbnail: (path: string, thumb: string) => void;
  setLoading: (path: string, loading: boolean) => void;
  getThumbnail: (path: string) => string | null;
  isLoading: (path: string) => boolean;
}

export const useThumbnailStore = create<ThumbnailState>()(
  immer((set, get) => ({
    thumbnails: {},
    loadingPaths: new Set(),
    
    setThumbnail: (path, thumb) => set((state: ThumbnailState) => {
      state.thumbnails[path] = thumb;
      state.loadingPaths.delete(path);
    }),
    
    setLoading: (path, loading) => set((state: ThumbnailState) => {
      if (loading) {
        state.loadingPaths.add(path);
      } else {
        state.loadingPaths.delete(path);
      }
    }),
    
    getThumbnail: (path) => {
      return get().thumbnails[path] || null;
    },
    
    isLoading: (path) => {
      return get().loadingPaths.has(path);
    },
  }))
);
