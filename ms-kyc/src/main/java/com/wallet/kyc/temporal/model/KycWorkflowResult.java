package com.wallet.kyc.temporal.model;

import com.wallet.kyc.domain.KycStatus;
import com.wallet.kyc.domain.WalletTier;
import java.io.Serializable;
import java.time.LocalDateTime;

public class KycWorkflowResult implements Serializable {
    private String applicationId;
    private String customerId;
    private KycStatus finalStatus;
    private WalletTier walletTier;
    private String rejectionReason;
    private Integer riskScore;
    private boolean requiresManualReview;
    private LocalDateTime completedAt;

    public KycWorkflowResult() {}

    public KycWorkflowResult(String applicationId, String customerId, KycStatus finalStatus,
                             WalletTier walletTier, String rejectionReason, Integer riskScore,
                             boolean requiresManualReview, LocalDateTime completedAt) {
        this.applicationId = applicationId;
        this.customerId = customerId;
        this.finalStatus = finalStatus;
        this.walletTier = walletTier;
        this.rejectionReason = rejectionReason;
        this.riskScore = riskScore;
        this.requiresManualReview = requiresManualReview;
        this.completedAt = completedAt;
    }

    public String getApplicationId() { return applicationId; }
    public void setApplicationId(String applicationId) { this.applicationId = applicationId; }

    public String getCustomerId() { return customerId; }
    public void setCustomerId(String customerId) { this.customerId = customerId; }

    public KycStatus getFinalStatus() { return finalStatus; }
    public void setFinalStatus(KycStatus finalStatus) { this.finalStatus = finalStatus; }

    public WalletTier getWalletTier() { return walletTier; }
    public void setWalletTier(WalletTier walletTier) { this.walletTier = walletTier; }

    public String getRejectionReason() { return rejectionReason; }
    public void setRejectionReason(String rejectionReason) { this.rejectionReason = rejectionReason; }

    public Integer getRiskScore() { return riskScore; }
    public void setRiskScore(Integer riskScore) { this.riskScore = riskScore; }

    public boolean isRequiresManualReview() { return requiresManualReview; }
    public void setRequiresManualReview(boolean requiresManualReview) { this.requiresManualReview = requiresManualReview; }

    public LocalDateTime getCompletedAt() { return completedAt; }
    public void setCompletedAt(LocalDateTime completedAt) { this.completedAt = completedAt; }
}
