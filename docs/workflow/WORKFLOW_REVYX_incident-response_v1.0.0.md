# WORKFLOW — REVYX Incident Response
**Document:** WORKFLOW_REVYX_incident-response_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S6 — Production Hardening Pre-Launch  
**Status:** APPROVED — Active from Launch Day  
**Date:** 2026-05-06  
**Authors:** SRE · Security · Legal/DPO  
**Reviewers:** Architect · Security · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Reliability | CRITICAL | Structured response reduces MTTR from hours to minutes |
| Legal / Compliance | CRITICAL | GDPR Art. 33: 72h breach notification to CNPDCP is legally binding |
| Tenant relations | HIGH | Transparent communication builds trust; silence damages it |
| Security | HIGH | Defined escalation prevents scope creep and unauthorized actions |
| Culture | MEDIUM | Blameless post-mortems enable learning without fear |

---

## 1. Incident Severity Matrix ★

| Severity | Definition | Examples | Response Time | Escalation |
|---|---|---|---|---|
| **SEV1 — Critical** | Full service outage OR data breach / leak | API 100% down; cross-tenant data exposure; RCE in production | **5 min** declaration; war-room in 15 min | CTO + VP Eng + DPO (if data breach) |
| **SEV2 — High** | Significant degradation OR security vulnerability unmitigated | >10% error rate; match engine v2 down; auth failures >5%; HIGH pen-test finding unpatched | **15 min** declaration; incident channel in 30 min | VP Eng + on-call manager |
| **SEV3 — Medium** | Limited degradation, workaround available | Single tenant impacted; p95 latency spike; scheduler job missed once; non-critical feature broken | **1 hour** acknowledgement | On-call engineer + team lead |
| **SEV4 — Low** | Minor / cosmetic / no user impact | Typo in error message; non-critical log noise; minor dashboard alert | **Next business day** | Ticket only — no escalation |

**Upgrade triggers:**
- SEV3 → SEV2: degradation spreading to more tenants OR no workaround found within 2h
- SEV2 → SEV1: error rate >25% OR any confirmed data breach OR RTO exceeded

---

## 2. On-Call Rotation ★

### 2.1 On-Call Schedule

| Week | Primary On-Call | Secondary (Backup) | Timezone |
|---|---|---|---|
| Week 1 | Engineer A | Engineer B | EET (UTC+3) |
| Week 2 | Engineer B | Engineer C | EET (UTC+3) |
| Week 3 | Engineer C | Engineer A | EET (UTC+3) |

Rotation repeats every 3 weeks. Schedule managed in PagerDuty.

**Coverage hours:**
- Business hours (Mon-Fri 09:00-18:00 EET): Primary on-call; response SLA per severity
- Off-hours (nights/weekends): Primary on-call via PagerDuty; SEV1 only auto-pages

### 2.2 Escalation Tree

```
SEV4 ──────────────────────────────► Ticket queue (Jira)
                                          │
SEV3 ──────────────────► Primary on-call  │
                              │           │
                    No ack within 15min   │
                              │           │
                              ▼           │
SEV2 ──────────────► Secondary on-call   │
                         │               │
               No ack within 15min       │
                         │               │
                         ▼               │
SEV1 ──────────► Team Lead               │
                    │                    │
                    ▼                    │
                VP Engineering           │
                    │                    │
                    ▼                    │
                CTO                      │
                    │                    │
         [Data breach confirmed]         │
                    │                    │
                    ▼                    │
                DPO ─────────────────────┘
```

### 2.3 PagerDuty Policy

```yaml
# pagerduty-policy.yaml (managed in infra/oncall/)
service: revyx-production
escalation_policy:
  - delay: 0     target: primary_oncall
  - delay: 15min target: secondary_oncall
  - delay: 30min target: team_lead
  - delay: 45min target: vp_engineering
  - delay: 60min target: cto

sev1_rules:
  - notify: [primary, secondary, team_lead, vp_engineering]
  - channel: slack:#sev1-war-room
  - call: true  # phone call, not just push notification

sev2_rules:
  - notify: [primary, secondary]
  - channel: slack:#incidents
  - call: false
```

---

## 3. War-Room Protocol ★

### 3.1 Declaration

**SEV1 trigger:** Any engineer can declare SEV1. No approval needed. Err on the side of over-declaring.

```
Slack #incidents:
@channel 🚨 SEV1 DECLARED — [brief description]
Incident ID: INC-YYYY-MM-DD-001
IC: [your name]
War-room: [Zoom link]
Status page: https://status.revyx.app
Time declared: [timestamp UTC]
```

### 3.2 Incident Commander (IC) Role ★

IC role is a **new RBAC role** proposed in this document. IC has:
- Authority to deploy hotfixes without standard review process (emergency bypass — logged)
- Authority to rollback to previous version without approval
- Authority to disable non-critical features (feature flags) unilaterally
- Authority to communicate externally (status page, tenant emails)

IC does NOT have:
- Authority to delete production data
- Authority to modify billing or contracts
- Authority to speak to media (escalate to CEO/VP Comms)

**IC selection:** First senior engineer in war-room nominates themselves. If no senior, team lead assigns.

### 3.3 War-Room Roles

| Role | Responsibility |
|---|---|
| **Incident Commander (IC)** | Drives resolution; coordinates all roles; external comms |
| **Technical Lead** | Root cause investigation; implements fix |
| **Communications Lead** | Status page updates; tenant notifications; Slack #incidents |
| **Scribe** | Timeline log in Jira incident ticket; captures decisions |
| **Support Lead** | Monitors tenant tickets; relays impact information to IC |

### 3.4 War-Room Timeline Protocol

```
T+00min  IC declared. War-room open. Roles assigned.
T+05min  Initial impact assessment: which tenants? what data? what errors?
T+10min  Mitigation options identified (rollback / hotfix / disable feature)
T+15min  Decision: mitigation action chosen. Technical Lead executes.
T+20min  Status page updated: "Investigating [symptom]"
T+30min  Progress check: is mitigation working?
         → If YES: monitor; update status page "Monitoring fix"
         → If NO: escalate, try next option
T+60min  If not resolved: escalate to CTO. Request external support if needed.
         → RTO 30min: if not resolved by T+30 → SEV1 auto-escalates to CTO
Resolution  Status page: "Resolved at [time]. Duration [X]min. Root cause: [brief]"
           Post-mortem scheduled within 5 business days.
```

---

## 4. Post-Mortem Template ★

**Philosophy: Blameless culture.** Incidents are system failures, not people failures. Post-mortems seek to understand contributing factors and prevent recurrence — not assign blame.

---

### POST-MORTEM: INC-[ID] — [Title]

**Severity:** SEV[1/2/3]  
**Date:** [YYYY-MM-DD]  
**Duration:** [X]h [Y]min  
**Incident Commander:** [name]  
**Authors:** [names]  
**Status:** [Draft / Under Review / Final]

---

#### Summary (2-3 sentences)
[What happened, what was the user impact, how was it resolved]

#### Impact
- **Tenants affected:** [count and names]
- **User-facing impact:** [describe: errors, latency, data issues]
- **Data integrity:** [any data loss / corruption / breach?]
- **Revenue impact:** [estimate if applicable]
- **Duration:** [total downtime / degradation period]

#### Timeline (UTC)
| Time | Event |
|---|---|
| HH:MM | Incident first triggered (automated alert or user report) |
| HH:MM | On-call notified |
| HH:MM | IC declared, war-room opened |
| HH:MM | Root cause hypothesized |
| HH:MM | Mitigation applied |
| HH:MM | Service restored |
| HH:MM | Post-mortem scheduled |

#### Root Cause Analysis
[5-why analysis or fishbone diagram]

**Root cause:** [Single sentence root cause statement]

**Contributing factors:**
1. [Factor 1]
2. [Factor 2]
3. [Factor 3]

#### What Went Well
- [Process element that worked well]
- [Detection / alerting that worked]
- [Team communication]

#### What Could Be Improved
- [Process gap]
- [Missing alert or monitoring]
- [Documentation gap]

#### Action Items
| Action | Owner | Due Date | Priority |
|---|---|---|---|
| [Specific preventive action] | [name] | [date] | HIGH/MED/LOW |
| Add alert for [condition] | [name] | [date] | HIGH |
| Update runbook: [section] | [name] | [date] | MED |

#### Lessons Learned
[Key insight for the team that improves future response or prevention]

---

### 4.1 Post-Mortem Review Process

1. Draft circulated to all participants within 48h of resolution
2. Review meeting (30 min): discuss contributing factors, validate action items
3. Action items tracked in Jira with `postmortem` label
4. Published to internal wiki within 5 business days
5. SEV1 post-mortems: shared summary with affected tenants (sanitized)

---

## 5. Tenant Communication ★

### 5.1 Status Page (statuspage.io or equivalent)

**URL:** `https://status.revyx.app`

**Components monitored:**
- API (https://api.revyx.app)
- Match Engine
- Deal Closure Workflow
- Notifications (NBA/Email)
- Admin Panel

**Update frequency during incidents:**
- SEV1: every 15 minutes
- SEV2: every 30 minutes
- SEV3: single update at declaration + resolution

### 5.2 Tenant Email Templates

#### Template 1: Incident Notification (SEV1/SEV2)

```
Subject: [ACTION REQUIRED] Service Incident — REVYX Platform

Dear [Tenant Name],

We are writing to inform you of an ongoing service incident affecting the REVYX platform.

**Status:** Investigating
**Impact:** [Brief description: "Property matching queries are experiencing elevated latency"]
**Tenant Impact:** Your account [is / is not] affected.
**Estimated Resolution:** [Time or "We are working to resolve this as quickly as possible"]

We are actively working to resolve this issue. Updates will be posted to:
https://status.revyx.app

We apologize for the inconvenience and will follow up with a root cause summary once resolved.

REVYX Platform Team
status@revyx.app
```

#### Template 2: Resolution Notification

```
Subject: RESOLVED — Service Incident INC-[ID]

Dear [Tenant Name],

The service incident reported on [DATE] at [TIME] UTC has been resolved.

**Resolution time:** [TIME UTC]
**Duration:** [X hours Y minutes]
**Root cause:** [Brief, non-technical summary]
**Actions taken:** [What was done to fix it]
**Prevention:** [Brief statement on what we're doing to prevent recurrence]

A full post-mortem summary is available at: [link]

We apologize for any disruption this caused. If you have questions or noticed any
data inconsistencies, please contact support@revyx.app.

REVYX Platform Team
```

#### Template 3: Data Breach Notification (GDPR Art. 34)

```
Subject: Important Security Notice — [Brief Description]

Dear [Tenant Name],

We are writing to inform you of a personal data incident that may affect your account.

**Date/time of incident:** [DATE TIME UTC]
**Type of data affected:** [Categories of personal data]
**Approximate number of data subjects:** [Count or estimate]
**Likely consequences:** [Brief description of risks]
**Measures taken:** [What REVYX has done / is doing]

We have reported this incident to the Romanian National Supervisory Authority (CNPDCP)
as required by GDPR Article 33.

**What you should do:**
[Specific actions the tenant/data subjects should take]

**Contact:**
For questions about this incident, contact our Data Protection Officer:
dpo@revyx.app | +40 XXX XXX XXX

We take the security of your data extremely seriously and sincerely apologize for
this incident.

[Name], Data Protection Officer
REVYX Technologies
```

---

## 6. GDPR Art. 33-34: Breach Notification SLA ★

### 6.1 Definitions

| Term | Definition |
|---|---|
| **Personal data breach** | Accidental or unlawful destruction, loss, alteration, unauthorized disclosure of / access to personal data |
| **CNPDCP** | Autoritatea Națională de Supraveghere a Prelucrării Datelor cu Caracter Personal (Romanian DPA) |
| **Art. 33 notification** | Controller must notify CNPDCP within 72 hours of becoming aware |
| **Art. 34 notification** | High-risk breaches: must notify data subjects "without undue delay" |

### 6.2 Detection → Notification Timeline ★

```
T+0h    Breach detected (automated alert or report)
         → SEV1 declared, DPO notified immediately
         
T+1h    Initial assessment:
         - What data? How many records? Which tenants?
         - Is breach ongoing? (If yes: contain first)
         - Likely cause?
         
T+2h    Containment:
         - Revoke compromised credentials/tokens
         - Isolate affected systems if needed
         - Preserve evidence (don't overwrite logs)
         
T+4h    Preliminary CNPDCP notification (if high-risk — Art. 33 para 4 allows incomplete
         notification with supplement to follow)
         DPO submits via CNPDCP online portal
         
T+24h   Full CNPDCP notification submitted (if not yet complete)
         Includes: nature, categories, approximate number, consequences, measures
         
T+48h   Affected tenant notification (Art. 28 DPA obligation — REVYX as processor
         must notify controller without undue delay)
         
T+72h   **HARD DEADLINE:** Full Art. 33 notification to CNPDCP
         If breach is HIGH RISK to individuals → Art. 34 data subject notification begins
         
T+72h+  Post-breach review + DPIA update
```

### 6.3 CNPDCP Notification Content (Art. 33(3))

Required fields:
1. Nature of the personal data breach (categories and approximate number of data subjects + records)
2. Name and contact details of DPO or other contact point
3. Likely consequences of the breach
4. Measures taken or proposed by the controller to address the breach, including mitigation

**CNPDCP notification portal:** https://www.dataprotection.ro/servlet/ViewDocument?id=1924

### 6.4 High-Risk Assessment (Art. 34 trigger)

Art. 34 notification to data subjects required if breach is "likely to result in a high risk to the rights and freedoms of natural persons."

High-risk indicators (any one triggers Art. 34):
- Financial data exposed (IBAN, payment info)
- Authentication credentials (passwords, tokens)
- Special categories data (health, biometric, political views)
- Large-scale exposure (>1000 data subjects)
- Data enabling identity theft or fraud
- Vulnerable population affected (minors)

Low-risk indicators (Art. 34 NOT required):
- Data already publicly available
- Data encrypted and key not compromised
- Small number of data subjects (<10) with minimal risk

### 6.5 Breach Response Checklist

```
□ DPO notified within 1h of detection
□ Breach contained (revoke access, patch vector)
□ Evidence preserved (logs, DB snapshots before patch)
□ Preliminary CNPDCP notification submitted (T+4h if high-risk)
□ Affected tenants notified (T+48h)
□ Full CNPDCP notification submitted (T+72h deadline)
□ Art. 34 data subject notification (if high-risk, "without undue delay")
□ Internal post-mortem + DPIA update
□ Remediation measures documented and implemented
□ CNPDCP follow-up response (if authority requests additional info)
□ Lessons learned incorporated into security controls
```

---

## 7. Runbooks (Index)

| Runbook | Location | Covers |
|---|---|---|
| Match engine fail-back to v1 | `docs/runbooks/match-engine-failback.md` | pgvector degradation → BM25 fallback |
| Database failover | `docs/runbooks/db-failover.md` | Primary failure → WAL-E restore to DR |
| JWT key rotation | `docs/runbooks/jwt-rotation.md` | Emergency key rotation |
| Cross-tenant isolation breach | `docs/runbooks/cross-tenant-breach.md` | Data leakage investigation |
| Error budget depleted | `docs/runbooks/error-budget.md` | SLO budget exhaustion response |
| Embedding cost spike | `docs/runbooks/embedding-cost-spike.md` | OpenAI abuse or bulk reindex runaway |
| GDPR data subject request | `docs/runbooks/gdpr-dsar.md` | DSAR processing procedure |
| Admin account compromise | `docs/runbooks/admin-compromise.md` | Admin credential breach |

---

## 8. Audit Checkpoint — S6 Incident Response ★

**Architect:** SEV matrix is clear and actionable. War-room roles prevent "too many cooks" problem. IC role with emergency bypass authority (logged) is the right balance of speed vs governance. ✅

**Security:** BYPASSRLS for IC (emergency deploy) creates a security gap — mitigated by mandatory audit logging of all IC actions. Compromise of IC credentials would be catastrophic → MFA required for IC role. ✅ (Note: IC credential protection added to pre-launch checklist S6.1 as item S8.)

**Compliance:** GDPR Art. 33 72h deadline with preliminary notification at T+4h is the correct conservative approach. Art. 34 high-risk indicators comprehensive. CNPDCP portal URL documented. T+48h tenant notification under DPA Art. 28 compliant. ✅

**Product:** Status page + tenant email templates ready-to-use. Blameless post-mortem culture explicitly stated. SEV3/4 handled without disrupting business operations. ✅

**Audit Lead:** **Blocking item: status page must be live before launch** (referenced in pre-launch checklist P4). Post-mortem process must be demonstrated — first simulated incident drill recommended in month 1 post-launch. ✅

---

*End of WORKFLOW_REVYX_incident-response_v1.0.0.md*
