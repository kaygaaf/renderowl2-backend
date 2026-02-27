# Renderowl 2.0 - Frontend/Backend API Contracts

## Overview
This document defines the API contracts between the Frontend (Next.js 15) and Backend teams for Renderowl 2.0.

## Base URL
```
Production: https://api.renderowl.com/v1
Development: http://localhost:8000/v1
```

## Authentication
All API requests require a Bearer token:
```
Authorization: Bearer <token>
```

---

## Projects API

### List Projects
```http
GET /projects
```

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "proj_123",
      "name": "My Video Project",
      "width": 1920,
      "height": 1080,
      "fps": 30,
      "duration": 120.5,
      "thumbnailUrl": "https://cdn.renderowl.com/thumbs/proj_123.jpg",
      "createdAt": "2026-02-27T10:00:00Z",
      "updatedAt": "2026-02-27T15:30:00Z"
    }
  ],
  "meta": {
    "total": 42,
    "page": 1,
    "perPage": 20
  }
}
```

### Create Project
```http
POST /projects
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "New Project",
  "width": 1920,
  "height": 1080,
  "fps": 30
}
```

**Response** (201 Created):
```json
{
  "id": "proj_456",
  "name": "New Project",
  "width": 1920,
  "height": 1080,
  "fps": 30,
  "duration": 0,
  "createdAt": "2026-02-27T19:00:00Z",
  "updatedAt": "2026-02-27T19:00:00Z"
}
```

### Get Project with Timeline
```http
GET /projects/:id
```

**Response** (200 OK):
```json
{
  "id": "proj_123",
  "name": "My Video Project",
  "width": 1920,
  "height": 1080,
  "fps": 30,
  "duration": 120.5,
  "tracks": [
    {
      "id": "track_1",
      "type": "video",
      "name": "Video Track 1",
      "zIndex": 1,
      "clips": [
        {
          "id": "clip_1",
          "trackId": "track_1",
          "type": "video",
          "sourceId": "asset_123",
          "startTime": 0,
          "duration": 10.5,
          "sourceStart": 0,
          "sourceEnd": 10.5,
          "effects": [],
          "transitions": {}
        }
      ]
    }
  ],
  "assets": [
    {
      "id": "asset_123",
      "type": "video",
      "name": "intro.mp4",
      "url": "https://cdn.renderowl.com/assets/asset_123.mp4",
      "duration": 15.0,
      "width": 1920,
      "height": 1080
    }
  ]
}
```

### Update Project Timeline
```http
PATCH /projects/:id/timeline
Content-Type: application/json
```

**Request Body**:
```json
{
  "tracks": [
    {
      "id": "track_1",
      "type": "video",
      "name": "Video Track 1",
      "clips": [
        {
          "id": "clip_1",
          "trackId": "track_1",
          "type": "video",
          "sourceId": "asset_123",
          "startTime": 0,
          "duration": 10.5,
          "sourceStart": 0,
          "sourceEnd": 10.5
        }
      ]
    }
  ]
}
```

**Response** (200 OK):
```json
{
  "id": "proj_123",
  "updatedAt": "2026-02-27T19:05:00Z",
  "tracks": [...]
}
```

---

## Media Assets API

### Upload Asset (Presigned URL)
```http
POST /assets/upload-url
Content-Type: application/json
```

**Request Body**:
```json
{
  "filename": "intro.mp4",
  "contentType": "video/mp4",
  "size": 52428800
}
```

**Response** (200 OK):
```json
{
  "uploadUrl": "https://storage.renderowl.com/upload/...",
  "assetId": "asset_456",
  "expiresAt": "2026-02-27T19:15:00Z"
}
```

### Confirm Upload
```http
POST /assets/:id/confirm
```

**Response** (200 OK):
```json
{
  "id": "asset_456",
  "status": "processing"
}
```

### Get Asset
```http
GET /assets/:id
```

**Response** (200 OK):
```json
{
  "id": "asset_456",
  "type": "video",
  "name": "intro.mp4",
  "url": "https://cdn.renderowl.com/assets/asset_456.mp4",
  "thumbnailUrl": "https://cdn.renderowl.com/thumbs/asset_456.jpg",
  "duration": 15.0,
  "width": 1920,
  "height": 1080,
  "status": "ready",
  "createdAt": "2026-02-27T19:00:00Z"
}
```

---

## Export API

### Create Export Job
```http
POST /exports
Content-Type: application/json
```

**Request Body**:
```json
{
  "projectId": "proj_123",
  "format": "mp4",
  "quality": "1080p",
  "settings": {
    "codec": "h264",
    "bitrate": "5000k"
  }
}
```

**Response** (202 Accepted):
```json
{
  "id": "export_789",
  "projectId": "proj_123",
  "status": "queued",
  "progress": 0,
  "createdAt": "2026-02-27T19:10:00Z"
}
```

### Get Export Status
```http
GET /exports/:id
```

**Response** (200 OK):
```json
{
  "id": "export_789",
  "projectId": "proj_123",
  "status": "processing",
  "progress": 45,
  "downloadUrl": null,
  "expiresAt": null,
  "createdAt": "2026-02-27T19:10:00Z",
  "updatedAt": "2026-02-27T19:12:00Z"
}
```

**Status Values**:
- `queued` - Waiting to process
- `processing` - Currently rendering
- `completed` - Ready for download
- `failed` - Error occurred

### List Exports
```http
GET /projects/:id/exports
```

---

## AI Generation API

### Generate Scene
```http
POST /ai/generate-scene
Content-Type: application/json
```

**Request Body**:
```json
{
  "projectId": "proj_123",
  "script": "A beautiful sunset over mountains",
  "options": {
    "style": "cinematic",
    "duration": 5.0,
    "aspectRatio": "16:9"
  }
}
```

**Response** (202 Accepted):
```json
{
  "id": "gen_abc",
  "status": "processing",
  "estimatedDuration": 30
}
```

### Generate Captions
```http
POST /ai/generate-captions
Content-Type: application/json
```

**Request Body**:
```json
{
  "projectId": "proj_123",
  "assetId": "asset_123",
  "options": {
    "language": "en",
    "style": "subtitle"
  }
}
```

---

## WebSocket Events (Real-time)

### Connection
```
wss://api.renderowl.com/ws?token=<token>
```

### Events from Server

#### Export Progress
```json
{
  "type": "export.progress",
  "data": {
    "exportId": "export_789",
    "progress": 67,
    "status": "processing"
  }
}
```

#### Asset Processing Complete
```json
{
  "type": "asset.ready",
  "data": {
    "assetId": "asset_456",
    "thumbnailUrl": "https://cdn.renderowl.com/thumbs/asset_456.jpg"
  }
}
```

#### AI Generation Complete
```json
{
  "type": "ai.complete",
  "data": {
    "generationId": "gen_abc",
    "assetId": "asset_new",
    "url": "https://cdn.renderowl.com/assets/asset_new.mp4"
  }
}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "name",
        "message": "Name is required"
      }
    ]
  }
}
```

**Error Codes**:
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid/missing token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (e.g., duplicate name)
- `422` - Unprocessable Entity
- `429` - Rate Limited
- `500` - Internal Server Error

---

## TypeScript Types (Frontend Reference)

```typescript
// types/api.ts

export interface Project {
  id: string;
  name: string;
  width: number;
  height: number;
  fps: number;
  duration: number;
  tracks: Track[];
  assets: Asset[];
  createdAt: string;
  updatedAt: string;
}

export interface Track {
  id: string;
  type: 'video' | 'audio' | 'caption';
  name: string;
  zIndex: number;
  clips: Clip[];
}

export interface Clip {
  id: string;
  trackId: string;
  type: 'video' | 'audio' | 'caption';
  sourceId: string;
  startTime: number;
  duration: number;
  sourceStart: number;
  sourceEnd: number;
  effects: Effect[];
  transitions: {
    in?: Transition;
    out?: Transition;
  };
}

export interface Asset {
  id: string;
  type: 'video' | 'audio' | 'image';
  name: string;
  url: string;
  thumbnailUrl?: string;
  duration?: number;
  width?: number;
  height?: number;
  status: 'uploading' | 'processing' | 'ready' | 'error';
}

export interface ExportJob {
  id: string;
  projectId: string;
  status: 'queued' | 'processing' | 'completed' | 'failed';
  progress: number;
  downloadUrl?: string;
  createdAt: string;
}
```

---

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| General API | 100 req/min |
| Asset Upload | 10 req/min |
| AI Generation | 5 req/min |

---

**Status**: Draft - Pending Backend Review  
**Last Updated**: 2026-02-27  
**Frontend Contact**: Frontend Lead  
**Backend Contact**: TBD
