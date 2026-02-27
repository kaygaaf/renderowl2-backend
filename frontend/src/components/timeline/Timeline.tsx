'use client';

import React from 'react';
import {
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  closestCenter,
} from '@dnd-kit/core';
import {
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { useTimelineStore, useTracks } from '@/store/timelineStore';
import { TimelineTrack } from './TimelineTrack';
import { TimelineRuler } from './TimelineRuler';
import { TimelinePlayhead } from './TimelinePlayhead';

export interface TimelineProps {
  className?: string;
}

export const Timeline: React.FC<TimelineProps> = ({ className = '' }) => {
  const tracks = useTracks();
  const moveTrack = useTimelineStore((state) => state.moveTrack);
  const moveClip = useTimelineStore((state) => state.moveClip);
  const setCurrentTime = useTimelineStore((state) => state.setCurrentTime);
  const [activeId, setActiveId] = React.useState<string | null>(null);

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveId(null);

    if (!over) return;

    // Handle track reordering
    if (active.id !== over.id) {
      const oldIndex = tracks.findIndex((t) => t.id === active.id);
      const newIndex = tracks.findIndex((t) => t.id === over.id);
      
      if (oldIndex !== -1 && newIndex !== -1) {
        moveTrack(active.id as string, newIndex);
      }
    }
  };

  const handleTimelineClick = (event: React.MouseEvent<HTMLDivElement>) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    // TODO: Convert pixels to time based on zoom
    // setCurrentTime(x / zoom);
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
        </div>
        
        <div className="flex items-center gap-2">
          <button
            onClick={() => useTimelineStore.getState().addTrack('video')}
            className="px-2 py-1 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90"
          >
            + Video Track
          </button>
          <button
            onClick={() => useTimelineStore.getState().addTrack('audio')}
            className="px-2 py-1 text-xs bg-secondary text-secondary-foreground rounded hover:bg-secondary/90"
          >
            + Audio Track
          </button>
        </div>
      </div>

      {/* Timeline ruler */}
      <TimelineRuler />

      {/* Tracks area */}
      <DndContext
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <div className="relative flex-1 overflow-auto">
          <SortableContext
            items={tracks.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            <div className="relative min-h-[200px]" onClick={handleTimelineClick}>
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

          <DragOverlay>
            {activeId ? (
              <div className="bg-primary/20 border-2 border-primary rounded p-2">
                Moving...
              </div>
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
          Use drag & drop to reorder tracks
        </div>
      </div>
    </div>
  );
};
