import { Timeline } from '@/components/timeline';

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">Renderowl 2.0</h1>
              <p className="text-sm text-muted-foreground">
                Next-gen video editor
              </p>
            </div>
            
            <div className="flex items-center gap-2">
              <span className="text-xs px-2 py-1 bg-primary/10 text-primary rounded">
                v2.0.0-alpha
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="container mx-auto px-4 py-6">
        <div className="space-y-6">
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
      <footer className="border-t mt-auto">
        <div className="container mx-auto px-4 py-4">
          <p className="text-sm text-muted-foreground text-center">
            Renderowl 2.0 — Built with Next.js 15, Tailwind CSS, shadcn/ui, Zustand & @dnd-kit
          </p>
        </div>
      </footer>
    </div>
  );
}
