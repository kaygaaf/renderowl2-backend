'use client';

import React, { useState, useCallback, useRef } from 'react';
import { useDraggable } from '@dnd-kit/core';
import { CSS } from '@dnd-kit/utilities';
import { TimelineClip as TimelineClipType } from '@/types/timeline';
import { useTimelineStore, useZoom, useSelectedClips } from '@/store/timelineStore';

interface TimelineClipProps {
  clip: TimelineClipType;
  trackId: string;
  trackType: 'video' | 'audio';
}

export const TimelineClip: React.FC<TimelineClipProps> = ({ clip, trackId, trackType }) => {
  const zoom = useZoom();
  const selectedClipIds = useSelectedClips();
  const { 
    selectClip, 
    updateClip, 
    dragClip, 
    endClipDrag,
    resizeClip,
    endClipResize 
  } = useTimelineStore();
  
  const isSelected = selectedClipIds.includes(clip.id);
  
  // Resize state
  const [isResizing, setIsResizing] = useState(false);
  const [resizeSide, setResizeSide] = useState<'left' | 'right' | null>(null);
  const resizeStartData = useRef<{
    startX: number;
    originalStartTime: number;
    originalDuration: number;
  } | null>(null);

  // Drag setup for moving clips within timeline
  const { 
    attributes, 
    listeners, 
    setNodeRef, 
    transform, 
    isDragging 
  } = useDraggable({
    id: clip.id,
    data: { clip, trackId },
    disabled: isResizing,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    left: `${clip.startTime * zoom}px`,
    width: `${clip.duration * zoom}px`,
    opacity: isDragging ? 0.5 : 1,
  };

  // Handle resize start
  const handleResizeStart = useCallback((e: React.MouseEvent, side: 'left' | 'right') => {
    e.stopPropagation();
    e.preventDefault();
    
    setIsResizing(true);
    setResizeSide(side);
    resizeStartData.current = {
      startX: e.clientX,
      originalStartTime: clip.startTime,
      originalDuration: clip.duration,
    };
  }, [clip.startTime, clip.duration]);

  // Handle resize move
  const handleResizeMove = useCallback((e: MouseEvent) => {
    if (!isResizing || !resizeStartData.current) return;
    
    const { startX, originalStartTime, originalDuration } = resizeStartData.current;
    const deltaX = e.clientX - startX;
    const deltaTime = deltaX / zoom;
    
    if (resizeSide === 'right') {
      // Resize from right - adjust duration
      const newDuration = Math.max(0.5, originalDuration + deltaTime);
      resizeClip(clip.id, newDuration);
    } else if (resizeSide === 'left') {
      // Resize from left - adjust start time and duration
      const maxDelta = originalDuration - 0.5;
      const clampedDelta = Math.max(-originalStartTime, Math.min(maxDelta, deltaTime));
      const newStartTime = originalStartTime + clampedDelta;
      const newDuration = originalDuration - clampedDelta;
      resizeClip(clip.id, newDuration, newStartTime);
    }
  }, [isResizing, resizeSide, clip.id, zoom, resizeClip]);

  // Handle resize end
  const handleResizeEnd = useCallback(() => {
    if (isResizing) {
      setIsResizing(false);
      setResizeSide(null);
      resizeStartData.current = null;
      endClipResize();
    }
  }, [isResizing, endClipResize]);

  // Add/remove resize event listeners
  React.useEffect(() => {
    if (isResizing) {
      window.addEventListener('mousemove', handleResizeMove);
      window.addEventListener('mouseup', handleResizeEnd);
      return () => {
        window.removeEventListener('mousemove', handleResizeMove);
        window.removeEventListener('mouseup', handleResizeEnd);
      };
    }
  }, [isResizing, handleResizeMove, handleResizeEnd]);

  // Get clip color based on type
  const getClipColor = () => {
    switch (clip.type) {
      case 'video':
        return 'bg-blue-500/20 border-blue-500 hover:bg-blue-500/30';
      case 'audio':
        return 'bg-green-500/20 border-green-500 hover:bg-green-500/30';
      case 'image':
        return 'bg-purple-500/20 border-purple-500 hover:bg-purple-500/30';
      case 'text':
        return 'bg-yellow-500/20 border-yellow-500 hover:bg-yellow-500/30';
      default:
        return 'bg-gray-500/20 border-gray-500 hover:bg-gray-500/30';
    }
  };

  return (
    <div
      ref={setNodeRef}
      className={`
        absolute top-1 bottom-1 rounded border-2 cursor-pointer transition-all
        ${getClipColor()}
        ${isSelected ? 'ring-2 ring-primary ring-offset-1 z-10' : 'z-0'}
        ${isDragging ? 'z-50 shadow-lg' : ''}
        ${isResizing ? 'select-none' : ''}
      `}
      style={style}
      onClick={(e) => {
        e.stopPropagation();
        selectClip(clip.id, e.shiftKey);
      }}
      {...attributes}
      {...listeners}
    >
      {/* Clip content */}
      <div className="relative h-full overflow-hidden">
        {/* Left resize handle */}
        <div
          className={`
            absolute left-0 top-0 bottom-0 w-2 cursor-w-resize
            hover:bg-white/30 transition-colors flex items-center justify-center
            ${isResizing && resizeSide === 'left' ? 'bg-white/50' : ''}
          `}
          onMouseDown={(e) => handleResizeStart(e, 'left')}
        >
          <div className="w-0.5 h-4 bg-white/60 rounded" />
        </div>

        {/* Clip label */}
        <div className="px-4 py-1 text-xs truncate select-none">
          {clip.name}
        </div>

        {/* Right resize handle */}
        <div
          className={`
            absolute right-0 top-0 bottom-0 w-2 cursor-e-resize
            hover:bg-white/30 transition-colors flex items-center justify-center
            ${isResizing && resizeSide === 'right' ? 'bg-white/50' : ''}
          `}
          onMouseDown={(e) => handleResizeStart(e, 'right')}
        >
          <div className="w-0.5 h-4 bg-white/60 rounded" />
        </div>

        {/* Duration indicator */}
        <div className="absolute bottom-0.5 right-1 text-[10px] opacity-60">
          {clip.duration.toFixed(1)}s
        </div>
      </div>
    </div>
  );
};
