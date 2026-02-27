# Renderowl 2.0 — Frontend

Next-generation video editor built with **Next.js 15**, **TypeScript**, **Tailwind CSS**, **shadcn/ui**, **Zustand**, and **@dnd-kit**.

## 🚀 Tech Stack

- **Framework:** Next.js 15 (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS v4
- **UI Components:** shadcn/ui (zinc base)
- **State Management:** Zustand
- **Drag & Drop:** @dnd-kit
- **Icons:** Lucide React

## 📁 Project Structure

```
/projects/renderowl2.0/frontend/
├── src/
│   ├── app/                 # Next.js App Router
│   │   ├── globals.css      # Global styles + CSS variables
│   │   ├── layout.tsx       # Root layout
│   │   └── page.tsx         # Home page with timeline
│   ├── components/
│   │   └── timeline/        # Timeline components
│   │       ├── Timeline.tsx         # Main timeline container
│   │       ├── TimelineTrack.tsx    # Individual track
│   │       ├── TimelineRuler.tsx    # Time ruler/marks
│   │       ├── TimelinePlayhead.tsx # Playhead indicator
│   │       └── index.ts             # Barrel export
│   ├── store/
│   │   └── timelineStore.ts # Zustand store for timeline state
│   ├── types/
│   │   └── timeline.ts      # TypeScript interfaces
│   └── lib/
│       └── utils.ts         # Utility functions (shadcn)
├── public/                  # Static assets
├── tests/                   # Test files
├── components.json          # shadcn/ui config
├── next.config.ts           # Next.js config
├── tailwind.config.ts       # Tailwind config
├── tsconfig.json            # TypeScript config
└── package.json
```

## 🛠️ Setup Instructions

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Navigate to project directory
cd /projects/renderowl2.0/frontend

# Install dependencies
npm install

# Run development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to view the app.

## 🎯 Features

### Timeline Component

- ✅ **Multi-track support** — Video and audio tracks
- ✅ **Drag & drop** — Reorder tracks with @dnd-kit
- ✅ **Track controls** — Mute, lock, hide/show, delete
- ✅ **Playhead** — Visual indicator with red marker
- ✅ **Time ruler** — Seconds-based timeline ruler
- ✅ **Zoom control** — Adjustable timeline zoom (10-200 px/sec)
- ✅ **Clip visualization** — Color-coded clips (blue=video, green=audio)
- ✅ **Zustand state** — Centralized timeline state management

### State Management (Zustand)

The timeline store includes:

```typescript
// Tracks
addTrack(type)
removeTrack(trackId)
moveTrack(trackId, newIndex)
updateTrack(trackId, updates)

// Clips
addClip(trackId, clip)
removeClip(clipId)
moveClip(clipId, newTrackId, newStartTime)
updateClip(clipId, updates)
splitClip(clipId, splitTime)
trimClip(clipId, startOffset, endOffset)

// Selection
selectClip(clipId, multi?)
deselectClip(clipId)
clearSelection()
selectTrack(trackId)

// Playback
setCurrentTime(time)
setPlaying(playing)
setZoom(zoom)
```

## 🧪 Available Scripts

```bash
npm run dev      # Start development server
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

## 📦 Key Dependencies

```json
{
  "next": "^15.x",
  "react": "^19.x",
  "react-dom": "^19.x",
  "zustand": "^5.x",
  "@dnd-kit/core": "^6.x",
  "@dnd-kit/sortable": "^10.x",
  "@dnd-kit/utilities": "^3.x",
  "lucide-react": "^0.x",
  "tailwindcss": "^4.x",
  "shadcn/ui": "latest"
}
```

## 🎨 UI Components

The project uses shadcn/ui components. To add more:

```bash
npx shadcn add button
npx shadcn add slider
npx shadcn add dropdown-menu
# etc.
```

## 🚧 Next Steps

- [ ] Add clip import from file system
- [ ] Implement actual video/audio playback
- [ ] Add export/render functionality
- [ ] Implement undo/redo history
- [ ] Add keyboard shortcuts
- [ ] Create clip trimming UI
- [ ] Add transitions and effects
- [ ] Implement preview scrubbing

## 📝 Notes

- v1.x is **FROZEN** — all new development on v2.0
- This is a fresh start with modern tooling
- Tailwind v4 uses CSS-based configuration (no tailwind.config.js needed)

---

Built with ❤️ for Renderowl 2.0
