# Renderowl 2.0 - Timeline Editor Architecture

## Frontend Stack Decision
**Framework**: Next.js 15 + React 19 + TypeScript

## 1. Timeline Editor Architecture Overview

### Core Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           TIMELINE EDITOR LAYOUT                            │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     VIDEO PREVIEW CANVAS                            │   │
│  │                    (Fabric.js + HTML5 Video)                        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      TIMELINE TRACKS                                │   │
│  │  ┌─────────┬─────────┬─────────┬─────────┬─────────┬────────────┐  │   │
│  │  │ Video 1 │ Video 2 │         │  Video  │         │            │  │   │
│  │  │ Track   │ Track   │         │  Track  │         │            │  │   │
│  │  └─────────┴─────────┴─────────┴─────────┴─────────┴────────────┘  │   │
│  │  ┌─────────┬─────────┬─────────┬─────────┬─────────┬────────────┐  │   │
│  │  │ Audio 1 │         │ Audio 2 │         │ Audio 3 │            │  │   │
│  │  │ Track   │         │ Track   │         │ Track   │            │  │   │
│  │  └─────────┴─────────┴─────────┴─────────┴─────────┴────────────┘  │   │
│  │  ┌─────────┬─────────┬─────────┬─────────┬─────────┬────────────┐  │   │
│  │  │Caption 1│         │Caption 2│         │         │            │  │   │
│  │  │ Track   │         │ Track   │         │         │            │  │   │
│  │  └─────────┴─────────┴─────────┴─────────┴─────────┴────────────┘  │   │
│  │  ┌─────────────────────────────────────────────────────────────┐  │   │
│  │  │                   PLAYHEAD / TIME RULER                      │  │   │
│  │  └─────────────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     PROPERTY PANEL / TOOLS                          │   │
│  │              (Effects, Transitions, Export Settings)                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 2. State Management Architecture

### Global Store (Zustand)
```typescript
// stores/timelineStore.ts
interface TimelineState {
  // Timeline data
  tracks: Track[];
  clips: Clip[];
  duration: number;
  currentTime: number;
  zoom: number;
  
  // Playback state
  isPlaying: boolean;
  playbackSpeed: number;
  
  // Selection state
  selectedClipId: string | null;
  selectedTrackId: string | null;
  
  // History for undo/redo
  history: TimelineSnapshot[];
  historyIndex: number;
  
  // Actions
  addClip: (clip: Clip) => void;
  removeClip: (id: string) => void;
  updateClip: (id: string, updates: Partial<Clip>) => void;
  moveClip: (id: string, trackId: string, startTime: number) => void;
  splitClip: (id: string, time: number) => void;
  setCurrentTime: (time: number) => void;
  setZoom: (zoom: number) => void;
  undo: () => void;
  redo: () => void;
}
```

### Server State (TanStack Query)
- Project metadata
- Media assets
- Export jobs
- User preferences

## 3. Component Hierarchy

```
TimelineEditor (page)
├── VideoPreview
│   ├── CanvasOverlay (Fabric.js)
│   └── VideoPlayer (react-player)
├── TimelineContainer
│   ├── TimeRuler
│   ├── TrackList
│   │   ├── Track (VideoTrack, AudioTrack, CaptionTrack)
│   │   │   ├── TrackHeader
│   │   │   └── TrackClips
│   │   │       └── ClipItem
│   │   └── TrackDivider (for resizing)
│   └── Playhead
├── Toolbar
│   ├── PlaybackControls
│   ├── ZoomControls
│   └── ToolSelector
└── PropertyPanel
    ├── ClipProperties
    ├── EffectControls
    └── ExportSettings
```

## 4. Key Libraries & Dependencies

### Timeline Editing
| Library | Purpose | Pros | Cons |
|---------|---------|------|------|
| @designcombo/react-timeline | Timeline UI component | Purpose-built, React-native | Limited customization |
| react-timeline-editor | Drag-drop timeline | Good for simple cases | Not maintained actively |
| Custom + Dnd-kit | Full control | Maximum flexibility | Higher dev cost |

**Decision**: Start with @designcombo/react-timeline for rapid development, evaluate custom solution if needed.

### Video Preview
| Library | Purpose | Notes |
|---------|---------|-------|
| react-player | Video playback | Wrapper for multiple players |
| Fabric.js | Canvas overlays | Text, shapes, annotations |
| HTML5 Canvas | Custom rendering | Full control |

### State Management
| Library | Purpose |
|---------|---------|
| Zustand | Global client state |
| TanStack Query | Server state |
| Immer | Immutable updates |

### Animation & UI
| Library | Purpose |
|---------|---------|
| Framer Motion | UI animations |
| Tailwind CSS | Styling |
| shadcn/ui | Component library |

## 5. Data Models

### Track
```typescript
interface Track {
  id: string;
  type: 'video' | 'audio' | 'caption';
  name: string;
  isMuted: boolean;
  isVisible: boolean;
  clips: Clip[];
  zIndex: number;
}
```

### Clip
```typescript
interface Clip {
  id: string;
  trackId: string;
  type: 'video' | 'audio' | 'caption';
  sourceId: string; // Reference to media asset
  startTime: number; // Timeline position (seconds)
  duration: number;
  sourceStart: number; // Start in source media
  sourceEnd: number; // End in source media
  effects: Effect[];
  transitions: {
    in?: Transition;
    out?: Transition;
  };
}
```

### Project
```typescript
interface Project {
  id: string;
  name: string;
  width: number;
  height: number;
  fps: number;
  duration: number;
  tracks: Track[];
  assets: MediaAsset[];
  createdAt: Date;
  updatedAt: Date;
}
```

## 6. Real-time Preview Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    PREVIEW ENGINE                          │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────┐      ┌──────────────┐                   │
│  │  Zustand     │─────▶│  Preview     │                   │
│  │  Store       │      │  Orchestrator│                   │
│  └──────────────┘      └──────┬───────┘                   │
│                               │                           │
│              ┌────────────────┼────────────────┐          │
│              ▼                ▼                ▼          │
│        ┌──────────┐     ┌──────────┐    ┌──────────┐     │
│        │  Video   │     │  Audio   │    │ Canvas   │     │
│        │  Player  │     │  Mixer   │    │ Overlay  │     │
│        └──────────┘     └──────────┘    └──────────┘     │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Preview Flow
1. User scrubs timeline → Update `currentTime` in store
2. Preview orchestrator calculates active clips at time T
3. Video player seeks to position
4. Canvas overlay renders effects/transitions
5. Audio mixer prepares volume levels

## 7. Export Integration

The timeline state feeds into the backend export pipeline:

```typescript
interface ExportJob {
  projectId: string;
  timeline: TimelineState;
  outputFormat: 'mp4' | 'webm' | 'mov';
  quality: '720p' | '1080p' | '4k';
  // References to media assets for backend processing
}
```

## 8. Performance Considerations

1. **Virtualization**: Only render visible timeline segments
2. **Debouncing**: Throttle timeline scrubbing updates
3. **Memoization**: Use React.memo for Clip components
4. **Web Workers**: Offload heavy calculations
5. **Lazy Loading**: Load media assets on demand

## 9. Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Space | Play/Pause |
| J | Previous frame |
| K | Play/Pause |
| L | Next frame |
| I | Set in point |
| O | Set out point |
| Delete | Remove selected clip |
| Ctrl+Z | Undo |
| Ctrl+Shift+Z | Redo |

## 10. Next Steps

1. [ ] Set up Next.js 15 project structure
2. [ ] Install core dependencies
3. [ ] Implement basic Timeline component
4. [ ] Create VideoPreview integration
5. [ ] Build Zustand store
6. [ ] Design API contracts with Backend team
7. [ ] Prototype drag-and-drop clips

---
**Document Owner**: Frontend Lead
**Last Updated**: 2026-02-27
**Status**: In Progress
