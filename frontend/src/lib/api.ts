import { TimelineTrack, TimelineClip } from '@/types/timeline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface ApiError {
  message: string;
  status?: number;
}

export interface TimelineData {
  id?: number;
  title: string;
  description?: string;
  tracks: TimelineTrack[];
  totalDuration?: number;
}

export interface TimelineResponse {
  id: number;
  user_id: number;
  title: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
  track_count?: number;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw {
          message: errorData.error || `HTTP ${response.status}: ${response.statusText}`,
          status: response.status,
        } as ApiError;
      }

      // Handle 204 No Content
      if (response.status === 204) {
        return undefined as T;
      }

      return await response.json() as T;
    } catch (error) {
      if ((error as ApiError).message) {
        throw error;
      }
      throw {
        message: error instanceof Error ? error.message : 'Network error occurred',
      } as ApiError;
    }
  }

  // Timeline endpoints
  async createTimeline(data: { title: string; description?: string }): Promise<TimelineResponse> {
    return this.request('/timeline', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getTimeline(id: number): Promise<TimelineResponse> {
    return this.request(`/timeline/${id}`, {
      method: 'GET',
    });
  }

  async updateTimeline(id: number, data: Partial<TimelineResponse>): Promise<TimelineResponse> {
    return this.request(`/timeline/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteTimeline(id: number): Promise<void> {
    return this.request(`/timeline/${id}`, {
      method: 'DELETE',
    });
  }

  async listTimelines(limit = 20, offset = 0): Promise<TimelineResponse[]> {
    return this.request(`/timelines?limit=${limit}&offset=${offset}`, {
      method: 'GET',
    });
  }

  async getUserTimelines(): Promise<TimelineResponse[]> {
    return this.request('/timelines/me', {
      method: 'GET',
    });
  }

  // Full timeline data sync (custom endpoint for frontend state)
  async saveTimelineData(id: number, data: TimelineData): Promise<TimelineData> {
    // For now, we'll use the update endpoint
    // In a full implementation, this would save tracks and clips too
    return this.request(`/timeline/${id}/data`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }).catch(() => {
      // Fallback to regular update if data endpoint doesn't exist
      return this.updateTimeline(id, {
        title: data.title,
        description: data.description,
      }) as unknown as TimelineData;
    });
  }
}

// Export singleton instance
export const api = new ApiClient();

// Export for custom base URL if needed
export const createApiClient = (baseUrl: string) => new ApiClient(baseUrl);
