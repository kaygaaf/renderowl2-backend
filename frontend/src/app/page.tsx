'use client';

import { useEffect } from 'react';
import { Timeline } from '@/components/timeline';
import { useTimelineStore, useTimelineLoading, useTimelineSaving, useTimelineError, useLastSaved } from '@/store/timelineStore';

export default function Home() {
  const { 
    loadTimeline, 
    createTimeline, 
    saveTimeline, 
    timelineId, 
    clearError,
    addClip,
  } = useTimelineStore();
  
  const isLoading = useTimelineLoading();
  const isSaving = useTimelineSaving();
  const error = useTimelineError();
  const lastSaved = useLastSaved();

  // Load timeline on mount (demo: create new if none exists)
  useEffect(() => {
    const initTimeline = async () => {
      // For demo purposes, create a new timeline if none loaded
      if (!timelineId) {
        // Add some demo clips
        setTimeout(() => {
          addClip('track-1', {
            name: 'Intro Video',
            startTime: 0,
            duration: 5,
            trackId: 'track-1',
            type: 'video',
          });
          addClip('track-1', {
            name: 'Main Content',
            startTime: 5.5,
            duration: 10,
            trackId: 'track-1',
            type: 'video',
          });
          addClip('track-2', {
            name: 'Background Music',
            startTime: 0,
            duration: 15.5,
            trackId: 'track-2',
            type: 'audio',
          });
        }, 100);
      }
    };
    
    initTimeline();
  }, []);

  const handleCreateTimeline = async () => {
    await createTimeline('New Project', 'Created from frontend');
  };

  const handleLoadTimeline = async () => {
    // For demo, load timeline with ID 1
    await loadTimeline(1);
  };

  const handleSaveTimeline = async () => {
    await saveTimeline();
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div>
                <h1 className="text-2xl font-bold">Renderowl 2.0</h1>
                <p className="text-sm text-muted-foreground">
                  Next-gen video editor
                </p>
              </div>
              
              <div className="flex items-center gap-2">
                <button
                  onClick={handleCreateTimeline}
                  disabled={isLoading}
                  className="px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
                >
                  New Timeline
                </button>
                <button
                  onClick={handleLoadTimeline}
                  disabled={isLoading}
                  className="px-3 py-1.5 text-xs bg-secondary text-secondary-foreground rounded hover:bg-secondary/90 disabled:opacity-50"
                >
                  Load Timeline
                </button>
                <button
                  onClick={handleSaveTimeline}
                  disabled={isSaving || !timelineId}
                  className="px-3 py-1.5 text-xs bg-accent text-accent-foreground rounded hover:bg-accent/90 disabled:opacity-50"
                >
                  {isSaving ? 'Saving...' : 'Save Timeline'}
                </button>
              </div>
            </div>
            
            <div className="flex items-center gap-2">
              {lastSaved && (
                <span className="text-xs text-muted-foreground">
                  Saved {lastSaved.toLocaleTimeString()}
                </span>
              )}
              <span className="text-xs px-2 py-1 bg-primary/10 text-primary rounded">
                v2.0.0-alpha
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="flex-1 container mx-auto px-4 py-6">
        <div className="space-y-6">
          {/* Status banner */}
          {error && (
            <div className="px-4 py-3 bg-destructive/10 border border-destructive/20 rounded-lg flex items-center justify-between">
              <span className="text-sm text-destructive">{error}</span>
              <button
                onClick={clearError}
                className="text-sm text-destructive hover:underline"
              >
                Dismiss
              </button>
            </div>
          )}

          {/* Preview area placeholder */}
          <div className="aspect-video bg-muted rounded-lg flex items-center justify-center border-2 border-dashed">
            <div className="text-center">
              <p className="text-lg font-medium text-muted-foreground">
                Video Preview
              </p>
              <p className="text-sm text-muted-foreground">
                Load a project to see the preview
              </p>
            </div>
          </div>

          {/* Timeline */}
          <Timeline className="h-[400px]" />
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t">
        <div className="container mx-auto px-4 py-4">
          <p className="text-sm text-muted-foreground text-center">
            Renderowl 2.0 — Connected to Backend API • Built with Next.js 15, Tailwind CSS, shadcn/ui, Zustand & @dnd-kit
          </p>
        </div>
      </footer>
    </div>
  );
}
