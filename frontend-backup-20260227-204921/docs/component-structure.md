# Renderowl 2.0 - Component Structure

## Directory Structure

```
app/
├── (editor)/                           # Route group for editor pages
│   ├── layout.tsx                      # Editor layout with sidebar
│   ├── page.tsx                        # Main timeline editor
│   ├── project/
│   │   └── [id]/
│   │       └── page.tsx                # Open existing project
│   └── new/
│       └── page.tsx                    # Create new project
├── api/                                # API routes (if needed)
├── layout.tsx                          # Root layout
└── page.tsx                            # Landing page

components/
├── timeline/                           # Timeline editor components
│   ├── TimelineContainer.tsx           # Main timeline wrapper
│   ├── TimelineProvider.tsx            # Context provider
│   ├── tracks/
│   │   ├── TrackList.tsx               # List of all tracks
│   │   ├── Track.tsx                   # Individual track
│   │   ├── TrackHeader.tsx             # Track controls
│   │   └── TrackDivider.tsx            # Resize handle
│   ├── clips/
│   │   ├── ClipItem.tsx                # Single clip component
│   │   ├── ClipResizeHandle.tsx        # Resize handles
│   │   ├── VideoClip.tsx               # Video clip variant
│   │   ├── AudioClip.tsx               # Audio clip variant
│   │   └── CaptionClip.tsx             # Caption clip variant
│   ├── ruler/
│   │   ├── TimeRuler.tsx               # Time markers
│   │   └── Playhead.tsx                # Current position indicator
│   └── controls/
│       ├── ZoomControls.tsx            # Zoom in/out
│       ├── PlaybackControls.tsx        # Play, pause, seek
│       └── TimelineToolbar.tsx         # Top toolbar
├── preview/                            # Video preview components
│   ├── VideoPreview.tsx                # Main preview container
│   ├── CanvasOverlay.tsx               # Fabric.js canvas
│   ├── VideoPlayer.tsx                 # react-player wrapper
│   └── PreviewControls.tsx             # Preview toolbar
├── properties/                         # Property panels
│   ├── PropertyPanel.tsx               # Main panel container
│   ├── ClipProperties.tsx              # Selected clip settings
│   ├── TrackProperties.tsx             # Track settings
│   ├── EffectControls.tsx              # Effects panel
│   └── ExportSettings.tsx              # Export configuration
├── media/                              # Media management
│   ├── MediaBrowser.tsx                # Asset library
│   ├── MediaUploader.tsx               # Upload component
│   └── MediaItem.tsx                   # Individual asset
├── layout/                             # Layout components
│   ├── EditorLayout.tsx                # Editor page layout
│   ├── Sidebar.tsx                     # Left sidebar
│   ├── Panel.tsx                       # Resizable panel
│   └── Splitter.tsx                    # Splitter handle
└── ui/                                 # shadcn/ui components
    ├── button.tsx
    ├── slider.tsx
    ├── tabs.tsx
    └── ...

hooks/                                  # Custom React hooks
├── useTimeline.ts                      # Timeline state hook
├── usePlayback.ts                      # Playback controls
├── useMedia.ts                         # Media management
├── useProject.ts                       # Project operations
├── useKeyboardShortcuts.ts             # Keyboard handling
└── useExport.ts                        # Export operations

stores/                                 # Zustand stores
├── timelineStore.ts                    # Timeline state
├── projectStore.ts                     # Project metadata
├── playbackStore.ts                    # Playback state
└── uiStore.ts                          # UI state (panels, etc.)

lib/                                    # Utilities and configs
├── utils.ts                            # General utilities
├── constants.ts                        # App constants
├── fabric.ts                           # Fabric.js setup
├── queryClient.ts                      # TanStack Query config
└── api.ts                              # API client

types/                                  # TypeScript types
├── timeline.ts                         # Timeline types
├── project.ts                          # Project types
├── media.ts                            # Media types
└── api.ts                              # API types

services/                               # API services
├── projectService.ts                   # Project CRUD
├── mediaService.ts                     # Media operations
├── exportService.ts                    # Export jobs
└── aiService.ts                        # AI generation
```

## Component Details

### Timeline Components

#### TimelineContainer
```typescript
interface TimelineContainerProps {
  projectId: string;
  initialTracks?: Track[];
  onChange?: (timeline: TimelineState) => void;
}
```

#### TrackList
```typescript
interface TrackListProps {
  tracks: Track[];
  selectedTrackId?: string;
  onTrackSelect: (trackId: string) => void;
  onTrackReorder: (trackIds: string[]) => void;
}
```

#### ClipItem
```typescript
interface ClipItemProps {
  clip: Clip;
  isSelected: boolean;
  zoom: number;
  onSelect: () => void;
  onMove: (newStartTime: number) => void;
  onResize: (newDuration: number) => void;
  onSplit: (time: number) => void;
}
```

### Preview Components

#### VideoPreview
```typescript
interface VideoPreviewProps {
  currentTime: number;
  isPlaying: boolean;
  tracks: Track[];
  width: number;
  height: number;
  onTimeUpdate: (time: number) => void;
}
```

#### CanvasOverlay
```typescript
interface CanvasOverlayProps {
  width: number;
  height: number;
  overlays: Overlay[];
  currentTime: number;
}
```

## State Flow

```
User Action
    │
    ▼
┌─────────────────┐
│  Component      │
│  (e.g., Clip)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Zustand Store  │
│  (timelineStore)│
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐ ┌───────────┐
│ Local │ │  Server   │
│ State │ │  (Query)  │
└───────┘ └───────────┘
```

## Component Composition Example

```tsx
// app/(editor)/page.tsx
export default function EditorPage() {
  return (
    <EditorLayout>
      <VideoPreviewPanel>
        <VideoPreview />
      </VideoPreviewPanel>
      
      <TimelinePanel>
        <TimelineToolbar />
        <TimelineContainer>
          <TimeRuler />
          <TrackList>
            {tracks.map(track => (
              <Track key={track.id}>
                <TrackHeader {...track} />
                <TrackClips>
                  {track.clips.map(clip => (
                    <ClipItem key={clip.id} {...clip} />
                  ))}
                </TrackClips>
              </Track>
            ))}
          </TrackList>
          <Playhead />
        </TimelineContainer>
      </TimelinePanel>
      
      <PropertyPanel>
        <ClipProperties />
      </PropertyPanel>
    </EditorLayout>
  );
}
```

## Export Pattern

All components use named exports:

```typescript
// components/timeline/index.ts
export { TimelineContainer } from './TimelineContainer';
export { TrackList } from './tracks/TrackList';
export { Track } from './tracks/Track';
export { ClipItem } from './clips/ClipItem';
// ...

// Usage
import { TimelineContainer, TrackList } from '@/components/timeline';
```

---

**Status**: Draft  
**Last Updated**: 2026-02-27  
**Owner**: Frontend Lead
