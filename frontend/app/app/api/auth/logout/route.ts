import { NextResponse } from 'next/server'
export async function POST() {
  const res = NextResponse.json({ ok:true })
  res.cookies.delete('nv_access')
  res.cookies.delete('nv_refresh')
  return res
}
