'use client';

import React from 'react';
import {
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
  closestCenter,
  pointerWithin,
} from '@dnd-kit/core';
import {
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { useTimelineStore, useTracks, useZoom } from '@/store/timelineStore';
import { TimelineTrack } from './TimelineTrack';
import { TimelineRuler } from './TimelineRuler';
import { TimelinePlayhead } from './TimelinePlayhead';
import { TimelineClip as TimelineClipComponent } from './TimelineClip';
import { TimelineClip } from '@/types/timeline';

export interface TimelineProps {
  className?: string;
}

export const Timeline: React.FC<TimelineProps> = ({ className = '' }) => {
  const tracks = useTracks();
  const zoom = useZoom();
  const { 
    moveTrack, 
    dragClip, 
    endClipDrag,
    addTrack,
    isLoading,
    isSaving,
    error,
    clearError,
    timelineId,
  } = useTimelineStore();
  
  const [activeId, setActiveId] = React.useState<string | null>(null);
  const [activeClip, setActiveClip] = React.useState<TimelineClip | null>(null);
  const [dragOverlayPos, setDragOverlayPos] = React.useState<{ x: number; y: number } | null>(null);

  // Configure sensors for drag detection
  const sensors = useSensors(
    useSensor(MouseSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 250,
        tolerance: 5,
      },
    })
  );

  const handleDragStart = (event: DragStartEvent) => {
    const { active } = event;
    setActiveId(active.id as string);
    
    // Check if dragging a clip
    if (active.data.current?.clip) {
      setActiveClip(active.data.current.clip);
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over, delta } = event;
    setActiveId(null);
    setActiveClip(null);

    if (!over) return;

    // Handle track reordering
    const oldTrackIndex = tracks.findIndex((t) => t.id === active.id);
    const newTrackIndex = tracks.findIndex((t) => t.id === over.id);
    
    if (oldTrackIndex !== -1 && newTrackIndex !== -1 && active.id !== over.id) {
      moveTrack(active.id as string, newTrackIndex);
      return;
    }

    // Handle clip drag between tracks
    if (active.data.current?.clip && over.id) {
      const clip = active.data.current.clip as TimelineClip;
      const sourceTrackId = active.data.current.trackId as string;
      
      // Determine target track
      let targetTrackId: string;
      
      if (over.id.toString().startsWith('track-')) {
        targetTrackId = over.id.toString().replace('track-', '');
      } else {
        // Find track by clip ID
        const track = tracks.find(t => t.clips.some(c => c.id === over.id));
        if (track) {
          targetTrackId = track.id;
        } else {
          return;
        }
      }

      // Calculate new start time based on drag delta
      const deltaTime = delta.x / zoom;
      const newStartTime = Math.max(0, clip.startTime + deltaTime);
      
      // Move clip
      dragClip(clip.id, targetTrackId, newStartTime);
      
      // Save to backend
      await endClipDrag();
    }
  };

  // Calculate grid position for drag overlay
  const handleDragMove = (event: DragEndEvent) => {
    if (activeClip) {
      // Snap to grid (0.5 second increments)
      // This is just visual feedback
    }
  };

  // Status indicator component
  const StatusIndicator = () => {
    if (isLoading) {
      return (
        <span className="flex items-center gap-1 text-xs text-muted-foreground">
          <svg className="animate-spin h-3 w-3" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          Loading...
        </span>
      );
    }
    
    if (isSaving) {
      return (
        <span className="flex items-center gap-1 text-xs text-muted-foreground">
          <svg className="animate-spin h-3 w-3" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          Saving...
        </span>
      );
    }
    
    return null;
  };

  return (
    <div className={`flex flex-col bg-background border rounded-lg overflow-hidden ${className}`}>
      {/* Header with controls */}
      <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/50">
        <div className="flex items-center gap-2">
          <h2 className="text-sm font-semibold">Timeline</h2>
          <span className="text-xs text-muted-foreground">
            {tracks.length} tracks
          </span>
          <StatusIndicator />
        </div>
        
        <div className="flex items-center gap-2">
          <button
            onClick={() => addTrack('video')}
            className="px-2 py-1 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90"
          >
            + Video Track
          </button>
          <button
            onClick={() => addTrack('audio')}
            className="px-2 py-1 text-xs bg-secondary text-secondary-foreground rounded hover:bg-secondary/90"
          >
            + Audio Track
          </button>
        </div>
      </div>

      {/* Error banner */}
      {error && (
        <div className="px-4 py-2 bg-destructive/10 border-b border-destructive/20 flex items-center justify-between">
          <span className="text-xs text-destructive">{error}</span>
          <button
            onClick={clearError}
            className="text-xs text-destructive hover:underline"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Timeline ruler */}
      <TimelineRuler />

      {/* Tracks area */}
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragMove={handleDragMove}
      >
        <div className="relative flex-1 overflow-auto">
          <SortableContext
            items={tracks.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            <div className="relative min-h-[200px]">
              {/* Playhead */}
              <TimelinePlayhead />

              {/* Tracks */}
              {tracks.map((track) => (
                <TimelineTrack key={track.id} track={track} />
              ))}

              {tracks.length === 0 && (
                <div className="flex items-center justify-center h-32 text-muted-foreground">
                  <p>No tracks yet. Add a track to get started.</p>
                </div>
              )}
            </div>
          </SortableContext>

          <DragOverlay dropAnimation={null}>
            {activeId ? (
              activeClip ? (
                <div
                  className={`
                    rounded border-2 border-dashed border-primary/50 bg-primary/20
                    flex items-center justify-center
                  `}
                  style={{
                    width: `${activeClip.duration * zoom}px`,
                    height: '40px',
                  }}
                >
                  <span className="text-xs text-primary">{activeClip.name}</span>
                </div>
              ) : (
                <div className="bg-primary/20 border-2 border-primary rounded p-2">
                  Moving Track...
                </div>
              )
            ) : null}
          </DragOverlay>
        </div>
      </DndContext>

      {/* Footer with zoom controls */}
      <div className="flex items-center justify-between px-4 py-2 border-t bg-muted/50">
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground">Zoom:</span>
          <input
            type="range"
            min="10"
            max="200"
            defaultValue="50"
            onChange={(e) => useTimelineStore.getState().setZoom(Number(e.target.value))}
            className="w-24"
          />
        </div>
        
        <div className="text-xs text-muted-foreground">
          Drag clips to move • Drag edges to resize
        </div>
      </div>
    </div>
  );
};
