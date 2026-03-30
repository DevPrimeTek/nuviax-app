# docs/user-workflow.md — NuviaX User Workflow

> Version: 10.5.0 | Last updated: 2026-03-30

---

## 1. User Journey (End-to-End)

### 1.1 Registration & Onboarding
### 1.2 First Login
### 1.3 Goal Setup
### 1.4 Daily Loop
### 1.5 Progress Review
### 1.6 Growth & Leveling

---

## 2. Goal Creation Flow

### 2.1 Onboarding Goal Wizard
### 2.2 AI Category Suggestion
### 2.3 Manual Category Selection
### 2.4 Behavior Model Selection (G-11)
### 2.5 Goal Submission & Validation
### 2.6 Initial Score Calculation

---

## 3. Daily Execution Flow

### 3.1 Today Page Overview
### 3.2 Task List Rendering
### 3.3 Task Completion
### 3.4 Daily Check-in
### 3.5 Score Recalculation Trigger
### 3.6 Progress Feedback (Grade, %)

---

## 4. SRM Flow (L1–L3)

### 4.1 SRM Entry Points
### 4.2 L1 — Daily Review
### 4.3 L2 — Weekly Review
### 4.4 L3 — Monthly Review
### 4.5 SRM Submission
### 4.6 Score Impact After SRM

---

## 5. Achievement Flow

### 5.1 Achievement Trigger Conditions
### 5.2 Ceremony Tiers (Tier 1–3)
### 5.3 Badge Award & Display
### 5.4 Achievement History (/profile)

---

## 6. Visualization Flow

### 6.1 Progress Bar & Grade Display
### 6.2 Activity Heatmap (52-week)
### 6.3 Goal Progress Cards
### 6.4 Profile Stats Overview
### 6.5 Dark/Light Theme Rendering

---

## 7. Test Scenarios

### 7.1 Happy Path — New User Full Journey
### 7.2 User With No Goals
### 7.3 Missed Daily Check-in
### 7.4 AI Suggestion Timeout (Fallback)
### 7.5 SRM Completion After Missed Period
### 7.6 Achievement Unlock Edge Case
### 7.7 Theme Persistence Across Sessions
### 7.8 Language Switch (EN / RU / RO)

---

## 8. Critical Checkpoints

### 8.1 Server-Side Calculation Enforcement
### 8.2 Opaque API Response Validation
### 8.3 JWT Auth on All Protected Routes
### 8.4 Admin 404 (Non-Admin Access)
### 8.5 Graceful Degradation (AI / Email Down)
### 8.6 Timing-Safe Forgot Password
