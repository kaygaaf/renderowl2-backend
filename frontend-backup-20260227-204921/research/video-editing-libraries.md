# React Video Editing Libraries Research

## Executive Summary

For Renderowl 2.0's timeline editor, I evaluated multiple React-based video editing libraries. Based on our requirements for drag-and-drop timeline editing, real-time preview, and multi-track support, here are my findings:

## Recommended Approach

### Primary: @designcombo/react-timeline
**Best fit for rapid development with good customization options.**

- **GitHub**: designcombo/react-timeline
- **License**: MIT
- **Stars**: ~500+

**Pros**:
- Purpose-built for React
- Supports multiple track types
- Drag and drop built-in
- TypeScript support
- Active development

**Cons**:
- Smaller community than generic DnD libraries
- Some styling customization required

**Installation**:
```bash
npm install @designcombo/react-timeline
```

**Basic Usage**:
```tsx
import { Timeline, Track, Clip } from '@designcombo/react-timeline';

<Timeline
  tracks={tracks}
  onClipMove={handleClipMove}
  onClipResize={handleClipResize}
  zoom={zoomLevel}
  currentTime={currentTime}
/>
```

---

## Alternative Options

### Option A: react-timeline-editor
- **Status**: ⚠️ Not actively maintained
- **Best for**: Simple use cases only
- **Verdict**: Skip - maintenance risk

### Option B: Dnd-kit + Custom Timeline
- **Status**: ✅ Actively maintained
- **Best for**: Maximum customization
- **Verdict**: Good fallback if @designcombo doesn't meet needs

**Key libraries**:
```bash
npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

### Option C: Remotion (Video Rendering)
- **Status**: ✅ Actively maintained, strong community
- **Best for**: Programmatic video generation
- **Verdict**: Use for export/preview, not timeline UI

**Note**: Remotion is excellent for rendering but doesn't provide a timeline editor UI out of the box.

---

## Supporting Libraries

### Video Playback
| Library | Purpose | Recommendation |
|---------|---------|----------------|
| react-player | Multi-format video player | ✅ Use |
| video.js | Advanced controls | Consider for advanced features |
| hls.js | HLS streaming | Future consideration |

### Canvas/Overlay
| Library | Purpose | Recommendation |
|---------|---------|----------------|
| fabric.js | Canvas manipulation | ✅ Use for overlays |
| konva.js | React canvas | Alternative to Fabric |
| react-konva | Konva React wrapper | If using Konva |

### State Management
| Library | Purpose | Recommendation |
|---------|---------|----------------|
| Zustand | Global state | ✅ Use |
| TanStack Query | Server state | ✅ Use |
| Immer | Immutable updates | ✅ Use |

---

## Implementation Strategy

### Phase 1: MVP Timeline (Week 1-2)
1. Set up @designcombo/react-timeline
2. Implement basic track rendering
3. Add clip drag-and-drop
4. Connect to Zustand store

### Phase 2: Preview Integration (Week 3)
1. Integrate react-player for video preview
2. Sync playhead with timeline
3. Add canvas overlay with Fabric.js

### Phase 3: Advanced Features (Week 4+)
1. Multi-track support
2. Transitions and effects
3. Zoom and pan controls
4. Keyboard shortcuts

---

## Dependencies to Install

```bash
# Core timeline
npm install @designcombo/react-timeline

# Video playback
npm install react-player

# Canvas overlays
npm install fabric
npm install @types/fabric --save-dev

# State management
npm install zustand immer
npm install @tanstack/react-query

# DnD (fallback/custom)
npm install @dnd-kit/core @dnd-kit/sortable

# Animation
npm install framer-motion

# Utilities
npm install date-fns uuid
npm install @types/uuid --save-dev
```

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| @designcombo limitations | Medium | Keep Dnd-kit as backup plan |
| Fabric.js bundle size | Low | Lazy load canvas component |
| Real-time sync issues | High | Throttle updates, use refs |
| Browser compatibility | Low | Test on target browsers early |

---

## Recommendation

**Go with @designcombo/react-timeline as the primary timeline component.**

If we hit limitations during implementation, we'll evaluate:
1. Forking and customizing @designcombo
2. Building custom with Dnd-kit
3. Hybrid approach

---

**Research Date**: 2026-02-27  
**Researcher**: Frontend Lead  
**Status**: Complete ✅
