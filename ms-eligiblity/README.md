# ms-eligiblity

Customer Onboarding Eligibility Microservice for Internet E-Wallet applications.

`ms-eligiblity` evaluates customer qualification during registration/onboarding before triggering KYC identity verification and account creation. It provides both a synchronous REST API and a durable Temporal Activity Worker.

---

## Features
- **REST API with Chi Router**: Synchronous checks, check retrieval, and health status.
- **Temporal Worker & Activities**: Durable activities (`EvaluateEligibilityActivity`, `QueryEligibilityStatusActivity`) for onboarding workflow sagas on `CUSTOMER_ONBOARDING_TASK_QUEUE`.
- **Modular Rules Engine**:
  - `AgeVerificationRule`: Enforces minimum age (default: 18).
  - `JurisdictionSanctionsRule`: Checks FATF blacklisted and sanctioned jurisdictions (e.g. KP, IR, SY, CU, MM).
  - `SanctionsAndBlacklistRule`: Checks customer names and IDs against AML/Sanctions blacklist.
  - `DuplicateIdentityRule`: Prevents duplicate account creation for existing national IDs.
  - `PoliticallyExposedPersonRule`: Flags PEPs for Enhanced Due Diligence (EDD) manual review.
  - `WalletTierRule`: Verifies eligibility for requested wallet tiers (e.g. Premium tier income requirements).
- **PostgreSQL Persistence**: Stores evaluation audit trails, individual rule execution results, and sanction blacklist tables.
- **Embedded Migrations**: Automatically applies database schema on startup via Go `//go:embed`.

---

## API Endpoints

### 1. Check Eligibility
`POST /api/v1/eligibility/check`

**Request Body:**
```json
{
  "customer_id": "cust-vn-1001",
  "national_id": "VN012345678",
  "id_type": "NATIONAL_ID",
  "first_name": "Nguyen",
  "last_name": "Van An",
  "date_of_birth": "1995-05-15T00:00:00Z",
  "nationality": "VN",
  "residence_country": "VN",
  "email": "van.an@example.com",
  "phone_number": "+84901234567",
  "declared_pep": false,
  "wallet_tier_requested": "TIER_1_BASIC"
}
```

**Response (200 OK):**
```json
{
  "id": "1a9be76f-4da3-4748-8a64-9d14fe414887",
  "customer_id": "cust-vn-1001",
  "status": "APPROVED",
  "risk_score": 10,
  "risk_level": "LOW",
  "approved_tier": "TIER_1_BASIC",
  "rejection_reasons": [],
  "rule_results": [
    {
      "rule_name": "AgeVerificationRule",
      "passed": true,
      "severity": "INFO",
      "reason": "Customer meets minimum age requirement (31 >= 18)"
    }
  ],
  "evaluated_at": "2026-09-27T06:31:39.773592Z"
}
```

### 2. Query Check by ID
`GET /api/v1/eligibility/checks/{id}`

### 3. Query Latest Check by Customer ID
`GET /api/v1/eligibility/customers/{customerId}/latest`

### 4. List Active Rules
`GET /api/v1/eligibility/rules`

### 5. Health Check
`GET /healthz`

---

## Temporal Integration

`ms-eligiblity` registers the following activities on task queue `CUSTOMER_ONBOARDING_TASK_QUEUE`:
- `EvaluateEligibilityActivity(ctx context.Context, req domain.EligibilityCheckRequest) (*domain.EligibilityCheckResult, error)`
- `QueryEligibilityStatusActivity(ctx context.Context, checkID string) (*domain.EligibilityCheckResult, error)`

### Invoking from an Onboarding Workflow
```go
ao := workflow.ActivityOptions{
    TaskQueue:           "CUSTOMER_ONBOARDING_TASK_QUEUE",
    StartToCloseTimeout: 10 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
        NonRetryableErrorTypes: []string{
            "ERR_ELIGIBILITY_EVALUATION",
            "ERR_INVALID_CHECK_ID",
            "ERR_CHECK_NOT_FOUND",
        },
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)

var result domain.EligibilityCheckResult
err := workflow.ExecuteActivity(ctx, acts.EvaluateEligibilityActivity, req).Get(ctx, &result)
if err != nil {
    // Handle failure
}

if result.Status == domain.StatusApproved {
    // Proceed to KYC verification
}
```

---

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `PORT` | `8081` | HTTP port |
| `DATABASE_URL` | `postgres://temporal:temporal@localhost:5432/postgres?sslmode=disable` | PostgreSQL connection string |
| `TEMPORAL_HOST_PORT` | `localhost:7233` | Temporal server gRPC address |
| `TEMPORAL_NAMESPACE` | `default` | Temporal namespace |
| `TEMPORAL_TASK_QUEUE` | `CUSTOMER_ONBOARDING_TASK_QUEUE` | Temporal worker task queue |
| `MIN_ELIGIBILITY_AGE` | `18` | Minimum required age for onboarding |
| `RUN_MODE` | `all` | Process mode: `all`, `http`, or `worker` |

---

## Commands

```bash
# Run tests
make test

# Build executable
make build

# Run service
make run
```
