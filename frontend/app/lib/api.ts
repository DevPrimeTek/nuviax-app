const BASE = (process.env.NEXT_PUBLIC_API_URL || 'https://api.nuviax.app') + '/api'

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

/* ── Types ── */
export interface Goal {
  id: string; name: string; category: string
  target_value: number; current_value: number; unit: string
  status: 'active'|'waiting'|'completed'|'paused'
  current_sprint: number; total_sprints: number
  sprint_days_left: number; overall_score: number
  progress_pct: number; color: string
}

export interface Task {
  id: string; goal_id: string; text: string
  type: 'main'|'optional'|'personal'
  estimated_min: number; done: boolean; date: string
}

export interface DashboardData {
  greeting: string; date_label: string; streak: number
  sprint_name: string; sprint_days_left: number; sprint_pct: number
  tasks_done: number; tasks_total: number; milestone_pct: number
  mrr_current: number; active_goals: Goal[]; today_tasks: Task[]
}

export interface TodayData {
  goal_name: string; sprint_label: string
  milestone: string; milestone_pct: number
  tasks_done: number; tasks_total: number; tasks: Task[]
}

export interface RecapData {
  sprint_name: string; score: number; grade: string
  days_active: number; days_total: number; streak: number
  mrr_delta: number; next_sprint_name: string
}

/* ── Endpoints ── */
export const dashApi   = { get: (t:string) => req<DashboardData>('/v1/dashboard',{},t) }
export const todayApi  = { get: (t:string) => req<TodayData>('/v1/today',{},t),
  complete: (t:string,id:string) => req(`/v1/today/complete/${id}`,{method:'POST'},t),
  addPersonal: (t:string,text:string,min:number) =>
    req<Task>('/v1/today/personal',{method:'POST',body:JSON.stringify({text,estimated_min:min})},t),
}
export const goalsApi  = {
  list: (t:string) => req<{goals:Goal[];waiting:Goal[]}>('/v1/goals',{},t),
  get:  (t:string,id:string) => req<Goal>(`/v1/goals/${id}`,{},t),
  create: (t:string,data:Record<string,unknown>) =>
    req<Goal>('/v1/goals',{method:'POST',body:JSON.stringify(data)},t),
}
export const recapApi  = {
  get:    (t:string,gid:string) => req<RecapData>(`/v1/goals/${gid}/recap`,{},t),
  submit: (t:string,gid:string,answers:Record<string,unknown>) =>
    req(`/v1/goals/${gid}/recap`,{method:'POST',body:JSON.stringify(answers)},t),
}
export const userApi   = {
  me:          (t:string) => req<{id:string;name:string;email:string;lang:string}>('/v1/user/me',{},t),
  updatePrefs: (t:string,p:Record<string,unknown>) =>
    req('/v1/user/prefs',{method:'PATCH',body:JSON.stringify(p)},t),
}
