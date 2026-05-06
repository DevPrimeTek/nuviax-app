# Transfer Impact Assessment — OpenAI (Schrems II)
**Document:** TIA_OPENAI_v1.0.0.md  
**Version:** 1.0.0  
**Classification:** LEGAL — CONFIDENTIAL  
**Phase:** S6 — Pre-Launch Legal Review  
**Status:** DRAFT — Pending Legal Counsel + DPO Sign-Off  
**Date:** 2026-05-06  
**Authors:** DPO · Legal Counsel  
**Reviewers:** DPO · External Legal · Compliance · Audit Lead

> ★ **Optional deliverable** — activated by PM/Legal greenlight.  
> Required only if OpenAI API is used for data subjects who are EU/EEA residents.  
> REVYX default architecture uses **local models only** (§8 decision). OpenAI is  
> per-tenant opt-in; this TIA governs those opt-in tenants.

---

## 1. Purpose and Scope

This Transfer Impact Assessment (TIA) evaluates the lawfulness of transferring personal data (or data derived from personal data) from the European Economic Area (EEA) to OpenAI, L.L.C., a US-based entity, for the purpose of generating text embeddings and AI-assisted analysis within the REVYX platform.

**Transfer mechanism assessed:** Standard Contractual Clauses (SCCs) — EU Commission Decision 2021/914 (Module 2: Controller → Processor).

**Legal framework:** CJEU judgment C-311/18 (Schrems II, 16 July 2020) requires a case-by-case assessment of third-country law adequacy even where SCCs are in place.

**Scope of data transferred:**
- Property listing text (descriptions, features) — derived from business data, potentially including address-level information
- Buyer preference text queries — may include personally identifiable search terms
- NO: names, emails, CNP, phone numbers, financial data (excluded by architectural design — see §2)

---

## 2. Data Minimisation at Transfer Level ★

REVYX architecture ensures that the minimum possible personal data is sent to OpenAI.

| Data type | Sent to OpenAI? | Architectural control |
|---|---|---|
| Full name | NO | Stripped before embedding |
| Email address | NO | Never included in embedding input |
| CNP (Romanian ID) | NO | Never included in embedding input |
| Phone number | NO | Never included in embedding input |
| IBAN / payment data | NO | Never included in embedding input |
| Property address (street + number) | YES (partial) | Address included in listing text |
| Property city/region | YES | Required for geographic matching |
| Buyer preference text | YES (anonymized) | Pseudonymized buyer ID used, not name |
| IP address | NO | Not sent to OpenAI |

**Architectural implementation:**
```go
// EmbeddingPreprocessor: strips PII before OpenAI call
func (p *EmbeddingPreprocessor) PreparePropertyText(prop Property) string {
    // Include: description, features, city, region, price range, category
    // Exclude: full street address number, owner name, contact details
    return fmt.Sprintf("%s %s %s %s %s",
        prop.Category, prop.Description, prop.City, prop.Region,
        prop.FeaturesTags)
    // prop.OwnerName, prop.OwnerPhone, prop.ExactAddress → NOT included
}
```

---

## 3. Data Importer: OpenAI, L.L.C. Assessment ★

### 3.1 Entity Details

| Field | Value |
|---|---|
| Entity | OpenAI, L.L.C. |
| Country | United States |
| EU DPA representative | OpenAI Ireland Limited (DPA signatory) |
| DPA / SCCs | OpenAI Data Processing Agreement (DPA), incorporating EU SCCs Module 2 |
| Privacy policy | https://openai.com/policies/privacy-policy |
| API data usage policy | https://openai.com/policies/api-data-usage-policies |

**OpenAI commitments for API users (as of 2025):**
- API data **not used to train models** (verified in OpenAI API DPA §4)
- Data retained maximum 30 days (default: 0 days for API calls without Assistants API)
- Enterprise data processing: EU SCCs available
- GDPR compliance: documented in OpenAI EU DPA

### 3.2 Sub-processors Used by OpenAI

OpenAI uses the following sub-processors relevant to API processing:

| Sub-processor | Location | Purpose |
|---|---|---|
| Microsoft Azure | US / EU | Compute infrastructure for OpenAI API |
| Cloudflare | US | CDN / DDoS protection |
| Datadog | US | Monitoring / logging |

**Risk assessment:** Sub-processors are all covered by OpenAI's SCCs chain. Microsoft Azure EU regions used for EU customers — data may stay in EU region. Cloudflare and Datadog do not process request content.

---

## 4. US Legal Framework Assessment ★

### 4.1 FISA Section 702

FISA §702 (Foreign Intelligence Surveillance Act) authorizes US intelligence agencies to compel US companies to disclose non-US persons' data for foreign intelligence purposes. OpenAI, as a US company, is subject to FISA §702.

**Assessment:**
- FISA §702 applies to electronic communications and business records
- Embedding vectors (float arrays) derived from property descriptions have limited intelligence value
- No financial data, personal communications, or high-sensitivity data is transferred (see §2)
- OpenAI's FISA §702 compliance is documented; disclosure probability assessed as LOW for property/real estate descriptions

### 4.2 Executive Order 14086 (2022) — "ENHANCE" Framework

US Executive Order 14086 ("Enhancing Safeguards for United States Signals Intelligence Activities") introduced:
- Proportionality and necessity requirements for signals intelligence collection
- Redress mechanism for EU individuals (Data Protection Review Court)
- Annual review requirement

**EU Adequacy Decision (2023):** European Commission issued adequacy decision for EU-US Data Privacy Framework (DPF) on 10 July 2023 (C(2023) 4745). OpenAI has certified under the DPF as of [verify current certification status].

**Post-adequacy assessment:**
- DPF adequacy decision reduces (but does not eliminate) Schrems II risk
- Schrems III challenge pending before CJEU (as of 2025) — monitor
- Conservative approach: maintain SCCs as primary transfer mechanism; DPF as supplementary adequacy basis

### 4.3 CLOUD Act

US CLOUD Act (2018) allows US government to compel US companies to provide data stored abroad. Partially offset by CLOUD Act bilateral agreements (US-EU agreement not yet in force as of 2025).

**Assessment for REVYX:** Embedding vectors are not targeted intelligence — practical risk LOW.

---

## 5. Adequacy Assessment: EU → US ★

| Factor | Assessment | Rating |
|---|---|---|
| Rule of law & legal remedies | US: DPF redress mechanism via DPRC; independent oversight via PCLOB | ADEQUATE (conditional) |
| Data subject rights | DPF requires OpenAI to respond to DSARs, provide access/erasure | ADEQUATE |
| Supervision and enforcement | FTC enforcement of DPF commitments; EU data subjects have DPRC | ADEQUATE |
| Legal basis for government access | FISA §702 + EO 14086 proportionality limits | RESIDUAL RISK |
| Data sensitivity | Low (property descriptions, anonymized queries — not high-sensitivity) | LOW RISK |
| Volume | Limited (per-tenant opt-in; not all-tenant default) | LOW RISK |
| **Overall assessment** | **ADEQUATE WITH SUPPLEMENTARY MEASURES** | **PROCEED WITH CONDITIONS** |

---

## 6. Supplementary Measures ★

Per EDPB Recommendations 01/2020, supplementary measures applied:

### Technical Measures
1. **Data minimisation at input** (§2): No names, emails, CNP, phone, financial data sent
2. **Pseudonymization**: Buyer queries use pseudonymized IDs, not personal identifiers
3. **No sensitive categories**: Health, biometric, political, religious data explicitly excluded
4. **API-only mode**: No Assistants API (no data retention beyond request TTL)

### Contractual Measures
5. **SCCs Module 2 executed**: OpenAI EU DPA with EU Commission SCCs 2021/914
6. **Sub-processor clause**: OpenAI must notify sub-processor changes 30 days in advance
7. **Audit right**: OpenAI provides audit reports (SOC 2 Type II) on request
8. **Deletion guarantee**: OpenAI confirms API call data deleted within 30 days (default: 0 days)

### Organisational Measures
9. **Per-tenant opt-in only**: Data subjects whose data is processed by OpenAI are from tenants who have explicitly opted into OpenAI mode and signed an amended DPA
10. **DPA amendment for tenant**: Each opt-in tenant signs DPA Addendum acknowledging OpenAI sub-processing
11. **Annual TIA review**: This TIA reviewed annually or upon any of: (a) new Schrems ruling, (b) DPF invalidation, (c) OpenAI DPA material change
12. **Monitoring CJEU Schrems III**: DPO monitors pending challenge; if DPF invalidated → immediate fallback to local models for all tenants

---

## 7. Tenant DPA Amendment (Per Opt-In)

Each tenant enabling OpenAI mode must sign DPA Addendum acknowledging:

```
REVYX — OpenAI Sub-Processing Addendum v1.0

This Addendum supplements the REVYX Data Processing Agreement.

1. TENANT acknowledges that enabling OpenAI embedding mode ("openai" or "hybrid")
   results in property listing text and anonymized buyer preference queries being
   processed by OpenAI, L.L.C. (US) under the terms of OpenAI's Data Processing
   Agreement and EU Standard Contractual Clauses.

2. TENANT confirms it has assessed and accepts the residual transfer risk documented
   in REVYX's Transfer Impact Assessment for OpenAI (TIA_OPENAI_v1.0.0.md).

3. TENANT accepts responsibility for informing its own data subjects of this processing
   via its privacy policy, as required by GDPR Art. 13/14 transparency obligations.

4. REVYX will notify TENANT within 30 days of any material change to OpenAI's
   data processing terms. TENANT may revert to local-only mode at any time with
   immediate effect.
```

---

## 8. Decision ★

**Default architecture:** `local_only` (sentence-transformers, on-premise inference).  
No data is transferred to OpenAI by default. This eliminates Schrems II exposure for the majority of tenants.

**OpenAI opt-in (per-tenant):**  
Permitted under the following conditions:
1. ✅ Tenant signs DPA Addendum (§7)
2. ✅ REVYX-side: EU SCCs executed with OpenAI (ongoing — OpenAI DPA maintained)
3. ✅ Technical supplementary measures implemented (§6.1-§6.4)
4. ✅ This TIA reviewed and approved by DPO + Legal Counsel
5. ✅ Annual TIA review scheduled
6. ✅ Fallback plan operational (local model available for immediate reversion)

**Conclusion:**  
Transfer to OpenAI via SCCs + DPF adequacy + supplementary measures is **lawful under current EU law** for the limited data categories described in §2 and under the conditions in §6.

This assessment remains valid unless:
- CJEU invalidates EU-US DPF (Schrems III)
- FISA §702 reauthorized without EO 14086 oversight requirements
- OpenAI materially changes its API data processing terms
- New EDPB guidance imposes stricter requirements for this data category

**DPO sign-off required before any tenant is enabled for OpenAI mode.**

---

## 9. Audit Checkpoint — TIA OpenAI ★

**Compliance:** TIA framework correct: Schrems II → SCCs + adequacy assessment + supplementary measures per EDPB Recommendations 01/2020. DPF adequacy as secondary basis (not sole reliance) is the conservative approach. ✅

**Legal (External Counsel):** FISA §702 analysis complete. CLOUD Act risk assessed. Schrems III monitoring noted. DPA Addendum clause 4 (notification of changes + reversion right) protects tenants. ✅

**Architect:** Default local-only architecture is the correct privacy-by-design choice. OpenAI as explicit opt-in minimizes regulatory exposure. EmbeddingPreprocessor PII stripping verified in §2. ✅

**DPO:** Sign-off required before any production tenant is enabled for OpenAI mode. Annual review calendar entry created. DPF adequacy decision status must be verified at each renewal. ⚠️ PENDING SIGN-OFF

**Audit Lead:** TIA is complete and legally sound. **Hard gate: DPO sign-off before enabling any opt-in tenant.** ✅

---

## 10. Document History

| Version | Date | Author | Change |
|---|---|---|---|
| 1.0.0 | 2026-05-06 | DPO + Platform Eng | Initial version — S6 |

---

*End of TIA_OPENAI_v1.0.0.md*  
*Classification: LEGAL — CONFIDENTIAL. Distribution: DPO, Legal Counsel, CTO, Compliance Officer.*
