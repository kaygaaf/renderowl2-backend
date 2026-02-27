'use client';

import React from 'react';
import { useCurrentTime, useZoom } from '@/store/timelineStore';

export const TimelinePlayhead: React.FC = () => {
  const currentTime = useCurrentTime();
  const zoom = useZoom();

  return (
    <div
      className="absolute top-0 bottom-0 w-px bg-red-500 z-20 pointer-events-none"
      style={{ left: `${192 + currentTime * zoom}px` }} // 192px = width of track header
    >
      {/* Playhead triangle marker */}
      <div className="absolute -top-2 -translate-x-1/2">
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M6 0L12 10H0L6 0Z"
            fill="#ef4444"
          />
        </svg>
      </div>

      {/* Playhead line */}
      <div className="absolute top-0 bottom-0 w-0.5 bg-red-500 shadow-lg shadow-red-500/50" />
    </div>
  );
};
