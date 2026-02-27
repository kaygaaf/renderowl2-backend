'use client';

import React from 'react';
import { useZoom, useCurrentTime } from '@/store/timelineStore';

export const TimelineRuler: React.FC = () => {
  const zoom = useZoom();
  const currentTime = useCurrentTime();

  // Generate time markers every 5 seconds
  const markers = React.useMemo(() => {
    const totalDuration = 120; // 2 minutes for now
    const markerInterval = 5; // Every 5 seconds
    const markers = [];
    
    for (let time = 0; time <= totalDuration; time += markerInterval) {
      const isMajor = time % 10 === 0;
      markers.push({
        time,
        isMajor,
        left: time * zoom,
      });
    }
    
    return markers;
  }, [zoom]);

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="h-8 border-b bg-muted/30 relative overflow-hidden">
      <div className="absolute inset-0 flex items-end">
        {markers.map((marker) => (
          <div
            key={marker.time}
            className="absolute bottom-0 flex flex-col items-center"
            style={{ left: `${marker.left + 192}px` }} // 192px = width of track header
          >
            <div
              className={`w-px bg-border ${
                marker.isMajor ? 'h-4 bg-foreground' : 'h-2'
              }`}
            />
            {marker.isMajor && (
              <span className="text-xs text-muted-foreground mt-1">
                {formatTime(marker.time)}
              </span>
            )}
          </div>
        ))}
      </div>

      {/* Current time indicator */}
      <div
        className="absolute top-0 bottom-0 w-px bg-primary z-10 pointer-events-none"
        style={{ left: `${192 + currentTime * zoom}px` }}
      >
        <div className="absolute -top-1 -translate-x-1/2 bg-primary text-primary-foreground text-xs px-1 rounded">
          {formatTime(currentTime)}
        </div>
      </div>
    </div>
  );
};
