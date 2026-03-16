const BASE = (process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app') + '/api'

export class ApiError extends Error {
  constructor(public status: number, message: string) { super(message) }
}

async function req<T>(path: string, init: RequestInit = {}, token?: string): Promise<T> {
  const headers: Record<string,string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string,string>),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const r = await fetch(`${BASE}${path}`, { ...init, headers })
  if (!r.ok) {
    let msg = r.statusText
    try { const j = await r.json(); msg = j.error || j.message || msg } catch {}
    throw new ApiError(r.status, msg)
  }
  if (r.status === 204) return {} as T
  return r.json()
}

/* ── Auth ── */
export const authApi = {
  login:   (email: string, password: string) =>
    req<{access_token:string;refresh_token:string}>('/v1/auth/login',
      {method:'POST', body:JSON.stringify({email,password})}),
  register: (name: string, email: string, password: string) =>
    req<{access_token:string;refresh_token:string}>('/v1/auth/register',
      {method:'POST', body:JSON.stringify({name,email,password})}),
  logout: (token: string) =>
    req('/v1/auth/logout', {method:'POST'}, token),
}

/* ── Backend types (structura reală) ── */
export interface GoalSummary {
  id: string
  name: string
  status: 'ACTIVE'|'WAITING'|'PAUSED'|'COMPLETED'|'ARCHIVED'
  progress_score: number   // 0-1
  grade: string
  days_left: number
  sprint_number: number
  total_sprints: number
  start_date: string
  end_date: string
}

export interface DashboardData {
  user: { id: string; full_name: string; locale: string }
  active_goals: GoalSummary[]
  waiting_goals: GoalSummary[]
  today_tasks_count: number
}

export interface DailyTask {
  id: string
  goal_id: string
  text: string
  type: 'MAIN'|'PERSONAL'
  completed: boolean
  sort_order: number
  task_date: string
}

export interface TodayData {
  date: string
  goal_name: string
  day_number: number
  main_tasks: DailyTask[]
  personal_tasks: DailyTask[]
  done_count: number
  total_count: number
  streak_days: number
  checkpoint?: { name: string; progress_pct: number; status: string }
}

export interface Goal {
  id: string
  name: string
  description?: string
  status: 'ACTIVE'|'WAITING'|'PAUSED'|'COMPLETED'|'ARCHIVED'
  start_date: string
  end_date: string
  created_at: string
}

/* ── Endpoints (server-side - cu token) ── */
export const dashApi   = { get: (t:string) => req<DashboardData>('/v1/dashboard',{},t) }

export const goalsApi  = {
  list:   (t:string) => req<{goals:Goal[];waiting:Goal[]}>('/v1/goals',{},t),
  get:    (t:string, id:string) => req<Goal>(`/v1/goals/${id}`,{},t),
  create: (t:string, data:Record<string,unknown>) =>
    req<Goal>('/v1/goals',{method:'POST',body:JSON.stringify(data)},t),
}

export const settingsApi = {
  get:    (t:string) => req<{user_id:string;locale:string;notifications_on:boolean;sprint_reflection:boolean}>('/v1/settings',{},t),
  update: (t:string, data:{locale?:string}) =>
    req('/v1/settings',{method:'PATCH',body:JSON.stringify(data)},t),
}
