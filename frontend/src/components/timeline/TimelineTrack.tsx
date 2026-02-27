'use client';

import React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { useDroppable } from '@dnd-kit/core';
import { TimelineTrack as TimelineTrackType } from '@/types/timeline';
import { useTimelineStore, useZoom } from '@/store/timelineStore';
import { TimelineClip } from './TimelineClip';
import { GripVertical, Eye, EyeOff, Lock, Unlock, Volume2, VolumeX, Trash2 } from 'lucide-react';

interface TimelineTrackProps {
  track: TimelineTrackType;
}

export const TimelineTrack: React.FC<TimelineTrackProps> = ({ track }) => {
  const zoom = useZoom();
  const {
    attributes,
    listeners,
    setNodeRef: setSortableRef,
    transform,
    transition,
    isDragging: isTrackDragging,
  } = useSortable({ id: track.id });

  // Droppable for clips
  const { setNodeRef: setDroppableRef, isOver } = useDroppable({
    id: `track-${track.id}`,
    data: { trackId: track.id },
  });

  const { updateTrack, removeTrack, selectTrack, selectedTrackId } = useTimelineStore();
  const isSelected = selectedTrackId === track.id;

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isTrackDragging ? 0.5 : 1,
  };

  const trackHeight = track.type === 'video' ? 80 : 60;

  // Combine refs for sortable and droppable
  const setRefs = (el: HTMLElement | null) => {
    setSortableRef(el);
    setDroppableRef(el);
  };

  return (
    <div
      ref={setRefs}
      style={style}
      className={`
        flex border-b last:border-b-0 transition-colors
        ${isSelected ? 'bg-primary/5' : 'bg-background'}
        ${isTrackDragging ? 'z-50' : ''}
        ${isOver ? 'bg-primary/10' : ''}
      `}
    >
      {/* Track header */}
      <div
        className={`
          w-48 flex-shrink-0 border-r p-2 flex flex-col gap-1
          ${isSelected ? 'bg-primary/10' : 'bg-muted/30'}
        `}
        onClick={() => selectTrack(track.id)}
      >
        <div className="flex items-center gap-1">
          <button
            {...attributes}
            {...listeners}
            className="p-1 hover:bg-muted rounded cursor-grab active:cursor-grabbing"
          >
            <GripVertical className="w-4 h-4 text-muted-foreground" />
          </button>
          
          <span className="text-sm font-medium truncate flex-1">
            {track.name}
          </span>
          
          <button
            onClick={(e) => {
              e.stopPropagation();
              removeTrack(track.id);
            }}
            className="p-1 hover:bg-destructive/10 hover:text-destructive rounded opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <Trash2 className="w-3 h-3" />
          </button>
        </div>

        <div className="flex items-center gap-1">
          <button
            onClick={(e) => {
              e.stopPropagation();
              updateTrack(track.id, { isVisible: !track.isVisible });
            }}
            className={`p-1 rounded ${track.isVisible ? 'text-primary' : 'text-muted-foreground'}`}
            title={track.isVisible ? 'Hide track' : 'Show track'}
          >
            {track.isVisible ? <Eye className="w-3 h-3" /> : <EyeOff className="w-3 h-3" />}
          </button>

          <button
            onClick={(e) => {
              e.stopPropagation();
              updateTrack(track.id, { isLocked: !track.isLocked });
            }}
            className={`p-1 rounded ${track.isLocked ? 'text-primary' : 'text-muted-foreground'}`}
            title={track.isLocked ? 'Unlock track' : 'Lock track'}
          >
            {track.isLocked ? <Lock className="w-3 h-3" /> : <Unlock className="w-3 h-3" />}
          </button>

          <button
            onClick={(e) => {
              e.stopPropagation();
              updateTrack(track.id, { isMuted: !track.isMuted });
            }}
            className={`p-1 rounded ${track.isMuted ? 'text-muted-foreground' : 'text-primary'}`}
            title={track.isMuted ? 'Unmute track' : 'Mute track'}
          >
            {track.isMuted ? <VolumeX className="w-3 h-3" /> : <Volume2 className="w-3 h-3" />}
          </button>

          <span className="text-xs text-muted-foreground ml-auto uppercase">
            {track.type}
          </span>
        </div>
      </div>

      {/* Track clips area */}
      <div
        className="flex-1 relative group"
        style={{ height: trackHeight }}
        onClick={(e) => {
          e.stopPropagation();
          selectTrack(track.id);
        }}
      >
        {/* Grid lines */}
        <div className="absolute inset-0 pointer-events-none">
          {Array.from({ length: 100 }).map((_, i) => (
            <div
              key={i}
              className="absolute top-0 bottom-0 border-l border-border/30"
              style={{ left: `${i * zoom * 5}px` }}
            />
          ))}
        </div>

        {/* Clips */}
        {track.clips.map((clip) => (
          <TimelineClip
            key={clip.id}
            clip={clip}
            trackId={track.id}
            trackType={track.type}
          />
        ))}

        {track.clips.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center text-muted-foreground text-xs">
            Drop clips here
          </div>
        )}

        {/* Drop indicator */}
        {isOver && (
          <div className="absolute inset-0 border-2 border-dashed border-primary/50 bg-primary/5 rounded pointer-events-none" />
        )}
      </div>
    </div>
  );
};
