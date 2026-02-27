'use client';

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { TimelineState, TimelineActions, TimelineTrack, TimelineClip } from '@/types/timeline';
import { api, ApiError, TimelineData } from '@/lib/api';

// Generate unique IDs
const generateId = () => Math.random().toString(36).substring(2, 9);

// Initial state
const initialState: TimelineState = {
  tracks: [
    {
      id: 'track-1',
      name: 'Video 1',
      type: 'video',
      clips: [],
      isMuted: false,
      isLocked: false,
      isVisible: true,
    },
    {
      id: 'track-2',
      name: 'Audio 1',
      type: 'audio',
      clips: [],
      isMuted: false,
      isLocked: false,
      isVisible: true,
    },
  ],
  currentTime: 0,
  totalDuration: 60, // 60 seconds default
  zoom: 50, // 50 pixels per second
  selectedClipIds: [],
  selectedTrackId: null,
  isPlaying: false,
};

// Backend sync state
interface BackendState {
  timelineId: number | null;
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
  lastSaved: Date | null;
}

// Combine state and actions type
interface TimelineStore extends TimelineState, TimelineActions, BackendState {
  // Backend actions
  loadTimeline: (id: number) => Promise<void>;
  createTimeline: (title: string, description?: string) => Promise<void>;
  saveTimeline: () => Promise<void>;
  clearError: () => void;
  
  // Drag and drop
  dragClip: (clipId: string, newTrackId: string, newStartTime: number) => void;
  endClipDrag: () => Promise<void>;
  
  // Resize
  resizeClip: (clipId: string, newDuration: number, startTime?: number) => void;
  endClipResize: () => Promise<void>;
}

// Create the store
export const useTimelineStore = create<TimelineStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      timelineId: null,
      isLoading: false,
      isSaving: false,
      error: null,
      lastSaved: null,

      // Backend actions
      loadTimeline: async (id: number) => {
        set({ isLoading: true, error: null });
        try {
          const timeline = await api.getTimeline(id);
          set({ 
            timelineId: id,
            isLoading: false,
          });
        } catch (error) {
          const apiError = error as ApiError;
          set({ 
            error: apiError.message || 'Failed to load timeline',
            isLoading: false 
          });
        }
      },

      createTimeline: async (title: string, description?: string) => {
        set({ isLoading: true, error: null });
        try {
          const timeline = await api.createTimeline({ title, description });
          set({ 
            timelineId: timeline.id,
            isLoading: false,
            lastSaved: new Date(),
          });
        } catch (error) {
          const apiError = error as ApiError;
          set({ 
            error: apiError.message || 'Failed to create timeline',
            isLoading: false 
          });
        }
      },

      saveTimeline: async () => {
        const { timelineId } = get();
        if (!timelineId) {
          set({ error: 'No timeline loaded' });
          return;
        }
        
        set({ isSaving: true, error: null });
        try {
          const { tracks } = get();
          await api.saveTimelineData(timelineId, {
            title: 'Timeline',
            tracks,
          });
          set({ lastSaved: new Date() });
        } catch (error) {
          const apiError = error as ApiError;
          set({ error: apiError.message || 'Failed to save timeline' });
        } finally {
          set({ isSaving: false });
        }
      },

      clearError: () => set({ error: null }),

      // Track actions
      addTrack: (type) => {
        set((state) => {
          const trackCount = state.tracks.filter((t) => t.type === type).length;
          const newTrack: TimelineTrack = {
            id: generateId(),
            name: `${type === 'video' ? 'Video' : 'Audio'} ${trackCount + 1}`,
            type,
            clips: [],
            isMuted: false,
            isLocked: false,
            isVisible: true,
          };
          return { tracks: [...state.tracks, newTrack] };
        });
      },

      removeTrack: (trackId) => {
        set((state) => ({
          tracks: state.tracks.filter((t) => t.id !== trackId),
          selectedClipIds: state.selectedClipIds.filter(
            (id) => !state.tracks.find((t) => t.id === trackId)?.clips.some((c) => c.id === id)
          ),
          selectedTrackId: state.selectedTrackId === trackId ? null : state.selectedTrackId,
        }));
      },

      moveTrack: (trackId, newIndex) => {
        set((state) => {
          const tracks = [...state.tracks];
          const oldIndex = tracks.findIndex((t) => t.id === trackId);
          if (oldIndex === -1) return state;
          
          const [movedTrack] = tracks.splice(oldIndex, 1);
          tracks.splice(newIndex, 0, movedTrack);
          return { tracks };
        });
      },

      updateTrack: (trackId, updates) => {
        set((state) => ({
          tracks: state.tracks.map((t) =>
            t.id === trackId ? { ...t, ...updates } : t
          ),
        }));
      },

      // Clip actions
      addClip: (trackId, clip) => {
        set((state) => ({
          tracks: state.tracks.map((t) =>
            t.id === trackId
              ? { ...t, clips: [...t.clips, { ...clip, id: generateId() }] }
              : t
          ),
        }));
      },

      removeClip: (clipId) => {
        set((state) => ({
          tracks: state.tracks.map((t) => ({
            ...t,
            clips: t.clips.filter((c) => c.id !== clipId),
          })),
          selectedClipIds: state.selectedClipIds.filter((id) => id !== clipId),
        }));
      },

      moveClip: (clipId, newTrackId, newStartTime) => {
        set((state) => {
          let clipToMove: TimelineClip | undefined;
          
          const tracks = state.tracks.map((t) => {
            const clip = t.clips.find((c) => c.id === clipId);
            if (clip) {
              clipToMove = clip;
              return { ...t, clips: t.clips.filter((c) => c.id !== clipId) };
            }
            return t;
          });
          
          if (!clipToMove) return state;
          
          return {
            tracks: tracks.map((t) =>
              t.id === newTrackId
                ? {
                    ...t,
                    clips: [
                      ...t.clips,
                      { ...clipToMove!, trackId: newTrackId, startTime: newStartTime },
                    ],
                  }
                : t
            ),
          };
        });
      },

      updateClip: (clipId, updates) => {
        set((state) => ({
          tracks: state.tracks.map((t) => ({
            ...t,
            clips: t.clips.map((c) =>
              c.id === clipId ? { ...c, ...updates } : c
            ),
          })),
        }));
      },

      // Drag and drop with immediate save
      dragClip: (clipId, newTrackId, newStartTime) => {
        set((state) => {
          let clipToMove: TimelineClip | undefined;
          
          const tracks = state.tracks.map((t) => {
            const clip = t.clips.find((c) => c.id === clipId);
            if (clip) {
              clipToMove = clip;
              return { ...t, clips: t.clips.filter((c) => c.id !== clipId) };
            }
            return t;
          });
          
          if (!clipToMove) return state;
          
          return {
            tracks: tracks.map((t) =>
              t.id === newTrackId
                ? {
                    ...t,
                    clips: [
                      ...t.clips,
                      { ...clipToMove!, trackId: newTrackId, startTime: newStartTime },
                    ].sort((a, b) => a.startTime - b.startTime),
                  }
                : t
            ),
          };
        });
      },

      endClipDrag: async () => {
        await get().saveTimeline();
      },

      // Resize with immediate save
      resizeClip: (clipId, newDuration, startTime) => {
        set((state) => ({
          tracks: state.tracks.map((t) => ({
            ...t,
            clips: t.clips.map((c) =>
              c.id === clipId 
                ? { 
                    ...c, 
                    duration: newDuration,
                    ...(startTime !== undefined && { startTime })
                  } 
                : c
            ),
          })),
        }));
      },

      endClipResize: async () => {
        await get().saveTimeline();
      },

      // Selection actions
      selectClip: (clipId, multi = false) =>
        set((state) => ({
          selectedClipIds: multi
            ? [...state.selectedClipIds, clipId]
            : [clipId],
        })),

      deselectClip: (clipId) =>
        set((state) => ({
          selectedClipIds: state.selectedClipIds.filter((id) => id !== clipId),
        })),

      clearSelection: () =>
        set(() => ({
          selectedClipIds: [],
          selectedTrackId: null,
        })),

      selectTrack: (trackId) =>
        set(() => ({
          selectedTrackId: trackId,
          selectedClipIds: [],
        })),

      // Playback actions
      setCurrentTime: (time) =>
        set(() => ({
          currentTime: Math.max(0, time),
        })),

      setPlaying: (playing) =>
        set(() => ({
          isPlaying: playing,
        })),

      setZoom: (zoom) =>
        set(() => ({
          zoom: Math.max(10, Math.min(200, zoom)),
        })),

      // Timeline actions
      splitClip: (clipId, splitTime) =>
        set((state) => ({
          tracks: state.tracks.map((t) => ({
            ...t,
            clips: t.clips.flatMap((c) => {
              if (c.id !== clipId) return [c];
              
              const splitPoint = splitTime - c.startTime;
              if (splitPoint <= 0 || splitPoint >= c.duration) return [c];
              
              const firstPart: TimelineClip = {
                ...c,
                duration: splitPoint,
              };
              
              const secondPart: TimelineClip = {
                ...c,
                id: generateId(),
                startTime: c.startTime + splitPoint,
                duration: c.duration - splitPoint,
              };
              
              return [firstPart, secondPart];
            }),
          })),
        })),

      trimClip: (clipId, startOffset, endOffset) =>
        set((state) => ({
          tracks: state.tracks.map((t) => ({
            ...t,
            clips: t.clips.map((c) => {
              if (c.id !== clipId) return c;
              return {
                ...c,
                startTime: c.startTime + startOffset,
                duration: c.duration - startOffset - endOffset,
              };
            }),
          })),
        })),
    }),
    { name: 'TimelineStore' }
  )
);

// Selector hooks for better performance
export const useTracks = () => useTimelineStore((state) => state.tracks);
export const useCurrentTime = () => useTimelineStore((state) => state.currentTime);
export const useIsPlaying = () => useTimelineStore((state) => state.isPlaying);
export const useZoom = () => useTimelineStore((state) => state.zoom);
export const useSelectedClips = () => useTimelineStore((state) => state.selectedClipIds);
export const useSelectedTrack = () => useTimelineStore((state) => state.selectedTrackId);
export const useTimelineLoading = () => useTimelineStore((state) => state.isLoading);
export const useTimelineSaving = () => useTimelineStore((state) => state.isSaving);
export const useTimelineError = () => useTimelineStore((state) => state.error);
export const useLastSaved = () => useTimelineStore((state) => state.lastSaved);
