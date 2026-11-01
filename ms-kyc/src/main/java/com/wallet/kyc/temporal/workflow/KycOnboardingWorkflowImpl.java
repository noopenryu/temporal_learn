package com.wallet.kyc.temporal.workflow;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import com.wallet.kyc.temporal.activity.*;
import com.wallet.kyc.temporal.model.*;
import io.temporal.activity.ActivityOptions;
import io.temporal.common.RetryOptions;
import io.temporal.workflow.Workflow;
import org.slf4j.Logger;

import java.time.Duration;
import java.time.LocalDateTime;

public class KycOnboardingWorkflowImpl implements KycOnboardingWorkflow {

    private static final Logger log = Workflow.getLogger(KycOnboardingWorkflowImpl.class);

    private final ActivityOptions standardActivityOptions = ActivityOptions.newBuilder()
            .setStartToCloseTimeout(Duration.ofMinutes(2))
            .setRetryOptions(RetryOptions.newBuilder()
                    .setInitialInterval(Duration.ofSeconds(1))
                    .setBackoffCoefficient(2.0)
                    .setMaximumInterval(Duration.ofSeconds(10))
                    .setMaximumAttempts(3)
                    .build())
            .build();

    private final DocumentVerificationActivity documentActivity =
            Workflow.newActivityStub(DocumentVerificationActivity.class, standardActivityOptions);

    private final BiometricVerificationActivity biometricActivity =
            Workflow.newActivityStub(BiometricVerificationActivity.class, standardActivityOptions);

    private final SanctionsScreeningActivity sanctionsActivity =
            Workflow.newActivityStub(SanctionsScreeningActivity.class, standardActivityOptions);

    private final RiskAssessmentActivity riskActivity =
            Workflow.newActivityStub(RiskAssessmentActivity.class, standardActivityOptions);

    private final KycPersistenceActivity persistenceActivity =
            Workflow.newActivityStub(KycPersistenceActivity.class, standardActivityOptions);

    private final NotificationActivity notificationActivity =
            Workflow.newActivityStub(NotificationActivity.class, standardActivityOptions);

    private KycStatus currentStatus = KycStatus.PENDING;
    private String currentStage = "INITIATED";
    private int progressPercent = 0;
    private String statusMessage = "Workflow initialized";

    private ManualReviewDecision manualReviewDecision = null;

    @Override
    public KycWorkflowResult executeKyc(KycWorkflowRequest request) {
        String applicationId = request.getApplicationId();
        String customerId = request.getCustomerId();

        log.info("Starting KYC onboarding workflow for customerId={}, applicationId={}", customerId, applicationId);

        // 1. Mark as Processing
        this.currentStatus = KycStatus.PROCESSING;
        this.currentStage = "DOCUMENT_VERIFICATION";
        this.progressPercent = 15;
        this.statusMessage = "Verifying identity document";
        persistenceActivity.updateApplicationStatus(applicationId, KycStatus.PROCESSING, WalletTier.NONE, null, null);

        // 2. ID Document Verification
        DocumentVerificationResult docResult = documentActivity.verifyDocument(request);
        persistenceActivity.recordActivityAudit(applicationId, "DOCUMENT_VERIFICATION",
                docResult.isValid(), docResult.getConfidenceScore(),
                "{\"valid\":" + docResult.isValid() + ",\"reason\":\"" + (docResult.getFailureReason() == null ? "" : docResult.getFailureReason()) + "\"}");

        if (!docResult.isValid()) {
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    docResult.getFailureReason(), 100, false);
        }

        // 3. Biometric / Facial Match Verification
        this.currentStage = "BIOMETRIC_MATCH";
        this.progressPercent = 35;
        this.statusMessage = "Analyzing biometric face match and liveness";
        BiometricVerificationResult bioResult = biometricActivity.verifyBiometrics(request);
        persistenceActivity.recordActivityAudit(applicationId, "BIOMETRIC_MATCH",
                bioResult.isMatched(), bioResult.getFaceMatchScore(),
                "{\"matched\":" + bioResult.isMatched() + ",\"score\":" + bioResult.getFaceMatchScore() + "}");

        // 4. Sanctions & AML Screening
        this.currentStage = "SANCTIONS_PEP";
        this.progressPercent = 55;
        this.statusMessage = "Performing AML & Sanctions screening";
        SanctionsScreeningResult sanctionsResult = sanctionsActivity.screenSanctions(request);
        persistenceActivity.recordActivityAudit(applicationId, "SANCTIONS_PEP",
                sanctionsResult.isClear(), sanctionsResult.isClear() ? 100.0 : 0.0,
                "{\"clear\":" + sanctionsResult.isClear() + ",\"pep\":" + sanctionsResult.isPepMatch() + ",\"sanctions\":" + sanctionsResult.isSanctionsMatch() + "}");

        if (!sanctionsResult.isClear()) {
            String reason = "Sanctions / AML watchlist match detected: " + sanctionsResult.getMatchedList();
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    reason, 100, false);
        }

        // 5. Risk Assessment & Eligibility Check
        this.currentStage = "RISK_ASSESSMENT";
        this.progressPercent = 75;
        this.statusMessage = "Evaluating applicant risk and eligibility";
        RiskAssessmentResult riskResult = riskActivity.assessRisk(request, docResult, bioResult);
        persistenceActivity.recordActivityAudit(applicationId, "RISK_ASSESSMENT",
                riskResult.isAgeEligible() && !riskResult.isDuplicateDetected(),
                (double) riskResult.getRiskScore(),
                "{\"riskScore\":" + riskResult.getRiskScore() + ",\"ageEligible\":" + riskResult.isAgeEligible() + ",\"duplicate\":" + riskResult.isDuplicateDetected() + "}");

        if (!riskResult.isAgeEligible()) {
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    "Applicant does not meet minimum age eligibility requirement (18+)", riskResult.getRiskScore(), false);
        }

        if (riskResult.isDuplicateDetected()) {
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    "Identity document number is already registered to an active account", riskResult.getRiskScore(), false);
        }

        // 6. Decision Engine & Human-in-the-Loop Branching
        if (bioResult.getFaceMatchScore() < 60.0) {
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    "Biometric facial match confidence insufficient (" + bioResult.getFaceMatchScore() + "%)", riskResult.getRiskScore(), false);
        }

        // Auto-Approval condition: biometric score >= 80% and low risk (< 40)
        if (bioResult.getFaceMatchScore() >= 80.0 && riskResult.getRiskScore() < 40) {
            WalletTier tier = (riskResult.getRiskScore() < 20) ? WalletTier.TIER_2_PREMIUM : WalletTier.TIER_1_BASIC;
            return finalizeWorkflow(applicationId, customerId, KycStatus.APPROVED, tier,
                    null, riskResult.getRiskScore(), false);
        }

        // Borderline condition: Requires Manual Compliance Review
        this.currentStatus = KycStatus.REQUIRES_MANUAL_REVIEW;
        this.currentStage = "MANUAL_REVIEW";
        this.progressPercent = 85;
        this.statusMessage = "Application queued for compliance officer manual review";
        persistenceActivity.updateApplicationStatus(applicationId, KycStatus.REQUIRES_MANUAL_REVIEW,
                WalletTier.NONE, riskResult.getRiskScore(), "Pending manual review (Biometric match: " + bioResult.getFaceMatchScore() + "%, Risk: " + riskResult.getRiskScore() + ")");

        log.info("Application {} requires manual review. Waiting for signal...", applicationId);
        boolean signalReceived = Workflow.await(Duration.ofDays(7), () -> this.manualReviewDecision != null);

        if (!signalReceived) {
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    "Manual compliance review timed out after 7 days", riskResult.getRiskScore(), true);
        }

        persistenceActivity.recordManualReview(applicationId,
                manualReviewDecision.getReviewerId(),
                manualReviewDecision.isApproved() ? "APPROVE" : "REJECT",
                manualReviewDecision.getNotes());

        if (manualReviewDecision.isApproved()) {
            return finalizeWorkflow(applicationId, customerId, KycStatus.APPROVED, WalletTier.TIER_1_BASIC,
                    null, riskResult.getRiskScore(), true);
        } else {
            return finalizeWorkflow(applicationId, customerId, KycStatus.REJECTED, WalletTier.NONE,
                    "Manual review rejected: " + manualReviewDecision.getNotes(), riskResult.getRiskScore(), true);
        }
    }

    private KycWorkflowResult finalizeWorkflow(String applicationId, String customerId,
                                               KycStatus status, WalletTier tier,
                                               String rejectionReason, Integer riskScore,
                                               boolean hadManualReview) {
        this.currentStatus = status;
        this.currentStage = "COMPLETED";
        this.progressPercent = 100;
        this.statusMessage = "KYC process completed with status: " + status;

        persistenceActivity.updateApplicationStatus(applicationId, status, tier, riskScore, rejectionReason);

        String notificationMsg = status == KycStatus.APPROVED
                ? "Congratulations! Your e-wallet KYC verification is approved with tier " + tier
                : "Your e-wallet KYC verification was declined: " + rejectionReason;
        notificationActivity.sendCustomerNotification(customerId, applicationId, status, notificationMsg);

        return new KycWorkflowResult(applicationId, customerId, status, tier, rejectionReason, riskScore, hadManualReview, LocalDateTime.now());
    }

    @Override
    public void submitManualReview(ManualReviewDecision decision) {
        log.info("Received manual review decision: reviewer={}, approved={}, notes={}",
                decision.getReviewerId(), decision.isApproved(), decision.getNotes());
        this.manualReviewDecision = decision;
    }

    @Override
    public String getWorkflowStatus() {
        return currentStatus.name();
    }

    @Override
    public KycWorkflowProgress getProgress() {
        return new KycWorkflowProgress(currentStage, progressPercent, statusMessage);
    }
}
