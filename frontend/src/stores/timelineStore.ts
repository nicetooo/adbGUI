import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import type { Session } from './sessionStore';

interface TimelineState {
  sessionList: Session[];
  selectedSessionId: string | null;
  autoScroll: boolean;
  
  setSessionList: (sessions: Session[]) => void;
  setSelectedSessionId: (id: string | null) => void;
  setAutoScroll: (auto: boolean) => void;
  toggleAutoScroll: () => void;
}

export const useTimelineStore = create<TimelineState>()(
  immer((set) => ({
    sessionList: [],
    selectedSessionId: null,
    autoScroll: true,
    
    setSessionList: (sessions) => set((state: TimelineState) => {
      state.sessionList = sessions;
    }),
    
    setSelectedSessionId: (id) => set((state: TimelineState) => {
      state.selectedSessionId = id;
    }),
    
    setAutoScroll: (auto) => set((state: TimelineState) => {
      state.autoScroll = auto;
    }),
    
    toggleAutoScroll: () => set((state: TimelineState) => {
      state.autoScroll = !state.autoScroll;
    }),
  }))
);
